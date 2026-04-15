package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
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

// GetScheduleICalendarParams identifies the broadcaster whose schedule to export as iCalendar.
type GetScheduleICalendarParams struct {
	BroadcasterID string
}

// UpdateScheduleParams identifies the broadcaster schedule settings to update.
type UpdateScheduleParams struct {
	BroadcasterID     string
	IsVacationEnabled *bool
	VacationStartTime *time.Time
	VacationEndTime   *time.Time
	Timezone          string
}

// CreateScheduleSegmentParams identifies the broadcaster whose schedule segment should be created.
type CreateScheduleSegmentParams struct {
	BroadcasterID string
}

// CreateScheduleSegmentRequest contains the data needed to create a schedule segment.
type CreateScheduleSegmentRequest struct {
	StartTime   time.Time `json:"start_time"`
	Timezone    string    `json:"timezone"`
	IsRecurring bool      `json:"is_recurring"`
	Duration    int       `json:"duration"`
	CategoryID  string    `json:"category_id,omitempty"`
	Title       string    `json:"title,omitempty"`
}

// MarshalJSON encodes duration using Twitch's documented string representation.
func (r CreateScheduleSegmentRequest) MarshalJSON() ([]byte, error) {
	body := struct {
		StartTime   time.Time `json:"start_time"`
		Timezone    string    `json:"timezone"`
		IsRecurring bool      `json:"is_recurring"`
		Duration    string    `json:"duration"`
		CategoryID  string    `json:"category_id,omitempty"`
		Title       string    `json:"title,omitempty"`
	}{
		StartTime:   r.StartTime,
		Timezone:    r.Timezone,
		IsRecurring: r.IsRecurring,
		Duration:    strconv.Itoa(r.Duration),
		CategoryID:  r.CategoryID,
		Title:       r.Title,
	}
	return json.Marshal(body)
}

// UpdateScheduleSegmentParams identifies the schedule segment to update.
type UpdateScheduleSegmentParams struct {
	BroadcasterID string
	ID            string
}

// UpdateScheduleSegmentRequest contains the mutable fields for a schedule segment.
type UpdateScheduleSegmentRequest struct {
	StartTime  *time.Time `json:"start_time,omitempty"`
	Duration   *int       `json:"duration,omitempty"`
	CategoryID string     `json:"category_id,omitempty"`
	Title      string     `json:"title,omitempty"`
	IsCanceled *bool      `json:"is_canceled,omitempty"`
	Timezone   string     `json:"timezone,omitempty"`
}

// MarshalJSON encodes duration using Twitch's documented string representation.
func (r UpdateScheduleSegmentRequest) MarshalJSON() ([]byte, error) {
	body := struct {
		StartTime  *time.Time `json:"start_time,omitempty"`
		Duration   *string    `json:"duration,omitempty"`
		CategoryID string     `json:"category_id,omitempty"`
		Title      string     `json:"title,omitempty"`
		IsCanceled *bool      `json:"is_canceled,omitempty"`
		Timezone   string     `json:"timezone,omitempty"`
	}{
		StartTime:  r.StartTime,
		CategoryID: r.CategoryID,
		Title:      r.Title,
		IsCanceled: r.IsCanceled,
		Timezone:   r.Timezone,
	}
	if r.Duration != nil {
		duration := strconv.Itoa(*r.Duration)
		body.Duration = &duration
	}
	return json.Marshal(body)
}

// DeleteScheduleSegmentParams identifies the schedule segment to delete.
type DeleteScheduleSegmentParams struct {
	BroadcasterID string
	ID            string
}

// GetScheduleResponse is the typed response for Get Channel Stream Schedule.
type GetScheduleResponse struct {
	Data Schedule
}

// GetScheduleICalendarResponse contains the raw iCalendar representation of a broadcaster's schedule.
type GetScheduleICalendarResponse struct {
	Calendar string
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

// GetICalendar fetches a broadcaster's stream schedule in iCalendar format.
func (s *ScheduleService) GetICalendar(ctx context.Context, params GetScheduleICalendarParams) (*GetScheduleICalendarResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)

	httpResp, meta, err := s.client.doRaw(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/schedule/icalendar",
		Query:  query,
	})
	if err != nil {
		return nil, meta, err
	}
	body, err := replayableBody(httpResp)
	if err != nil {
		return nil, meta, err
	}
	return &GetScheduleICalendarResponse{Calendar: string(body)}, meta, nil
}

// UpdateSettings updates a broadcaster's schedule settings, such as vacation mode.
func (s *ScheduleService) UpdateSettings(ctx context.Context, params UpdateScheduleParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	if params.IsVacationEnabled != nil {
		query.Set("is_vacation_enabled", strconv.FormatBool(*params.IsVacationEnabled))
	}
	if params.VacationStartTime != nil {
		query.Set("vacation_start_time", params.VacationStartTime.Format(time.RFC3339))
	}
	if params.VacationEndTime != nil {
		query.Set("vacation_end_time", params.VacationEndTime.Format(time.RFC3339))
	}
	if params.Timezone != "" {
		query.Set("timezone", params.Timezone)
	}

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/schedule/settings",
		Query:  query,
	}, nil)
}

// CreateSegment adds a schedule segment to the broadcaster's stream schedule.
func (s *ScheduleService) CreateSegment(ctx context.Context, params CreateScheduleSegmentParams, req CreateScheduleSegmentRequest) (*GetScheduleResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)

	var data Schedule
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/schedule/segment",
		Query:  query,
		Body:   req,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetScheduleResponse{Data: data}, meta, nil
}

// UpdateSegment updates a schedule segment in the broadcaster's stream schedule.
func (s *ScheduleService) UpdateSegment(ctx context.Context, params UpdateScheduleSegmentParams, req UpdateScheduleSegmentRequest) (*GetScheduleResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("id", params.ID)

	var data Schedule
	meta, err := s.client.doData(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/schedule/segment",
		Query:  query,
		Body:   req,
	}, &data)
	if err != nil {
		return nil, meta, err
	}
	return &GetScheduleResponse{Data: data}, meta, nil
}

// DeleteSegment removes a schedule segment from the broadcaster's stream schedule.
func (s *ScheduleService) DeleteSegment(ctx context.Context, params DeleteScheduleSegmentParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("id", params.ID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/schedule/segment",
		Query:  query,
	}, nil)
}
