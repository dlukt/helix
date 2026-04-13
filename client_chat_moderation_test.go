package helix_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dlukt/helix"
)

func intPtr(v int) *int {
	return &v
}

func TestChatAndModerationServicesEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/chat/settings":
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
						"broadcaster_id":                    "123",
						"moderator_id":                      "456",
						"emote_mode":                        false,
						"slow_mode":                         true,
						"slow_mode_wait_time":               10,
						"follower_mode":                     true,
						"follower_mode_duration":            5,
						"subscriber_mode":                   false,
						"unique_chat_mode":                  true,
						"non_moderator_chat_delay":          true,
						"non_moderator_chat_delay_duration": 4,
					}},
				})
			case http.MethodPatch:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("patch broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("moderator_id"); got != "456" {
					t.Fatalf("patch moderator_id = %q, want 456", got)
				}
				var req helix.UpdateChatSettingsRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if req.SlowMode == nil || !*req.SlowMode {
					t.Fatalf("SlowMode = %#v, want true", req.SlowMode)
				}
				if req.SlowModeWaitTime == nil || *req.SlowModeWaitTime != 30 {
					t.Fatalf("SlowModeWaitTime = %#v, want 30", req.SlowModeWaitTime)
				}
				if req.UniqueChatMode == nil || !*req.UniqueChatMode {
					t.Fatalf("UniqueChatMode = %#v, want true", req.UniqueChatMode)
				}
				if req.FollowerMode != nil {
					t.Fatalf("FollowerMode = %#v, want omitted", req.FollowerMode)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"broadcaster_id":                    "123",
						"moderator_id":                      "456",
						"emote_mode":                        false,
						"slow_mode":                         true,
						"slow_mode_wait_time":               30,
						"follower_mode":                     true,
						"follower_mode_duration":            5,
						"subscriber_mode":                   false,
						"unique_chat_mode":                  true,
						"non_moderator_chat_delay":          true,
						"non_moderator_chat_delay_duration": 4,
					}},
				})
			default:
				t.Fatalf("unexpected method for /chat/settings: %s", r.Method)
			}
		case "/chat/chatters":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("moderator_id"); got != "456" {
				t.Fatalf("moderator_id = %q, want 456", got)
			}
			if got := r.URL.Query().Get("first"); got != "25" {
				t.Fatalf("first = %q, want 25", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"user_id":    "u-1",
					"user_login": "viewer1",
					"user_name":  "Viewer1",
				}},
				"total": 1,
				"pagination": map[string]any{
					"cursor": "next-chatters",
				},
			})
		case "/chat/announcements":
			if r.Method != http.MethodPost {
				t.Fatalf("announcement method = %s, want POST", r.Method)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("announcement broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("moderator_id"); got != "456" {
				t.Fatalf("announcement moderator_id = %q, want 456", got)
			}
			var req helix.SendAnnouncementRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got := req.Message; got != "Heads up!" {
				t.Fatalf("announcement message = %q, want Heads up!", got)
			}
			if got := req.Color; got != "purple" {
				t.Fatalf("announcement color = %q, want purple", got)
			}
			if req.ForSourceOnly == nil || *req.ForSourceOnly {
				t.Fatalf("announcement ForSourceOnly = %#v, want false", req.ForSourceOnly)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/chat/shoutouts":
			if r.Method != http.MethodPost {
				t.Fatalf("shoutout method = %s, want POST", r.Method)
			}
			if got := r.URL.Query().Get("from_broadcaster_id"); got != "123" {
				t.Fatalf("from_broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("to_broadcaster_id"); got != "789" {
				t.Fatalf("to_broadcaster_id = %q, want 789", got)
			}
			if got := r.URL.Query().Get("moderator_id"); got != "456" {
				t.Fatalf("shoutout moderator_id = %q, want 456", got)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/chat/messages":
			if r.Method != http.MethodPost {
				t.Fatalf("message method = %s, want POST", r.Method)
			}
			var req helix.SendMessageRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got := req.BroadcasterID; got != "123" {
				t.Fatalf("message broadcaster_id = %q, want 123", got)
			}
			if got := req.SenderID; got != "456" {
				t.Fatalf("message sender_id = %q, want 456", got)
			}
			if got := req.Message; got != "Hello, world!" {
				t.Fatalf("message text = %q, want Hello, world!", got)
			}
			if got := req.ReplyParentMessageID; got != "parent-1" {
				t.Fatalf("message reply_parent_message_id = %q, want parent-1", got)
			}
			if req.ForSourceOnly == nil || *req.ForSourceOnly {
				t.Fatalf("message ForSourceOnly = %#v, want false", req.ForSourceOnly)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"message_id":  "msg-1",
					"is_sent":     true,
					"drop_reason": nil,
				}},
			})
		case "/moderation/moderators":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query()["user_id"]; len(got) != 2 || got[0] != "456" || got[1] != "789" {
				t.Fatalf("user_id = %v, want [456 789]", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"user_id":    "456",
					"user_login": "mod1",
					"user_name":  "Mod1",
				}},
				"pagination": map[string]any{
					"cursor": "next-moderators",
				},
			})
		case "/moderation/banned":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query()["user_id"]; len(got) != 2 || got[0] != "777" || got[1] != "888" {
				t.Fatalf("user_id = %v, want [777 888]", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"user_id":         "777",
					"user_login":      "badviewer",
					"user_name":       "BadViewer",
					"expires_at":      "2024-04-15T02:00:28Z",
					"created_at":      "2024-04-15T01:30:28Z",
					"reason":          "spam",
					"moderator_id":    "456",
					"moderator_login": "mod1",
					"moderator_name":  "Mod1",
				}},
				"pagination": map[string]any{
					"cursor": "next-banned",
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

	settingsResp, _, err := client.Chat.GetSettings(context.Background(), helix.GetChatSettingsParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	})
	if err != nil {
		t.Fatalf("Chat.GetSettings() error = %v", err)
	}
	if got := settingsResp.Data[0].SlowMode; !got {
		t.Fatal("SlowMode = false, want true")
	}
	if settingsResp.Data[0].SlowModeWaitTime == nil || *settingsResp.Data[0].SlowModeWaitTime != 10 {
		t.Fatalf("SlowModeWaitTime = %v, want 10", settingsResp.Data[0].SlowModeWaitTime)
	}
	if settingsResp.Data[0].NonModeratorChatDelay == nil || !*settingsResp.Data[0].NonModeratorChatDelay {
		t.Fatalf("NonModeratorChatDelay = %v, want true", settingsResp.Data[0].NonModeratorChatDelay)
	}
	if settingsResp.Data[0].NonModeratorChatDelayDuration == nil || *settingsResp.Data[0].NonModeratorChatDelayDuration != 4 {
		t.Fatalf("NonModeratorChatDelayDuration = %v, want 4", settingsResp.Data[0].NonModeratorChatDelayDuration)
	}

	updatedResp, _, err := client.Chat.UpdateSettings(context.Background(), helix.UpdateChatSettingsParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	}, helix.UpdateChatSettingsRequest{
		SlowMode:         boolPtr(true),
		SlowModeWaitTime: intPtr(30),
		UniqueChatMode:   boolPtr(true),
	})
	if err != nil {
		t.Fatalf("Chat.UpdateSettings() error = %v", err)
	}
	if updatedResp.Data[0].SlowModeWaitTime == nil || *updatedResp.Data[0].SlowModeWaitTime != 30 {
		t.Fatalf("updated SlowModeWaitTime = %v, want 30", updatedResp.Data[0].SlowModeWaitTime)
	}

	chattersResp, chattersMeta, err := client.Chat.GetChatters(context.Background(), helix.GetChattersParams{
		CursorParams:  helix.CursorParams{First: 25},
		BroadcasterID: "123",
		ModeratorID:   "456",
	})
	if err != nil {
		t.Fatalf("Chat.GetChatters() error = %v", err)
	}
	if got := chattersResp.Total; got != 1 {
		t.Fatalf("Total = %d, want 1", got)
	}
	if got := chattersMeta.Pagination.Cursor; got != "next-chatters" {
		t.Fatalf("Chatters cursor = %q, want next-chatters", got)
	}

	announcementMeta, err := client.Chat.SendAnnouncement(context.Background(), helix.SendAnnouncementParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	}, helix.SendAnnouncementRequest{
		Message:       "Heads up!",
		Color:         "purple",
		ForSourceOnly: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("Chat.SendAnnouncement() error = %v", err)
	}
	if announcementMeta.StatusCode != http.StatusNoContent {
		t.Fatalf("announcement status = %d, want %d", announcementMeta.StatusCode, http.StatusNoContent)
	}

	shoutoutMeta, err := client.Chat.SendShoutout(context.Background(), helix.SendShoutoutParams{
		FromBroadcasterID: "123",
		ToBroadcasterID:   "789",
		ModeratorID:       "456",
	})
	if err != nil {
		t.Fatalf("Chat.SendShoutout() error = %v", err)
	}
	if shoutoutMeta.StatusCode != http.StatusNoContent {
		t.Fatalf("shoutout status = %d, want %d", shoutoutMeta.StatusCode, http.StatusNoContent)
	}

	messageResp, _, err := client.Chat.SendMessage(context.Background(), helix.SendMessageRequest{
		BroadcasterID:        "123",
		SenderID:             "456",
		Message:              "Hello, world!",
		ReplyParentMessageID: "parent-1",
		ForSourceOnly:        boolPtr(false),
	})
	if err != nil {
		t.Fatalf("Chat.SendMessage() error = %v", err)
	}
	if got := messageResp.Data[0].MessageID; got != "msg-1" {
		t.Fatalf("message id = %q, want msg-1", got)
	}
	if !messageResp.Data[0].IsSent {
		t.Fatal("IsSent = false, want true")
	}
	if messageResp.Data[0].DropReason != nil {
		t.Fatalf("DropReason = %#v, want nil", messageResp.Data[0].DropReason)
	}

	modsResp, modsMeta, err := client.Moderation.GetModerators(context.Background(), helix.GetModeratorsParams{
		BroadcasterID: "123",
		UserIDs:       []string{"456", "789"},
	})
	if err != nil {
		t.Fatalf("Moderation.GetModerators() error = %v", err)
	}
	if got := modsResp.Data[0].UserLogin; got != "mod1" {
		t.Fatalf("UserLogin = %q, want mod1", got)
	}
	if got := modsMeta.Pagination.Cursor; got != "next-moderators" {
		t.Fatalf("Moderators cursor = %q, want next-moderators", got)
	}

	bannedResp, bannedMeta, err := client.Moderation.GetBannedUsers(context.Background(), helix.GetBannedUsersParams{
		BroadcasterID: "123",
		UserIDs:       []string{"777", "888"},
	})
	if err != nil {
		t.Fatalf("Moderation.GetBannedUsers() error = %v", err)
	}
	if got := bannedResp.Data[0].Reason; got != "spam" {
		t.Fatalf("Reason = %q, want spam", got)
	}
	if got := bannedMeta.Pagination.Cursor; got != "next-banned" {
		t.Fatalf("Banned cursor = %q, want next-banned", got)
	}
}

func TestChatGetSettingsPreservesOmittedPrivilegedDelayFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/chat/settings" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
			t.Fatalf("broadcaster_id = %q, want 123", got)
		}
		if got := r.URL.Query().Get("moderator_id"); got != "" {
			t.Fatalf("moderator_id = %q, want empty", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{
				"broadcaster_id":         "123",
				"moderator_id":           "",
				"emote_mode":             false,
				"slow_mode":              false,
				"slow_mode_wait_time":    nil,
				"follower_mode":          false,
				"follower_mode_duration": nil,
				"subscriber_mode":        false,
				"unique_chat_mode":       false,
			}},
		})
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, _, err := client.Chat.GetSettings(context.Background(), helix.GetChatSettingsParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Chat.GetSettings() error = %v", err)
	}

	got := resp.Data[0]
	if got.NonModeratorChatDelay != nil {
		t.Fatalf("NonModeratorChatDelay = %v, want nil when field is omitted", *got.NonModeratorChatDelay)
	}
	if got.NonModeratorChatDelayDuration != nil {
		t.Fatalf("NonModeratorChatDelayDuration = %v, want nil when field is omitted", *got.NonModeratorChatDelayDuration)
	}
}

func TestChatGetSettingsPreservesDisabledPrivilegedDelayFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/chat/settings" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("moderator_id"); got != "456" {
			t.Fatalf("moderator_id = %q, want 456", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{
				"broadcaster_id":                    "123",
				"moderator_id":                      "456",
				"emote_mode":                        false,
				"slow_mode":                         false,
				"slow_mode_wait_time":               nil,
				"follower_mode":                     false,
				"follower_mode_duration":            nil,
				"subscriber_mode":                   false,
				"unique_chat_mode":                  false,
				"non_moderator_chat_delay":          false,
				"non_moderator_chat_delay_duration": nil,
			}},
		})
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, _, err := client.Chat.GetSettings(context.Background(), helix.GetChatSettingsParams{
		BroadcasterID: "123",
		ModeratorID:   "456",
	})
	if err != nil {
		t.Fatalf("Chat.GetSettings() error = %v", err)
	}

	got := resp.Data[0]
	if got.NonModeratorChatDelay == nil {
		t.Fatal("NonModeratorChatDelay = nil, want false")
	}
	if *got.NonModeratorChatDelay {
		t.Fatal("NonModeratorChatDelay = true, want false")
	}
	if got.NonModeratorChatDelayDuration != nil {
		t.Fatalf("NonModeratorChatDelayDuration = %v, want nil", *got.NonModeratorChatDelayDuration)
	}
}
