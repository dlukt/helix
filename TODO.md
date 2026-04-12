# TODO

This file lists what is still missing relative to the agreed implementation plan and the current code in this repository.

## Helix API coverage

- Add the rest of the typed Helix service groups on `*helix.Client`.
  - Currently implemented: `Users.Get`, `EventSub.Create`, `EventSub.List`, `EventSub.Delete`.
  - Missing from the plan: the broader Helix API surface such as streams, channels, chat, clips, videos, moderation, schedules, analytics, bits, ads, polls, predictions, raids, whispers, goals, hype train, subscriptions, teams, extensions, games, search, markers, and other documented Helix sections.
- Expand the checked-in manifest and generator so typed request and response code is generated for the Helix API itself, not just the small EventSub registry.
- Add more representative typed models.
  - `users.go` only models a subset of `Get Users` fields.
  - `eventsub_api.go` currently uses generic `map[string]string` for conditions instead of generated condition types per subscription type and version.

## Helix transport behavior

- Add a `RequestID` field to `helix.Response` and populate it from the response headers.
- Add `401`-aware retry behavior in `helix.Client` that cooperates with refresh-capable token sources.
  - Current behavior: the token source is called before each request, but `Do` does not detect an unauthorized response and force a token refresh/retry.
- Add stronger API error decoding.
  - Current behavior: `APIError` only keeps status code, raw body, and rate-limit info.
  - Missing: parse Twitch error payload fields such as error code/name and message into typed fields.
- Add typed pagination helpers beyond exposing `Pagination.Cursor`.
  - The plan intentionally excluded an auto-paginator for v1, but request params and cursor handling still need to be covered consistently across generated endpoints.
- Add tests for:
  - repeated query parameters on more than one endpoint shape
  - zero-body success responses beyond current EventSub delete coverage
  - non-JSON and malformed JSON error paths
  - request ID propagation

## OAuth package

- Add explicit support for implicit-flow helper structures beyond `AuthorizationURL`.
  - Current behavior: authorization URL generation always targets `response_type=code`.
  - Missing: helpers for the implicit-flow URL and parsing/representing the callback result shape.
- Add device-code polling orchestration.
  - Current behavior: `StartDeviceAuthorization` and `ExchangeDeviceCode` exist.
  - Missing: a helper that polls according to `interval` and handles documented responses such as `authorization_pending` and similar device-flow states.
- Add configurable token storage helpers if they are meant to ship in-package.
  - Current behavior: `TokenStore` is only an interface; no packaged implementations are provided.
- Add a source/decorator that can force a refresh after a `401` without requiring callers to replace the whole token source.
- Add tests for:
  - refresh-token URL-encoding edge cases
  - refresh failure modes and propagation
  - `401`-triggered refresh integration from the `helix` package
  - hourly validation rollover after more than one hour
  - revoke and validate non-2xx error payload handling

## EventSub webhook runtime

- Enforce the documented webhook timestamp freshness window.
  - Current behavior: signatures are verified, but old messages are not rejected based on timestamp age.
- Add explicit handling for missing required EventSub headers.
- Add a configurable deduplication policy and expiry.
  - Current behavior: `MemoryDeduplicator` keeps message IDs forever for the life of the process.
- Generate typed event structs and decoders for the wider EventSub reference.
  - Currently implemented: `channel.follow@2`, `stream.online@1`.
  - Missing: the rest of the documented EventSub subscription types and versions.
- Add revocation tests that assert typed callback delivery and status handling.

## EventSub WebSocket runtime

- Enforce keepalive timeout behavior.
  - Current behavior: `session_keepalive` messages are accepted, but there is no timeout watchdog that reconnects or fails when keepalives stop arriving.
- Handle documented close codes explicitly and surface typed errors for them.
- Preserve the old connection until the replacement connection is fully welcomed, then close it.
  - Current behavior: reconnect works for the tested happy path, but the code should be hardened and tested against partial failures during the handoff.
- Add support for hard-disconnect recovery and resubscription behavior using the Helix EventSub subscription API.
  - Current behavior: the WebSocket client only consumes messages from a provided URL; it does not recreate subscriptions after a terminal disconnect.
- Add support for revocation dispatch from WebSocket messages in tests.
- Add tests for:
  - keepalive timeout expiry
  - close-code handling
  - reconnect failure paths
  - unexpected message ordering
  - cancellation during reconnect handoff

## Generator and manifests

- Expand `internal/manifest` from the current minimal EventSub decoder manifest into the planned source of truth for:
  - Helix endpoints
  - request params
  - response shapes
  - EventSub subscription conditions
  - EventSub transport/token requirements
  - required scopes
- Harden `internal/cmd/genmanifest` so it validates manifest contents and fails clearly on malformed entries.
- Add tests for the generator output or golden-file checks for generated code.

## Documentation and examples

- Add package documentation for `helix`, `oauth`, and `eventsub`.
- Add runnable examples for:
  - app token client usage
  - user token client usage
  - webhook verification and dispatch
  - WebSocket EventSub consumption
- Add a README with package goals, supported features, and current coverage limits.

## CI and project hygiene

- Add CI configuration to run at least `go test ./...` and `go generate ./eventsub` verification.
- Decide whether generated files should be verified as up to date in CI.
- Add linting/static-analysis configuration if desired for the package.
