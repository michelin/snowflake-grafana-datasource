package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

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

type SnowflakeDatasource struct {
	db *sql.DB
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
	} else if len(authenticationSecret.Token) != 0 {
		params.Add("authenticator", "oauth")
		params.Add("token", authenticationSecret.Token)
	} else {
		userPass = url.QueryEscape(config.Username) + ":" + url.QueryEscape(authenticationSecret.Password) + "@"
	}
	return fmt.Sprintf("%s%s?%s&%s", userPass, config.Account, params.Encode(), config.ExtraConfig)
}

func NewDataSourceInstance(ctx context.Context, setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Info("Creating instance")
	datasource := &SnowflakeDatasource{}
	return datasource, nil
}

func (s *SnowflakeDatasource) Dispose() {
	log.DefaultLogger.Info("Disposing of instance")
}
