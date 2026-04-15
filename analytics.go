package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// AnalyticsService provides access to Twitch analytics APIs.
type AnalyticsService struct {
	client *Client
}

// GetExtensionAnalyticsParams filters Get Extension Analytics requests.
type GetExtensionAnalyticsParams struct {
	CursorParams
	ExtensionID string
	Type        string
	StartedAt   *time.Time
	EndedAt     *time.Time
}

// GetGameAnalyticsParams filters Get Game Analytics requests.
type GetGameAnalyticsParams struct {
	CursorParams
	GameID    string
	Type      string
	StartedAt *time.Time
	EndedAt   *time.Time
}

// AnalyticsDateRange describes the period covered by an analytics report row.
type AnalyticsDateRange struct {
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

// AnalyticsReport describes a downloadable analytics report.
type AnalyticsReport struct {
	ExtensionID string             `json:"extension_id"`
	GameID      string             `json:"game_id"`
	DateRange   AnalyticsDateRange `json:"date_range"`
	Type        string             `json:"type"`
	URL         string             `json:"URL"`
}

// GetAnalyticsResponse is the typed response for analytics endpoints.
type GetAnalyticsResponse struct {
	Data       []AnalyticsReport `json:"data"`
	Pagination Pagination        `json:"pagination"`
}

// GetExtensions fetches extension analytics reports.
func (s *AnalyticsService) GetExtensions(ctx context.Context, params GetExtensionAnalyticsParams) (*GetAnalyticsResponse, *Response, error) {
	query := url.Values{}
	if params.ExtensionID != "" {
		query.Set("extension_id", params.ExtensionID)
	}
	if params.Type != "" {
		query.Set("type", params.Type)
	}
	if params.StartedAt != nil {
		query.Set("started_at", params.StartedAt.UTC().Format(time.RFC3339))
	}
	if params.EndedAt != nil {
		query.Set("ended_at", params.EndedAt.UTC().Format(time.RFC3339))
	}
	addCursorParams(query, params.CursorParams)

	var resp GetAnalyticsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/analytics/extensions",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetGames fetches game analytics reports.
func (s *AnalyticsService) GetGames(ctx context.Context, params GetGameAnalyticsParams) (*GetAnalyticsResponse, *Response, error) {
	query := url.Values{}
	if params.GameID != "" {
		query.Set("game_id", params.GameID)
	}
	if params.Type != "" {
		query.Set("type", params.Type)
	}
	if params.StartedAt != nil {
		query.Set("started_at", params.StartedAt.UTC().Format(time.RFC3339))
	}
	if params.EndedAt != nil {
		query.Set("ended_at", params.EndedAt.UTC().Format(time.RFC3339))
	}
	addCursorParams(query, params.CursorParams)

	var resp GetAnalyticsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/analytics/games",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
