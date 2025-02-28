package oauth

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetToken(t *testing.T) {
	tcs := []struct {
		oauth Oauth
	}{
		{oauth: Oauth{ClientId: "", ClientSecret: "xxx", TokenEndpoint: "xxx"}},
		{oauth: Oauth{ClientId: "xx", ClientSecret: "", TokenEndpoint: "xxx"}},
		{oauth: Oauth{ClientId: "xx", ClientSecret: "xxx", TokenEndpoint: ""}},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			code, err := GetToken(tc.oauth, true)
			require.Nil(t, err)
			require.Empty(t, code)
		})
	}
}

func TestTokenSourceIsRecreatedWhenRequested(t *testing.T) {
	var callCount int
	// Mock token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"access_token": "test_access_token",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
		callCount++
	}))
	defer ts.Close()

	oauth := Oauth{
		ClientId:      "test-client-id",
		ClientSecret:  "test-client-secret",
		TokenEndpoint: ts.URL,
	}

	// First call to GetToken with recreate = true
	token1, err1 := GetToken(oauth, true)
	require.NotEmpty(t, token1)
	require.NoError(t, err1)

	require.Equal(t, 1, callCount)

	// Second call to GetToken with recreate = true
	token2, err2 := GetToken(oauth, true)
	require.NotEmpty(t, token2)
	require.NoError(t, err2)

	require.Equal(t, 2, callCount)
}

func TestTokenSourceIsNotRecreatedWhenNotRequested(t *testing.T) {
	var callCount int
	// Mock token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"access_token": "test_access_token",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
		callCount++
	}))
	defer ts.Close()

	oauth := Oauth{
		ClientId:      "test-client-id",
		ClientSecret:  "test-client-secret",
		TokenEndpoint: ts.URL,
	}

	// First call to GetToken with recreate = true
	token1, err1 := GetToken(oauth, true)
	require.NotEmpty(t, token1)
	require.NoError(t, err1)

	require.Equal(t, 1, callCount)

	// Second call to GetToken with recreate = false
	token2, err2 := GetToken(oauth, false)
	require.NotEmpty(t, token2)
	require.NoError(t, err2)

	require.Equal(t, 1, callCount)
}

func TestErrorWhenTokenCannotBeRetrieved(t *testing.T) {
	var callCount int
	// Mock token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "invalid_client"}`))
		callCount++
	}))
	defer ts.Close()
	oauth := Oauth{
		ClientId:      "invalid-client-id",
		ClientSecret:  "invalid-client-secret",
		TokenEndpoint: ts.URL,
	}

	// Call GetToken with recreate = true
	token, err := GetToken(oauth, true)
	require.Empty(t, token)
	require.Error(t, err)

	require.Contains(t, err.Error(), "oauth2: \"invalid_client\"")
}

func TestGetTokenMissingConfiguration(t *testing.T) {
	oauth := Oauth{}
	token, err := GetToken(oauth, true)
	require.Empty(t, token)
	require.NoError(t, err)
}
