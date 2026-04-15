package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// ExtensionsService provides access to Twitch extension APIs.
type ExtensionsService struct {
	client *Client
}

// GetExtensionTransactionsParams filters Get Extension Transactions requests.
type GetExtensionTransactionsParams struct {
	CursorParams
	ExtensionID    string
	TransactionIDs []string
}

// GetExtensionParams identifies an extension and optional version to fetch.
type GetExtensionParams struct {
	ExtensionID      string
	ExtensionVersion string
}

// GetExtensionConfigurationSegmentParams identifies the configuration segment to fetch.
type GetExtensionConfigurationSegmentParams struct {
	BroadcasterID string
	ExtensionID   string
	Segment       string
	Segments      []string
}

// SetExtensionConfigurationSegmentRequest updates a hosted extension configuration segment.
type SetExtensionConfigurationSegmentRequest struct {
	ExtensionID   string  `json:"extension_id"`
	Segment       string  `json:"segment"`
	BroadcasterID string  `json:"broadcaster_id,omitempty"`
	Content       *string `json:"content,omitempty"`
	Version       string  `json:"version,omitempty"`
}

// SetExtensionRequiredConfigurationRequest updates the required configuration string for an extension.
type SetExtensionRequiredConfigurationRequest struct {
	ExtensionID           string `json:"extension_id"`
	ExtensionVersion      string `json:"extension_version"`
	RequiredConfiguration string `json:"required_configuration"`
	BroadcasterID         string `json:"-"`
}

// SendExtensionPubSubMessageRequest sends a message to extension clients.
type SendExtensionPubSubMessageRequest struct {
	Target            []string `json:"target"`
	BroadcasterID     string   `json:"broadcaster_id,omitempty"`
	IsGlobalBroadcast *bool    `json:"is_global_broadcast,omitempty"`
	Message           string   `json:"message"`
}

// ExtensionTransactionCost describes the cost of an extension transaction.
type ExtensionTransactionCost struct {
	Amount int    `json:"amount"`
	Type   string `json:"type"`
}

// ExtensionTransactionProductData describes the purchased extension product.
type ExtensionTransactionProductData struct {
	Domain        string                   `json:"domain"`
	SKU           string                   `json:"sku"`
	InDevelopment bool                     `json:"inDevelopment"`
	DisplayName   string                   `json:"displayName"`
	Broadcast     bool                     `json:"broadcast"`
	Expiration    string                   `json:"expiration"`
	Cost          ExtensionTransactionCost `json:"cost"`
}

// ExtensionTransaction describes a single extension transaction.
type ExtensionTransaction struct {
	ID               string                          `json:"id"`
	Timestamp        time.Time                       `json:"timestamp"`
	BroadcasterID    string                          `json:"broadcaster_id"`
	BroadcasterLogin string                          `json:"broadcaster_login"`
	BroadcasterName  string                          `json:"broadcaster_name"`
	UserID           string                          `json:"user_id"`
	UserLogin        string                          `json:"user_login"`
	UserName         string                          `json:"user_name"`
	ProductType      string                          `json:"product_type"`
	ProductData      ExtensionTransactionProductData `json:"product_data"`
}

// GetExtensionTransactionsResponse is the typed response for Get Extension Transactions.
type GetExtensionTransactionsResponse struct {
	Data       []ExtensionTransaction `json:"data"`
	Pagination Pagination             `json:"pagination"`
}

// ExtensionViewConfig describes how an extension view renders.
type ExtensionViewConfig struct {
	ViewerURL              string `json:"viewer_url"`
	Height                 int    `json:"height,omitempty"`
	CanLinkExternalContent bool   `json:"can_link_external_content,omitempty"`
	AspectWidth            int    `json:"aspect_width,omitempty"`
	AspectHeight           int    `json:"aspect_height,omitempty"`
	AspectRatioX           int    `json:"aspect_ratio_x,omitempty"`
	AspectRatioY           int    `json:"aspect_ratio_y,omitempty"`
	Autoscale              bool   `json:"autoscale,omitempty"`
	ScalePixels            int    `json:"scale_pixels,omitempty"`
	TargetHeight           int    `json:"target_height,omitempty"`
	Size                   int    `json:"size,omitempty"`
	Zoom                   bool   `json:"zoom,omitempty"`
	ZoomPixels             int    `json:"zoom_pixels,omitempty"`
}

// ExtensionViews describes the various views exposed by an extension.
type ExtensionViews struct {
	Mobile       *ExtensionViewConfig `json:"mobile"`
	Panel        *ExtensionViewConfig `json:"panel"`
	VideoOverlay *ExtensionViewConfig `json:"video_overlay"`
	Component    *ExtensionViewConfig `json:"component"`
	Config       *ExtensionViewConfig `json:"config"`
}

// Extension describes a Twitch extension.
type Extension struct {
	AuthorName                string            `json:"author_name"`
	BitsEnabled               bool              `json:"bits_enabled"`
	CanInstall                bool              `json:"can_install"`
	ConfigurationLocation     string            `json:"configuration_location"`
	Description               string            `json:"description"`
	EULATOSURL                string            `json:"eula_tos_url"`
	HasChatSupport            bool              `json:"has_chat_support"`
	IconURL                   string            `json:"icon_url"`
	IconURLs                  map[string]string `json:"icon_urls"`
	ID                        string            `json:"id"`
	Name                      string            `json:"name"`
	PrivacyPolicyURL          string            `json:"privacy_policy_url"`
	RequestIdentityLink       bool              `json:"request_identity_link"`
	ScreenshotURLs            []string          `json:"screenshot_urls"`
	State                     string            `json:"state"`
	SubscriptionsSupportLevel string            `json:"subscriptions_support_level"`
	Summary                   string            `json:"summary"`
	SupportEmail              string            `json:"support_email"`
	Version                   string            `json:"version"`
	ViewerSummary             string            `json:"viewer_summary"`
	Views                     ExtensionViews    `json:"views"`
	AllowlistedConfigURLs     []string          `json:"allowlisted_config_urls"`
	AllowlistedPanelURLs      []string          `json:"allowlisted_panel_urls"`
}

// GetExtensionsResponse is the typed response for Get Extensions.
type GetExtensionsResponse struct {
	Data []Extension `json:"data"`
}

// ExtensionConfigurationSegment describes one hosted extension configuration segment.
type ExtensionConfigurationSegment struct {
	Content       string `json:"content"`
	Version       string `json:"version"`
	BroadcasterID string `json:"broadcaster_id"`
	ExtensionID   string `json:"extension_id"`
	Segment       string `json:"segment"`
}

// GetExtensionConfigurationSegmentResponse is the typed response for Get Extension Configuration Segment.
type GetExtensionConfigurationSegmentResponse struct {
	Data []ExtensionConfigurationSegment `json:"data"`
}

// GetTransactions fetches an extension's transactions.
func (s *ExtensionsService) GetTransactions(ctx context.Context, params GetExtensionTransactionsParams) (*GetExtensionTransactionsResponse, *Response, error) {
	query := url.Values{}
	if params.ExtensionID != "" {
		query.Set("extension_id", params.ExtensionID)
	}
	addRepeated(query, "id", params.TransactionIDs)
	addCursorParams(query, params.CursorParams)

	var resp GetExtensionTransactionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/extensions/transactions",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// Get fetches metadata for an extension.
func (s *ExtensionsService) Get(ctx context.Context, params GetExtensionParams) (*GetExtensionsResponse, *Response, error) {
	query := url.Values{}
	query.Set("extension_id", params.ExtensionID)
	if params.ExtensionVersion != "" {
		query.Set("extension_version", params.ExtensionVersion)
	}

	var resp GetExtensionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/extensions",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetConfigurationSegment fetches a hosted extension configuration segment.
func (s *ExtensionsService) GetConfigurationSegment(ctx context.Context, params GetExtensionConfigurationSegmentParams) (*GetExtensionConfigurationSegmentResponse, *Response, error) {
	query := url.Values{}
	if params.BroadcasterID != "" {
		query.Set("broadcaster_id", params.BroadcasterID)
	}
	query.Set("extension_id", params.ExtensionID)
	segments := append([]string{}, params.Segments...)
	if params.Segment != "" {
		segments = append([]string{params.Segment}, segments...)
	}
	addRepeated(query, "segment", segments)

	var resp GetExtensionConfigurationSegmentResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/extensions/configurations",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// SetConfigurationSegment updates a hosted extension configuration segment.
func (s *ExtensionsService) SetConfigurationSegment(ctx context.Context, req SetExtensionConfigurationSegmentRequest) (*Response, error) {
	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPut,
		Path:   "/extensions/configurations",
		Body:   req,
	}, nil)
}

// SetRequiredConfiguration updates the required configuration string for an extension version.
func (s *ExtensionsService) SetRequiredConfiguration(ctx context.Context, req SetExtensionRequiredConfigurationRequest) (*Response, error) {
	query := url.Values{}
	if req.BroadcasterID != "" {
		query.Set("broadcaster_id", req.BroadcasterID)
	}

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPut,
		Path:   "/extensions/required_configuration",
		Query:  query,
		Body:   req,
	}, nil)
}

// SendPubSubMessage sends a PubSub message to extension clients.
func (s *ExtensionsService) SendPubSubMessage(ctx context.Context, req SendExtensionPubSubMessageRequest) (*Response, error) {
	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/extensions/pubsub",
		Body:   req,
	}, nil)
}
