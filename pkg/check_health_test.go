package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckHealthWithValidConnection(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	req := &backend.CheckHealthRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				JSONData:                []byte("{\"account\":\"test\",\"username\":\"user\"}"),
				DecryptedSecureJSONData: map[string]string{"password": "pass"},
			},
		},
	}
	ctx := context.Background()
	td := &SnowflakeDatasource{db: db}
	result, err := td.CheckHealth(ctx, req)
	require.NoError(t, err)
	require.Equal(t, backend.HealthStatusOk, result.Status)
	require.Equal(t, "Data source is working", result.Message)
}

func TestCheckHealthWithInvalidConnection(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT 1").WillReturnError(sql.ErrConnDone)

	req := &backend.CheckHealthRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				JSONData:                []byte("{\"account\":\"invalid\",\"username\":\"user\"}"),
				DecryptedSecureJSONData: map[string]string{"password": "pass"},
			},
		},
	}
	ctx := context.Background()
	td := &SnowflakeDatasource{db: db}
	result, err := td.CheckHealth(ctx, req)
	require.NoError(t, err)
	require.Equal(t, backend.HealthStatusError, result.Status)
	require.Contains(t, result.Message, "Validation query error")
}

func TestCheckHealthWithMissingPasswordAndPrivateKey(t *testing.T) {
	req := &backend.CheckHealthRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				JSONData:                []byte("{\"account\":\"test\",\"username\":\"user\"}"),
				DecryptedSecureJSONData: map[string]string{},
			},
		},
	}
	ctx := context.Background()
	td := &SnowflakeDatasource{}
	result, err := td.CheckHealth(ctx, req)
	require.NoError(t, err)
	require.Equal(t, backend.HealthStatusError, result.Status)
	require.Equal(t, "Password or private key or Oauth fields are required.", result.Message)
}

func TestCheckHealthWithInvalidJSONData(t *testing.T) {
	req := &backend.CheckHealthRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				JSONData:                []byte("{"),
				DecryptedSecureJSONData: map[string]string{"password": "pass"},
			},
		},
	}
	ctx := context.Background()
	td := &SnowflakeDatasource{}
	result, err := td.CheckHealth(ctx, req)
	require.NoError(t, err)
	require.Equal(t, backend.HealthStatusError, result.Status)
	require.Equal(t, "Error getting config: unexpected end of JSON input", result.Message)
}

func TestCreateAndValidationConnectionString(t *testing.T) {

	tcs := []struct {
		name             string
		request          *backend.CheckHealthRequest
		result           *backend.CheckHealthResult
		connectionString string
	}{
		{
			name: "Missing Authentication",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{}"),
						DecryptedSecureJSONData: map[string]string{"password": ""},
					},
				},
			},
			result: &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "Password or private key or Oauth fields are required."},
		},
		{
			name: "Bad Json Configuration",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{"),
						DecryptedSecureJSONData: map[string]string{"password": "pass"},
					},
				},
			},
			result: &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "Error getting config: unexpected end of JSON input"},
		},
		{
			name: "missing Account",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{}"),
						DecryptedSecureJSONData: map[string]string{"password": "pass"},
					},
				},
			},
			result: &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "Account not provided"},
		},
		{
			name: "missing Username",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{\"account\":\"test\"}"),
						DecryptedSecureJSONData: map[string]string{"password": "pass"},
					},
				},
			},
			result: &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "Username not provided"},
		},
		{
			name: "multiple Auth Methods Pass And Key",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{\"account\":\"test\"}"),
						DecryptedSecureJSONData: map[string]string{"password": "pass", "privateKey": "xxxxx"},
					},
				},
			},
			result: &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "Only one authentication method must be provided."},
		},
		{
			name: "multiple Auth Methods Pass And Oauth",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{\"account\":\"test\"}"),
						DecryptedSecureJSONData: map[string]string{"password": "pass", "clientSecret": "s"},
					},
				},
			},
			result: &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "Only one authentication method must be provided."},
		},
		{
			name: "multiple Auth Methods Key And Oauth",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{\"account\":\"test\"}"),
						DecryptedSecureJSONData: map[string]string{"clientSecret": "t", "privateKey": "xxxxx"},
					},
				},
			},
			result: &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "Only one authentication method must be provided."},
		},
		{
			name: "valid User Password Auth",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{\"account\":\"test\",\"username\":\"user\"}"),
						DecryptedSecureJSONData: map[string]string{"password": "pass"},
					},
				},
			},
			connectionString: "user:pass@test?database=&role=&schema=&warehouse=&validateDefaultParameters=true",
		},
		{
			name: "missing ClientId And Token Endpoint",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{\"account\":\"test\"}"),
						DecryptedSecureJSONData: map[string]string{"clientSecret": "t"},
					},
				},
			},
			result: &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "All OAuth fields are mandatory. Please click the 'Login with Snowflake' button to proceed before saving the datasource."},
		},
		{
			name: "missing Token Endpoint",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{\"account\":\"test\"}"),
						DecryptedSecureJSONData: map[string]string{"clientId": "t", "clientSecret": "t"},
					},
				},
			},
			result: &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "All OAuth fields are mandatory. Please click the 'Login with Snowflake' button to proceed before saving the datasource."},
		},
		{
			name: "missing ClientId",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{\"account\":\"test\"}"),
						DecryptedSecureJSONData: map[string]string{"tokenEndpoint": "t", "clientSecret": "t"},
					},
				},
			},
			result: &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: "All OAuth fields are mandatory. Please click the 'Login with Snowflake' button to proceed before saving the datasource."},
		},
		{
			name: "valid User Password Auth And ExtraConfig",
			request: &backend.CheckHealthRequest{
				PluginContext: backend.PluginContext{
					DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
						JSONData:                []byte("{\"account\":\"test\",\"username\":\"user\",\"extraConfig\":\"config=conf\"}"),
						DecryptedSecureJSONData: map[string]string{"password": "pass"},
					},
				},
			},
			connectionString: "user:pass@test?database=&role=&schema=&warehouse=&config=conf&validateDefaultParameters=true",
		},
	}
	for _, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %s", tc.name), func(t *testing.T) {
			con, result := createAndValidationConnectionString(tc.request)
			if result == nil {
				require.Equal(t, tc.connectionString, con)
			} else {
				require.Equal(t, tc.result, result)
			}
		})
	}
}

func TestCreateAndValidationConnectionStringWithOauth(t *testing.T) {
	// Mock token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"access_token": "test_access_token",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
	}))
	defer ts.Close()

	req := &backend.CheckHealthRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				JSONData:                []byte("{\"account\":\"test\",\"extraConfig\":\"config=conf\",\"clientId\": \"t\", \"tokenEndpoint\": \"" + ts.URL + "\", \"redirectUrl\": \"redirect\"}"),
				DecryptedSecureJSONData: map[string]string{"code": "xxx", "clientSecret": "t"},
			},
		},
	}
	con, result := createAndValidationConnectionString(req)
	require.Equal(t, "test?authenticator=oauth&database=&role=&schema=&token=test_access_token&warehouse=&config=conf&validateDefaultParameters=true", con)
	require.Nil(t, result)
}

func TestOauthTokenIssue(t *testing.T) {
	// Mock token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "invalid_request"}`))
	}))
	defer ts.Close()

	req := &backend.CheckHealthRequest{
		PluginContext: backend.PluginContext{
			DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
				JSONData:                []byte("{\"account\":\"test\",\"clientId\": \"t\", \"tokenEndpoint\": \"" + ts.URL + "\", \"redirectUrl\": \"redirect\"}"),
				DecryptedSecureJSONData: map[string]string{"clientSecret": "t", "code": "xxx"},
			},
		},
	}
	con, result := createAndValidationConnectionString(req)
	require.Empty(t, con)
	require.Equal(t, result.Status, backend.HealthStatusError)
	require.Equal(t, "Error getting token: oauth2: \"invalid_request\"", result.Message)
}
