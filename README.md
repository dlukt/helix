# helix

`helix` is a Go package for Twitch's Helix API, OAuth flows, and EventSub runtimes.

## Current coverage

- Helix client transport with typed response metadata, rate-limit parsing, request ID propagation, 401 retry with invalidating token sources, and typed API errors
- Helix service groups for users, channels, streams, chat, moderation, raids, schedule, markers, games, search, analytics, bits, hype train, entitlements, whispers, teams, content classification labels, extensions, channel points, guest star, polls, predictions, goals, charity, subscriptions, clips, videos, and EventSub subscription management
- Live Helix reference coverage is effectively complete apart from Twitch's deprecated stream-tags endpoints and small naming-polish gaps where one SDK method covers multiple closely related reference pages
- OAuth client helpers for authorization-code, client-credentials, implicit-flow URL generation, device authorization, device-code polling, refresh, validate, and revoke
- EventSub webhook verification and dispatch with timestamp validation, deduplication, and a generated typed registry covering 82 subscription type/version pairs
- EventSub WebSocket consumption with reconnect handling, typed close errors, optional hard-disconnect recovery via subscription recreation, and Helix conduit management support

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
