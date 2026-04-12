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

func TestEventSubServiceCreateListAndDeleteSubscriptions(t *testing.T) {
	t.Parallel()

	var createSeen, listSeen, deleteSeen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/eventsub/subscriptions":
			createSeen = true
			var req helix.CreateEventSubSubscriptionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got := req.Type; got != "channel.follow" {
				t.Fatalf("Type = %q, want %q", got, "channel.follow")
			}
			if got := req.Transport.Method; got != "webhook" {
				t.Fatalf("Transport.Method = %q, want %q", got, "webhook")
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":      "sub-1",
						"status":  "enabled",
						"type":    "channel.follow",
						"version": "2",
						"cost":    1,
						"condition": map[string]any{
							"broadcaster_user_id": "123",
							"moderator_user_id":   "456",
						},
						"transport": map[string]any{
							"method":   "webhook",
							"callback": "https://example.com/eventsub",
						},
					},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/eventsub/subscriptions":
			listSeen = true
			if got := r.URL.Query().Get("type"); got != "channel.follow" {
				t.Fatalf("type = %q, want %q", got, "channel.follow")
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":      "sub-1",
						"status":  "enabled",
						"type":    "channel.follow",
						"version": "2",
						"cost":    1,
						"transport": map[string]any{
							"method":   "webhook",
							"callback": "https://example.com/eventsub",
						},
					},
				},
				"pagination": map[string]any{
					"cursor": "next-page",
				},
				"total":          1,
				"total_cost":     1,
				"max_total_cost": 10000,
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/eventsub/subscriptions":
			deleteSeen = true
			if got := r.URL.Query().Get("id"); got != "sub-1" {
				t.Fatalf("id = %q, want %q", got, "sub-1")
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
		TokenSource: oauth.StaticSource{
			Value: oauth.Token{AccessToken: "access-token"},
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	created, _, err := client.EventSub.Create(context.Background(), helix.CreateEventSubSubscriptionRequest{
		Type:    "channel.follow",
		Version: "2",
		Condition: helix.EventSubCondition{
			"broadcaster_user_id": "123",
			"moderator_user_id":   "456",
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: "https://example.com/eventsub",
			Secret:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got := created.Data[0].ID; got != "sub-1" {
		t.Fatalf("created.Data[0].ID = %q, want %q", got, "sub-1")
	}

	listed, meta, err := client.EventSub.List(context.Background(), helix.ListEventSubSubscriptionsParams{
		Type: "channel.follow",
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if got := listed.Total; got != 1 {
		t.Fatalf("Total = %d, want 1", got)
	}
	if got := meta.Pagination.Cursor; got != "next-page" {
		t.Fatalf("meta.Pagination.Cursor = %q, want %q", got, "next-page")
	}

	meta, err = client.EventSub.Delete(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if got := meta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", got, http.StatusNoContent)
	}
	if !createSeen || !listSeen || !deleteSeen {
		t.Fatalf("createSeen=%t listSeen=%t deleteSeen=%t, want all true", createSeen, listSeen, deleteSeen)
	}
}

func TestEventSubServiceCreateAndListUseSharedRetryPath(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		calls++
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/eventsub/subscriptions":
			if calls == 1 {
				if got := r.Header.Get("Authorization"); got != "Bearer stale-token" {
					t.Fatalf("Authorization on create first attempt = %q, want stale token", got)
				}
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"Unauthorized","status":401}`))
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer fresh-token" {
				t.Fatalf("Authorization on create retry = %q, want fresh token", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":      "sub-1",
						"status":  "enabled",
						"type":    "channel.follow",
						"version": "2",
						"cost":    1,
						"transport": map[string]any{
							"method":   "webhook",
							"callback": "https://example.com/eventsub",
						},
					},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/eventsub/subscriptions":
			if calls == 3 {
				if got := r.Header.Get("Authorization"); got != "Bearer fresh-token" {
					t.Fatalf("Authorization on list first attempt = %q, want fresh token", got)
				}
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"error":"service unavailable"}`))
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer fresh-token" {
				t.Fatalf("Authorization on list retry = %q, want fresh token", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":      "sub-1",
						"status":  "enabled",
						"type":    "channel.follow",
						"version": "2",
						"cost":    1,
						"transport": map[string]any{
							"method":   "webhook",
							"callback": "https://example.com/eventsub",
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	source := &rotatingTokenSource{current: "stale-token"}
	client, err := helix.New(helix.Config{
		ClientID:    "client-id",
		BaseURL:     server.URL,
		TokenSource: source,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	created, _, err := client.EventSub.Create(context.Background(), helix.CreateEventSubSubscriptionRequest{
		Type:    "channel.follow",
		Version: "2",
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: "https://example.com/eventsub",
			Secret:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got := created.Data[0].ID; got != "sub-1" {
		t.Fatalf("created.Data[0].ID = %q, want %q", got, "sub-1")
	}

	listed, _, err := client.EventSub.List(context.Background(), helix.ListEventSubSubscriptionsParams{Type: "channel.follow"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if got := len(listed.Data); got != 1 {
		t.Fatalf("len(listed.Data) = %d, want 1", got)
	}
	if got := source.invalidations.Load(); got != 1 {
		t.Fatalf("invalidations = %d, want 1", got)
	}
}

func TestEventSubServiceListReturnsMetadataOnDecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Ratelimit-Limit", "800")
		w.Header().Set("Ratelimit-Remaining", "799")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[`))
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, meta, err := client.EventSub.List(context.Background(), helix.ListEventSubSubscriptionsParams{})
	if err == nil {
		t.Fatal("List() error = nil, want decode error")
	}
	if meta == nil {
		t.Fatal("meta = nil, want response metadata on decode error")
	}
	if got := meta.StatusCode; got != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", got, http.StatusOK)
	}
	if got := meta.RateLimit.Remaining; got != 799 {
		t.Fatalf("RateLimit.Remaining = %d, want 799", got)
	}
}

func TestEventSubServiceCreateDoesNotRetry503(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		calls++
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"service unavailable"}`))
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
		TokenSource: oauth.StaticSource{
			Value: oauth.Token{AccessToken: "access-token"},
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, _, err = client.EventSub.Create(context.Background(), helix.CreateEventSubSubscriptionRequest{
		Type:    "channel.follow",
		Version: "2",
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: "https://example.com/eventsub",
			Secret:   "secret",
		},
	})
	if err == nil {
		t.Fatal("Create() error = nil, want APIError")
	}
	if got := calls; got != 1 {
		t.Fatalf("calls = %d, want 1", got)
	}
}
