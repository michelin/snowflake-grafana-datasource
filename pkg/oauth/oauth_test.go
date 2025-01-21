package oauth

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTokenFromCodeMissingParameter(t *testing.T) {
	tcs := []struct {
		oauth Oauth
	}{
		{oauth: Oauth{ClientId: "", ClientSecret: "xxx", TokenEndpoint: "xxx", RedirectUrl: "xxx", Code: "xx"}},
		{oauth: Oauth{ClientId: "xx", ClientSecret: "", TokenEndpoint: "xxx", RedirectUrl: "xxx", Code: "xx"}},
		{oauth: Oauth{ClientId: "xx", ClientSecret: "xxx", TokenEndpoint: "", RedirectUrl: "xxx", Code: "xx"}},
		{oauth: Oauth{ClientId: "xx", ClientSecret: "xxx", TokenEndpoint: "xxx", RedirectUrl: "", Code: "xx"}},
		{oauth: Oauth{ClientId: "xx", ClientSecret: "xxx", TokenEndpoint: "xxx", RedirectUrl: "xxx", Code: ""}},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			code, err := GetTokenFromCode(tc.oauth)
			require.Nil(t, err)
			require.Empty(t, code)
		})
	}
}

func TestGetTokenFromCodeValid(t *testing.T) {
	// Mock token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Authorization"), "Basic Y2xpZW50SWQ6Y2xpZW50U2VjcmV0"; got != want {
			t.Errorf("Authorization header = %q; want %q", got, want)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed reading request body: %s.", err)
		}
		if string(body) != "code=code&grant_type=authorization_code&redirect_uri=redirect" {
			t.Errorf("Unexpected exchange payload; got %q", body)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"access_token": "test_access_token",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
	}))
	defer ts.Close()

	oauth := Oauth{ClientId: "clientId", ClientSecret: "clientSecret", TokenEndpoint: ts.URL, RedirectUrl: "redirect", Code: "code"}

	code, err := GetTokenFromCode(oauth)
	require.Nil(t, err)
	require.Equal(t, "test_access_token", code)
}

func TestGetTokenFromCodeError(t *testing.T) {
	// Mock token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error_message": "invalid_client"}`))
	}))
	defer ts.Close()

	oauth := Oauth{ClientId: "clientId", ClientSecret: "clientSecret", TokenEndpoint: ts.URL, RedirectUrl: "redirect", Code: "code"}

	code, err := GetTokenFromCode(oauth)
	require.Empty(t, code)
	require.Error(t, err, "")
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
