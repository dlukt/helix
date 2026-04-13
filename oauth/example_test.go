package oauth_test

import (
	"context"

	"github.com/dlukt/helix/oauth"
)

func ExampleClient_ImplicitAuthorizationURL() {
	client := oauth.NewClient(oauth.Config{
		ClientID: "client-id",
	})

	_ = client.ImplicitAuthorizationURL(oauth.ImplicitAuthorizationURLParams{
		RedirectURI: "https://example.com/callback",
		Scopes:      []string{"chat:read", "chat:edit"},
		State:       "opaque-state",
	})
}

func ExampleNewForceRefreshSource() {
	source := oauth.NewForceRefreshSource(oauth.StaticSource{
		Value: oauth.Token{AccessToken: "access-token"},
	}, func(context.Context, string) bool {
		return true
	})

	_, _ = source.Token(context.Background())
}

func ExampleNewMemoryTokenStore() {
	store := oauth.NewMemoryTokenStore(oauth.Token{
		AccessToken: "access-token",
	})

	_, _ = store.Load(context.Background())
}
