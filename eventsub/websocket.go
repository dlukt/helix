package eventsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultReconnectWelcomeTimeout = 10 * time.Second
	reconnectDrainGrace            = 100 * time.Millisecond
	reconnectRetryDelay            = 100 * time.Millisecond
)

// WebSocketClientConfig configures a WebSocket EventSub client.
type WebSocketClientConfig struct {
	URL                string
	Dialer             *websocket.Dialer
	WelcomeTimeout     time.Duration
	Registry           Registry
	OnSessionWelcome   func(context.Context, WebSocketSession) error
	OnSessionReconnect func(context.Context, WebSocketSession) error
	OnNotification     func(context.Context, Notification) error
	OnRevocation       func(context.Context, Revocation) error
}

// WebSocketSession describes a WebSocket EventSub session.
type WebSocketSession struct {
	ID                      string `json:"id"`
	Status                  string `json:"status"`
	KeepaliveTimeoutSeconds int    `json:"keepalive_timeout_seconds"`
	ReconnectURL            string `json:"reconnect_url"`
}

// WebSocketClient consumes EventSub WebSocket events.
type WebSocketClient struct {
	url                string
	dialer             *websocket.Dialer
	welcomeTimeout     time.Duration
	registry           Registry
	onSessionWelcome   func(context.Context, WebSocketSession) error
	onSessionReconnect func(context.Context, WebSocketSession) error
	onNotification     func(context.Context, Notification) error
	onRevocation       func(context.Context, Revocation) error
}

// NewWebSocketClient creates a new WebSocket EventSub client.
func NewWebSocketClient(cfg WebSocketClientConfig) *WebSocketClient {
	registry := cfg.Registry
	if registry == nil {
		registry = DefaultRegistry()
	}
	dialer := cfg.Dialer
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}
	return &WebSocketClient{
		url:                cfg.URL,
		dialer:             dialer,
		welcomeTimeout:     cfg.WelcomeTimeout,
		registry:           registry,
		onSessionWelcome:   cfg.OnSessionWelcome,
		onSessionReconnect: cfg.OnSessionReconnect,
		onNotification:     cfg.OnNotification,
		onRevocation:       cfg.OnRevocation,
	}
}

// Run connects to Twitch EventSub over WebSocket and dispatches messages until the context ends.
func (c *WebSocketClient) Run(ctx context.Context) error {
	conn, err := c.dial(ctx, c.url)
	if err != nil {
		return err
	}
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	var keepaliveTimeout time.Duration
	awaitingWelcome := true

	for {
		readTimeout := keepaliveTimeout
		if awaitingWelcome {
			readTimeout = c.resolveWelcomeTimeout()
		}
		message, err := c.readMessage(ctx, conn, readTimeout)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if isTimeout(err) {
				if awaitingWelcome {
					return err
				}
				nextConn, session, err := c.connectAndAwaitWelcome(ctx, c.url, c.reconnectWelcomeTimeout(keepaliveTimeout))
				if err != nil {
					return err
				}
				if c.onSessionWelcome != nil {
					if err := c.onSessionWelcome(ctx, session); err != nil {
						nextConn.Close()
						return err
					}
				}
				conn.Close()
				conn = nextConn
				keepaliveTimeout = keepaliveDuration(session)
				continue
			}
			return err
		}
		awaitingWelcome = false
		if next := keepaliveDuration(message.Payload.Session); next > 0 {
			keepaliveTimeout = next
		}
		if message.Metadata.MessageType == "session_reconnect" {
			nextConn, session, err := c.awaitReconnect(ctx, conn, keepaliveTimeout, message.Payload.Session.ReconnectURL)
			if err != nil {
				return err
			}
			if c.onSessionReconnect != nil {
				if err := c.onSessionReconnect(ctx, session); err != nil {
					nextConn.Close()
					return err
				}
			}
			conn.Close()
			conn = nextConn
			keepaliveTimeout = keepaliveDuration(session)
			continue
		}

		if err := c.dispatchMessage(ctx, message); err != nil {
			return err
		}
	}
}

func (c *WebSocketClient) awaitReconnect(ctx context.Context, conn *websocket.Conn, keepaliveTimeout time.Duration, reconnectURL string) (*websocket.Conn, WebSocketSession, error) {
	type reconnectResult struct {
		conn    *websocket.Conn
		session WebSocketSession
		err     error
	}

	var reconnectCancel context.CancelFunc
	var reconnectCh <-chan reconnectResult
	startReconnect := func() {
		reconnectCtx, cancelReconnectAttempt := context.WithCancel(ctx)
		reconnectCancel = cancelReconnectAttempt
		attemptCh := make(chan reconnectResult, 1)
		reconnectCh = attemptCh
		go func() {
			nextConn, session, err := c.connectAndAwaitWelcome(reconnectCtx, reconnectURL, c.reconnectWelcomeTimeout(keepaliveTimeout))
			attemptCh <- reconnectResult{conn: nextConn, session: session, err: err}
		}()
	}
	startReconnect()
	ownedByCaller := false
	var pendingReconnect *reconnectResult
	var retryCh <-chan time.Time
	var lastReconnectErr error
	defer func() {
		if reconnectCancel != nil {
			reconnectCancel()
		}
		if ownedByCaller {
			return
		}
		if pendingReconnect != nil {
			if pendingReconnect.conn != nil {
				pendingReconnect.conn.Close()
			}
			return
		}
		if reconnectCh == nil {
			return
		}
		select {
		case reconnect, ok := <-reconnectCh:
			if ok && reconnect.conn != nil {
				reconnect.conn.Close()
			}
		default:
			go func(reconnectCh <-chan reconnectResult) {
				reconnect := <-reconnectCh
				if reconnect.conn != nil {
					reconnect.conn.Close()
				}
			}(reconnectCh)
		}
	}()

	currentKeepalive := keepaliveTimeout
	readCh, err := startRead(conn, currentKeepalive)
	if err != nil {
		return nil, WebSocketSession{}, err
	}

	for {
		if pendingReconnect != nil {
			timer := time.NewTimer(reconnectDrainGrace)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return nil, WebSocketSession{}, ctx.Err()
			case result := <-readCh:
				if !timer.Stop() {
					<-timer.C
				}
				if result.err != nil {
					ownedByCaller = true
					return pendingReconnect.conn, pendingReconnect.session, nil
				}
				if next := keepaliveDuration(result.envelope.Payload.Session); next > 0 {
					currentKeepalive = next
				}
				if result.envelope.Metadata.MessageType != "session_reconnect" {
					if err := c.dispatchMessage(ctx, result.envelope); err != nil {
						return nil, WebSocketSession{}, err
					}
				}
				readCh, err = startRead(conn, currentKeepalive)
				if err != nil {
					return nil, WebSocketSession{}, err
				}
				continue
			case <-timer.C:
				ownedByCaller = true
				return pendingReconnect.conn, pendingReconnect.session, nil
			}
		}

		select {
		case <-ctx.Done():
			return nil, WebSocketSession{}, ctx.Err()
		case <-retryCh:
			retryCh = nil
			startReconnect()
		case reconnect := <-reconnectCh:
			reconnectCh = nil
			if reconnectCancel != nil {
				reconnectCancel()
				reconnectCancel = nil
			}
			if reconnect.err != nil {
				lastReconnectErr = reconnect.err
				retryCh = time.After(reconnectRetryDelay)
				continue
			}
			pendingReconnect = &reconnect
		case result := <-readCh:
			if result.err != nil {
				if retryCh != nil {
					retryCh = nil
					startReconnect()
				}
				if reconnectCh == nil {
					if lastReconnectErr != nil {
						return nil, WebSocketSession{}, lastReconnectErr
					}
					return nil, WebSocketSession{}, result.err
				}
				select {
				case <-ctx.Done():
					return nil, WebSocketSession{}, ctx.Err()
				case reconnect := <-reconnectCh:
					reconnectCh = nil
					if reconnectCancel != nil {
						reconnectCancel()
						reconnectCancel = nil
					}
					if reconnect.err != nil {
						return nil, WebSocketSession{}, reconnect.err
					}
					ownedByCaller = true
					return reconnect.conn, reconnect.session, nil
				}
			}
			if next := keepaliveDuration(result.envelope.Payload.Session); next > 0 {
				currentKeepalive = next
			}
			if result.envelope.Metadata.MessageType != "session_reconnect" {
				if err := c.dispatchMessage(ctx, result.envelope); err != nil {
					return nil, WebSocketSession{}, err
				}
			}
			readCh, err = startRead(conn, currentKeepalive)
			if err != nil {
				return nil, WebSocketSession{}, err
			}
		}
	}
}

func (c *WebSocketClient) connectAndAwaitWelcome(ctx context.Context, targetURL string, welcomeTimeout time.Duration) (*websocket.Conn, WebSocketSession, error) {
	conn, err := c.dial(ctx, targetURL)
	if err != nil {
		return nil, WebSocketSession{}, err
	}

	message, err := c.readMessage(ctx, conn, welcomeTimeout)
	if err != nil {
		conn.Close()
		return nil, WebSocketSession{}, err
	}
	if message.Metadata.MessageType != "session_welcome" {
		conn.Close()
		return nil, WebSocketSession{}, fmt.Errorf("eventsub: expected session_welcome on reconnect, got %q", message.Metadata.MessageType)
	}
	return conn, message.Payload.Session, nil
}

func (c *WebSocketClient) dial(ctx context.Context, targetURL string) (*websocket.Conn, error) {
	if targetURL == "" {
		return nil, errors.New("eventsub: websocket url is required")
	}

	conn, _, err := c.dialer.DialContext(ctx, targetURL, nil)
	if err != nil {
		return nil, err
	}
	conn.SetPingHandler(func(appData string) error {
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(5*time.Second))
	})
	return conn, nil
}

func (c *WebSocketClient) readMessage(ctx context.Context, conn *websocket.Conn, keepaliveTimeout time.Duration) (webSocketEnvelope, error) {
	done, err := startRead(conn, keepaliveTimeout)
	if err != nil {
		return webSocketEnvelope{}, err
	}

	select {
	case <-ctx.Done():
		conn.Close()
		return webSocketEnvelope{}, ctx.Err()
	case result := <-done:
		return result.envelope, result.err
	}
}

func (c *WebSocketClient) dispatchMessage(ctx context.Context, message webSocketEnvelope) error {
	switch message.Metadata.MessageType {
	case "session_welcome":
		if c.onSessionWelcome != nil {
			if err := c.onSessionWelcome(ctx, message.Payload.Session); err != nil {
				return err
			}
		}
	case "notification":
		event, err := c.registry.Decode(message.Payload.Subscription.Type, message.Payload.Subscription.Version, message.Payload.Event)
		if err != nil {
			return err
		}
		if c.onNotification != nil {
			if err := c.onNotification(ctx, Notification{
				MessageID:    message.Metadata.MessageID,
				Subscription: message.Payload.Subscription,
				Event:        event,
			}); err != nil {
				return err
			}
		}
	case "revocation":
		if c.onRevocation != nil {
			if err := c.onRevocation(ctx, Revocation{
				MessageID:    message.Metadata.MessageID,
				Subscription: message.Payload.Subscription,
			}); err != nil {
				return err
			}
		}
	case "session_keepalive":
		return nil
	default:
		return fmt.Errorf("eventsub: unsupported websocket message type %q", message.Metadata.MessageType)
	}
	return nil
}

type readResult struct {
	envelope webSocketEnvelope
	err      error
}

func startRead(conn *websocket.Conn, keepaliveTimeout time.Duration) (<-chan readResult, error) {
	if keepaliveTimeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(keepaliveTimeout)); err != nil {
			return nil, err
		}
	} else {
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			return nil, err
		}
	}

	done := make(chan readResult, 1)
	go func() {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			done <- readResult{err: err}
			return
		}
		var envelope webSocketEnvelope
		if err := json.Unmarshal(raw, &envelope); err != nil {
			done <- readResult{err: err}
			return
		}
		done <- readResult{envelope: envelope}
	}()
	return done, nil
}

func (c *WebSocketClient) reconnectWelcomeTimeout(keepaliveTimeout time.Duration) time.Duration {
	if keepaliveTimeout > 0 {
		return keepaliveTimeout
	}
	return c.resolveWelcomeTimeout()
}

func (c *WebSocketClient) resolveWelcomeTimeout() time.Duration {
	if c.welcomeTimeout > 0 {
		return c.welcomeTimeout
	}
	return defaultReconnectWelcomeTimeout
}

func keepaliveDuration(session WebSocketSession) time.Duration {
	if session.KeepaliveTimeoutSeconds <= 0 {
		return 0
	}
	return time.Duration(session.KeepaliveTimeoutSeconds) * time.Second
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

type webSocketEnvelope struct {
	Metadata struct {
		MessageID   string `json:"message_id"`
		MessageType string `json:"message_type"`
	} `json:"metadata"`
	Payload struct {
		Session      WebSocketSession `json:"session"`
		Subscription Subscription     `json:"subscription"`
		Event        json.RawMessage  `json:"event"`
	} `json:"payload"`
}
