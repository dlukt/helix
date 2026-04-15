package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// EntitlementsService provides access to Twitch entitlements APIs.
type EntitlementsService struct {
	client *Client
}

// GetDropsEntitlementsParams filters Get Drops Entitlements requests.
type GetDropsEntitlementsParams struct {
	CursorParams
	IDs               []string
	UserID            string
	GameID            string
	FulfillmentStatus string
}

// DropsEntitlement describes a single drop entitlement.
type DropsEntitlement struct {
	ID                string    `json:"id"`
	BenefitID         string    `json:"benefit_id"`
	Timestamp         time.Time `json:"timestamp"`
	UserID            string    `json:"user_id"`
	GameID            string    `json:"game_id"`
	FulfillmentStatus string    `json:"fulfillment_status"`
	LastUpdated       time.Time `json:"last_updated"`
}

// GetDropsEntitlementsResponse is the typed response for Get Drops Entitlements.
type GetDropsEntitlementsResponse struct {
	Data       []DropsEntitlement `json:"data"`
	Pagination Pagination         `json:"pagination"`
}

// UpdateDropsEntitlementsRequest is the request body for Update Drops Entitlements.
type UpdateDropsEntitlementsRequest struct {
	EntitlementIDs    []string `json:"entitlement_ids,omitempty"`
	FulfillmentStatus string   `json:"fulfillment_status,omitempty"`
}

// EntitlementUpdateResult describes the result of attempting to update one or more entitlements.
type EntitlementUpdateResult struct {
	Status string   `json:"status"`
	IDs    []string `json:"ids"`
}

// UpdateDropsEntitlementsResponse is the typed response for Update Drops Entitlements.
type UpdateDropsEntitlementsResponse struct {
	Data []EntitlementUpdateResult `json:"data"`
}

// GetDrops fetches drop entitlements for the current app or user context.
func (s *EntitlementsService) GetDrops(ctx context.Context, params GetDropsEntitlementsParams) (*GetDropsEntitlementsResponse, *Response, error) {
	query := url.Values{}
	addRepeated(query, "id", params.IDs)
	if params.UserID != "" {
		query.Set("user_id", params.UserID)
	}
	if params.GameID != "" {
		query.Set("game_id", params.GameID)
	}
	if params.FulfillmentStatus != "" {
		query.Set("fulfillment_status", params.FulfillmentStatus)
	}
	addCursorParams(query, params.CursorParams)

	var resp GetDropsEntitlementsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/entitlements/drops",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// UpdateDrops updates the fulfillment status for one or more drop entitlements.
func (s *EntitlementsService) UpdateDrops(ctx context.Context, req UpdateDropsEntitlementsRequest) (*UpdateDropsEntitlementsResponse, *Response, error) {
	var resp UpdateDropsEntitlementsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/entitlements/drops",
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
