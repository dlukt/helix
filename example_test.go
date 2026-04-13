package helix_test

import (
	"context"
	"net/http"

	"github.com/dlukt/helix"
	"github.com/dlukt/helix/oauth"
)

func ExampleNew_appTokenClient() {
	oauthClient := oauth.NewClient(oauth.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		HTTPClient:   http.DefaultClient,
	})

	tokenSource := oauth.NewAppSource(oauthClient, nil)

	_, _ = helix.New(helix.Config{
		ClientID:    "client-id",
		HTTPClient:  http.DefaultClient,
		TokenSource: tokenSource,
		UserAgent:   "helix-example",
	})
}

func ExampleClient_Users() {
	client, _ := helix.New(helix.Config{
		ClientID: "client-id",
		TokenSource: oauth.StaticSource{
			Value: oauth.Token{AccessToken: "access-token"},
		},
	})

	_, _, _ = client.Users.Get(context.Background(), helix.GetUsersParams{
		Logins: []string{"twitchdev"},
	})
}
