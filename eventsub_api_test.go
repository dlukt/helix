package helix_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dlukt/helix"
	"github.com/dlukt/helix/oauth"
)

func TestEventSubServiceCreateListAndDeleteSubscriptions(t *testing.T) {
	t.Parallel()

	var createSeen, listSeen, deleteSeen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/eventsub/subscriptions":
			createSeen = true
			var req helix.CreateEventSubSubscriptionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got := req.Type; got != "channel.follow" {
				t.Fatalf("Type = %q, want %q", got, "channel.follow")
			}
			if got := req.Transport.Method; got != "webhook" {
				t.Fatalf("Transport.Method = %q, want %q", got, "webhook")
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":      "sub-1",
						"status":  "enabled",
						"type":    "channel.follow",
						"version": "2",
						"cost":    1,
						"condition": map[string]any{
							"broadcaster_user_id": "123",
							"moderator_user_id":   "456",
						},
						"transport": map[string]any{
							"method":   "webhook",
							"callback": "https://example.com/eventsub",
						},
					},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/eventsub/subscriptions":
			listSeen = true
			if got := r.URL.Query().Get("type"); got != "channel.follow" {
				t.Fatalf("type = %q, want %q", got, "channel.follow")
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":      "sub-1",
						"status":  "enabled",
						"type":    "channel.follow",
						"version": "2",
						"cost":    1,
						"transport": map[string]any{
							"method":   "webhook",
							"callback": "https://example.com/eventsub",
						},
					},
				},
				"pagination": map[string]any{
					"cursor": "next-page",
				},
				"total":          1,
				"total_cost":     1,
				"max_total_cost": 10000,
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/eventsub/subscriptions":
			deleteSeen = true
			if got := r.URL.Query().Get("id"); got != "sub-1" {
				t.Fatalf("id = %q, want %q", got, "sub-1")
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
		TokenSource: oauth.StaticSource{
			Value: oauth.Token{AccessToken: "access-token"},
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	created, _, err := client.EventSub.Create(context.Background(), helix.CreateEventSubSubscriptionRequest{
		Type:    "channel.follow",
		Version: "2",
		Condition: helix.EventSubCondition{
			"broadcaster_user_id": "123",
			"moderator_user_id":   "456",
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: "https://example.com/eventsub",
			Secret:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got := created.Data[0].ID; got != "sub-1" {
		t.Fatalf("created.Data[0].ID = %q, want %q", got, "sub-1")
	}

	listed, meta, err := client.EventSub.List(context.Background(), helix.ListEventSubSubscriptionsParams{
		Type: "channel.follow",
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if got := listed.Total; got != 1 {
		t.Fatalf("Total = %d, want 1", got)
	}
	if got := meta.Pagination.Cursor; got != "next-page" {
		t.Fatalf("meta.Pagination.Cursor = %q, want %q", got, "next-page")
	}

	meta, err = client.EventSub.Delete(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if got := meta.StatusCode; got != http.StatusNoContent {
		t.Fatalf("StatusCode = %d, want %d", got, http.StatusNoContent)
	}
	if !createSeen || !listSeen || !deleteSeen {
		t.Fatalf("createSeen=%t listSeen=%t deleteSeen=%t, want all true", createSeen, listSeen, deleteSeen)
	}
}

func TestEventSubServiceCreateTypedSubscriptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		wantType      string
		wantVersion   string
		wantCondition helix.EventSubCondition
		create        func(context.Context, *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error)
	}{
		{
			name:        "channel follow",
			wantType:    "channel.follow",
			wantVersion: "2",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
				"moderator_user_id":   "456",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelFollowV2(ctx, helix.CreateChannelFollowV2Request{
					Condition: helix.ChannelFollowV2Condition{
						BroadcasterUserID: "123",
						ModeratorUserID:   "456",
					},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel update",
			wantType:    "channel.update",
			wantVersion: "2",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelUpdateV2(ctx, helix.CreateChannelUpdateV2Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "stream online",
			wantType:    "stream.online",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateStreamOnlineV1(ctx, helix.CreateStreamOnlineV1Request{
					Condition: helix.StreamOnlineV1Condition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "stream offline",
			wantType:    "stream.offline",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateStreamOfflineV1(ctx, helix.CreateStreamOfflineV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel subscribe",
			wantType:    "channel.subscribe",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelSubscribeV1(ctx, helix.CreateChannelSubscribeV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel subscription gift",
			wantType:    "channel.subscription.gift",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelSubscriptionGiftV1(ctx, helix.CreateChannelSubscriptionGiftV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel subscription end",
			wantType:    "channel.subscription.end",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelSubscriptionEndV1(ctx, helix.CreateChannelSubscriptionEndV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel subscription message",
			wantType:    "channel.subscription.message",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelSubscriptionMessageV1(ctx, helix.CreateChannelSubscriptionMessageV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel goal begin",
			wantType:    "channel.goal.begin",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelGoalBeginV1(ctx, helix.CreateChannelGoalBeginV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel goal progress",
			wantType:    "channel.goal.progress",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelGoalProgressV1(ctx, helix.CreateChannelGoalProgressV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel goal end",
			wantType:    "channel.goal.end",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelGoalEndV1(ctx, helix.CreateChannelGoalEndV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel poll begin",
			wantType:    "channel.poll.begin",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelPollBeginV1(ctx, helix.CreateChannelPollBeginV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel poll progress",
			wantType:    "channel.poll.progress",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelPollProgressV1(ctx, helix.CreateChannelPollProgressV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel poll end",
			wantType:    "channel.poll.end",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelPollEndV1(ctx, helix.CreateChannelPollEndV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel prediction begin",
			wantType:    "channel.prediction.begin",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelPredictionBeginV1(ctx, helix.CreateChannelPredictionBeginV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel prediction progress",
			wantType:    "channel.prediction.progress",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelPredictionProgressV1(ctx, helix.CreateChannelPredictionProgressV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel prediction lock",
			wantType:    "channel.prediction.lock",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelPredictionLockV1(ctx, helix.CreateChannelPredictionLockV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel prediction end",
			wantType:    "channel.prediction.end",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelPredictionEndV1(ctx, helix.CreateChannelPredictionEndV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel charity campaign donate",
			wantType:    "channel.charity_campaign.donate",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelCharityCampaignDonateV1(ctx, helix.CreateChannelCharityCampaignDonateV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel charity campaign start",
			wantType:    "channel.charity_campaign.start",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelCharityCampaignStartV1(ctx, helix.CreateChannelCharityCampaignStartV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel charity campaign progress",
			wantType:    "channel.charity_campaign.progress",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelCharityCampaignProgressV1(ctx, helix.CreateChannelCharityCampaignProgressV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel charity campaign stop",
			wantType:    "channel.charity_campaign.stop",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelCharityCampaignStopV1(ctx, helix.CreateChannelCharityCampaignStopV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel hype train begin",
			wantType:    "channel.hype_train.begin",
			wantVersion: "2",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelHypeTrainBeginV2(ctx, helix.CreateChannelHypeTrainBeginV2Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel hype train progress",
			wantType:    "channel.hype_train.progress",
			wantVersion: "2",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelHypeTrainProgressV2(ctx, helix.CreateChannelHypeTrainProgressV2Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel hype train end",
			wantType:    "channel.hype_train.end",
			wantVersion: "2",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelHypeTrainEndV2(ctx, helix.CreateChannelHypeTrainEndV2Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel shared chat begin",
			wantType:    "channel.shared_chat.begin",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelSharedChatBeginV1(ctx, helix.CreateChannelSharedChatBeginV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel shared chat update",
			wantType:    "channel.shared_chat.update",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelSharedChatUpdateV1(ctx, helix.CreateChannelSharedChatUpdateV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel shared chat end",
			wantType:    "channel.shared_chat.end",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelSharedChatEndV1(ctx, helix.CreateChannelSharedChatEndV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel shield mode begin",
			wantType:    "channel.shield_mode.begin",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
				"moderator_user_id":   "456",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelShieldModeBeginV1(ctx, helix.CreateChannelShieldModeBeginV1Request{
					Condition: helix.BroadcasterModeratorUserIDCondition{
						BroadcasterUserID: "123",
						ModeratorUserID:   "456",
					},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel shield mode end",
			wantType:    "channel.shield_mode.end",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
				"moderator_user_id":   "456",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelShieldModeEndV1(ctx, helix.CreateChannelShieldModeEndV1Request{
					Condition: helix.BroadcasterModeratorUserIDCondition{
						BroadcasterUserID: "123",
						ModeratorUserID:   "456",
					},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel shoutout create",
			wantType:    "channel.shoutout.create",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
				"moderator_user_id":   "456",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelShoutoutCreateV1(ctx, helix.CreateChannelShoutoutCreateV1Request{
					Condition: helix.BroadcasterModeratorUserIDCondition{
						BroadcasterUserID: "123",
						ModeratorUserID:   "456",
					},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel shoutout receive",
			wantType:    "channel.shoutout.receive",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
				"moderator_user_id":   "456",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelShoutoutReceiveV1(ctx, helix.CreateChannelShoutoutReceiveV1Request{
					Condition: helix.BroadcasterModeratorUserIDCondition{
						BroadcasterUserID: "123",
						ModeratorUserID:   "456",
					},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel warning acknowledge",
			wantType:    "channel.warning.acknowledge",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
				"moderator_user_id":   "456",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelWarningAcknowledgeV1(ctx, helix.CreateChannelWarningAcknowledgeV1Request{
					Condition: helix.BroadcasterModeratorUserIDCondition{
						BroadcasterUserID: "123",
						ModeratorUserID:   "456",
					},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel warning send",
			wantType:    "channel.warning.send",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
				"moderator_user_id":   "456",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelWarningSendV1(ctx, helix.CreateChannelWarningSendV1Request{
					Condition: helix.BroadcasterModeratorUserIDCondition{
						BroadcasterUserID: "123",
						ModeratorUserID:   "456",
					},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel suspicious user update",
			wantType:    "channel.suspicious_user.update",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
				"moderator_user_id":   "456",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelSuspiciousUserUpdateV1(ctx, helix.CreateChannelSuspiciousUserUpdateV1Request{
					Condition: helix.BroadcasterModeratorUserIDCondition{
						BroadcasterUserID: "123",
						ModeratorUserID:   "456",
					},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel suspicious user message",
			wantType:    "channel.suspicious_user.message",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
				"moderator_user_id":   "456",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelSuspiciousUserMessageV1(ctx, helix.CreateChannelSuspiciousUserMessageV1Request{
					Condition: helix.BroadcasterModeratorUserIDCondition{
						BroadcasterUserID: "123",
						ModeratorUserID:   "456",
					},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel cheer",
			wantType:    "channel.cheer",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelCheerV1(ctx, helix.CreateChannelCheerV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel ban",
			wantType:    "channel.ban",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelBanV1(ctx, helix.CreateChannelBanV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel unban",
			wantType:    "channel.unban",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelUnbanV1(ctx, helix.CreateChannelUnbanV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel moderator add",
			wantType:    "channel.moderator.add",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelModeratorAddV1(ctx, helix.CreateChannelModeratorAddV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel moderator remove",
			wantType:    "channel.moderator.remove",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelModeratorRemoveV1(ctx, helix.CreateChannelModeratorRemoveV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel vip add",
			wantType:    "channel.vip.add",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelVIPAddV1(ctx, helix.CreateChannelVIPAddV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel vip remove",
			wantType:    "channel.vip.remove",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelVIPRemoveV1(ctx, helix.CreateChannelVIPRemoveV1Request{
					Condition: helix.BroadcasterUserIDCondition{BroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
		{
			name:        "channel raid",
			wantType:    "channel.raid",
			wantVersion: "1",
			wantCondition: helix.EventSubCondition{
				"to_broadcaster_user_id": "123",
			},
			create: func(ctx context.Context, client *helix.Client) (*helix.CreateEventSubSubscriptionResponse, *helix.Response, error) {
				return client.EventSub.CreateChannelRaidV1(ctx, helix.CreateChannelRaidV1Request{
					Condition: helix.ChannelRaidV1Condition{ToBroadcasterUserID: "123"},
					Transport: helix.EventSubTransport{
						Method:   "webhook",
						Callback: "https://example.com/eventsub",
						Secret:   "secret",
					},
				})
			},
		},
	}

	var requests []helix.CreateEventSubSubscriptionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		var req helix.CreateEventSubSubscriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		requests = append(requests, req)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "sub-1", "status": "enabled", "type": req.Type, "version": req.Version},
			},
		})
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for _, tt := range tests {
		resp, _, err := tt.create(context.Background(), client)
		if err != nil {
			t.Fatalf("%s: create error = %v", tt.name, err)
		}
		if got := resp.Data[0].ID; got != "sub-1" {
			t.Fatalf("%s: resp.Data[0].ID = %q, want %q", tt.name, got, "sub-1")
		}
	}

	if got, want := len(requests), len(tests); got != want {
		t.Fatalf("len(requests) = %d, want %d", got, want)
	}
	for i, tt := range tests {
		req := requests[i]
		if got := req.Type; got != tt.wantType {
			t.Fatalf("%s: Type = %q, want %q", tt.name, got, tt.wantType)
		}
		if got := req.Version; got != tt.wantVersion {
			t.Fatalf("%s: Version = %q, want %q", tt.name, got, tt.wantVersion)
		}
		if got := len(req.Condition); got != len(tt.wantCondition) {
			t.Fatalf("%s: len(Condition) = %d, want %d", tt.name, got, len(tt.wantCondition))
		}
		for key, want := range tt.wantCondition {
			if got := req.Condition[key]; got != want {
				t.Fatalf("%s: Condition[%q] = %q, want %q", tt.name, key, got, want)
			}
		}
	}
}

func TestEventSubServiceCreateChannelRaidV1ValidatesCondition(t *testing.T) {
	t.Parallel()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  "https://example.com",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tests := []struct {
		name      string
		condition helix.ChannelRaidV1Condition
	}{
		{
			name:      "missing both IDs",
			condition: helix.ChannelRaidV1Condition{},
		},
		{
			name: "both IDs set",
			condition: helix.ChannelRaidV1Condition{
				FromBroadcasterUserID: "123",
				ToBroadcasterUserID:   "456",
			},
		},
	}

	for _, tt := range tests {
		_, _, err := client.EventSub.CreateChannelRaidV1(context.Background(), helix.CreateChannelRaidV1Request{
			Condition: tt.condition,
			Transport: helix.EventSubTransport{Method: "webhook"},
		})
		if err == nil {
			t.Fatalf("%s: error = nil, want validation error", tt.name)
		}
		if !strings.Contains(err.Error(), "exactly one") {
			t.Fatalf("%s: error = %q, want exactly-one validation message", tt.name, err.Error())
		}
	}
}

func TestEventSubServiceCreateAndListUseSharedRetryPath(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		calls++
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/eventsub/subscriptions":
			if calls == 1 {
				if got := r.Header.Get("Authorization"); got != "Bearer stale-token" {
					t.Fatalf("Authorization on create first attempt = %q, want stale token", got)
				}
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"Unauthorized","status":401}`))
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer fresh-token" {
				t.Fatalf("Authorization on create retry = %q, want fresh token", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":      "sub-1",
						"status":  "enabled",
						"type":    "channel.follow",
						"version": "2",
						"cost":    1,
						"transport": map[string]any{
							"method":   "webhook",
							"callback": "https://example.com/eventsub",
						},
					},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/eventsub/subscriptions":
			if calls == 3 {
				if got := r.Header.Get("Authorization"); got != "Bearer fresh-token" {
					t.Fatalf("Authorization on list first attempt = %q, want fresh token", got)
				}
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"error":"service unavailable"}`))
				return
			}
			if got := r.Header.Get("Authorization"); got != "Bearer fresh-token" {
				t.Fatalf("Authorization on list retry = %q, want fresh token", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":      "sub-1",
						"status":  "enabled",
						"type":    "channel.follow",
						"version": "2",
						"cost":    1,
						"transport": map[string]any{
							"method":   "webhook",
							"callback": "https://example.com/eventsub",
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	source := &rotatingTokenSource{current: "stale-token"}
	client, err := helix.New(helix.Config{
		ClientID:    "client-id",
		BaseURL:     server.URL,
		TokenSource: source,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	created, _, err := client.EventSub.Create(context.Background(), helix.CreateEventSubSubscriptionRequest{
		Type:    "channel.follow",
		Version: "2",
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: "https://example.com/eventsub",
			Secret:   "secret",
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got := created.Data[0].ID; got != "sub-1" {
		t.Fatalf("created.Data[0].ID = %q, want %q", got, "sub-1")
	}

	listed, _, err := client.EventSub.List(context.Background(), helix.ListEventSubSubscriptionsParams{Type: "channel.follow"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if got := len(listed.Data); got != 1 {
		t.Fatalf("len(listed.Data) = %d, want 1", got)
	}
	if got := source.invalidations.Load(); got != 1 {
		t.Fatalf("invalidations = %d, want 1", got)
	}
}

func TestEventSubServiceListReturnsMetadataOnDecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		w.Header().Set("Ratelimit-Limit", "800")
		w.Header().Set("Ratelimit-Remaining", "799")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[`))
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, meta, err := client.EventSub.List(context.Background(), helix.ListEventSubSubscriptionsParams{})
	if err == nil {
		t.Fatal("List() error = nil, want decode error")
	}
	if meta == nil {
		t.Fatal("meta = nil, want response metadata on decode error")
	}
	if got := meta.StatusCode; got != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", got, http.StatusOK)
	}
	if got := meta.RateLimit.Remaining; got != 799 {
		t.Fatalf("RateLimit.Remaining = %d, want 799", got)
	}
}

func TestEventSubServiceCreateDoesNotRetry503(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		calls++
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"service unavailable"}`))
	}))
	defer server.Close()

	client, err := helix.New(helix.Config{
		ClientID: "client-id",
		BaseURL:  server.URL,
		TokenSource: oauth.StaticSource{
			Value: oauth.Token{AccessToken: "access-token"},
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, _, err = client.EventSub.Create(context.Background(), helix.CreateEventSubSubscriptionRequest{
		Type:    "channel.follow",
		Version: "2",
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: "https://example.com/eventsub",
			Secret:   "secret",
		},
	})
	if err == nil {
		t.Fatal("Create() error = nil, want APIError")
	}
	if got := calls; got != 1 {
		t.Fatalf("calls = %d, want 1", got)
	}
}
