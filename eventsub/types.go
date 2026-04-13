package eventsub

import "time"

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

// StreamOnlineEvent is emitted for stream.online version 1 subscriptions.
type StreamOnlineEvent struct {
	ID                   string    `json:"id"`
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	Type                 string    `json:"type"`
	StartedAt            time.Time `json:"started_at"`
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
