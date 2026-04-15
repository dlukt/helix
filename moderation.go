package helix

import (
	"context"
	"fmt"
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

// GetModeratedChannelsParams filters Get Moderated Channels requests.
type GetModeratedChannelsParams struct {
	CursorParams
	UserID string
}

// Moderator describes a moderator in a broadcaster's channel.
type Moderator struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// ModeratedChannel describes a channel that the user can moderate.
type ModeratedChannel struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
}

// GetModeratedChannelsResponse is the typed response for Get Moderated Channels.
type GetModeratedChannelsResponse struct {
	Data       []ModeratedChannel `json:"data"`
	Pagination Pagination         `json:"pagination"`
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

// WarnUserParams identifies the broadcaster and moderator for a warning action.
type WarnUserParams struct {
	BroadcasterID string
	ModeratorID   string
}

// CheckAutoModStatusParams identifies the broadcaster whose AutoMod rules should be evaluated.
type CheckAutoModStatusParams struct {
	BroadcasterID string
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

// AutoModSettingsParams identifies the broadcaster and moderator for AutoMod settings operations.
type AutoModSettingsParams struct {
	BroadcasterID string
	ModeratorID   string
}

// GetUnbanRequestsParams filters Get Unban Requests requests.
type GetUnbanRequestsParams struct {
	CursorParams
	BroadcasterID string
	ModeratorID   string
	Status        string
	UserID        string
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

// AutoModCheckMessage describes a message to evaluate against AutoMod rules.
type AutoModCheckMessage struct {
	MsgID   string `json:"msg_id"`
	UserID  string `json:"user_id"`
	MsgText string `json:"msg_text"`
}

// CheckAutoModStatusRequest is the request body for Check AutoMod Status.
type CheckAutoModStatusRequest struct {
	Data []AutoModCheckMessage `json:"data"`
}

// AutoModCheckResult reports whether a message would be allowed through AutoMod.
type AutoModCheckResult struct {
	MsgID       string `json:"msg_id"`
	IsPermitted bool   `json:"is_permitted"`
}

// CheckAutoModStatusResponse is the typed response for Check AutoMod Status.
type CheckAutoModStatusResponse struct {
	Data []AutoModCheckResult `json:"data"`
}

// WarningPayload contains the user and reason for a warning.
type WarningPayload struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

// WarnUserRequest is the request body for Warn Chat User.
type WarnUserRequest struct {
	Data WarningPayload `json:"data"`
}

// Warning describes the resulting warning state for a user.
type Warning struct {
	BroadcasterID string `json:"broadcaster_id"`
	UserID        string `json:"user_id"`
	ModeratorID   string `json:"moderator_id"`
	Reason        string `json:"reason"`
}

// WarnUserResponse is the typed response for Warn Chat User.
type WarnUserResponse struct {
	Data []Warning `json:"data"`
}

// ManageHeldAutoModMessageRequest is the request body for Manage Held AutoMod Messages.
type ManageHeldAutoModMessageRequest struct {
	UserID string `json:"user_id"`
	MsgID  string `json:"msg_id"`
	Action string `json:"action"`
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

// SuspiciousUserStatusChangeRequest is the request body for Add Suspicious Status to Chat User.
type SuspiciousUserStatusChangeRequest struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

// SuspiciousUserStatusChange describes the result of applying or removing a suspicious-user status.
type SuspiciousUserStatusChange struct {
	UserID        string    `json:"user_id"`
	BroadcasterID string    `json:"broadcaster_id"`
	ModeratorID   string    `json:"moderator_id"`
	UpdatedAt     time.Time `json:"updated_at"`
	Status        string    `json:"status"`
	Types         []string  `json:"types"`
}

// SuspiciousUserStatusChangeResponse is the typed response for suspicious-user status changes.
type SuspiciousUserStatusChangeResponse struct {
	Data []SuspiciousUserStatusChange `json:"data"`
}

// AutoModSettings describes a broadcaster's AutoMod configuration.
type AutoModSettings struct {
	BroadcasterID           string `json:"broadcaster_id"`
	ModeratorID             string `json:"moderator_id"`
	OverallLevel            *int   `json:"overall_level"`
	Disability              int    `json:"disability"`
	Aggression              int    `json:"aggression"`
	SexualitySexOrGender    int    `json:"sexuality_sex_or_gender"`
	Misogyny                int    `json:"misogyny"`
	Bullying                int    `json:"bullying"`
	Swearing                int    `json:"swearing"`
	RaceEthnicityOrReligion int    `json:"race_ethnicity_or_religion"`
	SexBasedTerms           int    `json:"sex_based_terms"`
}

// GetAutoModSettingsResponse is the typed response for Get AutoMod Settings.
type GetAutoModSettingsResponse struct {
	Data []AutoModSettings `json:"data"`
}

// UpdateAutoModSettingsRequest is the request body for Update AutoMod Settings.
type UpdateAutoModSettingsRequest struct {
	Aggression              *int `json:"aggression,omitempty"`
	Bullying                *int `json:"bullying,omitempty"`
	Disability              *int `json:"disability,omitempty"`
	Misogyny                *int `json:"misogyny,omitempty"`
	OverallLevel            *int `json:"overall_level,omitempty"`
	RaceEthnicityOrReligion *int `json:"race_ethnicity_or_religion,omitempty"`
	SexBasedTerms           *int `json:"sex_based_terms,omitempty"`
	SexualitySexOrGender    *int `json:"sexuality_sex_or_gender,omitempty"`
	Swearing                *int `json:"swearing,omitempty"`
}

// UpdateAutoModSettingsResponse is the typed response for Update AutoMod Settings.
type UpdateAutoModSettingsResponse struct {
	Data []AutoModSettings `json:"data"`
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

// RemoveSuspiciousUserParams identifies the suspicious-user status to remove.
type RemoveSuspiciousUserParams struct {
	BroadcasterID string
	ModeratorID   string
	UserID        string
}

// UnbanRequest describes a pending or resolved unban request.
type UnbanRequest struct {
	ID               string     `json:"id"`
	BroadcasterID    string     `json:"broadcaster_id"`
	BroadcasterLogin string     `json:"broadcaster_login"`
	BroadcasterName  string     `json:"broadcaster_name"`
	ModeratorID      string     `json:"moderator_id"`
	ModeratorLogin   string     `json:"moderator_login"`
	ModeratorName    string     `json:"moderator_name"`
	UserID           string     `json:"user_id"`
	UserLogin        string     `json:"user_login"`
	UserName         string     `json:"user_name"`
	Text             string     `json:"text"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	ResolvedAt       *time.Time `json:"resolved_at"`
	ResolutionText   *string    `json:"resolution_text"`
}

// GetUnbanRequestsResponse is the typed response for Get Unban Requests.
type GetUnbanRequestsResponse struct {
	Data       []UnbanRequest `json:"data"`
	Pagination Pagination     `json:"pagination"`
}

// ResolveUnbanRequestParams identifies the unban request to resolve and how to resolve it.
type ResolveUnbanRequestParams struct {
	BroadcasterID  string
	ModeratorID    string
	UnbanRequestID string
	Status         string
	ResolutionText string
}

// ResolveUnbanRequestResponse is the typed response for Resolve Unban Request.
type ResolveUnbanRequestResponse struct {
	Data []UnbanRequest `json:"data"`
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

// GetModeratedChannels fetches the channels that the specified user can moderate.
func (s *ModerationService) GetModeratedChannels(ctx context.Context, params GetModeratedChannelsParams) (*GetModeratedChannelsResponse, *Response, error) {
	query := url.Values{}
	query.Set("user_id", params.UserID)
	addCursorParams(query, params.CursorParams)

	var resp GetModeratedChannelsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/moderation/channels",
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

// CheckAutoModStatus checks whether AutoMod would flag the specified messages for review.
func (s *ModerationService) CheckAutoModStatus(ctx context.Context, params CheckAutoModStatusParams, req CheckAutoModStatusRequest) (*CheckAutoModStatusResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)

	var resp CheckAutoModStatusResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/moderation/enforcements/status",
		Query:  query,
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// WarnUser warns a user in the broadcaster's chat room.
func (s *ModerationService) WarnUser(ctx context.Context, params WarnUserParams, req WarnUserRequest) (*WarnUserResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	var resp WarnUserResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/moderation/warnings",
		Query:  query,
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// ManageHeldAutoModMessage allows or denies a message held by AutoMod.
func (s *ModerationService) ManageHeldAutoModMessage(ctx context.Context, req ManageHeldAutoModMessageRequest) (*Response, error) {
	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/moderation/automod/message",
		Body:   req,
	}, nil)
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

// GetAutoModSettings fetches the broadcaster's AutoMod settings.
func (s *ModerationService) GetAutoModSettings(ctx context.Context, params AutoModSettingsParams) (*GetAutoModSettingsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	var resp GetAutoModSettingsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/moderation/automod/settings",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// AddSuspiciousUserStatus applies a suspicious-user status to a chatter.
func (s *ModerationService) AddSuspiciousUserStatus(ctx context.Context, params ShieldModeParams, req SuspiciousUserStatusChangeRequest) (*SuspiciousUserStatusChangeResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	var resp SuspiciousUserStatusChangeResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/moderation/suspicious_users",
		Query:  query,
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// UpdateAutoModSettings updates the broadcaster's AutoMod settings.
func (s *ModerationService) UpdateAutoModSettings(ctx context.Context, params AutoModSettingsParams, req UpdateAutoModSettingsRequest) (*UpdateAutoModSettingsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	mergedReq, meta, err := s.mergeAutoModSettingsRequest(ctx, params, req)
	if err != nil {
		return nil, meta, err
	}

	var resp UpdateAutoModSettingsResponse
	meta, err = s.client.Do(ctx, RawRequest{
		Method: http.MethodPut,
		Path:   "/moderation/automod/settings",
		Query:  query,
		Body:   mergedReq,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

func (s *ModerationService) mergeAutoModSettingsRequest(ctx context.Context, params AutoModSettingsParams, req UpdateAutoModSettingsRequest) (UpdateAutoModSettingsRequest, *Response, error) {
	if req.OverallLevel != nil || !req.hasCategorySettings() {
		return req, nil, nil
	}

	currentResp, meta, err := s.GetAutoModSettings(ctx, params)
	if err != nil {
		return req, meta, err
	}
	if len(currentResp.Data) == 0 {
		return req, meta, fmt.Errorf("helix: get automod settings returned no data")
	}

	current := currentResp.Data[0]
	merged := req
	if merged.Aggression == nil {
		merged.Aggression = &current.Aggression
	}
	if merged.Bullying == nil {
		merged.Bullying = &current.Bullying
	}
	if merged.Disability == nil {
		merged.Disability = &current.Disability
	}
	if merged.Misogyny == nil {
		merged.Misogyny = &current.Misogyny
	}
	if merged.RaceEthnicityOrReligion == nil {
		merged.RaceEthnicityOrReligion = &current.RaceEthnicityOrReligion
	}
	if merged.SexBasedTerms == nil {
		merged.SexBasedTerms = &current.SexBasedTerms
	}
	if merged.SexualitySexOrGender == nil {
		merged.SexualitySexOrGender = &current.SexualitySexOrGender
	}
	if merged.Swearing == nil {
		merged.Swearing = &current.Swearing
	}

	return merged, nil, nil
}

func (r UpdateAutoModSettingsRequest) hasCategorySettings() bool {
	return r.Aggression != nil ||
		r.Bullying != nil ||
		r.Disability != nil ||
		r.Misogyny != nil ||
		r.RaceEthnicityOrReligion != nil ||
		r.SexBasedTerms != nil ||
		r.SexualitySexOrGender != nil ||
		r.Swearing != nil
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

// GetUnbanRequests fetches unban requests for a broadcaster's channel.
func (s *ModerationService) GetUnbanRequests(ctx context.Context, params GetUnbanRequestsParams) (*GetUnbanRequestsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("status", params.Status)
	if params.UserID != "" {
		query.Set("user_id", params.UserID)
	}
	addCursorParams(query, params.CursorParams)

	var resp GetUnbanRequestsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/moderation/unban_requests",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// RemoveSuspiciousUserStatus removes a suspicious-user status from a chatter.
func (s *ModerationService) RemoveSuspiciousUserStatus(ctx context.Context, params RemoveSuspiciousUserParams) (*SuspiciousUserStatusChangeResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("user_id", params.UserID)

	var resp SuspiciousUserStatusChangeResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/moderation/suspicious_users",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// ResolveUnbanRequest approves or denies an unban request.
func (s *ModerationService) ResolveUnbanRequest(ctx context.Context, params ResolveUnbanRequestParams) (*ResolveUnbanRequestResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("unban_request_id", params.UnbanRequestID)
	query.Set("status", params.Status)
	if params.ResolutionText != "" {
		query.Set("resolution_text", params.ResolutionText)
	}

	var resp ResolveUnbanRequestResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/moderation/unban_requests",
		Query:  query,
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
