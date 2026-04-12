package helix

import (
	"context"
	"net/url"
)

// EventSubService provides access to EventSub subscription APIs.
type EventSubService struct {
	client *Client
}

// EventSubCondition contains subscription condition fields.
type EventSubCondition map[string]string

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
