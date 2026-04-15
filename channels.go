package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

// ChannelsService provides access to the channels API.
type ChannelsService struct {
	client *Client
}

// GetChannelsParams filters Get Channel Information requests.
type GetChannelsParams struct {
	BroadcasterIDs []string
}

// GetChannelEditorsParams identifies the broadcaster whose editors to list.
type GetChannelEditorsParams struct {
	BroadcasterID string
}

// GetFollowedChannelsParams filters Get Followed Channels requests.
type GetFollowedChannelsParams struct {
	CursorParams
	UserID        string
	BroadcasterID string
}

// GetChannelFollowersParams filters Get Channel Followers requests.
type GetChannelFollowersParams struct {
	CursorParams
	BroadcasterID string
	UserID        string
}

// GetVIPsParams filters Get VIPs requests.
type GetVIPsParams struct {
	CursorParams
	BroadcasterID string
	UserIDs       []string
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

// ChannelEditor describes a user with the editor role for a channel.
type ChannelEditor struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	UserLogin string `json:"user_login"`
	CreatedAt string `json:"created_at"`
}

// GetChannelEditorsResponse is the typed response for Get Channel Editors.
type GetChannelEditorsResponse struct {
	Data []ChannelEditor `json:"data"`
}

// FollowedChannel describes a broadcaster followed by a user.
type FollowedChannel struct {
	BroadcasterID    string    `json:"broadcaster_id"`
	BroadcasterLogin string    `json:"broadcaster_login"`
	BroadcasterName  string    `json:"broadcaster_name"`
	FollowedAt       time.Time `json:"followed_at"`
}

// GetFollowedChannelsResponse is the typed response for Get Followed Channels.
type GetFollowedChannelsResponse struct {
	Data       []FollowedChannel `json:"data"`
	Pagination Pagination        `json:"pagination"`
	Total      int               `json:"total"`
}

// ChannelFollower describes a user following a broadcaster.
type ChannelFollower struct {
	FollowedAt time.Time `json:"followed_at"`
	UserID     string    `json:"user_id"`
	UserLogin  string    `json:"user_login"`
	UserName   string    `json:"user_name"`
}

// GetChannelFollowersResponse is the typed response for Get Channel Followers.
type GetChannelFollowersResponse struct {
	Data       []ChannelFollower `json:"data"`
	Pagination Pagination        `json:"pagination"`
	Total      int               `json:"total"`
}

// VIP describes a VIP in a broadcaster's channel.
type VIP struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	UserLogin string `json:"user_login"`
}

// GetVIPsResponse is the typed response for Get VIPs.
type GetVIPsResponse struct {
	Data       []VIP      `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// UpdateChannelParams identifies which broadcaster channel to update.
type UpdateChannelParams struct {
	BroadcasterID string
}

// UpdateChannelRequest contains mutable channel information fields.
type UpdateChannelRequest struct {
	GameID                      string                             `json:"game_id,omitempty"`
	BroadcasterLanguage         string                             `json:"broadcaster_language,omitempty"`
	Title                       string                             `json:"title,omitempty"`
	Delay                       *int                               `json:"delay,omitempty"`
	Tags                        []string                           `json:"tags,omitempty"`
	ContentClassificationLabels []ContentClassificationLabelStatus `json:"content_classification_labels,omitempty"`
	IsBrandedContent            *bool                              `json:"is_branded_content,omitempty"`
}

// ContentClassificationLabelStatus describes whether a channel content classification label is enabled.
type ContentClassificationLabelStatus struct {
	ID        string `json:"id"`
	IsEnabled bool   `json:"is_enabled"`
}

// MarshalJSON preserves explicitly empty slices so callers can clear tags and labels.
func (r UpdateChannelRequest) MarshalJSON() ([]byte, error) {
	body := struct {
		GameID                      string                              `json:"game_id,omitempty"`
		BroadcasterLanguage         string                              `json:"broadcaster_language,omitempty"`
		Title                       string                              `json:"title,omitempty"`
		Delay                       *int                                `json:"delay,omitempty"`
		Tags                        *[]string                           `json:"tags,omitempty"`
		ContentClassificationLabels *[]ContentClassificationLabelStatus `json:"content_classification_labels,omitempty"`
		IsBrandedContent            *bool                               `json:"is_branded_content,omitempty"`
	}{
		GameID:              r.GameID,
		BroadcasterLanguage: r.BroadcasterLanguage,
		Title:               r.Title,
		Delay:               r.Delay,
		IsBrandedContent:    r.IsBrandedContent,
	}
	if r.Tags != nil {
		body.Tags = &r.Tags
	}
	if r.ContentClassificationLabels != nil {
		body.ContentClassificationLabels = &r.ContentClassificationLabels
	}
	return json.Marshal(body)
}

// AddVIPParams identifies the broadcaster and user to add as a VIP.
type AddVIPParams struct {
	BroadcasterID string
	UserID        string
}

// RemoveVIPParams identifies the broadcaster and user to remove as a VIP.
type RemoveVIPParams struct {
	BroadcasterID string
	UserID        string
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

// GetEditors fetches users who have the editor role for a broadcaster.
func (s *ChannelsService) GetEditors(ctx context.Context, params GetChannelEditorsParams) (*GetChannelEditorsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)

	var resp GetChannelEditorsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/channels/editors",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetFollowedChannels fetches broadcasters followed by the specified user.
func (s *ChannelsService) GetFollowedChannels(ctx context.Context, params GetFollowedChannelsParams) (*GetFollowedChannelsResponse, *Response, error) {
	query := url.Values{}
	query.Set("user_id", params.UserID)
	if params.BroadcasterID != "" {
		query.Set("broadcaster_id", params.BroadcasterID)
	}
	addCursorParams(query, params.CursorParams)

	var resp GetFollowedChannelsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/channels/followed",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetFollowers fetches users following the specified broadcaster.
func (s *ChannelsService) GetFollowers(ctx context.Context, params GetChannelFollowersParams) (*GetChannelFollowersResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	if params.UserID != "" {
		query.Set("user_id", params.UserID)
	}
	addCursorParams(query, params.CursorParams)

	var resp GetChannelFollowersResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/channels/followers",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetVIPs fetches VIPs for a broadcaster.
func (s *ChannelsService) GetVIPs(ctx context.Context, params GetVIPsParams) (*GetVIPsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	addCursorParams(query, params.CursorParams)
	addRepeated(query, "user_id", params.UserIDs)

	var resp GetVIPsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/channels/vips",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// Update modifies mutable channel information for a broadcaster.
func (s *ChannelsService) Update(ctx context.Context, params UpdateChannelParams, req UpdateChannelRequest) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/channels",
		Query:  query,
		Body:   req,
	}, nil)
}

// AddVIP adds the specified user as a VIP in the broadcaster's channel.
func (s *ChannelsService) AddVIP(ctx context.Context, params AddVIPParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("user_id", params.UserID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/channels/vips",
		Query:  query,
	}, nil)
}

// RemoveVIP removes the specified user as a VIP in the broadcaster's channel.
func (s *ChannelsService) RemoveVIP(ctx context.Context, params RemoveVIPParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("user_id", params.UserID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/channels/vips",
		Query:  query,
	}, nil)
}
