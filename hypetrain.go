package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// HypeTrainService provides access to Twitch hype train APIs.
type HypeTrainService struct {
	client *Client
}

// GetHypeTrainStatusParams identifies the broadcaster whose hype train status to fetch.
type GetHypeTrainStatusParams struct {
	BroadcasterID string
}

// HypeTrainContribution describes one contribution in a hype train status response.
type HypeTrainContribution struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Type      string `json:"type"`
	Total     int    `json:"total"`
}

// HypeTrainParticipant describes a broadcaster participating in a shared train.
type HypeTrainParticipant struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

// HypeTrainCurrent describes the currently active hype train, if any.
type HypeTrainCurrent struct {
	ID                      string                  `json:"id"`
	BroadcasterUserID       string                  `json:"broadcaster_user_id"`
	BroadcasterUserLogin    string                  `json:"broadcaster_user_login"`
	BroadcasterUserName     string                  `json:"broadcaster_user_name"`
	Total                   int                     `json:"total"`
	Progress                int                     `json:"progress"`
	Goal                    int                     `json:"goal"`
	TopContributions        []HypeTrainContribution `json:"top_contributions"`
	SharedTrainParticipants []HypeTrainParticipant  `json:"shared_train_participants"`
	Level                   int                     `json:"level"`
	StartedAt               time.Time               `json:"started_at"`
	ExpiresAt               time.Time               `json:"expires_at"`
	IsSharedTrain           bool                    `json:"is_shared_train"`
	Type                    string                  `json:"type"`
}

// HypeTrainRecord describes an all-time-high record in the hype train response.
type HypeTrainRecord struct {
	Level      int       `json:"level"`
	Total      int       `json:"total"`
	AchievedAt time.Time `json:"achieved_at"`
}

// HypeTrainStatus describes the broadcaster's current and historical hype train state.
type HypeTrainStatus struct {
	Current           *HypeTrainCurrent `json:"current"`
	AllTimeHigh       *HypeTrainRecord  `json:"all_time_high"`
	SharedAllTimeHigh *HypeTrainRecord  `json:"shared_all_time_high"`
}

// GetHypeTrainStatusResponse is the typed response for Get Hype Train Status.
type GetHypeTrainStatusResponse struct {
	Data []HypeTrainStatus `json:"data"`
}

// GetStatus fetches the broadcaster's hype train status.
func (s *HypeTrainService) GetStatus(ctx context.Context, params GetHypeTrainStatusParams) (*GetHypeTrainStatusResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)

	var resp GetHypeTrainStatusResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/hypetrain/status",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
