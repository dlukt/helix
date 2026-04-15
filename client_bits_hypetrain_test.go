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

func TestBitsAndHypeTrainServicesEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/bits/leaderboard":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("count"); got != "5" {
				t.Fatalf("count = %q, want 5", got)
			}
			if got := r.URL.Query().Get("period"); got != "month" {
				t.Fatalf("period = %q, want month", got)
			}
			if got := r.URL.Query().Get("started_at"); got != startedAt.Format(time.RFC3339) {
				t.Fatalf("started_at = %q, want %q", got, startedAt.Format(time.RFC3339))
			}
			if got := r.URL.Query().Get("user_id"); got != "456" {
				t.Fatalf("user_id = %q, want 456", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"user_id":    "456",
					"user_login": "cheerer",
					"user_name":  "Cheerer",
					"rank":       1,
					"score":      1200,
				}},
				"date_range": map[string]any{
					"started_at": "2024-04-01T00:00:00Z",
					"ended_at":   "2024-04-30T23:59:59Z",
				},
				"total": 1,
			})
		case "/bits/cheermotes":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"prefix": "Cheer",
					"tiers": []map[string]any{{
						"min_bits": 1,
						"id":       "1",
						"color":    "#979797",
						"images": map[string]any{
							"dark": map[string]any{
								"animated": map[string]any{
									"1": "https://example.com/cheer-dark-animated-1.gif",
								},
							},
						},
						"can_cheer":         true,
						"show_in_bits_card": true,
					}},
					"type":          "global_first_party",
					"order":         1,
					"last_updated":  "2024-04-15T10:00:00Z",
					"is_charitable": false,
				}},
			})
		case "/hypetrain/status":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"current": map[string]any{
						"id":                     "train-1",
						"broadcaster_user_id":    "123",
						"broadcaster_user_login": "caster",
						"broadcaster_user_name":  "Caster",
						"total":                  3200,
						"progress":               1800,
						"goal":                   2000,
						"top_contributions": []map[string]any{{
							"user_id":    "456",
							"user_login": "viewer1",
							"user_name":  "Viewer1",
							"type":       "BITS",
							"total":      1000,
						}},
						"shared_train_participants": []map[string]any{{
							"broadcaster_user_id":    "789",
							"broadcaster_user_login": "friendcaster",
							"broadcaster_user_name":  "FriendCaster",
						}},
						"level":           4,
						"started_at":      "2024-04-15T10:00:00Z",
						"expires_at":      "2024-04-15T10:05:00Z",
						"is_shared_train": true,
						"type":            "golden_kappa",
					},
					"all_time_high": map[string]any{
						"level":       8,
						"total":       9000,
						"achieved_at": "2024-04-10T12:00:00Z",
					},
					"shared_all_time_high": map[string]any{
						"level":       10,
						"total":       15000,
						"achieved_at": "2024-04-12T13:00:00Z",
					},
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

	leaderboardResp, _, err := client.Bits.GetLeaderboard(context.Background(), helix.GetBitsLeaderboardParams{
		Count:     5,
		Period:    "month",
		StartedAt: &startedAt,
		UserID:    "456",
	})
	if err != nil {
		t.Fatalf("Bits.GetLeaderboard() error = %v", err)
	}
	if got := leaderboardResp.Total; got != 1 {
		t.Fatalf("Leaderboard total = %d, want 1", got)
	}
	if got := leaderboardResp.Data[0].Score; got != 1200 {
		t.Fatalf("Leaderboard score = %d, want 1200", got)
	}
	if got := leaderboardResp.DateRange.EndedAt; got != "2024-04-30T23:59:59Z" {
		t.Fatalf("Leaderboard ended_at = %q, want 2024-04-30T23:59:59Z", got)
	}

	cheermotesResp, _, err := client.Bits.GetCheermotes(context.Background(), helix.GetCheermotesParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Bits.GetCheermotes() error = %v", err)
	}
	if got := cheermotesResp.Data[0].Prefix; got != "Cheer" {
		t.Fatalf("Cheermote prefix = %q, want Cheer", got)
	}
	if got := cheermotesResp.Data[0].Tiers[0].Images["dark"]["animated"]["1"]; got != "https://example.com/cheer-dark-animated-1.gif" {
		t.Fatalf("Cheermote image = %q, want dark animated 1x gif", got)
	}

	hypeResp, _, err := client.HypeTrain.GetStatus(context.Background(), helix.GetHypeTrainStatusParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("HypeTrain.GetStatus() error = %v", err)
	}
	if hypeResp.Data[0].Current == nil {
		t.Fatal("Current = nil, want active train")
	}
	if got := hypeResp.Data[0].Current.Progress; got != 1800 {
		t.Fatalf("Current progress = %d, want 1800", got)
	}
	if got := hypeResp.Data[0].AllTimeHigh.Total; got != 9000 {
		t.Fatalf("AllTimeHigh total = %d, want 9000", got)
	}
	if got := len(hypeResp.Data[0].Current.SharedTrainParticipants); got != 1 {
		t.Fatalf("SharedTrainParticipants len = %d, want 1", got)
	}
}
