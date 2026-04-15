package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// GoalsService provides access to Twitch creator goal APIs.
type GoalsService struct {
	client *Client
}

// GetGoalsParams identifies the broadcaster whose goals to fetch.
type GetGoalsParams struct {
	BroadcasterID string
}

// Goal describes a creator goal.
type Goal struct {
	ID               string    `json:"id"`
	BroadcasterID    string    `json:"broadcaster_id"`
	BroadcasterName  string    `json:"broadcaster_name"`
	BroadcasterLogin string    `json:"broadcaster_login"`
	Type             string    `json:"type"`
	Description      string    `json:"description"`
	CurrentAmount    int       `json:"current_amount"`
	TargetAmount     int       `json:"target_amount"`
	CreatedAt        time.Time `json:"created_at"`
}

// GetGoalsResponse is the typed response for Get Creator Goals.
type GetGoalsResponse struct {
	Data []Goal `json:"data"`
}

// Get fetches active goals for a broadcaster.
func (s *GoalsService) Get(ctx context.Context, params GetGoalsParams) (*GetGoalsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)

	var resp GetGoalsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/goals",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
