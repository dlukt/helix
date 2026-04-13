package eventsub_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/dlukt/helix/eventsub"
)

func TestWebhookHandlerRespondsToChallengeAndDispatchesTypedNotifications(t *testing.T) {
	t.Parallel()

	var challenges []string
	var notifications []eventsub.Notification

	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret: "super-secret",
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
		OnChallenge: func(_ context.Context, challenge eventsub.Challenge) error {
			challenges = append(challenges, challenge.Challenge)
			return nil
		},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			notifications = append(notifications, notification)
			return nil
		},
	})

	challengeBody := []byte(`{
		"challenge":"abc123",
		"subscription":{"type":"channel.follow","version":"2"}
	}`)
	challengeReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(challengeBody))
	signWebhook(challengeReq, "super-secret", "message-1", "2024-04-11T12:00:00Z", challengeBody, "webhook_callback_verification")
	challengeRec := httptest.NewRecorder()

	handler.ServeHTTP(challengeRec, challengeReq)

	if got := challengeRec.Code; got != http.StatusOK {
		t.Fatalf("challenge status = %d, want %d", got, http.StatusOK)
	}
	if got := challengeRec.Body.String(); got != "abc123" {
		t.Fatalf("challenge body = %q, want %q", got, "abc123")
	}

	notificationBody := []byte(`{
		"subscription":{"id":"sub-1","type":"channel.follow","version":"2","status":"enabled"},
		"event":{"user_id":"777","user_login":"viewer","user_name":"Viewer","broadcaster_user_id":"123","broadcaster_user_login":"caster","broadcaster_user_name":"Caster","followed_at":"2024-04-11T12:00:00Z"}
	}`)
	notificationReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(notificationBody))
	signWebhook(notificationReq, "super-secret", "message-2", "2024-04-11T12:00:01Z", notificationBody, "notification")
	notificationRec := httptest.NewRecorder()

	handler.ServeHTTP(notificationRec, notificationReq)

	if got := notificationRec.Code; got != http.StatusNoContent {
		t.Fatalf("notification status = %d, want %d", got, http.StatusNoContent)
	}
	if len(challenges) != 1 || challenges[0] != "abc123" {
		t.Fatalf("challenges = %v, want [abc123]", challenges)
	}
	if len(notifications) != 1 {
		t.Fatalf("len(notifications) = %d, want 1", len(notifications))
	}
	if got := notifications[0].Subscription.Type; got != "channel.follow" {
		t.Fatalf("Subscription.Type = %q, want %q", got, "channel.follow")
	}
	if got := notifications[0].MessageID; got != "message-2" {
		t.Fatalf("MessageID = %q, want %q", got, "message-2")
	}

	followEvent, ok := notifications[0].Event.(eventsub.ChannelFollowEvent)
	if !ok {
		t.Fatalf("Event type = %T, want ChannelFollowEvent", notifications[0].Event)
	}
	if got := followEvent.UserID; got != "777" {
		t.Fatalf("followEvent.UserID = %q, want %q", got, "777")
	}
}

func TestWebhookHandlerDispatchesDropEntitlementGrantBatchNotifications(t *testing.T) {
	t.Parallel()

	var notifications []eventsub.Notification

	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret: "super-secret",
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			notifications = append(notifications, notification)
			return nil
		},
	})

	notificationBody := []byte(`{
		"subscription":{"id":"sub-drop","type":"drop.entitlement.grant","version":"1","status":"enabled"},
		"events":[
			{
				"id":"grant-1",
				"data":{
					"organization_id":"9001",
					"category_id":"9002",
					"category_name":"Fortnite",
					"campaign_id":"9003",
					"user_id":"1234",
					"user_name":"Cool_User",
					"user_login":"cool_user",
					"entitlement_id":"ent-1",
					"benefit_id":"benefit-1",
					"created_at":"2019-01-28T04:17:53.325Z"
				}
			}
		]
	}`)
	notificationReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(notificationBody))
	signWebhook(notificationReq, "super-secret", "message-drop-1", "2024-04-11T12:00:01Z", notificationBody, "notification")
	notificationRec := httptest.NewRecorder()

	handler.ServeHTTP(notificationRec, notificationReq)

	if got := notificationRec.Code; got != http.StatusNoContent {
		t.Fatalf("notification status = %d, want %d", got, http.StatusNoContent)
	}
	if len(notifications) != 1 {
		t.Fatalf("len(notifications) = %d, want 1", len(notifications))
	}
	if got := notifications[0].Subscription.Type; got != "drop.entitlement.grant" {
		t.Fatalf("Subscription.Type = %q, want %q", got, "drop.entitlement.grant")
	}

	batch, ok := notifications[0].Event.(eventsub.DropEntitlementGrantBatch)
	if !ok {
		t.Fatalf("Event type = %T, want DropEntitlementGrantBatch", notifications[0].Event)
	}
	if got := len(batch); got != 1 {
		t.Fatalf("len(batch) = %d, want 1", got)
	}
	if batch[0].Data.CategoryName == nil || *batch[0].Data.CategoryName != "Fortnite" {
		t.Fatalf("CategoryName = %#v, want %q", batch[0].Data.CategoryName, "Fortnite")
	}
}

func TestWebhookHandlerPropagatesRevocationMessageID(t *testing.T) {
	t.Parallel()

	var revocations []eventsub.Revocation
	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret: "super-secret",
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
		OnRevocation: func(_ context.Context, revocation eventsub.Revocation) error {
			revocations = append(revocations, revocation)
			return nil
		},
	})

	body := []byte(`{
		"subscription":{"id":"sub-1","type":"stream.online","version":"1","status":"authorization_revoked"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(req, "super-secret", "message-revoke", "2024-04-11T12:00:01Z", body, "revocation")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Code; got != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", got, http.StatusNoContent)
	}
	if len(revocations) != 1 {
		t.Fatalf("len(revocations) = %d, want 1", len(revocations))
	}
	if got := revocations[0].MessageID; got != "message-revoke" {
		t.Fatalf("MessageID = %q, want %q", got, "message-revoke")
	}
}

func TestWebhookHandlerDecodesAutomodSettingsUpdateWrappedEvent(t *testing.T) {
	t.Parallel()

	var notifications []eventsub.Notification
	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret: "super-secret",
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			notifications = append(notifications, notification)
			return nil
		},
	})

	body := []byte(`{
		"subscription":{"id":"sub-automod","type":"automod.settings.update","version":"1","status":"enabled"},
		"event":{"data":[{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","moderator_user_id":"4242","moderator_user_login":"cool_mod","moderator_user_name":"CoolMod","bullying":3,"overall_level":null,"disability":2,"race_ethnicity_or_religion":1,"misogyny":2,"sexuality_sex_or_gender":3,"aggression":4,"sex_based_terms":1,"swearing":0}]}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(req, "super-secret", "message-automod-settings", "2024-04-11T12:00:01Z", body, "notification")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Code; got != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", got, http.StatusNoContent)
	}
	if len(notifications) != 1 {
		t.Fatalf("len(notifications) = %d, want 1", len(notifications))
	}
	typed, ok := notifications[0].Event.(eventsub.AutomodSettingsUpdateEvent)
	if !ok {
		t.Fatalf("decoded event type = %T, want AutomodSettingsUpdateEvent", notifications[0].Event)
	}
	if got := typed.ModeratorUserLogin; got != "cool_mod" {
		t.Fatalf("ModeratorUserLogin = %q, want %q", got, "cool_mod")
	}
	if typed.OverallLevel != nil {
		t.Fatalf("OverallLevel = %v, want nil", typed.OverallLevel)
	}
	if got := typed.Aggression; got != 4 {
		t.Fatalf("Aggression = %d, want 4", got)
	}
}

func TestWebhookHandlerDecodesAutomodMessageNotifications(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		body   []byte
		assert func(*testing.T, any)
	}{
		{
			name: "hold",
			body: []byte(`{
				"subscription":{"id":"sub-automod-hold","type":"automod.message.hold","version":"1","status":"enabled"},
				"event":{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","user_id":"7734","user_login":"uncool_viewer","user_name":"Uncool_viewer","message_id":"msg-a1","message":{"text":"cheer1 hello","fragments":[{"type":"cheermote","text":"cheer1","cheermote":{"prefix":"cheer","bits":1,"tier":1},"emote":null},{"type":"text","text":" ","cheermote":null,"emote":null},{"type":"emote","text":"hello","cheermote":null,"emote":{"id":"25","emote_set_id":"1"}}]},"category":"aggressive","level":4,"held_at":"2024-04-11T12:00:00Z"}
			}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.AutomodMessageHoldEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want AutomodMessageHoldEvent", event)
				}
				if got := typed.Message.Text; got != "cheer1 hello" {
					t.Fatalf("Message.Text = %q, want %q", got, "cheer1 hello")
				}
				if got := len(typed.Message.Fragments); got != 3 {
					t.Fatalf("len(Message.Fragments) = %d, want 3", got)
				}
				if got := typed.Message.Fragments[0].Cheermote.Bits; got != 1 {
					t.Fatalf("Message.Fragments[0].Cheermote.Bits = %d, want 1", got)
				}
				if got := typed.Message.Fragments[2].Emote.EmoteSetID; got != "1" {
					t.Fatalf("Message.Fragments[2].Emote.EmoteSetID = %q, want %q", got, "1")
				}
			},
		},
		{
			name: "update",
			body: []byte(`{
				"subscription":{"id":"sub-automod-update","type":"automod.message.update","version":"1","status":"enabled"},
				"event":{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","user_id":"7734","user_login":"uncool_viewer","user_name":"Uncool_viewer","moderator_user_id":"4242","moderator_user_name":"CoolMod","moderator_user_login":"cool_mod","message_id":"msg-a2","message":{"text":"Kappa hi","fragments":[{"type":"emote","text":"Kappa","cheermote":null,"emote":{"id":"25","emote_set_id":"1"}},{"type":"text","text":" hi","cheermote":null,"emote":null}]},"category":"bullying","level":3,"status":"approved","held_at":"2024-04-11T12:00:00Z"}
			}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.AutomodMessageUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want AutomodMessageUpdateEvent", event)
				}
				if got := typed.Message.Text; got != "Kappa hi" {
					t.Fatalf("Message.Text = %q, want %q", got, "Kappa hi")
				}
				if got := len(typed.Message.Fragments); got != 2 {
					t.Fatalf("len(Message.Fragments) = %d, want 2", got)
				}
				if got := typed.Message.Fragments[0].Emote.ID; got != "25" {
					t.Fatalf("Message.Fragments[0].Emote.ID = %q, want %q", got, "25")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var notifications []eventsub.Notification
			handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
				Secret: "super-secret",
				Now: func() time.Time {
					return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
				},
				OnNotification: func(_ context.Context, notification eventsub.Notification) error {
					notifications = append(notifications, notification)
					return nil
				},
			})

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tt.body))
			signWebhook(req, "super-secret", "message-"+tt.name, "2024-04-11T12:00:01Z", tt.body, "notification")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if got := rec.Code; got != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", got, http.StatusNoContent)
			}
			if len(notifications) != 1 {
				t.Fatalf("len(notifications) = %d, want 1", len(notifications))
			}
			tt.assert(t, notifications[0].Event)
		})
	}
}

func TestWebhookHandlerReturnsNoContentWhenDedupMarkFailsAfterSuccessfulCallback(t *testing.T) {
	t.Parallel()

	var notifications int
	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret: "super-secret",
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
		Deduplicator: markFailDeduplicator{},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			notifications++
			return nil
		},
	})

	body := []byte(`{
		"subscription":{"id":"sub-1","type":"channel.follow","version":"2","status":"enabled"},
		"event":{"user_id":"777","user_login":"viewer","user_name":"Viewer","broadcaster_user_id":"123","broadcaster_user_login":"caster","broadcaster_user_name":"Caster","followed_at":"2024-04-11T12:00:00Z"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(req, "super-secret", "message-mark-fail", "2024-04-11T12:00:01Z", body, "notification")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Code; got != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", got, http.StatusNoContent)
	}
	if got := notifications; got != 1 {
		t.Fatalf("notifications = %d, want 1", got)
	}
}

func TestWebhookHandlerRejectsBadSignaturesAndSuppressesDuplicates(t *testing.T) {
	t.Parallel()

	var delivered int
	dedup := eventsub.NewMemoryDeduplicator()
	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret:       "super-secret",
		Deduplicator: dedup,
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			delivered++
			return nil
		},
	})

	body := []byte(`{
		"subscription":{"id":"sub-1","type":"channel.follow","version":"2","status":"enabled"},
		"event":{"user_id":"777","user_login":"viewer","user_name":"Viewer","broadcaster_user_id":"123","broadcaster_user_login":"caster","broadcaster_user_name":"Caster","followed_at":"2024-04-11T12:00:00Z"}
	}`)

	badReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	badReq.Header.Set("Twitch-Eventsub-Message-Id", "message-1")
	badReq.Header.Set("Twitch-Eventsub-Message-Timestamp", "2024-04-11T12:00:00Z")
	badReq.Header.Set("Twitch-Eventsub-Message-Type", "notification")
	badReq.Header.Set("Twitch-Eventsub-Message-Signature", "sha256=deadbeef")
	badRec := httptest.NewRecorder()

	handler.ServeHTTP(badRec, badReq)

	if got := badRec.Code; got != http.StatusUnauthorized {
		t.Fatalf("bad signature status = %d, want %d", got, http.StatusUnauthorized)
	}

	for range 2 {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		signWebhook(req, "super-secret", "message-2", "2024-04-11T12:00:01Z", body, "notification")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if got := rec.Code; got != http.StatusNoContent {
			t.Fatalf("duplicate notification status = %d, want %d", got, http.StatusNoContent)
		}
	}

	if got := delivered; got != 1 {
		t.Fatalf("delivered = %d, want 1", got)
	}
}

func TestWebhookHandlerSuppressesDuplicatesByDefault(t *testing.T) {
	t.Parallel()

	var delivered int
	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret: "super-secret",
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			delivered++
			return nil
		},
	})

	body := []byte(`{
		"subscription":{"id":"sub-1","type":"channel.follow","version":"2","status":"enabled"},
		"event":{"user_id":"777","user_login":"viewer","user_name":"Viewer","broadcaster_user_id":"123","broadcaster_user_login":"caster","broadcaster_user_name":"Caster","followed_at":"2024-04-11T12:00:00Z"}
	}`)

	for range 2 {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		signWebhook(req, "super-secret", "message-default-dedup", "2024-04-11T12:00:01Z", body, "notification")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if got := rec.Code; got != http.StatusNoContent {
			t.Fatalf("duplicate notification status = %d, want %d", got, http.StatusNoContent)
		}
	}

	if got := delivered; got != 1 {
		t.Fatalf("delivered = %d, want 1", got)
	}
}

func TestWebhookHandlerSuppressesConcurrentDuplicateDeliveries(t *testing.T) {
	t.Parallel()

	started := make(chan struct{}, 1)
	release := make(chan struct{})
	var delivered int
	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret:       "super-secret",
		Deduplicator: eventsub.NewMemoryDeduplicator(),
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			delivered++
			select {
			case started <- struct{}{}:
			default:
			}
			<-release
			return nil
		},
	})

	body := []byte(`{
		"subscription":{"id":"sub-1","type":"channel.follow","version":"2","status":"enabled"},
		"event":{"user_id":"777","user_login":"viewer","user_name":"Viewer","broadcaster_user_id":"123","broadcaster_user_login":"caster","broadcaster_user_name":"Caster","followed_at":"2024-04-11T12:00:00Z"}
	}`)

	firstReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(firstReq, "super-secret", "message-concurrent", "2024-04-11T12:00:01Z", body, "notification")
	firstRec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(firstRec, firstReq)
		close(done)
	}()

	<-started

	secondReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(secondReq, "super-secret", "message-concurrent", "2024-04-11T12:00:01Z", body, "notification")
	secondRec := httptest.NewRecorder()
	handler.ServeHTTP(secondRec, secondReq)

	if got := secondRec.Code; got != http.StatusServiceUnavailable {
		t.Fatalf("second status = %d, want %d", got, http.StatusServiceUnavailable)
	}

	close(release)
	<-done

	if got := firstRec.Code; got != http.StatusNoContent {
		t.Fatalf("first status = %d, want %d", got, http.StatusNoContent)
	}
	if got := delivered; got != 1 {
		t.Fatalf("delivered = %d, want 1", got)
	}
}

func TestWebhookHandlerRetriesConcurrentDuplicateAfterOriginalFailure(t *testing.T) {
	t.Parallel()

	started := make(chan struct{}, 1)
	release := make(chan struct{})
	var attempts int
	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret:       "super-secret",
		Deduplicator: eventsub.NewMemoryDeduplicator(),
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			attempts++
			if attempts == 1 {
				select {
				case started <- struct{}{}:
				default:
				}
				<-release
				return assertiveError("transient failure")
			}
			return nil
		},
	})

	body := []byte(`{
		"subscription":{"id":"sub-1","type":"channel.follow","version":"2","status":"enabled"},
		"event":{"user_id":"777","user_login":"viewer","user_name":"Viewer","broadcaster_user_id":"123","broadcaster_user_login":"caster","broadcaster_user_name":"Caster","followed_at":"2024-04-11T12:00:00Z"}
	}`)

	firstReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(firstReq, "super-secret", "message-retry-concurrent", "2024-04-11T12:00:01Z", body, "notification")
	firstRec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(firstRec, firstReq)
		close(done)
	}()

	<-started

	secondReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(secondReq, "super-secret", "message-retry-concurrent", "2024-04-11T12:00:01Z", body, "notification")
	secondRec := httptest.NewRecorder()
	handler.ServeHTTP(secondRec, secondReq)

	if got := secondRec.Code; got != http.StatusServiceUnavailable {
		t.Fatalf("second status = %d, want %d", got, http.StatusServiceUnavailable)
	}

	close(release)
	<-done

	if got := firstRec.Code; got != http.StatusInternalServerError {
		t.Fatalf("first status = %d, want %d", got, http.StatusInternalServerError)
	}

	thirdReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(thirdReq, "super-secret", "message-retry-concurrent", "2024-04-11T12:00:01Z", body, "notification")
	thirdRec := httptest.NewRecorder()
	handler.ServeHTTP(thirdRec, thirdReq)

	if got := thirdRec.Code; got != http.StatusNoContent {
		t.Fatalf("third status = %d, want %d", got, http.StatusNoContent)
	}
	if got := attempts; got != 2 {
		t.Fatalf("attempts = %d, want 2", got)
	}
}

func TestWebhookHandlerRetriesSameNotificationAfterCallbackFailure(t *testing.T) {
	t.Parallel()

	var attempts int
	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret:       "super-secret",
		Deduplicator: eventsub.NewMemoryDeduplicator(),
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			attempts++
			if attempts == 1 {
				return assertiveError("transient failure")
			}
			return nil
		},
	})

	body := []byte(`{
		"subscription":{"id":"sub-1","type":"channel.follow","version":"2","status":"enabled"},
		"event":{"user_id":"777","user_login":"viewer","user_name":"Viewer","broadcaster_user_id":"123","broadcaster_user_login":"caster","broadcaster_user_name":"Caster","followed_at":"2024-04-11T12:00:00Z"}
	}`)

	for i := range 2 {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		signWebhook(req, "super-secret", "message-2", "2024-04-11T12:00:01Z", body, "notification")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if i == 0 {
			if got := rec.Code; got != http.StatusInternalServerError {
				t.Fatalf("first status = %d, want %d", got, http.StatusInternalServerError)
			}
			continue
		}
		if got := rec.Code; got != http.StatusNoContent {
			t.Fatalf("second status = %d, want %d", got, http.StatusNoContent)
		}
	}

	if got := attempts; got != 2 {
		t.Fatalf("attempts = %d, want 2", got)
	}
}

func TestWebhookHandlerRetriesUnsupportedMessageTypeAfterBadRequest(t *testing.T) {
	t.Parallel()

	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret:       "super-secret",
		Deduplicator: eventsub.NewMemoryDeduplicator(),
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
	})

	body := []byte(`{"subscription":{"type":"channel.follow","version":"2"}}`)
	for i := range 2 {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		signWebhook(req, "super-secret", "message-unsupported", "2024-04-11T12:00:01Z", body, "unsupported")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if got := rec.Code; got != http.StatusBadRequest {
			t.Fatalf("attempt %d status = %d, want %d", i+1, got, http.StatusBadRequest)
		}
	}
}

func TestWebhookHandlerRejectsStaleTimestamps(t *testing.T) {
	t.Parallel()

	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret: "super-secret",
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 20, 0, 0, time.UTC)
		},
	})

	body := []byte(`{
		"subscription":{"type":"channel.follow","version":"2"},
		"event":{"user_id":"777"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(req, "super-secret", "message-1", "2024-04-11T12:00:00Z", body, "notification")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Code; got != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", got, http.StatusUnauthorized)
	}
}

func TestWebhookHandlerRejectsRequestsMissingRequiredHeaders(t *testing.T) {
	t.Parallel()

	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret: "super-secret",
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
	})

	body := []byte(`{
		"subscription":{"type":"channel.follow","version":"2"},
		"event":{"user_id":"777"}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(req, "super-secret", "message-1", "2024-04-11T12:00:00Z", body, "notification")
	req.Header.Del("Twitch-Eventsub-Message-Type")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Code; got != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", got, http.StatusBadRequest)
	}
}

func TestWebhookHandlerRejectsTimestampsOutsideDefaultTenMinuteWindow(t *testing.T) {
	t.Parallel()

	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret: "super-secret",
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
	})

	body := []byte(`{
		"subscription":{"type":"channel.follow","version":"2"},
		"event":{"user_id":"777"}
	}`)

	tests := []struct {
		name      string
		timestamp string
	}{
		{
			name:      "too old",
			timestamp: "2024-04-11T11:54:59Z",
		},
		{
			name:      "too far in future",
			timestamp: "2024-04-11T12:15:01Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			signWebhook(req, "super-secret", "message-"+tt.name, tt.timestamp, body, "notification")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if got := rec.Code; got != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", got, http.StatusUnauthorized)
			}
		})
	}
}

func TestWebhookHandlerUsesConfiguredTimestampSkew(t *testing.T) {
	t.Parallel()

	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret:           "super-secret",
		MaxTimestampSkew: time.Minute,
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
	})

	body := []byte(`{
		"subscription":{"type":"channel.follow","version":"2"},
		"event":{"user_id":"777"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(req, "super-secret", "message-1", "2024-04-11T12:03:30Z", body, "notification")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Code; got != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", got, http.StatusUnauthorized)
	}
}

func TestWebhookHandlerRejectsRequestsWhenSecretIsEmpty(t *testing.T) {
	t.Parallel()

	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{})
	body := []byte(`{
		"subscription":{"type":"channel.follow","version":"2"},
		"event":{"user_id":"777"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	signWebhook(req, "", "message-1", time.Now().UTC().Format(time.RFC3339), body, "notification")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Code; got != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", got, http.StatusUnauthorized)
	}
}

func TestWebhookHandlerRejectsOversizedBodies(t *testing.T) {
	t.Parallel()

	var called bool
	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret: "super-secret",
		Now: func() time.Time {
			return time.Date(2024, 4, 11, 12, 5, 0, 0, time.UTC)
		},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			called = true
			return nil
		},
	})

	body := bytes.Repeat([]byte("a"), 2<<20)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Code; got != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", got, http.StatusRequestEntityTooLarge)
	}
	if called {
		t.Fatal("notification callback was called for oversized body")
	}
}

func signWebhook(req *http.Request, secret, messageID, timestamp string, body []byte, messageType string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Twitch-Eventsub-Message-Id", messageID)
	req.Header.Set("Twitch-Eventsub-Message-Timestamp", timestamp)
	req.Header.Set("Twitch-Eventsub-Message-Type", messageType)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(messageID))
	mac.Write([]byte(timestamp))
	mac.Write(body)
	req.Header.Set("Twitch-Eventsub-Message-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
}

type assertiveError string

func (e assertiveError) Error() string {
	return string(e)
}

type markFailDeduplicator struct{}

func (markFailDeduplicator) Reserve(context.Context, string) (eventsub.ReservationState, error) {
	return eventsub.ReservationAcquired, nil
}

func (markFailDeduplicator) Complete(context.Context, string) error {
	return assertiveError("mark failed")
}

func (markFailDeduplicator) Forget(context.Context, string) error {
	return nil
}

func TestDefaultRegistryDecodesKnownEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		subscriptionType string
		version          string
		raw              json.RawMessage
		assert           func(*testing.T, any)
	}{
		{
			name:             "channel guest star session begin beta",
			subscriptionType: "channel.guest_star_session.begin",
			version:          "beta",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_name":"Cool_User","broadcaster_user_login":"cool_user","moderator_user_id":"1312","moderator_user_name":"Mod_User","moderator_user_login":"mod_user","session_id":"session-1","started_at":"2023-04-11T10:11:52.123Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelGuestStarSessionBeginEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelGuestStarSessionBeginEvent", event)
				}
				if got := typed.SessionID; got != "session-1" {
					t.Fatalf("SessionID = %q, want %q", got, "session-1")
				}
			},
		},
		{
			name:             "channel guest star session end beta",
			subscriptionType: "channel.guest_star_session.end",
			version:          "beta",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_name":"Cool_User","broadcaster_user_login":"cool_user","moderator_user_id":"1312","moderator_user_name":"Mod_User","moderator_user_login":"mod_user","session_id":"session-1","started_at":"2023-04-11T10:11:52.123Z","ended_at":"2023-04-11T10:41:52.123Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelGuestStarSessionEndEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelGuestStarSessionEndEvent", event)
				}
				if got := typed.ModeratorUserLogin; got != "mod_user" {
					t.Fatalf("ModeratorUserLogin = %q, want %q", got, "mod_user")
				}
			},
		},
		{
			name:             "channel guest star guest update beta",
			subscriptionType: "channel.guest_star_guest.update",
			version:          "beta",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_name":"Cool_User","broadcaster_user_login":"cool_user","session_id":"session-1","moderator_user_id":"1312","moderator_user_name":"Mod_User","moderator_user_login":"mod_user","guest_user_id":"4242","guest_user_name":"Guest_User","guest_user_login":"guest_user","slot_id":"slot-1","state":"accepted","host_video_enabled":true,"host_audio_enabled":false,"host_volume":80}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelGuestStarGuestUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelGuestStarGuestUpdateEvent", event)
				}
				if got := typed.State; got != "accepted" {
					t.Fatalf("State = %q, want %q", got, "accepted")
				}
				if got := typed.HostVolume; got != 80 {
					t.Fatalf("HostVolume = %d, want 80", got)
				}
			},
		},
		{
			name:             "channel guest star settings update beta",
			subscriptionType: "channel.guest_star_settings.update",
			version:          "beta",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_name":"Cool_User","broadcaster_user_login":"cool_user","is_moderator_send_live_enabled":true,"slot_count":5,"is_browser_source_audio_enabled":true,"group_layout":"tiled"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelGuestStarSettingsUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelGuestStarSettingsUpdateEvent", event)
				}
				if got := typed.GroupLayout; got != "tiled" {
					t.Fatalf("GroupLayout = %q, want %q", got, "tiled")
				}
				if got := typed.SlotCount; got != 5 {
					t.Fatalf("SlotCount = %d, want 5", got)
				}
			},
		},
		{
			name:             "extension bits transaction create",
			subscriptionType: "extension.bits_transaction.create",
			version:          "1",
			raw:              json.RawMessage(`{"id":"bits-tx-id","extension_client_id":"deadbeef","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","user_name":"Coolest_User","user_login":"coolest_user","user_id":"1236","product":{"name":"great_product","sku":"skuskusku","bits":1234,"in_development":false}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ExtensionBitsTransactionCreateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ExtensionBitsTransactionCreateEvent", event)
				}
				if got := typed.ExtensionClientID; got != "deadbeef" {
					t.Fatalf("ExtensionClientID = %q, want %q", got, "deadbeef")
				}
				if got := typed.Product.Bits; got != 1234 {
					t.Fatalf("Product.Bits = %d, want 1234", got)
				}
				if typed.Product.InDevelopment {
					t.Fatal("Product.InDevelopment = true, want false")
				}
			},
		},
		{
			name:             "conduit shard disabled",
			subscriptionType: "conduit.shard.disabled",
			version:          "1",
			raw:              json.RawMessage(`{"conduit_id":"conduit-1","shard_id":"0","status":"websocket_disconnected","transport":{"method":"websocket","callback":null,"session_id":"session-1","connected_at":"2024-04-11T12:00:00Z","disconnected_at":"2024-04-11T12:05:00Z"}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ConduitShardDisabledEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ConduitShardDisabledEvent", event)
				}
				if got := typed.ConduitID; got != "conduit-1" {
					t.Fatalf("ConduitID = %q, want %q", got, "conduit-1")
				}
				if got := typed.Transport.Method; got != "websocket" {
					t.Fatalf("Transport.Method = %q, want %q", got, "websocket")
				}
				if typed.Transport.SessionID == nil || *typed.Transport.SessionID != "session-1" {
					t.Fatalf("Transport.SessionID = %v, want session-1", typed.Transport.SessionID)
				}
				if typed.Transport.ConnectedAt == nil || typed.Transport.DisconnectedAt == nil {
					t.Fatalf("Transport timestamps = %#v, want both timestamps", typed.Transport)
				}
			},
		},
		{
			name:             "drop entitlement grant",
			subscriptionType: "drop.entitlement.grant",
			version:          "1",
			raw:              json.RawMessage(`[{"id":"grant-1","data":{"organization_id":"9001","category_id":"9002","category_name":"Fortnite","campaign_id":"9003","user_id":"1234","user_name":"Cool_User","user_login":"cool_user","entitlement_id":"ent-1","benefit_id":"benefit-1","created_at":"2019-01-28T04:17:53.325Z"}}]`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.DropEntitlementGrantBatch)
				if !ok {
					t.Fatalf("decoded event type = %T, want DropEntitlementGrantBatch", event)
				}
				if got := len(typed); got != 1 {
					t.Fatalf("len(batch) = %d, want 1", got)
				}
				if got := typed[0].ID; got != "grant-1" {
					t.Fatalf("batch[0].ID = %q, want %q", got, "grant-1")
				}
				if typed[0].Data.CampaignID == nil || *typed[0].Data.CampaignID != "9003" {
					t.Fatalf("CampaignID = %#v, want %q", typed[0].Data.CampaignID, "9003")
				}
			},
		},
		{
			name:             "channel moderate v1",
			subscriptionType: "channel.moderate",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_login":"glowillig","broadcaster_user_name":"glowillig","moderator_user_id":"424596340","moderator_user_login":"quotrok","moderator_user_name":"quotrok","action":"mod","followers":null,"slow":null,"vip":null,"unvip":null,"mod":{"user_id":"141981764","user_login":"twitchdev","user_name":"TwitchDev"},"unmod":null,"ban":null,"unban":null,"timeout":null,"untimeout":null,"raid":null,"unraid":null,"delete":null,"automod_terms":null,"unban_request":null,"shared_chat_ban":null,"shared_chat_unban":null,"shared_chat_timeout":null,"shared_chat_untimeout":null,"shared_chat_delete":null}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelModerateEventV1)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelModerateEventV1", event)
				}
				if got := typed.Action; got != "mod" {
					t.Fatalf("Action = %q, want %q", got, "mod")
				}
				if typed.Mod == nil || typed.Mod.UserLogin != "twitchdev" {
					t.Fatalf("Mod = %#v, want twitchdev", typed.Mod)
				}
			},
		},
		{
			name:             "channel moderate v2",
			subscriptionType: "channel.moderate",
			version:          "2",
			raw:              json.RawMessage(`{"broadcaster_user_id":"423374343","broadcaster_user_login":"glowillig","broadcaster_user_name":"glowillig","moderator_user_id":"424596340","moderator_user_login":"quotrok","moderator_user_name":"quotrok","action":"warn","followers":null,"slow":null,"vip":null,"unvip":null,"warn":{"user_id":"141981764","user_login":"twitchdev","user_name":"TwitchDev","reason":"cut it out","chat_rules_cited":null},"mod":null,"unmod":null,"ban":null,"unban":null,"timeout":null,"untimeout":null,"raid":null,"unraid":null,"delete":null,"automod_terms":null,"unban_request":null,"shared_chat_ban":null,"shared_chat_unban":null,"shared_chat_timeout":null,"shared_chat_untimeout":null,"shared_chat_delete":null}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelModerateEventV2)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelModerateEventV2", event)
				}
				if got := typed.Action; got != "warn" {
					t.Fatalf("Action = %q, want %q", got, "warn")
				}
				if typed.Warn == nil || typed.Warn.UserLogin != "twitchdev" {
					t.Fatalf("Warn = %#v, want twitchdev", typed.Warn)
				}
				if typed.Warn.Reason == nil || *typed.Warn.Reason != "cut it out" {
					t.Fatalf("Warn.Reason = %v, want cut it out", typed.Warn.Reason)
				}
			},
		},
		{
			name:             "channel chat notification",
			subscriptionType: "channel.chat.notification",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1971641","broadcaster_user_login":"streamer","broadcaster_user_name":"streamer","chatter_user_id":"49912639","chatter_user_login":"viewer23","chatter_user_name":"viewer23","chatter_is_anonymous":false,"color":"","badges":[],"system_message":"viewer23 subscribed at Tier 1.","message_id":"d62235c8-47ff-a4f4-84e8-5a29a65a9c03","message":{"text":"","fragments":[]},"notice_type":"sub","sub":{"sub_tier":"1000","is_prime":false,"duration_months":1},"resub":null,"sub_gift":null,"community_sub_gift":null,"gift_paid_upgrade":null,"prime_paid_upgrade":null,"pay_it_forward":null,"raid":null,"unraid":null,"announcement":null,"bits_badge_tier":null,"charity_donation":null,"watch_streak":null,"shared_chat_sub":null,"shared_chat_resub":null,"shared_chat_sub_gift":null,"shared_chat_community_sub_gift":null,"shared_chat_gift_paid_upgrade":null,"shared_chat_prime_paid_upgrade":null,"shared_chat_pay_it_forward":null,"shared_chat_raid":null,"shared_chat_unraid":null,"shared_chat_announcement":null,"shared_chat_bits_badge_tier":null,"shared_chat_charity_donation":null,"source_broadcaster_user_id":null,"source_broadcaster_user_login":null,"source_broadcaster_user_name":null,"source_message_id":null,"source_badges":null,"is_source_only":null}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelChatNotificationEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelChatNotificationEvent", event)
				}
				if got := typed.NoticeType; got != "sub" {
					t.Fatalf("NoticeType = %q, want %q", got, "sub")
				}
				if typed.Sub == nil {
					t.Fatal("Sub = nil, want subscription notice")
				}
				if got := typed.Sub.SubTier; got != "1000" {
					t.Fatalf("Sub.SubTier = %q, want %q", got, "1000")
				}
				if typed.SourceBroadcasterUserID != nil {
					t.Fatalf("SourceBroadcasterUserID = %#v, want nil", typed.SourceBroadcasterUserID)
				}
			},
		},
		{
			name:             "channel chat notification shared chat",
			subscriptionType: "channel.chat.notification",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1971641","broadcaster_user_login":"streamer","broadcaster_user_name":"streamer","chatter_user_id":"49912639","chatter_user_login":"viewer23","chatter_user_name":"viewer23","chatter_is_anonymous":false,"color":"","badges":[],"system_message":"viewer23 subscribed at Tier 1. They've subscribed for 10 months!","message_id":"d62235c8-47ff-a4f4-84e8-5a29a65a9c03","message":{"text":"","fragments":[]},"notice_type":"shared_chat_resub","sub":null,"resub":null,"sub_gift":null,"community_sub_gift":null,"gift_paid_upgrade":null,"prime_paid_upgrade":null,"pay_it_forward":null,"raid":null,"unraid":null,"announcement":null,"bits_badge_tier":null,"charity_donation":null,"watch_streak":null,"shared_chat_sub":null,"shared_chat_resub":{"cumulative_months":10,"duration_months":0,"streak_months":null,"sub_tier":"1000","is_prime":false,"is_gift":false,"gifter_is_anonymous":null,"gifter_user_id":null,"gifter_user_name":null,"gifter_user_login":null},"shared_chat_sub_gift":null,"shared_chat_community_sub_gift":null,"shared_chat_gift_paid_upgrade":null,"shared_chat_prime_paid_upgrade":null,"shared_chat_pay_it_forward":null,"shared_chat_raid":null,"shared_chat_unraid":null,"shared_chat_announcement":null,"shared_chat_bits_badge_tier":null,"shared_chat_charity_donation":null,"source_broadcaster_user_id":"112233","source_broadcaster_user_login":"streamer33","source_broadcaster_user_name":"streamer33","source_message_id":"2be7193d-0366-4453-b6ec-b288ce9f2c39","source_badges":[{"set_id":"subscriber","id":"3","info":"3"}],"is_source_only":true}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelChatNotificationEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelChatNotificationEvent", event)
				}
				if got := typed.NoticeType; got != "shared_chat_resub" {
					t.Fatalf("NoticeType = %q, want %q", got, "shared_chat_resub")
				}
				if typed.SharedChatResub == nil {
					t.Fatal("SharedChatResub = nil, want shared chat resub notice")
				}
				if got := typed.SharedChatResub.CumulativeMonths; got != 10 {
					t.Fatalf("SharedChatResub.CumulativeMonths = %d, want 10", got)
				}
				if typed.SourceBroadcasterUserLogin == nil || *typed.SourceBroadcasterUserLogin != "streamer33" {
					t.Fatalf("SourceBroadcasterUserLogin = %#v, want %q", typed.SourceBroadcasterUserLogin, "streamer33")
				}
				if got := len(typed.SourceBadges); got != 1 {
					t.Fatalf("len(SourceBadges) = %d, want 1", got)
				}
				if typed.IsSourceOnly == nil || !*typed.IsSourceOnly {
					t.Fatalf("IsSourceOnly = %#v, want true", typed.IsSourceOnly)
				}
			},
		},
		{
			name:             "channel chat notification charity donation",
			subscriptionType: "channel.chat.notification",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1971641","broadcaster_user_login":"streamer","broadcaster_user_name":"streamer","chatter_user_id":"49912639","chatter_user_login":"viewer23","chatter_user_name":"viewer23","chatter_is_anonymous":false,"color":"","badges":[],"system_message":"viewer23 donated $5.00 to Example Charity.","message_id":"d62235c8-47ff-a4f4-84e8-5a29a65a9c03","message":{"text":"","fragments":[]},"notice_type":"charity_donation","sub":null,"resub":null,"sub_gift":null,"community_sub_gift":null,"gift_paid_upgrade":null,"prime_paid_upgrade":null,"pay_it_forward":null,"raid":null,"unraid":null,"announcement":null,"bits_badge_tier":null,"charity_donation":{"charity_name":"Example Charity","amount":{"value":500,"decimal_places":2,"currency":"USD"}},"watch_streak":null,"shared_chat_sub":null,"shared_chat_resub":null,"shared_chat_sub_gift":null,"shared_chat_community_sub_gift":null,"shared_chat_gift_paid_upgrade":null,"shared_chat_prime_paid_upgrade":null,"shared_chat_pay_it_forward":null,"shared_chat_raid":null,"shared_chat_unraid":null,"shared_chat_announcement":null,"shared_chat_bits_badge_tier":null,"shared_chat_charity_donation":null,"source_broadcaster_user_id":null,"source_broadcaster_user_login":null,"source_broadcaster_user_name":null,"source_message_id":null,"source_badges":null,"is_source_only":null}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelChatNotificationEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelChatNotificationEvent", event)
				}
				if got := typed.NoticeType; got != "charity_donation" {
					t.Fatalf("NoticeType = %q, want %q", got, "charity_donation")
				}
				if typed.CharityDonation == nil {
					t.Fatal("CharityDonation = nil, want charity donation notice")
				}
				if got := typed.CharityDonation.CharityName; got != "Example Charity" {
					t.Fatalf("CharityDonation.CharityName = %q, want %q", got, "Example Charity")
				}
				if got := typed.CharityDonation.Amount.Value; got != 500 {
					t.Fatalf("CharityDonation.Amount.Value = %d, want 500", got)
				}
				if got := typed.CharityDonation.Amount.DecimalPlace; got != 2 {
					t.Fatalf("CharityDonation.Amount.DecimalPlace = %d, want 2", got)
				}
				if got := typed.CharityDonation.Amount.Currency; got != "USD" {
					t.Fatalf("CharityDonation.Amount.Currency = %q, want %q", got, "USD")
				}
			},
		},
		{
			name:             "automod message hold",
			subscriptionType: "automod.message.hold",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","user_id":"7734","user_login":"uncool_viewer","user_name":"Uncool_viewer","message_id":"msg-a1","message":{"text":"cheer1 hello","fragments":[{"type":"cheermote","text":"cheer1","cheermote":{"prefix":"cheer","bits":1,"tier":1},"emote":null},{"type":"text","text":" ","cheermote":null,"emote":null},{"type":"emote","text":"hello","cheermote":null,"emote":{"id":"25","emote_set_id":"1"}}]},"category":"aggressive","level":4,"held_at":"2024-04-11T12:00:00Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.AutomodMessageHoldEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want AutomodMessageHoldEvent", event)
				}
				if got := typed.Category; got != "aggressive" {
					t.Fatalf("Category = %q, want %q", got, "aggressive")
				}
				if got := typed.Message.Text; got != "cheer1 hello" {
					t.Fatalf("Message.Text = %q, want %q", got, "cheer1 hello")
				}
				if got := len(typed.Message.Fragments); got != 3 {
					t.Fatalf("len(Message.Fragments) = %d, want 3", got)
				}
				if typed.Message.Fragments[0].Cheermote.Bits != 1 {
					t.Fatalf("Message.Fragments[0].Cheermote.Bits = %d, want 1", typed.Message.Fragments[0].Cheermote.Bits)
				}
				if got := typed.Message.Fragments[2].Emote.EmoteSetID; got != "1" {
					t.Fatalf("Message.Fragments[2].Emote.EmoteSetID = %q, want %q", got, "1")
				}
			},
		},
		{
			name:             "automod message hold v2",
			subscriptionType: "automod.message.hold",
			version:          "2",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_login":"blah","broadcaster_user_name":"blahblah","user_id":"4242","user_login":"baduser","user_name":"badbaduser","message_id":"bad-message-id","message":{"text":"This is a bad message… pogchamp","fragments":[{"type":"text","text":"This is a bad message… ","cheermote":null,"emote":null,"mention":null},{"type":"cheermote","text":"pogchamp","cheermote":{"prefix":"pogchamp","bits":1000,"tier":1},"emote":null,"mention":null}]},"reason":"automod","automod":{"category":"aggressive","level":1,"boundaries":[{"start_pos":0,"end_pos":10},{"start_pos":20,"end_pos":30}]},"blocked_term":null,"held_at":"2022-12-02T15:00:00Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.AutomodMessageHoldV2Event)
				if !ok {
					t.Fatalf("decoded event type = %T, want AutomodMessageHoldV2Event", event)
				}
				if got := typed.Reason; got != "automod" {
					t.Fatalf("Reason = %q, want %q", got, "automod")
				}
				if typed.Automod == nil || typed.Automod.Level != 1 {
					t.Fatalf("Automod = %#v, want level 1", typed.Automod)
				}
				if got := len(typed.Message.Fragments); got != 2 {
					t.Fatalf("len(Message.Fragments) = %d, want 2", got)
				}
				if typed.BlockedTerm != nil {
					t.Fatalf("BlockedTerm = %#v, want nil", typed.BlockedTerm)
				}
			},
		},
		{
			name:             "automod message update",
			subscriptionType: "automod.message.update",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","user_id":"7734","user_login":"uncool_viewer","user_name":"Uncool_viewer","moderator_user_id":"4242","moderator_user_name":"CoolMod","moderator_user_login":"cool_mod","message_id":"msg-a2","message":{"text":"Kappa hi","fragments":[{"type":"emote","text":"Kappa","cheermote":null,"emote":{"id":"25","emote_set_id":"1"}},{"type":"text","text":" hi","cheermote":null,"emote":null}]},"category":"bullying","level":3,"status":"approved","held_at":"2024-04-11T12:00:00Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.AutomodMessageUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want AutomodMessageUpdateEvent", event)
				}
				if got := typed.Status; got != "approved" {
					t.Fatalf("Status = %q, want %q", got, "approved")
				}
				if got := typed.ModeratorUserLogin; got != "cool_mod" {
					t.Fatalf("ModeratorUserLogin = %q, want %q", got, "cool_mod")
				}
				if got := typed.Message.Text; got != "Kappa hi" {
					t.Fatalf("Message.Text = %q, want %q", got, "Kappa hi")
				}
				if got := len(typed.Message.Fragments); got != 2 {
					t.Fatalf("len(Message.Fragments) = %d, want 2", got)
				}
				if got := typed.Message.Fragments[0].Emote.ID; got != "25" {
					t.Fatalf("Message.Fragments[0].Emote.ID = %q, want %q", got, "25")
				}
			},
		},
		{
			name:             "automod message update v2",
			subscriptionType: "automod.message.update",
			version:          "2",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_login":"blah","broadcaster_user_name":"blahblah","moderator_user_id":"9001","moderator_user_login":"the_mod","moderator_user_name":"The_Mod","user_id":"4242","user_login":"baduser","user_name":"badbaduser","message_id":"bad-message-id","message":{"text":"This is a bad message… pogchamp","fragments":[{"type":"text","text":"This is a bad message… ","cheermote":null,"emote":null,"mention":null},{"type":"cheermote","text":"pogchamp","cheermote":{"prefix":"pogchamp","bits":1000,"tier":1},"emote":null,"mention":null}]},"reason":"blocked_term","automod":null,"blocked_term":{"terms_found":[{"term_id":"123","owner_broadcaster_user_id":"1337","owner_broadcaster_user_login":"blah","owner_broadcaster_user_name":"blahblah","boundary":{"start_pos":0,"end_pos":30}}]},"status":"approved","held_at":"2022-12-02T15:00:00Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.AutomodMessageUpdateV2Event)
				if !ok {
					t.Fatalf("decoded event type = %T, want AutomodMessageUpdateV2Event", event)
				}
				if got := typed.Status; got != "approved" {
					t.Fatalf("Status = %q, want %q", got, "approved")
				}
				if got := typed.Reason; got != "blocked_term" {
					t.Fatalf("Reason = %q, want %q", got, "blocked_term")
				}
				if typed.BlockedTerm == nil || len(typed.BlockedTerm.TermsFound) != 1 {
					t.Fatalf("BlockedTerm = %#v, want one term", typed.BlockedTerm)
				}
				if typed.Automod != nil {
					t.Fatalf("Automod = %#v, want nil", typed.Automod)
				}
			},
		},
		{
			name:             "automod settings update",
			subscriptionType: "automod.settings.update",
			version:          "1",
			raw:              json.RawMessage(`{"data":[{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","moderator_user_id":"4242","moderator_user_login":"cool_mod","moderator_user_name":"CoolMod","bullying":3,"overall_level":null,"disability":2,"race_ethnicity_or_religion":1,"misogyny":2,"sexuality_sex_or_gender":3,"aggression":4,"sex_based_terms":1,"swearing":0}]}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.AutomodSettingsUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want AutomodSettingsUpdateEvent", event)
				}
				if typed.OverallLevel != nil {
					t.Fatalf("OverallLevel = %v, want nil", typed.OverallLevel)
				}
				if got := typed.Aggression; got != 4 {
					t.Fatalf("Aggression = %d, want 4", got)
				}
			},
		},
		{
			name:             "automod terms update",
			subscriptionType: "automod.terms.update",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","moderator_user_id":"4242","moderator_user_login":"cool_mod","moderator_user_name":"CoolMod","action":"add_blocked","from_automod":false,"terms":["spoiler","blocked phrase"]}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.AutomodTermsUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want AutomodTermsUpdateEvent", event)
				}
				if got := typed.Action; got != "add_blocked" {
					t.Fatalf("Action = %q, want %q", got, "add_blocked")
				}
				if got := len(typed.Terms); got != 2 {
					t.Fatalf("len(Terms) = %d, want 2", got)
				}
			},
		},
		{
			name:             "channel follow",
			subscriptionType: "channel.follow",
			version:          "2",
			raw:              json.RawMessage(`{"user_id":"777","user_login":"viewer","user_name":"Viewer","broadcaster_user_id":"123","broadcaster_user_login":"caster","broadcaster_user_name":"Caster","followed_at":"2024-04-11T12:00:00Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelFollowEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelFollowEvent", event)
				}
				if got := typed.UserID; got != "777" {
					t.Fatalf("UserID = %q, want %q", got, "777")
				}
			},
		},
		{
			name:             "channel update",
			subscriptionType: "channel.update",
			version:          "2",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","title":"Best Stream Ever","language":"en","category_id":"12453","category_name":"Grand Theft Auto","content_classification_labels":["MatureGame"]}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelUpdateEvent", event)
				}
				if got := typed.Title; got != "Best Stream Ever" {
					t.Fatalf("Title = %q, want %q", got, "Best Stream Ever")
				}
				if len(typed.ContentClassificationLabels) != 1 || typed.ContentClassificationLabels[0] != "MatureGame" {
					t.Fatalf("ContentClassificationLabels = %#v, want [MatureGame]", typed.ContentClassificationLabels)
				}
			},
		},
		{
			name:             "channel ad break begin",
			subscriptionType: "channel.ad_break.begin",
			version:          "1",
			raw:              json.RawMessage(`{"duration_seconds":"60","started_at":"2019-11-16T10:11:12.634234626Z","is_automatic":"false","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","requester_user_id":"1337","requester_user_login":"cool_user","requester_user_name":"Cool_User"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelAdBreakBeginEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelAdBreakBeginEvent", event)
				}
				if got := typed.DurationSeconds; got != 60 {
					t.Fatalf("DurationSeconds = %d, want 60", got)
				}
				if typed.IsAutomatic {
					t.Fatal("IsAutomatic = true, want false")
				}
			},
		},
		{
			name:             "channel ad break begin native types",
			subscriptionType: "channel.ad_break.begin",
			version:          "1",
			raw:              json.RawMessage(`{"duration_seconds":90,"started_at":"2019-11-16T10:11:12.634234626Z","is_automatic":true,"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","requester_user_id":"1337","requester_user_login":"cool_user","requester_user_name":"Cool_User"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelAdBreakBeginEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelAdBreakBeginEvent", event)
				}
				if got := typed.DurationSeconds; got != 90 {
					t.Fatalf("DurationSeconds = %d, want 90", got)
				}
				if !typed.IsAutomatic {
					t.Fatal("IsAutomatic = false, want true")
				}
			},
		},
		{
			name:             "channel bits use",
			subscriptionType: "channel.bits.use",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1234","user_login":"cool_user","user_name":"Cool_User","broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User","bits":2,"type":"cheer","power_up":null,"custom_power_up":null,"message":{"text":"cheer1 hi Kappa","fragments":[{"type":"cheermote","text":"cheer1","cheermote":{"prefix":"cheer","bits":1,"tier":1},"emote":null},{"type":"text","text":" hi ","cheermote":null,"emote":null},{"type":"emote","text":"Kappa","cheermote":null,"emote":{"id":"25","emote_set_id":"1","owner_id":"42","format":["static","animated"]}}]}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelBitsUseEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelBitsUseEvent", event)
				}
				if got := typed.Bits; got != 2 {
					t.Fatalf("Bits = %d, want 2", got)
				}
				if string(typed.PowerUp) != "null" || string(typed.CustomPowerUp) != "null" {
					t.Fatalf("PowerUp/CustomPowerUp = %s/%s, want null/null", typed.PowerUp, typed.CustomPowerUp)
				}
				if got := len(typed.Message.Fragments); got != 3 {
					t.Fatalf("len(Message.Fragments) = %d, want 3", got)
				}
				if typed.Message.Fragments[0].Cheermote == nil || typed.Message.Fragments[0].Cheermote.Bits != 1 {
					t.Fatalf("first fragment cheermote = %#v, want 1 bit", typed.Message.Fragments[0].Cheermote)
				}
				emote := typed.Message.Fragments[2].Emote
				if emote == nil {
					t.Fatal("third fragment emote = nil, want emote metadata")
				}
				if got := emote.OwnerID; got != "42" {
					t.Fatalf("Emote.OwnerID = %q, want %q", got, "42")
				}
				if got := emote.Format; len(got) != 2 || got[0] != "static" || got[1] != "animated" {
					t.Fatalf("Emote.Format = %#v, want []string{\"static\", \"animated\"}", got)
				}
			},
		},
		{
			name:             "channel points automatic reward redemption add v1",
			subscriptionType: "channel.channel_points_automatic_reward_redemption.add",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"12826","broadcaster_user_name":"Twitch","broadcaster_user_login":"twitch","user_id":"141981764","user_name":"TwitchDev","user_login":"twitchdev","id":"f024099a-e0fe-4339-9a0a-a706fb59f353","reward":{"type":"send_highlighted_message","cost":100,"unlocked_emote":null},"message":{"text":"Hello world! VoHiYo","emotes":[{"id":"81274","begin":13,"end":18}]},"user_input":"Hello world! VoHiYo ","redeemed_at":"2024-02-23T21:14:34.260398045Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPointsAutomaticRewardRedemptionAddEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPointsAutomaticRewardRedemptionAddEvent", event)
				}
				if got := typed.Reward.Cost; got != 100 {
					t.Fatalf("Reward.Cost = %d, want 100", got)
				}
				if got := len(typed.Message.Emotes); got != 1 {
					t.Fatalf("len(Message.Emotes) = %d, want 1", got)
				}
				if got := typed.Message.Emotes[0].ID; got != "81274" {
					t.Fatalf("Message.Emotes[0].ID = %q, want %q", got, "81274")
				}
			},
		},
		{
			name:             "channel points automatic reward redemption add v2",
			subscriptionType: "channel.channel_points_automatic_reward_redemption.add",
			version:          "2",
			raw:              json.RawMessage(`{"broadcaster_user_id":"12826","broadcaster_user_name":"Twitch","broadcaster_user_login":"twitch","user_id":"141981764","user_name":"TwitchDev","user_login":"twitchdev","id":"f024099a-e0fe-4339-9a0a-a706fb59f353","reward":{"type":"send_highlighted_message","channel_points":100,"emote":null},"message":{"text":"Hello world! VoHiYo","fragments":[{"type":"text","text":"Hello world! ","emote":null},{"type":"emote","text":"VoHiYo","emote":{"id":"81274"}}]},"redeemed_at":"2024-08-12T21:14:34.260398045Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPointsAutomaticRewardRedemptionAddV2Event)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPointsAutomaticRewardRedemptionAddV2Event", event)
				}
				if got := typed.Reward.ChannelPoints; got != 100 {
					t.Fatalf("Reward.ChannelPoints = %d, want 100", got)
				}
				if typed.Message == nil {
					t.Fatal("Message = nil, want fragments")
				}
				if got := len(typed.Message.Fragments); got != 2 {
					t.Fatalf("len(Message.Fragments) = %d, want 2", got)
				}
				if typed.Message.Fragments[1].Emote == nil || typed.Message.Fragments[1].Emote.ID != "81274" {
					t.Fatalf("Message.Fragments[1].Emote = %#v, want id 81274", typed.Message.Fragments[1].Emote)
				}
			},
		},
		{
			name:             "channel points custom reward add",
			subscriptionType: "channel.channel_points_custom_reward.add",
			version:          "1",
			raw:              json.RawMessage(`{"id":"reward-1","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","is_enabled":true,"is_paused":false,"is_in_stock":true,"title":"Cool Reward","cost":100,"prompt":"reward prompt","is_user_input_required":true,"should_redemptions_skip_request_queue":false,"cooldown_expires_at":null,"redemptions_redeemed_current_stream":null,"max_per_stream":{"is_enabled":true,"value":1000},"max_per_user_per_stream":{"is_enabled":true,"value":10},"global_cooldown":{"is_enabled":true,"seconds":1000},"background_color":"#FA1ED2","image":{"url_1x":"https://static-cdn.jtvnw.net/image-1.png","url_2x":"https://static-cdn.jtvnw.net/image-2.png","url_4x":"https://static-cdn.jtvnw.net/image-4.png"},"default_image":{"url_1x":"https://static-cdn.jtvnw.net/default-1.png","url_2x":"https://static-cdn.jtvnw.net/default-2.png","url_4x":"https://static-cdn.jtvnw.net/default-4.png"}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPointsCustomRewardAddEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPointsCustomRewardAddEvent", event)
				}
				if got := typed.Title; got != "Cool Reward" {
					t.Fatalf("Title = %q, want %q", got, "Cool Reward")
				}
				if typed.CooldownExpiresAt != nil {
					t.Fatalf("CooldownExpiresAt = %v, want nil", typed.CooldownExpiresAt)
				}
				if typed.Image == nil || typed.Image.URL2x != "https://static-cdn.jtvnw.net/image-2.png" {
					t.Fatalf("Image = %#v, want custom image URLs", typed.Image)
				}
			},
		},
		{
			name:             "channel points custom reward update",
			subscriptionType: "channel.channel_points_custom_reward.update",
			version:          "1",
			raw:              json.RawMessage(`{"id":"reward-1","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","is_enabled":true,"is_paused":false,"is_in_stock":true,"title":"Cool Reward","cost":250,"prompt":"reward prompt","is_user_input_required":true,"should_redemptions_skip_request_queue":true,"cooldown_expires_at":"2024-04-11T12:30:00Z","redemptions_redeemed_current_stream":123,"max_per_stream":{"is_enabled":true,"value":1000},"max_per_user_per_stream":{"is_enabled":true,"value":10},"global_cooldown":{"is_enabled":true,"seconds":1000},"background_color":"#FA1ED2","image":null,"default_image":{"url_1x":"https://static-cdn.jtvnw.net/default-1.png","url_2x":"https://static-cdn.jtvnw.net/default-2.png","url_4x":"https://static-cdn.jtvnw.net/default-4.png"}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPointsCustomRewardUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPointsCustomRewardUpdateEvent", event)
				}
				if got := typed.Cost; got != 250 {
					t.Fatalf("Cost = %d, want 250", got)
				}
				if typed.CooldownExpiresAt == nil {
					t.Fatal("CooldownExpiresAt = nil, want timestamp")
				}
				if typed.RedemptionsRedeemedCurrentStream == nil || *typed.RedemptionsRedeemedCurrentStream != 123 {
					t.Fatalf("RedemptionsRedeemedCurrentStream = %v, want 123", typed.RedemptionsRedeemedCurrentStream)
				}
				if typed.Image != nil {
					t.Fatalf("Image = %#v, want nil", typed.Image)
				}
			},
		},
		{
			name:             "channel points custom reward remove",
			subscriptionType: "channel.channel_points_custom_reward.remove",
			version:          "1",
			raw:              json.RawMessage(`{"id":"reward-1","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","is_enabled":false,"is_paused":false,"is_in_stock":false,"title":"Removed Reward","cost":100,"prompt":"reward prompt","is_user_input_required":false,"should_redemptions_skip_request_queue":false,"cooldown_expires_at":null,"redemptions_redeemed_current_stream":null,"max_per_stream":{"is_enabled":false,"value":0},"max_per_user_per_stream":{"is_enabled":false,"value":0},"global_cooldown":{"is_enabled":false,"seconds":0},"background_color":"#FA1ED2","image":null,"default_image":{"url_1x":"https://static-cdn.jtvnw.net/default-1.png","url_2x":"https://static-cdn.jtvnw.net/default-2.png","url_4x":"https://static-cdn.jtvnw.net/default-4.png"}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPointsCustomRewardRemoveEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPointsCustomRewardRemoveEvent", event)
				}
				if got := typed.Title; got != "Removed Reward" {
					t.Fatalf("Title = %q, want %q", got, "Removed Reward")
				}
				if typed.IsEnabled {
					t.Fatal("IsEnabled = true, want false")
				}
			},
		},
		{
			name:             "channel points custom reward redemption add",
			subscriptionType: "channel.channel_points_custom_reward_redemption.add",
			version:          "1",
			raw:              json.RawMessage(`{"id":"redemption-1","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","user_id":"9001","user_login":"cooler_user","user_name":"Cooler_User","user_input":"pogchamp","status":"unfulfilled","reward":{"id":"reward-1","title":"Cool Reward","cost":100,"prompt":"reward prompt"},"redeemed_at":"2020-07-15T17:16:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPointsCustomRewardRedemptionAddEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPointsCustomRewardRedemptionAddEvent", event)
				}
				if got := typed.Status; got != "unfulfilled" {
					t.Fatalf("Status = %q, want %q", got, "unfulfilled")
				}
				if got := typed.Reward.Cost; got != 100 {
					t.Fatalf("Reward.Cost = %d, want 100", got)
				}
			},
		},
		{
			name:             "channel points custom reward redemption update",
			subscriptionType: "channel.channel_points_custom_reward_redemption.update",
			version:          "1",
			raw:              json.RawMessage(`{"id":"redemption-1","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","user_id":"9001","user_login":"cooler_user","user_name":"Cooler_User","user_input":"pogchamp","status":"fulfilled","reward":{"id":"reward-1","title":"Cool Reward","cost":100,"prompt":"reward prompt"},"redeemed_at":"2020-07-15T17:16:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPointsCustomRewardRedemptionUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPointsCustomRewardRedemptionUpdateEvent", event)
				}
				if got := typed.Status; got != "fulfilled" {
					t.Fatalf("Status = %q, want %q", got, "fulfilled")
				}
				if got := typed.UserInput; got != "pogchamp" {
					t.Fatalf("UserInput = %q, want %q", got, "pogchamp")
				}
			},
		},
		{
			name:             "stream online",
			subscriptionType: "stream.online",
			version:          "1",
			raw:              json.RawMessage(`{"id":"9001","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","type":"live","started_at":"2020-10-11T10:11:12.123Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.StreamOnlineEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want StreamOnlineEvent", event)
				}
				if got := typed.ID; got != "9001" {
					t.Fatalf("ID = %q, want %q", got, "9001")
				}
			},
		},
		{
			name:             "stream offline",
			subscriptionType: "stream.offline",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.StreamOfflineEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want StreamOfflineEvent", event)
				}
				if got := typed.BroadcasterUserLogin; got != "cool_user" {
					t.Fatalf("BroadcasterUserLogin = %q, want %q", got, "cool_user")
				}
			},
		},
		{
			name:             "channel chat clear",
			subscriptionType: "channel.chat.clear",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_name":"Cool_User","broadcaster_user_login":"cool_user"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelChatClearEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelChatClearEvent", event)
				}
				if got := typed.BroadcasterUserLogin; got != "cool_user" {
					t.Fatalf("BroadcasterUserLogin = %q, want %q", got, "cool_user")
				}
			},
		},
		{
			name:             "channel chat clear user messages",
			subscriptionType: "channel.chat.clear_user_messages",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_name":"Cool_User","broadcaster_user_login":"cool_user","target_user_id":"7734","target_user_name":"Uncool_viewer","target_user_login":"uncool_viewer"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelChatClearUserMessagesEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelChatClearUserMessagesEvent", event)
				}
				if got := typed.TargetUserLogin; got != "uncool_viewer" {
					t.Fatalf("TargetUserLogin = %q, want %q", got, "uncool_viewer")
				}
			},
		},
		{
			name:             "channel chat message",
			subscriptionType: "channel.chat.message",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_name":"Cool_User","broadcaster_user_login":"cool_user","chatter_user_id":"7734","chatter_user_name":"Uncool_viewer","chatter_user_login":"uncool_viewer","message_id":"msg-0","message":{"text":"@caster Kappa","fragments":[{"type":"mention","text":"@caster","cheermote":null,"emote":null,"mention":{"user_id":"1337","user_name":"Cool_User","user_login":"cool_user"}},{"type":"text","text":" ","cheermote":null,"emote":null,"mention":null},{"type":"emote","text":"Kappa","cheermote":null,"emote":{"id":"25","emote_set_id":"1","owner_id":"1337","format":["static"]},"mention":null}]},"message_type":"text","badges":[{"set_id":"subscriber","id":"12","info":"12"}],"cheer":null,"color":"#FF0000","reply":{"parent_message_id":"parent-1","parent_message_body":"hello","parent_user_id":"42","parent_user_name":"Other","parent_user_login":"other","thread_message_id":"thread-1","thread_user_id":"42","thread_user_name":"Other","thread_user_login":"other"},"channel_points_custom_reward_id":null,"source_broadcaster_user_id":null,"source_broadcaster_user_name":null,"source_broadcaster_user_login":null,"source_message_id":null,"source_badges":null,"is_source_only":null}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelChatMessageEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelChatMessageEvent", event)
				}
				if got := typed.MessageType; got != "text" {
					t.Fatalf("MessageType = %q, want %q", got, "text")
				}
				if got := len(typed.Message.Fragments); got != 3 {
					t.Fatalf("len(Message.Fragments) = %d, want 3", got)
				}
				if typed.Message.Fragments[0].Mention == nil || typed.Message.Fragments[0].Mention.UserID != "1337" {
					t.Fatalf("first fragment mention = %#v, want user 1337", typed.Message.Fragments[0].Mention)
				}
				if typed.Message.Fragments[2].Emote == nil || typed.Message.Fragments[2].Emote.OwnerID != "1337" {
					t.Fatalf("third fragment emote = %#v, want owner 1337", typed.Message.Fragments[2].Emote)
				}
				if typed.Reply == nil || typed.Reply.ParentMessageID != "parent-1" {
					t.Fatalf("Reply = %#v, want parent-1", typed.Reply)
				}
				if got := len(typed.Badges); got != 1 {
					t.Fatalf("len(Badges) = %d, want 1", got)
				}
			},
		},
		{
			name:             "channel chat message delete",
			subscriptionType: "channel.chat.message_delete",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_name":"Cool_User","broadcaster_user_login":"cool_user","target_user_id":"7734","target_user_name":"Uncool_viewer","target_user_login":"uncool_viewer","message_id":"msg-1"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelChatMessageDeleteEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelChatMessageDeleteEvent", event)
				}
				if got := typed.MessageID; got != "msg-1" {
					t.Fatalf("MessageID = %q, want %q", got, "msg-1")
				}
			},
		},
		{
			name:             "channel chat settings update",
			subscriptionType: "channel.chat_settings.update",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","emote_mode":true,"follower_mode":false,"follower_mode_duration_minutes":null,"slow_mode":true,"slow_mode_wait_time_seconds":30,"subscriber_mode":true,"unique_chat_mode":false}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelChatSettingsUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelChatSettingsUpdateEvent", event)
				}
				if typed.FollowerModeDurationMinutes != nil {
					t.Fatalf("FollowerModeDurationMinutes = %v, want nil", typed.FollowerModeDurationMinutes)
				}
				if typed.SlowModeWaitTimeSeconds == nil || *typed.SlowModeWaitTimeSeconds != 30 {
					t.Fatalf("SlowModeWaitTimeSeconds = %v, want 30", typed.SlowModeWaitTimeSeconds)
				}
			},
		},
		{
			name:             "channel chat user message hold",
			subscriptionType: "channel.chat.user_message_hold",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","user_id":"7734","user_login":"uncool_viewer","user_name":"Uncool_viewer","message_id":"msg-2","message":{"text":"cheer1 hello","fragments":[{"type":"cheermote","text":"cheer1","cheermote":{"prefix":"cheer","bits":1,"tier":1},"emote":null},{"type":"text","text":" hello","cheermote":null,"emote":null}]}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelChatUserMessageHoldEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelChatUserMessageHoldEvent", event)
				}
				if got := typed.MessageID; got != "msg-2" {
					t.Fatalf("MessageID = %q, want %q", got, "msg-2")
				}
				if got := len(typed.Message.Fragments); got != 2 {
					t.Fatalf("len(Message.Fragments) = %d, want 2", got)
				}
				if typed.Message.Fragments[0].Cheermote == nil || typed.Message.Fragments[0].Cheermote.Bits != 1 {
					t.Fatalf("first fragment cheermote = %#v, want 1 bit", typed.Message.Fragments[0].Cheermote)
				}
			},
		},
		{
			name:             "channel chat user message update",
			subscriptionType: "channel.chat.user_message_update",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","user_id":"7734","user_login":"uncool_viewer","user_name":"Uncool_viewer","status":"approved","message_id":"msg-3","message":{"text":"Kappa hi","fragments":[{"type":"emote","text":"Kappa","cheermote":null,"emote":{"id":"25","emote_set_id":"1"}},{"type":"text","text":" hi","cheermote":null,"emote":null}]}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelChatUserMessageUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelChatUserMessageUpdateEvent", event)
				}
				if got := typed.Status; got != "approved" {
					t.Fatalf("Status = %q, want %q", got, "approved")
				}
				if got := len(typed.Message.Fragments); got != 2 {
					t.Fatalf("len(Message.Fragments) = %d, want 2", got)
				}
				if typed.Message.Fragments[0].Emote == nil || typed.Message.Fragments[0].Emote.ID != "25" {
					t.Fatalf("first fragment emote = %#v, want id 25", typed.Message.Fragments[0].Emote)
				}
			},
		},
		{
			name:             "channel subscribe",
			subscriptionType: "channel.subscribe",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1234","user_login":"cool_user","user_name":"Cool_User","broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User","tier":"1000","is_gift":false}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelSubscribeEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelSubscribeEvent", event)
				}
				if got := typed.Tier; got != "1000" {
					t.Fatalf("Tier = %q, want %q", got, "1000")
				}
			},
		},
		{
			name:             "channel subscription gift",
			subscriptionType: "channel.subscription.gift",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1234","user_login":"cool_user","user_name":"Cool_User","broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User","total":2,"tier":"1000","cumulative_total":284,"is_anonymous":false}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelSubscriptionGiftEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelSubscriptionGiftEvent", event)
				}
				if typed.UserID == nil || *typed.UserID != "1234" {
					t.Fatalf("UserID = %v, want 1234", typed.UserID)
				}
				if typed.CumulativeTotal == nil || *typed.CumulativeTotal != 284 {
					t.Fatalf("CumulativeTotal = %v, want 284", typed.CumulativeTotal)
				}
			},
		},
		{
			name:             "channel subscription end",
			subscriptionType: "channel.subscription.end",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1234","user_login":"cool_user","user_name":"Cool_User","broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User","tier":"1000","is_gift":false}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelSubscriptionEndEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelSubscriptionEndEvent", event)
				}
				if got := typed.UserName; got != "Cool_User" {
					t.Fatalf("UserName = %q, want %q", got, "Cool_User")
				}
			},
		},
		{
			name:             "channel subscription message",
			subscriptionType: "channel.subscription.message",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1234","user_login":"cool_user","user_name":"Cool_User","broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User","tier":"1000","message":{"text":"Love the stream! FevziGG","emotes":[{"begin":23,"end":30,"id":"302976485"}]},"cumulative_months":15,"streak_months":1,"duration_months":6}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelSubscriptionMessageEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelSubscriptionMessageEvent", event)
				}
				if got := typed.Message.Text; got != "Love the stream! FevziGG" {
					t.Fatalf("Message.Text = %q, want %q", got, "Love the stream! FevziGG")
				}
				if len(typed.Message.Emotes) != 1 || typed.Message.Emotes[0].ID != "302976485" {
					t.Fatalf("Message.Emotes = %#v, want one emote with id 302976485", typed.Message.Emotes)
				}
				if typed.StreakMonths == nil || *typed.StreakMonths != 1 {
					t.Fatalf("StreakMonths = %v, want 1", typed.StreakMonths)
				}
			},
		},
		{
			name:             "channel goal begin",
			subscriptionType: "channel.goal.begin",
			version:          "1",
			raw:              json.RawMessage(`{"id":"12345-cool-event","broadcaster_user_id":"141981764","broadcaster_user_name":"TwitchDev","broadcaster_user_login":"twitchdev","type":"subscription","description":"Help me get partner!","current_amount":100,"target_amount":220,"started_at":"2021-07-15T17:16:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelGoalBeginEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelGoalBeginEvent", event)
				}
				if got := typed.TargetAmount; got != 220 {
					t.Fatalf("TargetAmount = %d, want 220", got)
				}
			},
		},
		{
			name:             "channel goal progress",
			subscriptionType: "channel.goal.progress",
			version:          "1",
			raw:              json.RawMessage(`{"id":"12345-cool-event","broadcaster_user_id":"141981764","broadcaster_user_name":"TwitchDev","broadcaster_user_login":"twitchdev","type":"subscription","description":"Help me get partner!","current_amount":120,"target_amount":220,"started_at":"2021-07-15T17:16:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelGoalProgressEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelGoalProgressEvent", event)
				}
				if got := typed.CurrentAmount; got != 120 {
					t.Fatalf("CurrentAmount = %d, want 120", got)
				}
			},
		},
		{
			name:             "channel goal end",
			subscriptionType: "channel.goal.end",
			version:          "1",
			raw:              json.RawMessage(`{"id":"12345-abc-678-defgh","broadcaster_user_id":"141981764","broadcaster_user_name":"TwitchDev","broadcaster_user_login":"twitchdev","type":"subscription","description":"Help me get partner!","is_achieved":false,"current_amount":180,"target_amount":220,"started_at":"2021-07-15T17:16:03.17106713Z","ended_at":"2020-07-16T17:16:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelGoalEndEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelGoalEndEvent", event)
				}
				if typed.IsAchieved {
					t.Fatal("IsAchieved = true, want false")
				}
				if got := typed.Description; got != "Help me get partner!" {
					t.Fatalf("Description = %q, want %q", got, "Help me get partner!")
				}
			},
		},
		{
			name:             "channel poll begin",
			subscriptionType: "channel.poll.begin",
			version:          "1",
			raw:              json.RawMessage(`{"id":"1243456","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","title":"Aren’t shoes just really hard socks?","choices":[{"id":"123","title":"Yeah!"},{"id":"124","title":"No!"},{"id":"125","title":"Maybe!"}],"bits_voting":{"is_enabled":true,"amount_per_vote":10},"channel_points_voting":{"is_enabled":true,"amount_per_vote":10},"started_at":"2020-07-15T17:16:03.17106713Z","ends_at":"2020-07-15T17:16:08.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPollBeginEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPollBeginEvent", event)
				}
				if got := len(typed.Choices); got != 3 {
					t.Fatalf("len(Choices) = %d, want 3", got)
				}
				if !typed.BitsVoting.IsEnabled {
					t.Fatal("BitsVoting.IsEnabled = false, want true")
				}
			},
		},
		{
			name:             "channel poll progress",
			subscriptionType: "channel.poll.progress",
			version:          "1",
			raw:              json.RawMessage(`{"id":"1243456","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","title":"Aren’t shoes just really hard socks?","choices":[{"id":"123","title":"Yeah!","bits_votes":5,"channel_points_votes":7,"votes":12},{"id":"124","title":"No!","bits_votes":10,"channel_points_votes":4,"votes":14},{"id":"125","title":"Maybe!","bits_votes":0,"channel_points_votes":7,"votes":7}],"bits_voting":{"is_enabled":true,"amount_per_vote":10},"channel_points_voting":{"is_enabled":true,"amount_per_vote":10},"started_at":"2020-07-15T17:16:03.17106713Z","ends_at":"2020-07-15T17:16:08.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPollProgressEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPollProgressEvent", event)
				}
				if got := typed.Choices[1].Votes; got != 14 {
					t.Fatalf("Choices[1].Votes = %d, want 14", got)
				}
			},
		},
		{
			name:             "channel poll end",
			subscriptionType: "channel.poll.end",
			version:          "1",
			raw:              json.RawMessage(`{"id":"1243456","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","title":"Aren’t shoes just really hard socks?","choices":[{"id":"123","title":"Blue","bits_votes":50,"channel_points_votes":70,"votes":120},{"id":"124","title":"Yellow","bits_votes":100,"channel_points_votes":40,"votes":140},{"id":"125","title":"Green","bits_votes":10,"channel_points_votes":70,"votes":80}],"bits_voting":{"is_enabled":true,"amount_per_vote":10},"channel_points_voting":{"is_enabled":true,"amount_per_vote":10},"status":"completed","started_at":"2020-07-15T17:16:03.17106713Z","ended_at":"2020-07-15T17:16:11.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPollEndEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPollEndEvent", event)
				}
				if got := typed.Status; got != "completed" {
					t.Fatalf("Status = %q, want %q", got, "completed")
				}
				if got := typed.Choices[0].BitsVotes; got != 50 {
					t.Fatalf("Choices[0].BitsVotes = %d, want 50", got)
				}
			},
		},
		{
			name:             "channel prediction begin",
			subscriptionType: "channel.prediction.begin",
			version:          "1",
			raw:              json.RawMessage(`{"id":"1243456","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","title":"Aren’t shoes just really hard socks?","outcomes":[{"id":"1243456","title":"Yeah!","color":"blue"},{"id":"2243456","title":"No!","color":"pink"}],"started_at":"2020-07-15T17:16:03.17106713Z","locks_at":"2020-07-15T17:21:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPredictionBeginEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPredictionBeginEvent", event)
				}
				if got := len(typed.Outcomes); got != 2 {
					t.Fatalf("len(Outcomes) = %d, want 2", got)
				}
				if got := typed.Outcomes[0].Color; got != "blue" {
					t.Fatalf("Outcomes[0].Color = %q, want %q", got, "blue")
				}
			},
		},
		{
			name:             "channel prediction progress",
			subscriptionType: "channel.prediction.progress",
			version:          "1",
			raw:              json.RawMessage(`{"id":"1243456","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","title":"Aren’t shoes just really hard socks?","outcomes":[{"id":"1243456","title":"Yeah!","color":"blue","users":10,"channel_points":15000,"top_predictors":[{"user_name":"Cool_User","user_login":"cool_user","user_id":"1234","channel_points_won":null,"channel_points_used":500}]},{"id":"2243456","title":"No!","color":"pink","users":3,"channel_points":5000,"top_predictors":[{"user_name":"Cooler_User","user_login":"cooler_user","user_id":"12345","channel_points_won":null,"channel_points_used":5000}]}],"started_at":"2020-07-15T17:16:03.17106713Z","locks_at":"2020-07-15T17:21:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPredictionProgressEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPredictionProgressEvent", event)
				}
				if got := typed.Outcomes[0].TopPredictors[0].ChannelPointsUsed; got != 500 {
					t.Fatalf("TopPredictors[0].ChannelPointsUsed = %d, want 500", got)
				}
			},
		},
		{
			name:             "channel prediction lock",
			subscriptionType: "channel.prediction.lock",
			version:          "1",
			raw:              json.RawMessage(`{"id":"1243456","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","title":"Aren’t shoes just really hard socks?","outcomes":[{"id":"1243456","title":"Yeah!","color":"blue","users":10,"channel_points":15000,"top_predictors":[{"user_name":"Cool_User","user_login":"cool_user","user_id":"1234","channel_points_won":null,"channel_points_used":500}]},{"id":"2243456","title":"No!","color":"pink","users":3,"channel_points":5000,"top_predictors":[{"user_name":"Cooler_User","user_login":"cooler_user","user_id":"12345","channel_points_won":null,"channel_points_used":5000}]}],"started_at":"2020-07-15T17:16:03.17106713Z","locked_at":"2020-07-15T17:21:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPredictionLockEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPredictionLockEvent", event)
				}
				if got := typed.Outcomes[1].Users; got != 3 {
					t.Fatalf("Outcomes[1].Users = %d, want 3", got)
				}
			},
		},
		{
			name:             "channel prediction end",
			subscriptionType: "channel.prediction.end",
			version:          "1",
			raw:              json.RawMessage(`{"id":"1243456","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","title":"Aren’t shoes just really hard socks?","winning_outcome_id":"12345","outcomes":[{"id":"12345","title":"Yeah!","color":"blue","users":2,"channel_points":15000,"top_predictors":[{"user_name":"Cool_User","user_login":"cool_user","user_id":"1234","channel_points_won":10000,"channel_points_used":500}]},{"id":"22435","title":"No!","color":"pink","users":2,"channel_points":200,"top_predictors":[{"user_name":"Cooler_User","user_login":"cooler_user","user_id":"12345","channel_points_won":null,"channel_points_used":100}]}],"status":"resolved","started_at":"2020-07-15T17:16:03.17106713Z","ended_at":"2020-07-15T17:16:11.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelPredictionEndEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelPredictionEndEvent", event)
				}
				if typed.WinningOutcomeID == nil || *typed.WinningOutcomeID != "12345" {
					t.Fatalf("WinningOutcomeID = %v, want 12345", typed.WinningOutcomeID)
				}
				if got := typed.Outcomes[0].TopPredictors[0].ChannelPointsWon; got == nil || *got != 10000 {
					t.Fatalf("ChannelPointsWon = %v, want 10000", got)
				}
			},
		},
		{
			name:             "channel charity campaign donate",
			subscriptionType: "channel.charity_campaign.donate",
			version:          "1",
			raw:              json.RawMessage(`{"id":"a1b2c3-aabb-4455-d1e2f3","campaign_id":"123-abc-456-def","broadcaster_user_id":"123456","broadcaster_user_name":"SunnySideUp","broadcaster_user_login":"sunnysideup","user_id":"654321","user_login":"generoususer1","user_name":"GenerousUser1","charity_name":"Example name","charity_description":"Example description","charity_logo":"https://abc.cloudfront.net/ppgf/1000/100.png","charity_website":"https://www.example.com","amount":{"value":10000,"decimal_places":2,"currency":"USD"}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelCharityCampaignDonateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelCharityCampaignDonateEvent", event)
				}
				if got := typed.Amount.Value; got != 10000 {
					t.Fatalf("Amount.Value = %d, want 10000", got)
				}
				if got := typed.UserLogin; got != "generoususer1" {
					t.Fatalf("UserLogin = %q, want %q", got, "generoususer1")
				}
			},
		},
		{
			name:             "channel charity campaign start",
			subscriptionType: "channel.charity_campaign.start",
			version:          "1",
			raw:              json.RawMessage(`{"id":"123-abc-456-def","broadcaster_id":"123456","broadcaster_name":"SunnySideUp","broadcaster_login":"sunnysideup","charity_name":"Example name","charity_description":"Example description","charity_logo":"https://abc.cloudfront.net/ppgf/1000/100.png","charity_website":"https://www.example.com","current_amount":{"value":0,"decimal_places":2,"currency":"USD"},"target_amount":{"value":1500000,"decimal_places":2,"currency":"USD"},"started_at":"2022-07-26T17:00:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelCharityCampaignStartEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelCharityCampaignStartEvent", event)
				}
				if got := typed.TargetAmount.Currency; got != "USD" {
					t.Fatalf("TargetAmount.Currency = %q, want %q", got, "USD")
				}
				if got := typed.BroadcasterID; got != "123456" {
					t.Fatalf("BroadcasterID = %q, want %q", got, "123456")
				}
			},
		},
		{
			name:             "channel charity campaign progress",
			subscriptionType: "channel.charity_campaign.progress",
			version:          "1",
			raw:              json.RawMessage(`{"id":"123-abc-456-def","broadcaster_id":"123456","broadcaster_name":"SunnySideUp","broadcaster_login":"sunnysideup","charity_name":"Example name","charity_description":"Example description","charity_logo":"https://abc.cloudfront.net/ppgf/1000/100.png","charity_website":"https://www.example.com","current_amount":{"value":260000,"decimal_places":2,"currency":"USD"},"target_amount":{"value":1500000,"decimal_places":2,"currency":"USD"}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelCharityCampaignProgressEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelCharityCampaignProgressEvent", event)
				}
				if got := typed.CurrentAmount.Value; got != 260000 {
					t.Fatalf("CurrentAmount.Value = %d, want 260000", got)
				}
				if got := typed.BroadcasterName; got != "SunnySideUp" {
					t.Fatalf("BroadcasterName = %q, want %q", got, "SunnySideUp")
				}
			},
		},
		{
			name:             "channel charity campaign stop",
			subscriptionType: "channel.charity_campaign.stop",
			version:          "1",
			raw:              json.RawMessage(`{"id":"123-abc-456-def","broadcaster_id":"123456","broadcaster_name":"SunnySideUp","broadcaster_login":"sunnysideup","charity_name":"Example name","charity_description":"Example description","charity_logo":"https://abc.cloudfront.net/ppgf/1000/100.png","charity_website":"https://www.example.com","current_amount":{"value":1450000,"decimal_places":2,"currency":"USD"},"target_amount":{"value":1500000,"decimal_places":2,"currency":"USD"},"stopped_at":"2022-07-26T22:00:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelCharityCampaignStopEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelCharityCampaignStopEvent", event)
				}
				if got := typed.BroadcasterLogin; got != "sunnysideup" {
					t.Fatalf("BroadcasterLogin = %q, want %q", got, "sunnysideup")
				}
			},
		},
		{
			name:             "channel hype train begin",
			subscriptionType: "channel.hype_train.begin",
			version:          "2",
			raw:              json.RawMessage(`{"id":"1b0AsbInCHZW2SQFQkCzqN07Ib2","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","total":137,"progress":137,"goal":500,"top_contributions":[{"user_id":"123","user_login":"pogchamp","user_name":"PogChamp","type":"bits","total":50}],"shared_train_participants":null,"level":1,"started_at":"2020-07-15T17:16:03.17106713Z","expires_at":"2020-07-15T17:16:11.17106713Z","is_shared_train":false,"type":"regular","all_time_high_level":4,"all_time_high_total":2845}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelHypeTrainBeginEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelHypeTrainBeginEvent", event)
				}
				if got := typed.AllTimeHighLevel; got != 4 {
					t.Fatalf("AllTimeHighLevel = %d, want 4", got)
				}
				if got := typed.TopContributions[0].UserLogin; got != "pogchamp" {
					t.Fatalf("TopContributions[0].UserLogin = %q, want %q", got, "pogchamp")
				}
				if got := typed.Progress; got != 137 {
					t.Fatalf("Progress = %d, want 137", got)
				}
			},
		},
		{
			name:             "channel hype train progress",
			subscriptionType: "channel.hype_train.progress",
			version:          "2",
			raw:              json.RawMessage(`{"id":"1b0AsbInCHZW2SQFQkCzqN07Ib2","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","total":137,"progress":137,"goal":500,"top_contributions":[{"user_id":"123","user_login":"pogchamp","user_name":"PogChamp","type":"bits","total":50}],"shared_train_participants":null,"level":1,"started_at":"2020-07-15T17:16:03.17106713Z","expires_at":"2020-07-15T17:16:11.17106713Z","is_shared_train":false,"type":"regular"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelHypeTrainProgressEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelHypeTrainProgressEvent", event)
				}
				if got := typed.Goal; got != 500 {
					t.Fatalf("Goal = %d, want 500", got)
				}
			},
		},
		{
			name:             "channel hype train end",
			subscriptionType: "channel.hype_train.end",
			version:          "2",
			raw:              json.RawMessage(`{"id":"1b0AsbInCHZW2SQFQkCzqN07Ib2","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","total":137,"top_contributions":[{"user_id":"123","user_login":"pogchamp","user_name":"PogChamp","type":"bits","total":50}],"shared_train_participants":null,"level":1,"started_at":"2020-07-15T17:16:03.17106713Z","ended_at":"2020-07-15T17:16:11.17106713Z","cooldown_ends_at":"2020-07-16T17:16:11.17106713Z","is_shared_train":false,"type":"regular"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelHypeTrainEndEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelHypeTrainEndEvent", event)
				}
				if got := typed.CooldownEndsAt.Year(); got != 2020 {
					t.Fatalf("CooldownEndsAt.Year = %d, want 2020", got)
				}
				if _, ok := reflect.TypeOf(typed).FieldByName("Progress"); ok {
					t.Fatal("ChannelHypeTrainEndEvent unexpectedly exposes Progress")
				}
				if _, ok := reflect.TypeOf(typed).FieldByName("Goal"); ok {
					t.Fatal("ChannelHypeTrainEndEvent unexpectedly exposes Goal")
				}
			},
		},
		{
			name:             "channel shared chat begin",
			subscriptionType: "channel.shared_chat.begin",
			version:          "1",
			raw:              json.RawMessage(`{"session_id":"2b64a92a-dbb8-424e-b1c3-304423ba1b6f","broadcaster_user_id":"1971641","broadcaster_user_login":"streamer","broadcaster_user_name":"streamer","host_broadcaster_user_id":"1971641","host_broadcaster_user_login":"streamer","host_broadcaster_user_name":"streamer","participants":[{"broadcaster_user_id":"1971641","broadcaster_user_name":"streamer","broadcaster_user_login":"streamer"},{"broadcaster_user_id":"112233","broadcaster_user_name":"streamer33","broadcaster_user_login":"streamer33"}]}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelSharedChatBeginEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelSharedChatBeginEvent", event)
				}
				if got := len(typed.Participants); got != 2 {
					t.Fatalf("len(Participants) = %d, want 2", got)
				}
			},
		},
		{
			name:             "channel shared chat update",
			subscriptionType: "channel.shared_chat.update",
			version:          "1",
			raw:              json.RawMessage(`{"session_id":"2b64a92a-dbb8-424e-b1c3-304423ba1b6f","broadcaster_user_id":"1971641","broadcaster_user_login":"streamer","broadcaster_user_name":"streamer","host_broadcaster_user_id":"1971641","host_broadcaster_user_login":"streamer","host_broadcaster_user_name":"streamer","participants":[{"broadcaster_user_id":"1971641","broadcaster_user_name":"streamer","broadcaster_user_login":"streamer"},{"broadcaster_user_id":"112233","broadcaster_user_name":"streamer33","broadcaster_user_login":"streamer33"},{"broadcaster_user_id":"332211","broadcaster_user_name":"streamer11","broadcaster_user_login":"streamer11"}]}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelSharedChatUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelSharedChatUpdateEvent", event)
				}
				if got := typed.Participants[2].BroadcasterUserLogin; got != "streamer11" {
					t.Fatalf("Participants[2].BroadcasterUserLogin = %q, want %q", got, "streamer11")
				}
			},
		},
		{
			name:             "channel shared chat end",
			subscriptionType: "channel.shared_chat.end",
			version:          "1",
			raw:              json.RawMessage(`{"session_id":"2b64a92a-dbb8-424e-b1c3-304423ba1b6f","broadcaster_user_id":"1971641","broadcaster_user_login":"streamer","broadcaster_user_name":"streamer","host_broadcaster_user_id":"1971641","host_broadcaster_user_login":"streamer","host_broadcaster_user_name":"streamer"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelSharedChatEndEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelSharedChatEndEvent", event)
				}
				if got := typed.SessionID; got != "2b64a92a-dbb8-424e-b1c3-304423ba1b6f" {
					t.Fatalf("SessionID = %q, want session id", got)
				}
			},
		},
		{
			name:             "channel shield mode begin",
			subscriptionType: "channel.shield_mode.begin",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"12345","broadcaster_user_name":"SimplySimple","broadcaster_user_login":"simplysimple","moderator_user_id":"98765","moderator_user_name":"ParticularlyParticular123","moderator_user_login":"particularlyparticular123","started_at":"2022-07-26T17:00:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelShieldModeBeginEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelShieldModeBeginEvent", event)
				}
				if got := typed.ModeratorUserID; got != "98765" {
					t.Fatalf("ModeratorUserID = %q, want %q", got, "98765")
				}
			},
		},
		{
			name:             "channel shield mode end",
			subscriptionType: "channel.shield_mode.end",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"12345","broadcaster_user_name":"SimplySimple","broadcaster_user_login":"simplysimple","moderator_user_id":"98765","moderator_user_name":"ParticularlyParticular123","moderator_user_login":"particularlyparticular123","ended_at":"2022-07-27T01:30:23.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelShieldModeEndEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelShieldModeEndEvent", event)
				}
				if got := typed.BroadcasterUserLogin; got != "simplysimple" {
					t.Fatalf("BroadcasterUserLogin = %q, want %q", got, "simplysimple")
				}
			},
		},
		{
			name:             "channel shoutout create",
			subscriptionType: "channel.shoutout.create",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"12345","broadcaster_user_name":"SimplySimple","broadcaster_user_login":"simplysimple","moderator_user_id":"98765","moderator_user_name":"ParticularlyParticular123","moderator_user_login":"particularlyparticular123","to_broadcaster_user_id":"626262","to_broadcaster_user_name":"SandySanderman","to_broadcaster_user_login":"sandysanderman","started_at":"2022-07-26T17:00:03.17106713Z","viewer_count":860,"cooldown_ends_at":"2022-07-26T17:02:03.17106713Z","target_cooldown_ends_at":"2022-07-26T18:00:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelShoutoutCreateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelShoutoutCreateEvent", event)
				}
				if got := typed.ToBroadcasterUserLogin; got != "sandysanderman" {
					t.Fatalf("ToBroadcasterUserLogin = %q, want %q", got, "sandysanderman")
				}
				if got := typed.ViewerCount; got != 860 {
					t.Fatalf("ViewerCount = %d, want 860", got)
				}
			},
		},
		{
			name:             "channel shoutout receive",
			subscriptionType: "channel.shoutout.receive",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"626262","broadcaster_user_name":"SandySanderman","broadcaster_user_login":"sandysanderman","from_broadcaster_user_id":"12345","from_broadcaster_user_name":"SimplySimple","from_broadcaster_user_login":"simplysimple","viewer_count":860,"started_at":"2022-07-26T17:00:03.17106713Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelShoutoutReceiveEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelShoutoutReceiveEvent", event)
				}
				if got := typed.FromBroadcasterUserName; got != "SimplySimple" {
					t.Fatalf("FromBroadcasterUserName = %q, want %q", got, "SimplySimple")
				}
			},
		},
		{
			name:             "channel warning acknowledge",
			subscriptionType: "channel.warning.acknowledge",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"423374343","broadcaster_user_login":"glowillig","broadcaster_user_name":"glowillig","user_id":"141981764","user_login":"twitchdev","user_name":"TwitchDev"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelWarningAcknowledgeEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelWarningAcknowledgeEvent", event)
				}
				if got := typed.UserLogin; got != "twitchdev" {
					t.Fatalf("UserLogin = %q, want %q", got, "twitchdev")
				}
			},
		},
		{
			name:             "channel warning send",
			subscriptionType: "channel.warning.send",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"423374343","broadcaster_user_login":"glowillig","broadcaster_user_name":"glowillig","moderator_user_id":"424596340","moderator_user_login":"quotrok","moderator_user_name":"quotrok","user_id":"141981764","user_login":"twitchdev","user_name":"TwitchDev","reason":"cut it out","chat_rules_cited":null}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelWarningSendEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelWarningSendEvent", event)
				}
				if typed.Reason == nil || *typed.Reason != "cut it out" {
					t.Fatalf("Reason = %#v, want %q", typed.Reason, "cut it out")
				}
				if typed.ChatRulesCited != nil {
					t.Fatalf("ChatRulesCited = %#v, want nil", typed.ChatRulesCited)
				}
			},
		},
		{
			name:             "channel warning send null reason",
			subscriptionType: "channel.warning.send",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"423374343","broadcaster_user_login":"glowillig","broadcaster_user_name":"glowillig","moderator_user_id":"424596340","moderator_user_login":"quotrok","moderator_user_name":"quotrok","user_id":"141981764","user_login":"twitchdev","user_name":"TwitchDev","reason":null,"chat_rules_cited":["Be kind"]}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelWarningSendEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelWarningSendEvent", event)
				}
				if typed.Reason != nil {
					t.Fatalf("Reason = %#v, want nil", typed.Reason)
				}
				if got := len(typed.ChatRulesCited); got != 1 {
					t.Fatalf("len(ChatRulesCited) = %d, want 1", got)
				}
			},
		},
		{
			name:             "channel unban request create",
			subscriptionType: "channel.unban_request.create",
			version:          "1",
			raw:              json.RawMessage(`{"id":"60","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","user_id":"1339","user_login":"not_cool_user","user_name":"Not_Cool_User","text":"unban me","created_at":"2023-11-16T10:11:12.634234626Z"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelUnbanRequestCreateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelUnbanRequestCreateEvent", event)
				}
				if got := typed.Text; got != "unban me" {
					t.Fatalf("Text = %q, want %q", got, "unban me")
				}
			},
		},
		{
			name:             "channel unban request resolve",
			subscriptionType: "channel.unban_request.resolve",
			version:          "1",
			raw:              json.RawMessage(`{"id":"60","broadcaster_user_id":"1337","broadcaster_user_login":"cool_user","broadcaster_user_name":"Cool_User","moderator_user_id":"1337","moderator_user_login":"cool_user","moderator_user_name":"Cool_User","user_id":"1339","user_login":"not_cool_user","user_name":"Not_Cool_User","resolution_text":"no","status":"denied"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelUnbanRequestResolveEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelUnbanRequestResolveEvent", event)
				}
				if got := typed.Status; got != "denied" {
					t.Fatalf("Status = %q, want %q", got, "denied")
				}
				if got := typed.ResolutionText; got != "no" {
					t.Fatalf("ResolutionText = %q, want %q", got, "no")
				}
				if got := typed.ModeratorUserID; got != "1337" {
					t.Fatalf("ModeratorUserID = %q, want %q", got, "1337")
				}
				if got := typed.ModeratorUserLogin; got != "cool_user" {
					t.Fatalf("ModeratorUserLogin = %q, want %q", got, "cool_user")
				}
				if got := typed.ModeratorUserName; got != "Cool_User" {
					t.Fatalf("ModeratorUserName = %q, want %q", got, "Cool_User")
				}
			},
		},
		{
			name:             "user authorization grant",
			subscriptionType: "user.authorization.grant",
			version:          "1",
			raw:              json.RawMessage(`{"client_id":"crq72vsaoijkc83xx42hz6i37","user_id":"1337","user_login":"cool_user","user_name":"Cool_User"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.UserAuthorizationGrantEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want UserAuthorizationGrantEvent", event)
				}
				if got := typed.ClientID; got != "crq72vsaoijkc83xx42hz6i37" {
					t.Fatalf("ClientID = %q, want %q", got, "crq72vsaoijkc83xx42hz6i37")
				}
			},
		},
		{
			name:             "user authorization revoke",
			subscriptionType: "user.authorization.revoke",
			version:          "1",
			raw:              json.RawMessage(`{"client_id":"crq72vsaoijkc83xx42hz6i37","user_id":"1337","user_login":null,"user_name":null}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.UserAuthorizationRevokeEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want UserAuthorizationRevokeEvent", event)
				}
				if typed.UserLogin != nil || typed.UserName != nil {
					t.Fatalf("UserLogin/UserName = %#v/%#v, want nil", typed.UserLogin, typed.UserName)
				}
			},
		},
		{
			name:             "user update",
			subscriptionType: "user.update",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1337","user_login":"cool_user","user_name":"Cool_User","email":"user@email.com","email_verified":true,"description":"cool description"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.UserUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want UserUpdateEvent", event)
				}
				if typed.Email == nil || *typed.Email != "user@email.com" {
					t.Fatalf("Email = %v, want user@email.com", typed.Email)
				}
				if !typed.EmailVerified {
					t.Fatal("EmailVerified = false, want true")
				}
			},
		},
		{
			name:             "user whisper message",
			subscriptionType: "user.whisper.message",
			version:          "1",
			raw:              json.RawMessage(`{"from_user_id":"423374343","from_user_login":"glowillig","from_user_name":"glowillig","to_user_id":"424596340","to_user_login":"quotrok","to_user_name":"quotrok","whisper_id":"some-whisper-id","whisper":{"text":"a secret"}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.UserWhisperMessageEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want UserWhisperMessageEvent", event)
				}
				if got := typed.Whisper.Text; got != "a secret" {
					t.Fatalf("Whisper.Text = %q, want %q", got, "a secret")
				}
				if got := typed.FromUserLogin; got != "glowillig" {
					t.Fatalf("FromUserLogin = %q, want %q", got, "glowillig")
				}
			},
		},
		{
			name:             "channel suspicious user update",
			subscriptionType: "channel.suspicious_user.update",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1050263435","broadcaster_user_name":"77f111cbb75341449f5","broadcaster_user_login":"77f111cbb75341449f5","moderator_user_id":"1050263436","moderator_user_name":"29087e59dfc441968f6","moderator_user_login":"29087e59dfc441968f6","user_id":"1050263437","user_name":"06fbcc75952245c5a87","user_login":"06fbcc75952245c5a87","low_trust_status":"restricted"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelSuspiciousUserUpdateEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelSuspiciousUserUpdateEvent", event)
				}
				if got := typed.LowTrustStatus; got != "restricted" {
					t.Fatalf("LowTrustStatus = %q, want %q", got, "restricted")
				}
			},
		},
		{
			name:             "channel suspicious user message",
			subscriptionType: "channel.suspicious_user.message",
			version:          "1",
			raw:              json.RawMessage(`{"broadcaster_user_id":"1050263432","broadcaster_user_name":"dcf9dd9336034d23b65","broadcaster_user_login":"dcf9dd9336034d23b65","user_id":"1050263434","user_name":"4a46e2cf2e2f4d6a9e6","user_login":"4a46e2cf2e2f4d6a9e6","low_trust_status":"active_monitoring","shared_ban_channel_ids":["100","200"],"types":["ban_evader"],"ban_evasion_evaluation":"likely","message":{"message_id":"101010","text":"bad stuff pogchamp","fragments":[{"type":"emote","text":"bad stuff","cheermote":null,"emote":{"id":"899","emote_set_id":"1"}},{"type":"cheermote","text":"pogchamp","cheermote":{"prefix":"pogchamp","bits":100,"tier":1},"emote":null}]}}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelSuspiciousUserMessageEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelSuspiciousUserMessageEvent", event)
				}
				if got := typed.BanEvasionEvaluation; got != "likely" {
					t.Fatalf("BanEvasionEvaluation = %q, want %q", got, "likely")
				}
				if got := len(typed.Message.Fragments); got != 2 {
					t.Fatalf("len(Message.Fragments) = %d, want 2", got)
				}
				if typed.Message.Fragments[0].Emote == nil || typed.Message.Fragments[0].Emote.ID != "899" {
					t.Fatalf("first fragment emote = %#v, want id 899", typed.Message.Fragments[0].Emote)
				}
				if typed.Message.Fragments[1].Cheermote == nil || typed.Message.Fragments[1].Cheermote.Bits != 100 {
					t.Fatalf("second fragment cheermote = %#v, want 100 bits", typed.Message.Fragments[1].Cheermote)
				}
			},
		},
		{
			name:             "channel cheer anonymous",
			subscriptionType: "channel.cheer",
			version:          "1",
			raw:              json.RawMessage(`{"is_anonymous":true,"user_id":null,"user_login":null,"user_name":null,"broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User","message":"pogchamp","bits":1000}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelCheerEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelCheerEvent", event)
				}
				if !typed.IsAnonymous {
					t.Fatal("IsAnonymous = false, want true")
				}
				if typed.UserID != nil || typed.UserLogin != nil || typed.UserName != nil {
					t.Fatalf("anonymous cheer user fields = %#v/%#v/%#v, want nil", typed.UserID, typed.UserLogin, typed.UserName)
				}
			},
		},
		{
			name:             "channel ban",
			subscriptionType: "channel.ban",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1234","user_login":"cool_user","user_name":"Cool_User","broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User","moderator_user_id":"1339","moderator_user_login":"mod_user","moderator_user_name":"Mod_User","reason":"Offensive language","banned_at":"2020-07-15T18:15:11.17106713Z","ends_at":"2020-07-15T18:16:11.17106713Z","is_permanent":false}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelBanEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelBanEvent", event)
				}
				if got := typed.Reason; got != "Offensive language" {
					t.Fatalf("Reason = %q, want %q", got, "Offensive language")
				}
				if typed.EndsAt == nil {
					t.Fatal("EndsAt = nil, want timeout end time")
				}
			},
		},
		{
			name:             "channel unban",
			subscriptionType: "channel.unban",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1234","user_login":"cool_user","user_name":"Cool_User","broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User","moderator_user_id":"1339","moderator_user_login":"mod_user","moderator_user_name":"Mod_User"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelUnbanEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelUnbanEvent", event)
				}
				if got := typed.ModeratorUserName; got != "Mod_User" {
					t.Fatalf("ModeratorUserName = %q, want %q", got, "Mod_User")
				}
			},
		},
		{
			name:             "channel moderator add",
			subscriptionType: "channel.moderator.add",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1234","user_login":"mod_user","user_name":"Mod_User","broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelModeratorAddEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelModeratorAddEvent", event)
				}
				if got := typed.UserLogin; got != "mod_user" {
					t.Fatalf("UserLogin = %q, want %q", got, "mod_user")
				}
			},
		},
		{
			name:             "channel moderator remove",
			subscriptionType: "channel.moderator.remove",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1234","user_login":"not_mod_user","user_name":"Not_Mod_User","broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelModeratorRemoveEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelModeratorRemoveEvent", event)
				}
				if got := typed.UserName; got != "Not_Mod_User" {
					t.Fatalf("UserName = %q, want %q", got, "Not_Mod_User")
				}
			},
		},
		{
			name:             "channel vip add",
			subscriptionType: "channel.vip.add",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1234","user_login":"mod_user","user_name":"Mod_User","broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelVIPAddEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelVIPAddEvent", event)
				}
				if got := typed.BroadcasterUserName; got != "Cooler_User" {
					t.Fatalf("BroadcasterUserName = %q, want %q", got, "Cooler_User")
				}
			},
		},
		{
			name:             "channel vip remove",
			subscriptionType: "channel.vip.remove",
			version:          "1",
			raw:              json.RawMessage(`{"user_id":"1234","user_login":"mod_user","user_name":"Mod_User","broadcaster_user_id":"1337","broadcaster_user_login":"cooler_user","broadcaster_user_name":"Cooler_User"}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelVIPRemoveEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelVIPRemoveEvent", event)
				}
				if got := typed.UserID; got != "1234" {
					t.Fatalf("UserID = %q, want %q", got, "1234")
				}
			},
		},
		{
			name:             "channel raid",
			subscriptionType: "channel.raid",
			version:          "1",
			raw:              json.RawMessage(`{"from_broadcaster_user_id":"1234","from_broadcaster_user_login":"cool_user","from_broadcaster_user_name":"Cool_User","to_broadcaster_user_id":"1337","to_broadcaster_user_login":"cooler_user","to_broadcaster_user_name":"Cooler_User","viewers":9001}`),
			assert: func(t *testing.T, event any) {
				t.Helper()
				typed, ok := event.(eventsub.ChannelRaidEvent)
				if !ok {
					t.Fatalf("decoded event type = %T, want ChannelRaidEvent", event)
				}
				if got := typed.Viewers; got != 9001 {
					t.Fatalf("Viewers = %d, want 9001", got)
				}
			},
		},
	}

	for _, tt := range tests {
		event, err := eventsub.DefaultRegistry().Decode(tt.subscriptionType, tt.version, tt.raw)
		if err != nil {
			t.Fatalf("%s: Decode() error = %v", tt.name, err)
		}
		tt.assert(t, event)
	}
}
