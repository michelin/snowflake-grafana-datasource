package main

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
		clientId:      "test-client-id",
		clientSecret:  "test-client-secret",
		tokenEndpoint: ts.URL,
	}

	// First call to getToken with recreate = true
	token1, err1 := getToken(oauth, true)
	require.NotEmpty(t, token1)
	require.NoError(t, err1)

	require.Equal(t, 1, callCount)

	// Second call to getToken with recreate = true
	token2, err2 := getToken(oauth, true)
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
		clientId:      "test-client-id",
		clientSecret:  "test-client-secret",
		tokenEndpoint: ts.URL,
	}

	// First call to getToken with recreate = true
	token1, err1 := getToken(oauth, true)
	require.NotEmpty(t, token1)
	require.NoError(t, err1)

	require.Equal(t, 1, callCount)

	// Second call to getToken with recreate = false
	token2, err2 := getToken(oauth, false)
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
		clientId:      "invalid-client-id",
		clientSecret:  "invalid-client-secret",
		tokenEndpoint: ts.URL,
	}

	// Call getToken with recreate = true
	token, err := getToken(oauth, true)
	require.Empty(t, token)
	require.Error(t, err)

	require.Contains(t, err.Error(), "oauth2: \"invalid_client\"")
}

func TestGetTokenMissingConfiguration(t *testing.T) {
	oauth := Oauth{}
	token, err := getToken(oauth, true)
	require.Empty(t, token)
	require.NoError(t, err)
}
