package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// VideosService provides access to the videos API.
type VideosService struct {
	client *Client
}

// GetVideosParams filters Get Videos requests.
type GetVideosParams struct {
	CursorParams
	IDs      []string
	UserID   string
	GameID   string
	Language string
	Period   string
	Sort     string
	Type     string
}

// VideoMutedSegment describes a muted segment in a video.
type VideoMutedSegment struct {
	Duration int `json:"duration"`
	Offset   int `json:"offset"`
}

// Video describes a Twitch video.
type Video struct {
	ID            string              `json:"id"`
	StreamID      string              `json:"stream_id"`
	UserID        string              `json:"user_id"`
	UserLogin     string              `json:"user_login"`
	UserName      string              `json:"user_name"`
	Title         string              `json:"title"`
	Description   string              `json:"description"`
	CreatedAt     time.Time           `json:"created_at"`
	PublishedAt   time.Time           `json:"published_at"`
	URL           string              `json:"url"`
	ThumbnailURL  string              `json:"thumbnail_url"`
	Viewable      string              `json:"viewable"`
	ViewCount     int                 `json:"view_count"`
	Language      string              `json:"language"`
	Type          string              `json:"type"`
	Duration      string              `json:"duration"`
	MutedSegments []VideoMutedSegment `json:"muted_segments"`
}

// GetVideosResponse is the typed response for Get Videos.
type GetVideosResponse struct {
	Data []Video
}

// DeleteVideosParams identifies which videos to delete.
type DeleteVideosParams struct {
	IDs []string
}

// DeleteVideosResponse is the typed response for Delete Videos.
type DeleteVideosResponse struct {
	Data []string `json:"data"`
}

// Get fetches videos by ID or filter.
func (s *VideosService) Get(ctx context.Context, params GetVideosParams) (*GetVideosResponse, *Response, error) {
	if err := validateGetVideosParams(params); err != nil {
		return nil, nil, err
	}

	query := url.Values{}
	addCursorParams(query, params.CursorParams)
	addRepeated(query, "id", params.IDs)
	if params.UserID != "" {
		query.Set("user_id", params.UserID)
	}
	if params.GameID != "" {
		query.Set("game_id", params.GameID)
	}
	if params.Language != "" {
		query.Set("language", params.Language)
	}
	if params.Period != "" {
		query.Set("period", params.Period)
	}
	if params.Sort != "" {
		query.Set("sort", params.Sort)
	}
	if params.Type != "" {
		query.Set("type", params.Type)
	}

	var data []Video
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/videos",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetVideosResponse{Data: data}, meta, nil
}

// Delete deletes one or more videos.
func (s *VideosService) Delete(ctx context.Context, params DeleteVideosParams) (*DeleteVideosResponse, *Response, error) {
	if len(params.IDs) == 0 {
		return nil, nil, fmt.Errorf("helix: videos delete requires at least one id")
	}
	if len(params.IDs) > 5 {
		return nil, nil, fmt.Errorf("helix: videos delete supports at most 5 ids")
	}

	query := url.Values{}
	addRepeated(query, "id", params.IDs)

	var resp DeleteVideosResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/videos",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

func validateGetVideosParams(params GetVideosParams) error {
	hasIDs := len(params.IDs) > 0
	hasUserID := params.UserID != ""
	hasGameID := params.GameID != ""

	selectors := 0
	if hasIDs {
		selectors++
	}
	if hasUserID {
		selectors++
	}
	if hasGameID {
		selectors++
	}

	switch {
	case selectors == 0:
		return fmt.Errorf("helix: videos get requires exactly one of id, user_id, or game_id")
	case selectors > 1:
		return fmt.Errorf("helix: videos id, user_id, and game_id parameters are mutually exclusive")
	}

	if (params.Period != "" || params.Sort != "" || params.Type != "" || params.First > 0) && !(hasUserID || hasGameID) {
		return fmt.Errorf("helix: videos period, sort, type, and first filters require user_id or game_id")
	}
	if params.Language != "" && !hasGameID {
		return fmt.Errorf("helix: videos language filter requires game_id")
	}
	if (params.After != "" || params.Before != "") && !hasUserID {
		return fmt.Errorf("helix: videos after and before cursors require user_id")
	}

	return nil
}
