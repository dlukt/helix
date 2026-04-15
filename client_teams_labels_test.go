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

func TestTeamsAndContentClassificationLabelsServicesEncodeRequestsAndDecodeResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/teams/channel":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
				t.Fatalf("broadcaster_id = %q, want 123", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"broadcaster_id":       "123",
					"broadcaster_login":    "caster",
					"broadcaster_name":     "Caster",
					"background_image_url": nil,
					"banner":               nil,
					"created_at":           "2024-04-10T10:00:00Z",
					"updated_at":           "2024-04-12T10:00:00Z",
					"info":                 "A great team",
					"thumbnail_url":        "https://example.com/team.png",
					"team_name":            "teamrocket",
					"team_display_name":    "Team Rocket",
					"id":                   "team-1",
				}},
			})
		case "/teams":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("name"); got != "teamrocket" {
				t.Fatalf("name = %q, want teamrocket", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{
					"users": []map[string]any{{
						"user_id":    "123",
						"user_name":  "Caster",
						"user_login": "caster",
					}},
					"background_image_url": "https://example.com/bg.png",
					"banner":               "https://example.com/banner.png",
					"created_at":           "2024-04-10T10:00:00Z",
					"updated_at":           "2024-04-12T10:00:00Z",
					"info":                 "A great team",
					"thumbnail_url":        "https://example.com/team.png",
					"team_name":            "teamrocket",
					"team_display_name":    "Team Rocket",
					"id":                   "team-1",
				}},
			})
		case "/content_classification_labels":
			if got := r.Method; got != http.MethodGet {
				t.Fatalf("method = %q, want GET", got)
			}
			if got := r.URL.Query().Get("locale"); got != "" {
				t.Fatalf("locale = %q, want empty", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":          "DrugsIntoxication",
						"description": "Contains drug use or intoxication",
						"name":        "Drugs, Intoxication, or Excessive Tobacco Use",
					},
					{
						"id":          "ProfanityVulgarity",
						"description": "Contains profanity or vulgarity",
						"name":        "Significant Profanity or Vulgarity",
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

	channelTeamsResp, _, err := client.Teams.GetChannel(context.Background(), helix.GetChannelTeamsParams{
		BroadcasterID: "123",
	})
	if err != nil {
		t.Fatalf("Teams.GetChannel() error = %v", err)
	}
	if got := channelTeamsResp.Data[0].TeamName; got != "teamrocket" {
		t.Fatalf("Channel team_name = %q, want teamrocket", got)
	}
	if got := channelTeamsResp.Data[0].BroadcasterID; got != "123" {
		t.Fatalf("Channel broadcaster_id = %q, want 123", got)
	}
	if got := channelTeamsResp.Data[0].BroadcasterLogin; got != "caster" {
		t.Fatalf("Channel broadcaster_login = %q, want caster", got)
	}
	if channelTeamsResp.Data[0].BackgroundImage != nil {
		t.Fatalf("Channel BackgroundImage = %#v, want nil", channelTeamsResp.Data[0].BackgroundImage)
	}
	if got := channelTeamsResp.Data[0].CreatedAt.UTC(); !got.Equal(time.Date(2024, 4, 10, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("Channel CreatedAt = %v, want 2024-04-10T10:00:00Z", got)
	}

	teamsResp, _, err := client.Teams.Get(context.Background(), helix.GetTeamsParams{
		Name: "teamrocket",
	})
	if err != nil {
		t.Fatalf("Teams.Get() error = %v", err)
	}
	if got := teamsResp.Data[0].Users[0].UserLogin; got != "caster" {
		t.Fatalf("Team user_login = %q, want caster", got)
	}
	if teamsResp.Data[0].Banner == nil || *teamsResp.Data[0].Banner != "https://example.com/banner.png" {
		t.Fatalf("Team Banner = %#v, want https://example.com/banner.png", teamsResp.Data[0].Banner)
	}

	labelsResp, _, err := client.ContentClassificationLabels.Get(context.Background())
	if err != nil {
		t.Fatalf("ContentClassificationLabels.Get() error = %v", err)
	}
	if got := len(labelsResp.Data); got != 2 {
		t.Fatalf("Labels len = %d, want 2", got)
	}
	if got := labelsResp.Data[0].ID; got != "DrugsIntoxication" {
		t.Fatalf("First label ID = %q, want DrugsIntoxication", got)
	}
}

func TestContentClassificationLabelsServiceSendsLocale(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if got := r.Method; got != http.MethodGet {
			t.Fatalf("method = %q, want GET", got)
		}
		if got := r.URL.Path; got != "/content_classification_labels" {
			t.Fatalf("path = %q, want /content_classification_labels", got)
		}
		if got := r.URL.Query().Get("locale"); got != "de-DE" {
			t.Fatalf("locale = %q, want de-DE", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":          "Gambling",
					"description": "Enthaelt Gluecksspielinhalte",
					"name":        "Gluecksspiel",
				},
			},
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

	labelsResp, _, err := client.ContentClassificationLabels.Get(context.Background(), helix.GetContentClassificationLabelsParams{Locale: "de-DE"})
	if err != nil {
		t.Fatalf("ContentClassificationLabels.Get() error = %v", err)
	}
	if got := len(labelsResp.Data); got != 1 {
		t.Fatalf("Labels len = %d, want 1", got)
	}
	if got := labelsResp.Data[0].Name; got != "Gluecksspiel" {
		t.Fatalf("First label Name = %q, want Gluecksspiel", got)
	}
}

func TestContentClassificationLabelsServiceDecodesWrappedResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		if got := r.Method; got != http.MethodGet {
			t.Fatalf("method = %q, want GET", got)
		}
		if got := r.URL.Path; got != "/content_classification_labels" {
			t.Fatalf("path = %q, want /content_classification_labels", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"content_classification_labels": []map[string]any{
						{
							"id":          "MatureGame",
							"description": "Contains mature game content",
							"name":        "Mature-rated game",
						},
					},
				},
			},
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

	labelsResp, _, err := client.ContentClassificationLabels.Get(context.Background())
	if err != nil {
		t.Fatalf("ContentClassificationLabels.Get() error = %v", err)
	}
	if got := len(labelsResp.Data); got != 1 {
		t.Fatalf("Labels len = %d, want 1", got)
	}
	if got := labelsResp.Data[0].ID; got != "MatureGame" {
		t.Fatalf("First label ID = %q, want MatureGame", got)
	}
	if got := labelsResp.Data[0].Name; got != "Mature-rated game" {
		t.Fatalf("First label Name = %q, want Mature-rated game", got)
	}
}

func TestTeamsServiceGetValidatesSelectorParameters(t *testing.T) {
	t.Parallel()

	client, err := helix.New(helix.Config{ClientID: "client-id", BaseURL: "http://example.com"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tests := []struct {
		name    string
		params  helix.GetTeamsParams
		wantErr string
	}{
		{
			name:    "missing selectors",
			params:  helix.GetTeamsParams{},
			wantErr: "helix: teams get requires exactly one of id or name",
		},
		{
			name: "both selectors",
			params: helix.GetTeamsParams{
				ID:   "team-1",
				Name: "teamrocket",
			},
			wantErr: "helix: teams id and name parameters are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := client.Teams.Get(context.Background(), tt.params)
			if err == nil {
				t.Fatal("Teams.Get() error = nil, want validation error")
			}
			if got := err.Error(); got != tt.wantErr {
				t.Fatalf("Teams.Get() error = %q, want %q", got, tt.wantErr)
			}
		})
	}
}
