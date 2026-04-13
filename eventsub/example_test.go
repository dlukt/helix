package eventsub_test

import (
	"context"
	"net/http"

	"github.com/dlukt/helix/eventsub"
)

func ExampleNewWebhookHandler() {
	handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
		Secret: "super-secret",
		OnNotification: func(context.Context, eventsub.Notification) error {
			return nil
		},
	})

	_ = http.Handler(handler)
}

func ExampleNewWebSocketClient() {
	client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
		URL: "wss://eventsub.wss.twitch.tv/ws",
		OnNotification: func(context.Context, eventsub.Notification) error {
			return nil
		},
	})

	_ = client
}
