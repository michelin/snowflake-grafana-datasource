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
	// openDB overrides sql.Open for testing. If nil, sql.Open is used.
	openDB func(driverName, dsn string) (*sql.DB, error)
	// onRetireClose is called after a retired pool is closed. For testing only.
	onRetireClose func()
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

func (td *SnowflakeDatasource) openDatabase(connectionString string) (*sql.DB, error) {
	if td.openDB != nil {
		return td.openDB("snowflake", connectionString)
	}
	return sql.Open("snowflake", connectionString)
}

// getDB returns a cached *sql.DB or creates a new one if the connection string has changed.
// When the connection string changes (e.g. OAuth token refresh), the old pool is retired
// gracefully: it stays usable for in-flight queries and is closed after a grace period.
// The lock is released during open+ping so cached-pool callers are not blocked by network I/O.
func (td *SnowflakeDatasource) getDB(ctx context.Context, connectionString string) (*sql.DB, error) {
	// Fast path: return the cached pool without blocking on network I/O.
	td.mu.Lock()
	if td.db != nil && td.connString == connectionString {
		db := td.db
		td.mu.Unlock()
		return db, nil
	}
	td.mu.Unlock()

	// Slow path: open and validate a new pool without holding the lock,
	// so concurrent callers hitting the fast path are not blocked.
	newDB, err := td.openDatabase(connectionString)
	if err != nil {
		return nil, err
	}
	newDB.SetConnMaxLifetime(30 * time.Minute)

	// Validate the new pool before retiring the old one.
	// This avoids swapping a working pool for one that will fail on first use
	// (e.g., invalid/expired OAuth token).
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := newDB.PingContext(pingCtx); err != nil {
		newDB.Close()
		td.mu.Lock()
		oldDB := td.db
		td.mu.Unlock()
		if oldDB != nil {
			log.DefaultLogger.Warn("New connection pool validation failed, keeping existing pool", "err", err)
			return oldDB, nil
		}
		return nil, fmt.Errorf("connection pool validation failed: %w", err)
	}

	// Re-acquire lock to swap the pool. Another goroutine may have already
	// swapped while we were opening/pinging — check again.
	td.mu.Lock()
	defer td.mu.Unlock()

	if td.db != nil && td.connString == connectionString {
		// Another goroutine already updated; discard our new pool.
		newDB.Close()
		return td.db, nil
	}

	// Only retire the old pool after the new one is validated.
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
		td.mu.Lock()
		defer td.mu.Unlock()
		td.removeRetired(db)
		if td.onRetireClose != nil {
			td.onRetireClose()
		}
	})
	td.retired = append(td.retired, retiredPool{db: db, timer: timer})
}

// removeRetired removes the entry for the given db from the retired slice.
// Must be called with td.mu held.
func (td *SnowflakeDatasource) removeRetired(db *sql.DB) {
	for i, r := range td.retired {
		if r.db == db {
			td.retired = append(td.retired[:i], td.retired[i+1:]...)
			return
		}
	}
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
