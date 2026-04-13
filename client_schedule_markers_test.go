package helix_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dlukt/helix"
)

func TestScheduleAndMarkersServicesEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	startTime := time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/schedule":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query()["id"]; len(got) != 2 || got[0] != "seg-1" || got[1] != "seg-2" {
				t.Fatalf("id = %v, want [seg-1 seg-2]", got)
			}
			if got := r.URL.Query().Get("start_time"); got != startTime.Format(time.RFC3339) {
				t.Fatalf("start_time = %q, want %q", got, startTime.Format(time.RFC3339))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"segments": []map[string]any{
						{
							"id":             "seg-1",
							"start_time":     "2024-04-15T08:00:00Z",
							"end_time":       "2024-04-15T10:00:00Z",
							"title":          "Morning stream",
							"canceled_until": nil,
							"category": map[string]any{
								"id":   "509658",
								"name": "Just Chatting",
							},
							"is_recurring": true,
						},
					},
					"broadcaster_id":    "123",
					"broadcaster_name":  "Caster",
					"broadcaster_login": "caster",
					"vacation": map[string]any{
						"start_time": "2024-08-01T00:00:00Z",
						"end_time":   "2024-08-10T23:59:59Z",
					},
				},
				"pagination": map[string]any{"cursor": "next-schedule"},
			})
		case "/streams/markers":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("user_id"); got != "123" {
					t.Fatalf("user_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("first"); got != "5" {
					t.Fatalf("first = %q, want 5", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{
						{
							"user_id":    "123",
							"user_name":  "Caster",
							"user_login": "caster",
							"videos": []map[string]any{
								{
									"video_id": "video-1",
									"markers": []map[string]any{
										{
											"id":               "marker-1",
											"created_at":       "2024-04-15T08:30:00Z",
											"description":      "great moment",
											"position_seconds": 1800,
											"url":              "https://www.twitch.tv/videos/video-1?t=0h30m0s",
										},
									},
								},
							},
						},
					},
					"pagination": map[string]any{"cursor": "next-markers"},
				})
			case http.MethodPost:
				var req helix.CreateMarkerRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.UserID; got != "123" {
					t.Fatalf("UserID = %q, want 123", got)
				}
				if got := req.Description; got != "big play" {
					t.Fatalf("Description = %q, want big play", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{
						{
							"id":               "marker-2",
							"created_at":       "2024-04-15T08:45:00Z",
							"description":      "big play",
							"position_seconds": 2700,
						},
					},
				})
			default:
				t.Fatalf("unexpected method %s", r.Method)
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

	scheduleResp, scheduleMeta, err := client.Schedule.Get(context.Background(), helix.GetScheduleParams{
		BroadcasterID: "123",
		IDs:           []string{"seg-1", "seg-2"},
		StartTime:     &startTime,
	})
	if err != nil {
		t.Fatalf("Schedule.Get() error = %v", err)
	}
	if got := scheduleResp.Data.BroadcasterName; got != "Caster" {
		t.Fatalf("BroadcasterName = %q, want Caster", got)
	}
	if got := scheduleResp.Data.Segments[0].Category.Name; got != "Just Chatting" {
		t.Fatalf("Category.Name = %q, want Just Chatting", got)
	}
	if got := scheduleMeta.Pagination.Cursor; got != "next-schedule" {
		t.Fatalf("Schedule cursor = %q, want next-schedule", got)
	}

	markersResp, markersMeta, err := client.Markers.Get(context.Background(), helix.GetMarkersParams{
		CursorParams: helix.CursorParams{First: 5},
		UserID:       "123",
	})
	if err != nil {
		t.Fatalf("Markers.Get() error = %v", err)
	}
	if got := markersResp.Data[0].Videos[0].Markers[0].Description; got != "great moment" {
		t.Fatalf("Description = %q, want great moment", got)
	}
	if got := markersResp.Data[0].Videos[0].Markers[0].URL; got != "https://www.twitch.tv/videos/video-1?t=0h30m0s" {
		t.Fatalf("URL = %q, want marker URL", got)
	}
	if got := markersMeta.Pagination.Cursor; got != "next-markers" {
		t.Fatalf("Markers cursor = %q, want next-markers", got)
	}

	createdResp, _, err := client.Markers.Create(context.Background(), helix.CreateMarkerRequest{
		UserID:      "123",
		Description: "big play",
	})
	if err != nil {
		t.Fatalf("Markers.Create() error = %v", err)
	}
	if got := createdResp.Data[0].ID; got != "marker-2" {
		t.Fatalf("ID = %q, want marker-2", got)
	}
}

func TestMarkersGetRejectsInvalidParameterCombinations(t *testing.T) {
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
		params  helix.GetMarkersParams
		wantErr string
	}{
		{
			name:    "missing selector",
			params:  helix.GetMarkersParams{},
			wantErr: "requires exactly one of user_id or video_id",
		},
		{
			name: "mixed selectors",
			params: helix.GetMarkersParams{
				UserID:  "123",
				VideoID: "video-1",
			},
			wantErr: "mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, meta, err := client.Markers.Get(context.Background(), tt.params)
			if err == nil {
				t.Fatal("Markers.Get() error = nil, want validation error")
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
