package main

import (
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetConfig(t *testing.T) {

	tcs := []struct {
		json     string
		config   pluginConfig
		response string
		err      string
	}{
		{json: "{}", config: pluginConfig{}},
		{json: "{\"account\":\"test\"}", config: pluginConfig{Account: "test"}},
		{json: "{", err: "unexpected end of JSON input"},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			configStruct := backend.DataSourceInstanceSettings{
				JSONData: []byte(tc.json),
			}
			conf, err := getConfig(&configStruct)
			if tc.err == "" {
				require.NoError(t, err)
				require.Equal(t, tc.config, conf)
			} else {
				require.Error(t, err)
				require.Equal(t, tc.err, err.Error())
			}
		})
	}
}

func TestGetConnectionString(t *testing.T) {

	config := pluginConfig{
		Account:     "account",
		Database:    "database",
		Role:        "role",
		Schema:      "schema",
		Username:    "username",
		Warehouse:   "warehouse",
		ExtraConfig: "conf=xxx",
	}

	t.Run("testcase", func(t *testing.T) {
		connectionString := getConnectionString(&config, "password")
		require.Equal(t, "username:password@account/database/schema?warehouse=warehouse&role=role&conf=xxx", connectionString)
	})
}
