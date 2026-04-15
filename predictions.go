package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// PredictionsService provides access to Twitch prediction APIs.
type PredictionsService struct {
	client *Client
}

// GetPredictionsParams filters Get Predictions requests.
type GetPredictionsParams struct {
	CursorParams
	BroadcasterID string
	IDs           []string
}

// PredictionPredictor describes a top predictor for an outcome.
type PredictionPredictor struct {
	UserID            string `json:"user_id"`
	UserName          string `json:"user_name"`
	UserLogin         string `json:"user_login"`
	ChannelPointsUsed int    `json:"channel_points_used"`
	ChannelPointsWon  int    `json:"channel_points_won"`
}

// PredictionOutcome describes one prediction outcome.
type PredictionOutcome struct {
	ID            string                `json:"id"`
	Title         string                `json:"title"`
	Users         int                   `json:"users"`
	ChannelPoints int                   `json:"channel_points"`
	TopPredictors []PredictionPredictor `json:"top_predictors"`
	Color         string                `json:"color"`
}

// Prediction describes a Twitch prediction.
type Prediction struct {
	ID               string              `json:"id"`
	BroadcasterID    string              `json:"broadcaster_id"`
	BroadcasterName  string              `json:"broadcaster_name"`
	BroadcasterLogin string              `json:"broadcaster_login"`
	Title            string              `json:"title"`
	WinningOutcomeID string              `json:"winning_outcome_id"`
	Outcomes         []PredictionOutcome `json:"outcomes"`
	PredictionWindow int                 `json:"prediction_window"`
	Status           string              `json:"status"`
	CreatedAt        time.Time           `json:"created_at"`
	EndedAt          *time.Time          `json:"ended_at"`
	LockedAt         *time.Time          `json:"locked_at"`
}

// GetPredictionsResponse is the typed response for Get Predictions.
type GetPredictionsResponse struct {
	Data       []Prediction `json:"data"`
	Pagination Pagination   `json:"pagination"`
}

// CreatePredictionOutcome identifies an outcome when creating a prediction.
type CreatePredictionOutcome struct {
	Title string `json:"title"`
}

// CreatePredictionRequest is the request body for Create Prediction.
type CreatePredictionRequest struct {
	BroadcasterID    string                    `json:"broadcaster_id"`
	Title            string                    `json:"title"`
	Outcomes         []CreatePredictionOutcome `json:"outcomes"`
	PredictionWindow int                       `json:"prediction_window"`
}

// CreatePredictionResponse is the typed response for Create Prediction.
type CreatePredictionResponse struct {
	Data []Prediction `json:"data"`
}

// EndPredictionRequest is the request body for End Prediction.
type EndPredictionRequest struct {
	BroadcasterID    string `json:"broadcaster_id"`
	ID               string `json:"id"`
	Status           string `json:"status"`
	WinningOutcomeID string `json:"winning_outcome_id,omitempty"`
}

// EndPredictionResponse is the typed response for End Prediction.
type EndPredictionResponse struct {
	Data []Prediction `json:"data"`
}

// Get fetches one or more predictions for a broadcaster.
func (s *PredictionsService) Get(ctx context.Context, params GetPredictionsParams) (*GetPredictionsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	addRepeated(query, "id", params.IDs)
	addCursorParams(query, params.CursorParams)

	var resp GetPredictionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/predictions",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// Create starts a new prediction for a broadcaster.
func (s *PredictionsService) Create(ctx context.Context, req CreatePredictionRequest) (*CreatePredictionResponse, *Response, error) {
	var resp CreatePredictionResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/predictions",
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// End resolves, locks, or cancels an active prediction.
func (s *PredictionsService) End(ctx context.Context, req EndPredictionRequest) (*EndPredictionResponse, *Response, error) {
	var resp EndPredictionResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/predictions",
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
