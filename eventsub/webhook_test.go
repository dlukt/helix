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
