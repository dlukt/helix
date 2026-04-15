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

func TestGuestStarServiceEncodesRequestsAndDecodesResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/guest_star/channel_settings":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("moderator_id = %q, want 456", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"is_moderator_send_live_enabled":  true,
						"slot_count":                      4,
						"is_browser_source_audio_enabled": true,
						"layout":                          "TILED_LAYOUT",
						"browser_source_token":            "browser-token",
					}},
				})
			case http.MethodPut:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("is_moderator_send_live_enabled"); got != "false" {
					t.Fatalf("is_moderator_send_live_enabled = %q, want false", got)
				}
				if got := r.URL.Query().Get("slot_count"); got != "6" {
					t.Fatalf("slot_count = %q, want 6", got)
				}
				if got := r.URL.Query().Get("is_browser_source_audio_enabled"); got != "true" {
					t.Fatalf("is_browser_source_audio_enabled = %q, want true", got)
				}
				if got := r.URL.Query().Get("group_layout"); got != "HORIZONTAL_LAYOUT" {
					t.Fatalf("group_layout = %q, want HORIZONTAL_LAYOUT", got)
				}
				if got := r.URL.Query().Get("regenerate_browser_sources"); got != "true" {
					t.Fatalf("regenerate_browser_sources = %q, want true", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method for /guest_star/channel_settings: %s", r.Method)
			}
		case "/guest_star/session":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("moderator_id = %q, want 456", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id": "session-1",
						"guests": []map[string]any{
							{
								"id":                "0",
								"user_id":           "123",
								"user_display_name": "Caster",
								"user_login":        "caster",
								"is_live":           true,
								"volume":            100,
								"assigned_at":       "2024-04-15T09:00:00Z",
								"audio_settings": map[string]any{
									"is_available":     true,
									"is_host_enabled":  true,
									"is_guest_enabled": true,
								},
								"video_settings": map[string]any{
									"is_available":     true,
									"is_host_enabled":  true,
									"is_guest_enabled": true,
								},
							},
							{
								"slot_id":           "1",
								"user_id":           "789",
								"user_display_name": "Guest",
								"user_login":        "guest",
								"is_live":           false,
								"volume":            80,
								"assigned_at":       "2024-04-15T09:05:00Z",
								"audio_settings": map[string]any{
									"is_available":     true,
									"is_host_enabled":  true,
									"is_guest_enabled": false,
								},
								"video_settings": map[string]any{
									"is_available":     true,
									"is_host_enabled":  false,
									"is_guest_enabled": true,
								},
							},
						},
					}},
				})
			case http.MethodPost:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id": "session-2",
						"guests": []map[string]any{{
							"id":                "0",
							"user_id":           "123",
							"user_display_name": "Caster",
							"user_login":        "caster",
							"is_live":           true,
							"volume":            100,
							"assigned_at":       "2024-04-15T10:00:00Z",
							"audio_settings": map[string]any{
								"is_available":     true,
								"is_host_enabled":  true,
								"is_guest_enabled": true,
							},
							"video_settings": map[string]any{
								"is_available":     true,
								"is_host_enabled":  true,
								"is_guest_enabled": true,
							},
						}},
					}},
				})
			case http.MethodDelete:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("session_id"); got != "session-2" {
					t.Fatalf("session_id = %q, want session-2", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":     "session-2",
						"guests": []map[string]any{},
					}},
				})
			default:
				t.Fatalf("unexpected method for /guest_star/session: %s", r.Method)
			}
		case "/guest_star/invites":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("moderator_id = %q, want 456", got)
				}
				if got := r.URL.Query().Get("session_id"); got != "session-1" {
					t.Fatalf("session_id = %q, want session-1", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"user_id":            "789",
						"invited_at":         "2024-04-15T09:02:00Z",
						"status":             "READY",
						"is_audio_enabled":   true,
						"is_video_enabled":   false,
						"is_audio_available": true,
						"is_video_available": true,
					}},
				})
			case http.MethodPost:
				if got := r.URL.Query().Get("guest_id"); got != "789" {
					t.Fatalf("guest_id = %q, want 789", got)
				}
				w.WriteHeader(http.StatusNoContent)
			case http.MethodDelete:
				if got := r.URL.Query().Get("guest_id"); got != "789" {
					t.Fatalf("guest_id = %q, want 789", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method for /guest_star/invites: %s", r.Method)
			}
		case "/guest_star/slot":
			switch r.Method {
			case http.MethodPost:
				if got := r.URL.Query().Get("guest_id"); got != "789" {
					t.Fatalf("guest_id = %q, want 789", got)
				}
				if got := r.URL.Query().Get("slot_id"); got != "1" {
					t.Fatalf("slot_id = %q, want 1", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"code": "USER_NOT_READY",
					},
				})
			case http.MethodPatch:
				if got := r.URL.Query().Get("source_slot_id"); got != "1" {
					t.Fatalf("source_slot_id = %q, want 1", got)
				}
				if got := r.URL.Query().Get("destination_slot_id"); got != "2" {
					t.Fatalf("destination_slot_id = %q, want 2", got)
				}
				w.WriteHeader(http.StatusNoContent)
			case http.MethodDelete:
				if got := r.URL.Query().Get("slot_id"); got != "1" {
					t.Fatalf("slot_id = %q, want 1", got)
				}
				if got := r.URL.Query().Get("should_reinvite_guest"); got != "true" {
					t.Fatalf("should_reinvite_guest = %q, want true", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method for /guest_star/slot: %s", r.Method)
			}
		case "/guest_star/slot_settings":
			if got := r.Method; got != http.MethodPatch {
				t.Fatalf("method = %q, want PATCH", got)
			}
			if got := r.URL.Query().Get("slot_id"); got != "1" {
				t.Fatalf("slot_id = %q, want 1", got)
			}
			if got := r.URL.Query().Get("is_audio_enabled"); got != "false" {
				t.Fatalf("is_audio_enabled = %q, want false", got)
			}
			if got := r.URL.Query().Get("is_video_enabled"); got != "true" {
				t.Fatalf("is_video_enabled = %q, want true", got)
			}
			if got := r.URL.Query().Get("is_live"); got != "true" {
				t.Fatalf("is_live = %q, want true", got)
			}
			if got := r.URL.Query().Get("volume"); got != "65" {
				t.Fatalf("volume = %q, want 65", got)
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

	settingsResp, _, err := client.GuestStar.GetChannelSettings(context.Background(), helix.GetGuestStarChannelSettingsParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	})
	if err != nil {
		t.Fatalf("GuestStar.GetChannelSettings() error = %v", err)
	}
	if got := settingsResp.Data[0].GroupLayout; got != "TILED_LAYOUT" {
		t.Fatalf("GroupLayout = %q, want TILED_LAYOUT", got)
	}
	if got := settingsResp.Data[0].BrowserSourceToken; got != "browser-token" {
		t.Fatalf("BrowserSourceToken = %q, want browser-token", got)
	}

	moderatorLive := false
	slotCount := 6
	audioEnabled := true
	regenerate := true
	if _, err := client.GuestStar.UpdateChannelSettings(context.Background(), helix.UpdateGuestStarChannelSettingsParams{
		BroadcasterID:                "123",
		IsModeratorSendLiveEnabled:   &moderatorLive,
		SlotCount:                    &slotCount,
		IsBrowserSourceAudioEnabled:  &audioEnabled,
		GroupLayout:                  "HORIZONTAL_LAYOUT",
		RegenerateBrowserSourceToken: &regenerate,
	}); err != nil {
		t.Fatalf("GuestStar.UpdateChannelSettings() error = %v", err)
	}

	sessionResp, _, err := client.GuestStar.GetSession(context.Background(), helix.GetGuestStarSessionParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	})
	if err != nil {
		t.Fatalf("GuestStar.GetSession() error = %v", err)
	}
	if got := sessionResp.Data[0].Guests[0].SlotID; got != "0" {
		t.Fatalf("Host SlotID = %q, want 0", got)
	}
	if got := sessionResp.Data[0].Guests[1].SlotID; got != "1" {
		t.Fatalf("Guest SlotID = %q, want 1", got)
	}
	if got := sessionResp.Data[0].Guests[1].AssignedAt.UTC(); !got.Equal(time.Date(2024, 4, 15, 9, 5, 0, 0, time.UTC)) {
		t.Fatalf("AssignedAt = %v, want 2024-04-15T09:05:00Z", got)
	}
	if got := sessionResp.Data[0].Guests[1].VideoSettings.IsHostEnabled; got {
		t.Fatalf("VideoSettings.IsHostEnabled = %v, want false", got)
	}

	createdSessionResp, _, err := client.GuestStar.CreateSession(context.Background(), "123")
	if err != nil {
		t.Fatalf("GuestStar.CreateSession() error = %v", err)
	}
	if got := createdSessionResp.Data[0].ID; got != "session-2" {
		t.Fatalf("Created session ID = %q, want session-2", got)
	}

	endedSessionResp, _, err := client.GuestStar.EndSession(context.Background(), helix.EndGuestStarSessionParams{
		BroadcasterID: "123",
		SessionID:     "session-2",
	})
	if err != nil {
		t.Fatalf("GuestStar.EndSession() error = %v", err)
	}
	if got := endedSessionResp.Data[0].ID; got != "session-2" {
		t.Fatalf("Ended session ID = %q, want session-2", got)
	}

	invitesResp, _, err := client.GuestStar.GetInvites(context.Background(), helix.GuestStarInvitesParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
		SessionID:     "session-1",
	})
	if err != nil {
		t.Fatalf("GuestStar.GetInvites() error = %v", err)
	}
	if got := invitesResp.Data[0].Status; got != "READY" {
		t.Fatalf("Invite status = %q, want READY", got)
	}
	if got := invitesResp.Data[0].InvitedAt.UTC(); !got.Equal(time.Date(2024, 4, 15, 9, 2, 0, 0, time.UTC)) {
		t.Fatalf("InvitedAt = %v, want 2024-04-15T09:02:00Z", got)
	}

	if _, err := client.GuestStar.SendInvite(context.Background(), helix.SendGuestStarInviteParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
		SessionID:     "session-1",
		GuestID:       "789",
	}); err != nil {
		t.Fatalf("GuestStar.SendInvite() error = %v", err)
	}

	if _, err := client.GuestStar.DeleteInvite(context.Background(), helix.DeleteGuestStarInviteParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
		SessionID:     "session-1",
		GuestID:       "789",
	}); err != nil {
		t.Fatalf("GuestStar.DeleteInvite() error = %v", err)
	}

	assignResp, _, err := client.GuestStar.AssignSlot(context.Background(), helix.AssignGuestStarSlotParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
		SessionID:     "session-1",
		GuestID:       "789",
		SlotID:        "1",
	})
	if err != nil {
		t.Fatalf("GuestStar.AssignSlot() error = %v", err)
	}
	if got := assignResp.Data.Code; got != "USER_NOT_READY" {
		t.Fatalf("AssignSlot code = %q, want USER_NOT_READY", got)
	}

	if _, err := client.GuestStar.UpdateSlot(context.Background(), helix.UpdateGuestStarSlotParams{
		BroadcasterID:     "123",
		ModeratorID:       "456",
		SessionID:         "session-1",
		SourceSlotID:      "1",
		DestinationSlotID: "2",
	}); err != nil {
		t.Fatalf("GuestStar.UpdateSlot() error = %v", err)
	}

	reinvite := true
	if _, err := client.GuestStar.DeleteSlot(context.Background(), helix.DeleteGuestStarSlotParams{
		BroadcasterID:       "123",
		ModeratorID:         "456",
		SessionID:           "session-1",
		GuestID:             "789",
		SlotID:              "1",
		ShouldReinviteGuest: &reinvite,
	}); err != nil {
		t.Fatalf("GuestStar.DeleteSlot() error = %v", err)
	}

	shareAudio := false
	shareVideo := true
	isLive := true
	volume := 65
	if _, err := client.GuestStar.UpdateSlotSettings(context.Background(), helix.UpdateGuestStarSlotSettingsParams{
		BroadcasterID:  "123",
		ModeratorID:    "456",
		SessionID:      "session-1",
		SlotID:         "1",
		IsAudioEnabled: &shareAudio,
		IsVideoEnabled: &shareVideo,
		IsLive:         &isLive,
		Volume:         &volume,
	}); err != nil {
		t.Fatalf("GuestStar.UpdateSlotSettings() error = %v", err)
	}
}
