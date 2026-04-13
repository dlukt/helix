package helix

import (
	"context"
	"net/http"
	"net/url"
)

// ChatService provides access to Twitch chat APIs.
type ChatService struct {
	client *Client
}

// GetChatSettingsParams filters Get Chat Settings requests.
type GetChatSettingsParams struct {
	BroadcasterID string
	ModeratorID   string
}

// ChatSettings describes a broadcaster's chat room settings.
type ChatSettings struct {
	BroadcasterID                 string `json:"broadcaster_id"`
	ModeratorID                   string `json:"moderator_id"`
	EmoteMode                     bool   `json:"emote_mode"`
	SlowMode                      bool   `json:"slow_mode"`
	SlowModeWaitTime              *int   `json:"slow_mode_wait_time"`
	FollowerMode                  bool   `json:"follower_mode"`
	FollowerModeDuration          *int   `json:"follower_mode_duration"`
	SubscriberMode                bool   `json:"subscriber_mode"`
	UniqueChatMode                bool   `json:"unique_chat_mode"`
	NonModeratorChatDelay         *bool  `json:"non_moderator_chat_delay"`
	NonModeratorChatDelayDuration *int   `json:"non_moderator_chat_delay_duration"`
}

// GetChatSettingsResponse is the typed response for Get Chat Settings.
type GetChatSettingsResponse struct {
	Data []ChatSettings `json:"data"`
}

// GetChattersParams filters Get Chatters requests.
type GetChattersParams struct {
	CursorParams
	BroadcasterID string
	ModeratorID   string
}

// Chatter describes a user currently in chat.
type Chatter struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// GetChattersResponse is the typed response for Get Chatters.
type GetChattersResponse struct {
	Data       []Chatter  `json:"data"`
	Total      int        `json:"total"`
	Pagination Pagination `json:"pagination"`
}

// DeleteChatMessagesParams filters Delete Chat Messages requests.
type DeleteChatMessagesParams struct {
	BroadcasterID string
	ModeratorID   string
	MessageID     string
}

// GetSettings fetches the broadcaster's chat settings.
func (s *ChatService) GetSettings(ctx context.Context, params GetChatSettingsParams) (*GetChatSettingsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	if params.ModeratorID != "" {
		query.Set("moderator_id", params.ModeratorID)
	}

	var resp GetChatSettingsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/chat/settings",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetChatters fetches users currently connected to the broadcaster's chat room.
func (s *ChatService) GetChatters(ctx context.Context, params GetChattersParams) (*GetChattersResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	addCursorParams(query, params.CursorParams)

	var resp GetChattersResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/chat/chatters",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// DeleteChatMessages deletes a specific chat message or clears chat if MessageID is empty.
func (s *ChatService) DeleteChatMessages(ctx context.Context, params DeleteChatMessagesParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	if params.MessageID != "" {
		query.Set("message_id", params.MessageID)
	}

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/moderation/chat",
		Query:  query,
	}, nil)
}
