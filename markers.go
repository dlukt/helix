package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// MarkersService provides access to stream marker APIs.
type MarkersService struct {
	client *Client
}

// StreamMarker describes a single marker.
type StreamMarker struct {
	ID              string    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	Description     string    `json:"description"`
	PositionSeconds int       `json:"position_seconds"`
	URL             string    `json:"url"`
}

// MarkerVideo groups markers by video.
type MarkerVideo struct {
	VideoID string         `json:"video_id"`
	Markers []StreamMarker `json:"markers"`
}

// MarkerGroup groups markers by user.
type MarkerGroup struct {
	UserID    string        `json:"user_id"`
	UserName  string        `json:"user_name"`
	UserLogin string        `json:"user_login"`
	Videos    []MarkerVideo `json:"videos"`
}

// GetMarkersParams filters Get Stream Markers requests.
type GetMarkersParams struct {
	CursorParams
	UserID  string
	VideoID string
}

// GetMarkersResponse is the typed response for Get Stream Markers.
type GetMarkersResponse struct {
	Data []MarkerGroup
}

// CreateMarkerRequest creates a stream marker.
type CreateMarkerRequest struct {
	UserID      string `json:"user_id"`
	Description string `json:"description,omitempty"`
}

// CreateMarkerResponse is the typed response for Create Stream Marker.
type CreateMarkerResponse struct {
	Data []StreamMarker
}

// Get fetches markers from a broadcaster's latest stream or a specific video.
func (s *MarkersService) Get(ctx context.Context, params GetMarkersParams) (*GetMarkersResponse, *Response, error) {
	if err := validateGetMarkersParams(params); err != nil {
		return nil, nil, err
	}

	query := url.Values{}
	addCursorParams(query, params.CursorParams)
	if params.UserID != "" {
		query.Set("user_id", params.UserID)
	}
	if params.VideoID != "" {
		query.Set("video_id", params.VideoID)
	}

	var data []MarkerGroup
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/streams/markers",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetMarkersResponse{Data: data}, meta, nil
}

func validateGetMarkersParams(params GetMarkersParams) error {
	switch {
	case params.UserID == "" && params.VideoID == "":
		return fmt.Errorf("helix: markers get requires exactly one of user_id or video_id")
	case params.UserID != "" && params.VideoID != "":
		return fmt.Errorf("helix: markers user_id and video_id parameters are mutually exclusive")
	}

	return nil
}

// Create adds a marker to the current live stream.
func (s *MarkersService) Create(ctx context.Context, req CreateMarkerRequest) (*CreateMarkerResponse, *Response, error) {
	var data []StreamMarker
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/streams/markers",
		Body:   req,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &CreateMarkerResponse{Data: data}, meta, nil
}
