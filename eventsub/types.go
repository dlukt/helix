package eventsub

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Subscription identifies an EventSub subscription.
type Subscription struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Cost      int               `json:"cost,omitempty"`
	Condition map[string]string `json:"condition,omitempty"`
}

// Challenge is delivered during webhook callback verification.
type Challenge struct {
	Subscription Subscription `json:"subscription"`
	Challenge    string       `json:"challenge"`
}

// Revocation is delivered when a subscription is revoked.
type Revocation struct {
	MessageID    string
	Subscription Subscription `json:"subscription"`
}

// Notification wraps a typed EventSub notification.
type Notification struct {
	MessageID    string
	Subscription Subscription
	Event        any
}

// ChannelFollowEvent is emitted for channel.follow version 2 subscriptions.
type ChannelFollowEvent struct {
	UserID               string    `json:"user_id"`
	UserLogin            string    `json:"user_login"`
	UserName             string    `json:"user_name"`
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	FollowedAt           time.Time `json:"followed_at"`
}

// ChannelUpdateEvent is emitted for channel.update version 2 subscriptions.
type ChannelUpdateEvent struct {
	BroadcasterUserID           string   `json:"broadcaster_user_id"`
	BroadcasterUserLogin        string   `json:"broadcaster_user_login"`
	BroadcasterUserName         string   `json:"broadcaster_user_name"`
	Title                       string   `json:"title"`
	Language                    string   `json:"language"`
	CategoryID                  string   `json:"category_id"`
	CategoryName                string   `json:"category_name"`
	ContentClassificationLabels []string `json:"content_classification_labels"`
}

// AutomodMessageHoldEvent is emitted for automod.message.hold version 1 subscriptions.
type AutomodMessageHoldEvent struct {
	BroadcasterUserID    string                  `json:"broadcaster_user_id"`
	BroadcasterUserLogin string                  `json:"broadcaster_user_login"`
	BroadcasterUserName  string                  `json:"broadcaster_user_name"`
	UserID               string                  `json:"user_id"`
	UserLogin            string                  `json:"user_login"`
	UserName             string                  `json:"user_name"`
	MessageID            string                  `json:"message_id"`
	Message              AutomodMessageV1Message `json:"message"`
	Category             string                  `json:"category"`
	Level                int                     `json:"level"`
	HeldAt               time.Time               `json:"held_at"`
}

// AutomodBoundary identifies the matched text span inside an AutoMod v2 message.
type AutomodBoundary struct {
	StartPos int `json:"start_pos"`
	EndPos   int `json:"end_pos"`
}

// AutomodV2Details contains the AutoMod category and matched boundaries for a v2 event.
type AutomodV2Details struct {
	Category   string            `json:"category"`
	Level      int               `json:"level"`
	Boundaries []AutomodBoundary `json:"boundaries"`
}

// AutomodBlockedTermFound contains one blocked term match in an AutoMod v2 event.
type AutomodBlockedTermFound struct {
	TermID                    string          `json:"term_id"`
	OwnerBroadcasterUserID    string          `json:"owner_broadcaster_user_id"`
	OwnerBroadcasterUserLogin string          `json:"owner_broadcaster_user_login"`
	OwnerBroadcasterUserName  string          `json:"owner_broadcaster_user_name"`
	Boundary                  AutomodBoundary `json:"boundary"`
}

// AutomodBlockedTermDetails contains the blocked-term matches for an AutoMod v2 event.
type AutomodBlockedTermDetails struct {
	TermsFound []AutomodBlockedTermFound `json:"terms_found"`
}

// AutomodMessageHoldV2Event is emitted for automod.message.hold version 2 subscriptions.
type AutomodMessageHoldV2Event struct {
	BroadcasterUserID    string                     `json:"broadcaster_user_id"`
	BroadcasterUserLogin string                     `json:"broadcaster_user_login"`
	BroadcasterUserName  string                     `json:"broadcaster_user_name"`
	UserID               string                     `json:"user_id"`
	UserLogin            string                     `json:"user_login"`
	UserName             string                     `json:"user_name"`
	MessageID            string                     `json:"message_id"`
	Message              ChatMessage                `json:"message"`
	Reason               string                     `json:"reason"`
	Automod              *AutomodV2Details          `json:"automod"`
	BlockedTerm          *AutomodBlockedTermDetails `json:"blocked_term"`
	HeldAt               time.Time                  `json:"held_at"`
}

// AutomodMessageUpdateEvent is emitted for automod.message.update version 1 subscriptions.
type AutomodMessageUpdateEvent struct {
	BroadcasterUserID    string                  `json:"broadcaster_user_id"`
	BroadcasterUserLogin string                  `json:"broadcaster_user_login"`
	BroadcasterUserName  string                  `json:"broadcaster_user_name"`
	UserID               string                  `json:"user_id"`
	UserLogin            string                  `json:"user_login"`
	UserName             string                  `json:"user_name"`
	ModeratorUserID      string                  `json:"moderator_user_id"`
	ModeratorUserName    string                  `json:"moderator_user_name"`
	ModeratorUserLogin   string                  `json:"moderator_user_login"`
	MessageID            string                  `json:"message_id"`
	Message              AutomodMessageV1Message `json:"message"`
	Category             string                  `json:"category"`
	Level                int                     `json:"level"`
	Status               string                  `json:"status"`
	HeldAt               time.Time               `json:"held_at"`
}

// AutomodMessageUpdateV2Event is emitted for automod.message.update version 2 subscriptions.
type AutomodMessageUpdateV2Event struct {
	BroadcasterUserID    string                     `json:"broadcaster_user_id"`
	BroadcasterUserLogin string                     `json:"broadcaster_user_login"`
	BroadcasterUserName  string                     `json:"broadcaster_user_name"`
	ModeratorUserID      string                     `json:"moderator_user_id"`
	ModeratorUserLogin   string                     `json:"moderator_user_login"`
	ModeratorUserName    string                     `json:"moderator_user_name"`
	UserID               string                     `json:"user_id"`
	UserLogin            string                     `json:"user_login"`
	UserName             string                     `json:"user_name"`
	MessageID            string                     `json:"message_id"`
	Message              ChatMessage                `json:"message"`
	Reason               string                     `json:"reason"`
	Automod              *AutomodV2Details          `json:"automod"`
	BlockedTerm          *AutomodBlockedTermDetails `json:"blocked_term"`
	Status               string                     `json:"status"`
	HeldAt               time.Time                  `json:"held_at"`
}

// AutomodMessageV1Message contains the AutoMod v1 message text and attached fragments.
type AutomodMessageV1Message struct {
	Text      string                     `json:"text"`
	Fragments []AutomodMessageV1Fragment `json:"fragments"`
}

// AutomodMessageV1FragmentEmote describes an emote fragment in an AutoMod v1 message.
type AutomodMessageV1FragmentEmote struct {
	ID         string `json:"id"`
	EmoteSetID string `json:"emote_set_id"`
}

// AutomodMessageV1FragmentCheermote describes a cheermote fragment in an AutoMod v1 message.
type AutomodMessageV1FragmentCheermote struct {
	Prefix string `json:"prefix"`
	Bits   int    `json:"bits"`
	Tier   int    `json:"tier"`
}

// AutomodMessageV1Fragment describes one fragment in an AutoMod v1 message.
type AutomodMessageV1Fragment struct {
	Type      string                             `json:"type"`
	Text      string                             `json:"text"`
	Cheermote *AutomodMessageV1FragmentCheermote `json:"cheermote"`
	Emote     *AutomodMessageV1FragmentEmote     `json:"emote"`
}

// AutomodSettingsUpdateEvent is emitted for automod.settings.update version 1 subscriptions.
type AutomodSettingsUpdateEvent struct {
	BroadcasterUserID       string `json:"broadcaster_user_id"`
	BroadcasterUserLogin    string `json:"broadcaster_user_login"`
	BroadcasterUserName     string `json:"broadcaster_user_name"`
	ModeratorUserID         string `json:"moderator_user_id"`
	ModeratorUserLogin      string `json:"moderator_user_login"`
	ModeratorUserName       string `json:"moderator_user_name"`
	Bullying                int    `json:"bullying"`
	OverallLevel            *int   `json:"overall_level"`
	Disability              int    `json:"disability"`
	RaceEthnicityOrReligion int    `json:"race_ethnicity_or_religion"`
	Misogyny                int    `json:"misogyny"`
	SexualitySexOrGender    int    `json:"sexuality_sex_or_gender"`
	Aggression              int    `json:"aggression"`
	SexBasedTerms           int    `json:"sex_based_terms"`
	Swearing                int    `json:"swearing"`
}

// UnmarshalJSON accepts both Twitch's documented {"data":[...]} payload wrapper and a flat event object.
func (e *AutomodSettingsUpdateEvent) UnmarshalJSON(data []byte) error {
	type rawAutomodSettingsUpdateEvent AutomodSettingsUpdateEvent

	var wrapped struct {
		Data *json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	if wrapped.Data != nil {
		var items []rawAutomodSettingsUpdateEvent
		if err := json.Unmarshal(*wrapped.Data, &items); err != nil {
			return err
		}
		if len(items) == 0 {
			return fmt.Errorf("eventsub: automod settings update missing data item")
		}
		*e = AutomodSettingsUpdateEvent(items[0])
		return nil
	}

	var flat rawAutomodSettingsUpdateEvent
	if err := json.Unmarshal(data, &flat); err != nil {
		return err
	}
	*e = AutomodSettingsUpdateEvent(flat)
	return nil
}

// AutomodTermsUpdateEvent is emitted for automod.terms.update version 1 subscriptions.
type AutomodTermsUpdateEvent struct {
	BroadcasterUserID    string   `json:"broadcaster_user_id"`
	BroadcasterUserLogin string   `json:"broadcaster_user_login"`
	BroadcasterUserName  string   `json:"broadcaster_user_name"`
	ModeratorUserID      string   `json:"moderator_user_id"`
	ModeratorUserLogin   string   `json:"moderator_user_login"`
	ModeratorUserName    string   `json:"moderator_user_name"`
	Action               string   `json:"action"`
	FromAutomod          bool     `json:"from_automod"`
	Terms                []string `json:"terms"`
}

// GuestStarSessionEvent contains the shared fields for guest star session lifecycle events.
type GuestStarSessionEvent struct {
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	ModeratorUserID      string    `json:"moderator_user_id"`
	ModeratorUserName    string    `json:"moderator_user_name"`
	ModeratorUserLogin   string    `json:"moderator_user_login"`
	SessionID            string    `json:"session_id"`
	StartedAt            time.Time `json:"started_at"`
}

// ChannelGuestStarSessionBeginEvent is emitted for channel.guest_star_session.begin version beta subscriptions.
type ChannelGuestStarSessionBeginEvent struct {
	GuestStarSessionEvent
}

// ChannelGuestStarSessionEndEvent is emitted for channel.guest_star_session.end version beta subscriptions.
type ChannelGuestStarSessionEndEvent struct {
	GuestStarSessionEvent
	EndedAt time.Time `json:"ended_at"`
}

// ChannelGuestStarGuestUpdateEvent is emitted for channel.guest_star_guest.update version beta subscriptions.
type ChannelGuestStarGuestUpdateEvent struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	SessionID            string `json:"session_id"`
	ModeratorUserID      string `json:"moderator_user_id"`
	ModeratorUserName    string `json:"moderator_user_name"`
	ModeratorUserLogin   string `json:"moderator_user_login"`
	GuestUserID          string `json:"guest_user_id"`
	GuestUserName        string `json:"guest_user_name"`
	GuestUserLogin       string `json:"guest_user_login"`
	SlotID               string `json:"slot_id"`
	State                string `json:"state"`
	HostVideoEnabled     bool   `json:"host_video_enabled"`
	HostAudioEnabled     bool   `json:"host_audio_enabled"`
	HostVolume           int    `json:"host_volume"`
}

// ChannelGuestStarSettingsUpdateEvent is emitted for channel.guest_star_settings.update version beta subscriptions.
type ChannelGuestStarSettingsUpdateEvent struct {
	BroadcasterUserID           string `json:"broadcaster_user_id"`
	BroadcasterUserName         string `json:"broadcaster_user_name"`
	BroadcasterUserLogin        string `json:"broadcaster_user_login"`
	IsModeratorSendLiveEnabled  bool   `json:"is_moderator_send_live_enabled"`
	SlotCount                   int    `json:"slot_count"`
	IsBrowserSourceAudioEnabled bool   `json:"is_browser_source_audio_enabled"`
	GroupLayout                 string `json:"group_layout"`
}

// ExtensionBitsTransactionProduct describes the extension product involved in a Bits transaction.
type ExtensionBitsTransactionProduct struct {
	Name          string `json:"name"`
	SKU           string `json:"sku"`
	Bits          int    `json:"bits"`
	InDevelopment bool   `json:"in_development"`
}

// ExtensionBitsTransactionCreateEvent is emitted for extension.bits_transaction.create version 1 subscriptions.
type ExtensionBitsTransactionCreateEvent struct {
	ID                   string                          `json:"id"`
	ExtensionClientID    string                          `json:"extension_client_id"`
	BroadcasterUserID    string                          `json:"broadcaster_user_id"`
	BroadcasterUserLogin string                          `json:"broadcaster_user_login"`
	BroadcasterUserName  string                          `json:"broadcaster_user_name"`
	UserName             string                          `json:"user_name"`
	UserLogin            string                          `json:"user_login"`
	UserID               string                          `json:"user_id"`
	Product              ExtensionBitsTransactionProduct `json:"product"`
}

// ConduitShardDisabledTransport describes the transport affected by a disabled conduit shard.
type ConduitShardDisabledTransport struct {
	Method         string     `json:"method"`
	Callback       *string    `json:"callback"`
	SessionID      *string    `json:"session_id"`
	ConnectedAt    *time.Time `json:"connected_at"`
	DisconnectedAt *time.Time `json:"disconnected_at"`
}

// ConduitShardDisabledEvent is emitted for conduit.shard.disabled version 1 subscriptions.
type ConduitShardDisabledEvent struct {
	ConduitID string                        `json:"conduit_id"`
	ShardID   string                        `json:"shard_id"`
	Status    string                        `json:"status"`
	Transport ConduitShardDisabledTransport `json:"transport"`
}

// DropEntitlementGrantData describes one granted entitlement in a drop batch notification.
type DropEntitlementGrantData struct {
	OrganizationID string    `json:"organization_id"`
	CategoryID     *string   `json:"category_id"`
	CategoryName   *string   `json:"category_name"`
	CampaignID     *string   `json:"campaign_id"`
	UserID         string    `json:"user_id"`
	UserName       string    `json:"user_name"`
	UserLogin      string    `json:"user_login"`
	EntitlementID  string    `json:"entitlement_id"`
	BenefitID      string    `json:"benefit_id"`
	CreatedAt      time.Time `json:"created_at"`
}

// DropEntitlementGrantEvent describes one item in a drop entitlement batch.
type DropEntitlementGrantEvent struct {
	ID   string                   `json:"id"`
	Data DropEntitlementGrantData `json:"data"`
}

// DropEntitlementGrantBatch is emitted for drop.entitlement.grant version 1 subscriptions.
type DropEntitlementGrantBatch []DropEntitlementGrantEvent

// ChannelModerateUser identifies a user referenced by a moderation action.
type ChannelModerateUser struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// ChannelModerateFollowers describes followers-only settings set by a moderation action.
type ChannelModerateFollowers struct {
	FollowDurationMinutes int `json:"follow_duration_minutes"`
}

// ChannelModerateSlow describes slow-mode settings set by a moderation action.
type ChannelModerateSlow struct {
	WaitTimeSeconds int `json:"wait_time_seconds"`
}

// ChannelModerateBanAction describes a ban moderation action target.
type ChannelModerateBanAction struct {
	ChannelModerateUser
	Reason string `json:"reason"`
}

// ChannelModerateTimeoutAction describes a timeout moderation action target.
type ChannelModerateTimeoutAction struct {
	ChannelModerateUser
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// ChannelModerateRaidAction describes a raid or unraid moderation action target.
type ChannelModerateRaidAction struct {
	UserID          string `json:"user_id"`
	UserLogin       string `json:"user_login"`
	UserName        string `json:"user_name"`
	ViewerCount     int    `json:"viewer_count"`
	ProfileImageURL string `json:"profile_image_url"`
}

// ChannelModerateDeleteAction describes a deleted message moderation action target.
type ChannelModerateDeleteAction struct {
	ChannelModerateUser
	MessageID   string `json:"message_id"`
	MessageBody string `json:"message_body"`
}

// ChannelModerateAutomodTermsAction describes an add/remove blocked or permitted term action.
type ChannelModerateAutomodTermsAction struct {
	Action      string   `json:"action"`
	List        string   `json:"list"`
	Terms       []string `json:"terms"`
	FromAutomod bool     `json:"from_automod"`
}

// ChannelModerateUnbanRequestAction describes an approve/deny unban request action.
type ChannelModerateUnbanRequestAction struct {
	IsApproved       bool   `json:"is_approved"`
	UserID           string `json:"user_id"`
	UserLogin        string `json:"user_login"`
	UserName         string `json:"user_name"`
	ModeratorMessage string `json:"moderator_message"`
}

// ChannelModerateWarnAction describes a warn action in channel.moderate version 2.
type ChannelModerateWarnAction struct {
	ChannelModerateUser
	Reason         *string  `json:"reason"`
	ChatRulesCited []string `json:"chat_rules_cited"`
}

// ChannelModerateEvent contains the shared moderation action payload fields.
type ChannelModerateEvent struct {
	BroadcasterUserID          string                             `json:"broadcaster_user_id"`
	BroadcasterUserLogin       string                             `json:"broadcaster_user_login"`
	BroadcasterUserName        string                             `json:"broadcaster_user_name"`
	SourceBroadcasterUserID    *string                            `json:"source_broadcaster_user_id"`
	SourceBroadcasterUserLogin *string                            `json:"source_broadcaster_user_login"`
	SourceBroadcasterUserName  *string                            `json:"source_broadcaster_user_name"`
	ModeratorUserID            string                             `json:"moderator_user_id"`
	ModeratorUserLogin         string                             `json:"moderator_user_login"`
	ModeratorUserName          string                             `json:"moderator_user_name"`
	Action                     string                             `json:"action"`
	Followers                  *ChannelModerateFollowers          `json:"followers"`
	Slow                       *ChannelModerateSlow               `json:"slow"`
	VIP                        *ChannelModerateUser               `json:"vip"`
	UnVIP                      *ChannelModerateUser               `json:"unvip"`
	Mod                        *ChannelModerateUser               `json:"mod"`
	Unmod                      *ChannelModerateUser               `json:"unmod"`
	Ban                        *ChannelModerateBanAction          `json:"ban"`
	Unban                      *ChannelModerateUser               `json:"unban"`
	Timeout                    *ChannelModerateTimeoutAction      `json:"timeout"`
	Untimeout                  *ChannelModerateUser               `json:"untimeout"`
	Raid                       *ChannelModerateRaidAction         `json:"raid"`
	Unraid                     *ChannelModerateRaidAction         `json:"unraid"`
	Delete                     *ChannelModerateDeleteAction       `json:"delete"`
	AutomodTerms               *ChannelModerateAutomodTermsAction `json:"automod_terms"`
	UnbanRequest               *ChannelModerateUnbanRequestAction `json:"unban_request"`
	SharedChatBan              *ChannelModerateBanAction          `json:"shared_chat_ban"`
	SharedChatUnban            *ChannelModerateUser               `json:"shared_chat_unban"`
	SharedChatTimeout          *ChannelModerateTimeoutAction      `json:"shared_chat_timeout"`
	SharedChatUntimeout        *ChannelModerateUser               `json:"shared_chat_untimeout"`
	SharedChatDelete           *ChannelModerateDeleteAction       `json:"shared_chat_delete"`
}

// ChannelModerateEventV1 is emitted for channel.moderate version 1 subscriptions.
type ChannelModerateEventV1 struct {
	ChannelModerateEvent
}

// ChannelModerateEventV2 is emitted for channel.moderate version 2 subscriptions.
type ChannelModerateEventV2 struct {
	ChannelModerateEvent
	Warn *ChannelModerateWarnAction `json:"warn"`
}

// ChannelAdBreakBeginEvent is emitted for channel.ad_break.begin version 1 subscriptions.
type ChannelAdBreakBeginEvent struct {
	DurationSeconds      int       `json:"duration_seconds"`
	StartedAt            time.Time `json:"started_at"`
	IsAutomatic          bool      `json:"is_automatic"`
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	RequesterUserID      string    `json:"requester_user_id"`
	RequesterUserLogin   string    `json:"requester_user_login"`
	RequesterUserName    string    `json:"requester_user_name"`
}

// BitsUseMessageCheermote describes a cheermote fragment in a bits-use message.
type BitsUseMessageCheermote struct {
	Prefix string `json:"prefix"`
	Bits   int    `json:"bits"`
	Tier   int    `json:"tier"`
}

// BitsUseMessageEmote describes an emote fragment in a bits-use message.
type BitsUseMessageEmote struct {
	ID         string   `json:"id"`
	EmoteSetID string   `json:"emote_set_id"`
	OwnerID    string   `json:"owner_id"`
	Format     []string `json:"format"`
}

// BitsUseMessageFragment describes one fragment in a bits-use message.
type BitsUseMessageFragment struct {
	Type      string                   `json:"type"`
	Text      string                   `json:"text"`
	Cheermote *BitsUseMessageCheermote `json:"cheermote"`
	Emote     *BitsUseMessageEmote     `json:"emote"`
}

// BitsUseMessage contains the message payload included with a bits-use event.
type BitsUseMessage struct {
	Text      string                   `json:"text"`
	Fragments []BitsUseMessageFragment `json:"fragments"`
}

// ChannelBitsUseEvent is emitted for channel.bits.use version 1 subscriptions.
type ChannelBitsUseEvent struct {
	UserID               string          `json:"user_id"`
	UserLogin            string          `json:"user_login"`
	UserName             string          `json:"user_name"`
	BroadcasterUserID    string          `json:"broadcaster_user_id"`
	BroadcasterUserLogin string          `json:"broadcaster_user_login"`
	BroadcasterUserName  string          `json:"broadcaster_user_name"`
	Bits                 int             `json:"bits"`
	Type                 string          `json:"type"`
	PowerUp              json.RawMessage `json:"power_up"`
	CustomPowerUp        json.RawMessage `json:"custom_power_up"`
	Message              BitsUseMessage  `json:"message"`
}

// ChannelPointsRewardImage contains reward image URLs in multiple sizes.
type ChannelPointsRewardImage struct {
	URL1x string `json:"url_1x"`
	URL2x string `json:"url_2x"`
	URL4x string `json:"url_4x"`
}

// ChannelPointsRewardLimit describes a reward usage limit.
type ChannelPointsRewardLimit struct {
	IsEnabled bool `json:"is_enabled"`
	Value     int  `json:"value"`
}

// ChannelPointsRewardGlobalCooldown describes a reward cooldown policy.
type ChannelPointsRewardGlobalCooldown struct {
	IsEnabled bool `json:"is_enabled"`
	Seconds   int  `json:"seconds"`
}

// ChannelPointsCustomReward contains the full reward configuration from custom reward events.
type ChannelPointsCustomReward struct {
	ID                                string                            `json:"id"`
	BroadcasterUserID                 string                            `json:"broadcaster_user_id"`
	BroadcasterUserLogin              string                            `json:"broadcaster_user_login"`
	BroadcasterUserName               string                            `json:"broadcaster_user_name"`
	IsEnabled                         bool                              `json:"is_enabled"`
	IsPaused                          bool                              `json:"is_paused"`
	IsInStock                         bool                              `json:"is_in_stock"`
	Title                             string                            `json:"title"`
	Cost                              int                               `json:"cost"`
	Prompt                            string                            `json:"prompt"`
	IsUserInputRequired               bool                              `json:"is_user_input_required"`
	ShouldRedemptionsSkipRequestQueue bool                              `json:"should_redemptions_skip_request_queue"`
	CooldownExpiresAt                 *time.Time                        `json:"cooldown_expires_at"`
	RedemptionsRedeemedCurrentStream  *int                              `json:"redemptions_redeemed_current_stream"`
	MaxPerStream                      ChannelPointsRewardLimit          `json:"max_per_stream"`
	MaxPerUserPerStream               ChannelPointsRewardLimit          `json:"max_per_user_per_stream"`
	GlobalCooldown                    ChannelPointsRewardGlobalCooldown `json:"global_cooldown"`
	BackgroundColor                   string                            `json:"background_color"`
	Image                             *ChannelPointsRewardImage         `json:"image"`
	DefaultImage                      *ChannelPointsRewardImage         `json:"default_image"`
}

// ChannelPointsCustomRewardAddEvent is emitted for channel.channel_points_custom_reward.add version 1 subscriptions.
type ChannelPointsCustomRewardAddEvent struct {
	ChannelPointsCustomReward
}

// ChannelPointsCustomRewardUpdateEvent is emitted for channel.channel_points_custom_reward.update version 1 subscriptions.
type ChannelPointsCustomRewardUpdateEvent struct {
	ChannelPointsCustomReward
}

// ChannelPointsCustomRewardRemoveEvent is emitted for channel.channel_points_custom_reward.remove version 1 subscriptions.
type ChannelPointsCustomRewardRemoveEvent struct {
	ChannelPointsCustomReward
}

// ChannelPointsCustomRewardRedemptionReward contains the reward snapshot on a redemption event.
type ChannelPointsCustomRewardRedemptionReward struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Cost   int    `json:"cost"`
	Prompt string `json:"prompt"`
}

// ChannelPointsCustomRewardRedemption contains the shared payload for custom reward redemption events.
type ChannelPointsCustomRewardRedemption struct {
	ID                   string                                    `json:"id"`
	BroadcasterUserID    string                                    `json:"broadcaster_user_id"`
	BroadcasterUserLogin string                                    `json:"broadcaster_user_login"`
	BroadcasterUserName  string                                    `json:"broadcaster_user_name"`
	UserID               string                                    `json:"user_id"`
	UserLogin            string                                    `json:"user_login"`
	UserName             string                                    `json:"user_name"`
	UserInput            string                                    `json:"user_input"`
	Status               string                                    `json:"status"`
	Reward               ChannelPointsCustomRewardRedemptionReward `json:"reward"`
	RedeemedAt           time.Time                                 `json:"redeemed_at"`
}

// ChannelPointsCustomRewardRedemptionAddEvent is emitted for channel.channel_points_custom_reward_redemption.add version 1 subscriptions.
type ChannelPointsCustomRewardRedemptionAddEvent struct {
	ChannelPointsCustomRewardRedemption
}

// ChannelPointsCustomRewardRedemptionUpdateEvent is emitted for channel.channel_points_custom_reward_redemption.update version 1 subscriptions.
type ChannelPointsCustomRewardRedemptionUpdateEvent struct {
	ChannelPointsCustomRewardRedemption
}

// ChannelPointsAutomaticRewardRedemptionV1UnlockedEmote describes an emote unlocked by an automatic reward redemption.
type ChannelPointsAutomaticRewardRedemptionV1UnlockedEmote struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ChannelPointsAutomaticRewardRedemptionV1Reward contains the redeemed automatic reward details in version 1.
type ChannelPointsAutomaticRewardRedemptionV1Reward struct {
	Type          string                                                 `json:"type"`
	Cost          int                                                    `json:"cost"`
	UnlockedEmote *ChannelPointsAutomaticRewardRedemptionV1UnlockedEmote `json:"unlocked_emote"`
}

// ChannelPointsAutomaticRewardRedemptionV1MessageEmote describes an emote span in a version 1 automatic reward message.
type ChannelPointsAutomaticRewardRedemptionV1MessageEmote struct {
	ID    string `json:"id"`
	Begin int    `json:"begin"`
	End   int    `json:"end"`
}

// ChannelPointsAutomaticRewardRedemptionV1Message contains the version 1 message payload.
type ChannelPointsAutomaticRewardRedemptionV1Message struct {
	Text   string                                                 `json:"text"`
	Emotes []ChannelPointsAutomaticRewardRedemptionV1MessageEmote `json:"emotes"`
}

// ChannelPointsAutomaticRewardRedemptionAddEvent is emitted for channel.channel_points_automatic_reward_redemption.add version 1 subscriptions.
type ChannelPointsAutomaticRewardRedemptionAddEvent struct {
	BroadcasterUserID    string                                          `json:"broadcaster_user_id"`
	BroadcasterUserLogin string                                          `json:"broadcaster_user_login"`
	BroadcasterUserName  string                                          `json:"broadcaster_user_name"`
	UserID               string                                          `json:"user_id"`
	UserLogin            string                                          `json:"user_login"`
	UserName             string                                          `json:"user_name"`
	ID                   string                                          `json:"id"`
	Reward               ChannelPointsAutomaticRewardRedemptionV1Reward  `json:"reward"`
	Message              ChannelPointsAutomaticRewardRedemptionV1Message `json:"message"`
	UserInput            string                                          `json:"user_input"`
	RedeemedAt           time.Time                                       `json:"redeemed_at"`
}

// ChannelPointsAutomaticRewardRedemptionV2Emote describes the emote metadata in a version 2 automatic reward message fragment.
type ChannelPointsAutomaticRewardRedemptionV2Emote struct {
	ID string `json:"id"`
}

// ChannelPointsAutomaticRewardRedemptionV2MessageFragment describes one fragment in a version 2 automatic reward message.
type ChannelPointsAutomaticRewardRedemptionV2MessageFragment struct {
	Type  string                                         `json:"type"`
	Text  string                                         `json:"text"`
	Emote *ChannelPointsAutomaticRewardRedemptionV2Emote `json:"emote"`
}

// ChannelPointsAutomaticRewardRedemptionV2Message contains the version 2 message payload.
type ChannelPointsAutomaticRewardRedemptionV2Message struct {
	Text      string                                                    `json:"text"`
	Fragments []ChannelPointsAutomaticRewardRedemptionV2MessageFragment `json:"fragments"`
}

// ChannelPointsAutomaticRewardRedemptionV2RewardEmote describes the reward emote for version 2.
type ChannelPointsAutomaticRewardRedemptionV2RewardEmote struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ChannelPointsAutomaticRewardRedemptionV2Reward contains the redeemed automatic reward details in version 2.
type ChannelPointsAutomaticRewardRedemptionV2Reward struct {
	Type          string                                               `json:"type"`
	ChannelPoints int                                                  `json:"channel_points"`
	Emote         *ChannelPointsAutomaticRewardRedemptionV2RewardEmote `json:"emote"`
}

// ChannelPointsAutomaticRewardRedemptionAddV2Event is emitted for channel.channel_points_automatic_reward_redemption.add version 2 subscriptions.
type ChannelPointsAutomaticRewardRedemptionAddV2Event struct {
	BroadcasterUserID    string                                           `json:"broadcaster_user_id"`
	BroadcasterUserLogin string                                           `json:"broadcaster_user_login"`
	BroadcasterUserName  string                                           `json:"broadcaster_user_name"`
	UserID               string                                           `json:"user_id"`
	UserLogin            string                                           `json:"user_login"`
	UserName             string                                           `json:"user_name"`
	ID                   string                                           `json:"id"`
	Reward               ChannelPointsAutomaticRewardRedemptionV2Reward   `json:"reward"`
	Message              *ChannelPointsAutomaticRewardRedemptionV2Message `json:"message"`
	RedeemedAt           time.Time                                        `json:"redeemed_at"`
}

// UnmarshalJSON accepts both the native JSON types and the quoted values shown in Twitch's example payload.
func (e *ChannelAdBreakBeginEvent) UnmarshalJSON(data []byte) error {
	type rawEvent struct {
		DurationSeconds      json.RawMessage `json:"duration_seconds"`
		StartedAt            time.Time       `json:"started_at"`
		IsAutomatic          json.RawMessage `json:"is_automatic"`
		BroadcasterUserID    string          `json:"broadcaster_user_id"`
		BroadcasterUserLogin string          `json:"broadcaster_user_login"`
		BroadcasterUserName  string          `json:"broadcaster_user_name"`
		RequesterUserID      string          `json:"requester_user_id"`
		RequesterUserLogin   string          `json:"requester_user_login"`
		RequesterUserName    string          `json:"requester_user_name"`
	}

	var raw rawEvent
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	durationSeconds, err := parseJSONInt(raw.DurationSeconds)
	if err != nil {
		return fmt.Errorf("eventsub: decode ad break duration_seconds: %w", err)
	}
	isAutomatic, err := parseJSONBool(raw.IsAutomatic)
	if err != nil {
		return fmt.Errorf("eventsub: decode ad break is_automatic: %w", err)
	}

	*e = ChannelAdBreakBeginEvent{
		DurationSeconds:      durationSeconds,
		StartedAt:            raw.StartedAt,
		IsAutomatic:          isAutomatic,
		BroadcasterUserID:    raw.BroadcasterUserID,
		BroadcasterUserLogin: raw.BroadcasterUserLogin,
		BroadcasterUserName:  raw.BroadcasterUserName,
		RequesterUserID:      raw.RequesterUserID,
		RequesterUserLogin:   raw.RequesterUserLogin,
		RequesterUserName:    raw.RequesterUserName,
	}
	return nil
}

// StreamOnlineEvent is emitted for stream.online version 1 subscriptions.
type StreamOnlineEvent struct {
	ID                   string    `json:"id"`
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	Type                 string    `json:"type"`
	StartedAt            time.Time `json:"started_at"`
}

// StreamOfflineEvent is emitted for stream.offline version 1 subscriptions.
type StreamOfflineEvent struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

// ChannelChatClearEvent is emitted for channel.chat.clear version 1 subscriptions.
type ChannelChatClearEvent struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
}

// ChannelChatClearUserMessagesEvent is emitted for channel.chat.clear_user_messages version 1 subscriptions.
type ChannelChatClearUserMessagesEvent struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	TargetUserID         string `json:"target_user_id"`
	TargetUserName       string `json:"target_user_name"`
	TargetUserLogin      string `json:"target_user_login"`
}

// ChannelChatMessageDeleteEvent is emitted for channel.chat.message_delete version 1 subscriptions.
type ChannelChatMessageDeleteEvent struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	TargetUserID         string `json:"target_user_id"`
	TargetUserName       string `json:"target_user_name"`
	TargetUserLogin      string `json:"target_user_login"`
	MessageID            string `json:"message_id"`
}

// ChatMessageCheermote describes a cheermote fragment in a chat message.
type ChatMessageCheermote struct {
	Prefix string `json:"prefix"`
	Bits   int    `json:"bits"`
	Tier   int    `json:"tier"`
}

// ChatMessageEmote describes an emote fragment in a chat message.
type ChatMessageEmote struct {
	ID         string   `json:"id"`
	EmoteSetID string   `json:"emote_set_id"`
	OwnerID    string   `json:"owner_id"`
	Format     []string `json:"format"`
}

// ChatMessageMention describes a mentioned user in a chat message fragment.
type ChatMessageMention struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	UserLogin string `json:"user_login"`
}

// ChatMessageFragment describes one fragment in a chat message.
type ChatMessageFragment struct {
	Type      string                `json:"type"`
	Text      string                `json:"text"`
	Cheermote *ChatMessageCheermote `json:"cheermote"`
	Emote     *ChatMessageEmote     `json:"emote"`
	Mention   *ChatMessageMention   `json:"mention"`
}

// ChatMessage contains a structured chat message body.
type ChatMessage struct {
	Text      string                `json:"text"`
	Fragments []ChatMessageFragment `json:"fragments"`
}

// ChatBadge describes one badge attached to a chat user.
type ChatBadge struct {
	SetID string `json:"set_id"`
	ID    string `json:"id"`
	Info  string `json:"info"`
}

// ChatMessageCheer describes bits metadata attached to a chat message.
type ChatMessageCheer struct {
	Bits int `json:"bits"`
}

// ChatMessageReply describes reply metadata attached to a chat message.
type ChatMessageReply struct {
	ParentMessageID   string `json:"parent_message_id"`
	ParentMessageBody string `json:"parent_message_body"`
	ParentUserID      string `json:"parent_user_id"`
	ParentUserName    string `json:"parent_user_name"`
	ParentUserLogin   string `json:"parent_user_login"`
	ThreadMessageID   string `json:"thread_message_id"`
	ThreadUserID      string `json:"thread_user_id"`
	ThreadUserName    string `json:"thread_user_name"`
	ThreadUserLogin   string `json:"thread_user_login"`
}

// ChannelChatMessageEvent is emitted for channel.chat.message version 1 subscriptions.
type ChannelChatMessageEvent struct {
	BroadcasterUserID           string            `json:"broadcaster_user_id"`
	BroadcasterUserName         string            `json:"broadcaster_user_name"`
	BroadcasterUserLogin        string            `json:"broadcaster_user_login"`
	ChatterUserID               string            `json:"chatter_user_id"`
	ChatterUserName             string            `json:"chatter_user_name"`
	ChatterUserLogin            string            `json:"chatter_user_login"`
	MessageID                   string            `json:"message_id"`
	Message                     ChatMessage       `json:"message"`
	MessageType                 string            `json:"message_type"`
	Badges                      []ChatBadge       `json:"badges"`
	Cheer                       *ChatMessageCheer `json:"cheer"`
	Color                       string            `json:"color"`
	Reply                       *ChatMessageReply `json:"reply"`
	ChannelPointsCustomRewardID *string           `json:"channel_points_custom_reward_id"`
	SourceBroadcasterUserID     *string           `json:"source_broadcaster_user_id"`
	SourceBroadcasterUserName   *string           `json:"source_broadcaster_user_name"`
	SourceBroadcasterUserLogin  *string           `json:"source_broadcaster_user_login"`
	SourceMessageID             *string           `json:"source_message_id"`
	SourceBadges                []ChatBadge       `json:"source_badges"`
	IsSourceOnly                *bool             `json:"is_source_only"`
}

// ChannelChatNotificationSubNotice describes a new subscription notice shown in chat.
type ChannelChatNotificationSubNotice struct {
	SubTier        string `json:"sub_tier"`
	IsPrime        bool   `json:"is_prime"`
	DurationMonths int    `json:"duration_months"`
}

// ChannelChatNotificationResubNotice describes a resubscription notice shown in chat.
type ChannelChatNotificationResubNotice struct {
	CumulativeMonths  int     `json:"cumulative_months"`
	DurationMonths    int     `json:"duration_months"`
	StreakMonths      *int    `json:"streak_months"`
	SubTier           string  `json:"sub_tier"`
	IsPrime           *bool   `json:"is_prime"`
	IsGift            bool    `json:"is_gift"`
	GifterIsAnonymous *bool   `json:"gifter_is_anonymous"`
	GifterUserID      *string `json:"gifter_user_id"`
	GifterUserName    *string `json:"gifter_user_name"`
	GifterUserLogin   *string `json:"gifter_user_login"`
}

// ChannelChatNotificationSubGiftNotice describes a gifted subscription notice shown in chat.
type ChannelChatNotificationSubGiftNotice struct {
	DurationMonths     int     `json:"duration_months"`
	CumulativeTotal    *int    `json:"cumulative_total"`
	RecipientUserID    string  `json:"recipient_user_id"`
	RecipientUserName  string  `json:"recipient_user_name"`
	RecipientUserLogin string  `json:"recipient_user_login"`
	SubTier            string  `json:"sub_tier"`
	CommunityGiftID    *string `json:"community_gift_id"`
}

// ChannelChatNotificationCommunitySubGiftNotice describes a community gift notice shown in chat.
type ChannelChatNotificationCommunitySubGiftNotice struct {
	ID              string `json:"id"`
	Total           int    `json:"total"`
	SubTier         string `json:"sub_tier"`
	CumulativeTotal *int   `json:"cumulative_total"`
}

// ChannelChatNotificationGiftPaidUpgradeNotice describes a gift paid upgrade notice shown in chat.
type ChannelChatNotificationGiftPaidUpgradeNotice struct {
	GifterIsAnonymous bool    `json:"gifter_is_anonymous"`
	GifterUserID      *string `json:"gifter_user_id"`
	GifterUserName    *string `json:"gifter_user_name"`
	GifterUserLogin   *string `json:"gifter_user_login"`
}

// ChannelChatNotificationPrimePaidUpgradeNotice describes a Prime paid upgrade notice shown in chat.
type ChannelChatNotificationPrimePaidUpgradeNotice struct {
	SubTier string `json:"sub_tier"`
}

// ChannelChatNotificationPayItForwardNotice describes a pay-it-forward notice shown in chat.
type ChannelChatNotificationPayItForwardNotice struct {
	GifterIsAnonymous bool    `json:"gifter_is_anonymous"`
	GifterUserID      *string `json:"gifter_user_id"`
	GifterUserName    *string `json:"gifter_user_name"`
	GifterUserLogin   *string `json:"gifter_user_login"`
}

// ChannelChatNotificationRaidNotice describes a raid notice shown in chat.
type ChannelChatNotificationRaidNotice struct {
	UserID          string `json:"user_id"`
	UserName        string `json:"user_name"`
	UserLogin       string `json:"user_login"`
	ViewerCount     int    `json:"viewer_count"`
	ProfileImageURL string `json:"profile_image_url"`
}

// ChannelChatNotificationUnraidNotice describes an unraid notice shown in chat.
type ChannelChatNotificationUnraidNotice struct{}

// ChannelChatNotificationAnnouncementNotice describes an announcement notice shown in chat.
type ChannelChatNotificationAnnouncementNotice struct {
	Color string `json:"color"`
}

// ChannelChatNotificationBitsBadgeTierNotice describes a Bits badge tier notice shown in chat.
type ChannelChatNotificationBitsBadgeTierNotice struct {
	Tier int `json:"tier"`
}

// ChannelChatNotificationCharityAmount describes the donation amount in a charity donation notice.
type ChannelChatNotificationCharityAmount struct {
	Value        int    `json:"value"`
	DecimalPlace int    `json:"decimal_places"`
	Currency     string `json:"currency"`
}

// ChannelChatNotificationCharityDonationNotice describes a charity donation notice shown in chat.
type ChannelChatNotificationCharityDonationNotice struct {
	CharityName string                               `json:"charity_name"`
	Amount      ChannelChatNotificationCharityAmount `json:"amount"`
}

// ChannelChatNotificationWatchStreakNotice describes a watch streak notice shown in chat.
type ChannelChatNotificationWatchStreakNotice struct {
	StreakCount          int `json:"streak_count"`
	ChannelPointsAwarded int `json:"channel_points_awarded"`
}

// ChannelChatNotificationEvent is emitted for channel.chat.notification version 1 subscriptions.
type ChannelChatNotificationEvent struct {
	BroadcasterUserID          string                                         `json:"broadcaster_user_id"`
	BroadcasterUserLogin       string                                         `json:"broadcaster_user_login"`
	BroadcasterUserName        string                                         `json:"broadcaster_user_name"`
	ChatterUserID              string                                         `json:"chatter_user_id"`
	ChatterUserLogin           string                                         `json:"chatter_user_login"`
	ChatterUserName            string                                         `json:"chatter_user_name"`
	ChatterIsAnonymous         bool                                           `json:"chatter_is_anonymous"`
	Color                      string                                         `json:"color"`
	Badges                     []ChatBadge                                    `json:"badges"`
	SystemMessage              string                                         `json:"system_message"`
	MessageID                  string                                         `json:"message_id"`
	Message                    ChatMessage                                    `json:"message"`
	NoticeType                 string                                         `json:"notice_type"`
	Sub                        *ChannelChatNotificationSubNotice              `json:"sub"`
	Resub                      *ChannelChatNotificationResubNotice            `json:"resub"`
	SubGift                    *ChannelChatNotificationSubGiftNotice          `json:"sub_gift"`
	CommunitySubGift           *ChannelChatNotificationCommunitySubGiftNotice `json:"community_sub_gift"`
	GiftPaidUpgrade            *ChannelChatNotificationGiftPaidUpgradeNotice  `json:"gift_paid_upgrade"`
	PrimePaidUpgrade           *ChannelChatNotificationPrimePaidUpgradeNotice `json:"prime_paid_upgrade"`
	PayItForward               *ChannelChatNotificationPayItForwardNotice     `json:"pay_it_forward"`
	Raid                       *ChannelChatNotificationRaidNotice             `json:"raid"`
	Unraid                     *ChannelChatNotificationUnraidNotice           `json:"unraid"`
	Announcement               *ChannelChatNotificationAnnouncementNotice     `json:"announcement"`
	BitsBadgeTier              *ChannelChatNotificationBitsBadgeTierNotice    `json:"bits_badge_tier"`
	CharityDonation            *ChannelChatNotificationCharityDonationNotice  `json:"charity_donation"`
	WatchStreak                *ChannelChatNotificationWatchStreakNotice      `json:"watch_streak"`
	SharedChatSub              *ChannelChatNotificationSubNotice              `json:"shared_chat_sub"`
	SharedChatResub            *ChannelChatNotificationResubNotice            `json:"shared_chat_resub"`
	SharedChatSubGift          *ChannelChatNotificationSubGiftNotice          `json:"shared_chat_sub_gift"`
	SharedChatCommunitySubGift *ChannelChatNotificationCommunitySubGiftNotice `json:"shared_chat_community_sub_gift"`
	SharedChatGiftPaidUpgrade  *ChannelChatNotificationGiftPaidUpgradeNotice  `json:"shared_chat_gift_paid_upgrade"`
	SharedChatPrimePaidUpgrade *ChannelChatNotificationPrimePaidUpgradeNotice `json:"shared_chat_prime_paid_upgrade"`
	SharedChatPayItForward     *ChannelChatNotificationPayItForwardNotice     `json:"shared_chat_pay_it_forward"`
	SharedChatRaid             *ChannelChatNotificationRaidNotice             `json:"shared_chat_raid"`
	SharedChatUnraid           *ChannelChatNotificationUnraidNotice           `json:"shared_chat_unraid"`
	SharedChatAnnouncement     *ChannelChatNotificationAnnouncementNotice     `json:"shared_chat_announcement"`
	SharedChatBitsBadgeTier    *ChannelChatNotificationBitsBadgeTierNotice    `json:"shared_chat_bits_badge_tier"`
	SharedChatCharityDonation  *ChannelChatNotificationCharityDonationNotice  `json:"shared_chat_charity_donation"`
	SourceBroadcasterUserID    *string                                        `json:"source_broadcaster_user_id"`
	SourceBroadcasterUserLogin *string                                        `json:"source_broadcaster_user_login"`
	SourceBroadcasterUserName  *string                                        `json:"source_broadcaster_user_name"`
	SourceMessageID            *string                                        `json:"source_message_id"`
	SourceBadges               []ChatBadge                                    `json:"source_badges"`
	IsSourceOnly               *bool                                          `json:"is_source_only"`
}

// ChannelChatSettingsUpdateEvent is emitted for channel.chat_settings.update version 1 subscriptions.
type ChannelChatSettingsUpdateEvent struct {
	BroadcasterUserID           string `json:"broadcaster_user_id"`
	BroadcasterUserLogin        string `json:"broadcaster_user_login"`
	BroadcasterUserName         string `json:"broadcaster_user_name"`
	EmoteMode                   bool   `json:"emote_mode"`
	FollowerMode                bool   `json:"follower_mode"`
	FollowerModeDurationMinutes *int   `json:"follower_mode_duration_minutes"`
	SlowMode                    bool   `json:"slow_mode"`
	SlowModeWaitTimeSeconds     *int   `json:"slow_mode_wait_time_seconds"`
	SubscriberMode              bool   `json:"subscriber_mode"`
	UniqueChatMode              bool   `json:"unique_chat_mode"`
}

// ChannelChatUserMessageHoldEvent is emitted for channel.chat.user_message_hold version 1 subscriptions.
type ChannelChatUserMessageHoldEvent struct {
	BroadcasterUserID    string      `json:"broadcaster_user_id"`
	BroadcasterUserLogin string      `json:"broadcaster_user_login"`
	BroadcasterUserName  string      `json:"broadcaster_user_name"`
	UserID               string      `json:"user_id"`
	UserLogin            string      `json:"user_login"`
	UserName             string      `json:"user_name"`
	MessageID            string      `json:"message_id"`
	Message              ChatMessage `json:"message"`
}

// ChannelChatUserMessageUpdateEvent is emitted for channel.chat.user_message_update version 1 subscriptions.
type ChannelChatUserMessageUpdateEvent struct {
	BroadcasterUserID    string      `json:"broadcaster_user_id"`
	BroadcasterUserLogin string      `json:"broadcaster_user_login"`
	BroadcasterUserName  string      `json:"broadcaster_user_name"`
	UserID               string      `json:"user_id"`
	UserLogin            string      `json:"user_login"`
	UserName             string      `json:"user_name"`
	Status               string      `json:"status"`
	MessageID            string      `json:"message_id"`
	Message              ChatMessage `json:"message"`
}

// ChannelSubscribeEvent is emitted for channel.subscribe version 1 subscriptions.
type ChannelSubscribeEvent struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	Tier                 string `json:"tier"`
	IsGift               bool   `json:"is_gift"`
}

// ChannelSubscriptionEndEvent is emitted for channel.subscription.end version 1 subscriptions.
type ChannelSubscriptionEndEvent struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	Tier                 string `json:"tier"`
	IsGift               bool   `json:"is_gift"`
}

// ChannelSubscriptionGiftEvent is emitted for channel.subscription.gift version 1 subscriptions.
type ChannelSubscriptionGiftEvent struct {
	UserID               *string `json:"user_id"`
	UserLogin            *string `json:"user_login"`
	UserName             *string `json:"user_name"`
	BroadcasterUserID    string  `json:"broadcaster_user_id"`
	BroadcasterUserLogin string  `json:"broadcaster_user_login"`
	BroadcasterUserName  string  `json:"broadcaster_user_name"`
	Total                int     `json:"total"`
	Tier                 string  `json:"tier"`
	CumulativeTotal      *int    `json:"cumulative_total"`
	IsAnonymous          bool    `json:"is_anonymous"`
}

// ChannelCheerEvent is emitted for channel.cheer version 1 subscriptions.
type ChannelCheerEvent struct {
	IsAnonymous          bool    `json:"is_anonymous"`
	UserID               *string `json:"user_id"`
	UserLogin            *string `json:"user_login"`
	UserName             *string `json:"user_name"`
	BroadcasterUserID    string  `json:"broadcaster_user_id"`
	BroadcasterUserLogin string  `json:"broadcaster_user_login"`
	BroadcasterUserName  string  `json:"broadcaster_user_name"`
	Message              string  `json:"message"`
	Bits                 int     `json:"bits"`
}

// ChannelSubscriptionMessageEmote describes an emote in a resubscription chat message.
type ChannelSubscriptionMessageEmote struct {
	Begin int    `json:"begin"`
	End   int    `json:"end"`
	ID    string `json:"id"`
}

// ChannelSubscriptionMessage contains a resubscription chat message.
type ChannelSubscriptionMessage struct {
	Text   string                            `json:"text"`
	Emotes []ChannelSubscriptionMessageEmote `json:"emotes"`
}

// ChannelSubscriptionMessageEvent is emitted for channel.subscription.message version 1 subscriptions.
type ChannelSubscriptionMessageEvent struct {
	UserID               string                     `json:"user_id"`
	UserLogin            string                     `json:"user_login"`
	UserName             string                     `json:"user_name"`
	BroadcasterUserID    string                     `json:"broadcaster_user_id"`
	BroadcasterUserLogin string                     `json:"broadcaster_user_login"`
	BroadcasterUserName  string                     `json:"broadcaster_user_name"`
	Tier                 string                     `json:"tier"`
	Message              ChannelSubscriptionMessage `json:"message"`
	CumulativeMonths     int                        `json:"cumulative_months"`
	StreakMonths         *int                       `json:"streak_months"`
	DurationMonths       int                        `json:"duration_months"`
}

// ChannelGoalEvent contains the shared goal payload fields.
type ChannelGoalEvent struct {
	ID                   string    `json:"id"`
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	Type                 string    `json:"type"`
	Description          string    `json:"description"`
	CurrentAmount        int       `json:"current_amount"`
	TargetAmount         int       `json:"target_amount"`
	StartedAt            time.Time `json:"started_at"`
}

// ChannelGoalBeginEvent is emitted for channel.goal.begin version 1 subscriptions.
type ChannelGoalBeginEvent struct {
	ChannelGoalEvent
}

// ChannelGoalProgressEvent is emitted for channel.goal.progress version 1 subscriptions.
type ChannelGoalProgressEvent struct {
	ChannelGoalEvent
}

// ChannelGoalEndEvent is emitted for channel.goal.end version 1 subscriptions.
type ChannelGoalEndEvent struct {
	ChannelGoalEvent
	IsAchieved bool      `json:"is_achieved"`
	EndedAt    time.Time `json:"ended_at"`
}

// PollChoice describes one poll option.
type PollChoice struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	BitsVotes          int    `json:"bits_votes"`
	ChannelPointsVotes int    `json:"channel_points_votes"`
	Votes              int    `json:"votes"`
}

// PollVoting describes whether a paid voting mode is enabled.
type PollVoting struct {
	IsEnabled     bool `json:"is_enabled"`
	AmountPerVote int  `json:"amount_per_vote"`
}

// PollEvent contains the shared poll payload fields.
type PollEvent struct {
	ID                   string       `json:"id"`
	BroadcasterUserID    string       `json:"broadcaster_user_id"`
	BroadcasterUserLogin string       `json:"broadcaster_user_login"`
	BroadcasterUserName  string       `json:"broadcaster_user_name"`
	Title                string       `json:"title"`
	Choices              []PollChoice `json:"choices"`
	BitsVoting           PollVoting   `json:"bits_voting"`
	ChannelPointsVoting  PollVoting   `json:"channel_points_voting"`
	StartedAt            time.Time    `json:"started_at"`
}

// ChannelPollBeginEvent is emitted for channel.poll.begin version 1 subscriptions.
type ChannelPollBeginEvent struct {
	PollEvent
	EndsAt time.Time `json:"ends_at"`
}

// ChannelPollProgressEvent is emitted for channel.poll.progress version 1 subscriptions.
type ChannelPollProgressEvent struct {
	PollEvent
	EndsAt time.Time `json:"ends_at"`
}

// ChannelPollEndEvent is emitted for channel.poll.end version 1 subscriptions.
type ChannelPollEndEvent struct {
	PollEvent
	Status  string    `json:"status"`
	EndedAt time.Time `json:"ended_at"`
}

// PredictionTopPredictor describes one of the top participants in a prediction outcome.
type PredictionTopPredictor struct {
	UserName          string `json:"user_name"`
	UserLogin         string `json:"user_login"`
	UserID            string `json:"user_id"`
	ChannelPointsWon  *int   `json:"channel_points_won"`
	ChannelPointsUsed int    `json:"channel_points_used"`
}

// PredictionOutcome describes one prediction outcome.
type PredictionOutcome struct {
	ID            string                   `json:"id"`
	Title         string                   `json:"title"`
	Color         string                   `json:"color"`
	Users         int                      `json:"users"`
	ChannelPoints int                      `json:"channel_points"`
	TopPredictors []PredictionTopPredictor `json:"top_predictors"`
}

// PredictionEvent contains the shared prediction payload fields.
type PredictionEvent struct {
	ID                   string              `json:"id"`
	BroadcasterUserID    string              `json:"broadcaster_user_id"`
	BroadcasterUserLogin string              `json:"broadcaster_user_login"`
	BroadcasterUserName  string              `json:"broadcaster_user_name"`
	Title                string              `json:"title"`
	Outcomes             []PredictionOutcome `json:"outcomes"`
	StartedAt            time.Time           `json:"started_at"`
}

// ChannelPredictionBeginEvent is emitted for channel.prediction.begin version 1 subscriptions.
type ChannelPredictionBeginEvent struct {
	PredictionEvent
	LocksAt time.Time `json:"locks_at"`
}

// ChannelPredictionProgressEvent is emitted for channel.prediction.progress version 1 subscriptions.
type ChannelPredictionProgressEvent struct {
	PredictionEvent
	LocksAt time.Time `json:"locks_at"`
}

// ChannelPredictionLockEvent is emitted for channel.prediction.lock version 1 subscriptions.
type ChannelPredictionLockEvent struct {
	PredictionEvent
	LockedAt time.Time `json:"locked_at"`
}

// ChannelPredictionEndEvent is emitted for channel.prediction.end version 1 subscriptions.
type ChannelPredictionEndEvent struct {
	PredictionEvent
	WinningOutcomeID *string   `json:"winning_outcome_id"`
	Status           string    `json:"status"`
	EndedAt          time.Time `json:"ended_at"`
}

// CharityAmount describes a charity campaign monetary amount.
type CharityAmount struct {
	Value         int    `json:"value"`
	DecimalPlaces int    `json:"decimal_places"`
	Currency      string `json:"currency"`
}

// CharityCampaignEvent contains the shared charity campaign payload fields.
type CharityCampaignEvent struct {
	ID                 string `json:"id"`
	BroadcasterID      string `json:"broadcaster_id"`
	BroadcasterName    string `json:"broadcaster_name"`
	BroadcasterLogin   string `json:"broadcaster_login"`
	CharityName        string `json:"charity_name"`
	CharityDescription string `json:"charity_description"`
	CharityLogo        string `json:"charity_logo"`
	CharityWebsite     string `json:"charity_website"`
}

// ChannelCharityCampaignDonateEvent is emitted for channel.charity_campaign.donate version 1 subscriptions.
type ChannelCharityCampaignDonateEvent struct {
	ID                   string        `json:"id"`
	CampaignID           string        `json:"campaign_id"`
	BroadcasterUserID    string        `json:"broadcaster_user_id"`
	BroadcasterUserName  string        `json:"broadcaster_user_name"`
	BroadcasterUserLogin string        `json:"broadcaster_user_login"`
	UserID               string        `json:"user_id"`
	UserLogin            string        `json:"user_login"`
	UserName             string        `json:"user_name"`
	CharityName          string        `json:"charity_name"`
	CharityDescription   string        `json:"charity_description"`
	CharityLogo          string        `json:"charity_logo"`
	CharityWebsite       string        `json:"charity_website"`
	Amount               CharityAmount `json:"amount"`
}

// ChannelCharityCampaignStartEvent is emitted for channel.charity_campaign.start version 1 subscriptions.
type ChannelCharityCampaignStartEvent struct {
	CharityCampaignEvent
	CurrentAmount CharityAmount `json:"current_amount"`
	TargetAmount  CharityAmount `json:"target_amount"`
	StartedAt     time.Time     `json:"started_at"`
}

// ChannelCharityCampaignProgressEvent is emitted for channel.charity_campaign.progress version 1 subscriptions.
type ChannelCharityCampaignProgressEvent struct {
	CharityCampaignEvent
	CurrentAmount CharityAmount `json:"current_amount"`
	TargetAmount  CharityAmount `json:"target_amount"`
}

// ChannelCharityCampaignStopEvent is emitted for channel.charity_campaign.stop version 1 subscriptions.
type ChannelCharityCampaignStopEvent struct {
	CharityCampaignEvent
	CurrentAmount CharityAmount `json:"current_amount"`
	TargetAmount  CharityAmount `json:"target_amount"`
	StoppedAt     time.Time     `json:"stopped_at"`
}

// HypeTrainContribution describes a top hype train contribution.
type HypeTrainContribution struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Type      string `json:"type"`
	Total     int    `json:"total"`
}

// HypeTrainParticipant describes a shared hype train participant.
type HypeTrainParticipant struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

// HypeTrainEvent contains shared hype train payload fields.
type HypeTrainEvent struct {
	ID                      string                  `json:"id"`
	BroadcasterUserID       string                  `json:"broadcaster_user_id"`
	BroadcasterUserLogin    string                  `json:"broadcaster_user_login"`
	BroadcasterUserName     string                  `json:"broadcaster_user_name"`
	Total                   int                     `json:"total"`
	TopContributions        []HypeTrainContribution `json:"top_contributions"`
	SharedTrainParticipants []HypeTrainParticipant  `json:"shared_train_participants"`
	Level                   int                     `json:"level"`
	StartedAt               time.Time               `json:"started_at"`
	IsSharedTrain           bool                    `json:"is_shared_train"`
	Type                    string                  `json:"type"`
}

// ChannelHypeTrainBeginEvent is emitted for channel.hype_train.begin version 2 subscriptions.
type ChannelHypeTrainBeginEvent struct {
	HypeTrainEvent
	Progress         int       `json:"progress"`
	Goal             int       `json:"goal"`
	ExpiresAt        time.Time `json:"expires_at"`
	AllTimeHighLevel int       `json:"all_time_high_level"`
	AllTimeHighTotal int       `json:"all_time_high_total"`
}

// ChannelHypeTrainProgressEvent is emitted for channel.hype_train.progress version 2 subscriptions.
type ChannelHypeTrainProgressEvent struct {
	HypeTrainEvent
	Progress  int       `json:"progress"`
	Goal      int       `json:"goal"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ChannelHypeTrainEndEvent is emitted for channel.hype_train.end version 2 subscriptions.
type ChannelHypeTrainEndEvent struct {
	HypeTrainEvent
	EndedAt        time.Time `json:"ended_at"`
	CooldownEndsAt time.Time `json:"cooldown_ends_at"`
}

// SharedChatParticipant describes a broadcaster participating in a shared chat session.
type SharedChatParticipant struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
}

// SharedChatSessionEvent contains the shared session fields for shared chat events.
type SharedChatSessionEvent struct {
	SessionID                string `json:"session_id"`
	BroadcasterUserID        string `json:"broadcaster_user_id"`
	BroadcasterUserLogin     string `json:"broadcaster_user_login"`
	BroadcasterUserName      string `json:"broadcaster_user_name"`
	HostBroadcasterUserID    string `json:"host_broadcaster_user_id"`
	HostBroadcasterUserLogin string `json:"host_broadcaster_user_login"`
	HostBroadcasterUserName  string `json:"host_broadcaster_user_name"`
}

// ChannelSharedChatBeginEvent is emitted for channel.shared_chat.begin version 1 subscriptions.
type ChannelSharedChatBeginEvent struct {
	SharedChatSessionEvent
	Participants []SharedChatParticipant `json:"participants"`
}

// ChannelSharedChatUpdateEvent is emitted for channel.shared_chat.update version 1 subscriptions.
type ChannelSharedChatUpdateEvent struct {
	SharedChatSessionEvent
	Participants []SharedChatParticipant `json:"participants"`
}

// ChannelSharedChatEndEvent is emitted for channel.shared_chat.end version 1 subscriptions.
type ChannelSharedChatEndEvent struct {
	SharedChatSessionEvent
}

// ChannelShieldModeBeginEvent is emitted for channel.shield_mode.begin version 1 subscriptions.
type ChannelShieldModeBeginEvent struct {
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	ModeratorUserID      string    `json:"moderator_user_id"`
	ModeratorUserName    string    `json:"moderator_user_name"`
	ModeratorUserLogin   string    `json:"moderator_user_login"`
	StartedAt            time.Time `json:"started_at"`
}

// ChannelShieldModeEndEvent is emitted for channel.shield_mode.end version 1 subscriptions.
type ChannelShieldModeEndEvent struct {
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	ModeratorUserID      string    `json:"moderator_user_id"`
	ModeratorUserName    string    `json:"moderator_user_name"`
	ModeratorUserLogin   string    `json:"moderator_user_login"`
	EndedAt              time.Time `json:"ended_at"`
}

// ChannelShoutoutCreateEvent is emitted for channel.shoutout.create version 1 subscriptions.
type ChannelShoutoutCreateEvent struct {
	BroadcasterUserID      string    `json:"broadcaster_user_id"`
	BroadcasterUserName    string    `json:"broadcaster_user_name"`
	BroadcasterUserLogin   string    `json:"broadcaster_user_login"`
	ModeratorUserID        string    `json:"moderator_user_id"`
	ModeratorUserName      string    `json:"moderator_user_name"`
	ModeratorUserLogin     string    `json:"moderator_user_login"`
	ToBroadcasterUserID    string    `json:"to_broadcaster_user_id"`
	ToBroadcasterUserName  string    `json:"to_broadcaster_user_name"`
	ToBroadcasterUserLogin string    `json:"to_broadcaster_user_login"`
	ViewerCount            int       `json:"viewer_count"`
	StartedAt              time.Time `json:"started_at"`
	CooldownEndsAt         time.Time `json:"cooldown_ends_at"`
	TargetCooldownEndsAt   time.Time `json:"target_cooldown_ends_at"`
}

// ChannelShoutoutReceiveEvent is emitted for channel.shoutout.receive version 1 subscriptions.
type ChannelShoutoutReceiveEvent struct {
	BroadcasterUserID        string    `json:"broadcaster_user_id"`
	BroadcasterUserName      string    `json:"broadcaster_user_name"`
	BroadcasterUserLogin     string    `json:"broadcaster_user_login"`
	FromBroadcasterUserID    string    `json:"from_broadcaster_user_id"`
	FromBroadcasterUserName  string    `json:"from_broadcaster_user_name"`
	FromBroadcasterUserLogin string    `json:"from_broadcaster_user_login"`
	ViewerCount              int       `json:"viewer_count"`
	StartedAt                time.Time `json:"started_at"`
}

// ChannelWarningAcknowledgeEvent is emitted for channel.warning.acknowledge version 1 subscriptions.
type ChannelWarningAcknowledgeEvent struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
}

// ChannelWarningSendEvent is emitted for channel.warning.send version 1 subscriptions.
type ChannelWarningSendEvent struct {
	BroadcasterUserID    string   `json:"broadcaster_user_id"`
	BroadcasterUserLogin string   `json:"broadcaster_user_login"`
	BroadcasterUserName  string   `json:"broadcaster_user_name"`
	ModeratorUserID      string   `json:"moderator_user_id"`
	ModeratorUserLogin   string   `json:"moderator_user_login"`
	ModeratorUserName    string   `json:"moderator_user_name"`
	UserID               string   `json:"user_id"`
	UserLogin            string   `json:"user_login"`
	UserName             string   `json:"user_name"`
	Reason               *string  `json:"reason"`
	ChatRulesCited       []string `json:"chat_rules_cited"`
}

// ChannelUnbanRequestCreateEvent is emitted for channel.unban_request.create version 1 subscriptions.
type ChannelUnbanRequestCreateEvent struct {
	ID                   string    `json:"id"`
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	UserID               string    `json:"user_id"`
	UserLogin            string    `json:"user_login"`
	UserName             string    `json:"user_name"`
	Text                 string    `json:"text"`
	CreatedAt            time.Time `json:"created_at"`
}

// ChannelUnbanRequestResolveEvent is emitted for channel.unban_request.resolve version 1 subscriptions.
type ChannelUnbanRequestResolveEvent struct {
	ID                   string `json:"id"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	ModeratorUserID      string `json:"moderator_user_id"`
	ModeratorUserLogin   string `json:"moderator_user_login"`
	ModeratorUserName    string `json:"moderator_user_name"`
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	ResolutionText       string `json:"resolution_text"`
	Status               string `json:"status"`
}

// UserAuthorizationGrantEvent is emitted for user.authorization.grant version 1 subscriptions.
type UserAuthorizationGrantEvent struct {
	ClientID  string `json:"client_id"`
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

// UserAuthorizationRevokeEvent is emitted for user.authorization.revoke version 1 subscriptions.
type UserAuthorizationRevokeEvent struct {
	ClientID  string  `json:"client_id"`
	UserID    string  `json:"user_id"`
	UserLogin *string `json:"user_login"`
	UserName  *string `json:"user_name"`
}

// UserUpdateEvent is emitted for user.update version 1 subscriptions.
type UserUpdateEvent struct {
	UserID        string  `json:"user_id"`
	UserLogin     string  `json:"user_login"`
	UserName      string  `json:"user_name"`
	Email         *string `json:"email"`
	EmailVerified bool    `json:"email_verified"`
	Description   string  `json:"description"`
}

// Whisper contains the text of a received whisper.
type Whisper struct {
	Text string `json:"text"`
}

// UserWhisperMessageEvent is emitted for user.whisper.message version 1 subscriptions.
type UserWhisperMessageEvent struct {
	FromUserID    string  `json:"from_user_id"`
	FromUserLogin string  `json:"from_user_login"`
	FromUserName  string  `json:"from_user_name"`
	ToUserID      string  `json:"to_user_id"`
	ToUserLogin   string  `json:"to_user_login"`
	ToUserName    string  `json:"to_user_name"`
	WhisperID     string  `json:"whisper_id"`
	Whisper       Whisper `json:"whisper"`
}

// SuspiciousUserMessageCheermote describes a cheermote fragment in a suspicious user message.
type SuspiciousUserMessageCheermote struct {
	Prefix string `json:"prefix"`
	Bits   int    `json:"bits"`
	Tier   int    `json:"tier"`
}

// SuspiciousUserMessageEmote describes an emote fragment in a suspicious user message.
type SuspiciousUserMessageEmote struct {
	ID         string `json:"id"`
	EmoteSetID string `json:"emote_set_id"`
}

// SuspiciousUserMessageFragment describes one fragment in a suspicious user message.
type SuspiciousUserMessageFragment struct {
	Type      string                          `json:"type"`
	Text      string                          `json:"text"`
	Cheermote *SuspiciousUserMessageCheermote `json:"cheermote"`
	Emote     *SuspiciousUserMessageEmote     `json:"emote"`
}

// SuspiciousUserMessage contains a suspicious user's chat message.
type SuspiciousUserMessage struct {
	MessageID string                          `json:"message_id"`
	Text      string                          `json:"text"`
	Fragments []SuspiciousUserMessageFragment `json:"fragments"`
}

// ChannelSuspiciousUserUpdateEvent is emitted for channel.suspicious_user.update version 1 subscriptions.
type ChannelSuspiciousUserUpdateEvent struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	ModeratorUserID      string `json:"moderator_user_id"`
	ModeratorUserName    string `json:"moderator_user_name"`
	ModeratorUserLogin   string `json:"moderator_user_login"`
	UserID               string `json:"user_id"`
	UserName             string `json:"user_name"`
	UserLogin            string `json:"user_login"`
	LowTrustStatus       string `json:"low_trust_status"`
}

// ChannelSuspiciousUserMessageEvent is emitted for channel.suspicious_user.message version 1 subscriptions.
type ChannelSuspiciousUserMessageEvent struct {
	BroadcasterUserID    string                `json:"broadcaster_user_id"`
	BroadcasterUserName  string                `json:"broadcaster_user_name"`
	BroadcasterUserLogin string                `json:"broadcaster_user_login"`
	UserID               string                `json:"user_id"`
	UserName             string                `json:"user_name"`
	UserLogin            string                `json:"user_login"`
	LowTrustStatus       string                `json:"low_trust_status"`
	SharedBanChannelIDs  []string              `json:"shared_ban_channel_ids"`
	Types                []string              `json:"types"`
	BanEvasionEvaluation string                `json:"ban_evasion_evaluation"`
	Message              SuspiciousUserMessage `json:"message"`
}

// ChannelModeratorAddEvent is emitted for channel.moderator.add version 1 subscriptions.
type ChannelModeratorAddEvent struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

// ChannelModeratorRemoveEvent is emitted for channel.moderator.remove version 1 subscriptions.
type ChannelModeratorRemoveEvent struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

// ChannelBanEvent is emitted for channel.ban version 1 subscriptions.
type ChannelBanEvent struct {
	UserID               string     `json:"user_id"`
	UserLogin            string     `json:"user_login"`
	UserName             string     `json:"user_name"`
	BroadcasterUserID    string     `json:"broadcaster_user_id"`
	BroadcasterUserLogin string     `json:"broadcaster_user_login"`
	BroadcasterUserName  string     `json:"broadcaster_user_name"`
	ModeratorUserID      string     `json:"moderator_user_id"`
	ModeratorUserLogin   string     `json:"moderator_user_login"`
	ModeratorUserName    string     `json:"moderator_user_name"`
	Reason               string     `json:"reason"`
	BannedAt             time.Time  `json:"banned_at"`
	EndsAt               *time.Time `json:"ends_at"`
	IsPermanent          bool       `json:"is_permanent"`
}

// ChannelUnbanEvent is emitted for channel.unban version 1 subscriptions.
type ChannelUnbanEvent struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	ModeratorUserID      string `json:"moderator_user_id"`
	ModeratorUserLogin   string `json:"moderator_user_login"`
	ModeratorUserName    string `json:"moderator_user_name"`
}

// ChannelVIPAddEvent is emitted for channel.vip.add version 1 subscriptions.
type ChannelVIPAddEvent struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

// ChannelVIPRemoveEvent is emitted for channel.vip.remove version 1 subscriptions.
type ChannelVIPRemoveEvent struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

// ChannelRaidEvent is emitted for channel.raid version 1 subscriptions.
type ChannelRaidEvent struct {
	FromBroadcasterUserID    string `json:"from_broadcaster_user_id"`
	FromBroadcasterUserLogin string `json:"from_broadcaster_user_login"`
	FromBroadcasterUserName  string `json:"from_broadcaster_user_name"`
	ToBroadcasterUserID      string `json:"to_broadcaster_user_id"`
	ToBroadcasterUserLogin   string `json:"to_broadcaster_user_login"`
	ToBroadcasterUserName    string `json:"to_broadcaster_user_name"`
	Viewers                  int    `json:"viewers"`
}

func parseJSONInt(raw json.RawMessage) (int, error) {
	if len(raw) == 0 {
		return 0, fmt.Errorf("missing integer")
	}

	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, nil
	}

	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, err
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func parseJSONBool(raw json.RawMessage) (bool, error) {
	if len(raw) == 0 {
		return false, fmt.Errorf("missing boolean")
	}

	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return b, nil
	}

	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return false, err
	}

	value, err := strconv.ParseBool(s)
	if err != nil {
		return false, err
	}
	return value, nil
}
