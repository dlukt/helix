// Package eventsub provides Twitch EventSub webhook and WebSocket runtimes.
//
// It includes typed notification decoding through a registry, signature
// verification for webhook deliveries, deduplication helpers, reconnect-aware
// WebSocket consumption, and optional hard-disconnect recovery via Helix
// subscription recreation.
package eventsub
