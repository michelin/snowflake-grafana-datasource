package main

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/michelin/snowflake-grafana-datasource/pkg/data"
	_oauth "github.com/michelin/snowflake-grafana-datasource/pkg/oauth"
	_ "github.com/snowflakedb/gosnowflake"
)

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *SnowflakeDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	_, result := createAndValidationConnectionString(req)
	if result != nil {
		return result, nil
	}
	i, err := td.im.Get(ctx, req.PluginContext)
	if err != nil {
		return nil, err
	}
	instance := i.(*instanceSettings)
	db := instance.db

	row, err := db.QueryContext(ctx, "SELECT 1")
	if err != nil {
		return createHealthError(fmt.Sprintf("Validation query error : %s", err)), nil
	}

	defer row.Close()

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}

// createHealthError creates an error result
func createHealthError(message string) *backend.CheckHealthResult {
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusError,
		Message: message,
	}
}

// validateAuthFields validates the authentication fields
func validateAuthFields(password, privateKey string, oauth _oauth.Oauth) *backend.CheckHealthResult {
	if password == "" && privateKey == "" && oauth.ClientSecret == "" {
		return createHealthError("Password or private key or Oauth fields are required.")
	}

	if (password != "" && (privateKey != "" || oauth.ClientSecret != "")) || (privateKey != "" && oauth.ClientSecret != "") {
		return createHealthError("Only one authentication method must be provided.")
	}

	if password == "" && privateKey == "" && (oauth.ClientSecret == "" || oauth.ClientId == "" || oauth.TokenEndpoint == "") {
		return createHealthError("All OAuth fields are mandatory.")
	}

	return nil
}

// createAndValidationConnectionString creates a connection string and validates the configuration
func createAndValidationConnectionString(req *backend.CheckHealthRequest) (string, *backend.CheckHealthResult) {

	config, err := getConfig(req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return "", createHealthError(fmt.Sprintf("Error getting config: %s", err))
	}

	password := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["password"]
	privateKey := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["privateKey"]

	oauth := _oauth.Oauth{
		ClientId:      config.ClientId,
		ClientSecret:  req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["clientSecret"],
		TokenEndpoint: config.TokenEndpoint,
	}

	if validationResult := validateAuthFields(password, privateKey, oauth); validationResult != nil {
		return "", validationResult
	}

	if config.Account == "" {
		return "", createHealthError("Account not provided")
	}

	if config.Username == "" && (password != "" || privateKey != "") {
		return "", createHealthError("Username not provided")
	}

	if len(config.ExtraConfig) > 0 {
		config.ExtraConfig = config.ExtraConfig + "&validateDefaultParameters=true"
	} else {
		config.ExtraConfig = "validateDefaultParameters=true"
	}

	token, err := _oauth.GetToken(oauth, true)
	if err != nil {
		return "", createHealthError(fmt.Sprintf("Error getting token: %s", err))
	}

	authenticationSecret := data.AuthenticationSecret{Password: password, PrivateKey: privateKey, Token: token}

	connectionString := getConnectionString(&config, authenticationSecret)
	return connectionString, nil
}
