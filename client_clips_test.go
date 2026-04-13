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

func TestClipsActionEndpointsEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	liveDuration := 30.0
	vodDuration := 30.0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/clips":
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("method = %q, want POST", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "141981764" {
				t.Fatalf("broadcaster_id = %q, want 141981764", got)
			}
			if got := r.URL.Query().Get("has_delay"); got != "" {
				t.Fatalf("has_delay = %q, want empty", got)
			}
			if got := r.URL.Query().Get("title"); got != "Live clip" {
				t.Fatalf("title = %q, want Live clip", got)
			}
			if got := r.URL.Query().Get("duration"); got != "30" {
				t.Fatalf("duration = %q, want 30", got)
			}
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":       "FiveWordsForClipSlug",
					"edit_url": "https://www.twitch.tv/twitchdev/clip/FiveWordsForClipSlug",
				}},
			})
		case "/videos/clips":
			if got := r.Method; got != http.MethodPost {
				t.Fatalf("method = %q, want POST", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "141981764" {
				t.Fatalf("broadcaster_id = %q, want 141981764", got)
			}
			if got := r.URL.Query().Get("vod_id"); got != "2277656159" {
				t.Fatalf("vod_id = %q, want 2277656159", got)
			}
			if got := r.URL.Query().Get("vod_offset"); got != "90" {
				t.Fatalf("vod_offset = %q, want 90", got)
			}
			if got := r.URL.Query().Get("duration"); got != "30" {
				t.Fatalf("duration = %q, want 30", got)
			}
			if got := r.URL.Query().Get("editor_id"); got != "12826" {
				t.Fatalf("editor_id = %q, want 12826", got)
			}
			if got := r.URL.Query().Get("title"); got != "VOD clip" {
				t.Fatalf("title = %q, want VOD clip", got)
			}
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"id":       "VODClipSlug",
					"edit_url": "https://www.twitch.tv/twitchdev/clip/VODClipSlug",
				}},
			})
		case "/clips/download":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("editor_id"); got != "141981764" {
				t.Fatalf("editor_id = %q, want 141981764", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "141981764" {
				t.Fatalf("broadcaster_id = %q, want 141981764", got)
			}
			if got := r.URL.Query()["clip_id"]; len(got) != 2 || got[0] != "clip-1" || got[1] != "clip-2" {
				t.Fatalf("clip_id = %v, want [clip-1 clip-2]", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"clip_id":                "clip-1",
						"landscape_download_url": "https://production.assets.clips.twitchcdn.net/clip1.mp4",
						"portrait_download_url":  nil,
					},
					{
						"clip_id":                "clip-2",
						"landscape_download_url": nil,
						"portrait_download_url":  "https://production.assets.clips.twitchcdn.net/clip2-portrait.mp4",
					},
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

	createResp, createMeta, err := client.Clips.Create(context.Background(), helix.CreateClipParams{
		BroadcasterID: "141981764",
		Duration:      &liveDuration,
		Title:         "Live clip",
	})
	if err != nil {
		t.Fatalf("Clips.Create() error = %v", err)
	}
	if got := createResp.Data[0].ID; got != "FiveWordsForClipSlug" {
		t.Fatalf("Create clip ID = %q, want FiveWordsForClipSlug", got)
	}
	if got := createMeta.StatusCode; got != http.StatusAccepted {
		t.Fatalf("Create clip status = %d, want %d", got, http.StatusAccepted)
	}

	createFromVODResp, createFromVODMeta, err := client.Clips.CreateFromVOD(context.Background(), helix.CreateClipFromVODParams{
		EditorID:      "12826",
		BroadcasterID: "141981764",
		VODID:         "2277656159",
		VODOffset:     90,
		Duration:      &vodDuration,
		Title:         "VOD clip",
	})
	if err != nil {
		t.Fatalf("Clips.CreateFromVOD() error = %v", err)
	}
	if got := createFromVODResp.Data[0].ID; got != "VODClipSlug" {
		t.Fatalf("Create VOD clip ID = %q, want VODClipSlug", got)
	}
	if got := createFromVODMeta.StatusCode; got != http.StatusAccepted {
		t.Fatalf("Create VOD clip status = %d, want %d", got, http.StatusAccepted)
	}

	downloadsResp, downloadsMeta, err := client.Clips.GetDownloads(context.Background(), helix.GetClipsDownloadParams{
		EditorID:      "141981764",
		BroadcasterID: "141981764",
		ClipIDs:       []string{"clip-1", "clip-2"},
	})
	if err != nil {
		t.Fatalf("Clips.GetDownloads() error = %v", err)
	}
	if got := len(downloadsResp.Data); got != 2 {
		t.Fatalf("len(downloads.Data) = %d, want 2", got)
	}
	if downloadsResp.Data[0].LandscapeDownloadURL == nil || *downloadsResp.Data[0].LandscapeDownloadURL != "https://production.assets.clips.twitchcdn.net/clip1.mp4" {
		t.Fatalf("clip 1 landscape url = %#v, want clip1 download URL", downloadsResp.Data[0].LandscapeDownloadURL)
	}
	if downloadsResp.Data[0].PortraitDownloadURL != nil {
		t.Fatalf("clip 1 portrait url = %#v, want nil", downloadsResp.Data[0].PortraitDownloadURL)
	}
	if downloadsResp.Data[1].PortraitDownloadURL == nil || *downloadsResp.Data[1].PortraitDownloadURL != "https://production.assets.clips.twitchcdn.net/clip2-portrait.mp4" {
		t.Fatalf("clip 2 portrait url = %#v, want clip2 portrait URL", downloadsResp.Data[1].PortraitDownloadURL)
	}
	if got := downloadsMeta.StatusCode; got != http.StatusOK {
		t.Fatalf("downloads status = %d, want %d", got, http.StatusOK)
	}
}

func TestClipsRejectInvalidParameterCombinations(t *testing.T) {
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

	t.Run("get validation", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			params  helix.GetClipsParams
			wantErr string
		}{
			{
				name:    "missing selector",
				params:  helix.GetClipsParams{},
				wantErr: "requires exactly one of id, broadcaster_id, or game_id",
			},
			{
				name: "mixed selectors",
				params: helix.GetClipsParams{
					BroadcasterID: "123",
					GameID:        "456",
				},
				wantErr: "mutually exclusive",
			},
			{
				name: "ids with pagination",
				params: helix.GetClipsParams{
					IDs:          []string{"clip-1"},
					CursorParams: helix.CursorParams{First: 1},
				},
				wantErr: "pagination parameters require broadcaster_id or game_id",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, meta, err := client.Clips.Get(context.Background(), tt.params)
				if err == nil {
					t.Fatal("Clips.Get() error = nil, want validation error")
				}
				if meta != nil {
					t.Fatalf("meta = %#v, want nil", meta)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want substring %q", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("create from vod validation", func(t *testing.T) {
		t.Parallel()

		duration := 30.0
		tests := []struct {
			name    string
			params  helix.CreateClipFromVODParams
			wantErr string
		}{
			{
				name: "missing editor",
				params: helix.CreateClipFromVODParams{
					BroadcasterID: "141981764",
					VODID:         "2277656159",
					VODOffset:     90,
					Duration:      &duration,
					Title:         "VOD clip",
				},
				wantErr: "requires editor_id",
			},
			{
				name: "missing title",
				params: helix.CreateClipFromVODParams{
					EditorID:      "12826",
					BroadcasterID: "141981764",
					VODID:         "2277656159",
					VODOffset:     90,
					Duration:      &duration,
				},
				wantErr: "requires title",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, meta, err := client.Clips.CreateFromVOD(context.Background(), tt.params)
				if err == nil {
					t.Fatal("Clips.CreateFromVOD() error = nil, want validation error")
				}
				if meta != nil {
					t.Fatalf("meta = %#v, want nil", meta)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want substring %q", err, tt.wantErr)
				}
			})
		}
	})

	if got := calls.Load(); got != 0 {
		t.Fatalf("unexpected HTTP calls = %d, want 0", got)
	}
}

func TestClipsCreateFromVODAllowsDefaultDuration(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		calls.Add(1)
		if got := r.URL.Path; got != "/videos/clips" {
			t.Fatalf("path = %q, want /videos/clips", got)
		}
		if got := r.Method; got != http.MethodPost {
			t.Fatalf("method = %q, want POST", got)
		}
		if got := r.URL.Query().Get("duration"); got != "" {
			t.Fatalf("duration = %q, want omitted", got)
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{
				"id":       "VODClipSlug",
				"edit_url": "https://www.twitch.tv/twitchdev/clip/VODClipSlug",
			}},
		})
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{ClientID: "client-id", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, meta, err := client.Clips.CreateFromVOD(context.Background(), helix.CreateClipFromVODParams{
		EditorID:      "12826",
		BroadcasterID: "141981764",
		VODID:         "2277656159",
		VODOffset:     90,
		Title:         "VOD clip",
	})
	if err != nil {
		t.Fatalf("Clips.CreateFromVOD() error = %v", err)
	}
	if got := resp.Data[0].ID; got != "VODClipSlug" {
		t.Fatalf("Create VOD clip ID = %q, want VODClipSlug", got)
	}
	if got := meta.StatusCode; got != http.StatusAccepted {
		t.Fatalf("Create VOD clip status = %d, want %d", got, http.StatusAccepted)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("HTTP calls = %d, want 1", got)
	}
}
