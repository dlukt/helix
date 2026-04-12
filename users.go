package helix

import (
	"context"
	"net/http"
	"net/url"
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
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
}

// GetUsersResponse is the typed response for Get Users.
type GetUsersResponse struct {
	Data []User
}

// Get fetches users by id or login.
func (s *UsersService) Get(ctx context.Context, params GetUsersParams) (*GetUsersResponse, *Response, error) {
	query := url.Values{}
	for _, id := range params.IDs {
		query.Add("id", id)
	}
	for _, login := range params.Logins {
		query.Add("login", login)
	}

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
