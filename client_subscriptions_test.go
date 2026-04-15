package helix_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dlukt/helix"
)

func TestSubscriptionsServiceEncodesRequestsAndDecodesResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/subscriptions":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query()["user_id"]; len(got) != 2 || got[0] != "456" || got[1] != "789" {
				t.Fatalf("user_id = %v, want [456 789]", got)
			}
			if got := r.URL.Query().Get("first"); got != "10" {
				t.Fatalf("first = %q, want 10", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"total":  2,
				"points": 3,
				"data": []map[string]any{{
					"broadcaster_id":    "123",
					"broadcaster_login": "caster",
					"broadcaster_name":  "Caster",
					"gifter_id":         "",
					"gifter_login":      "",
					"gifter_name":       "",
					"is_gift":           false,
					"tier":              "2000",
					"plan_name":         "Tier 2",
					"user_id":           "456",
					"user_name":         "SubscriberOne",
					"user_login":        "subscriberone",
				}},
				"pagination": map[string]any{
					"cursor": "next-subscriptions",
				},
			})
		case "/subscriptions/user":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("user_id"); got != "456" {
				t.Fatalf("user_id = %q, want 456", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"broadcaster_id":    "123",
					"broadcaster_login": "caster",
					"broadcaster_name":  "Caster",
					"is_gift":           false,
					"tier":              "2000",
				}},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	broadcasterResp, broadcasterMeta, err := client.Subscriptions.GetBroadcaster(context.Background(), helix.GetBroadcasterSubscriptionsParams{
		CursorParams:  helix.CursorParams{First: 10},
		BroadcasterID: "123",
		UserIDs:       []string{"456", "789"},
	})
	if err != nil {
		t.Fatalf("Subscriptions.GetBroadcaster() error = %v", err)
	}
	if got := broadcasterResp.Total; got != 2 {
		t.Fatalf("Total = %d, want 2", got)
	}
	if got := broadcasterResp.Points; got != 3 {
		t.Fatalf("Points = %d, want 3", got)
	}
	if got := broadcasterResp.Data[0].PlanName; got != "Tier 2" {
		t.Fatalf("PlanName = %q, want Tier 2", got)
	}
	if got := broadcasterMeta.Pagination.Cursor; got != "next-subscriptions" {
		t.Fatalf("Subscriptions cursor = %q, want next-subscriptions", got)
	}

	userResp, _, err := client.Subscriptions.CheckUser(context.Background(), helix.CheckUserSubscriptionParams{
		BroadcasterID: "123",
		UserID:        "456",
	})
	if err != nil {
		t.Fatalf("Subscriptions.CheckUser() error = %v", err)
	}
	if got := userResp.Data[0].Tier; got != "2000" {
		t.Fatalf("Tier = %q, want 2000", got)
	}
	if got := userResp.Data[0].BroadcasterLogin; got != "caster" {
		t.Fatalf("BroadcasterLogin = %q, want caster", got)
	}
}
