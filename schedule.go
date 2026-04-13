package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// ScheduleService provides access to the channel schedule API.
type ScheduleService struct {
	client *Client
}

// ScheduleSegmentCategory identifies a schedule segment's category.
type ScheduleSegmentCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ScheduleSegment describes a scheduled broadcast segment.
type ScheduleSegment struct {
	ID            string                   `json:"id"`
	StartTime     time.Time                `json:"start_time"`
	EndTime       time.Time                `json:"end_time"`
	Title         string                   `json:"title"`
	CanceledUntil *time.Time               `json:"canceled_until"`
	Category      *ScheduleSegmentCategory `json:"category"`
	IsRecurring   bool                     `json:"is_recurring"`
}

// ScheduleVacation describes a broadcaster vacation window.
type ScheduleVacation struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// Schedule contains a broadcaster's schedule.
type Schedule struct {
	Segments         []ScheduleSegment `json:"segments"`
	BroadcasterID    string            `json:"broadcaster_id"`
	BroadcasterName  string            `json:"broadcaster_name"`
	BroadcasterLogin string            `json:"broadcaster_login"`
	Vacation         *ScheduleVacation `json:"vacation"`
}

// GetScheduleParams filters Get Channel Stream Schedule requests.
type GetScheduleParams struct {
	CursorParams
	BroadcasterID string
	IDs           []string
	StartTime     *time.Time
}

// GetScheduleResponse is the typed response for Get Channel Stream Schedule.
type GetScheduleResponse struct {
	Data Schedule
}

// Get fetches a broadcaster's stream schedule.
func (s *ScheduleService) Get(ctx context.Context, params GetScheduleParams) (*GetScheduleResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	addCursorParams(query, params.CursorParams)
	addRepeated(query, "id", params.IDs)
	if params.StartTime != nil {
		query.Set("start_time", params.StartTime.UTC().Format(time.RFC3339))
	}

	var data Schedule
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/schedule",
		Query:  query,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetScheduleResponse{Data: data}, meta, nil
}
