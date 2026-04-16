package helix

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
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

// GetExtensionLiveChannelsParams filters Get Extension Live Channels requests.
type GetExtensionLiveChannelsParams struct {
	CursorParams
	ExtensionID string
}

// GetExtensionBitsProductsParams filters Get Extension Bits Products requests.
type GetExtensionBitsProductsParams struct {
	ShouldIncludeAll *bool
}

// GetExtensionSecretsParams identifies the extension whose JWT secrets are fetched.
type GetExtensionSecretsParams struct {
	ExtensionID string
}

// CreateExtensionSecretParams identifies the extension secret to create.
type CreateExtensionSecretParams struct {
	ExtensionID string
	Delay       int
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

// SendExtensionChatMessageRequest sends a message to a broadcaster's chat as the extension.
type SendExtensionChatMessageRequest struct {
	BroadcasterID    string `json:"-"`
	Text             string `json:"text"`
	ExtensionID      string `json:"extension_id"`
	ExtensionVersion string `json:"extension_version"`
}

// UpdateExtensionBitsProductRequest adds or updates a Bits product for an extension.
type UpdateExtensionBitsProductRequest struct {
	SKU           string                   `json:"sku"`
	Cost          ExtensionTransactionCost `json:"cost"`
	DisplayName   string                   `json:"display_name"`
	InDevelopment *bool                    `json:"in_development,omitempty"`
	Expiration    *time.Time               `json:"expiration,omitempty"`
	IsBroadcast   *bool                    `json:"is_broadcast,omitempty"`
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

// ExtensionBitsProduct describes a Bits-in-Extensions product.
type ExtensionBitsProduct struct {
	SKU           string                   `json:"sku"`
	Cost          ExtensionTransactionCost `json:"cost"`
	InDevelopment bool                     `json:"in_development"`
	DisplayName   string                   `json:"display_name"`
	Expiration    *time.Time               `json:"expiration"`
	IsBroadcast   bool                     `json:"is_broadcast"`
}

// UnmarshalJSON accepts empty-string expirations for non-expiring products.
func (p *ExtensionBitsProduct) UnmarshalJSON(data []byte) error {
	type wire struct {
		SKU           string                   `json:"sku"`
		Cost          ExtensionTransactionCost `json:"cost"`
		InDevelopment bool                     `json:"in_development"`
		DisplayName   string                   `json:"display_name"`
		Expiration    *string                  `json:"expiration"`
		IsBroadcast   bool                     `json:"is_broadcast"`
	}

	var raw wire
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*p = ExtensionBitsProduct{
		SKU:           raw.SKU,
		Cost:          raw.Cost,
		InDevelopment: raw.InDevelopment,
		DisplayName:   raw.DisplayName,
		IsBroadcast:   raw.IsBroadcast,
	}

	if raw.Expiration == nil || *raw.Expiration == "" {
		return nil
	}

	expiration, err := time.Parse(time.RFC3339, *raw.Expiration)
	if err != nil {
		return err
	}
	p.Expiration = &expiration

	return nil
}

// GetExtensionBitsProductsResponse is the typed response for extension Bits product endpoints.
type GetExtensionBitsProductsResponse struct {
	Data []ExtensionBitsProduct `json:"data"`
}

// ExtensionLiveChannel describes a live channel that has an extension installed or activated.
type ExtensionLiveChannel struct {
	BroadcasterID   string `json:"broadcaster_id"`
	BroadcasterName string `json:"broadcaster_name"`
	GameName        string `json:"game_name"`
	GameID          string `json:"game_id"`
	Title           string `json:"title"`
}

// GetExtensionLiveChannelsResponse is the typed response for Get Extension Live Channels.
type GetExtensionLiveChannelsResponse struct {
	Data       []ExtensionLiveChannel `json:"data"`
	Pagination Pagination             `json:"pagination"`
}

// ExtensionSecret describes a JWT secret for an extension.
type ExtensionSecret struct {
	Content   string    `json:"content"`
	ActiveAt  time.Time `json:"active_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ExtensionSecretDefinition describes one versioned secret payload returned by Twitch.
type ExtensionSecretDefinition struct {
	FormatVersion int               `json:"format_version"`
	Secrets       []ExtensionSecret `json:"secrets"`
}

// GetExtensionSecretsResponse is the typed response for Get/Create Extension Secret.
type GetExtensionSecretsResponse struct {
	Data []ExtensionSecretDefinition `json:"data"`
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

// GetBitsProducts fetches the Bits products owned by the extension.
func (s *ExtensionsService) GetBitsProducts(ctx context.Context, params ...GetExtensionBitsProductsParams) (*GetExtensionBitsProductsResponse, *Response, error) {
	query := url.Values{}
	if len(params) > 0 && params[0].ShouldIncludeAll != nil {
		query.Set("should_include_all", strconv.FormatBool(*params[0].ShouldIncludeAll))
	}

	var resp GetExtensionBitsProductsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/bits/extensions",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetLiveChannels fetches channels that are live and have the extension installed or activated.
func (s *ExtensionsService) GetLiveChannels(ctx context.Context, params GetExtensionLiveChannelsParams) (*GetExtensionLiveChannelsResponse, *Response, error) {
	query := url.Values{}
	query.Set("extension_id", params.ExtensionID)
	addCursorParams(query, params.CursorParams)

	httpResp, meta, err := s.client.doRaw(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/extensions/live",
		Query:  query,
	})
	if err != nil {
		return nil, meta, err
	}
	body, err := replayableBody(httpResp)
	if err != nil {
		return nil, meta, err
	}

	var resp struct {
		Data       []ExtensionLiveChannel `json:"data"`
		Pagination json.RawMessage        `json:"pagination"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, meta, err
	}
	pagination, err := decodeExtensionsLivePagination(resp.Pagination)
	if err != nil {
		return nil, meta, err
	}
	meta.Pagination = pagination

	return &GetExtensionLiveChannelsResponse{
		Data:       resp.Data,
		Pagination: pagination,
	}, meta, nil
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

// GetReleased fetches metadata for a released extension.
func (s *ExtensionsService) GetReleased(ctx context.Context, params GetExtensionParams) (*GetExtensionsResponse, *Response, error) {
	query := url.Values{}
	query.Set("extension_id", params.ExtensionID)
	if params.ExtensionVersion != "" {
		query.Set("extension_version", params.ExtensionVersion)
	}

	var resp GetExtensionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/extensions/released",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetSecrets fetches the JWT secrets associated with an extension.
func (s *ExtensionsService) GetSecrets(ctx context.Context, params GetExtensionSecretsParams) (*GetExtensionSecretsResponse, *Response, error) {
	query := url.Values{}
	query.Set("extension_id", params.ExtensionID)

	var resp GetExtensionSecretsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/extensions/jwt/secrets",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// CreateSecret creates a new JWT secret for an extension.
func (s *ExtensionsService) CreateSecret(ctx context.Context, params CreateExtensionSecretParams) (*GetExtensionSecretsResponse, *Response, error) {
	query := url.Values{}
	query.Set("extension_id", params.ExtensionID)
	if params.Delay > 0 {
		query.Set("delay", strconv.Itoa(params.Delay))
	}

	var resp GetExtensionSecretsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/extensions/jwt/secrets",
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

// UpdateBitsProduct adds or updates a Bits product for the extension.
func (s *ExtensionsService) UpdateBitsProduct(ctx context.Context, req UpdateExtensionBitsProductRequest) (*GetExtensionBitsProductsResponse, *Response, error) {
	var resp GetExtensionBitsProductsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPut,
		Path:   "/bits/extensions",
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// SendChatMessage sends a chat message as the extension to the broadcaster's chat room.
func (s *ExtensionsService) SendChatMessage(ctx context.Context, req SendExtensionChatMessageRequest) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", req.BroadcasterID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/extensions/chat",
		Query:  query,
		Body: struct {
			Text             string `json:"text"`
			ExtensionID      string `json:"extension_id"`
			ExtensionVersion string `json:"extension_version"`
		}{
			Text:             req.Text,
			ExtensionID:      req.ExtensionID,
			ExtensionVersion: req.ExtensionVersion,
		},
	}, nil)
}

func decodeExtensionsLivePagination(raw json.RawMessage) (Pagination, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return Pagination{}, nil
	}
	if raw[0] == '"' {
		var cursor string
		if err := json.Unmarshal(raw, &cursor); err != nil {
			return Pagination{}, err
		}
		return Pagination{Cursor: cursor}, nil
	}
	var pagination Pagination
	if err := json.Unmarshal(raw, &pagination); err != nil {
		return Pagination{}, err
	}
	return pagination, nil
}
