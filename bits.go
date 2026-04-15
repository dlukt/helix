package helix

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// BitsService provides access to Twitch bits APIs.
type BitsService struct {
	client *Client
}

// GetBitsLeaderboardParams filters Get Bits Leaderboard requests.
type GetBitsLeaderboardParams struct {
	Count     int
	Period    string
	StartedAt *time.Time
	UserID    string
}

// GetCheermotesParams filters Get Cheermotes requests.
type GetCheermotesParams struct {
	BroadcasterID string
}

// BitsDateRange describes the date range used for a leaderboard response.
type BitsDateRange struct {
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
}

// BitsLeaderboardEntry describes one user in the bits leaderboard.
type BitsLeaderboardEntry struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Rank      int    `json:"rank"`
	Score     int    `json:"score"`
}

// GetBitsLeaderboardResponse is the typed response for Get Bits Leaderboard.
type GetBitsLeaderboardResponse struct {
	Data      []BitsLeaderboardEntry `json:"data"`
	DateRange BitsDateRange          `json:"date_range"`
	Total     int                    `json:"total"`
}

// CheermoteImages describes the nested cheermote asset map by theme, animation, and scale.
type CheermoteImages map[string]map[string]map[string]string

// CheermoteTier describes one tier of a cheermote.
type CheermoteTier struct {
	MinBits        int             `json:"min_bits"`
	ID             string          `json:"id"`
	Color          string          `json:"color"`
	Images         CheermoteImages `json:"images"`
	CanCheer       bool            `json:"can_cheer"`
	ShowInBitsCard bool            `json:"show_in_bits_card"`
}

// Cheermote describes one cheermote definition.
type Cheermote struct {
	Prefix       string          `json:"prefix"`
	Tiers        []CheermoteTier `json:"tiers"`
	Type         string          `json:"type"`
	Order        int             `json:"order"`
	LastUpdated  time.Time       `json:"last_updated"`
	IsCharitable bool            `json:"is_charitable"`
}

// GetCheermotesResponse is the typed response for Get Cheermotes.
type GetCheermotesResponse struct {
	Data []Cheermote `json:"data"`
}

// GetLeaderboard fetches the current bits leaderboard.
func (s *BitsService) GetLeaderboard(ctx context.Context, params GetBitsLeaderboardParams) (*GetBitsLeaderboardResponse, *Response, error) {
	query := url.Values{}
	if params.Count > 0 {
		query.Set("count", strconv.Itoa(params.Count))
	}
	if params.Period != "" {
		query.Set("period", params.Period)
	}
	if params.StartedAt != nil {
		query.Set("started_at", params.StartedAt.UTC().Format(time.RFC3339))
	}
	if params.UserID != "" {
		query.Set("user_id", params.UserID)
	}

	var resp GetBitsLeaderboardResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/bits/leaderboard",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetCheermotes fetches cheermote definitions, optionally scoped to a broadcaster.
func (s *BitsService) GetCheermotes(ctx context.Context, params GetCheermotesParams) (*GetCheermotesResponse, *Response, error) {
	query := url.Values{}
	if params.BroadcasterID != "" {
		query.Set("broadcaster_id", params.BroadcasterID)
	}

	var resp GetCheermotesResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/bits/cheermotes",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
