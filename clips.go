package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// ClipsService provides access to the clips API.
type ClipsService struct {
	client *Client
}

// GetClipsParams filters Get Clips requests.
type GetClipsParams struct {
	CursorParams
	IDs           []string
	BroadcasterID string
	GameID        string
	StartedAt     *time.Time
	EndedAt       *time.Time
	IsFeatured    *bool
}

// Clip describes a Twitch clip.
type Clip struct {
	ID              string    `json:"id"`
	URL             string    `json:"url"`
	EmbedURL        string    `json:"embed_url"`
	BroadcasterID   string    `json:"broadcaster_id"`
	BroadcasterName string    `json:"broadcaster_name"`
	CreatorID       string    `json:"creator_id"`
	CreatorName     string    `json:"creator_name"`
	VideoID         string    `json:"video_id"`
	GameID          string    `json:"game_id"`
	Language        string    `json:"language"`
	Title           string    `json:"title"`
	ViewCount       int       `json:"view_count"`
	CreatedAt       time.Time `json:"created_at"`
	ThumbnailURL    string    `json:"thumbnail_url"`
	Duration        float64   `json:"duration"`
	VodOffset       *int      `json:"vod_offset"`
	IsFeatured      bool      `json:"is_featured"`
}

// GetClipsResponse is the typed response for Get Clips.
type GetClipsResponse struct {
	Data []Clip
}

// Get fetches clips by ID or filter.
func (s *ClipsService) Get(ctx context.Context, params GetClipsParams) (*GetClipsResponse, *Response, error) {
	query := url.Values{}
	addCursorParams(query, params.CursorParams)
	addRepeated(query, "id", params.IDs)
	if params.BroadcasterID != "" {
		query.Set("broadcaster_id", params.BroadcasterID)
	}
	if params.GameID != "" {
		query.Set("game_id", params.GameID)
	}
	if params.StartedAt != nil {
		query.Set("started_at", params.StartedAt.UTC().Format(time.RFC3339))
	}
	if params.EndedAt != nil {
		query.Set("ended_at", params.EndedAt.UTC().Format(time.RFC3339))
	}
	if params.IsFeatured != nil {
		if *params.IsFeatured {
			query.Set("is_featured", "true")
		} else {
			query.Set("is_featured", "false")
		}
	}

	var data []Clip
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/clips",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetClipsResponse{Data: data}, meta, nil
}
