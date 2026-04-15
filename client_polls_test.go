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

func TestPollsServiceEncodesRequestsAndDecodesResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/polls":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query()["id"]; len(got) != 2 || got[0] != "poll-1" || got[1] != "poll-2" {
					t.Fatalf("id = %v, want [poll-1 poll-2]", got)
				}
				if got := r.URL.Query().Get("first"); got != "5" {
					t.Fatalf("first = %q, want 5", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":                "poll-1",
						"broadcaster_id":    "123",
						"broadcaster_name":  "Caster",
						"broadcaster_login": "caster",
						"title":             "What should we play next?",
						"choices": []map[string]any{
							{
								"id":                   "choice-1",
								"title":                "Racing",
								"votes":                3,
								"channel_points_votes": 1,
								"bits_votes":           0,
							},
							{
								"id":                   "choice-2",
								"title":                "Puzzle",
								"votes":                2,
								"channel_points_votes": 0,
								"bits_votes":           0,
							},
						},
						"bits_voting_enabled":           false,
						"bits_per_vote":                 0,
						"channel_points_voting_enabled": true,
						"channel_points_per_vote":       25,
						"status":                        "ACTIVE",
						"duration":                      300,
						"started_at":                    "2024-04-15T12:00:00Z",
					}},
					"pagination": map[string]any{
						"cursor": "next-polls",
					},
				})
			case http.MethodPost:
				var req helix.CreatePollRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.BroadcasterID; got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := req.Title; got != "What should we play next?" {
					t.Fatalf("title = %q, want question", got)
				}
				if got := len(req.Choices); got != 2 {
					t.Fatalf("len(choices) = %d, want 2", got)
				}
				if req.ChannelPointsVotingEnabled == nil || !*req.ChannelPointsVotingEnabled {
					t.Fatalf("channel_points_voting_enabled = %#v, want true", req.ChannelPointsVotingEnabled)
				}
				if req.ChannelPointsPerVote == nil || *req.ChannelPointsPerVote != 25 {
					t.Fatalf("channel_points_per_vote = %#v, want 25", req.ChannelPointsPerVote)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":                "poll-3",
						"broadcaster_id":    "123",
						"broadcaster_name":  "Caster",
						"broadcaster_login": "caster",
						"title":             "What should we play next?",
						"choices": []map[string]any{
							{
								"id":                   "choice-1",
								"title":                "Racing",
								"votes":                0,
								"channel_points_votes": 0,
								"bits_votes":           0,
							},
							{
								"id":                   "choice-2",
								"title":                "Puzzle",
								"votes":                0,
								"channel_points_votes": 0,
								"bits_votes":           0,
							},
						},
						"bits_voting_enabled":           false,
						"bits_per_vote":                 0,
						"channel_points_voting_enabled": true,
						"channel_points_per_vote":       25,
						"status":                        "ACTIVE",
						"duration":                      300,
						"started_at":                    "2024-04-15T12:05:00Z",
					}},
				})
			case http.MethodPatch:
				var req helix.EndPollRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.BroadcasterID; got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := req.ID; got != "poll-3" {
					t.Fatalf("id = %q, want poll-3", got)
				}
				if got := req.Status; got != "TERMINATED" {
					t.Fatalf("status = %q, want TERMINATED", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":                "poll-3",
						"broadcaster_id":    "123",
						"broadcaster_name":  "Caster",
						"broadcaster_login": "caster",
						"title":             "What should we play next?",
						"choices": []map[string]any{
							{
								"id":                   "choice-1",
								"title":                "Racing",
								"votes":                4,
								"channel_points_votes": 1,
								"bits_votes":           0,
							},
							{
								"id":                   "choice-2",
								"title":                "Puzzle",
								"votes":                2,
								"channel_points_votes": 0,
								"bits_votes":           0,
							},
						},
						"bits_voting_enabled":           false,
						"bits_per_vote":                 0,
						"channel_points_voting_enabled": true,
						"channel_points_per_vote":       25,
						"status":                        "TERMINATED",
						"duration":                      300,
						"started_at":                    "2024-04-15T12:05:00Z",
						"ended_at":                      "2024-04-15T12:07:00Z",
					}},
				})
			default:
				t.Fatalf("unexpected method for /polls: %s", r.Method)
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

	getResp, getMeta, err := client.Polls.Get(context.Background(), helix.GetPollsParams{
		CursorParams:  helix.CursorParams{First: 5},
		BroadcasterID: "123",
		IDs:           []string{"poll-1", "poll-2"},
	})
	if err != nil {
		t.Fatalf("Polls.Get() error = %v", err)
	}
	if got := getResp.Data[0].Title; got != "What should we play next?" {
		t.Fatalf("Poll title = %q, want question", got)
	}
	if got := getResp.Data[0].Choices[0].ChannelPointsVotes; got != 1 {
		t.Fatalf("Choice channel_points_votes = %d, want 1", got)
	}
	if getResp.Data[0].EndedAt != nil {
		t.Fatalf("EndedAt = %v, want nil for active poll", getResp.Data[0].EndedAt)
	}
	if got := getMeta.Pagination.Cursor; got != "next-polls" {
		t.Fatalf("Polls cursor = %q, want next-polls", got)
	}

	createResp, _, err := client.Polls.Create(context.Background(), helix.CreatePollRequest{
		BroadcasterID:              "123",
		Title:                      "What should we play next?",
		Choices:                    []helix.CreatePollChoice{{Title: "Racing"}, {Title: "Puzzle"}},
		Duration:                   300,
		ChannelPointsVotingEnabled: boolPtr(true),
		ChannelPointsPerVote:       intPtr(25),
	})
	if err != nil {
		t.Fatalf("Polls.Create() error = %v", err)
	}
	if got := createResp.Data[0].ID; got != "poll-3" {
		t.Fatalf("Created poll ID = %q, want poll-3", got)
	}
	if got := createResp.Data[0].StartedAt; !got.Equal(time.Date(2024, 4, 15, 12, 5, 0, 0, time.UTC)) {
		t.Fatalf("Created poll StartedAt = %v, want 2024-04-15T12:05:00Z", got)
	}

	endResp, _, err := client.Polls.End(context.Background(), helix.EndPollRequest{
		BroadcasterID: "123",
		ID:            "poll-3",
		Status:        "TERMINATED",
	})
	if err != nil {
		t.Fatalf("Polls.End() error = %v", err)
	}
	if got := endResp.Data[0].Status; got != "TERMINATED" {
		t.Fatalf("Ended poll status = %q, want TERMINATED", got)
	}
	if endResp.Data[0].EndedAt == nil {
		t.Fatal("EndedAt = nil, want timestamp")
	}
}
