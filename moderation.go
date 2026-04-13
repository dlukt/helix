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

// BanUserParams identifies the broadcaster and moderator for a ban action.
type BanUserParams struct {
	BroadcasterID string
	ModeratorID   string
}

// BanUserData identifies the user and optional timeout details.
type BanUserData struct {
	UserID   string `json:"user_id"`
	Duration int    `json:"duration,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// BanUserRequest is the request body for Ban User.
type BanUserRequest struct {
	Data BanUserData `json:"data"`
}

// BanAction describes the resulting ban or timeout.
type BanAction struct {
	BroadcasterID string     `json:"broadcaster_id"`
	ModeratorID   string     `json:"moderator_id"`
	UserID        string     `json:"user_id"`
	CreatedAt     time.Time  `json:"created_at"`
	EndTime       *time.Time `json:"end_time"`
}

// BanUserResponse is the typed response for Ban User.
type BanUserResponse struct {
	Data []BanAction `json:"data"`
}

// UnbanUserParams identifies the broadcaster, moderator, and user to unban.
type UnbanUserParams struct {
	BroadcasterID string
	ModeratorID   string
	UserID        string
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

// BanUser bans a user or puts them in a timeout.
func (s *ModerationService) BanUser(ctx context.Context, params BanUserParams, req BanUserRequest) (*BanUserResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	var resp BanUserResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/moderation/bans",
		Query:  query,
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// UnbanUser removes a ban or timeout from a user.
func (s *ModerationService) UnbanUser(ctx context.Context, params UnbanUserParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("user_id", params.UserID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/moderation/bans",
		Query:  query,
	}, nil)
}
