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
func AddQueryTagInfos(ctx context.Context) context.Context {
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

	queryTagData := map[string]interface{}{
		"pluginVersion":  pluginConfig.PluginVersion,
		"grafanaVersion": grafanaVersion,
		"grafanaHost":    grafanaHost,
		"grafanaOrgId":   pluginConfig.OrgID,
		"datasourceId":   grafanaDatasourceID,
	}
	queryTag, err := json.Marshal(queryTagData)
	if err != nil {
		log.DefaultLogger.Error("could not marshal json: %s\n", err)
		return ctx
	}
	return gosnowflake.WithQueryTag(ctx, string(queryTag))
}

// Enrich the query with some data in the context
func EnrichQueryWithContext(qc *data.QueryConfigStruct, ctx context.Context) string {
	// Extract plugin config
	pluginConfig := backend.PluginConfigFromContext(ctx)

	// User
	var grafanaUser = ""
	if pluginConfig.User != nil {
		grafanaUser = pluginConfig.User.Login
	}

	queryTagData := map[string]interface{}{
		"queryType":   qc.QueryType,
		"from":        qc.TimeRange.From.Format(time.RFC3339),
		"to":          qc.TimeRange.To.Format(time.RFC3339),
		"grafanaUser": grafanaUser,
	}
	queryInfo, err := json.Marshal(queryTagData)
	if err != nil {
		log.DefaultLogger.Error("could not marshal json: %s\n", err)
		return qc.FinalQuery
	}
	return qc.FinalQuery + "\n--" + string(queryInfo)
}
