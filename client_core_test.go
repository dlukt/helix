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

func TestCreatorCoreServicesEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2024, 4, 11, 12, 0, 0, 0, time.UTC)
	endedAt := time.Date(2024, 4, 11, 13, 0, 0, 0, time.UTC)
	isFeatured := true
	channelPatchRequests := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/channels":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query()["broadcaster_id"]; len(got) != 2 || got[0] != "123" || got[1] != "456" {
					t.Fatalf("broadcaster_id = %v, want [123 456]", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"broadcaster_id":                "123",
						"broadcaster_login":             "caster",
						"broadcaster_name":              "Caster",
						"broadcaster_language":          "en",
						"game_id":                       "509658",
						"game_name":                     "Just Chatting",
						"title":                         "Live now",
						"delay":                         0,
						"tags":                          []string{"english"},
						"content_classification_labels": []string{},
						"is_branded_content":            false,
					}},
				})
			case http.MethodPatch:
				channelPatchRequests++
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("patch broadcaster_id = %q, want 123", got)
				}
				var req map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				switch channelPatchRequests {
				case 1:
					var gameID, language, title string
					var delay int
					var tags []string
					var labels []struct {
						ID        string `json:"id"`
						IsEnabled bool   `json:"is_enabled"`
					}
					var branded bool
					if err := json.Unmarshal(req["game_id"], &gameID); err != nil || gameID != "509658" {
						t.Fatalf("game_id = %q, unmarshal err = %v, want 509658", gameID, err)
					}
					if err := json.Unmarshal(req["broadcaster_language"], &language); err != nil || language != "en" {
						t.Fatalf("broadcaster_language = %q, unmarshal err = %v, want en", language, err)
					}
					if err := json.Unmarshal(req["title"], &title); err != nil || title != "New title" {
						t.Fatalf("title = %q, unmarshal err = %v, want New title", title, err)
					}
					if err := json.Unmarshal(req["delay"], &delay); err != nil || delay != 10 {
						t.Fatalf("delay = %d, unmarshal err = %v, want 10", delay, err)
					}
					if err := json.Unmarshal(req["tags"], &tags); err != nil || len(tags) != 2 || tags[0] != "english" || tags[1] != "speedrun" {
						t.Fatalf("tags = %#v, unmarshal err = %v, want [english speedrun]", tags, err)
					}
					if err := json.Unmarshal(req["content_classification_labels"], &labels); err != nil || len(labels) != 1 || labels[0].ID != "DrugsIntoxication" || !labels[0].IsEnabled {
						t.Fatalf("content_classification_labels = %#v, unmarshal err = %v, want [{ID:DrugsIntoxication IsEnabled:true}]", labels, err)
					}
					if err := json.Unmarshal(req["is_branded_content"], &branded); err != nil || !branded {
						t.Fatalf("is_branded_content = %t, unmarshal err = %v, want true", branded, err)
					}
				case 2:
					var tags []string
					var labels []struct {
						ID        string `json:"id"`
						IsEnabled bool   `json:"is_enabled"`
					}
					if err := json.Unmarshal(req["tags"], &tags); err != nil || tags == nil || len(tags) != 0 {
						t.Fatalf("tags = %#v, unmarshal err = %v, want empty slice", tags, err)
					}
					if err := json.Unmarshal(req["content_classification_labels"], &labels); err != nil || labels == nil || len(labels) != 0 {
						t.Fatalf("content_classification_labels = %#v, unmarshal err = %v, want empty slice", labels, err)
					}
					if _, ok := req["title"]; ok {
						t.Fatalf("title should be omitted when unset")
					}
				default:
					t.Fatalf("unexpected channel patch request count %d", channelPatchRequests)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method for /channels: %s", r.Method)
			}
		case "/channels/editors":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("editors broadcaster_id = %q, want 123", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"user_id":    "456",
					"user_name":  "EditorOne",
					"user_login": "editorone",
					"created_at": "2024-04-11T12:00:00Z",
				}},
			})
		case "/channels/followed":
			if got := r.URL.Query().Get("user_id"); got != "123" {
				t.Fatalf("followed user_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "456" {
				t.Fatalf("followed broadcaster_id = %q, want 456", got)
			}
			if got := r.URL.Query().Get("first"); got != "5" {
				t.Fatalf("followed first = %q, want 5", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"total": 8,
				"data": []map[string]any{{
					"broadcaster_id":    "456",
					"broadcaster_login": "othercaster",
					"broadcaster_name":  "OtherCaster",
					"followed_at":       "2024-04-10T12:00:00Z",
				}},
				"pagination": map[string]any{"cursor": "next-followed-channels"},
			})
		case "/channels/followers":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("followers broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("user_id"); got != "789" {
				t.Fatalf("followers user_id = %q, want 789", got)
			}
			if got := r.URL.Query().Get("first"); got != "5" {
				t.Fatalf("followers first = %q, want 5", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"total": 42,
				"data": []map[string]any{{
					"followed_at": "2024-04-09T12:00:00Z",
					"user_id":     "789",
					"user_login":  "viewerthree",
					"user_name":   "ViewerThree",
				}},
				"pagination": map[string]any{"cursor": "next-channel-followers"},
			})
		case "/channels/commercial":
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("method = %q, want POST", got)
			}
			var req helix.CommercialStartRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got := req.BroadcasterID; got != "123" {
				t.Fatalf("commercial broadcaster_id = %q, want 123", got)
			}
			if got := req.Length; got != 60 {
				t.Fatalf("commercial length = %d, want 60", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"length":      60,
					"message":     "commercial started successfully",
					"retry_after": 480,
				}},
			})
		case "/channels/ads":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("ads broadcaster_id = %q, want 123", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"snooze_count":      2,
					"snooze_refresh_at": "2024-04-11T12:20:00Z",
					"next_ad_at":        "2024-04-11T12:30:00Z",
					"duration":          90,
					"last_ad_at":        "2024-04-11T12:00:00Z",
					"preroll_free_time": 1200,
				}},
			})
		case "/channels/ads/schedule/snooze":
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("method = %q, want POST", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("snooze broadcaster_id = %q, want 123", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"snooze_count":      1,
					"snooze_enabled":    true,
					"snooze_refresh_at": "2024-04-11T12:50:00Z",
					"next_ad_at":        "2024-04-11T12:45:00Z",
				}},
			})
		case "/channels/vips":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("vips broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query()["user_id"]; len(got) != 2 || got[0] != "456" || got[1] != "789" {
					t.Fatalf("vips user_id = %v, want [456 789]", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"user_id":    "456",
						"user_name":  "VIPOne",
						"user_login": "vipone",
					}},
					"pagination": map[string]any{"cursor": "next-vips"},
				})
			case http.MethodPost:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("add vip broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("user_id"); got != "789" {
					t.Fatalf("add vip user_id = %q, want 789", got)
				}
				w.WriteHeader(http.StatusNoContent)
			case http.MethodDelete:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("remove vip broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("user_id"); got != "789" {
					t.Fatalf("remove vip user_id = %q, want 789", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method for /channels/vips: %s", r.Method)
			}
		case "/streams":
			if got := r.URL.Query()["user_id"]; len(got) != 2 || got[0] != "123" || got[1] != "456" {
				t.Fatalf("user_id = %v, want [123 456]", got)
			}
			if got := r.URL.Query()["language"]; len(got) != 2 || got[0] != "en" || got[1] != "de" {
				t.Fatalf("language = %v, want [en de]", got)
			}
			if got := r.URL.Query().Get("first"); got != "20" {
				t.Fatalf("first = %q, want 20", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":            "stream-1",
					"user_id":       "123",
					"user_login":    "caster",
					"user_name":     "Caster",
					"game_id":       "509658",
					"game_name":     "Just Chatting",
					"type":          "live",
					"title":         "hello",
					"viewer_count":  42,
					"started_at":    "2024-04-11T12:00:00Z",
					"language":      "en",
					"thumbnail_url": "https://example.com/thumb.jpg",
					"is_mature":     false,
				}},
				"pagination": map[string]any{"cursor": "next-streams"},
			})
		case "/games":
			if got := r.URL.Query()["id"]; len(got) != 2 || got[0] != "1" || got[1] != "2" {
				t.Fatalf("id = %v, want [1 2]", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":          "1",
					"name":        "Game",
					"box_art_url": "https://example.com/box.jpg",
					"igdb_id":     "99",
				}},
			})
		case "/games/top":
			if got := r.URL.Query().Get("after"); got != "cursor-1" {
				t.Fatalf("after = %q, want cursor-1", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":          "509658",
					"name":        "Just Chatting",
					"box_art_url": "https://example.com/box.jpg",
				}},
				"pagination": map[string]any{"cursor": "next-games"},
			})
		case "/search/categories":
			if got := r.URL.Query().Get("query"); got != "chat" {
				t.Fatalf("query = %q, want chat", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":          "509658",
					"name":        "Just Chatting",
					"box_art_url": "https://example.com/box.jpg",
				}},
			})
		case "/search/channels":
			if got := r.URL.Query().Get("live_only"); got != "true" {
				t.Fatalf("live_only = %q, want true", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":                "123",
					"broadcaster_login": "caster",
					"display_name":      "Caster",
					"game_id":           "509658",
					"game_name":         "Just Chatting",
					"title":             "hello",
					"thumbnail_url":     "https://example.com/thumb.jpg",
					"is_live":           true,
					"started_at":        "2024-04-11T12:00:00Z",
					"tags":              []string{"english"},
				}},
			})
		case "/clips":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("started_at"); got != startedAt.Format(time.RFC3339) {
				t.Fatalf("started_at = %q, want %q", got, startedAt.Format(time.RFC3339))
			}
			if got := r.URL.Query().Get("ended_at"); got != endedAt.Format(time.RFC3339) {
				t.Fatalf("ended_at = %q, want %q", got, endedAt.Format(time.RFC3339))
			}
			if got := r.URL.Query().Get("is_featured"); got != "true" {
				t.Fatalf("is_featured = %q, want true", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":               "clip-1",
					"url":              "https://example.com/clip",
					"embed_url":        "https://example.com/embed",
					"broadcaster_id":   "123",
					"broadcaster_name": "Caster",
					"creator_id":       "999",
					"creator_name":     "Viewer",
					"video_id":         "vid-1",
					"game_id":          "509658",
					"language":         "en",
					"title":            "great clip",
					"view_count":       10,
					"created_at":       "2024-04-11T12:30:00Z",
					"thumbnail_url":    "https://example.com/thumb.jpg",
					"duration":         12.5,
					"is_featured":      true,
				}},
			})
		case "/videos":
			if got := r.URL.Query().Get("game_id"); got != "509658" {
				t.Fatalf("game_id = %q, want 509658", got)
			}
			if got := r.URL.Query().Get("sort"); got != "views" {
				t.Fatalf("sort = %q, want views", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":            "v1",
					"stream_id":     "stream-1",
					"user_id":       "123",
					"user_login":    "caster",
					"user_name":     "Caster",
					"title":         "video title",
					"description":   "video description",
					"created_at":    "2024-04-11T12:00:00Z",
					"published_at":  "2024-04-11T12:05:00Z",
					"url":           "https://example.com/video",
					"thumbnail_url": "https://example.com/thumb.jpg",
					"viewable":      "public",
					"view_count":    100,
					"language":      "en",
					"type":          "archive",
					"duration":      "2h3m",
					"muted_segments": []map[string]any{{
						"duration": 30,
						"offset":   90,
					}},
				}},
				"pagination": map[string]any{"cursor": "next-videos"},
			})
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{ClientID: "client-id", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	channelsResp, _, err := client.Channels.Get(context.Background(), helix.GetChannelsParams{
		BroadcasterIDs: []string{"123", "456"},
	})
	if err != nil {
		t.Fatalf("Channels.Get() error = %v", err)
	}
	if got := channelsResp.Data[0].BroadcasterName; got != "Caster" {
		t.Fatalf("BroadcasterName = %q, want Caster", got)
	}

	editorsResp, _, err := client.Channels.GetEditors(context.Background(), helix.GetChannelEditorsParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Channels.GetEditors() error = %v", err)
	}
	if got := editorsResp.Data[0].UserLogin; got != "editorone" {
		t.Fatalf("Editor UserLogin = %q, want editorone", got)
	}

	followedChannelsResp, followedChannelsMeta, err := client.Channels.GetFollowedChannels(context.Background(), helix.GetFollowedChannelsParams{
		CursorParams:  helix.CursorParams{First: 5},
		UserID:        "123",
		BroadcasterID: "456",
	})
	if err != nil {
		t.Fatalf("Channels.GetFollowedChannels() error = %v", err)
	}
	if got := followedChannelsResp.Total; got != 8 {
		t.Fatalf("Followed channels total = %d, want 8", got)
	}
	if got := followedChannelsResp.Data[0].BroadcasterLogin; got != "othercaster" {
		t.Fatalf("Followed broadcaster login = %q, want othercaster", got)
	}
	if got := followedChannelsMeta.Pagination.Cursor; got != "next-followed-channels" {
		t.Fatalf("Followed channels cursor = %q, want next-followed-channels", got)
	}

	followersResp, followersMeta, err := client.Channels.GetFollowers(context.Background(), helix.GetChannelFollowersParams{
		CursorParams:  helix.CursorParams{First: 5},
		BroadcasterID: "123",
		UserID:        "789",
	})
	if err != nil {
		t.Fatalf("Channels.GetFollowers() error = %v", err)
	}
	if got := followersResp.Total; got != 42 {
		t.Fatalf("Followers total = %d, want 42", got)
	}
	if got := followersResp.Data[0].UserLogin; got != "viewerthree" {
		t.Fatalf("Follower UserLogin = %q, want viewerthree", got)
	}
	if got := followersMeta.Pagination.Cursor; got != "next-channel-followers" {
		t.Fatalf("Followers cursor = %q, want next-channel-followers", got)
	}

	commercialResp, _, err := client.Channels.StartCommercial(context.Background(), helix.CommercialStartRequest{
		BroadcasterID: "123",
		Length:        60,
	})
	if err != nil {
		t.Fatalf("Channels.StartCommercial() error = %v", err)
	}
	if got := commercialResp.Data[0].RetryAfter; got != 480 {
		t.Fatalf("Commercial retry_after = %d, want 480", got)
	}
	if got := commercialResp.Data[0].Message; got != "commercial started successfully" {
		t.Fatalf("Commercial message = %q, want commercial started successfully", got)
	}

	adScheduleResp, _, err := client.Channels.GetAdSchedule(context.Background(), helix.GetAdScheduleParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Channels.GetAdSchedule() error = %v", err)
	}
	if got := adScheduleResp.Data[0].Duration; got != 90 {
		t.Fatalf("Ad schedule duration = %d, want 90", got)
	}
	if got := adScheduleResp.Data[0].NextAdAt; got != "2024-04-11T12:30:00Z" {
		t.Fatalf("Ad schedule next_ad_at = %q, want 2024-04-11T12:30:00Z", got)
	}

	snoozeResp, _, err := client.Channels.SnoozeNextAd(context.Background(), helix.SnoozeNextAdParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Channels.SnoozeNextAd() error = %v", err)
	}
	if got := snoozeResp.Data[0].SnoozeCount; got != 1 {
		t.Fatalf("Snooze count = %d, want 1", got)
	}
	if got := snoozeResp.Data[0].SnoozeEnabled; !got {
		t.Fatalf("Snooze enabled = %t, want true", got)
	}
	if got := snoozeResp.Data[0].NextAdAt; got != "2024-04-11T12:45:00Z" {
		t.Fatalf("Snooze next_ad_at = %q, want 2024-04-11T12:45:00Z", got)
	}

	vipsResp, vipsMeta, err := client.Channels.GetVIPs(context.Background(), helix.GetVIPsParams{
		BroadcasterID: "123",
		UserIDs:       []string{"456", "789"},
	})
	if err != nil {
		t.Fatalf("Channels.GetVIPs() error = %v", err)
	}
	if got := vipsResp.Data[0].UserLogin; got != "vipone" {
		t.Fatalf("VIP UserLogin = %q, want vipone", got)
	}
	if got := vipsMeta.Pagination.Cursor; got != "next-vips" {
		t.Fatalf("VIPs cursor = %q, want next-vips", got)
	}

	updateMeta, err := client.Channels.Update(context.Background(), helix.UpdateChannelParams{
		BroadcasterID: "123",
	}, helix.UpdateChannelRequest{
		GameID:                      "509658",
		BroadcasterLanguage:         "en",
		Title:                       "New title",
		Delay:                       intPtr(10),
		Tags:                        []string{"english", "speedrun"},
		ContentClassificationLabels: []helix.ContentClassificationLabelStatus{{ID: "DrugsIntoxication", IsEnabled: true}},
		IsBrandedContent:            boolPtr(true),
	})
	if err != nil {
		t.Fatalf("Channels.Update() error = %v", err)
	}
	if got := updateMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("Channels.Update() status = %d, want %d", got, http.StatusNoContent)
	}

	updateMeta, err = client.Channels.Update(context.Background(), helix.UpdateChannelParams{
		BroadcasterID: "123",
	}, helix.UpdateChannelRequest{
		Tags:                        []string{},
		ContentClassificationLabels: []helix.ContentClassificationLabelStatus{},
	})
	if err != nil {
		t.Fatalf("Channels.Update() clear values error = %v", err)
	}
	if got := updateMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("Channels.Update() clear values status = %d, want %d", got, http.StatusNoContent)
	}

	addVIPMeta, err := client.Channels.AddVIP(context.Background(), helix.AddVIPParams{
		BroadcasterID: "123",
		UserID:        "789",
	})
	if err != nil {
		t.Fatalf("Channels.AddVIP() error = %v", err)
	}
	if got := addVIPMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("Channels.AddVIP() status = %d, want %d", got, http.StatusNoContent)
	}

	removeVIPMeta, err := client.Channels.RemoveVIP(context.Background(), helix.RemoveVIPParams{
		BroadcasterID: "123",
		UserID:        "789",
	})
	if err != nil {
		t.Fatalf("Channels.RemoveVIP() error = %v", err)
	}
	if got := removeVIPMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("Channels.RemoveVIP() status = %d, want %d", got, http.StatusNoContent)
	}

	streamsResp, streamsMeta, err := client.Streams.Get(context.Background(), helix.GetStreamsParams{
		CursorParams: helix.CursorParams{First: 20},
		UserIDs:      []string{"123", "456"},
		Languages:    []string{"en", "de"},
	})
	if err != nil {
		t.Fatalf("Streams.Get() error = %v", err)
	}
	if got := streamsResp.Data[0].ViewerCount; got != 42 {
		t.Fatalf("ViewerCount = %d, want 42", got)
	}
	if got := streamsMeta.Pagination.Cursor; got != "next-streams" {
		t.Fatalf("Streams cursor = %q, want next-streams", got)
	}

	gamesResp, _, err := client.Games.Get(context.Background(), helix.GetGamesParams{
		IDs: []string{"1", "2"},
	})
	if err != nil {
		t.Fatalf("Games.Get() error = %v", err)
	}
	if got := gamesResp.Data[0].Name; got != "Game" {
		t.Fatalf("Name = %q, want Game", got)
	}

	topResp, topMeta, err := client.Games.Top(context.Background(), helix.GetTopGamesParams{
		CursorParams: helix.CursorParams{After: "cursor-1"},
	})
	if err != nil {
		t.Fatalf("Games.Top() error = %v", err)
	}
	if got := topResp.Data[0].ID; got != "509658" {
		t.Fatalf("Top game ID = %q, want 509658", got)
	}
	if got := topMeta.Pagination.Cursor; got != "next-games" {
		t.Fatalf("Top games cursor = %q, want next-games", got)
	}

	categoriesResp, _, err := client.Search.Categories(context.Background(), helix.SearchCategoriesParams{
		Query: "chat",
	})
	if err != nil {
		t.Fatalf("Search.Categories() error = %v", err)
	}
	if got := categoriesResp.Data[0].Name; got != "Just Chatting" {
		t.Fatalf("Category name = %q, want Just Chatting", got)
	}

	searchChannelsResp, _, err := client.Search.Channels(context.Background(), helix.SearchChannelsParams{
		Query:    "caster",
		LiveOnly: true,
	})
	if err != nil {
		t.Fatalf("Search.Channels() error = %v", err)
	}
	if got := searchChannelsResp.Data[0].DisplayName; got != "Caster" {
		t.Fatalf("DisplayName = %q, want Caster", got)
	}
	if got := searchChannelsResp.Data[0].StartedAt; got == nil || !got.Equal(startedAt) {
		t.Fatalf("StartedAt = %v, want %v", got, startedAt)
	}

	offlineSearchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if r.URL.Path != "/search/channels" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("query"); got != "offline" {
			t.Fatalf("query = %q, want offline", got)
		}
		if got := r.URL.Query().Get("live_only"); got != "" {
			t.Fatalf("live_only = %q, want empty", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":                "123",
					"broadcaster_login": "caster",
					"display_name":      "Caster",
					"game_id":           "509658",
					"game_name":         "Just Chatting",
					"title":             "offline now",
					"thumbnail_url":     "https://example.com/thumb.jpg",
					"is_live":           false,
					"started_at":        "",
					"tags":              []string{"english"},
				},
				{
					"id":                "456",
					"broadcaster_login": "livecaster",
					"display_name":      "LiveCaster",
					"game_id":           "509658",
					"game_name":         "Just Chatting",
					"title":             "live now",
					"thumbnail_url":     "https://example.com/thumb-live.jpg",
					"is_live":           true,
					"started_at":        "2024-04-11T12:00:00Z",
					"tags":              []string{"english"},
				},
			},
		})
	}))
	defer offlineSearchServer.Close()

	offlineSearchClient, err := helix.New(helix.Config{ClientID: "client-id", BaseURL: offlineSearchServer.URL})
	if err != nil {
		t.Fatalf("New() offline search client error = %v", err)
	}

	offlineSearchResp, _, err := offlineSearchClient.Search.Channels(context.Background(), helix.SearchChannelsParams{
		Query: "offline",
	})
	if err != nil {
		t.Fatalf("Search.Channels() offline error = %v", err)
	}
	if got := offlineSearchResp.Data[0].StartedAt; got != nil {
		t.Fatalf("offline StartedAt = %v, want nil", got)
	}
	if got := offlineSearchResp.Data[1].StartedAt; got == nil || !got.Equal(startedAt) {
		t.Fatalf("live StartedAt = %v, want %v", got, startedAt)
	}

	clipsResp, _, err := client.Clips.Get(context.Background(), helix.GetClipsParams{
		BroadcasterID: "123",
		StartedAt:     &startedAt,
		EndedAt:       &endedAt,
		IsFeatured:    &isFeatured,
	})
	if err != nil {
		t.Fatalf("Clips.Get() error = %v", err)
	}
	if got := clipsResp.Data[0].ViewCount; got != 10 {
		t.Fatalf("Clip ViewCount = %d, want 10", got)
	}

	videosResp, videosMeta, err := client.Videos.Get(context.Background(), helix.GetVideosParams{
		GameID: "509658",
		Sort:   "views",
	})
	if err != nil {
		t.Fatalf("Videos.Get() error = %v", err)
	}
	if got := videosResp.Data[0].Title; got != "video title" {
		t.Fatalf("Video Title = %q, want video title", got)
	}
	if got := videosMeta.Pagination.Cursor; got != "next-videos" {
		t.Fatalf("Videos cursor = %q, want next-videos", got)
	}
}
