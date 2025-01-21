package main

import (
	"context"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/michelin/snowflake-grafana-datasource/pkg/data"
	sf "github.com/snowflakedb/gosnowflake"
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

	t.Run("with User/pass", func(t *testing.T) {
		connectionString := getConnectionString(&config, data.AuthenticationSecret{Password: "password", PrivateKey: "", Token: ""})
		require.Equal(t, "username:password@account?database=database&role=role&schema=schema&warehouse=warehouse&conf=xxx", connectionString)
	})

	t.Run("with private key", func(t *testing.T) {
		connectionString := getConnectionString(&config, data.AuthenticationSecret{Password: "", PrivateKey: "privateKey", Token: ""})
		require.Equal(t, "username@account?authenticator=SNOWFLAKE_JWT&database=database&privateKey=privateKey&role=role&schema=schema&warehouse=warehouse&conf=xxx", connectionString)
	})

	t.Run("with User/pass special char", func(t *testing.T) {
		connectionString := getConnectionString(&config, data.AuthenticationSecret{Password: "p@sswor/d", PrivateKey: "", Token: ""})
		require.Equal(t, "username:p%40sswor%2Fd@account?database=database&role=role&schema=schema&warehouse=warehouse&conf=xxx", connectionString)
	})

	t.Run("with token", func(t *testing.T) {
		connectionString := getConnectionString(&config, data.AuthenticationSecret{Password: "", PrivateKey: "", Token: "xxxxxxxx"})
		require.Equal(t, "account?authenticator=oauth&database=database&role=role&schema=schema&token=xxxxxxxx&warehouse=warehouse&conf=xxx", connectionString)
	})

	config = pluginConfig{
		Account:     "account", // account not escaped, can't have special chars
		Database:    "dat@base",
		Role:        "ro@le",
		Schema:      "sch@ema",
		Username:    "user@name",
		Warehouse:   "ware@house",
		ExtraConfig: "conf=xxx",
	}

	t.Run("with string to escape", func(t *testing.T) {
		passwordIn := "pa$$s+&"
		connectionString := getConnectionString(&config, data.AuthenticationSecret{Password: passwordIn, PrivateKey: "", Token: ""})
		require.Equal(t, "user%40name:pa%24%24s%2B%26@account?database=dat%40base&role=ro%40le&schema=sch%40ema&warehouse=ware%40house&conf=xxx", connectionString)

		dsnParsed, err := sf.ParseDSN(connectionString)
		require.Nil(t, err)
		require.Equal(t, passwordIn, dsnParsed.Password)
		require.Equal(t, config.Account, dsnParsed.Account)
		require.Equal(t, config.Username, dsnParsed.User)
	})
}

func TestCreatesNewDataSourceInstance(t *testing.T) {
	settings := backend.DataSourceInstanceSettings{}
	instance, err := newDataSourceInstance(context.Background(), settings)
	require.NoError(t, err)
	require.NotNil(t, instance)
}

func TestDisposesInstanceWithoutError(t *testing.T) {
	instance := &instanceSettings{}
	require.NotPanics(t, func() {
		instance.Dispose()
	})
}
