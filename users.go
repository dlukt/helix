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
