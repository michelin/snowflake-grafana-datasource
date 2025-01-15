package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/michelin/snowflake-grafana-datasource/pkg/data"
	"github.com/michelin/snowflake-grafana-datasource/pkg/utils"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	_ "github.com/snowflakedb/gosnowflake"
)

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *SnowflakeDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {

	connectionString, result := createAndValidationConnectionString(req)
	if result != nil {
		return result, nil
	}
	// Use the existing db field instead of opening a new connection
	if td.db == nil || td.db.Ping() != nil {
		var err error
		td.db, err = sql.Open("snowflake", connectionString)
		if err != nil {
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: fmt.Sprintf("Connection issue : %s", err),
			}, nil
		}
	}
	defer td.db.Close()

	row, err := td.db.QueryContext(utils.AddQueryTagInfos(ctx, &data.QueryConfigStruct{}), "SELECT 1")
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Validation query error : %s", err),
		}, nil
	}

	defer row.Close()

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}

func createAndValidationConnectionString(req *backend.CheckHealthRequest) (string, *backend.CheckHealthResult) {

	config, err := getConfig(req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Error getting config: %s", err),
		}
	}

	password := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["password"]
	privateKey := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["privateKey"]

	oauth := Oauth{
		clientId:      config.ClientId,
		clientSecret:  req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["clientSecret"],
		tokenEndpoint: "https://" + config.Account + "/oauth/token-request",
		code:          req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["code"],
	}

	if password == "" && privateKey == "" && oauth.clientSecret == "" {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Password or private key or Oauth fields are required.",
		}
	}

	if (password != "" && (privateKey != "" || oauth.clientSecret != "")) || (privateKey != "" && oauth.clientSecret != "") {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Only one authentication method must be provided.",
		}
	}

	if password == "" && privateKey == "" && (oauth.clientSecret == "" || oauth.clientId == "" || oauth.tokenEndpoint == "") {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "All Oauth fields are required.",
		}
	}

	if config.Account == "" {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Account not provided",
		}
	}

	if config.Username == "" && (password != "" || privateKey != "") {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Username not provided",
		}
	}

	if len(config.ExtraConfig) > 0 {
		config.ExtraConfig = config.ExtraConfig + "&validateDefaultParameters=true"
	} else {
		config.ExtraConfig = "validateDefaultParameters=true"
	}

	token, err := getTokenFromCode(oauth)
	if err != nil {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Error getting token: %s", err),
		}
	}

	connectionString := getConnectionString(&config, password, privateKey, token)
	return connectionString, nil
}
