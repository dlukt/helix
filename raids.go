package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// RaidsService provides access to raid APIs.
type RaidsService struct {
	client *Client
}

// StartRaidParams identifies the broadcaster and target of a raid.
type StartRaidParams struct {
	FromBroadcasterID string
	ToBroadcasterID   string
}

// Raid contains information about a scheduled raid.
type Raid struct {
	CreatedAt time.Time `json:"created_at"`
	IsMature  bool      `json:"is_mature"`
}

// StartRaidResponse is the typed response for Start a Raid.
type StartRaidResponse struct {
	Data []Raid `json:"data"`
}

// CancelRaidParams identifies the broadcaster whose pending raid should be canceled.
type CancelRaidParams struct {
	BroadcasterID string
}

// Start schedules a raid from one broadcaster to another.
func (s *RaidsService) Start(ctx context.Context, params StartRaidParams) (*StartRaidResponse, *Response, error) {
	query := url.Values{}
	query.Set("from_broadcaster_id", params.FromBroadcasterID)
	query.Set("to_broadcaster_id", params.ToBroadcasterID)

	var resp StartRaidResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/raids",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// Cancel cancels a pending raid.
func (s *RaidsService) Cancel(ctx context.Context, params CancelRaidParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/raids",
		Query:  query,
	}, nil)
}
