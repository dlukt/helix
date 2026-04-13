package helix

import (
	"context"
	"net/http"
	"net/url"
)

// ChannelsService provides access to the channels API.
type ChannelsService struct {
	client *Client
}

// GetChannelsParams filters Get Channel Information requests.
type GetChannelsParams struct {
	BroadcasterIDs []string
}

// Channel describes Twitch channel information.
type Channel struct {
	BroadcasterID               string   `json:"broadcaster_id"`
	BroadcasterLogin            string   `json:"broadcaster_login"`
	BroadcasterName             string   `json:"broadcaster_name"`
	BroadcasterLanguage         string   `json:"broadcaster_language"`
	GameID                      string   `json:"game_id"`
	GameName                    string   `json:"game_name"`
	Title                       string   `json:"title"`
	Delay                       int      `json:"delay"`
	Tags                        []string `json:"tags"`
	ContentClassificationLabels []string `json:"content_classification_labels"`
	IsBrandedContent            bool     `json:"is_branded_content"`
}

// GetChannelsResponse is the typed response for Get Channel Information.
type GetChannelsResponse struct {
	Data []Channel
}

// Get fetches channel information for one or more broadcasters.
func (s *ChannelsService) Get(ctx context.Context, params GetChannelsParams) (*GetChannelsResponse, *Response, error) {
	query := url.Values{}
	addRepeated(query, "broadcaster_id", params.BroadcasterIDs)

	var data []Channel
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/channels",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetChannelsResponse{Data: data}, meta, nil
}
