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

func TestEntitlementsAndWhispersServicesEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/entitlements/drops":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query()["id"]; len(got) != 2 || got[0] != "ent-1" || got[1] != "ent-2" {
					t.Fatalf("id = %v, want [ent-1 ent-2]", got)
				}
				if got := r.URL.Query().Get("user_id"); got != "456" {
					t.Fatalf("user_id = %q, want 456", got)
				}
				if got := r.URL.Query().Get("game_id"); got != "789" {
					t.Fatalf("game_id = %q, want 789", got)
				}
				if got := r.URL.Query().Get("fulfillment_status"); got != "CLAIMED" {
					t.Fatalf("fulfillment_status = %q, want CLAIMED", got)
				}
				if got := r.URL.Query().Get("first"); got != "25" {
					t.Fatalf("first = %q, want 25", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":                 "ent-1",
						"benefit_id":         "benefit-1",
						"timestamp":          "2024-04-15T10:00:00Z",
						"user_id":            "456",
						"game_id":            "789",
						"fulfillment_status": "CLAIMED",
						"last_updated":       "2024-04-15T10:05:00Z",
					}},
					"pagination": map[string]any{
						"cursor": "next-entitlements",
					},
				})
			case http.MethodPatch:
				var req helix.UpdateDropsEntitlementsRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := len(req.EntitlementIDs); got != 2 {
					t.Fatalf("len(entitlement_ids) = %d, want 2", got)
				}
				if got := req.FulfillmentStatus; got != "FULFILLED" {
					t.Fatalf("fulfillment_status = %q, want FULFILLED", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"status": "SUCCESS",
						"ids":    []string{"ent-1", "ent-2"},
					}},
				})
			default:
				t.Fatalf("unexpected method for /entitlements/drops: %s", r.Method)
			}
		case "/whispers":
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("method = %q, want POST", got)
			}
			if got := r.URL.Query().Get("from_user_id"); got != "123" {
				t.Fatalf("from_user_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("to_user_id"); got != "456" {
				t.Fatalf("to_user_id = %q, want 456", got)
			}
			var req helix.SendWhisperRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got := req.Message; got != "hello there" {
				t.Fatalf("message = %q, want hello there", got)
			}
			w.WriteHeader(http.StatusNoContent)
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

	entitlementsResp, entitlementsMeta, err := client.Entitlements.GetDrops(context.Background(), helix.GetDropsEntitlementsParams{
		CursorParams:      helix.CursorParams{First: 25},
		IDs:               []string{"ent-1", "ent-2"},
		UserID:            "456",
		GameID:            "789",
		FulfillmentStatus: "CLAIMED",
	})
	if err != nil {
		t.Fatalf("Entitlements.GetDrops() error = %v", err)
	}
	if got := entitlementsResp.Data[0].BenefitID; got != "benefit-1" {
		t.Fatalf("BenefitID = %q, want benefit-1", got)
	}
	if got := entitlementsResp.Data[0].Timestamp.UTC(); !got.Equal(time.Date(2024, 4, 15, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("Timestamp = %v, want 2024-04-15T10:00:00Z", got)
	}
	if got := entitlementsMeta.Pagination.Cursor; got != "next-entitlements" {
		t.Fatalf("Entitlements cursor = %q, want next-entitlements", got)
	}

	updateResp, _, err := client.Entitlements.UpdateDrops(context.Background(), helix.UpdateDropsEntitlementsRequest{
		EntitlementIDs:    []string{"ent-1", "ent-2"},
		FulfillmentStatus: "FULFILLED",
	})
	if err != nil {
		t.Fatalf("Entitlements.UpdateDrops() error = %v", err)
	}
	if got := updateResp.Data[0].Status; got != "SUCCESS" {
		t.Fatalf("Update status = %q, want SUCCESS", got)
	}
	if got := len(updateResp.Data[0].IDs); got != 2 {
		t.Fatalf("Updated ids len = %d, want 2", got)
	}

	whisperMeta, err := client.Whispers.Send(context.Background(), helix.SendWhisperParams{
		FromUserID: "123",
		ToUserID:   "456",
	}, helix.SendWhisperRequest{
		Message: "hello there",
	})
	if err != nil {
		t.Fatalf("Whispers.Send() error = %v", err)
	}
	if got := whisperMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("Whispers.Send() status = %d, want %d", got, http.StatusNoContent)
	}
}
