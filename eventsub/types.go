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
