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

func TestChannelPointsServiceEncodesRequestsAndDecodesResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/channel_points/custom_rewards":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query()["id"]; len(got) != 1 || got[0] != "reward-1" {
					t.Fatalf("id = %v, want [reward-1]", got)
				}
				if got := r.URL.Query().Get("only_manageable_rewards"); got != "true" {
					t.Fatalf("only_manageable_rewards = %q, want true", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"broadcaster_name":       "Caster",
						"broadcaster_login":      "caster",
						"broadcaster_id":         "123",
						"id":                     "reward-1",
						"image":                  nil,
						"background_color":       "#00E5CB",
						"is_enabled":             true,
						"cost":                   100,
						"title":                  "Hydrate",
						"prompt":                 "Drink water",
						"is_user_input_required": false,
						"max_per_stream_setting": map[string]any{
							"is_enabled":     true,
							"max_per_stream": 5,
						},
						"max_per_user_per_stream_setting": map[string]any{
							"is_enabled":              true,
							"max_per_user_per_stream": 1,
						},
						"global_cooldown_setting": map[string]any{
							"is_enabled":              true,
							"global_cooldown_seconds": 60,
						},
						"is_paused":                             false,
						"is_in_stock":                           true,
						"default_image":                         map[string]any{"url_1x": "https://example.com/1x.png", "url_2x": "https://example.com/2x.png", "url_4x": "https://example.com/4x.png"},
						"should_redemptions_skip_request_queue": false,
						"redemptions_redeemed_current_stream":   2,
						"cooldown_expires_at":                   "2024-04-15T10:05:00Z",
					}},
				})
			case http.MethodPost:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				var req helix.CreateCustomRewardRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.Title; got != "Hydrate" {
					t.Fatalf("title = %q, want Hydrate", got)
				}
				if req.IsEnabled == nil || !*req.IsEnabled {
					t.Fatalf("is_enabled = %#v, want true", req.IsEnabled)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"broadcaster_name":                      "Caster",
						"broadcaster_login":                     "caster",
						"broadcaster_id":                        "123",
						"id":                                    "reward-2",
						"image":                                 nil,
						"background_color":                      "#00E5CB",
						"is_enabled":                            true,
						"cost":                                  100,
						"title":                                 "Hydrate",
						"prompt":                                "Drink water",
						"is_user_input_required":                false,
						"max_per_stream_setting":                map[string]any{"is_enabled": true, "max_per_stream": 5},
						"max_per_user_per_stream_setting":       map[string]any{"is_enabled": true, "max_per_user_per_stream": 1},
						"global_cooldown_setting":               map[string]any{"is_enabled": true, "global_cooldown_seconds": 60},
						"is_paused":                             false,
						"is_in_stock":                           true,
						"default_image":                         map[string]any{"url_1x": "https://example.com/1x.png", "url_2x": "https://example.com/2x.png", "url_4x": "https://example.com/4x.png"},
						"should_redemptions_skip_request_queue": false,
						"redemptions_redeemed_current_stream":   0,
						"cooldown_expires_at":                   nil,
					}},
				})
			case http.MethodPatch:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("id"); got != "reward-2" {
					t.Fatalf("id = %q, want reward-2", got)
				}
				var req helix.UpdateCustomRewardRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				switch {
				case req.Title != nil:
					if got := *req.Title; got != "Hydrate Now" {
						t.Fatalf("title = %q, want Hydrate Now", got)
					}
					if req.IsPaused == nil || !*req.IsPaused {
						t.Fatalf("is_paused = %#v, want true", req.IsPaused)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{
						"data": []map[string]any{{
							"broadcaster_name":                      "Caster",
							"broadcaster_login":                     "caster",
							"broadcaster_id":                        "123",
							"id":                                    "reward-2",
							"image":                                 nil,
							"background_color":                      "#00E5CB",
							"is_enabled":                            true,
							"cost":                                  100,
							"title":                                 "Hydrate Now",
							"prompt":                                "Drink water",
							"is_user_input_required":                false,
							"max_per_stream_setting":                map[string]any{"is_enabled": true, "max_per_stream": 5},
							"max_per_user_per_stream_setting":       map[string]any{"is_enabled": true, "max_per_user_per_stream": 1},
							"global_cooldown_setting":               map[string]any{"is_enabled": true, "global_cooldown_seconds": 60},
							"is_paused":                             true,
							"is_in_stock":                           true,
							"default_image":                         map[string]any{"url_1x": "https://example.com/1x.png", "url_2x": "https://example.com/2x.png", "url_4x": "https://example.com/4x.png"},
							"should_redemptions_skip_request_queue": false,
							"redemptions_redeemed_current_stream":   0,
							"cooldown_expires_at":                   nil,
						}},
					})
				case req.Prompt != nil:
					if got := *req.Prompt; got != "" {
						t.Fatalf("prompt = %q, want empty string", got)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{
						"data": []map[string]any{{
							"broadcaster_name":                      "Caster",
							"broadcaster_login":                     "caster",
							"broadcaster_id":                        "123",
							"id":                                    "reward-2",
							"image":                                 nil,
							"background_color":                      "#00E5CB",
							"is_enabled":                            true,
							"cost":                                  100,
							"title":                                 "Hydrate Now",
							"prompt":                                "",
							"is_user_input_required":                false,
							"max_per_stream_setting":                map[string]any{"is_enabled": true, "max_per_stream": 5},
							"max_per_user_per_stream_setting":       map[string]any{"is_enabled": true, "max_per_user_per_stream": 1},
							"global_cooldown_setting":               map[string]any{"is_enabled": true, "global_cooldown_seconds": 60},
							"is_paused":                             true,
							"is_in_stock":                           true,
							"default_image":                         map[string]any{"url_1x": "https://example.com/1x.png", "url_2x": "https://example.com/2x.png", "url_4x": "https://example.com/4x.png"},
							"should_redemptions_skip_request_queue": false,
							"redemptions_redeemed_current_stream":   0,
							"cooldown_expires_at":                   nil,
						}},
					})
				default:
					t.Fatal("expected title or prompt update payload")
				}
			case http.MethodDelete:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("id"); got != "reward-2" {
					t.Fatalf("id = %q, want reward-2", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method for /channel_points/custom_rewards: %s", r.Method)
			}
		case "/channel_points/custom_rewards/redemptions":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("reward_id"); got != "reward-1" {
					t.Fatalf("reward_id = %q, want reward-1", got)
				}
				if got := r.URL.Query()["id"]; len(got) != 1 || got[0] != "redemption-1" {
					t.Fatalf("id = %v, want [redemption-1]", got)
				}
				if got := r.URL.Query().Get("status"); got != "UNFULFILLED" {
					t.Fatalf("status = %q, want UNFULFILLED", got)
				}
				if got := r.URL.Query().Get("sort"); got != "NEWEST" {
					t.Fatalf("sort = %q, want NEWEST", got)
				}
				if got := r.URL.Query().Get("first"); got != "10" {
					t.Fatalf("first = %q, want 10", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"broadcaster_name":  "Caster",
						"broadcaster_login": "caster",
						"broadcaster_id":    "123",
						"id":                "redemption-1",
						"user_id":           "456",
						"user_login":        "viewer1",
						"user_name":         "Viewer1",
						"user_input":        "please drink",
						"status":            "UNFULFILLED",
						"redeemed_at":       "2024-04-15T10:10:00Z",
						"reward": map[string]any{
							"id":     "reward-1",
							"title":  "Hydrate",
							"prompt": "Drink water",
							"cost":   100,
						},
					}},
					"pagination": map[string]any{"cursor": "next-redemptions"},
				})
			case http.MethodPatch:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("reward_id"); got != "reward-1" {
					t.Fatalf("reward_id = %q, want reward-1", got)
				}
				if got := r.URL.Query()["id"]; len(got) != 1 || got[0] != "redemption-1" {
					t.Fatalf("id = %v, want [redemption-1]", got)
				}
				var req helix.UpdateCustomRewardRedemptionStatusRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.Status; got != "FULFILLED" {
					t.Fatalf("status = %q, want FULFILLED", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"broadcaster_name":  "Caster",
						"broadcaster_login": "caster",
						"broadcaster_id":    "123",
						"id":                "redemption-1",
						"user_id":           "456",
						"user_login":        "viewer1",
						"user_name":         "Viewer1",
						"user_input":        "please drink",
						"status":            "FULFILLED",
						"redeemed_at":       "2024-04-15T10:10:00Z",
						"reward": map[string]any{
							"id":     "reward-1",
							"title":  "Hydrate",
							"prompt": "Drink water",
							"cost":   100,
						},
					}},
				})
			default:
				t.Fatalf("unexpected method for /channel_points/custom_rewards/redemptions: %s", r.Method)
			}
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

	manageable := true
	rewardsResp, _, err := client.ChannelPoints.GetRewards(context.Background(), helix.GetCustomRewardsParams{
		BroadcasterID:         "123",
		IDs:                   []string{"reward-1"},
		OnlyManageableRewards: &manageable,
	})
	if err != nil {
		t.Fatalf("ChannelPoints.GetRewards() error = %v", err)
	}
	if got := rewardsResp.Data[0].Title; got != "Hydrate" {
		t.Fatalf("Reward title = %q, want Hydrate", got)
	}
	if rewardsResp.Data[0].Image != nil {
		t.Fatalf("Reward image = %#v, want nil", rewardsResp.Data[0].Image)
	}
	if rewardsResp.Data[0].CooldownExpiresAt == nil {
		t.Fatal("CooldownExpiresAt = nil, want timestamp")
	}

	createdResp, _, err := client.ChannelPoints.CreateReward(context.Background(), "123", helix.CreateCustomRewardRequest{
		Title:                             "Hydrate",
		Cost:                              100,
		Prompt:                            "Drink water",
		IsEnabled:                         boolPtr(true),
		IsMaxPerStreamEnabled:             boolPtr(true),
		MaxPerStream:                      intPtr(5),
		IsMaxPerUserPerStreamEnabled:      boolPtr(true),
		MaxPerUserPerStream:               intPtr(1),
		IsGlobalCooldownEnabled:           boolPtr(true),
		GlobalCooldownSeconds:             intPtr(60),
		ShouldRedemptionsSkipRequestQueue: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("ChannelPoints.CreateReward() error = %v", err)
	}
	if got := createdResp.Data[0].ID; got != "reward-2" {
		t.Fatalf("Created reward ID = %q, want reward-2", got)
	}

	updatedResp, _, err := client.ChannelPoints.UpdateReward(context.Background(), helix.UpdateCustomRewardParams{
		BroadcasterID: "123",
		ID:            "reward-2",
	}, helix.UpdateCustomRewardRequest{
		Title:    stringPtr("Hydrate Now"),
		IsPaused: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("ChannelPoints.UpdateReward() error = %v", err)
	}
	if got := updatedResp.Data[0].Title; got != "Hydrate Now" {
		t.Fatalf("Updated reward title = %q, want Hydrate Now", got)
	}
	if !updatedResp.Data[0].IsPaused {
		t.Fatal("Updated reward IsPaused = false, want true")
	}

	clearedPromptResp, _, err := client.ChannelPoints.UpdateReward(context.Background(), helix.UpdateCustomRewardParams{
		BroadcasterID: "123",
		ID:            "reward-2",
	}, helix.UpdateCustomRewardRequest{
		Prompt: stringPtr(""),
	})
	if err != nil {
		t.Fatalf("ChannelPoints.UpdateReward() clear prompt error = %v", err)
	}
	if got := clearedPromptResp.Data[0].Prompt; got != "" {
		t.Fatalf("Cleared reward prompt = %q, want empty string", got)
	}

	deleteMeta, err := client.ChannelPoints.DeleteReward(context.Background(), helix.DeleteCustomRewardParams{
		BroadcasterID: "123",
		ID:            "reward-2",
	})
	if err != nil {
		t.Fatalf("ChannelPoints.DeleteReward() error = %v", err)
	}
	if got := deleteMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("DeleteReward status = %d, want %d", got, http.StatusNoContent)
	}

	redemptionsResp, redemptionsMeta, err := client.ChannelPoints.GetRedemptions(context.Background(), helix.GetCustomRewardRedemptionsParams{
		CursorParams:  helix.CursorParams{First: 10},
		BroadcasterID: "123",
		RewardID:      "reward-1",
		IDs:           []string{"redemption-1"},
		Status:        "UNFULFILLED",
		Sort:          "NEWEST",
	})
	if err != nil {
		t.Fatalf("ChannelPoints.GetRedemptions() error = %v", err)
	}
	if got := redemptionsResp.Data[0].UserInput; got != "please drink" {
		t.Fatalf("Redemption user_input = %q, want please drink", got)
	}
	if got := redemptionsMeta.Pagination.Cursor; got != "next-redemptions" {
		t.Fatalf("Redemptions cursor = %q, want next-redemptions", got)
	}

	updatedRedemptionResp, _, err := client.ChannelPoints.UpdateRedemptionStatus(context.Background(), helix.UpdateCustomRewardRedemptionStatusParams{
		BroadcasterID: "123",
		RewardID:      "reward-1",
		IDs:           []string{"redemption-1"},
	}, helix.UpdateCustomRewardRedemptionStatusRequest{
		Status: "FULFILLED",
	})
	if err != nil {
		t.Fatalf("ChannelPoints.UpdateRedemptionStatus() error = %v", err)
	}
	if got := updatedRedemptionResp.Data[0].Status; got != "FULFILLED" {
		t.Fatalf("Updated redemption status = %q, want FULFILLED", got)
	}
	if got := updatedRedemptionResp.Data[0].RedeemedAt.UTC(); !got.Equal(time.Date(2024, 4, 15, 10, 10, 0, 0, time.UTC)) {
		t.Fatalf("Updated redemption RedeemedAt = %v, want 2024-04-15T10:10:00Z", got)
	}
}
