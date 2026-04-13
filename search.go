package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// SearchService provides access to the search API.
type SearchService struct {
	client *Client
}

// SearchCategoriesParams filters Search Categories requests.
type SearchCategoriesParams struct {
	CursorParams
	Query string
}

// SearchChannelsParams filters Search Channels requests.
type SearchChannelsParams struct {
	CursorParams
	Query    string
	LiveOnly bool
}

// CategorySearchResult describes a category search result.
type CategorySearchResult struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
	IGDBID    string `json:"igdb_id"`
}

// ChannelSearchResult describes a channel search result.
type ChannelSearchResult struct {
	ID               string     `json:"id"`
	BroadcasterLogin string     `json:"broadcaster_login"`
	DisplayName      string     `json:"display_name"`
	GameID           string     `json:"game_id"`
	GameName         string     `json:"game_name"`
	Title            string     `json:"title"`
	ThumbnailURL     string     `json:"thumbnail_url"`
	IsLive           bool       `json:"is_live"`
	StartedAt        *time.Time `json:"started_at"`
	Tags             []string   `json:"tags"`
}

// UnmarshalJSON accepts Twitch's empty started_at for offline channels.
func (r *ChannelSearchResult) UnmarshalJSON(data []byte) error {
	type channelSearchResult ChannelSearchResult
	type channelSearchResultWire struct {
		channelSearchResult
		StartedAt string `json:"started_at"`
	}

	var wire channelSearchResultWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	*r = ChannelSearchResult(wire.channelSearchResult)
	if wire.StartedAt == "" {
		r.StartedAt = nil
		return nil
	}

	startedAt, err := time.Parse(time.RFC3339, wire.StartedAt)
	if err != nil {
		return err
	}
	r.StartedAt = &startedAt
	return nil
}

// SearchCategoriesResponse is the typed response for Search Categories.
type SearchCategoriesResponse struct {
	Data []CategorySearchResult
}

// SearchChannelsResponse is the typed response for Search Channels.
type SearchChannelsResponse struct {
	Data []ChannelSearchResult
}

// Categories searches categories by query.
func (s *SearchService) Categories(ctx context.Context, params SearchCategoriesParams) (*SearchCategoriesResponse, *Response, error) {
	query := url.Values{}
	query.Set("query", params.Query)
	addCursorParams(query, params.CursorParams)

	var data []CategorySearchResult
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/search/categories",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &SearchCategoriesResponse{Data: data}, meta, nil
}

// Channels searches channels by query.
func (s *SearchService) Channels(ctx context.Context, params SearchChannelsParams) (*SearchChannelsResponse, *Response, error) {
	query := url.Values{}
	query.Set("query", params.Query)
	addCursorParams(query, params.CursorParams)
	if params.LiveOnly {
		query.Set("live_only", strconv.FormatBool(true))
	}

	var data []ChannelSearchResult
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/search/channels",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &SearchChannelsResponse{Data: data}, meta, nil
}
