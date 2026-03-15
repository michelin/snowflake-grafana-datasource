package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/michelin/snowflake-grafana-datasource/pkg/data"
	sf "github.com/snowflakedb/gosnowflake"
	"github.com/stretchr/testify/require"
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

	t.Run("with PAT", func(t *testing.T) {
		connectionString := getConnectionString(&config, data.AuthenticationSecret{Password: "", PrivateKey: "", Token: "", PAT: "my-pat-token"})
		require.Equal(t, "username@account?authenticator=programmatic_access_token&database=database&role=role&schema=schema&token=my-pat-token&warehouse=warehouse&conf=xxx", connectionString)
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
	instance, err := NewDataSourceInstance(context.Background(), settings)
	require.NoError(t, err)
	require.NotNil(t, instance)
}

func TestDisposesInstanceWithoutError(t *testing.T) {
	instance := &SnowflakeDatasource{}
	require.NotPanics(t, func() {
		instance.Dispose()
	})
}

func TestGetDBReturnsCachedDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	td := &SnowflakeDatasource{db: db, connString: "conn1"}
	defer td.Dispose()

	mock.ExpectClose()

	// getDB with the same connection string should return the cached db
	got, err := td.getDB("conn1")
	require.NoError(t, err)
	require.Equal(t, db, got)
}

func TestGetDBReusesCachedDBOnMultipleCalls(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	td := &SnowflakeDatasource{db: db, connString: "conn1"}
	defer td.Dispose()

	mock.ExpectClose()

	// Multiple calls with same connection string should return the same db
	db1, err := td.getDB("conn1")
	require.NoError(t, err)
	db2, err := td.getDB("conn1")
	require.NoError(t, err)
	require.Equal(t, db1, db2)
}

func TestGetDBRetiresOldPoolGracefully(t *testing.T) {
	db1, mock1, err := sqlmock.New()
	require.NoError(t, err)

	// Use a short grace period so the test doesn't leak goroutines
	td := &SnowflakeDatasource{
		db:                db1,
		connString:        "conn1",
		retireGracePeriod: 50 * time.Millisecond,
	}

	// Simulate a connection string change by swapping to a new mock db
	db2, mock2, err := sqlmock.New()
	require.NoError(t, err)
	td.db = db2
	td.connString = "conn2"

	// Retire old pool — it should NOT be closed immediately
	mock1.ExpectClose()
	td.mu.Lock()
	td.retireDB(db1)
	td.mu.Unlock()
	require.NoError(t, db1.Ping(), "retired pool should still be usable during grace period")

	// Wait for the grace period to expire and verify the old pool was closed
	time.Sleep(100 * time.Millisecond)
	require.NoError(t, mock1.ExpectationsWereMet())

	// Verify new db is returned with new connection string
	got, err := td.getDB("conn2")
	require.NoError(t, err)
	require.Equal(t, db2, got)

	mock2.ExpectClose()
	td.Dispose()
	require.NoError(t, mock2.ExpectationsWereMet())
}

func TestGetDBNoRetirementWhenConnStringUnchanged(t *testing.T) {
	db1, mock1, err := sqlmock.New()
	require.NoError(t, err)

	td := &SnowflakeDatasource{
		db:                db1,
		connString:        "conn1",
		retireGracePeriod: 50 * time.Millisecond,
	}

	// Calling getDB with the same connection string should reuse the pool
	// without triggering any retirement.
	got, err := td.getDB("conn1")
	require.NoError(t, err)
	require.Equal(t, db1, got, "same connString should return same pool")
	require.Empty(t, td.retired, "no retirement when connString unchanged")

	mock1.ExpectClose()
	td.Dispose()
	require.NoError(t, mock1.ExpectationsWereMet())
}

func TestGetDBIsConcurrencySafe(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)

	td := &SnowflakeDatasource{db: db, connString: "conn1"}
	defer td.Dispose()

	const goroutines = 50
	errs := make(chan error, goroutines)

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := td.getDB("conn1")
			if err != nil {
				errs <- err
				return
			}
			if got == nil {
				errs <- fmt.Errorf("getDB returned nil")
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("goroutine error: %v", err)
	}
}

func TestDisposeClosesDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	mock.ExpectClose()

	td := &SnowflakeDatasource{db: db, connString: "conn1"}
	td.Dispose()

	require.Nil(t, td.db)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDisposeStopsRetireTimersAndClosesPools(t *testing.T) {
	dbActive, mockActive, err := sqlmock.New()
	require.NoError(t, err)

	dbRetired, mockRetired, err := sqlmock.New()
	require.NoError(t, err)

	// Use a long grace period so the timer won't fire before Dispose
	td := &SnowflakeDatasource{
		db:                dbActive,
		connString:        "conn2",
		retireGracePeriod: 10 * time.Minute,
	}

	// Simulate a retired pool with a pending timer
	mockRetired.ExpectClose()
	td.mu.Lock()
	td.retireDB(dbRetired)
	td.mu.Unlock()

	// Dispose should stop the timer and close both pools
	mockActive.ExpectClose()
	td.Dispose()

	require.Nil(t, td.db)
	require.Nil(t, td.retired)
	require.NoError(t, mockActive.ExpectationsWereMet())
	require.NoError(t, mockRetired.ExpectationsWereMet())
}

func TestDisposeWithNilDB(t *testing.T) {
	td := &SnowflakeDatasource{}
	require.NotPanics(t, func() {
		td.Dispose()
	})
	require.Nil(t, td.db)
}
