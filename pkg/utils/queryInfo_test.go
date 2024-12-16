package utils

import (
	"context"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/useragent"
	"github.com/michelin/snowflake-grafana-datasource/pkg/data"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAddQueryTagInfosWithValidPluginConfig(t *testing.T) {
	ctx := context.Background()

	useragent, _ := useragent.New("8.0.0", "darwin", "amd64")

	config := map[string]string{
		"GF_APP_URL": "http://localhost:3000",
	}
	pluginConfig := &backend.PluginContext{
		PluginVersion: "1.0.0",
		UserAgent:     useragent,
		GrafanaConfig: backend.NewGrafanaCfg(config),
		OrgID:         1,
		DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
			UID: "datasource-uid",
		},
	}
	ctx = backend.WithPluginContext(ctx, *pluginConfig)
	ctx = AddQueryTagInfos(ctx)
	queryTag := fmt.Sprint(ctx)
	expectedTag := `{"datasourceId":"datasource-uid","grafanaHost":"http://localhost:3000","grafanaOrgId":1,"grafanaVersion":"8.0.0","pluginVersion":"1.0.0"}`
	require.Contains(t, queryTag, expectedTag)
}

func TestAddQueryTagInfosWithNilConfig(t *testing.T) {
	ctx := context.Background()

	pluginConfig := &backend.PluginContext{
		PluginVersion:              "1.0.0",
		UserAgent:                  nil,
		GrafanaConfig:              nil,
		OrgID:                      1,
		DataSourceInstanceSettings: nil,
	}
	ctx = backend.WithPluginContext(ctx, *pluginConfig)
	ctx = AddQueryTagInfos(ctx)
	queryTag := fmt.Sprint(ctx)
	expectedTag := `{"datasourceId":"","grafanaHost":"","grafanaOrgId":1,"grafanaVersion":"","pluginVersion":"1.0.0"}`
	require.Contains(t, queryTag, expectedTag)
}

func TestEnrichQueryWithContext(t *testing.T) {
	ctx := context.Background()

	pluginConfig := &backend.PluginContext{
		User: &backend.User{
			Login: "test-user",
		},
	}
	ctx = backend.WithPluginContext(ctx, *pluginConfig)

	timeRange := backend.TimeRange{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	qc := &data.QueryConfigStruct{
		QueryType:  "table",
		TimeRange:  timeRange,
		FinalQuery: "SELECT * FROM test_table",
	}

	expectedQueryInfo := `{"from":"2024-01-01T00:00:00Z","grafanaUser":"test-user","queryType":"table","to":"2024-01-02T00:00:00Z"}`

	enrichedQuery := EnrichQueryWithContext(qc, ctx)
	expectedQuery := qc.FinalQuery + "\n--" + expectedQueryInfo

	require.Equal(t, expectedQuery, enrichedQuery)
}
