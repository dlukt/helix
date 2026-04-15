package helix

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ChannelPointsService provides access to Twitch channel points APIs.
type ChannelPointsService struct {
	client *Client
}

// GetCustomRewardsParams filters Get Custom Reward requests.
type GetCustomRewardsParams struct {
	BroadcasterID         string
	IDs                   []string
	OnlyManageableRewards *bool
}

// CreateCustomRewardRequest is the request body for Create Custom Rewards.
type CreateCustomRewardRequest struct {
	Title                             string `json:"title"`
	Cost                              int    `json:"cost"`
	Prompt                            string `json:"prompt,omitempty"`
	IsEnabled                         *bool  `json:"is_enabled,omitempty"`
	BackgroundColor                   string `json:"background_color,omitempty"`
	IsUserInputRequired               *bool  `json:"is_user_input_required,omitempty"`
	IsMaxPerStreamEnabled             *bool  `json:"is_max_per_stream_enabled,omitempty"`
	MaxPerStream                      *int   `json:"max_per_stream,omitempty"`
	IsMaxPerUserPerStreamEnabled      *bool  `json:"is_max_per_user_per_stream_enabled,omitempty"`
	MaxPerUserPerStream               *int   `json:"max_per_user_per_stream,omitempty"`
	IsGlobalCooldownEnabled           *bool  `json:"is_global_cooldown_enabled,omitempty"`
	GlobalCooldownSeconds             *int   `json:"global_cooldown_seconds,omitempty"`
	ShouldRedemptionsSkipRequestQueue *bool  `json:"should_redemptions_skip_request_queue,omitempty"`
}

// UpdateCustomRewardParams identifies the broadcaster and reward to update.
type UpdateCustomRewardParams struct {
	BroadcasterID string
	ID            string
}

// UpdateCustomRewardRequest is the request body for Update Custom Reward.
type UpdateCustomRewardRequest struct {
	Title                             *string `json:"title,omitempty"`
	Cost                              *int    `json:"cost,omitempty"`
	Prompt                            *string `json:"prompt,omitempty"`
	IsEnabled                         *bool   `json:"is_enabled,omitempty"`
	BackgroundColor                   *string `json:"background_color,omitempty"`
	IsUserInputRequired               *bool   `json:"is_user_input_required,omitempty"`
	IsPaused                          *bool   `json:"is_paused,omitempty"`
	IsMaxPerStreamEnabled             *bool   `json:"is_max_per_stream_enabled,omitempty"`
	MaxPerStream                      *int    `json:"max_per_stream,omitempty"`
	IsMaxPerUserPerStreamEnabled      *bool   `json:"is_max_per_user_per_stream_enabled,omitempty"`
	MaxPerUserPerStream               *int    `json:"max_per_user_per_stream,omitempty"`
	IsGlobalCooldownEnabled           *bool   `json:"is_global_cooldown_enabled,omitempty"`
	GlobalCooldownSeconds             *int    `json:"global_cooldown_seconds,omitempty"`
	ShouldRedemptionsSkipRequestQueue *bool   `json:"should_redemptions_skip_request_queue,omitempty"`
}

// DeleteCustomRewardParams identifies the broadcaster and reward to delete.
type DeleteCustomRewardParams struct {
	BroadcasterID string
	ID            string
}

// GetCustomRewardRedemptionsParams filters Get Custom Reward Redemption requests.
type GetCustomRewardRedemptionsParams struct {
	CursorParams
	BroadcasterID string
	RewardID      string
	IDs           []string
	Status        string
	Sort          string
}

// UpdateCustomRewardRedemptionStatusParams identifies the redemptions to update.
type UpdateCustomRewardRedemptionStatusParams struct {
	BroadcasterID string
	RewardID      string
	IDs           []string
}

// UpdateCustomRewardRedemptionStatusRequest is the request body for Update Redemption Status.
type UpdateCustomRewardRedemptionStatusRequest struct {
	Status string `json:"status"`
}

// CustomRewardImage describes custom reward image URLs.
type CustomRewardImage struct {
	URL1x string `json:"url_1x"`
	URL2x string `json:"url_2x"`
	URL4x string `json:"url_4x"`
}

// MaxPerStreamSetting describes the per-stream limit for a custom reward.
type MaxPerStreamSetting struct {
	IsEnabled    bool `json:"is_enabled"`
	MaxPerStream int  `json:"max_per_stream"`
}

// MaxPerUserPerStreamSetting describes the per-user per-stream limit for a custom reward.
type MaxPerUserPerStreamSetting struct {
	IsEnabled           bool `json:"is_enabled"`
	MaxPerUserPerStream int  `json:"max_per_user_per_stream"`
}

// GlobalCooldownSetting describes the global cooldown for a custom reward.
type GlobalCooldownSetting struct {
	IsEnabled             bool `json:"is_enabled"`
	GlobalCooldownSeconds int  `json:"global_cooldown_seconds"`
}

// CustomReward describes a channel points custom reward.
type CustomReward struct {
	BroadcasterName                   string                     `json:"broadcaster_name"`
	BroadcasterLogin                  string                     `json:"broadcaster_login"`
	BroadcasterID                     string                     `json:"broadcaster_id"`
	ID                                string                     `json:"id"`
	Image                             *CustomRewardImage         `json:"image"`
	BackgroundColor                   string                     `json:"background_color"`
	IsEnabled                         bool                       `json:"is_enabled"`
	Cost                              int                        `json:"cost"`
	Title                             string                     `json:"title"`
	Prompt                            string                     `json:"prompt"`
	IsUserInputRequired               bool                       `json:"is_user_input_required"`
	MaxPerStreamSetting               MaxPerStreamSetting        `json:"max_per_stream_setting"`
	MaxPerUserPerStreamSetting        MaxPerUserPerStreamSetting `json:"max_per_user_per_stream_setting"`
	GlobalCooldownSetting             GlobalCooldownSetting      `json:"global_cooldown_setting"`
	IsPaused                          bool                       `json:"is_paused"`
	IsInStock                         bool                       `json:"is_in_stock"`
	DefaultImage                      CustomRewardImage          `json:"default_image"`
	ShouldRedemptionsSkipRequestQueue bool                       `json:"should_redemptions_skip_request_queue"`
	RedemptionsRedeemedCurrentStream  *int                       `json:"redemptions_redeemed_current_stream"`
	CooldownExpiresAt                 *time.Time                 `json:"cooldown_expires_at"`
}

// GetCustomRewardsResponse is the typed response for custom reward list endpoints.
type GetCustomRewardsResponse struct {
	Data []CustomReward `json:"data"`
}

// CustomRewardRedemptionReward describes the nested reward metadata on a redemption.
type CustomRewardRedemptionReward struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Prompt string `json:"prompt"`
	Cost   int    `json:"cost"`
}

// CustomRewardRedemption describes a custom reward redemption.
type CustomRewardRedemption struct {
	BroadcasterName  string                       `json:"broadcaster_name"`
	BroadcasterLogin string                       `json:"broadcaster_login"`
	BroadcasterID    string                       `json:"broadcaster_id"`
	ID               string                       `json:"id"`
	UserID           string                       `json:"user_id"`
	UserLogin        string                       `json:"user_login"`
	UserName         string                       `json:"user_name"`
	UserInput        string                       `json:"user_input"`
	Status           string                       `json:"status"`
	RedeemedAt       time.Time                    `json:"redeemed_at"`
	Reward           CustomRewardRedemptionReward `json:"reward"`
}

// GetCustomRewardRedemptionsResponse is the typed response for custom reward redemption endpoints.
type GetCustomRewardRedemptionsResponse struct {
	Data       []CustomRewardRedemption `json:"data"`
	Pagination Pagination               `json:"pagination"`
}

// GetRewards fetches custom rewards for a broadcaster.
func (s *ChannelPointsService) GetRewards(ctx context.Context, params GetCustomRewardsParams) (*GetCustomRewardsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	addRepeated(query, "id", params.IDs)
	if params.OnlyManageableRewards != nil {
		query.Set("only_manageable_rewards", strconv.FormatBool(*params.OnlyManageableRewards))
	}

	var resp GetCustomRewardsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/channel_points/custom_rewards",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// CreateReward creates a custom reward for a broadcaster.
func (s *ChannelPointsService) CreateReward(ctx context.Context, broadcasterID string, req CreateCustomRewardRequest) (*GetCustomRewardsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", broadcasterID)

	var resp GetCustomRewardsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/channel_points/custom_rewards",
		Query:  query,
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// UpdateReward updates a custom reward for a broadcaster.
func (s *ChannelPointsService) UpdateReward(ctx context.Context, params UpdateCustomRewardParams, req UpdateCustomRewardRequest) (*GetCustomRewardsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("id", params.ID)

	var resp GetCustomRewardsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/channel_points/custom_rewards",
		Query:  query,
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// DeleteReward deletes a custom reward for a broadcaster.
func (s *ChannelPointsService) DeleteReward(ctx context.Context, params DeleteCustomRewardParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("id", params.ID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/channel_points/custom_rewards",
		Query:  query,
	}, nil)
}

// GetRedemptions fetches redemptions for a custom reward.
func (s *ChannelPointsService) GetRedemptions(ctx context.Context, params GetCustomRewardRedemptionsParams) (*GetCustomRewardRedemptionsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("reward_id", params.RewardID)
	addRepeated(query, "id", params.IDs)
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.Sort != "" {
		query.Set("sort", params.Sort)
	}
	addCursorParams(query, params.CursorParams)

	var resp GetCustomRewardRedemptionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/channel_points/custom_rewards/redemptions",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// UpdateRedemptionStatus updates the status of one or more custom reward redemptions.
func (s *ChannelPointsService) UpdateRedemptionStatus(ctx context.Context, params UpdateCustomRewardRedemptionStatusParams, req UpdateCustomRewardRedemptionStatusRequest) (*GetCustomRewardRedemptionsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("reward_id", params.RewardID)
	addRepeated(query, "id", params.IDs)

	var resp GetCustomRewardRedemptionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/channel_points/custom_rewards/redemptions",
		Query:  query,
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
