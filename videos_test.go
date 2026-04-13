package helix_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/dlukt/helix"
)

func TestVideosGetRejectsInvalidParameterCombinations(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{ClientID: "client-id", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tests := []struct {
		name    string
		params  helix.GetVideosParams
		wantErr string
	}{
		{
			name: "missing selector",
			params: helix.GetVideosParams{
				Sort: "views",
			},
			wantErr: "requires exactly one of id, user_id, or game_id",
		},
		{
			name: "mixed selectors",
			params: helix.GetVideosParams{
				IDs:    []string{"v1"},
				UserID: "123",
			},
			wantErr: "mutually exclusive",
		},
		{
			name: "id selector with sort",
			params: helix.GetVideosParams{
				IDs:  []string{"v1"},
				Sort: "views",
			},
			wantErr: "period, sort, type, and first filters require user_id or game_id",
		},
		{
			name: "game selector with after cursor",
			params: helix.GetVideosParams{
				GameID:       "509658",
				CursorParams: helix.CursorParams{After: "cursor-1"},
			},
			wantErr: "after and before cursors require user_id",
		},
		{
			name: "user selector with language",
			params: helix.GetVideosParams{
				UserID:   "123",
				Language: "en",
			},
			wantErr: "language filter requires game_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, meta, err := client.Videos.Get(context.Background(), tt.params)
			if err == nil {
				t.Fatal("Videos.Get() error = nil, want validation error")
			}
			if meta != nil {
				t.Fatalf("meta = %#v, want nil", meta)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err, tt.wantErr)
			}
		})
	}

	if got := calls.Load(); got != 0 {
		t.Fatalf("unexpected HTTP calls = %d, want 0", got)
	}
}

func TestVideosGetAllowsSupportedSelectorSpecificFilters(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		call := calls.Add(1)
		if got := r.Method; got != http.MethodGet {
			t.Fatalf("method = %q, want GET", got)
		}
		if got := r.URL.Path; got != "/videos" {
			t.Fatalf("path = %q, want /videos", got)
		}

		switch call {
		case 1:
			if got := r.URL.Query().Get("game_id"); got != "509658" {
				t.Fatalf("game_id = %q, want 509658", got)
			}
			if got := r.URL.Query().Get("language"); got != "en" {
				t.Fatalf("language = %q, want en", got)
			}
		case 2:
			if got := r.URL.Query().Get("user_id"); got != "123" {
				t.Fatalf("user_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("after"); got != "cursor-1" {
				t.Fatalf("after = %q, want cursor-1", got)
			}
		default:
			t.Fatalf("unexpected HTTP call %d", call)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{},
		})
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{ClientID: "client-id", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, meta, err := client.Videos.Get(context.Background(), helix.GetVideosParams{
		GameID:   "509658",
		Language: "en",
	})
	if err != nil {
		t.Fatalf("Videos.Get() with game_id and language error = %v", err)
	}
	if meta == nil {
		t.Fatal("meta = nil, want response metadata")
	}

	_, meta, err = client.Videos.Get(context.Background(), helix.GetVideosParams{
		CursorParams: helix.CursorParams{After: "cursor-1"},
		UserID:       "123",
	})
	if err != nil {
		t.Fatalf("Videos.Get() with user_id and after error = %v", err)
	}
	if meta == nil {
		t.Fatal("meta = nil, want response metadata")
	}

	if got := calls.Load(); got != 2 {
		t.Fatalf("HTTP calls = %d, want 2", got)
	}
}

func TestVideosDeleteValidatesCountAndEncodesRequest(t *testing.T) {
	t.Parallel()

	t.Run("validation", func(t *testing.T) {
		t.Parallel()

		client, err := helix.New(helix.Config{ClientID: "client-id", BaseURL: "https://example.com"})
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		_, meta, err := client.Videos.Delete(context.Background(), helix.DeleteVideosParams{})
		if err == nil || !strings.Contains(err.Error(), "requires at least one id") {
			t.Fatalf("error = %v, want missing id validation", err)
		}
		if meta != nil {
			t.Fatalf("meta = %#v, want nil", meta)
		}

		_, meta, err = client.Videos.Delete(context.Background(), helix.DeleteVideosParams{
			IDs: []string{"1", "2", "3", "4", "5", "6"},
		})
		if err == nil || !strings.Contains(err.Error(), "at most 5 ids") {
			t.Fatalf("error = %v, want max ids validation", err)
		}
		if meta != nil {
			t.Fatalf("meta = %#v, want nil", meta)
		}
	})

	t.Run("request", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Helper()

			if got := r.Method; got != http.MethodDelete {
				t.Fatalf("method = %q, want DELETE", got)
			}
			if got := r.URL.Path; got != "/videos" {
				t.Fatalf("path = %q, want /videos", got)
			}
			if got := r.URL.Query()["id"]; len(got) != 2 || got[0] != "1234" || got[1] != "9876" {
				t.Fatalf("id = %v, want [1234 9876]", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []string{"1234", "9876"},
			})
		}))
		defer server.Close()

		client, err := helix.New(helix.Config{ClientID: "client-id", BaseURL: server.URL})
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		resp, meta, err := client.Videos.Delete(context.Background(), helix.DeleteVideosParams{
			IDs: []string{"1234", "9876"},
		})
		if err != nil {
			t.Fatalf("Videos.Delete() error = %v", err)
		}
		if got := len(resp.Data); got != 2 {
			t.Fatalf("len(resp.Data) = %d, want 2", got)
		}
		if got := resp.Data[0]; got != "1234" {
			t.Fatalf("resp.Data[0] = %q, want 1234", got)
		}
		if got := meta.StatusCode; got != http.StatusOK {
			t.Fatalf("meta.StatusCode = %d, want %d", got, http.StatusOK)
		}
	})
}
