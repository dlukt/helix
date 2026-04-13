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

// UpdateChatSettingsParams identifies which broadcaster chat settings to update.
type UpdateChatSettingsParams struct {
	BroadcasterID string
	ModeratorID   string
}

// UpdateChatSettingsRequest contains the chat settings to update.
type UpdateChatSettingsRequest struct {
	EmoteMode                     *bool `json:"emote_mode,omitempty"`
	SlowMode                      *bool `json:"slow_mode,omitempty"`
	SlowModeWaitTime              *int  `json:"slow_mode_wait_time,omitempty"`
	FollowerMode                  *bool `json:"follower_mode,omitempty"`
	FollowerModeDuration          *int  `json:"follower_mode_duration,omitempty"`
	SubscriberMode                *bool `json:"subscriber_mode,omitempty"`
	UniqueChatMode                *bool `json:"unique_chat_mode,omitempty"`
	NonModeratorChatDelay         *bool `json:"non_moderator_chat_delay,omitempty"`
	NonModeratorChatDelayDuration *int  `json:"non_moderator_chat_delay_duration,omitempty"`
}

// UpdateChatSettingsResponse is the typed response for Update Chat Settings.
type UpdateChatSettingsResponse struct {
	Data []ChatSettings `json:"data"`
}

// SendAnnouncementParams identifies where to send a chat announcement.
type SendAnnouncementParams struct {
	BroadcasterID string
	ModeratorID   string
}

// SendAnnouncementRequest is the request body for Send Chat Announcement.
type SendAnnouncementRequest struct {
	Message       string `json:"message"`
	Color         string `json:"color,omitempty"`
	ForSourceOnly *bool  `json:"for_source_only,omitempty"`
}

// DeleteChatMessagesParams filters Delete Chat Messages requests.
type DeleteChatMessagesParams struct {
	BroadcasterID string
	ModeratorID   string
	MessageID     string
}

// SendShoutoutParams identifies the broadcaster and moderator for a shoutout.
type SendShoutoutParams struct {
	FromBroadcasterID string
	ToBroadcasterID   string
	ModeratorID       string
}

// SendMessageRequest is the request body for Send Chat Message.
type SendMessageRequest struct {
	BroadcasterID        string `json:"broadcaster_id"`
	SenderID             string `json:"sender_id"`
	Message              string `json:"message"`
	ReplyParentMessageID string `json:"reply_parent_message_id,omitempty"`
	ForSourceOnly        *bool  `json:"for_source_only,omitempty"`
}

// SendMessageDropReason describes why a chat message was dropped.
type SendMessageDropReason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SendMessageResult describes the outcome of attempting to send a chat message.
type SendMessageResult struct {
	MessageID  string                 `json:"message_id"`
	IsSent     bool                   `json:"is_sent"`
	DropReason *SendMessageDropReason `json:"drop_reason"`
}

// SendMessageResponse is the typed response for Send Chat Message.
type SendMessageResponse struct {
	Data []SendMessageResult `json:"data"`
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

// UpdateSettings updates the broadcaster's chat settings.
func (s *ChatService) UpdateSettings(ctx context.Context, params UpdateChatSettingsParams, req UpdateChatSettingsRequest) (*UpdateChatSettingsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	var resp UpdateChatSettingsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/chat/settings",
		Query:  query,
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// SendAnnouncement sends an announcement to the broadcaster's chat room.
func (s *ChatService) SendAnnouncement(ctx context.Context, params SendAnnouncementParams, req SendAnnouncementRequest) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/chat/announcements",
		Query:  query,
		Body:   req,
	}, nil)
}

// SendShoutout sends a shoutout from one broadcaster to another.
func (s *ChatService) SendShoutout(ctx context.Context, params SendShoutoutParams) (*Response, error) {
	query := url.Values{}
	query.Set("from_broadcaster_id", params.FromBroadcasterID)
	query.Set("to_broadcaster_id", params.ToBroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/chat/shoutouts",
		Query:  query,
	}, nil)
}

// SendMessage sends a chat message to the broadcaster's chat room.
func (s *ChatService) SendMessage(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, *Response, error) {
	var resp SendMessageResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/chat/messages",
		Body:   req,
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
