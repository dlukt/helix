package eventsub_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/dlukt/helix/eventsub"
	"github.com/gorilla/websocket"
)

func TestWebSocketClientReceivesWelcomeAndTypedNotification(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var welcomes []string
	var notifications []eventsub.Notification

	server := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":           "notification-1",
				"message_type":         "notification",
				"message_timestamp":    "2024-04-11T12:00:01Z",
				"subscription_type":    "channel.follow",
				"subscription_version": "2",
			},
			"payload": map[string]any{
				"subscription": map[string]any{
					"id":      "sub-1",
					"type":    "channel.follow",
					"version": "2",
					"status":  "enabled",
				},
				"event": map[string]any{
					"user_id":                "777",
					"user_login":             "viewer",
					"user_name":              "Viewer",
					"broadcaster_user_id":    "123",
					"broadcaster_user_login": "caster",
					"broadcaster_user_name":  "Caster",
					"followed_at":            "2024-04-11T12:00:00Z",
				},
			},
		})
		time.Sleep(50 * time.Millisecond)
	})
	defer server.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server.URL),
		OnSessionWelcome: func(_ context.Context, session eventsub.WebSocketSession) error {
			mu.Lock()
			defer mu.Unlock()
			welcomes = append(welcomes, session.ID)
			return nil
		},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			mu.Lock()
			defer mu.Unlock()
			notifications = append(notifications, notification)
			cancel()
			return nil
		},
	})

	err := client.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("Run() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(welcomes) != 1 || welcomes[0] != "session-1" {
		t.Fatalf("welcomes = %v, want [session-1]", welcomes)
	}
	if len(notifications) != 1 {
		t.Fatalf("len(notifications) = %d, want 1", len(notifications))
	}
	if got := notifications[0].MessageID; got != "notification-1" {
		t.Fatalf("notification MessageID = %q, want %q", got, "notification-1")
	}
	if _, ok := notifications[0].Event.(eventsub.ChannelFollowEvent); !ok {
		t.Fatalf("notification event type = %T, want ChannelFollowEvent", notifications[0].Event)
	}
}

func TestWebSocketClientReconnectsUsingReconnectURL(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var welcomes []string
	var reconnects []string
	var notifications []eventsub.Notification

	server2 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		time.Sleep(150 * time.Millisecond)
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-2",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:02Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-2",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":           "notification-2",
				"message_type":         "notification",
				"message_timestamp":    "2024-04-11T12:00:03Z",
				"subscription_type":    "stream.online",
				"subscription_version": "1",
			},
			"payload": map[string]any{
				"subscription": map[string]any{
					"id":      "sub-2",
					"type":    "stream.online",
					"version": "1",
					"status":  "enabled",
				},
				"event": map[string]any{
					"id":                     "stream-1",
					"broadcaster_user_id":    "123",
					"broadcaster_user_login": "caster",
					"broadcaster_user_name":  "Caster",
					"type":                   "live",
					"started_at":             "2024-04-11T12:00:03Z",
				},
			},
		})
		time.Sleep(50 * time.Millisecond)
	})
	defer server2.Close()

	server1 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "reconnect-1",
				"message_type":      "session_reconnect",
				"message_timestamp": "2024-04-11T12:00:01Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "reconnecting",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             serverToWSURL(server2.URL),
				},
			},
		})
		time.Sleep(100 * time.Millisecond)
	})
	defer server1.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server1.URL),
		OnSessionWelcome: func(_ context.Context, session eventsub.WebSocketSession) error {
			mu.Lock()
			defer mu.Unlock()
			welcomes = append(welcomes, session.ID)
			return nil
		},
		OnSessionReconnect: func(_ context.Context, session eventsub.WebSocketSession) error {
			mu.Lock()
			defer mu.Unlock()
			reconnects = append(reconnects, session.ID)
			return nil
		},
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			mu.Lock()
			defer mu.Unlock()
			notifications = append(notifications, notification)
			cancel()
			return nil
		},
	})

	err := client.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("Run() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(welcomes) != 1 || welcomes[0] != "session-1" {
		t.Fatalf("welcomes = %v, want [session-1]", welcomes)
	}
	if len(reconnects) != 1 || reconnects[0] != "session-2" {
		t.Fatalf("reconnects = %v, want [session-2]", reconnects)
	}
	if len(notifications) != 1 {
		t.Fatalf("len(notifications) = %d, want 1", len(notifications))
	}
	if _, ok := notifications[0].Event.(eventsub.StreamOnlineEvent); !ok {
		t.Fatalf("notification event type = %T, want StreamOnlineEvent", notifications[0].Event)
	}
}

func TestWebSocketClientProcessesOldSocketNotificationsDuringReconnectHandoff(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var mu sync.Mutex
	var notifications []eventsub.Notification

	server2 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		time.Sleep(10 * time.Millisecond)
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-2",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:02Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-2",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
	})
	defer server2.Close()

	server1 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "reconnect-1",
				"message_type":      "session_reconnect",
				"message_timestamp": "2024-04-11T12:00:01Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "reconnecting",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             serverToWSURL(server2.URL),
				},
			},
		})
		time.Sleep(50 * time.Millisecond)
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":           "notification-old",
				"message_type":         "notification",
				"message_timestamp":    "2024-04-11T12:00:01Z",
				"subscription_type":    "channel.follow",
				"subscription_version": "2",
			},
			"payload": map[string]any{
				"subscription": map[string]any{
					"id":      "sub-1",
					"type":    "channel.follow",
					"version": "2",
					"status":  "enabled",
				},
				"event": map[string]any{
					"user_id":                "777",
					"user_login":             "viewer",
					"user_name":              "Viewer",
					"broadcaster_user_id":    "123",
					"broadcaster_user_login": "caster",
					"broadcaster_user_name":  "Caster",
					"followed_at":            "2024-04-11T12:00:00Z",
				},
			},
		})
		time.Sleep(25 * time.Millisecond)
		_ = conn.Close()
	})
	defer server1.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server1.URL),
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			mu.Lock()
			defer mu.Unlock()
			notifications = append(notifications, notification)
			cancel()
			return nil
		},
	})

	err := client.Run(ctx)
	if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		t.Fatalf("Run() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(notifications) != 1 {
		t.Fatalf("len(notifications) = %d, want 1", len(notifications))
	}
	if _, ok := notifications[0].Event.(eventsub.ChannelFollowEvent); !ok {
		t.Fatalf("notification event type = %T, want ChannelFollowEvent", notifications[0].Event)
	}
}

func TestWebSocketClientDrainsBufferedOldSocketNotificationsDuringReconnectHandoff(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var mu sync.Mutex
	var notifications []eventsub.Notification

	server2 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		time.Sleep(10 * time.Millisecond)
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-2",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:02Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-2",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		time.Sleep(200 * time.Millisecond)
	})
	defer server2.Close()

	server1 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "reconnect-1",
				"message_type":      "session_reconnect",
				"message_timestamp": "2024-04-11T12:00:01Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "reconnecting",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             serverToWSURL(server2.URL),
				},
			},
		})
		time.Sleep(50 * time.Millisecond)
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":           "notification-old-1",
				"message_type":         "notification",
				"message_timestamp":    "2024-04-11T12:00:01Z",
				"subscription_type":    "channel.follow",
				"subscription_version": "2",
			},
			"payload": map[string]any{
				"subscription": map[string]any{
					"id":      "sub-1",
					"type":    "channel.follow",
					"version": "2",
					"status":  "enabled",
				},
				"event": map[string]any{
					"user_id":                "777",
					"user_login":             "viewer",
					"user_name":              "Viewer",
					"broadcaster_user_id":    "123",
					"broadcaster_user_login": "caster",
					"broadcaster_user_name":  "Caster",
					"followed_at":            "2024-04-11T12:00:00Z",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":           "notification-old-2",
				"message_type":         "notification",
				"message_timestamp":    "2024-04-11T12:00:02Z",
				"subscription_type":    "channel.follow",
				"subscription_version": "2",
			},
			"payload": map[string]any{
				"subscription": map[string]any{
					"id":      "sub-1",
					"type":    "channel.follow",
					"version": "2",
					"status":  "enabled",
				},
				"event": map[string]any{
					"user_id":                "888",
					"user_login":             "viewer2",
					"user_name":              "Viewer Two",
					"broadcaster_user_id":    "123",
					"broadcaster_user_login": "caster",
					"broadcaster_user_name":  "Caster",
					"followed_at":            "2024-04-11T12:00:01Z",
				},
			},
		})
		time.Sleep(25 * time.Millisecond)
		_ = conn.Close()
	})
	defer server1.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server1.URL),
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			mu.Lock()
			defer mu.Unlock()
			notifications = append(notifications, notification)
			if len(notifications) == 2 {
				cancel()
			}
			return nil
		},
	})

	err := client.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("Run() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(notifications) != 2 {
		t.Fatalf("len(notifications) = %d, want 2", len(notifications))
	}
	if notifications[0].MessageID != "notification-old-1" || notifications[1].MessageID != "notification-old-2" {
		t.Fatalf("message IDs = [%s %s], want [notification-old-1 notification-old-2]", notifications[0].MessageID, notifications[1].MessageID)
	}
}

func TestWebSocketClientSwitchesToReplacementSocketAfterWelcome(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var mu sync.Mutex
	var notifications []eventsub.Notification

	server2 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-2",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:02Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-2",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		time.Sleep(20 * time.Millisecond)
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":           "notification-new",
				"message_type":         "notification",
				"message_timestamp":    "2024-04-11T12:00:03Z",
				"subscription_type":    "stream.online",
				"subscription_version": "1",
			},
			"payload": map[string]any{
				"subscription": map[string]any{
					"id":      "sub-2",
					"type":    "stream.online",
					"version": "1",
					"status":  "enabled",
				},
				"event": map[string]any{
					"id":                     "stream-1",
					"broadcaster_user_id":    "123",
					"broadcaster_user_login": "caster",
					"broadcaster_user_name":  "Caster",
					"type":                   "live",
					"started_at":             "2024-04-11T12:00:03Z",
				},
			},
		})
		time.Sleep(50 * time.Millisecond)
	})
	defer server2.Close()

	server1 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "reconnect-1",
				"message_type":      "session_reconnect",
				"message_timestamp": "2024-04-11T12:00:01Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "reconnecting",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             serverToWSURL(server2.URL),
				},
			},
		})
		time.Sleep(time.Second)
	})
	defer server1.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server1.URL),
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			mu.Lock()
			defer mu.Unlock()
			notifications = append(notifications, notification)
			cancel()
			return nil
		},
	})

	err := client.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("Run() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(notifications) != 1 {
		t.Fatalf("len(notifications) = %d, want 1", len(notifications))
	}
	if got := notifications[0].MessageID; got != "notification-new" {
		t.Fatalf("MessageID = %q, want %q", got, "notification-new")
	}
}

func TestWebSocketClientRetriesReconnectWhileOldSocketStaysHealthy(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var mu sync.Mutex
	var notifications []eventsub.Notification
	var replacementMu sync.Mutex
	var replacementConnections int

	server2 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		replacementMu.Lock()
		replacementConnections++
		firstAttempt := replacementConnections == 1
		replacementMu.Unlock()
		if firstAttempt {
			_ = conn.Close()
			return
		}

		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-2",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:02Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-2",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		time.Sleep(20 * time.Millisecond)
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":           "notification-new",
				"message_type":         "notification",
				"message_timestamp":    "2024-04-11T12:00:03Z",
				"subscription_type":    "stream.online",
				"subscription_version": "1",
			},
			"payload": map[string]any{
				"subscription": map[string]any{
					"id":      "sub-2",
					"type":    "stream.online",
					"version": "1",
					"status":  "enabled",
				},
				"event": map[string]any{
					"id":                     "stream-1",
					"broadcaster_user_id":    "123",
					"broadcaster_user_login": "caster",
					"broadcaster_user_name":  "Caster",
					"type":                   "live",
					"started_at":             "2024-04-11T12:00:03Z",
				},
			},
		})
		time.Sleep(50 * time.Millisecond)
	})
	defer server2.Close()

	server1 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "reconnect-1",
				"message_type":      "session_reconnect",
				"message_timestamp": "2024-04-11T12:00:01Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "reconnecting",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             serverToWSURL(server2.URL),
				},
			},
		})
		time.Sleep(150 * time.Millisecond)
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":           "notification-old",
				"message_type":         "notification",
				"message_timestamp":    "2024-04-11T12:00:02Z",
				"subscription_type":    "channel.follow",
				"subscription_version": "2",
			},
			"payload": map[string]any{
				"subscription": map[string]any{
					"id":      "sub-1",
					"type":    "channel.follow",
					"version": "2",
					"status":  "enabled",
				},
				"event": map[string]any{
					"user_id":                "777",
					"user_login":             "viewer",
					"user_name":              "Viewer",
					"broadcaster_user_id":    "123",
					"broadcaster_user_login": "caster",
					"broadcaster_user_name":  "Caster",
					"followed_at":            "2024-04-11T12:00:00Z",
				},
			},
		})
		time.Sleep(400 * time.Millisecond)
	})
	defer server1.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server1.URL),
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			mu.Lock()
			defer mu.Unlock()
			notifications = append(notifications, notification)
			if len(notifications) == 2 {
				cancel()
			}
			return nil
		},
	})

	err := client.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("Run() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(notifications) != 2 {
		t.Fatalf("len(notifications) = %d, want 2", len(notifications))
	}
	if notifications[0].MessageID != "notification-old" || notifications[1].MessageID != "notification-new" {
		t.Fatalf("message IDs = [%s %s], want [notification-old notification-new]", notifications[0].MessageID, notifications[1].MessageID)
	}
	replacementMu.Lock()
	defer replacementMu.Unlock()
	if replacementConnections < 2 {
		t.Fatalf("replacement connections = %d, want at least 2", replacementConnections)
	}
}

func TestWebSocketClientWaitsForReplacementWelcomeAfterOldSocketCloses(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var welcomes []string
	var reconnects []string

	server2 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		time.Sleep(150 * time.Millisecond)
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-2",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:02Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-2",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		time.Sleep(100 * time.Millisecond)
		cancel()
	})
	defer server2.Close()

	server1 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "reconnect-1",
				"message_type":      "session_reconnect",
				"message_timestamp": "2024-04-11T12:00:01Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "reconnecting",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             serverToWSURL(server2.URL),
				},
			},
		})
		time.Sleep(25 * time.Millisecond)
		_ = conn.Close()
	})
	defer server1.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server1.URL),
		OnSessionWelcome: func(_ context.Context, session eventsub.WebSocketSession) error {
			mu.Lock()
			defer mu.Unlock()
			welcomes = append(welcomes, session.ID)
			return nil
		},
		OnSessionReconnect: func(_ context.Context, session eventsub.WebSocketSession) error {
			mu.Lock()
			defer mu.Unlock()
			reconnects = append(reconnects, session.ID)
			return nil
		},
	})

	err := client.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("Run() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(welcomes) != 1 || welcomes[0] != "session-1" {
		t.Fatalf("welcomes = %v, want [session-1]", welcomes)
	}
	if len(reconnects) != 1 || reconnects[0] != "session-2" {
		t.Fatalf("reconnects = %v, want [session-2]", reconnects)
	}
}

func TestWebSocketClientClosesReplacementSocketWhenReconnectIsCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	replacementClosed := make(chan struct{}, 1)

	server2 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-2",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:02Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-2",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				select {
				case replacementClosed <- struct{}{}:
				default:
				}
				return
			}
		}
	})
	defer server2.Close()

	server1 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "reconnect-1",
				"message_type":      "session_reconnect",
				"message_timestamp": "2024-04-11T12:00:01Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "reconnecting",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             serverToWSURL(server2.URL),
				},
			},
		})
		time.Sleep(25 * time.Millisecond)
		cancel()
		time.Sleep(100 * time.Millisecond)
	})
	defer server1.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server1.URL),
	})

	err := client.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}

	select {
	case <-replacementClosed:
	case <-time.After(2 * time.Second):
		t.Fatal("replacement websocket was not closed after cancellation")
	}
}

func TestWebSocketClientPropagatesMessageIDOnRevocation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var revocations []eventsub.Revocation
	server := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "revocation-1",
				"message_type":      "revocation",
				"message_timestamp": "2024-04-11T12:00:01Z",
			},
			"payload": map[string]any{
				"subscription": map[string]any{
					"id":      "sub-1",
					"type":    "stream.online",
					"version": "1",
					"status":  "authorization_revoked",
				},
			},
		})
		time.Sleep(50 * time.Millisecond)
	})
	defer server.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server.URL),
		OnRevocation: func(_ context.Context, revocation eventsub.Revocation) error {
			revocations = append(revocations, revocation)
			cancel()
			return nil
		},
	})

	err := client.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("Run() error = %v", err)
	}

	if len(revocations) != 1 {
		t.Fatalf("len(revocations) = %d, want 1", len(revocations))
	}
	if got := revocations[0].MessageID; got != "revocation-1" {
		t.Fatalf("revocation MessageID = %q, want %q", got, "revocation-1")
	}
}

func TestWebSocketClientClosesReplacementSocketOnReconnectAbort(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	replacementClosed := make(chan struct{}, 1)

	server2 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-2",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:02Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-2",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				select {
				case replacementClosed <- struct{}{}:
				default:
				}
				return
			}
		}
	})
	defer server2.Close()

	server1 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "reconnect-1",
				"message_type":      "session_reconnect",
				"message_timestamp": "2024-04-11T12:00:01Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "reconnecting",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             serverToWSURL(server2.URL),
				},
			},
		})
		time.Sleep(25 * time.Millisecond)
		_ = conn.Close()
	})
	defer server1.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server1.URL),
	})

	err := client.Run(ctx)
	if err == nil || err.Error() != "websocket: close 1006 (abnormal closure): unexpected EOF" {
		t.Fatalf("Run() error = %v, want old-socket EOF during reconnect abort", err)
	}

	select {
	case <-replacementClosed:
	case <-time.After(2 * time.Second):
		t.Fatal("replacement websocket was not closed after reconnect abort")
	}
}

func TestWebSocketClientClosesPendingReplacementSocketOnCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	replacementWelcomed := make(chan struct{}, 1)
	replacementClosed := make(chan struct{}, 1)

	server2 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-2",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:02Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-2",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		select {
		case replacementWelcomed <- struct{}{}:
		default:
		}
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				select {
				case replacementClosed <- struct{}{}:
				default:
				}
				return
			}
		}
	})
	defer server2.Close()

	server1 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "reconnect-1",
				"message_type":      "session_reconnect",
				"message_timestamp": "2024-04-11T12:00:01Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "reconnecting",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             serverToWSURL(server2.URL),
				},
			},
		})
		time.Sleep(time.Second)
	})
	defer server1.Close()

	go func() {
		<-replacementWelcomed
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server1.URL),
	})

	err := client.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}

	select {
	case <-replacementClosed:
	case <-time.After(2 * time.Second):
		t.Fatal("pending replacement websocket was not closed after cancellation")
	}
}

func TestWebSocketClientReconnectsAfterKeepaliveTimeout(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var sessions []string

	var connections int
	server := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		connections++
		switch connections {
		case 1:
			writeWSMessage(t, conn, map[string]any{
				"metadata": map[string]any{
					"message_id":        "welcome-1",
					"message_type":      "session_welcome",
					"message_timestamp": "2024-04-11T12:00:00Z",
				},
				"payload": map[string]any{
					"session": map[string]any{
						"id":                        "session-1",
						"status":                    "connected",
						"keepalive_timeout_seconds": 1,
						"reconnect_url":             "",
					},
				},
			})
			time.Sleep(1500 * time.Millisecond)
		default:
			writeWSMessage(t, conn, map[string]any{
				"metadata": map[string]any{
					"message_id":        "welcome-2",
					"message_type":      "session_welcome",
					"message_timestamp": "2024-04-11T12:00:02Z",
				},
				"payload": map[string]any{
					"session": map[string]any{
						"id":                        "session-2",
						"status":                    "connected",
						"keepalive_timeout_seconds": 1,
						"reconnect_url":             "",
					},
				},
			})
			time.Sleep(100 * time.Millisecond)
			cancel()
		}
	})
	defer server.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server.URL),
		OnSessionWelcome: func(_ context.Context, session eventsub.WebSocketSession) error {
			mu.Lock()
			defer mu.Unlock()
			sessions = append(sessions, session.ID)
			return nil
		},
	})

	err := client.Run(ctx)
	if err != nil && err != context.Canceled {
		t.Fatalf("Run() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(sessions) < 2 || sessions[0] != "session-1" || sessions[1] != "session-2" {
		t.Fatalf("sessions = %v, want [session-1 session-2]", sessions)
	}
}

func TestWebSocketClientClosesReconnectedSessionOnLaterExit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	closed := make(chan struct{}, 1)

	server2 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-2",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:02Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-2",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":           "notification-2",
				"message_type":         "notification",
				"message_timestamp":    "2024-04-11T12:00:03Z",
				"subscription_type":    "stream.online",
				"subscription_version": "1",
			},
			"payload": map[string]any{
				"subscription": map[string]any{
					"id":      "sub-2",
					"type":    "stream.online",
					"version": "1",
					"status":  "enabled",
				},
				"event": map[string]any{
					"id":                     "stream-1",
					"broadcaster_user_id":    "123",
					"broadcaster_user_login": "caster",
					"broadcaster_user_name":  "Caster",
					"type":                   "live",
					"started_at":             "2024-04-11T12:00:03Z",
				},
			},
		})
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				select {
				case closed <- struct{}{}:
				default:
				}
				return
			}
		}
	})
	defer server2.Close()

	server1 := newTestWSServer(t, func(conn *websocket.Conn, serverURL string) {
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "welcome-1",
				"message_type":      "session_welcome",
				"message_timestamp": "2024-04-11T12:00:00Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "connected",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             "",
				},
			},
		})
		writeWSMessage(t, conn, map[string]any{
			"metadata": map[string]any{
				"message_id":        "reconnect-1",
				"message_type":      "session_reconnect",
				"message_timestamp": "2024-04-11T12:00:01Z",
			},
			"payload": map[string]any{
				"session": map[string]any{
					"id":                        "session-1",
					"status":                    "reconnecting",
					"keepalive_timeout_seconds": 10,
					"reconnect_url":             serverToWSURL(server2.URL),
				},
			},
		})
		time.Sleep(200 * time.Millisecond)
	})
	defer server1.Close()

	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: serverToWSURL(server1.URL),
		OnNotification: func(_ context.Context, notification eventsub.Notification) error {
			return assertiveWSError("stop after reconnect")
		},
	})

	err := client.Run(ctx)
	if err == nil || err.Error() != "stop after reconnect" {
		t.Fatalf("Run() error = %v, want stop after reconnect", err)
	}

	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("reconnected websocket was not closed on exit")
	}
}

func newTestWSServer(t *testing.T, onConnect func(conn *websocket.Conn, serverURL string)) *httptest.Server {
	t.Helper()

	upgrader := websocket.Upgrader{}
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Upgrade() error = %v", err)
			return
		}
		defer conn.Close()
		onConnect(conn, serverToWSURL(server.URL))
	}))
	return server
}

func writeWSMessage(t *testing.T, conn *websocket.Conn, payload any) {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, raw); err != nil {
		t.Fatalf("WriteMessage() error = %v", err)
	}
}

func serverToWSURL(raw string) string {
	return "ws" + raw[len("http"):]
}

type assertiveWSError string

func (e assertiveWSError) Error() string {
	return string(e)
}
