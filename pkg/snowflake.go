package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"net/url"

	"github.com/allegro/bigcache/v3"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type DBDataResponse struct {
	dataResponse backend.DataResponse
	refID        string
}

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
	im            instancemgmt.InstanceManager
	actQueryCount queryCounter
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SnowflakeDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {

	// create response struct
	result := backend.NewQueryDataResponse()

	/*password := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["password"]
	privateKey := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["privateKey"]

	config, err := getConfig(req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		log.DefaultLogger.Error("Could not get config for plugin", "err", err)
		return response, err
	}*/
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
		go td.query(ctx, &wg, ch, instance, query)
		//go e.executeQuery(query, &wg, ctx, ch, queryjson)
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
	Account     string `json:"account"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	Warehouse   string `json:"warehouse"`
	Database    string `json:"database"`
	Schema      string `json:"schema"`
	ExtraConfig string `json:"extraConfig"`
}

func getConfig(settings *backend.DataSourceInstanceSettings) (pluginConfig, error) {
	var config pluginConfig
	err := json.Unmarshal(settings.JSONData, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func getConnectionString(config *pluginConfig, password string, privateKey string) string {
	params := url.Values{}
	params.Add("role", config.Role)
	params.Add("warehouse", config.Warehouse)
	params.Add("database", config.Database)
	params.Add("schema", config.Schema)

	var userPass = ""
	if len(privateKey) != 0 {
		params.Add("authenticator", "SNOWFLAKE_JWT")
		params.Add("privateKey", privateKey)
		userPass = url.User(config.Username).String()
	} else {
		userPass = url.UserPassword(config.Username, password).String()
	}

	return fmt.Sprintf("%s@%s?%s&%s", userPass, config.Account, params.Encode(), config.ExtraConfig)
}

type instanceSettings struct {
	db    *sql.DB
	cache *bigcache.BigCache
}

func newDataSourceInstance(ctx context.Context, setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {

	log.DefaultLogger.Info("Creating instance")
	password := setting.DecryptedSecureJSONData["password"]
	privateKey := setting.DecryptedSecureJSONData["privateKey"]

	config, err := getConfig(&setting)
	if err != nil {
		log.DefaultLogger.Error("Could not get config for plugin", "err", err)
		return nil, err
	}

	connectionString := getConnectionString(&config, password, privateKey)
	db, err := sql.Open("snowflake", connectionString)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(100)                                  //config.DSInfo.JsonData.MaxOpenConns)
	db.SetMaxIdleConns(100)                                  //config.DSInfo.JsonData.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(3600) * time.Second) //time.Duration(14400) * time.Second) //time.Duration(config.DSInfo.JsonData.ConnMaxLifetime) * time.Second)
	cache, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(60*time.Minute))
	return &instanceSettings{db: db, cache: cache}, nil
}

func (s *instanceSettings) Dispose() {
	log.DefaultLogger.Info("Disposing of instance")
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			log.DefaultLogger.Error("Failed to dispose db", "error", err)
		}
	}
	if s.cache != nil {
		if err := s.cache.Close(); err != nil {
			log.DefaultLogger.Error("Failed to dispose db", "error", err)
		}
	}
	log.DefaultLogger.Debug("DB disposed")
}
