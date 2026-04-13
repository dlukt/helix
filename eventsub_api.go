package helix

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
)

// EventSubService provides access to EventSub subscription APIs.
type EventSubService struct {
	client *Client
}

// EventSubCondition contains subscription condition fields.
type EventSubCondition map[string]string

// ChannelFollowV2Condition targets channel.follow version 2 subscriptions.
type ChannelFollowV2Condition struct {
	BroadcasterUserID string `json:"broadcaster_user_id,omitempty"`
	ModeratorUserID   string `json:"moderator_user_id,omitempty"`
}

// StreamOnlineV1Condition targets stream.online version 1 subscriptions.
type StreamOnlineV1Condition struct {
	BroadcasterUserID string `json:"broadcaster_user_id,omitempty"`
}

// BroadcasterUserIDCondition targets subscriptions scoped by broadcaster_user_id.
type BroadcasterUserIDCondition struct {
	BroadcasterUserID string `json:"broadcaster_user_id,omitempty"`
}

// ChannelRaidV1Condition targets channel.raid version 1 subscriptions.
type ChannelRaidV1Condition struct {
	FromBroadcasterUserID string `json:"from_broadcaster_user_id,omitempty"`
	ToBroadcasterUserID   string `json:"to_broadcaster_user_id,omitempty"`
}

// EventSubTransport describes subscription delivery transport.
type EventSubTransport struct {
	Method    string `json:"method"`
	Callback  string `json:"callback,omitempty"`
	Secret    string `json:"secret,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	ConduitID string `json:"conduit_id,omitempty"`
}

// EventSubSubscription is a subscription resource.
type EventSubSubscription struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Cost      int               `json:"cost"`
	Condition EventSubCondition `json:"condition"`
	Transport EventSubTransport `json:"transport"`
}

// CreateEventSubSubscriptionRequest creates a subscription.
type CreateEventSubSubscriptionRequest struct {
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Condition EventSubCondition `json:"condition"`
	Transport EventSubTransport `json:"transport"`
}

// CreateChannelFollowV2Request creates a typed channel.follow@2 subscription.
type CreateChannelFollowV2Request struct {
	Condition ChannelFollowV2Condition
	Transport EventSubTransport
}

// CreateStreamOnlineV1Request creates a typed stream.online@1 subscription.
type CreateStreamOnlineV1Request struct {
	Condition StreamOnlineV1Condition
	Transport EventSubTransport
}

// CreateChannelSubscribeV1Request creates a typed channel.subscribe@1 subscription.
type CreateChannelSubscribeV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelSubscriptionGiftV1Request creates a typed channel.subscription.gift@1 subscription.
type CreateChannelSubscriptionGiftV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelSubscriptionEndV1Request creates a typed channel.subscription.end@1 subscription.
type CreateChannelSubscriptionEndV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelSubscriptionMessageV1Request creates a typed channel.subscription.message@1 subscription.
type CreateChannelSubscriptionMessageV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelCheerV1Request creates a typed channel.cheer@1 subscription.
type CreateChannelCheerV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelRaidV1Request creates a typed channel.raid@1 subscription.
type CreateChannelRaidV1Request struct {
	Condition ChannelRaidV1Condition
	Transport EventSubTransport
}

// CreateEventSubSubscriptionResponse is returned by Create.
type CreateEventSubSubscriptionResponse struct {
	Data []EventSubSubscription `json:"data"`
}

// ListEventSubSubscriptionsParams filters List.
type ListEventSubSubscriptionsParams struct {
	Status string
	Type   string
	After  string
}

// ListEventSubSubscriptionsResponse is returned by List.
type ListEventSubSubscriptionsResponse struct {
	Data         []EventSubSubscription `json:"data"`
	Total        int                    `json:"total"`
	TotalCost    int                    `json:"total_cost"`
	MaxTotalCost int                    `json:"max_total_cost"`
	Pagination   Pagination             `json:"pagination"`
}

// Create creates a new EventSub subscription.
func (s *EventSubService) Create(ctx context.Context, req CreateEventSubSubscriptionRequest) (*CreateEventSubSubscriptionResponse, *Response, error) {
	var resp CreateEventSubSubscriptionResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: "POST",
		Path:   "/eventsub/subscriptions",
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// CreateChannelFollowV2 creates a typed channel.follow version 2 subscription.
func (s *EventSubService) CreateChannelFollowV2(ctx context.Context, req CreateChannelFollowV2Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.follow",
		Version:   "2",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateStreamOnlineV1 creates a typed stream.online version 1 subscription.
func (s *EventSubService) CreateStreamOnlineV1(ctx context.Context, req CreateStreamOnlineV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "stream.online",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSubscribeV1 creates a typed channel.subscribe version 1 subscription.
func (s *EventSubService) CreateChannelSubscribeV1(ctx context.Context, req CreateChannelSubscribeV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.subscribe",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSubscriptionGiftV1 creates a typed channel.subscription.gift version 1 subscription.
func (s *EventSubService) CreateChannelSubscriptionGiftV1(ctx context.Context, req CreateChannelSubscriptionGiftV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.subscription.gift",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSubscriptionEndV1 creates a typed channel.subscription.end version 1 subscription.
func (s *EventSubService) CreateChannelSubscriptionEndV1(ctx context.Context, req CreateChannelSubscriptionEndV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.subscription.end",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSubscriptionMessageV1 creates a typed channel.subscription.message version 1 subscription.
func (s *EventSubService) CreateChannelSubscriptionMessageV1(ctx context.Context, req CreateChannelSubscriptionMessageV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.subscription.message",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelCheerV1 creates a typed channel.cheer version 1 subscription.
func (s *EventSubService) CreateChannelCheerV1(ctx context.Context, req CreateChannelCheerV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.cheer",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelRaidV1 creates a typed channel.raid version 1 subscription.
func (s *EventSubService) CreateChannelRaidV1(ctx context.Context, req CreateChannelRaidV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	if req.Condition.FromBroadcasterUserID == "" && req.Condition.ToBroadcasterUserID == "" {
		return nil, nil, errors.New("helix: channel.raid requires exactly one of from_broadcaster_user_id or to_broadcaster_user_id")
	}
	if req.Condition.FromBroadcasterUserID != "" && req.Condition.ToBroadcasterUserID != "" {
		return nil, nil, errors.New("helix: channel.raid requires exactly one of from_broadcaster_user_id or to_broadcaster_user_id")
	}

	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.raid",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// List lists EventSub subscriptions.
func (s *EventSubService) List(ctx context.Context, params ListEventSubSubscriptionsParams) (*ListEventSubSubscriptionsResponse, *Response, error) {
	query := url.Values{}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.Type != "" {
		query.Set("type", params.Type)
	}
	if params.After != "" {
		query.Set("after", params.After)
	}

	var resp ListEventSubSubscriptionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: "GET",
		Path:   "/eventsub/subscriptions",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// Delete deletes an EventSub subscription by ID.
func (s *EventSubService) Delete(ctx context.Context, id string) (*Response, error) {
	query := url.Values{}
	query.Set("id", id)
	return s.client.Do(ctx, RawRequest{
		Method: "DELETE",
		Path:   "/eventsub/subscriptions",
		Query:  query,
	}, nil)
}

func marshalCondition(value any) (EventSubCondition, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var condition EventSubCondition
	if err := json.Unmarshal(raw, &condition); err != nil {
		return nil, err
	}
	return condition, nil
}
