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

		switch r.URL.Path {
		case "/schedule":
			w.Header().Set("Content-Type", "application/json")
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
		case "/schedule/icalendar":
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			w.Header().Set("Content-Type", "text/calendar")
			_, _ = w.Write([]byte("BEGIN:VCALENDAR\nVERSION:2.0\nEND:VCALENDAR"))
		case "/schedule/settings":
			if got := r.Method; got != http.MethodPatch {
				t.Fatalf("method = %q, want PATCH", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			if got := r.URL.Query().Get("is_vacation_enabled"); got != "true" {
				t.Fatalf("is_vacation_enabled query = %q, want true", got)
			}
			if got := r.URL.Query().Get("vacation_start_time"); got != "2024-08-01T00:00:00Z" {
				t.Fatalf("vacation_start_time query = %q, want 2024-08-01T00:00:00Z", got)
			}
			if got := r.URL.Query().Get("vacation_end_time"); got != "2024-08-10T23:59:59Z" {
				t.Fatalf("vacation_end_time query = %q, want 2024-08-10T23:59:59Z", got)
			}
			if got := r.URL.Query().Get("timezone"); got != "America/New_York" {
				t.Fatalf("timezone query = %q, want America/New_York", got)
			}
			if got := r.Header.Get("Content-Type"); got != "" {
				t.Fatalf("content-type = %q, want omitted", got)
			}
			if got := r.ContentLength; got != 0 {
				t.Fatalf("content-length = %d, want 0", got)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/schedule/segment":
			w.Header().Set("Content-Type", "application/json")
			switch r.Method {
			case http.MethodPost:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				var req struct {
					StartTime   time.Time `json:"start_time"`
					Timezone    string    `json:"timezone"`
					IsRecurring bool      `json:"is_recurring"`
					Duration    string    `json:"duration"`
					CategoryID  string    `json:"category_id,omitempty"`
					Title       string    `json:"title,omitempty"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.StartTime.Format(time.RFC3339); got != "2024-04-16T08:00:00Z" {
					t.Fatalf("StartTime = %q, want 2024-04-16T08:00:00Z", got)
				}
				if got := req.Timezone; got != "America/New_York" {
					t.Fatalf("Timezone = %q, want America/New_York", got)
				}
				if !req.IsRecurring {
					t.Fatal("IsRecurring = false, want true")
				}
				if got := req.Duration; got != "60" {
					t.Fatalf("Duration = %q, want 60", got)
				}
				if got := req.CategoryID; got != "509658" {
					t.Fatalf("CategoryID = %q, want 509658", got)
				}
				if got := req.Title; got != "Recurring morning stream" {
					t.Fatalf("Title = %q, want Recurring morning stream", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"segments": []map[string]any{
							{
								"id":             "seg-3",
								"start_time":     "2024-04-16T08:00:00Z",
								"end_time":       "2024-04-16T09:00:00Z",
								"title":          "Recurring morning stream",
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
						"vacation":          nil,
					},
				})
			case http.MethodPatch:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("id"); got != "seg-3" {
					t.Fatalf("id = %q, want seg-3", got)
				}
				var req struct {
					StartTime  *time.Time `json:"start_time,omitempty"`
					Duration   *string    `json:"duration,omitempty"`
					CategoryID string     `json:"category_id,omitempty"`
					Title      string     `json:"title,omitempty"`
					IsCanceled *bool      `json:"is_canceled,omitempty"`
					Timezone   string     `json:"timezone,omitempty"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if req.Duration == nil || *req.Duration != "120" {
					t.Fatalf("Duration = %#v, want 120", req.Duration)
				}
				if got := req.Title; got != "Longer morning stream" {
					t.Fatalf("Title = %q, want Longer morning stream", got)
				}
				if got := req.CategoryID; got != "509660" {
					t.Fatalf("CategoryID = %q, want 509660", got)
				}
				if req.IsCanceled == nil || *req.IsCanceled {
					t.Fatalf("IsCanceled = %#v, want false", req.IsCanceled)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"segments": []map[string]any{
							{
								"id":             "seg-3",
								"start_time":     "2024-04-16T08:00:00Z",
								"end_time":       "2024-04-16T10:00:00Z",
								"title":          "Longer morning stream",
								"canceled_until": nil,
								"category": map[string]any{
									"id":   "509660",
									"name": "Art",
								},
								"is_recurring": true,
							},
						},
						"broadcaster_id":    "123",
						"broadcaster_name":  "Caster",
						"broadcaster_login": "caster",
						"vacation":          nil,
					},
				})
			case http.MethodDelete:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query().Get("id"); got != "seg-3" {
					t.Fatalf("id = %q, want seg-3", got)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				t.Fatalf("unexpected method %s", r.Method)
			}
		case "/streams/markers":
			w.Header().Set("Content-Type", "application/json")
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

	calendarResp, calendarMeta, err := client.Schedule.GetICalendar(context.Background(), helix.GetScheduleICalendarParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Schedule.GetICalendar() error = %v", err)
	}
	if got := calendarResp.Calendar; got != "BEGIN:VCALENDAR\nVERSION:2.0\nEND:VCALENDAR" {
		t.Fatalf("Calendar = %q, want iCalendar body", got)
	}
	if got := calendarMeta.StatusCode; got != http.StatusOK {
		t.Fatalf("Schedule.GetICalendar() status = %d, want %d", got, http.StatusOK)
	}

	vacationStart := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	vacationEnd := time.Date(2024, 8, 10, 23, 59, 59, 0, time.UTC)
	isVacationEnabled := true
	updateSettingsMeta, err := client.Schedule.UpdateSettings(context.Background(), helix.UpdateScheduleParams{
		BroadcasterID:     "123",
		IsVacationEnabled: &isVacationEnabled,
		VacationStartTime: &vacationStart,
		VacationEndTime:   &vacationEnd,
		Timezone:          "America/New_York",
	})
	if err != nil {
		t.Fatalf("Schedule.UpdateSettings() error = %v", err)
	}
	if got := updateSettingsMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("Schedule.UpdateSettings() status = %d, want %d", got, http.StatusNoContent)
	}

	segmentStart := time.Date(2024, 4, 16, 8, 0, 0, 0, time.UTC)
	createdSegmentResp, _, err := client.Schedule.CreateSegment(context.Background(), helix.CreateScheduleSegmentParams{
		BroadcasterID: "123",
	}, helix.CreateScheduleSegmentRequest{
		StartTime:   segmentStart,
		Timezone:    "America/New_York",
		IsRecurring: true,
		Duration:    60,
		CategoryID:  "509658",
		Title:       "Recurring morning stream",
	})
	if err != nil {
		t.Fatalf("Schedule.CreateSegment() error = %v", err)
	}
	if got := createdSegmentResp.Data.Segments[0].ID; got != "seg-3" {
		t.Fatalf("Created segment ID = %q, want seg-3", got)
	}

	duration := 120
	isCanceled := false
	updatedSegmentResp, _, err := client.Schedule.UpdateSegment(context.Background(), helix.UpdateScheduleSegmentParams{
		BroadcasterID: "123",
		ID:            "seg-3",
	}, helix.UpdateScheduleSegmentRequest{
		Duration:   &duration,
		CategoryID: "509660",
		Title:      "Longer morning stream",
		IsCanceled: &isCanceled,
	})
	if err != nil {
		t.Fatalf("Schedule.UpdateSegment() error = %v", err)
	}
	if got := updatedSegmentResp.Data.Segments[0].Title; got != "Longer morning stream" {
		t.Fatalf("Updated segment Title = %q, want Longer morning stream", got)
	}
	if got := updatedSegmentResp.Data.Segments[0].Category.Name; got != "Art" {
		t.Fatalf("Updated segment Category.Name = %q, want Art", got)
	}

	deleteSegmentMeta, err := client.Schedule.DeleteSegment(context.Background(), helix.DeleteScheduleSegmentParams{
		BroadcasterID: "123",
		ID:            "seg-3",
	})
	if err != nil {
		t.Fatalf("Schedule.DeleteSegment() error = %v", err)
	}
	if got := deleteSegmentMeta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("Schedule.DeleteSegment() status = %d, want %d", got, http.StatusNoContent)
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
