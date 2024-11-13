package main

import (
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type Oauth struct {
	clientId      string
	clientSecret  string
	tokenEndpoint string
}

var tokenSource oauth2.TokenSource

func getToken(oauth Oauth, recreate bool) (string, error) {
	if oauth.clientId == "" || oauth.clientSecret == "" || oauth.tokenEndpoint == "" {
		return "", nil
	}

	if recreate {
		config := &clientcredentials.Config{
			ClientID:     oauth.clientId,
			ClientSecret: oauth.clientSecret,
			TokenURL:     oauth.tokenEndpoint,
		}

		// Create a TokenSource that caches and refreshes the token automatically
		tokenSource = config.TokenSource(context.Background())
	}

	token, err := tokenSource.Token()
	if err != nil {
		log.DefaultLogger.Error("Could not get token", "err", err)
		return "", err
	}

	return token.AccessToken, nil
}
