package eventsub

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

const maxWebhookBodyBytes = 1 << 20

const defaultTimestampSkew = 10 * time.Minute

// ReservationState describes the outcome of reserving a message ID.
type ReservationState uint8

const (
	ReservationAcquired ReservationState = iota
	ReservationDuplicateInFlight
	ReservationDuplicateCompleted
)

// Deduplicator suppresses duplicate message delivery.
type Deduplicator interface {
	Reserve(context.Context, string) (ReservationState, error)
	Complete(context.Context, string) error
	Forget(context.Context, string) error
}

// MemoryDeduplicator tracks message IDs in memory.
type MemoryDeduplicator struct {
	mu   sync.Mutex
	now  func() time.Time
	ttl  time.Duration
	seen map[string]dedupEntry
}

type dedupEntry struct {
	completed bool
	at        time.Time
}

// NewMemoryDeduplicator creates an in-memory deduplicator.
func NewMemoryDeduplicator() *MemoryDeduplicator {
	return NewMemoryDeduplicatorWithTTL(10 * time.Minute)
}

// NewMemoryDeduplicatorWithTTL creates an in-memory deduplicator with a custom expiry.
func NewMemoryDeduplicatorWithTTL(ttl time.Duration) *MemoryDeduplicator {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &MemoryDeduplicator{
		now:  time.Now,
		ttl:  ttl,
		seen: map[string]dedupEntry{},
	}
}

// Reserve atomically records an in-flight message ID or reports a duplicate.
func (d *MemoryDeduplicator) Reserve(_ context.Context, messageID string) (ReservationState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cleanupExpiredLocked(d.now())
	if entry, ok := d.seen[messageID]; ok {
		if entry.completed {
			return ReservationDuplicateCompleted, nil
		}
		return ReservationDuplicateInFlight, nil
	}
	d.seen[messageID] = dedupEntry{at: d.now()}
	return ReservationAcquired, nil
}

// Complete records the message ID as successfully processed.
func (d *MemoryDeduplicator) Complete(_ context.Context, messageID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.now()
	d.cleanupExpiredLocked(now)
	d.seen[messageID] = dedupEntry{completed: true, at: now}
	return nil
}

// Forget removes an in-flight reservation so a later retry can process it again.
func (d *MemoryDeduplicator) Forget(_ context.Context, messageID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.seen, messageID)
	return nil
}

func (d *MemoryDeduplicator) cleanupExpiredLocked(now time.Time) {
	for messageID, entry := range d.seen {
		if now.Sub(entry.at) > d.ttl {
			delete(d.seen, messageID)
		}
	}
}

// WebhookHandlerConfig configures a WebhookHandler.
type WebhookHandlerConfig struct {
	Secret           string
	Registry         Registry
	Deduplicator     Deduplicator
	DeduplicationTTL time.Duration
	MaxTimestampSkew time.Duration
	OnChallenge      func(context.Context, Challenge) error
	OnNotification   func(context.Context, Notification) error
	OnRevocation     func(context.Context, Revocation) error
	Now              func() time.Time
}

// WebhookHandler verifies and dispatches EventSub webhook requests.
type WebhookHandler struct {
	secret           string
	registry         Registry
	deduplicator     Deduplicator
	onChallenge      func(context.Context, Challenge) error
	onNotification   func(context.Context, Notification) error
	onRevocation     func(context.Context, Revocation) error
	now              func() time.Time
	maxTimestampSkew time.Duration
}

// NewWebhookHandler constructs a webhook handler.
func NewWebhookHandler(cfg WebhookHandlerConfig) *WebhookHandler {
	registry := cfg.Registry
	if registry == nil {
		registry = DefaultRegistry()
	}
	deduplicator := cfg.Deduplicator
	if deduplicator == nil {
		deduplicator = NewMemoryDeduplicatorWithTTL(cfg.DeduplicationTTL)
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	maxTimestampSkew := cfg.MaxTimestampSkew
	if maxTimestampSkew <= 0 {
		maxTimestampSkew = defaultTimestampSkew
	}
	return &WebhookHandler{
		secret:           cfg.Secret,
		registry:         registry,
		deduplicator:     deduplicator,
		onChallenge:      cfg.OnChallenge,
		onNotification:   cfg.OnNotification,
		onRevocation:     cfg.OnRevocation,
		now:              now,
		maxTimestampSkew: maxTimestampSkew,
	}
}

// ServeHTTP handles EventSub webhook requests.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bodyReader := http.MaxBytesReader(w, r.Body, maxWebhookBodyBytes)
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}

	messageID := r.Header.Get("Twitch-Eventsub-Message-Id")
	timestamp := r.Header.Get("Twitch-Eventsub-Message-Timestamp")
	signature := r.Header.Get("Twitch-Eventsub-Message-Signature")
	messageType := r.Header.Get("Twitch-Eventsub-Message-Type")
	if messageID == "" || timestamp == "" || signature == "" || messageType == "" {
		http.Error(w, "missing required eventsub headers", http.StatusBadRequest)
		return
	}
	if !h.verify(messageID, timestamp, signature, body) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}
	if !h.timestampIsFresh(timestamp) {
		http.Error(w, "stale timestamp", http.StatusUnauthorized)
		return
	}

	shouldDeduplicate := messageType != "webhook_callback_verification"
	reserved := false
	if shouldDeduplicate && h.deduplicator != nil {
		state, err := h.deduplicator.Reserve(ctx, messageID)
		if err != nil {
			http.Error(w, "dedup failed", http.StatusInternalServerError)
			return
		}
		if state == ReservationDuplicateCompleted {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if state == ReservationDuplicateInFlight {
			http.Error(w, "duplicate delivery in flight", http.StatusServiceUnavailable)
			return
		}
		reserved = true
	}

	switch messageType {
	case "webhook_callback_verification":
		var challenge Challenge
		if err := json.Unmarshal(body, &challenge); err != nil {
			http.Error(w, "decode challenge", http.StatusBadRequest)
			return
		}
		if h.onChallenge != nil {
			if err := h.onChallenge(ctx, challenge); err != nil {
				http.Error(w, "challenge callback failed", http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(challenge.Challenge)); err != nil {
			return
		}
	case "notification":
		var envelope struct {
			Subscription Subscription    `json:"subscription"`
			Event        json.RawMessage `json:"event"`
			Events       json.RawMessage `json:"events"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			if reserved {
				_ = h.deduplicator.Forget(ctx, messageID)
			}
			http.Error(w, "decode notification", http.StatusBadRequest)
			return
		}
		payload := envelope.Event
		if len(envelope.Events) != 0 {
			payload = envelope.Events
		}
		if len(payload) == 0 {
			if reserved {
				_ = h.deduplicator.Forget(ctx, messageID)
			}
			http.Error(w, "missing notification payload", http.StatusBadRequest)
			return
		}
		event, err := h.registry.Decode(envelope.Subscription.Type, envelope.Subscription.Version, payload)
		if err != nil {
			if reserved {
				_ = h.deduplicator.Forget(ctx, messageID)
			}
			http.Error(w, "decode typed event", http.StatusBadRequest)
			return
		}
		if h.onNotification != nil {
			if err := h.onNotification(ctx, Notification{
				MessageID:    messageID,
				Subscription: envelope.Subscription,
				Event:        event,
			}); err != nil {
				if reserved {
					_ = h.deduplicator.Forget(ctx, messageID)
				}
				http.Error(w, "notification callback failed", http.StatusInternalServerError)
				return
			}
		}
		if shouldDeduplicate && h.deduplicator != nil {
			if err := h.deduplicator.Complete(ctx, messageID); err != nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	case "revocation":
		var revocation Revocation
		if err := json.Unmarshal(body, &revocation); err != nil {
			if reserved {
				_ = h.deduplicator.Forget(ctx, messageID)
			}
			http.Error(w, "decode revocation", http.StatusBadRequest)
			return
		}
		if h.onRevocation != nil {
			revocation.MessageID = messageID
			if err := h.onRevocation(ctx, revocation); err != nil {
				if reserved {
					_ = h.deduplicator.Forget(ctx, messageID)
				}
				http.Error(w, "revocation callback failed", http.StatusInternalServerError)
				return
			}
		}
		if shouldDeduplicate && h.deduplicator != nil {
			if err := h.deduplicator.Complete(ctx, messageID); err != nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		if reserved {
			_ = h.deduplicator.Forget(ctx, messageID)
		}
		http.Error(w, "unsupported message type", http.StatusBadRequest)
	}
}

func (h *WebhookHandler) verify(messageID, timestamp, signature string, body []byte) bool {
	if h.secret == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write([]byte(messageID))
	mac.Write([]byte(timestamp))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}

func (h *WebhookHandler) timestampIsFresh(raw string) bool {
	if raw == "" {
		return false
	}

	timestamp, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return false
	}

	now := h.now()
	return !timestamp.Before(now.Add(-h.maxTimestampSkew)) && !timestamp.After(now.Add(h.maxTimestampSkew))
}
