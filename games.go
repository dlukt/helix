package helix

import (
	"context"
	"net/http"
	"net/url"
)

// GamesService provides access to the games API.
type GamesService struct {
	client *Client
}

// Game describes a Twitch game or category.
type Game struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
	IGDBID    string `json:"igdb_id"`
}

// GetGamesParams filters Get Games requests.
type GetGamesParams struct {
	IDs     []string
	Names   []string
	IGDBIDs []string
}

// GetGamesResponse is the typed response for Get Games.
type GetGamesResponse struct {
	Data []Game
}

// GetTopGamesParams filters Get Top Games requests.
type GetTopGamesParams struct {
	CursorParams
}

// GetTopGamesResponse is the typed response for Get Top Games.
type GetTopGamesResponse struct {
	Data []Game
}

// Get fetches games by ID, name, or IGDB ID.
func (s *GamesService) Get(ctx context.Context, params GetGamesParams) (*GetGamesResponse, *Response, error) {
	query := url.Values{}
	addRepeated(query, "id", params.IDs)
	addRepeated(query, "name", params.Names)
	addRepeated(query, "igdb_id", params.IGDBIDs)

	var data []Game
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/games",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetGamesResponse{Data: data}, meta, nil
}

// Top fetches the most-viewed games.
func (s *GamesService) Top(ctx context.Context, params GetTopGamesParams) (*GetTopGamesResponse, *Response, error) {
	query := url.Values{}
	addCursorParams(query, params.CursorParams)

	var data []Game
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/games/top",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetTopGamesResponse{Data: data}, meta, nil
}
