package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// UsersService provides access to the users API.
type UsersService struct {
	client *Client
}

// GetUsersParams filters Get Users requests.
type GetUsersParams struct {
	IDs    []string
	Logins []string
}

// GetAuthorizationsByUserParams filters Get Authorization By User requests.
type GetAuthorizationsByUserParams struct {
	UserIDs []string
}

// GetUserActiveExtensionsParams filters Get User Active Extensions requests.
type GetUserActiveExtensionsParams struct {
	UserID string
}

// UpdateUserParams identifies the mutable user fields to update.
type UpdateUserParams struct {
	Description *string
}

// GetUserBlockListParams filters Get User Block List requests.
type GetUserBlockListParams struct {
	CursorParams
	UserID string
	// BroadcasterID is deprecated. Use UserID.
	BroadcasterID string
}

// BlockUserParams identifies the user to block and the optional block metadata.
type BlockUserParams struct {
	TargetUserID  string
	SourceContext string
	Reason        string
}

// UnblockUserParams identifies the user to unblock.
type UnblockUserParams struct {
	TargetUserID string
}

// User is a Twitch user.
type User struct {
	ID              string    `json:"id"`
	Login           string    `json:"login"`
	DisplayName     string    `json:"display_name"`
	Type            string    `json:"type"`
	BroadcasterType string    `json:"broadcaster_type"`
	Description     string    `json:"description"`
	ProfileImageURL string    `json:"profile_image_url"`
	OfflineImageURL string    `json:"offline_image_url"`
	ViewCount       int       `json:"view_count"`
	Email           string    `json:"email"`
	CreatedAt       time.Time `json:"created_at"`
}

// GetUsersResponse is the typed response for Get Users.
type GetUsersResponse struct {
	Data []User
}

// UserAuthorization describes the scopes a user has granted to the application.
type UserAuthorization struct {
	UserID    string   `json:"user_id"`
	UserName  string   `json:"user_name"`
	UserLogin string   `json:"user_login"`
	Scopes    []string `json:"scopes"`
}

// GetAuthorizationsByUserResponse is the typed response for Get Authorization By User.
type GetAuthorizationsByUserResponse struct {
	Data []UserAuthorization `json:"data"`
}

// UserExtension describes an installed extension.
type UserExtension struct {
	ID          string   `json:"id"`
	Version     string   `json:"version"`
	Name        string   `json:"name"`
	CanActivate bool     `json:"can_activate"`
	Type        []string `json:"type"`
}

// GetUserExtensionsResponse is the typed response for Get User Extensions.
type GetUserExtensionsResponse struct {
	Data []UserExtension `json:"data"`
}

// UserActiveExtensionSlot describes an active or configured extension slot.
type UserActiveExtensionSlot struct {
	Active  bool   `json:"active"`
	ID      string `json:"id,omitempty"`
	Version string `json:"version,omitempty"`
	Name    string `json:"name,omitempty"`
	X       *int   `json:"x,omitempty"`
	Y       *int   `json:"y,omitempty"`
}

// UserActiveExtensions groups active extensions by extension type and slot number.
type UserActiveExtensions struct {
	Panel     map[string]UserActiveExtensionSlot `json:"panel,omitempty"`
	Overlay   map[string]UserActiveExtensionSlot `json:"overlay,omitempty"`
	Component map[string]UserActiveExtensionSlot `json:"component,omitempty"`
}

// GetUserActiveExtensionsResponse is the typed response for Get User Active Extensions.
type GetUserActiveExtensionsResponse struct {
	Data UserActiveExtensions `json:"data"`
}

// UpdateUserExtensionsRequest is the request body for Update User Extensions.
type UpdateUserExtensionsRequest struct {
	Data UserActiveExtensions `json:"data"`
}

// UpdateUserExtensionsResponse is the typed response for Update User Extensions.
type UpdateUserExtensionsResponse struct {
	Data UserActiveExtensions `json:"data"`
}

// UserBlock describes a blocked user.
type UserBlock struct {
	UserID      string `json:"user_id"`
	UserLogin   string `json:"user_login"`
	DisplayName string `json:"display_name"`
}

// GetUserBlockListResponse is the typed response for Get User Block List.
type GetUserBlockListResponse struct {
	Data       []UserBlock `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// UpdateUserResponse is the typed response for Update User.
type UpdateUserResponse struct {
	Data []User `json:"data"`
}

// Get fetches users by id or login.
func (s *UsersService) Get(ctx context.Context, params GetUsersParams) (*GetUsersResponse, *Response, error) {
	query := url.Values{}
	addRepeated(query, "id", params.IDs)
	addRepeated(query, "login", params.Logins)

	var data []User
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/users",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetUsersResponse{Data: data}, meta, nil
}

// GetAuthorizationsByUser fetches the authorization scopes granted by one or more users.
func (s *UsersService) GetAuthorizationsByUser(ctx context.Context, params GetAuthorizationsByUserParams) (*GetAuthorizationsByUserResponse, *Response, error) {
	query := url.Values{}
	addRepeated(query, "user_id", params.UserIDs)

	var resp GetAuthorizationsByUserResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/authorization/users",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetExtensions fetches the extensions the authenticated user has installed.
func (s *UsersService) GetExtensions(ctx context.Context) (*GetUserExtensionsResponse, *Response, error) {
	var resp GetUserExtensionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/users/extensions/list",
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetActiveExtensions fetches the active extensions for a user or the authenticated user.
func (s *UsersService) GetActiveExtensions(ctx context.Context, params GetUserActiveExtensionsParams) (*GetUserActiveExtensionsResponse, *Response, error) {
	query := url.Values{}
	if params.UserID != "" {
		query.Set("user_id", params.UserID)
	}

	var resp GetUserActiveExtensionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/users/extensions",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// UpdateExtensions updates the authenticated user's active extensions configuration.
func (s *UsersService) UpdateExtensions(ctx context.Context, req UpdateUserExtensionsRequest) (*UpdateUserExtensionsResponse, *Response, error) {
	var resp UpdateUserExtensionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPut,
		Path:   "/users/extensions",
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// Update modifies the authenticated user's profile fields.
func (s *UsersService) Update(ctx context.Context, params UpdateUserParams) (*UpdateUserResponse, *Response, error) {
	query := url.Values{}
	if params.Description != nil {
		query.Set("description", *params.Description)
	}

	var resp UpdateUserResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPut,
		Path:   "/users",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetBlocks fetches the user's blocked users.
func (s *UsersService) GetBlocks(ctx context.Context, params GetUserBlockListParams) (*GetUserBlockListResponse, *Response, error) {
	query := url.Values{}
	userID := params.UserID
	if userID == "" {
		userID = params.BroadcasterID
	}
	query.Set("user_id", userID)
	addCursorParams(query, params.CursorParams)

	var resp GetUserBlockListResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/users/blocks",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// Block prevents the specified user from interacting with the broadcaster.
func (s *UsersService) Block(ctx context.Context, params BlockUserParams) (*Response, error) {
	query := url.Values{}
	query.Set("target_user_id", params.TargetUserID)
	if params.SourceContext != "" {
		query.Set("source_context", params.SourceContext)
	}
	if params.Reason != "" {
		query.Set("reason", params.Reason)
	}

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPut,
		Path:   "/users/blocks",
		Query:  query,
	}, nil)
}

// Unblock removes the specified user from the broadcaster's blocked users.
func (s *UsersService) Unblock(ctx context.Context, params UnblockUserParams) (*Response, error) {
	query := url.Values{}
	query.Set("target_user_id", params.TargetUserID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/users/blocks",
		Query:  query,
	}, nil)
}
