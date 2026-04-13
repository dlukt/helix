package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// StreamsService provides access to the streams API.
type StreamsService struct {
	client *Client
}

// GetStreamsParams filters Get Streams requests.
type GetStreamsParams struct {
	CursorParams
	UserIDs    []string
	UserLogins []string
	GameIDs    []string
	Languages  []string
	Type       string
}

// Stream describes a Twitch stream.
type Stream struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	GameID       string    `json:"game_id"`
	GameName     string    `json:"game_name"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
	IsMature     bool      `json:"is_mature"`
}

// GetStreamsResponse is the typed response for Get Streams.
type GetStreamsResponse struct {
	Data []Stream
}

// Get fetches live streams.
func (s *StreamsService) Get(ctx context.Context, params GetStreamsParams) (*GetStreamsResponse, *Response, error) {
	query := url.Values{}
	addCursorParams(query, params.CursorParams)
	addRepeated(query, "user_id", params.UserIDs)
	addRepeated(query, "user_login", params.UserLogins)
	addRepeated(query, "game_id", params.GameIDs)
	addRepeated(query, "language", params.Languages)
	if params.Type != "" {
		query.Set("type", params.Type)
	}

	var data []Stream
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/streams",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetStreamsResponse{Data: data}, meta, nil
}
