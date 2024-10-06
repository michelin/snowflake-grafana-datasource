package main

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
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
	password := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["password"]
	privateKey := req.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["privateKey"]

	if password == "" && privateKey == "" {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Password or private key are required.",
		}
	}

	config, err := getConfig(req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Error getting config: %s", err),
		}
	}

	if config.Account == "" {
		return "", &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Account not provided",
		}
	}

	if config.Username == "" {
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

	connectionString := getConnectionString(&config, password, privateKey)
	return connectionString, nil
}
