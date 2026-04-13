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
	BroadcasterID      string `json:"broadcaster_user_id"`
	BroadcasterName    string `json:"broadcaster_user_name"`
	BroadcasterLogin   string `json:"broadcaster_user_login"`
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
