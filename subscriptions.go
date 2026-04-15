package helix

import (
	"context"
	"net/http"
	"net/url"
)

// SubscriptionsService provides access to Twitch subscription APIs.
type SubscriptionsService struct {
	client *Client
}

// GetBroadcasterSubscriptionsParams filters Get Broadcaster Subscriptions requests.
type GetBroadcasterSubscriptionsParams struct {
	CursorParams
	BroadcasterID string
	UserIDs       []string
}

// CheckUserSubscriptionParams identifies a broadcaster and user for subscription lookup.
type CheckUserSubscriptionParams struct {
	BroadcasterID string
	UserID        string
}

// Subscription describes a user's subscription to a broadcaster.
type Subscription struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
	GifterID         string `json:"gifter_id"`
	GifterLogin      string `json:"gifter_login"`
	GifterName       string `json:"gifter_name"`
	IsGift           bool   `json:"is_gift"`
	Tier             string `json:"tier"`
	PlanName         string `json:"plan_name"`
	UserID           string `json:"user_id"`
	UserName         string `json:"user_name"`
	UserLogin        string `json:"user_login"`
}

// GetBroadcasterSubscriptionsResponse is the typed response for Get Broadcaster Subscriptions.
type GetBroadcasterSubscriptionsResponse struct {
	Data       []Subscription `json:"data"`
	Pagination Pagination     `json:"pagination"`
	Total      int            `json:"total"`
	Points     int            `json:"points"`
}

// CheckUserSubscriptionResponse is the typed response for Check User Subscription.
type CheckUserSubscriptionResponse struct {
	Data []Subscription `json:"data"`
}

// GetBroadcaster fetches subscriptions for a broadcaster.
func (s *SubscriptionsService) GetBroadcaster(ctx context.Context, params GetBroadcasterSubscriptionsParams) (*GetBroadcasterSubscriptionsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	addRepeated(query, "user_id", params.UserIDs)
	addCursorParams(query, params.CursorParams)

	var resp GetBroadcasterSubscriptionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/subscriptions",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// CheckUser fetches the subscription relationship for a specific user and broadcaster.
func (s *SubscriptionsService) CheckUser(ctx context.Context, params CheckUserSubscriptionParams) (*CheckUserSubscriptionResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("user_id", params.UserID)

	var resp CheckUserSubscriptionResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/subscriptions/user",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
