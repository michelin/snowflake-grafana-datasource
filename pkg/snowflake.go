package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	sf "github.com/snowflakedb/gosnowflake"
	"net/url"
)

// newDatasource returns datasource.ServeOpts.
func newDatasource() datasource.ServeOpts {
	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)
	ds := &SnowflakeDatasource{
		im: im,
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

type SnowflakeDatasource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager
	db *sql.DB
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SnowflakeDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {

	// create response struct
	response := backend.NewQueryDataResponse()

	password := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["password"]
	privateKey := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["privateKey"]
	oauth := Oauth{
		clientId:      req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["clientId"],
		clientSecret:  req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["clientSecret"],
		tokenEndpoint: req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["tokenEndpoint"],
	}

	config, err := getConfig(req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		log.DefaultLogger.Error("Could not get config for plugin", "err", err)
		return response, err
	}

	token, err := getToken(oauth, false)
	if err != nil {
		return response, err
	}

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = td.query(ctx, q, req, config, password, privateKey, token)
	}

	return response, nil
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
}

func getConfig(settings *backend.DataSourceInstanceSettings) (pluginConfig, error) {
	var config pluginConfig
	err := json.Unmarshal(settings.JSONData, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func getConnectionString(config *pluginConfig, password string, privateKey string, token string) string {
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
	if len(privateKey) != 0 {
		params.Add("authenticator", "SNOWFLAKE_JWT")
		params.Add("privateKey", privateKey)
		userPass = url.QueryEscape(config.Username) + "@"
	} else if len(token) != 0 {
		params.Add("authenticator", "oauth")
		params.Add("token", token)
	} else {
		userPass = url.QueryEscape(config.Username) + ":" + url.QueryEscape(password) + "@"
	}
	return fmt.Sprintf("%s%s?%s&%s", userPass, config.Account, params.Encode(), config.ExtraConfig)
}

type instanceSettings struct {
}

func newDataSourceInstance(ctx context.Context, setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Info("Creating instance")
	return &instanceSettings{}, nil
}

func (s *instanceSettings) Dispose() {
	log.DefaultLogger.Info("Disposing of instance")
}
