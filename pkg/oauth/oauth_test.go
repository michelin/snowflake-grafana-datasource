package oauth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetToken(t *testing.T) {
	tcs := []struct {
		oauth Oauth
	}{
		{oauth: Oauth{ClientId: "", ClientSecret: "xxx", TokenEndpoint: "xxx", Scopes: []string{}}},
		{oauth: Oauth{ClientId: "xx", ClientSecret: "", TokenEndpoint: "xxx", Scopes: []string{}}},
		{oauth: Oauth{ClientId: "xx", ClientSecret: "xxx", TokenEndpoint: "", Scopes: []string{}}},
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
		Scopes:        []string{"read", "write"},
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
		Scopes:        []string{"read", "write"},
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
		Scopes:        []string{},
	}

	// Call GetToken with recreate = true
	token, err := GetToken(oauth, true)
	require.Empty(t, token)
	require.Error(t, err)

	require.Contains(t, err.Error(), "oauth2: \"invalid_client\"")
}

func TestGetTokenMissingConfiguration(t *testing.T) {
	oauth := Oauth{Scopes: []string{}}
	token, err := GetToken(oauth, true)
	require.Empty(t, token)
	require.NoError(t, err)
}

func TestGetTokenWithScopes(t *testing.T) {
	var callCount int
	var receivedScopes string
	// Mock token endpoint that checks for scopes
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse form data to check scopes
		r.ParseForm()
		receivedScopes = r.Form.Get("scope")

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
		Scopes:        []string{"session:role:ACCOUNTADMIN", "refresh_token"},
	}

	// Call GetToken with recreate = true
	token, err := GetToken(oauth, true)
	require.NotEmpty(t, token)
	require.NoError(t, err)
	require.Equal(t, 1, callCount)
	require.Contains(t, receivedScopes, "session:role:ACCOUNTADMIN")
	require.Contains(t, receivedScopes, "refresh_token")
}

func TestGetTokenWithMultipleScopes(t *testing.T) {
	var receivedScopes string
	// Mock token endpoint that checks for scopes
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse form data to check scopes
		r.ParseForm()
		receivedScopes = r.Form.Get("scope")

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"access_token": "test_access_token",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
	}))
	defer ts.Close()

	testCases := []struct {
		name     string
		scopes   []string
		expected []string
	}{
		{
			name:     "single scope",
			scopes:   []string{"session:role:ACCOUNTADMIN"},
			expected: []string{"session:role:ACCOUNTADMIN"},
		},
		{
			name:     "multiple scopes",
			scopes:   []string{"session:role:ACCOUNTADMIN", "refresh_token", "openid"},
			expected: []string{"session:role:ACCOUNTADMIN", "refresh_token", "openid"},
		},
		{
			name:     "empty scopes",
			scopes:   []string{},
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oauth := Oauth{
				ClientId:      "test-client-id",
				ClientSecret:  "test-client-secret",
				TokenEndpoint: ts.URL,
				Scopes:        tc.scopes,
			}

			token, err := GetToken(oauth, true)
			require.NotEmpty(t, token)
			require.NoError(t, err)

			if len(tc.expected) > 0 {
				for _, expectedScope := range tc.expected {
					require.Contains(t, receivedScopes, expectedScope)
				}
			}
		})
	}
}
