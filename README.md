# helix

`helix` is a Go package for Twitch's Helix API, OAuth flows, and EventSub runtimes.

## Current coverage

- Helix client transport with typed response metadata, rate-limit parsing, request ID propagation, 401 retry with invalidating token sources, and typed API errors
- Helix service groups for users, channels, streams, games, search, clips, videos, and EventSub subscription management
- OAuth client helpers for authorization-code, client-credentials, implicit-flow URL generation, device authorization, device-code polling, refresh, validate, and revoke
- EventSub webhook verification and dispatch with timestamp validation and deduplication
- EventSub WebSocket consumption with reconnect handling, typed close errors, and optional hard-disconnect recovery via subscription recreation

## Install

```bash
go get github.com/dlukt/helix
```

## Quick start

```go
oauthClient := oauth.NewClient(oauth.Config{
    ClientID:     "client-id",
    ClientSecret: "client-secret",
})

tokenSource := oauth.NewAppSource(oauthClient, nil)

client, err := helix.New(helix.Config{
    ClientID:    "client-id",
    TokenSource: tokenSource,
})
if err != nil {
    panic(err)
}

users, _, err := client.Users.Get(ctx, helix.GetUsersParams{
    Logins: []string{"twitchdev"},
})
if err != nil {
    panic(err)
}

_ = users
```

## EventSub webhook

```go
handler := eventsub.NewWebhookHandler(eventsub.WebhookHandlerConfig{
    Secret: "super-secret",
    OnNotification: func(ctx context.Context, n eventsub.Notification) error {
        return nil
    },
})

http.Handle("/eventsub", handler)
```

## EventSub WebSocket

```go
client := eventsub.NewWebSocketClient(eventsub.WebSocketClientConfig{
    URL: "wss://eventsub.wss.twitch.tv/ws",
    OnNotification: func(ctx context.Context, n eventsub.Notification) error {
        return nil
    },
})

if err := client.Run(ctx); err != nil {
    panic(err)
}
```

## Development

```bash
go test ./...
go generate ./eventsub
```
