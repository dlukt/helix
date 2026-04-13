package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

// CreateClipParams identifies the broadcaster and optional metadata when creating a live clip.
type CreateClipParams struct {
	BroadcasterID string
	Duration      *float64
	Title         string
}

// CreateClipFromVODParams identifies the clip source and metadata when creating a clip from a VOD.
type CreateClipFromVODParams struct {
	EditorID      string
	BroadcasterID string
	VODID         string
	VODOffset     int
	Duration      *float64
	Title         string
}

// GetClipsDownloadParams filters Get Clips Download requests.
type GetClipsDownloadParams struct {
	EditorID      string
	BroadcasterID string
	ClipIDs       []string
}

// CreatedClip describes a newly created clip and its edit URL.
type CreatedClip struct {
	ID      string `json:"id"`
	EditURL string `json:"edit_url"`
}

// CreateClipResponse is the typed response for Create Clip and Create Clip From VOD.
type CreateClipResponse struct {
	Data []CreatedClip `json:"data"`
}

// ClipDownload describes the downloadable media URLs for a clip.
type ClipDownload struct {
	ClipID               string  `json:"clip_id"`
	LandscapeDownloadURL *string `json:"landscape_download_url"`
	PortraitDownloadURL  *string `json:"portrait_download_url"`
}

// GetClipsDownloadResponse is the typed response for Get Clips Download.
type GetClipsDownloadResponse struct {
	Data []ClipDownload `json:"data"`
}

// Get fetches clips by ID or filter.
func (s *ClipsService) Get(ctx context.Context, params GetClipsParams) (*GetClipsResponse, *Response, error) {
	if err := validateGetClipsParams(params); err != nil {
		return nil, nil, err
	}

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

// Create starts creating a clip from the broadcaster's live stream.
func (s *ClipsService) Create(ctx context.Context, params CreateClipParams) (*CreateClipResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	if params.Duration != nil {
		query.Set("duration", trimFloat(*params.Duration))
	}
	if params.Title != "" {
		query.Set("title", params.Title)
	}

	var resp CreateClipResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/clips",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// CreateFromVOD starts creating a clip from a broadcaster's VOD.
func (s *ClipsService) CreateFromVOD(ctx context.Context, params CreateClipFromVODParams) (*CreateClipResponse, *Response, error) {
	if err := validateCreateClipFromVODParams(params); err != nil {
		return nil, nil, err
	}

	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("vod_id", params.VODID)
	query.Set("vod_offset", trimFloat(float64(params.VODOffset)))
	query.Set("editor_id", params.EditorID)
	query.Set("title", params.Title)
	if params.Duration != nil {
		query.Set("duration", trimFloat(*params.Duration))
	}

	var resp CreateClipResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/videos/clips",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetDownloads fetches downloadable media URLs for one or more clips.
func (s *ClipsService) GetDownloads(ctx context.Context, params GetClipsDownloadParams) (*GetClipsDownloadResponse, *Response, error) {
	query := url.Values{}
	query.Set("editor_id", params.EditorID)
	query.Set("broadcaster_id", params.BroadcasterID)
	for _, clipID := range params.ClipIDs {
		query.Add("clip_id", clipID)
	}

	var resp GetClipsDownloadResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/clips/download",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

func trimFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func validateGetClipsParams(params GetClipsParams) error {
	selectors := 0
	hasIDs := len(params.IDs) > 0
	if len(params.IDs) > 0 {
		selectors++
	}
	if params.BroadcasterID != "" {
		selectors++
	}
	if params.GameID != "" {
		selectors++
	}

	switch {
	case selectors == 0:
		return fmt.Errorf("helix: clips get requires exactly one of id, broadcaster_id, or game_id")
	case selectors > 1:
		return fmt.Errorf("helix: clips id, broadcaster_id, and game_id parameters are mutually exclusive")
	}
	if hasIDs && (params.After != "" || params.Before != "" || params.First > 0) {
		return fmt.Errorf("helix: clips pagination parameters require broadcaster_id or game_id")
	}

	return nil
}

func validateCreateClipFromVODParams(params CreateClipFromVODParams) error {
	switch {
	case params.EditorID == "":
		return fmt.Errorf("helix: clips create from vod requires editor_id")
	case params.Title == "":
		return fmt.Errorf("helix: clips create from vod requires title")
	}

	return nil
}
