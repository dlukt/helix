package helix

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
)

// EventSubService provides access to EventSub subscription APIs.
type EventSubService struct {
	client *Client
}

// EventSubCondition contains subscription condition fields.
type EventSubCondition map[string]string

// ChannelFollowV2Condition targets channel.follow version 2 subscriptions.
type ChannelFollowV2Condition struct {
	BroadcasterUserID string `json:"broadcaster_user_id,omitempty"`
	ModeratorUserID   string `json:"moderator_user_id,omitempty"`
}

// StreamOnlineV1Condition targets stream.online version 1 subscriptions.
type StreamOnlineV1Condition struct {
	BroadcasterUserID string `json:"broadcaster_user_id,omitempty"`
}

// BroadcasterIDCondition targets subscriptions scoped by broadcaster_id.
type BroadcasterIDCondition struct {
	BroadcasterID string `json:"broadcaster_id,omitempty"`
}

// BroadcasterUserIDCondition targets subscriptions scoped by broadcaster_user_id.
type BroadcasterUserIDCondition struct {
	BroadcasterUserID string `json:"broadcaster_user_id,omitempty"`
}

// BroadcasterModeratorUserIDCondition targets subscriptions scoped by broadcaster and moderator IDs.
type BroadcasterModeratorUserIDCondition struct {
	BroadcasterUserID string `json:"broadcaster_user_id,omitempty"`
	ModeratorUserID   string `json:"moderator_user_id,omitempty"`
}

// BroadcasterUserIDViewerCondition targets subscriptions scoped by broadcaster and user IDs.
type BroadcasterUserIDViewerCondition struct {
	BroadcasterUserID string `json:"broadcaster_user_id,omitempty"`
	UserID            string `json:"user_id,omitempty"`
}

// ClientIDCondition targets subscriptions scoped by client_id.
type ClientIDCondition struct {
	ClientID string `json:"client_id,omitempty"`
}

// UserIDCondition targets subscriptions scoped by user_id.
type UserIDCondition struct {
	UserID string `json:"user_id,omitempty"`
}

// ChannelRaidV1Condition targets channel.raid version 1 subscriptions.
type ChannelRaidV1Condition struct {
	FromBroadcasterUserID string `json:"from_broadcaster_user_id,omitempty"`
	ToBroadcasterUserID   string `json:"to_broadcaster_user_id,omitempty"`
}

// EventSubTransport describes subscription delivery transport.
type EventSubTransport struct {
	Method    string `json:"method"`
	Callback  string `json:"callback,omitempty"`
	Secret    string `json:"secret,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	ConduitID string `json:"conduit_id,omitempty"`
}

// EventSubSubscription is a subscription resource.
type EventSubSubscription struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Cost      int               `json:"cost"`
	Condition EventSubCondition `json:"condition"`
	Transport EventSubTransport `json:"transport"`
}

// CreateEventSubSubscriptionRequest creates a subscription.
type CreateEventSubSubscriptionRequest struct {
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Condition EventSubCondition `json:"condition"`
	Transport EventSubTransport `json:"transport"`
}

// CreateAutomodMessageHoldV1Request creates a typed automod.message.hold@1 subscription.
type CreateAutomodMessageHoldV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateAutomodMessageUpdateV1Request creates a typed automod.message.update@1 subscription.
type CreateAutomodMessageUpdateV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateAutomodSettingsUpdateV1Request creates a typed automod.settings.update@1 subscription.
type CreateAutomodSettingsUpdateV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateAutomodTermsUpdateV1Request creates a typed automod.terms.update@1 subscription.
type CreateAutomodTermsUpdateV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateChannelFollowV2Request creates a typed channel.follow@2 subscription.
type CreateChannelFollowV2Request struct {
	Condition ChannelFollowV2Condition
	Transport EventSubTransport
}

// CreateChannelUpdateV2Request creates a typed channel.update@2 subscription.
type CreateChannelUpdateV2Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelAdBreakBeginV1Request creates a typed channel.ad_break.begin@1 subscription.
type CreateChannelAdBreakBeginV1Request struct {
	Condition BroadcasterIDCondition
	Transport EventSubTransport
}

// CreateChannelBitsUseV1Request creates a typed channel.bits.use@1 subscription.
type CreateChannelBitsUseV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateStreamOnlineV1Request creates a typed stream.online@1 subscription.
type CreateStreamOnlineV1Request struct {
	Condition StreamOnlineV1Condition
	Transport EventSubTransport
}

// CreateStreamOfflineV1Request creates a typed stream.offline@1 subscription.
type CreateStreamOfflineV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelChatClearV1Request creates a typed channel.chat.clear@1 subscription.
type CreateChannelChatClearV1Request struct {
	Condition BroadcasterUserIDViewerCondition
	Transport EventSubTransport
}

// CreateChannelChatClearUserMessagesV1Request creates a typed channel.chat.clear_user_messages@1 subscription.
type CreateChannelChatClearUserMessagesV1Request struct {
	Condition BroadcasterUserIDViewerCondition
	Transport EventSubTransport
}

// CreateChannelChatMessageV1Request creates a typed channel.chat.message@1 subscription.
type CreateChannelChatMessageV1Request struct {
	Condition BroadcasterUserIDViewerCondition
	Transport EventSubTransport
}

// CreateChannelChatMessageDeleteV1Request creates a typed channel.chat.message_delete@1 subscription.
type CreateChannelChatMessageDeleteV1Request struct {
	Condition BroadcasterUserIDViewerCondition
	Transport EventSubTransport
}

// CreateChannelChatSettingsUpdateV1Request creates a typed channel.chat_settings.update@1 subscription.
type CreateChannelChatSettingsUpdateV1Request struct {
	Condition BroadcasterUserIDViewerCondition
	Transport EventSubTransport
}

// CreateChannelChatUserMessageHoldV1Request creates a typed channel.chat.user_message_hold@1 subscription.
type CreateChannelChatUserMessageHoldV1Request struct {
	Condition BroadcasterUserIDViewerCondition
	Transport EventSubTransport
}

// CreateChannelChatUserMessageUpdateV1Request creates a typed channel.chat.user_message_update@1 subscription.
type CreateChannelChatUserMessageUpdateV1Request struct {
	Condition BroadcasterUserIDViewerCondition
	Transport EventSubTransport
}

// CreateChannelSubscribeV1Request creates a typed channel.subscribe@1 subscription.
type CreateChannelSubscribeV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelSubscriptionGiftV1Request creates a typed channel.subscription.gift@1 subscription.
type CreateChannelSubscriptionGiftV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelSubscriptionEndV1Request creates a typed channel.subscription.end@1 subscription.
type CreateChannelSubscriptionEndV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelSubscriptionMessageV1Request creates a typed channel.subscription.message@1 subscription.
type CreateChannelSubscriptionMessageV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelGoalBeginV1Request creates a typed channel.goal.begin@1 subscription.
type CreateChannelGoalBeginV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelGoalProgressV1Request creates a typed channel.goal.progress@1 subscription.
type CreateChannelGoalProgressV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelGoalEndV1Request creates a typed channel.goal.end@1 subscription.
type CreateChannelGoalEndV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelPollBeginV1Request creates a typed channel.poll.begin@1 subscription.
type CreateChannelPollBeginV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelPollProgressV1Request creates a typed channel.poll.progress@1 subscription.
type CreateChannelPollProgressV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelPollEndV1Request creates a typed channel.poll.end@1 subscription.
type CreateChannelPollEndV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelPredictionBeginV1Request creates a typed channel.prediction.begin@1 subscription.
type CreateChannelPredictionBeginV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelPredictionProgressV1Request creates a typed channel.prediction.progress@1 subscription.
type CreateChannelPredictionProgressV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelPredictionLockV1Request creates a typed channel.prediction.lock@1 subscription.
type CreateChannelPredictionLockV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelPredictionEndV1Request creates a typed channel.prediction.end@1 subscription.
type CreateChannelPredictionEndV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelCharityCampaignDonateV1Request creates a typed channel.charity_campaign.donate@1 subscription.
type CreateChannelCharityCampaignDonateV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelCharityCampaignStartV1Request creates a typed channel.charity_campaign.start@1 subscription.
type CreateChannelCharityCampaignStartV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelCharityCampaignProgressV1Request creates a typed channel.charity_campaign.progress@1 subscription.
type CreateChannelCharityCampaignProgressV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelCharityCampaignStopV1Request creates a typed channel.charity_campaign.stop@1 subscription.
type CreateChannelCharityCampaignStopV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelHypeTrainBeginV2Request creates a typed channel.hype_train.begin@2 subscription.
type CreateChannelHypeTrainBeginV2Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelHypeTrainProgressV2Request creates a typed channel.hype_train.progress@2 subscription.
type CreateChannelHypeTrainProgressV2Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelHypeTrainEndV2Request creates a typed channel.hype_train.end@2 subscription.
type CreateChannelHypeTrainEndV2Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelSharedChatBeginV1Request creates a typed channel.shared_chat.begin@1 subscription.
type CreateChannelSharedChatBeginV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelSharedChatUpdateV1Request creates a typed channel.shared_chat.update@1 subscription.
type CreateChannelSharedChatUpdateV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelSharedChatEndV1Request creates a typed channel.shared_chat.end@1 subscription.
type CreateChannelSharedChatEndV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelShieldModeBeginV1Request creates a typed channel.shield_mode.begin@1 subscription.
type CreateChannelShieldModeBeginV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateChannelShieldModeEndV1Request creates a typed channel.shield_mode.end@1 subscription.
type CreateChannelShieldModeEndV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateChannelShoutoutCreateV1Request creates a typed channel.shoutout.create@1 subscription.
type CreateChannelShoutoutCreateV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateChannelShoutoutReceiveV1Request creates a typed channel.shoutout.receive@1 subscription.
type CreateChannelShoutoutReceiveV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateChannelWarningAcknowledgeV1Request creates a typed channel.warning.acknowledge@1 subscription.
type CreateChannelWarningAcknowledgeV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateChannelWarningSendV1Request creates a typed channel.warning.send@1 subscription.
type CreateChannelWarningSendV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateChannelUnbanRequestCreateV1Request creates a typed channel.unban_request.create@1 subscription.
type CreateChannelUnbanRequestCreateV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateChannelUnbanRequestResolveV1Request creates a typed channel.unban_request.resolve@1 subscription.
type CreateChannelUnbanRequestResolveV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateUserAuthorizationGrantV1Request creates a typed user.authorization.grant@1 subscription.
type CreateUserAuthorizationGrantV1Request struct {
	Condition ClientIDCondition
	Transport EventSubTransport
}

// CreateUserAuthorizationRevokeV1Request creates a typed user.authorization.revoke@1 subscription.
type CreateUserAuthorizationRevokeV1Request struct {
	Condition ClientIDCondition
	Transport EventSubTransport
}

// CreateUserUpdateV1Request creates a typed user.update@1 subscription.
type CreateUserUpdateV1Request struct {
	Condition UserIDCondition
	Transport EventSubTransport
}

// CreateUserWhisperMessageV1Request creates a typed user.whisper.message@1 subscription.
type CreateUserWhisperMessageV1Request struct {
	Condition UserIDCondition
	Transport EventSubTransport
}

// CreateChannelSuspiciousUserUpdateV1Request creates a typed channel.suspicious_user.update@1 subscription.
type CreateChannelSuspiciousUserUpdateV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateChannelSuspiciousUserMessageV1Request creates a typed channel.suspicious_user.message@1 subscription.
type CreateChannelSuspiciousUserMessageV1Request struct {
	Condition BroadcasterModeratorUserIDCondition
	Transport EventSubTransport
}

// CreateChannelCheerV1Request creates a typed channel.cheer@1 subscription.
type CreateChannelCheerV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelBanV1Request creates a typed channel.ban@1 subscription.
type CreateChannelBanV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelUnbanV1Request creates a typed channel.unban@1 subscription.
type CreateChannelUnbanV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelModeratorAddV1Request creates a typed channel.moderator.add@1 subscription.
type CreateChannelModeratorAddV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelModeratorRemoveV1Request creates a typed channel.moderator.remove@1 subscription.
type CreateChannelModeratorRemoveV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelVIPAddV1Request creates a typed channel.vip.add@1 subscription.
type CreateChannelVIPAddV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelVIPRemoveV1Request creates a typed channel.vip.remove@1 subscription.
type CreateChannelVIPRemoveV1Request struct {
	Condition BroadcasterUserIDCondition
	Transport EventSubTransport
}

// CreateChannelRaidV1Request creates a typed channel.raid@1 subscription.
type CreateChannelRaidV1Request struct {
	Condition ChannelRaidV1Condition
	Transport EventSubTransport
}

// CreateEventSubSubscriptionResponse is returned by Create.
type CreateEventSubSubscriptionResponse struct {
	Data []EventSubSubscription `json:"data"`
}

// ListEventSubSubscriptionsParams filters List.
type ListEventSubSubscriptionsParams struct {
	Status string
	Type   string
	After  string
}

// ListEventSubSubscriptionsResponse is returned by List.
type ListEventSubSubscriptionsResponse struct {
	Data         []EventSubSubscription `json:"data"`
	Total        int                    `json:"total"`
	TotalCost    int                    `json:"total_cost"`
	MaxTotalCost int                    `json:"max_total_cost"`
	Pagination   Pagination             `json:"pagination"`
}

// Create creates a new EventSub subscription.
func (s *EventSubService) Create(ctx context.Context, req CreateEventSubSubscriptionRequest) (*CreateEventSubSubscriptionResponse, *Response, error) {
	var resp CreateEventSubSubscriptionResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: "POST",
		Path:   "/eventsub/subscriptions",
		Body:   req,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// CreateAutomodMessageHoldV1 creates a typed automod.message.hold version 1 subscription.
func (s *EventSubService) CreateAutomodMessageHoldV1(ctx context.Context, req CreateAutomodMessageHoldV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "automod.message.hold",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateAutomodMessageUpdateV1 creates a typed automod.message.update version 1 subscription.
func (s *EventSubService) CreateAutomodMessageUpdateV1(ctx context.Context, req CreateAutomodMessageUpdateV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "automod.message.update",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateAutomodSettingsUpdateV1 creates a typed automod.settings.update version 1 subscription.
func (s *EventSubService) CreateAutomodSettingsUpdateV1(ctx context.Context, req CreateAutomodSettingsUpdateV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "automod.settings.update",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateAutomodTermsUpdateV1 creates a typed automod.terms.update version 1 subscription.
func (s *EventSubService) CreateAutomodTermsUpdateV1(ctx context.Context, req CreateAutomodTermsUpdateV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "automod.terms.update",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelFollowV2 creates a typed channel.follow version 2 subscription.
func (s *EventSubService) CreateChannelFollowV2(ctx context.Context, req CreateChannelFollowV2Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.follow",
		Version:   "2",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelUpdateV2 creates a typed channel.update version 2 subscription.
func (s *EventSubService) CreateChannelUpdateV2(ctx context.Context, req CreateChannelUpdateV2Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.update",
		Version:   "2",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelAdBreakBeginV1 creates a typed channel.ad_break.begin version 1 subscription.
func (s *EventSubService) CreateChannelAdBreakBeginV1(ctx context.Context, req CreateChannelAdBreakBeginV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.ad_break.begin",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelBitsUseV1 creates a typed channel.bits.use version 1 subscription.
func (s *EventSubService) CreateChannelBitsUseV1(ctx context.Context, req CreateChannelBitsUseV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.bits.use",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateStreamOnlineV1 creates a typed stream.online version 1 subscription.
func (s *EventSubService) CreateStreamOnlineV1(ctx context.Context, req CreateStreamOnlineV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "stream.online",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateStreamOfflineV1 creates a typed stream.offline version 1 subscription.
func (s *EventSubService) CreateStreamOfflineV1(ctx context.Context, req CreateStreamOfflineV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "stream.offline",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelChatClearV1 creates a typed channel.chat.clear version 1 subscription.
func (s *EventSubService) CreateChannelChatClearV1(ctx context.Context, req CreateChannelChatClearV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.chat.clear",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelChatClearUserMessagesV1 creates a typed channel.chat.clear_user_messages version 1 subscription.
func (s *EventSubService) CreateChannelChatClearUserMessagesV1(ctx context.Context, req CreateChannelChatClearUserMessagesV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.chat.clear_user_messages",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelChatMessageV1 creates a typed channel.chat.message version 1 subscription.
func (s *EventSubService) CreateChannelChatMessageV1(ctx context.Context, req CreateChannelChatMessageV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.chat.message",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelChatMessageDeleteV1 creates a typed channel.chat.message_delete version 1 subscription.
func (s *EventSubService) CreateChannelChatMessageDeleteV1(ctx context.Context, req CreateChannelChatMessageDeleteV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.chat.message_delete",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelChatSettingsUpdateV1 creates a typed channel.chat_settings.update version 1 subscription.
func (s *EventSubService) CreateChannelChatSettingsUpdateV1(ctx context.Context, req CreateChannelChatSettingsUpdateV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.chat_settings.update",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelChatUserMessageHoldV1 creates a typed channel.chat.user_message_hold version 1 subscription.
func (s *EventSubService) CreateChannelChatUserMessageHoldV1(ctx context.Context, req CreateChannelChatUserMessageHoldV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.chat.user_message_hold",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelChatUserMessageUpdateV1 creates a typed channel.chat.user_message_update version 1 subscription.
func (s *EventSubService) CreateChannelChatUserMessageUpdateV1(ctx context.Context, req CreateChannelChatUserMessageUpdateV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.chat.user_message_update",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSubscribeV1 creates a typed channel.subscribe version 1 subscription.
func (s *EventSubService) CreateChannelSubscribeV1(ctx context.Context, req CreateChannelSubscribeV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.subscribe",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSubscriptionGiftV1 creates a typed channel.subscription.gift version 1 subscription.
func (s *EventSubService) CreateChannelSubscriptionGiftV1(ctx context.Context, req CreateChannelSubscriptionGiftV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.subscription.gift",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSubscriptionEndV1 creates a typed channel.subscription.end version 1 subscription.
func (s *EventSubService) CreateChannelSubscriptionEndV1(ctx context.Context, req CreateChannelSubscriptionEndV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.subscription.end",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSubscriptionMessageV1 creates a typed channel.subscription.message version 1 subscription.
func (s *EventSubService) CreateChannelSubscriptionMessageV1(ctx context.Context, req CreateChannelSubscriptionMessageV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.subscription.message",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelGoalBeginV1 creates a typed channel.goal.begin version 1 subscription.
func (s *EventSubService) CreateChannelGoalBeginV1(ctx context.Context, req CreateChannelGoalBeginV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.goal.begin",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelGoalProgressV1 creates a typed channel.goal.progress version 1 subscription.
func (s *EventSubService) CreateChannelGoalProgressV1(ctx context.Context, req CreateChannelGoalProgressV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.goal.progress",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelGoalEndV1 creates a typed channel.goal.end version 1 subscription.
func (s *EventSubService) CreateChannelGoalEndV1(ctx context.Context, req CreateChannelGoalEndV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.goal.end",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelPollBeginV1 creates a typed channel.poll.begin version 1 subscription.
func (s *EventSubService) CreateChannelPollBeginV1(ctx context.Context, req CreateChannelPollBeginV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.poll.begin",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelPollProgressV1 creates a typed channel.poll.progress version 1 subscription.
func (s *EventSubService) CreateChannelPollProgressV1(ctx context.Context, req CreateChannelPollProgressV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.poll.progress",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelPollEndV1 creates a typed channel.poll.end version 1 subscription.
func (s *EventSubService) CreateChannelPollEndV1(ctx context.Context, req CreateChannelPollEndV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.poll.end",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelPredictionBeginV1 creates a typed channel.prediction.begin version 1 subscription.
func (s *EventSubService) CreateChannelPredictionBeginV1(ctx context.Context, req CreateChannelPredictionBeginV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.prediction.begin",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelPredictionProgressV1 creates a typed channel.prediction.progress version 1 subscription.
func (s *EventSubService) CreateChannelPredictionProgressV1(ctx context.Context, req CreateChannelPredictionProgressV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.prediction.progress",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelPredictionLockV1 creates a typed channel.prediction.lock version 1 subscription.
func (s *EventSubService) CreateChannelPredictionLockV1(ctx context.Context, req CreateChannelPredictionLockV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.prediction.lock",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelPredictionEndV1 creates a typed channel.prediction.end version 1 subscription.
func (s *EventSubService) CreateChannelPredictionEndV1(ctx context.Context, req CreateChannelPredictionEndV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.prediction.end",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelCharityCampaignDonateV1 creates a typed channel.charity_campaign.donate version 1 subscription.
func (s *EventSubService) CreateChannelCharityCampaignDonateV1(ctx context.Context, req CreateChannelCharityCampaignDonateV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.charity_campaign.donate",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelCharityCampaignStartV1 creates a typed channel.charity_campaign.start version 1 subscription.
func (s *EventSubService) CreateChannelCharityCampaignStartV1(ctx context.Context, req CreateChannelCharityCampaignStartV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.charity_campaign.start",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelCharityCampaignProgressV1 creates a typed channel.charity_campaign.progress version 1 subscription.
func (s *EventSubService) CreateChannelCharityCampaignProgressV1(ctx context.Context, req CreateChannelCharityCampaignProgressV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.charity_campaign.progress",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelCharityCampaignStopV1 creates a typed channel.charity_campaign.stop version 1 subscription.
func (s *EventSubService) CreateChannelCharityCampaignStopV1(ctx context.Context, req CreateChannelCharityCampaignStopV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.charity_campaign.stop",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelHypeTrainBeginV2 creates a typed channel.hype_train.begin version 2 subscription.
func (s *EventSubService) CreateChannelHypeTrainBeginV2(ctx context.Context, req CreateChannelHypeTrainBeginV2Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.hype_train.begin",
		Version:   "2",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelHypeTrainProgressV2 creates a typed channel.hype_train.progress version 2 subscription.
func (s *EventSubService) CreateChannelHypeTrainProgressV2(ctx context.Context, req CreateChannelHypeTrainProgressV2Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.hype_train.progress",
		Version:   "2",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelHypeTrainEndV2 creates a typed channel.hype_train.end version 2 subscription.
func (s *EventSubService) CreateChannelHypeTrainEndV2(ctx context.Context, req CreateChannelHypeTrainEndV2Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.hype_train.end",
		Version:   "2",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSharedChatBeginV1 creates a typed channel.shared_chat.begin version 1 subscription.
func (s *EventSubService) CreateChannelSharedChatBeginV1(ctx context.Context, req CreateChannelSharedChatBeginV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.shared_chat.begin",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSharedChatUpdateV1 creates a typed channel.shared_chat.update version 1 subscription.
func (s *EventSubService) CreateChannelSharedChatUpdateV1(ctx context.Context, req CreateChannelSharedChatUpdateV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.shared_chat.update",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSharedChatEndV1 creates a typed channel.shared_chat.end version 1 subscription.
func (s *EventSubService) CreateChannelSharedChatEndV1(ctx context.Context, req CreateChannelSharedChatEndV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.shared_chat.end",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelShieldModeBeginV1 creates a typed channel.shield_mode.begin version 1 subscription.
func (s *EventSubService) CreateChannelShieldModeBeginV1(ctx context.Context, req CreateChannelShieldModeBeginV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.shield_mode.begin",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelShieldModeEndV1 creates a typed channel.shield_mode.end version 1 subscription.
func (s *EventSubService) CreateChannelShieldModeEndV1(ctx context.Context, req CreateChannelShieldModeEndV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.shield_mode.end",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelShoutoutCreateV1 creates a typed channel.shoutout.create version 1 subscription.
func (s *EventSubService) CreateChannelShoutoutCreateV1(ctx context.Context, req CreateChannelShoutoutCreateV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.shoutout.create",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelShoutoutReceiveV1 creates a typed channel.shoutout.receive version 1 subscription.
func (s *EventSubService) CreateChannelShoutoutReceiveV1(ctx context.Context, req CreateChannelShoutoutReceiveV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.shoutout.receive",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelWarningAcknowledgeV1 creates a typed channel.warning.acknowledge version 1 subscription.
func (s *EventSubService) CreateChannelWarningAcknowledgeV1(ctx context.Context, req CreateChannelWarningAcknowledgeV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.warning.acknowledge",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelWarningSendV1 creates a typed channel.warning.send version 1 subscription.
func (s *EventSubService) CreateChannelWarningSendV1(ctx context.Context, req CreateChannelWarningSendV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.warning.send",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelUnbanRequestCreateV1 creates a typed channel.unban_request.create version 1 subscription.
func (s *EventSubService) CreateChannelUnbanRequestCreateV1(ctx context.Context, req CreateChannelUnbanRequestCreateV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.unban_request.create",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelUnbanRequestResolveV1 creates a typed channel.unban_request.resolve version 1 subscription.
func (s *EventSubService) CreateChannelUnbanRequestResolveV1(ctx context.Context, req CreateChannelUnbanRequestResolveV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.unban_request.resolve",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateUserAuthorizationGrantV1 creates a typed user.authorization.grant version 1 subscription.
func (s *EventSubService) CreateUserAuthorizationGrantV1(ctx context.Context, req CreateUserAuthorizationGrantV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "user.authorization.grant",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateUserAuthorizationRevokeV1 creates a typed user.authorization.revoke version 1 subscription.
func (s *EventSubService) CreateUserAuthorizationRevokeV1(ctx context.Context, req CreateUserAuthorizationRevokeV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "user.authorization.revoke",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateUserUpdateV1 creates a typed user.update version 1 subscription.
func (s *EventSubService) CreateUserUpdateV1(ctx context.Context, req CreateUserUpdateV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "user.update",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateUserWhisperMessageV1 creates a typed user.whisper.message version 1 subscription.
func (s *EventSubService) CreateUserWhisperMessageV1(ctx context.Context, req CreateUserWhisperMessageV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "user.whisper.message",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSuspiciousUserUpdateV1 creates a typed channel.suspicious_user.update version 1 subscription.
func (s *EventSubService) CreateChannelSuspiciousUserUpdateV1(ctx context.Context, req CreateChannelSuspiciousUserUpdateV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.suspicious_user.update",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelSuspiciousUserMessageV1 creates a typed channel.suspicious_user.message version 1 subscription.
func (s *EventSubService) CreateChannelSuspiciousUserMessageV1(ctx context.Context, req CreateChannelSuspiciousUserMessageV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.suspicious_user.message",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelCheerV1 creates a typed channel.cheer version 1 subscription.
func (s *EventSubService) CreateChannelCheerV1(ctx context.Context, req CreateChannelCheerV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.cheer",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelBanV1 creates a typed channel.ban version 1 subscription.
func (s *EventSubService) CreateChannelBanV1(ctx context.Context, req CreateChannelBanV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.ban",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelUnbanV1 creates a typed channel.unban version 1 subscription.
func (s *EventSubService) CreateChannelUnbanV1(ctx context.Context, req CreateChannelUnbanV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.unban",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelModeratorAddV1 creates a typed channel.moderator.add version 1 subscription.
func (s *EventSubService) CreateChannelModeratorAddV1(ctx context.Context, req CreateChannelModeratorAddV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.moderator.add",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelModeratorRemoveV1 creates a typed channel.moderator.remove version 1 subscription.
func (s *EventSubService) CreateChannelModeratorRemoveV1(ctx context.Context, req CreateChannelModeratorRemoveV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.moderator.remove",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelVIPAddV1 creates a typed channel.vip.add version 1 subscription.
func (s *EventSubService) CreateChannelVIPAddV1(ctx context.Context, req CreateChannelVIPAddV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.vip.add",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelVIPRemoveV1 creates a typed channel.vip.remove version 1 subscription.
func (s *EventSubService) CreateChannelVIPRemoveV1(ctx context.Context, req CreateChannelVIPRemoveV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.vip.remove",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// CreateChannelRaidV1 creates a typed channel.raid version 1 subscription.
func (s *EventSubService) CreateChannelRaidV1(ctx context.Context, req CreateChannelRaidV1Request) (*CreateEventSubSubscriptionResponse, *Response, error) {
	if req.Condition.FromBroadcasterUserID == "" && req.Condition.ToBroadcasterUserID == "" {
		return nil, nil, errors.New("helix: channel.raid requires exactly one of from_broadcaster_user_id or to_broadcaster_user_id")
	}
	if req.Condition.FromBroadcasterUserID != "" && req.Condition.ToBroadcasterUserID != "" {
		return nil, nil, errors.New("helix: channel.raid requires exactly one of from_broadcaster_user_id or to_broadcaster_user_id")
	}

	condition, err := marshalCondition(req.Condition)
	if err != nil {
		return nil, nil, err
	}
	return s.Create(ctx, CreateEventSubSubscriptionRequest{
		Type:      "channel.raid",
		Version:   "1",
		Condition: condition,
		Transport: req.Transport,
	})
}

// List lists EventSub subscriptions.
func (s *EventSubService) List(ctx context.Context, params ListEventSubSubscriptionsParams) (*ListEventSubSubscriptionsResponse, *Response, error) {
	query := url.Values{}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.Type != "" {
		query.Set("type", params.Type)
	}
	if params.After != "" {
		query.Set("after", params.After)
	}

	var resp ListEventSubSubscriptionsResponse
	meta, err := s.client.Do(ctx, RawRequest{
		Method: "GET",
		Path:   "/eventsub/subscriptions",
		Query:  query,
	}, &resp)
	if err != nil {
		return nil, meta, err
	}
	return &resp, meta, nil
}

// Delete deletes an EventSub subscription by ID.
func (s *EventSubService) Delete(ctx context.Context, id string) (*Response, error) {
	query := url.Values{}
	query.Set("id", id)
	return s.client.Do(ctx, RawRequest{
		Method: "DELETE",
		Path:   "/eventsub/subscriptions",
		Query:  query,
	}, nil)
}

func marshalCondition(value any) (EventSubCondition, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var condition EventSubCondition
	if err := json.Unmarshal(raw, &condition); err != nil {
		return nil, err
	}
	return condition, nil
}
