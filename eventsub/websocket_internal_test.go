package eventsub

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestConnectAndAwaitWelcomeTimesOutWhenReplacementStaysSilent(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Upgrade() error = %v", err)
			return
		}
		defer conn.Close()
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	client := NewWebSocketClient(WebSocketClientConfig{
		URL: serverToWSURLInternal(server.URL),
	})

	start := time.Now()
	_, _, err := client.connectAndAwaitWelcome(context.Background(), serverToWSURLInternal(server.URL), 50*time.Millisecond)
	if err == nil {
		t.Fatal("connectAndAwaitWelcome() error = nil, want timeout")
	}
	if !isTimeout(err) {
		t.Fatalf("connectAndAwaitWelcome() error = %v, want timeout", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("connectAndAwaitWelcome() took %v, want < 500ms", elapsed)
	}
}

func TestRunTimesOutWhenInitialWelcomeNeverArrives(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Upgrade() error = %v", err)
			return
		}
		defer conn.Close()
		time.Sleep(500 * time.Millisecond)
	}))
	defer server.Close()

	client := NewWebSocketClient(WebSocketClientConfig{
		URL:            serverToWSURLInternal(server.URL),
		WelcomeTimeout: 50 * time.Millisecond,
	})

	done := make(chan error, 1)
	go func() {
		done <- client.Run(context.Background())
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Run() error = nil, want timeout")
		}
		if !isTimeout(err) {
			t.Fatalf("Run() error = %v, want timeout", err)
		}
	case <-time.After(300 * time.Millisecond):
		server.CloseClientConnections()
		t.Fatal("Run() did not time out waiting for initial welcome")
	}
}

func serverToWSURLInternal(raw string) string {
	return "ws" + raw[len("http"):]
}
