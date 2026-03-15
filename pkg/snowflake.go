package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/michelin/snowflake-grafana-datasource/pkg/data"
	_oauth "github.com/michelin/snowflake-grafana-datasource/pkg/oauth"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"net/url"

	sf "github.com/snowflakedb/gosnowflake"
)

var (
	_ backend.QueryDataHandler = (*SnowflakeDatasource)(nil)
)

// defaultRetireGracePeriod is the delay before closing a retired connection pool,
// giving in-flight queries time to complete.
const defaultRetireGracePeriod = 1 * time.Minute

type retiredPool struct {
	db    *sql.DB
	timer *time.Timer
}

type SnowflakeDatasource struct {
	mu                sync.Mutex
	db                *sql.DB
	connString        string
	retired           []retiredPool
	retireGracePeriod time.Duration
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SnowflakeDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {

	// create response struct
	response := backend.NewQueryDataResponse()

	config, err := getConfig(req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		log.DefaultLogger.Error("Could not get config for plugin", "err", err)
		return response, err
	}

	password := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["password"]
	privateKey := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["privateKey"]
	pat := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["pat"]
	oauth := _oauth.Oauth{
		ClientId:      config.ClientId,
		ClientSecret:  req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["clientSecret"],
		TokenEndpoint: config.TokenEndpoint,
		Scopes:        config.Scopes,
	}

	token, err := _oauth.GetToken(oauth, false)
	if err != nil {
		return response, err
	}

	authenticationSecret := data.AuthenticationSecret{
		Password:   password,
		PrivateKey: privateKey,
		Token:      token,
		PAT:        pat,
	}

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = td.query(ctx, q, req, config, authenticationSecret)
	}

	return response, nil
}

type pluginConfig struct {
	Account                  string   `json:"account"`
	Username                 string   `json:"username"`
	Role                     string   `json:"role"`
	Warehouse                string   `json:"warehouse"`
	Database                 string   `json:"database"`
	Schema                   string   `json:"schema"`
	ExtraConfig              string   `json:"extraConfig"`
	MaxChunkDownloadWorkers  string   `json:"maxChunkDownloadWorkers"`
	CustomJSONDecoderEnabled bool     `json:"customJSONDecoderEnabled"`
	ClientId                 string   `json:"clientId"`
	TokenEndpoint            string   `json:"tokenEndpoint"`
	RedirectUrl              string   `json:"redirectUrl"`
	Scopes                   []string `json:"scopes"`
}

func getConfig(settings *backend.DataSourceInstanceSettings) (pluginConfig, error) {
	var config pluginConfig
	err := json.Unmarshal(settings.JSONData, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func getConnectionString(config *pluginConfig, authenticationSecret data.AuthenticationSecret) string {
	params := url.Values{}
	params.Add("role", config.Role)
	params.Add("warehouse", config.Warehouse)
	params.Add("database", config.Database)
	params.Add("schema", config.Schema)

	if config.MaxChunkDownloadWorkers != "" {
		n0, err := strconv.Atoi(config.MaxChunkDownloadWorkers)
		if err != nil {
			log.DefaultLogger.Error("invalid value for MaxChunkDownloadWorkers: %v", config.MaxChunkDownloadWorkers)
		}
		sf.MaxChunkDownloadWorkers = n0
	}
	sf.CustomJSONDecoderEnabled = config.CustomJSONDecoderEnabled

	var userPass = ""
	if len(authenticationSecret.PrivateKey) != 0 {
		params.Add("authenticator", "SNOWFLAKE_JWT")
		params.Add("privateKey", authenticationSecret.PrivateKey)
		userPass = url.QueryEscape(config.Username) + "@"
	} else if len(authenticationSecret.PAT) != 0 {
		params.Add("authenticator", "programmatic_access_token")
		params.Add("token", authenticationSecret.PAT)
		userPass = url.QueryEscape(config.Username) + "@"
	} else if len(authenticationSecret.Token) != 0 {
		params.Add("authenticator", "oauth")
		params.Add("token", authenticationSecret.Token)
	} else {
		userPass = url.QueryEscape(config.Username) + ":" + url.QueryEscape(authenticationSecret.Password) + "@"
	}
	return fmt.Sprintf("%s%s?%s&%s", userPass, config.Account, params.Encode(), config.ExtraConfig)
}

func (td *SnowflakeDatasource) gracePeriod() time.Duration {
	if td.retireGracePeriod > 0 {
		return td.retireGracePeriod
	}
	return defaultRetireGracePeriod
}

// getDB returns a cached *sql.DB or creates a new one if the connection string has changed.
// When the connection string changes (e.g. OAuth token refresh), the old pool is retired
// gracefully: it stays usable for in-flight queries and is closed after a grace period.
func (td *SnowflakeDatasource) getDB(connectionString string) (*sql.DB, error) {
	td.mu.Lock()
	defer td.mu.Unlock()

	if td.db != nil && td.connString == connectionString {
		return td.db, nil
	}

	// Open the new pool first so that on failure the old pool remains active.
	newDB, err := sql.Open("snowflake", connectionString)
	if err != nil {
		return nil, err
	}
	newDB.SetConnMaxLifetime(30 * time.Minute)

	// Only retire the old pool after the new one is ready.
	if td.db != nil {
		td.retireDB(td.db)
	}

	td.db = newDB
	td.connString = connectionString
	return newDB, nil
}

// retireDB schedules the given pool to be closed after a grace period,
// allowing in-flight queries to complete. Must be called with td.mu held.
func (td *SnowflakeDatasource) retireDB(db *sql.DB) {
	timer := time.AfterFunc(td.gracePeriod(), func() {
		db.Close()
	})
	td.retired = append(td.retired, retiredPool{db: db, timer: timer})
}

func NewDataSourceInstance(ctx context.Context, setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Info("Creating instance")
	datasource := &SnowflakeDatasource{}
	return datasource, nil
}

func (td *SnowflakeDatasource) Dispose() {
	log.DefaultLogger.Info("Disposing of instance")
	td.mu.Lock()
	defer td.mu.Unlock()

	// Stop pending retire timers and close their pools immediately.
	for _, r := range td.retired {
		if r.timer.Stop() {
			// Timer hadn't fired: we must close the pool ourselves.
			r.db.Close()
		}
		// If Stop() returned false, the timer callback already closed
		// (or is about to close) the pool — no action needed.
	}
	td.retired = nil

	if td.db != nil {
		td.db.Close()
		td.db = nil
	}
}
