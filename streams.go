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

// GetFollowedStreamsParams filters Get Followed Streams requests.
type GetFollowedStreamsParams struct {
	CursorParams
	UserID string
}

// GetStreamKeyParams identifies the broadcaster whose stream key to fetch.
type GetStreamKeyParams struct {
	BroadcasterID string
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

// StreamKey contains a broadcaster's stream key.
type StreamKey struct {
	StreamKey string `json:"stream_key"`
}

// GetStreamKeyResponse is the typed response for Get Stream Key.
type GetStreamKeyResponse struct {
	Data []StreamKey
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

// GetFollowed fetches live streams for broadcasters that the specified user follows.
func (s *StreamsService) GetFollowed(ctx context.Context, params GetFollowedStreamsParams) (*GetStreamsResponse, *Response, error) {
	query := url.Values{}
	query.Set("user_id", params.UserID)
	addCursorParams(query, params.CursorParams)

	var data []Stream
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/streams/followed",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetStreamsResponse{Data: data}, meta, nil
}

// GetKey fetches the broadcaster's stream key.
func (s *StreamsService) GetKey(ctx context.Context, params GetStreamKeyParams) (*GetStreamKeyResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)

	var resp GetStreamKeyResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/streams/key",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
