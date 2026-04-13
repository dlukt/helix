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

// AddModeratorParams identifies the broadcaster and user to add as a moderator.
type AddModeratorParams struct {
	BroadcasterID string
	UserID        string
}

// RemoveModeratorParams identifies the broadcaster and user to remove as a moderator.
type RemoveModeratorParams struct {
	BroadcasterID string
	UserID        string
}

// GetBannedUsersParams filters Get Banned Users requests.
type GetBannedUsersParams struct {
	CursorParams
	BroadcasterID string
	UserIDs       []string
}

// GetBlockedTermsParams filters Get Blocked Terms requests.
type GetBlockedTermsParams struct {
	CursorParams
	BroadcasterID string
	ModeratorID   string
}

// BannedUser describes a ban or timeout entry.
type BannedUser struct {
	UserID         string     `json:"user_id"`
	UserLogin      string     `json:"user_login"`
	UserName       string     `json:"user_name"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
	Reason         string     `json:"reason"`
	ModeratorID    string     `json:"moderator_id"`
	ModeratorLogin string     `json:"moderator_login"`
	ModeratorName  string     `json:"moderator_name"`
}

// GetBannedUsersResponse is the typed response for Get Banned Users.
type GetBannedUsersResponse struct {
	Data []BannedUser `json:"data"`
}

// BlockedTerm describes a blocked term entry.
type BlockedTerm struct {
	ID        string     `json:"id"`
	Text      string     `json:"text"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// GetBlockedTermsResponse is the typed response for Get Blocked Terms.
type GetBlockedTermsResponse struct {
	Data       []BlockedTerm `json:"data"`
	Pagination Pagination    `json:"pagination"`
}

// ShieldModeParams identifies the broadcaster and moderator for Shield Mode operations.
type ShieldModeParams struct {
	BroadcasterID string
	ModeratorID   string
}

// ShieldModeStatus describes the current Shield Mode state.
type ShieldModeStatus struct {
	IsActive        bool       `json:"is_active"`
	ModeratorID     string     `json:"moderator_id"`
	ModeratorLogin  string     `json:"moderator_login"`
	ModeratorName   string     `json:"moderator_name"`
	LastActivatedAt *time.Time `json:"last_activated_at"`
}

// GetShieldModeStatusResponse is the typed response for Get Shield Mode Status.
type GetShieldModeStatusResponse struct {
	Data []ShieldModeStatus `json:"data"`
}

// UpdateShieldModeStatusRequest is the request body for Update Shield Mode Status.
type UpdateShieldModeStatusRequest struct {
	IsActive bool `json:"is_active"`
}

// UpdateShieldModeStatusResponse is the typed response for Update Shield Mode Status.
type UpdateShieldModeStatusResponse struct {
	Data []ShieldModeStatus `json:"data"`
}

// AddBlockedTermRequest is the request body for Add Blocked Term.
type AddBlockedTermRequest struct {
	Text string `json:"text"`
}

// AddBlockedTermResponse is the typed response for Add Blocked Term.
type AddBlockedTermResponse struct {
	Data []BlockedTerm `json:"data"`
}

// RemoveBlockedTermParams identifies the blocked term to remove.
type RemoveBlockedTermParams struct {
	BroadcasterID string
	ModeratorID   string
	ID            string
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

// AddModerator adds the specified user as a moderator in the broadcaster's channel.
func (s *ModerationService) AddModerator(ctx context.Context, params AddModeratorParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("user_id", params.UserID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/moderation/moderators",
		Query:  query,
	}, nil)
}

// RemoveModerator removes the specified user as a moderator in the broadcaster's channel.
func (s *ModerationService) RemoveModerator(ctx context.Context, params RemoveModeratorParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("user_id", params.UserID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/moderation/moderators",
		Query:  query,
	}, nil)
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

// GetBlockedTerms fetches a broadcaster's public blocked terms.
func (s *ModerationService) GetBlockedTerms(ctx context.Context, params GetBlockedTermsParams) (*GetBlockedTermsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	addCursorParams(query, params.CursorParams)

	var resp GetBlockedTermsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/moderation/blocked_terms",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetShieldModeStatus fetches the broadcaster's Shield Mode status.
func (s *ModerationService) GetShieldModeStatus(ctx context.Context, params ShieldModeParams) (*GetShieldModeStatusResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	var resp GetShieldModeStatusResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/moderation/shield_mode",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// AddBlockedTerm adds a term to the broadcaster's public blocked terms list.
func (s *ModerationService) AddBlockedTerm(ctx context.Context, params ShieldModeParams, req AddBlockedTermRequest) (*AddBlockedTermResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	var resp AddBlockedTermResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/moderation/blocked_terms",
		Query:  query,
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// RemoveBlockedTerm removes a blocked term from the broadcaster's public list.
func (s *ModerationService) RemoveBlockedTerm(ctx context.Context, params RemoveBlockedTermParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("id", params.ID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/moderation/blocked_terms",
		Query:  query,
	}, nil)
}

// UpdateShieldModeStatus activates or deactivates the broadcaster's Shield Mode.
func (s *ModerationService) UpdateShieldModeStatus(ctx context.Context, params ShieldModeParams, req UpdateShieldModeStatusRequest) (*UpdateShieldModeStatusResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	var resp UpdateShieldModeStatusResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPut,
		Path:   "/moderation/shield_mode",
		Query:  query,
		Body:   req,
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
