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

func TestChatAssetServicesEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/chat/emotes":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":           "emote-1",
					"name":         "PartyParrot",
					"images":       map[string]any{"url_1x": "https://example.com/e1-1x", "url_2x": "https://example.com/e1-2x", "url_4x": "https://example.com/e1-4x"},
					"tier":         "1000",
					"emote_type":   "subscriptions",
					"emote_set_id": "set-1",
					"owner_id":     "123",
					"format":       []string{"static", "animated"},
					"scale":        []string{"1.0", "2.0", "3.0"},
					"theme_mode":   []string{"light", "dark"},
				}},
				"template": "https://static-cdn.jtvnw.net/emoticons/v2/{id}/{format}/{theme_mode}/{scale}",
			})
		case "/chat/emotes/global":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":           "global-1",
					"name":         "HeyGuys",
					"images":       map[string]any{"url_1x": "https://example.com/g1-1x", "url_2x": "https://example.com/g1-2x", "url_4x": "https://example.com/g1-4x"},
					"emote_type":   "globals",
					"emote_set_id": "set-global",
					"owner_id":     "0",
					"format":       []string{"static"},
					"scale":        []string{"1.0", "2.0", "3.0"},
					"theme_mode":   []string{"light", "dark"},
				}},
				"template": "https://static-cdn.jtvnw.net/emoticons/v2/{id}/{format}/{theme_mode}/{scale}",
			})
		case "/chat/emotes/set":
			if got := r.URL.Query()["emote_set_id"]; len(got) != 2 || got[0] != "set-1" || got[1] != "set-2" {
				t.Fatalf("emote_set_id = %v, want [set-1 set-2]", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":           "set-emote-1",
					"name":         "SetSmile",
					"images":       map[string]any{"url_1x": "https://example.com/s1-1x", "url_2x": "https://example.com/s1-2x", "url_4x": "https://example.com/s1-4x"},
					"emote_type":   "follower",
					"emote_set_id": "set-2",
					"owner_id":     "456",
					"format":       []string{"static"},
					"scale":        []string{"1.0", "2.0", "3.0"},
					"theme_mode":   []string{"light"},
				}},
				"template": "https://static-cdn.jtvnw.net/emoticons/v2/{id}/{format}/{theme_mode}/{scale}",
			})
		case "/chat/emotes/user":
			if got := r.URL.Query().Get("user_id"); got != "456" {
				t.Fatalf("user_id = %q, want 456", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("after"); got != "cursor-1" {
				t.Fatalf("after = %q, want cursor-1", got)
			}
			if got := r.URL.Query().Get("first"); got != "20" {
				t.Fatalf("first = %q, want 20", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":           "user-emote-1",
					"name":         "FollowerLove",
					"images":       map[string]any{"url_1x": "https://example.com/u1-1x", "url_2x": "https://example.com/u1-2x", "url_4x": "https://example.com/u1-4x"},
					"tier":         "",
					"emote_type":   "follower",
					"emote_set_id": "set-user",
					"owner_id":     "123",
					"format":       []string{"static"},
					"scale":        []string{"1.0", "2.0", "3.0"},
					"theme_mode":   []string{"light", "dark"},
				}},
				"template": "https://static-cdn.jtvnw.net/emoticons/v2/{id}/{format}/{theme_mode}/{scale}",
				"pagination": map[string]any{
					"cursor": "next-user-emotes",
				},
			})
		case "/chat/badges":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"set_id": "subscriber",
					"versions": []map[string]any{{
						"id":           "12",
						"image_url_1x": "https://example.com/b1-1x",
						"image_url_2x": "https://example.com/b1-2x",
						"image_url_4x": "https://example.com/b1-4x",
						"title":        "1-Year Subscriber",
						"description":  "Channel subscriber for 1 year",
						"click_action": "subscribe_to_channel",
						"click_url":    "https://www.twitch.tv/subs/caster",
					}},
				}},
			})
		case "/chat/badges/global":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"set_id": "staff",
					"versions": []map[string]any{{
						"id":           "1",
						"image_url_1x": "https://example.com/staff-1x",
						"image_url_2x": "https://example.com/staff-2x",
						"image_url_4x": "https://example.com/staff-4x",
						"title":        "Staff",
						"description":  "Twitch staff member",
						"click_action": nil,
						"click_url":    nil,
					}},
				}},
			})
		case "/shared_chat/session":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"session_id":          "session-1",
					"host_broadcaster_id": "123",
					"participants": []map[string]any{
						{"broadcaster_id": "123"},
						{"broadcaster_id": "456"},
					},
					"created_at": "2024-04-15T10:00:00Z",
					"updated_at": "2024-04-15T10:05:00Z",
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

	channelEmotesResp, _, err := client.Chat.GetChannelEmotes(context.Background(), helix.GetChannelEmotesParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Chat.GetChannelEmotes() error = %v", err)
	}
	if got := channelEmotesResp.Data[0].Name; got != "PartyParrot" {
		t.Fatalf("Channel emote name = %q, want PartyParrot", got)
	}
	if got := channelEmotesResp.Data[0].OwnerID; got != "123" {
		t.Fatalf("Channel emote owner_id = %q, want 123", got)
	}
	if got := channelEmotesResp.Template; got == "" {
		t.Fatal("Channel emote template = empty, want value")
	}

	globalEmotesResp, _, err := client.Chat.GetGlobalEmotes(context.Background())
	if err != nil {
		t.Fatalf("Chat.GetGlobalEmotes() error = %v", err)
	}
	if got := globalEmotesResp.Data[0].EmoteType; got != "globals" {
		t.Fatalf("Global emote type = %q, want globals", got)
	}

	emoteSetsResp, _, err := client.Chat.GetEmoteSets(context.Background(), helix.GetEmoteSetsParams{
		EmoteSetIDs: []string{"set-1", "set-2"},
	})
	if err != nil {
		t.Fatalf("Chat.GetEmoteSets() error = %v", err)
	}
	if got := emoteSetsResp.Data[0].EmoteSetID; got != "set-2" {
		t.Fatalf("Emote set id = %q, want set-2", got)
	}

	userEmotesResp, userEmotesMeta, err := client.Chat.GetUserEmotes(context.Background(), helix.GetUserEmotesParams{
		CursorParams:  helix.CursorParams{After: "cursor-1", First: 20},
		UserID:        "456",
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Chat.GetUserEmotes() error = %v", err)
	}
	if got := userEmotesResp.Data[0].Name; got != "FollowerLove" {
		t.Fatalf("User emote name = %q, want FollowerLove", got)
	}
	if got := userEmotesMeta.Pagination.Cursor; got != "next-user-emotes" {
		t.Fatalf("User emotes cursor = %q, want next-user-emotes", got)
	}

	channelBadgesResp, _, err := client.Chat.GetChannelBadges(context.Background(), helix.GetChannelBadgesParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Chat.GetChannelBadges() error = %v", err)
	}
	if got := channelBadgesResp.Data[0].Versions[0].Title; got != "1-Year Subscriber" {
		t.Fatalf("Channel badge title = %q, want 1-Year Subscriber", got)
	}
	if got := channelBadgesResp.Data[0].Versions[0].ClickAction; got == nil || *got != "subscribe_to_channel" {
		t.Fatalf("Channel badge click_action = %#v, want subscribe_to_channel", got)
	}

	globalBadgesResp, _, err := client.Chat.GetGlobalBadges(context.Background())
	if err != nil {
		t.Fatalf("Chat.GetGlobalBadges() error = %v", err)
	}
	if got := globalBadgesResp.Data[0].SetID; got != "staff" {
		t.Fatalf("Global badge set_id = %q, want staff", got)
	}
	if got := globalBadgesResp.Data[0].Versions[0].ClickURL; got != nil {
		t.Fatalf("Global badge click_url = %#v, want nil", got)
	}

	sharedSessionResp, _, err := client.Chat.GetSharedChatSession(context.Background(), helix.GetSharedChatSessionParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Chat.GetSharedChatSession() error = %v", err)
	}
	if got := sharedSessionResp.Data[0].SessionID; got != "session-1" {
		t.Fatalf("Shared chat session id = %q, want session-1", got)
	}
	if got := len(sharedSessionResp.Data[0].Participants); got != 2 {
		t.Fatalf("Shared chat participant count = %d, want 2", got)
	}
	if got := sharedSessionResp.Data[0].CreatedAt; !got.Equal(time.Date(2024, 4, 15, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("Shared chat created_at = %v, want 2024-04-15T10:00:00Z", got)
	}
}
