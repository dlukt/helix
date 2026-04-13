package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// ModerationService provides access to Twitch moderation APIs.
type ModerationService struct {
	client *Client
}

// GetModeratorsParams filters Get Moderators requests.
type GetModeratorsParams struct {
	CursorParams
	BroadcasterID string
	UserIDs       []string
}

// Moderator describes a moderator in a broadcaster's channel.
type Moderator struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// GetModeratorsResponse is the typed response for Get Moderators.
type GetModeratorsResponse struct {
	Data []Moderator `json:"data"`
}

// GetBannedUsersParams filters Get Banned Users requests.
type GetBannedUsersParams struct {
	CursorParams
	BroadcasterID string
	UserIDs       []string
}

// BannedUser describes a ban or timeout entry.
type BannedUser struct {
	UserID         string    `json:"user_id"`
	UserLogin      string    `json:"user_login"`
	UserName       string    `json:"user_name"`
	ExpiresAt      string    `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	Reason         string    `json:"reason"`
	ModeratorID    string    `json:"moderator_id"`
	ModeratorLogin string    `json:"moderator_login"`
	ModeratorName  string    `json:"moderator_name"`
}

// GetBannedUsersResponse is the typed response for Get Banned Users.
type GetBannedUsersResponse struct {
	Data []BannedUser `json:"data"`
}

// GetModerators fetches moderators for a broadcaster.
func (s *ModerationService) GetModerators(ctx context.Context, params GetModeratorsParams) (*GetModeratorsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	addCursorParams(query, params.CursorParams)
	addRepeated(query, "user_id", params.UserIDs)

	var resp GetModeratorsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/moderation/moderators",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetBannedUsers fetches banned or timed-out users for a broadcaster.
func (s *ModerationService) GetBannedUsers(ctx context.Context, params GetBannedUsersParams) (*GetBannedUsersResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	addCursorParams(query, params.CursorParams)
	addRepeated(query, "user_id", params.UserIDs)

	var resp GetBannedUsersResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/moderation/banned",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
