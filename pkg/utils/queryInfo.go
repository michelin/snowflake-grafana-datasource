package utils

import (
	"context"
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/michelin/snowflake-grafana-datasource/pkg/data"
	"github.com/snowflakedb/gosnowflake"
	"time"
)

// AddQueryTagInfos Add Query Tag Infos to the context
func AddQueryTagInfos(ctx context.Context, qc *data.QueryConfigStruct) context.Context {
	// Extract plugin config
	pluginConfig := backend.PluginConfigFromContext(ctx)

	// User Agent
	var grafanaVersion = ""
	if pluginConfig.UserAgent != nil {
		grafanaVersion = pluginConfig.UserAgent.GrafanaVersion()
	}

	// Grafana Host
	var grafanaHost = ""
	if pluginConfig.GrafanaConfig != nil {
		grafanaHost = pluginConfig.GrafanaConfig.Get("GF_APP_URL")
	}

	// Datasource ID
	var grafanaDatasourceID = ""
	if pluginConfig.DataSourceInstanceSettings != nil {
		grafanaDatasourceID = pluginConfig.DataSourceInstanceSettings.UID
	}

	// User
	var grafanaUser = ""
	if pluginConfig.User != nil {
		grafanaUser = pluginConfig.User.Login
	}

	queryTagData := data.QueryTagStruct{
		PluginVersion: pluginConfig.PluginVersion,
		QueryType:     qc.QueryType,
		From:          qc.TimeRange.From.Format(time.RFC3339),
		To:            qc.TimeRange.To.Format(time.RFC3339),
		Grafana: data.QueryTagGrafanaStruct{
			Version:      grafanaVersion,
			Host:         grafanaHost,
			OrgId:        pluginConfig.OrgID,
			User:         grafanaUser,
			DatasourceId: grafanaDatasourceID,
		},
	}
	queryTag, err := json.Marshal(queryTagData)
	if err != nil {
		log.DefaultLogger.Error("could not marshal json: %s\n", err)
		return ctx
	}
	return gosnowflake.WithQueryTag(ctx, string(queryTag))
}
