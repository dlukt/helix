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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/channels":
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
