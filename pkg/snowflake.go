package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"net/url"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/michelin/snowflake-grafana-datasource/pkg/data"
	_oauth "github.com/michelin/snowflake-grafana-datasource/pkg/oauth"

	sf "github.com/snowflakedb/gosnowflake"
)

type DBDataResponse struct {
	dataResponse backend.DataResponse
	refID        string
}

var (
	_ backend.QueryDataHandler = (*SnowflakeDatasource)(nil)
)

type SnowflakeDatasource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SnowflakeDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {

	// create response struct
	result := backend.NewQueryDataResponse()
	i, err := td.im.Get(ctx, req.PluginContext)
	if err != nil {
		return nil, err
	}
	instance := i.(*instanceSettings)
	ch := make(chan DBDataResponse, len(req.Queries))
	var wg sync.WaitGroup
	// Execute each query in a goroutine and wait for them to finish afterwards
	for _, query := range req.Queries {
		wg.Add(1)
		go td.query(ctx, &wg, ch, req, instance, query)
	}

	wg.Wait()

	// Read results from channels
	close(ch)
	result.Responses = make(map[string]backend.DataResponse)
	for queryResult := range ch {
		result.Responses[queryResult.refID] = queryResult.dataResponse
	}

	return result, nil
}

type pluginConfig struct {
	Account                  string `json:"account"`
	Username                 string `json:"username"`
	Role                     string `json:"role"`
	Warehouse                string `json:"warehouse"`
	Database                 string `json:"database"`
	Schema                   string `json:"schema"`
	ExtraConfig              string `json:"extraConfig"`
	MaxChunkDownloadWorkers  string `json:"maxChunkDownloadWorkers"`
	CustomJSONDecoderEnabled bool   `json:"customJSONDecoderEnabled"`
	ClientId                 string `json:"clientId"`
	TokenEndpoint            string `json:"tokenEndpoint"`
	RedirectUrl              string `json:"redirectUrl"`
	MaxOpenConnections       string `json:"maxOpenConnections"`
	IntMaxOpenConnections    int64
	MaxQueuedQueries         string `json:"maxQueuedQueries"`
	IntMaxQueuedQueries      int64
	ConnectionLifetime       string `json:"connectionLifetime"`
	IntConnectionLifetime    int64
}

func getConfig(settings *backend.DataSourceInstanceSettings) (pluginConfig, error) {
	var config pluginConfig
	err := json.Unmarshal(settings.JSONData, &config)
	if config.MaxOpenConnections == "" {
		config.MaxOpenConnections = "100"
	}
	if config.ConnectionLifetime == "" {
		config.ConnectionLifetime = "60"
	}
	if config.MaxQueuedQueries == "" {
		config.MaxQueuedQueries = "400"
	}
	if MaxOpenConnections, err := strconv.Atoi(config.MaxOpenConnections); err == nil {
		config.IntMaxOpenConnections = int64(MaxOpenConnections)
	} else {
		return config, err
	}
	if ConnectionLifetime, err := strconv.Atoi(config.ConnectionLifetime); err == nil {
		config.IntConnectionLifetime = int64(ConnectionLifetime)
	} else {
		return config, err
	}
	if MaxQueuedQueries, err := strconv.Atoi(config.MaxQueuedQueries); err == nil {
		config.IntMaxQueuedQueries = int64(MaxQueuedQueries)
	} else {
		return config, err
	}
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
	} else if len(authenticationSecret.Token) != 0 {
		params.Add("authenticator", "oauth")
		params.Add("token", authenticationSecret.Token)
	} else {
		userPass = url.QueryEscape(config.Username) + ":" + url.QueryEscape(authenticationSecret.Password) + "@"
	}
	return fmt.Sprintf("%s%s?%s&%s", userPass, config.Account, params.Encode(), config.ExtraConfig)
}

type instanceSettings struct {
	db            *sql.DB
	config        *pluginConfig
	actQueryCount queryCounter
}

func NewDataSourceInstance(ctx context.Context, setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {

	log.DefaultLogger.Info("Creating instance")
	config, err := getConfig(&setting)
	password := setting.DecryptedSecureJSONData["password"]
	privateKey := setting.DecryptedSecureJSONData["privateKey"]
	oauth := _oauth.Oauth{
		ClientId:      config.ClientId,
		ClientSecret:  setting.DecryptedSecureJSONData["clientSecret"],
		TokenEndpoint: config.TokenEndpoint,
	}

	token, err := _oauth.GetToken(oauth, false)
	if err != nil {
		return nil, err
	}

	authenticationSecret := data.AuthenticationSecret{
		Password:   password,
		PrivateKey: privateKey,
		Token:      token,
	}

	if err != nil {
		log.DefaultLogger.Error("Could not get config for plugin", "err", err)
		return nil, err
	}

	connectionString := getConnectionString(&config, authenticationSecret)
	db, err := sql.Open("snowflake", connectionString)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(int(config.IntMaxOpenConnections))
	db.SetMaxIdleConns(int(config.IntMaxOpenConnections))
	db.SetConnMaxLifetime(time.Duration(int(config.IntConnectionLifetime)) * time.Minute)
	return &instanceSettings{db: db, config: &config}, nil
}
func (s *instanceSettings) Dispose() {
	log.DefaultLogger.Info("Disposing of instance")
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			log.DefaultLogger.Error("Failed to dispose db", "error", err)
		}
	}
	log.DefaultLogger.Debug("DB disposed")
}
