package helix_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dlukt/helix"
	"github.com/dlukt/helix/oauth"
)

func TestUsersGetSendsAuthenticatedRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if got := r.Method; got != http.MethodGet {
			t.Fatalf("method = %q, want %q", got, http.MethodGet)
		}
		if got := r.URL.Path; got != "/users" {
			t.Fatalf("path = %q, want %q", got, "/users")
		}
		if got := r.URL.Query()["id"]; len(got) != 2 || got[0] != "123" || got[1] != "456" {
			t.Fatalf("query id = %v, want [123 456]", got)
		}
		if got := r.Header.Get("Client-Id"); got != "client-id" {
			t.Fatalf("Client-Id = %q, want %q", got, "client-id")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer access-token")
		}
		if got := r.Header.Get("User-Agent"); got != "helix-test" {
			t.Fatalf("User-Agent = %q, want %q", got, "helix-test")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Ratelimit-Limit", "800")
		w.Header().Set("Ratelimit-Remaining", "799")
		w.Header().Set("Ratelimit-Reset", "1712846400")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":           "123",
					"login":        "darko",
					"display_name": "Darko",
				},
			},
			"pagination": map[string]any{
				"cursor": "next-cursor",
			},
		}); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID:  "client-id",
		BaseURL:   server.URL,
		UserAgent: "helix-test",
		TokenSource: oauth.StaticSource{
			Value: oauth.Token{AccessToken: "access-token"},
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, meta, err := client.Users.Get(context.Background(), helix.GetUsersParams{
		IDs: []string{"123", "456"},
	})
	if err != nil {
		t.Fatalf("Users.Get() error = %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("len(resp.Data) = %d, want 1", len(resp.Data))
	}
	if got := resp.Data[0].Login; got != "darko" {
		t.Fatalf("resp.Data[0].Login = %q, want %q", got, "darko")
	}
	if got := meta.Pagination.Cursor; got != "next-cursor" {
		t.Fatalf("meta.Pagination.Cursor = %q, want %q", got, "next-cursor")
	}
	if got := meta.RateLimit.Limit; got != 800 {
		t.Fatalf("meta.RateLimit.Limit = %d, want 800", got)
	}
	if got := meta.RateLimit.Remaining; got != 799 {
		t.Fatalf("meta.RateLimit.Remaining = %d, want 799", got)
	}
	if got := meta.StatusCode; got != http.StatusOK {
		t.Fatalf("meta.StatusCode = %d, want %d", got, http.StatusOK)
	}
}
