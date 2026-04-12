package eventsub

import (
	"encoding/json"
	"fmt"
	"sync"
)

//go:generate go run ../internal/cmd/genmanifest

// Registry decodes typed EventSub event payloads.
type Registry interface {
	Decode(subscriptionType, version string, raw json.RawMessage) (any, error)
}

type decodeFunc func(json.RawMessage) (any, error)

// DynamicRegistry stores event decoders keyed by type and version.
type DynamicRegistry struct {
	mu       sync.RWMutex
	decoders map[string]decodeFunc
}

// NewRegistry creates an empty registry.
func NewRegistry() *DynamicRegistry {
	return &DynamicRegistry{decoders: map[string]decodeFunc{}}
}

// Register adds a decoder for a subscription type and version.
func (r *DynamicRegistry) Register(subscriptionType, version string, fn decodeFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.decoders[key(subscriptionType, version)] = fn
}

// Decode decodes a typed event payload.
func (r *DynamicRegistry) Decode(subscriptionType, version string, raw json.RawMessage) (any, error) {
	r.mu.RLock()
	fn, ok := r.decoders[key(subscriptionType, version)]
	r.mu.RUnlock()
	if !ok {
		return UnknownEvent{Raw: append(json.RawMessage(nil), raw...), SubscriptionType: subscriptionType, Version: version}, nil
	}
	return fn(raw)
}

// UnknownEvent is returned when no typed decoder is registered.
type UnknownEvent struct {
	Raw              json.RawMessage
	SubscriptionType string
	Version          string
}

func key(subscriptionType, version string) string {
	return subscriptionType + "@" + version
}

var (
	defaultRegistry Registry
	defaultOnce     sync.Once
)

// DefaultRegistry returns the generated default registry.
func DefaultRegistry() Registry {
	defaultOnce.Do(func() {
		defaultRegistry = newDefaultRegistry()
	})
	return defaultRegistry
}

func decodeInto[T any](raw json.RawMessage) (any, error) {
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("eventsub: decode event: %w", err)
	}
	return value, nil
}
