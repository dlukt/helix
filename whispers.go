package helix

import (
	"context"
	"net/http"
	"net/url"
)

// WhispersService provides access to Twitch whispers APIs.
type WhispersService struct {
	client *Client
}

// SendWhisperParams identifies the source and destination users for a whisper.
type SendWhisperParams struct {
	FromUserID string
	ToUserID   string
}

// SendWhisperRequest is the request body for Send Whisper.
type SendWhisperRequest struct {
	Message string `json:"message"`
}

// Send sends a whisper from one user to another.
func (s *WhispersService) Send(ctx context.Context, params SendWhisperParams, req SendWhisperRequest) (*Response, error) {
	query := url.Values{}
	query.Set("from_user_id", params.FromUserID)
	query.Set("to_user_id", params.ToUserID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/whispers",
		Query:  query,
		Body:   req,
	}, nil)
}
