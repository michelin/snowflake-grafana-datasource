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
		User: &backend.User{
			Login: "test-user",
		},
	}

	timeRange := backend.TimeRange{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	qc := &data.QueryConfigStruct{
		QueryType:  "table",
		TimeRange:  timeRange,
		FinalQuery: "SELECT * FROM test_table",
	}

	ctx = backend.WithPluginContext(ctx, *pluginConfig)
	ctx = AddQueryTagInfos(ctx, qc)
	queryTag := fmt.Sprint(ctx)
	expectedTag := `{"datasourceId":"datasource-uid","from":"2024-01-01T00:00:00Z","grafanaHost":"http://localhost:3000","grafanaOrgId":1,"grafanaUser":"test-user","grafanaVersion":"8.0.0","pluginVersion":"1.0.0","queryType":"table","to":"2024-01-02T00:00:00Z"}`
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
		User:                       nil,
	}

	timeRange := backend.TimeRange{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	qc := &data.QueryConfigStruct{
		QueryType:  "table",
		TimeRange:  timeRange,
		FinalQuery: "SELECT * FROM test_table",
	}

	ctx = backend.WithPluginContext(ctx, *pluginConfig)
	ctx = AddQueryTagInfos(ctx, qc)
	queryTag := fmt.Sprint(ctx)
	expectedTag := `{"datasourceId":"","from":"2024-01-01T00:00:00Z","grafanaHost":"","grafanaOrgId":1,"grafanaUser":"","grafanaVersion":"","pluginVersion":"1.0.0","queryType":"table","to":"2024-01-02T00:00:00Z"}`
	require.Contains(t, queryTag, expectedTag)
}
