package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/michelin/snowflake-grafana-datasource/pkg/data"
	_oauth "github.com/michelin/snowflake-grafana-datasource/pkg/oauth"
	"github.com/michelin/snowflake-grafana-datasource/pkg/utils"
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

	oauth := _oauth.Oauth{
		ClientId:      config.ClientId,
		ClientSecret:  req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["clientSecret"],
		TokenEndpoint: config.TokenEndpoint,
	}

	if password == "" && privateKey == "" && oauth.ClientSecret == "" {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Password or private key or Oauth fields are required.",
		}
	}

	if (password != "" && (privateKey != "" || oauth.ClientSecret != "")) || (privateKey != "" && oauth.ClientSecret != "") {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Only one authentication method must be provided.",
		}
	}

	if password == "" && privateKey == "" && (oauth.ClientSecret == "" || oauth.ClientId == "" || oauth.TokenEndpoint == "") {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "All OAuth fields are mandatory.",
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

	token, err := _oauth.GetToken(oauth, true)
	if err != nil {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Error getting token: %s", err),
		}
	}

	authenticationSecret := data.AuthenticationSecret{Password: password, PrivateKey: privateKey, Token: token}

	connectionString := getConnectionString(&config, authenticationSecret)
	return connectionString, nil
}
