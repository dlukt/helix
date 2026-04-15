package helix_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dlukt/helix"
)

func TestGoalsAndCharityServicesEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/goals":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":                "goal-1",
					"broadcaster_id":    "123",
					"broadcaster_login": "caster",
					"broadcaster_name":  "Caster",
					"type":              "follower",
					"description":       "Road to 1,000 followers",
					"current_amount":    875,
					"target_amount":     1000,
					"created_at":        "2024-04-15T08:00:00Z",
				}},
			})
		case "/charity/campaigns":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":                  "campaign-1",
					"broadcaster_id":      "123",
					"broadcaster_login":   "caster",
					"broadcaster_name":    "Caster",
					"charity_name":        "Good Cause",
					"charity_description": "Helping people",
					"charity_logo":        "https://example.com/charity.png",
					"charity_website":     "https://example.org",
					"current_amount": map[string]any{
						"value":          25000,
						"decimal_places": 2,
						"currency":       "USD",
					},
					"target_amount": map[string]any{
						"value":          50000,
						"decimal_places": 2,
						"currency":       "USD",
					},
				}},
			})
		case "/charity/donations":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("first"); got != "10" {
				t.Fatalf("first = %q, want 10", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":                "donation-1",
					"campaign_id":       "campaign-1",
					"broadcaster_id":    "123",
					"broadcaster_login": "caster",
					"broadcaster_name":  "Caster",
					"user_id":           "456",
					"user_login":        "donor1",
					"user_name":         "Donor1",
					"amount": map[string]any{
						"value":          1500,
						"decimal_places": 2,
						"currency":       "USD",
					},
					"created_at": "2024-04-15T09:30:00Z",
				}},
				"pagination": map[string]any{
					"cursor": "next-donations",
				},
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

	goalsResp, _, err := client.Goals.Get(context.Background(), helix.GetGoalsParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Goals.Get() error = %v", err)
	}
	if got := goalsResp.Data[0].Type; got != "follower" {
		t.Fatalf("Goal type = %q, want follower", got)
	}
	if got := goalsResp.Data[0].CreatedAt.UTC(); !got.Equal(time.Date(2024, 4, 15, 8, 0, 0, 0, time.UTC)) {
		t.Fatalf("Goal created_at = %v, want 2024-04-15T08:00:00Z", got)
	}

	campaignResp, _, err := client.Charity.GetCampaign(context.Background(), helix.GetCharityCampaignParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Charity.GetCampaign() error = %v", err)
	}
	if got := campaignResp.Data[0].CharityName; got != "Good Cause" {
		t.Fatalf("Charity name = %q, want Good Cause", got)
	}
	if got := campaignResp.Data[0].CurrentAmount.Value; got != 25000 {
		t.Fatalf("CurrentAmount.Value = %d, want 25000", got)
	}

	donationsResp, donationsMeta, err := client.Charity.GetDonations(context.Background(), helix.GetCharityCampaignDonationsParams{
		CursorParams:  helix.CursorParams{First: 10},
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Charity.GetDonations() error = %v", err)
	}
	if got := donationsResp.Data[0].UserLogin; got != "donor1" {
		t.Fatalf("Donation user_login = %q, want donor1", got)
	}
	if got := donationsResp.Data[0].Amount.Value; got != 1500 {
		t.Fatalf("Donation amount value = %d, want 1500", got)
	}
	if got := donationsMeta.Pagination.Cursor; got != "next-donations" {
		t.Fatalf("Donations cursor = %q, want next-donations", got)
	}
}
