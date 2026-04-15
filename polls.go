package helix

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// PollsService provides access to Twitch poll APIs.
type PollsService struct {
	client *Client
}

// GetPollsParams filters Get Polls requests.
type GetPollsParams struct {
	CursorParams
	BroadcasterID string
	IDs           []string
}

// PollChoice describes one choice in a poll.
type PollChoice struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Votes              int    `json:"votes"`
	ChannelPointsVotes int    `json:"channel_points_votes"`
	BitsVotes          int    `json:"bits_votes"`
}

// Poll describes a Twitch poll.
type Poll struct {
	ID                         string       `json:"id"`
	BroadcasterID              string       `json:"broadcaster_id"`
	BroadcasterName            string       `json:"broadcaster_name"`
	BroadcasterLogin           string       `json:"broadcaster_login"`
	Title                      string       `json:"title"`
	Choices                    []PollChoice `json:"choices"`
	BitsVotingEnabled          bool         `json:"bits_voting_enabled"`
	BitsPerVote                int          `json:"bits_per_vote"`
	ChannelPointsVotingEnabled bool         `json:"channel_points_voting_enabled"`
	ChannelPointsPerVote       int          `json:"channel_points_per_vote"`
	Status                     string       `json:"status"`
	Duration                   int          `json:"duration"`
	StartedAt                  time.Time    `json:"started_at"`
	EndedAt                    *time.Time   `json:"ended_at"`
}

// GetPollsResponse is the typed response for Get Polls.
type GetPollsResponse struct {
	Data       []Poll     `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// CreatePollChoice identifies a choice when creating a poll.
type CreatePollChoice struct {
	Title string `json:"title"`
}

// CreatePollRequest is the request body for Create Poll.
type CreatePollRequest struct {
	BroadcasterID              string             `json:"broadcaster_id"`
	Title                      string             `json:"title"`
	Choices                    []CreatePollChoice `json:"choices"`
	Duration                   int                `json:"duration"`
	ChannelPointsVotingEnabled *bool              `json:"channel_points_voting_enabled,omitempty"`
	ChannelPointsPerVote       *int               `json:"channel_points_per_vote,omitempty"`
}

// CreatePollResponse is the typed response for Create Poll.
type CreatePollResponse struct {
	Data []Poll `json:"data"`
}

// EndPollRequest is the request body for End Poll.
type EndPollRequest struct {
	BroadcasterID string `json:"broadcaster_id"`
	ID            string `json:"id"`
	Status        string `json:"status"`
}

// EndPollResponse is the typed response for End Poll.
type EndPollResponse struct {
	Data []Poll `json:"data"`
}

// Get fetches one or more polls for a broadcaster.
func (s *PollsService) Get(ctx context.Context, params GetPollsParams) (*GetPollsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	addRepeated(query, "id", params.IDs)
	addCursorParams(query, params.CursorParams)

	var resp GetPollsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/polls",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// Create starts a new poll for a broadcaster.
func (s *PollsService) Create(ctx context.Context, req CreatePollRequest) (*CreatePollResponse, *Response, error) {
	var resp CreatePollResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/polls",
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// End terminates an active poll early and returns its final state.
func (s *PollsService) End(ctx context.Context, req EndPollRequest) (*EndPollResponse, *Response, error) {
	var resp EndPollResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/polls",
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}
