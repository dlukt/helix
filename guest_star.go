package helix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GuestStarService provides access to Twitch Guest Star APIs.
type GuestStarService struct {
	client *Client
}

// GetGuestStarChannelSettingsParams identifies the host and moderator for reading channel settings.
type GetGuestStarChannelSettingsParams struct {
	BroadcasterID string
	ModeratorID   string
}

// UpdateGuestStarChannelSettingsParams identifies the broadcaster whose Guest Star settings are updated.
type UpdateGuestStarChannelSettingsParams struct {
	BroadcasterID                string
	IsModeratorSendLiveEnabled   *bool
	SlotCount                    *int
	IsBrowserSourceAudioEnabled  *bool
	GroupLayout                  string
	RegenerateBrowserSourceToken *bool
}

// GetGuestStarSessionParams identifies the host and moderator for reading the active session.
type GetGuestStarSessionParams struct {
	BroadcasterID string
	ModeratorID   string
}

// EndGuestStarSessionParams identifies the broadcaster session to end.
type EndGuestStarSessionParams struct {
	BroadcasterID string
	SessionID     string
}

// GuestStarInvitesParams identifies the session whose invites are queried.
type GuestStarInvitesParams struct {
	BroadcasterID string
	ModeratorID   string
	SessionID     string
}

// SendGuestStarInviteParams identifies the guest invite to create.
type SendGuestStarInviteParams struct {
	BroadcasterID string
	ModeratorID   string
	SessionID     string
	GuestID       string
}

// DeleteGuestStarInviteParams identifies the guest invite to revoke.
type DeleteGuestStarInviteParams struct {
	BroadcasterID string
	ModeratorID   string
	SessionID     string
	GuestID       string
}

// AssignGuestStarSlotParams identifies the guest slot assignment to create.
type AssignGuestStarSlotParams struct {
	BroadcasterID string
	ModeratorID   string
	SessionID     string
	GuestID       string
	SlotID        string
}

// UpdateGuestStarSlotParams identifies the slot assignment move to perform.
type UpdateGuestStarSlotParams struct {
	BroadcasterID     string
	ModeratorID       string
	SessionID         string
	SourceSlotID      string
	DestinationSlotID string
}

// DeleteGuestStarSlotParams identifies the guest slot assignment to remove.
type DeleteGuestStarSlotParams struct {
	BroadcasterID       string
	ModeratorID         string
	SessionID           string
	GuestID             string
	SlotID              string
	ShouldReinviteGuest *bool
}

// UpdateGuestStarSlotSettingsParams identifies the slot settings to mutate.
type UpdateGuestStarSlotSettingsParams struct {
	BroadcasterID  string
	ModeratorID    string
	SessionID      string
	SlotID         string
	IsAudioEnabled *bool
	IsVideoEnabled *bool
	IsLive         *bool
	Volume         *int
}

// GuestStarChannelSettings describes the Guest Star settings for a broadcaster.
type GuestStarChannelSettings struct {
	IsModeratorSendLiveEnabled  bool   `json:"is_moderator_send_live_enabled"`
	SlotCount                   int    `json:"slot_count"`
	IsBrowserSourceAudioEnabled bool   `json:"is_browser_source_audio_enabled"`
	GroupLayout                 string `json:"group_layout"`
	BrowserSourceToken          string `json:"browser_source_token"`
}

// UnmarshalJSON accepts both documented group_layout and example layout field names.
func (s *GuestStarChannelSettings) UnmarshalJSON(data []byte) error {
	type alias GuestStarChannelSettings
	aux := struct {
		alias
		Layout string `json:"layout"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*s = GuestStarChannelSettings(aux.alias)
	if s.GroupLayout == "" {
		s.GroupLayout = aux.Layout
	}
	return nil
}

// GetGuestStarChannelSettingsResponse is the typed response for Get Channel Guest Star Settings.
type GetGuestStarChannelSettingsResponse struct {
	Data []GuestStarChannelSettings `json:"data"`
}

// GuestStarMediaSettings describes whether a guest's audio or video is available and enabled.
type GuestStarMediaSettings struct {
	IsHostEnabled  bool `json:"is_host_enabled"`
	IsGuestEnabled bool `json:"is_guest_enabled"`
	IsAvailable    bool `json:"is_available"`
}

// GuestStarSlot describes a guest currently participating in a session.
type GuestStarSlot struct {
	SlotID          string                 `json:"slot_id"`
	IsLive          bool                   `json:"is_live"`
	UserID          string                 `json:"user_id"`
	UserDisplayName string                 `json:"user_display_name"`
	UserLogin       string                 `json:"user_login"`
	Volume          int                    `json:"volume"`
	AssignedAt      time.Time              `json:"assigned_at"`
	AudioSettings   GuestStarMediaSettings `json:"audio_settings"`
	VideoSettings   GuestStarMediaSettings `json:"video_settings"`
}

// UnmarshalJSON accepts both slot_id and the older id field from Twitch examples.
func (s *GuestStarSlot) UnmarshalJSON(data []byte) error {
	type alias GuestStarSlot
	aux := struct {
		alias
		ID string `json:"id"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*s = GuestStarSlot(aux.alias)
	if s.SlotID == "" {
		s.SlotID = aux.ID
	}
	return nil
}

// GuestStarSession describes an active Guest Star session.
type GuestStarSession struct {
	ID     string          `json:"id"`
	Guests []GuestStarSlot `json:"guests"`
}

// GetGuestStarSessionResponse is the typed response for session endpoints.
type GetGuestStarSessionResponse struct {
	Data []GuestStarSession `json:"data"`
}

// GuestStarInvite describes an invite and the guest's waiting-room readiness.
type GuestStarInvite struct {
	UserID           string    `json:"user_id"`
	InvitedAt        time.Time `json:"invited_at"`
	Status           string    `json:"status"`
	IsAudioEnabled   bool      `json:"is_audio_enabled"`
	IsVideoEnabled   bool      `json:"is_video_enabled"`
	IsAudioAvailable bool      `json:"is_audio_available"`
	IsVideoAvailable bool      `json:"is_video_available"`
}

// GetGuestStarInvitesResponse is the typed response for Get Guest Star Invites.
type GetGuestStarInvitesResponse struct {
	Data []GuestStarInvite `json:"data"`
}

// AssignGuestStarSlotResult describes a non-empty Assign Guest Star Slot response.
type AssignGuestStarSlotResult struct {
	Code string `json:"code"`
}

// AssignGuestStarSlotResponse is the typed response for Assign Guest Star Slot.
type AssignGuestStarSlotResponse struct {
	Data AssignGuestStarSlotResult `json:"data"`
}

// GetChannelSettings fetches Guest Star settings for a broadcaster.
func (s *GuestStarService) GetChannelSettings(ctx context.Context, params GetGuestStarChannelSettingsParams) (*GetGuestStarChannelSettingsResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	var resp GetGuestStarChannelSettingsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/guest_star/channel_settings",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// UpdateChannelSettings updates Guest Star settings for a broadcaster.
func (s *GuestStarService) UpdateChannelSettings(ctx context.Context, params UpdateGuestStarChannelSettingsParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	if params.IsModeratorSendLiveEnabled != nil {
		query.Set("is_moderator_send_live_enabled", strconv.FormatBool(*params.IsModeratorSendLiveEnabled))
	}
	if params.SlotCount != nil {
		query.Set("slot_count", strconv.Itoa(*params.SlotCount))
	}
	if params.IsBrowserSourceAudioEnabled != nil {
		query.Set("is_browser_source_audio_enabled", strconv.FormatBool(*params.IsBrowserSourceAudioEnabled))
	}
	if params.GroupLayout != "" {
		query.Set("group_layout", params.GroupLayout)
	}
	if params.RegenerateBrowserSourceToken != nil {
		query.Set("regenerate_browser_sources", strconv.FormatBool(*params.RegenerateBrowserSourceToken))
	}

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPut,
		Path:   "/guest_star/channel_settings",
		Query:  query,
	}, nil)
}

// GetSession fetches the active Guest Star session for a broadcaster.
func (s *GuestStarService) GetSession(ctx context.Context, params GetGuestStarSessionParams) (*GetGuestStarSessionResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)

	var resp GetGuestStarSessionResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/guest_star/session",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// CreateSession creates a Guest Star session for a broadcaster.
func (s *GuestStarService) CreateSession(ctx context.Context, broadcasterID string) (*GetGuestStarSessionResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", broadcasterID)

	var resp GetGuestStarSessionResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/guest_star/session",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// EndSession ends a Guest Star session for a broadcaster.
func (s *GuestStarService) EndSession(ctx context.Context, params EndGuestStarSessionParams) (*GetGuestStarSessionResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("session_id", params.SessionID)

	var resp GetGuestStarSessionResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/guest_star/session",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// GetInvites fetches the current invite queue for a Guest Star session.
func (s *GuestStarService) GetInvites(ctx context.Context, params GuestStarInvitesParams) (*GetGuestStarInvitesResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("session_id", params.SessionID)

	var resp GetGuestStarInvitesResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodGet,
		Path:   "/guest_star/invites",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// SendInvite sends a Guest Star invite.
func (s *GuestStarService) SendInvite(ctx context.Context, params SendGuestStarInviteParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("session_id", params.SessionID)
	query.Set("guest_id", params.GuestID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/guest_star/invites",
		Query:  query,
	}, nil)
}

// DeleteInvite revokes a Guest Star invite.
func (s *GuestStarService) DeleteInvite(ctx context.Context, params DeleteGuestStarInviteParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("session_id", params.SessionID)
	query.Set("guest_id", params.GuestID)

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/guest_star/invites",
		Query:  query,
	}, nil)
}

// AssignSlot assigns a Guest Star guest to a slot.
func (s *GuestStarService) AssignSlot(ctx context.Context, params AssignGuestStarSlotParams) (*AssignGuestStarSlotResponse, *Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("session_id", params.SessionID)
	query.Set("guest_id", params.GuestID)
	query.Set("slot_id", params.SlotID)

	var resp AssignGuestStarSlotResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: http.MethodPost,
		Path:   "/guest_star/slot",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// UpdateSlot moves or swaps a Guest Star slot assignment.
func (s *GuestStarService) UpdateSlot(ctx context.Context, params UpdateGuestStarSlotParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("session_id", params.SessionID)
	query.Set("source_slot_id", params.SourceSlotID)
	if params.DestinationSlotID != "" {
		query.Set("destination_slot_id", params.DestinationSlotID)
	}

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/guest_star/slot",
		Query:  query,
	}, nil)
}

// DeleteSlot removes a Guest Star slot assignment.
func (s *GuestStarService) DeleteSlot(ctx context.Context, params DeleteGuestStarSlotParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("session_id", params.SessionID)
	query.Set("guest_id", params.GuestID)
	query.Set("slot_id", params.SlotID)
	if params.ShouldReinviteGuest != nil {
		query.Set("should_reinvite_guest", strconv.FormatBool(*params.ShouldReinviteGuest))
	}

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodDelete,
		Path:   "/guest_star/slot",
		Query:  query,
	}, nil)
}

// UpdateSlotSettings updates a Guest Star slot's media and live settings.
func (s *GuestStarService) UpdateSlotSettings(ctx context.Context, params UpdateGuestStarSlotSettingsParams) (*Response, error) {
	query := url.Values{}
	query.Set("broadcaster_id", params.BroadcasterID)
	query.Set("moderator_id", params.ModeratorID)
	query.Set("session_id", params.SessionID)
	query.Set("slot_id", params.SlotID)
	if params.IsAudioEnabled != nil {
		query.Set("is_audio_enabled", strconv.FormatBool(*params.IsAudioEnabled))
	}
	if params.IsVideoEnabled != nil {
		query.Set("is_video_enabled", strconv.FormatBool(*params.IsVideoEnabled))
	}
	if params.IsLive != nil {
		query.Set("is_live", strconv.FormatBool(*params.IsLive))
	}
	if params.Volume != nil {
		query.Set("volume", strconv.Itoa(*params.Volume))
	}

	return s.client.Do(ctx, RawRequest{
		Method: http.MethodPatch,
		Path:   "/guest_star/slot_settings",
		Query:  query,
	}, nil)
}
