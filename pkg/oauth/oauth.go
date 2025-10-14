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
	Scopes        []string
}

var tokenSource oauth2.TokenSource

// GetToken retrieves a token from the token endpoint (client credentials flow)
func GetToken(oauth Oauth, recreate bool) (string, error) {

	if oauth.ClientId == "" || oauth.ClientSecret == "" || oauth.TokenEndpoint == "" {
		return "", nil
	}

	if tokenSource == nil || recreate {
		config := &clientcredentials.Config{
			ClientID:     oauth.ClientId,
			ClientSecret: oauth.ClientSecret,
			TokenURL:     oauth.TokenEndpoint,
			AuthStyle:    oauth2.AuthStyleAutoDetect,
			Scopes:       oauth.Scopes,
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
