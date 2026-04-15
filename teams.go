package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// TeamsService provides access to Twitch team APIs.
type TeamsService struct {
	client *Client
}

// GetChannelTeamsParams identifies the broadcaster whose teams to fetch.
type GetChannelTeamsParams struct {
	BroadcasterID string
}

// GetTeamsParams identifies a team by ID or name.
type GetTeamsParams struct {
	ID   string
	Name string
}

// TeamUser describes a user in a Twitch team.
type TeamUser struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	UserLogin string `json:"user_login"`
}

// Team describes a Twitch team.
type Team struct {
	Users            []TeamUser `json:"users"`
	BackgroundImage  *string    `json:"background_image_url"`
	Banner           *string    `json:"banner"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Info             string     `json:"info"`
	ThumbnailURL     string     `json:"thumbnail_url"`
	TeamName         string     `json:"team_name"`
	TeamDisplayName  string     `json:"team_display_name"`
	ID               string     `json:"id"`
	BroadcasterID    string     `json:"broadcaster_id,omitempty"`
	BroadcasterLogin string     `json:"broadcaster_login,omitempty"`
	BroadcasterName  string     `json:"broadcaster_name,omitempty"`
}

// GetChannelTeamsResponse is the typed response for Get Channel Teams.
type GetChannelTeamsResponse struct {
	Data []Team `json:"data"`
}

// GetTeamsResponse is the typed response for Get Teams.
type GetTeamsResponse struct {
	Data []Team `json:"data"`
}

// GetChannel fetches teams that a broadcaster belongs to.
func (s *TeamsService) GetChannel(ctx context.Context, params GetChannelTeamsParams) (*GetChannelTeamsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)

	var resp GetChannelTeamsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/teams/channel",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// Get fetches a team by ID or name.
func (s *TeamsService) Get(ctx context.Context, params GetTeamsParams) (*GetTeamsResponse, *Response, error) {
	if err := validateGetTeamsParams(params); err != nil {
		return nil, nil, err
	}

	query := url.Values{}
	if params.ID != "" {
		query.Set("id", params.ID)
	}
	if params.Name != "" {
		query.Set("name", params.Name)
	}

	var resp GetTeamsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/teams",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

func validateGetTeamsParams(params GetTeamsParams) error {
	switch {
	case params.ID == "" && params.Name == "":
		return fmt.Errorf("helix: teams get requires exactly one of id or name")
	case params.ID != "" && params.Name != "":
		return fmt.Errorf("helix: teams id and name parameters are mutually exclusive")
	}

	return nil
}
