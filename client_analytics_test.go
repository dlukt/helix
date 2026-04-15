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

func TestAnalyticsServiceEncodesRequestsAndDecodesResponses(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	endedAt := time.Date(2024, 4, 30, 23, 59, 59, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/analytics/extensions":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("extension_id"); got != "ext-1" {
				t.Fatalf("extension_id = %q, want ext-1", got)
			}
			if got := r.URL.Query().Get("type"); got != "overview_v2" {
				t.Fatalf("type = %q, want overview_v2", got)
			}
			if got := r.URL.Query().Get("started_at"); got != startedAt.Format(time.RFC3339) {
				t.Fatalf("started_at = %q, want %q", got, startedAt.Format(time.RFC3339))
			}
			if got := r.URL.Query().Get("ended_at"); got != endedAt.Format(time.RFC3339) {
				t.Fatalf("ended_at = %q, want %q", got, endedAt.Format(time.RFC3339))
			}
			if got := r.URL.Query().Get("after"); got != "cursor-1" {
				t.Fatalf("after = %q, want cursor-1", got)
			}
			if got := r.URL.Query().Get("first"); got != "10" {
				t.Fatalf("first = %q, want 10", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"extension_id": "ext-1",
					"type":         "overview_v2",
					"date_range": map[string]any{
						"started_at": "2024-04-01T00:00:00Z",
						"ended_at":   "2024-04-30T23:59:59Z",
					},
					"URL": "https://example.com/extensions.csv",
				}},
				"pagination": map[string]any{
					"cursor": "next-extension-analytics",
				},
			})
		case "/analytics/games":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("game_id"); got != "509658" {
				t.Fatalf("game_id = %q, want 509658", got)
			}
			if got := r.URL.Query().Get("type"); got != "overview_v2" {
				t.Fatalf("type = %q, want overview_v2", got)
			}
			if got := r.URL.Query().Get("first"); got != "5" {
				t.Fatalf("first = %q, want 5", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"game_id": "509658",
					"type":    "overview_v2",
					"date_range": map[string]any{
						"started_at": "2024-04-01T00:00:00Z",
						"ended_at":   "2024-04-30T23:59:59Z",
					},
					"URL": "https://example.com/games.csv",
				}},
				"pagination": map[string]any{
					"cursor": "next-game-analytics",
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

	extensionsResp, extensionsMeta, err := client.Analytics.GetExtensions(context.Background(), helix.GetExtensionAnalyticsParams{
		CursorParams: helix.CursorParams{After: "cursor-1", First: 10},
		ExtensionID:  "ext-1",
		Type:         "overview_v2",
		StartedAt:    &startedAt,
		EndedAt:      &endedAt,
	})
	if err != nil {
		t.Fatalf("Analytics.GetExtensions() error = %v", err)
	}
	if got := extensionsResp.Data[0].URL; got != "https://example.com/extensions.csv" {
		t.Fatalf("Extensions URL = %q, want https://example.com/extensions.csv", got)
	}
	if got := extensionsResp.Data[0].ExtensionID; got != "ext-1" {
		t.Fatalf("Extensions extension_id = %q, want ext-1", got)
	}
	if got := extensionsMeta.Pagination.Cursor; got != "next-extension-analytics" {
		t.Fatalf("Extensions cursor = %q, want next-extension-analytics", got)
	}

	gamesResp, gamesMeta, err := client.Analytics.GetGames(context.Background(), helix.GetGameAnalyticsParams{
		CursorParams: helix.CursorParams{First: 5},
		GameID:       "509658",
		Type:         "overview_v2",
	})
	if err != nil {
		t.Fatalf("Analytics.GetGames() error = %v", err)
	}
	if got := gamesResp.Data[0].DateRange.StartedAt; !got.Equal(startedAt) {
		t.Fatalf("Games date_range.started_at = %v, want %v", got, startedAt)
	}
	if got := gamesResp.Data[0].URL; got != "https://example.com/games.csv" {
		t.Fatalf("Games URL = %q, want https://example.com/games.csv", got)
	}
	if got := gamesResp.Data[0].GameID; got != "509658" {
		t.Fatalf("Games game_id = %q, want 509658", got)
	}
	if got := gamesMeta.Pagination.Cursor; got != "next-game-analytics" {
		t.Fatalf("Games cursor = %q, want next-game-analytics", got)
	}
}
