package oauth

import (
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type Oauth struct {
	ClientId      string
	ClientSecret  string
	TokenEndpoint string
	RedirectUrl   string
	Code          string
}

var tokenSource oauth2.TokenSource

// GetTokenFromCode get token from code oauth flow
func GetTokenFromCode(oauth Oauth) (string, error) {
	if oauth.ClientId == "" || oauth.ClientSecret == "" || oauth.TokenEndpoint == "" || oauth.Code == "" || oauth.RedirectUrl == "" {
		return "", nil
	}

	config := &oauth2.Config{
		ClientID:     oauth.ClientId,
		ClientSecret: oauth.ClientSecret,
		RedirectURL:  oauth.RedirectUrl,
		Endpoint: oauth2.Endpoint{
			TokenURL:  oauth.TokenEndpoint,
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}

	token, err := config.Exchange(context.Background(), oauth.Code)
	if err != nil {
		log.DefaultLogger.Error("Could not get token", "err", err)
		return "", err
	}

	return token.AccessToken, nil
}

// Get
func GetToken(oauth Oauth, recreate bool) (string, error) {
	if oauth.ClientId == "" || oauth.ClientSecret == "" || oauth.TokenEndpoint == "" {
		return "", nil
	}

	if recreate {
		config := &clientcredentials.Config{
			ClientID:     oauth.ClientId,
			ClientSecret: oauth.ClientSecret,
			TokenURL:     oauth.TokenEndpoint,
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
