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

func TestPredictionsServiceEncodesRequestsAndDecodesResponses(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/predictions":
			switch r.Method {
			case http.MethodGet:
				if got := r.URL.Query().Get("broadcaster_id"); got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := r.URL.Query()["id"]; len(got) != 2 || got[0] != "prediction-1" || got[1] != "prediction-2" {
					t.Fatalf("id = %v, want [prediction-1 prediction-2]", got)
				}
				if got := r.URL.Query().Get("first"); got != "5" {
					t.Fatalf("first = %q, want 5", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":                 "prediction-1",
						"broadcaster_id":     "123",
						"broadcaster_name":   "Caster",
						"broadcaster_login":  "caster",
						"title":              "Will we beat the boss?",
						"winning_outcome_id": "",
						"outcomes": []map[string]any{
							{
								"id":             "outcome-1",
								"title":          "Yes",
								"users":          10,
								"channel_points": 2500,
								"top_predictors": []map[string]any{
									{
										"user_id":             "456",
										"user_name":           "ViewerOne",
										"user_login":          "viewerone",
										"channel_points_used": 500,
										"channel_points_won":  0,
									},
								},
								"color": "BLUE",
							},
							{
								"id":             "outcome-2",
								"title":          "No",
								"users":          4,
								"channel_points": 1000,
								"top_predictors": []map[string]any{},
								"color":          "PINK",
							},
						},
						"prediction_window": 120,
						"status":            "ACTIVE",
						"created_at":        "2024-04-15T12:10:00Z",
						"ended_at":          nil,
						"locked_at":         nil,
					}},
					"pagination": map[string]any{
						"cursor": "next-predictions",
					},
				})
			case http.MethodPost:
				var req helix.CreatePredictionRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.BroadcasterID; got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := req.Title; got != "Will we beat the boss?" {
					t.Fatalf("title = %q, want question", got)
				}
				if got := len(req.Outcomes); got != 2 {
					t.Fatalf("len(outcomes) = %d, want 2", got)
				}
				if got := req.PredictionWindow; got != 120 {
					t.Fatalf("prediction_window = %d, want 120", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":                 "prediction-3",
						"broadcaster_id":     "123",
						"broadcaster_name":   "Caster",
						"broadcaster_login":  "caster",
						"title":              "Will we beat the boss?",
						"winning_outcome_id": "",
						"outcomes": []map[string]any{
							{
								"id":             "outcome-1",
								"title":          "Yes",
								"users":          0,
								"channel_points": 0,
								"top_predictors": []map[string]any{},
								"color":          "BLUE",
							},
							{
								"id":             "outcome-2",
								"title":          "No",
								"users":          0,
								"channel_points": 0,
								"top_predictors": []map[string]any{},
								"color":          "PINK",
							},
						},
						"prediction_window": 120,
						"status":            "ACTIVE",
						"created_at":        "2024-04-15T12:15:00Z",
						"ended_at":          nil,
						"locked_at":         nil,
					}},
				})
			case http.MethodPatch:
				var req helix.EndPredictionRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("Decode() error = %v", err)
				}
				if got := req.BroadcasterID; got != "123" {
					t.Fatalf("broadcaster_id = %q, want 123", got)
				}
				if got := req.ID; got != "prediction-3" {
					t.Fatalf("id = %q, want prediction-3", got)
				}
				if got := req.Status; got != "RESOLVED" {
					t.Fatalf("status = %q, want RESOLVED", got)
				}
				if got := req.WinningOutcomeID; got != "outcome-1" {
					t.Fatalf("winning_outcome_id = %q, want outcome-1", got)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"data": []map[string]any{{
						"id":                 "prediction-3",
						"broadcaster_id":     "123",
						"broadcaster_name":   "Caster",
						"broadcaster_login":  "caster",
						"title":              "Will we beat the boss?",
						"winning_outcome_id": "outcome-1",
						"outcomes": []map[string]any{
							{
								"id":             "outcome-1",
								"title":          "Yes",
								"users":          20,
								"channel_points": 5000,
								"top_predictors": []map[string]any{
									{
										"user_id":             "456",
										"user_name":           "ViewerOne",
										"user_login":          "viewerone",
										"channel_points_used": 500,
										"channel_points_won":  900,
									},
								},
								"color": "BLUE",
							},
							{
								"id":             "outcome-2",
								"title":          "No",
								"users":          6,
								"channel_points": 1400,
								"top_predictors": []map[string]any{},
								"color":          "PINK",
							},
						},
						"prediction_window": 120,
						"status":            "RESOLVED",
						"created_at":        "2024-04-15T12:15:00Z",
						"ended_at":          "2024-04-15T12:18:00Z",
						"locked_at":         "2024-04-15T12:17:00Z",
					}},
				})
			default:
				t.Fatalf("unexpected method for /predictions: %s", r.Method)
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

	getResp, getMeta, err := client.Predictions.Get(context.Background(), helix.GetPredictionsParams{
		CursorParams:  helix.CursorParams{First: 5},
		BroadcasterID: "123",
		IDs:           []string{"prediction-1", "prediction-2"},
	})
	if err != nil {
		t.Fatalf("Predictions.Get() error = %v", err)
	}
	if got := getResp.Data[0].Outcomes[0].TopPredictors[0].UserLogin; got != "viewerone" {
		t.Fatalf("Top predictor user_login = %q, want viewerone", got)
	}
	if getResp.Data[0].EndedAt != nil {
		t.Fatalf("EndedAt = %v, want nil for active prediction", getResp.Data[0].EndedAt)
	}
	if got := getMeta.Pagination.Cursor; got != "next-predictions" {
		t.Fatalf("Predictions cursor = %q, want next-predictions", got)
	}

	createResp, _, err := client.Predictions.Create(context.Background(), helix.CreatePredictionRequest{
		BroadcasterID:    "123",
		Title:            "Will we beat the boss?",
		Outcomes:         []helix.CreatePredictionOutcome{{Title: "Yes"}, {Title: "No"}},
		PredictionWindow: 120,
	})
	if err != nil {
		t.Fatalf("Predictions.Create() error = %v", err)
	}
	if got := createResp.Data[0].ID; got != "prediction-3" {
		t.Fatalf("Created prediction ID = %q, want prediction-3", got)
	}
	if got := createResp.Data[0].Status; got != "ACTIVE" {
		t.Fatalf("Created prediction status = %q, want ACTIVE", got)
	}

	endResp, _, err := client.Predictions.End(context.Background(), helix.EndPredictionRequest{
		BroadcasterID:    "123",
		ID:               "prediction-3",
		Status:           "RESOLVED",
		WinningOutcomeID: "outcome-1",
	})
	if err != nil {
		t.Fatalf("Predictions.End() error = %v", err)
	}
	if got := endResp.Data[0].WinningOutcomeID; got != "outcome-1" {
		t.Fatalf("WinningOutcomeID = %q, want outcome-1", got)
	}
	if endResp.Data[0].EndedAt == nil {
		t.Fatal("EndedAt = nil, want timestamp")
	}
	if endResp.Data[0].LockedAt == nil {
		t.Fatal("LockedAt = nil, want timestamp")
	}
	if got := endResp.Data[0].LockedAt.UTC(); !got.Equal(time.Date(2024, 4, 15, 12, 17, 0, 0, time.UTC)) {
		t.Fatalf("LockedAt = %v, want 2024-04-15T12:17:00Z", got)
	}
}
