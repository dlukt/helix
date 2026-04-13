# Checkpoint

This file is no longer a stale list of missing pieces from the original plan.
It is the current project status and roadmap as of 2026-04-13.

## Current state

The repository now has a solid first implementation of a reusable Go package for
the Twitch Helix API, OAuth flows, and EventSub runtimes.

### Helix core

- `helix.Client` supports:
  - typed response metadata
  - request ID propagation
  - parsed rate-limit headers
  - richer API error decoding
  - `401` retry with refresh-capable token sources
  - low-level request execution through `RawRequest`
- Current service groups on `*helix.Client`:
  - `Users`
  - `Channels`
  - `Streams`
  - `Chat`
  - `Moderation`
  - `Raids`
  - `Schedule`
  - `Markers`
  - `Games`
  - `Search`
  - `Clips`
  - `Videos`
  - `EventSub`

### OAuth

- Package docs and examples are in place.
- Current coverage includes:
  - authorization-code flow helpers
  - client-credentials flow helpers
  - implicit-flow URL generation and callback parsing
  - device authorization start and polling
  - token refresh
  - token validate and revoke
  - memory-backed token storage helpers
  - force-refresh token-source decorator for `401` recovery

### EventSub webhook runtime

- Package docs and examples are in place.
- Current coverage includes:
  - signature verification
  - required-header validation
  - timestamp freshness validation
  - deduplication with configurable TTL
  - typed notification decoding through a generated registry
  - revocation dispatch

### EventSub WebSocket runtime

- Current coverage includes:
  - welcome / notification / revocation handling
  - keepalive timeout behavior
  - reconnect flow handling
  - typed close errors for documented close codes
  - optional hard-disconnect recovery via Helix subscription recreation

### Generator and manifests

- `internal/manifest/eventsub.json` is the source of truth for the generated
  EventSub decoder registry.
- `internal/cmd/genmanifest` validates manifest entries and duplicate keys.
- CI verifies:
  - `go test ./...`
  - `go generate ./eventsub`
  - generated files are up to date

### Documentation and project hygiene

- Root `README.md` exists.
- Package docs exist for:
  - `helix`
  - `oauth`
  - `eventsub`
- Runnable examples exist.
- GitHub Actions test workflow exists.

## Current EventSub coverage

The generated EventSub registry currently covers 82 typed subscription types and
versions from `internal/manifest/eventsub.json`.

This includes:

- AutoMod:
  - `automod.message.hold@1`
  - `automod.message.hold@2`
  - `automod.message.update@1`
  - `automod.message.update@2`
  - `automod.settings.update@1`
  - `automod.terms.update@1`
- Channel lifecycle and metadata:
  - `channel.follow@2`
  - `channel.update@2`
  - `stream.online@1`
  - `stream.offline@1`
- Ads and bits:
  - `channel.ad_break.begin@1`
  - `channel.bits.use@1`
  - `extension.bits_transaction.create@1`
  - `drop.entitlement.grant@1`
- Chat and moderation-adjacent chat events:
  - `channel.chat.clear@1`
  - `channel.chat.clear_user_messages@1`
  - `channel.chat.message@1`
  - `channel.chat.message_delete@1`
  - `channel.chat.notification@1`
  - `channel.chat_settings.update@1`
  - `channel.chat.user_message_hold@1`
  - `channel.chat.user_message_update@1`
- Subscriptions:
  - `channel.subscribe@1`
  - `channel.subscription.gift@1`
  - `channel.subscription.end@1`
  - `channel.subscription.message@1`
- Goals, polls, and predictions:
  - `channel.goal.begin@1`
  - `channel.goal.progress@1`
  - `channel.goal.end@1`
  - `channel.poll.begin@1`
  - `channel.poll.progress@1`
  - `channel.poll.end@1`
  - `channel.prediction.begin@1`
  - `channel.prediction.progress@1`
  - `channel.prediction.lock@1`
  - `channel.prediction.end@1`
- Charity and hype train:
  - `channel.charity_campaign.donate@1`
  - `channel.charity_campaign.start@1`
  - `channel.charity_campaign.progress@1`
  - `channel.charity_campaign.stop@1`
  - `channel.hype_train.begin@2`
  - `channel.hype_train.progress@2`
  - `channel.hype_train.end@2`
- Shared chat and safety:
  - `channel.shared_chat.begin@1`
  - `channel.shared_chat.update@1`
  - `channel.shared_chat.end@1`
  - `channel.shield_mode.begin@1`
  - `channel.shield_mode.end@1`
  - `channel.warning.acknowledge@1`
  - `channel.warning.send@1`
- Social, guest star, and moderation actions:
  - `channel.shoutout.create@1`
  - `channel.shoutout.receive@1`
  - `channel.guest_star_session.begin@beta`
  - `channel.guest_star_session.end@beta`
  - `channel.guest_star_guest.update@beta`
  - `channel.guest_star_settings.update@beta`
  - `channel.unban_request.create@1`
  - `channel.unban_request.resolve@1`
  - `channel.suspicious_user.update@1`
  - `channel.suspicious_user.message@1`
  - `channel.cheer@1`
  - `channel.ban@1`
  - `channel.unban@1`
  - `channel.moderate@1`
  - `channel.moderate@2`
  - `channel.moderator.add@1`
  - `channel.moderator.remove@1`
  - `channel.vip.add@1`
  - `channel.vip.remove@1`
  - `channel.raid@1`
- Channel points:
  - `channel.channel_points_custom_reward.add@1`
  - `channel.channel_points_custom_reward.update@1`
  - `channel.channel_points_custom_reward.remove@1`
  - `channel.channel_points_custom_reward_redemption.add@1`
  - `channel.channel_points_custom_reward_redemption.update@1`
  - `channel.channel_points_automatic_reward_redemption.add@1`
  - `channel.channel_points_automatic_reward_redemption.add@2`
- Infrastructure:
  - `conduit.shard.disabled@1`
- User-scoped events:
  - `user.authorization.grant@1`
  - `user.authorization.revoke@1`
  - `user.update@1`
  - `user.whisper.message@1`

## What is still ahead

The remaining work is now mostly about breadth and generation, not basic
transport or runtime foundations.

### 1. Finish broader EventSub reference coverage

This is the clearest next track.

Remaining work:

- add the rest of the documented EventSub subscription types and versions
- prioritize larger families that are not yet modeled
- keep the typed API helpers, manifest entries, generated registry, and tests in
  sync for each added family

Notes:

- the remaining EventSub gaps are now narrower and are more likely to be
  specialized payloads, newly documented versions, or transport-specific edge
  cases than the major chat/moderation families
- the unusual payload shapes are now mostly the exception cases rather than the
  norm, so it is worth periodically diffing the manifest against the live
  Twitch docs before choosing the next slice

### 2. Expand Helix API endpoint coverage

The client already has a useful set of service groups, but the endpoint surface
inside those services is still incomplete.

Remaining work:

- add more endpoints within existing services
- add missing typed request parameter structs where current coverage is thin
- add more representative response models
- add missing service groups if needed after existing groups are filled out

Recommended order:

1. deepen coverage inside already-created service groups
2. only then add entirely new service groups

### 3. Expand manifests and code generation beyond EventSub registry generation

This is the biggest structural roadmap item still open.

Remaining work:

- extend manifest data to describe Helix endpoints
- describe request params and query encoding in manifest form
- describe response shapes in manifest form
- describe EventSub condition schemas in manifest form
- optionally describe required scopes and transport constraints in manifest form
- generate more typed code from those manifests instead of continuing to add all
  API layers manually

This is the main step required to move from a hand-written implementation with a
generated EventSub registry to a fuller manifest-driven SDK.

### 4. Decide the long-term shape of typed EventSub conditions

We already have typed condition structs in `eventsub_api.go`, but they are still
maintained by hand.

Remaining work:

- either keep condition structs hand-written and continue incrementally
- or move condition definitions into manifest data and generate them

If generator expansion proceeds, generated conditions are the cleaner direction.

### 5. Continue hardening test coverage around new surface area

The foundation already has strong tests. The main remaining test work is tied to
future feature additions.

Remaining work:

- add tests alongside each new Helix endpoint
- add tests alongside each new EventSub family
- keep generator verification strict as manifest scope grows
- add golden or shape tests if generated Helix code is introduced

## Recommended roadmap

### Near term

1. Continue filling in the remaining EventSub families.
2. Keep additions small and review-friendly, one coherent family at a time.
3. Avoid starting Helix generator expansion until EventSub parity is in a
   clearly good place.

### Mid term

1. Broaden endpoint coverage within the existing Helix service groups.
2. Normalize request/response model patterns across services.
3. Decide the manifest format needed for Helix code generation.

### Long term

1. Build a richer manifest schema that can describe Helix endpoints and
   EventSub conditions.
2. Generate larger parts of the SDK from manifests.
3. Reduce hand-written duplication in:
   - EventSub typed create helpers
   - EventSub condition structs
   - Helix request/response boilerplate

## Non-goals for the checkpoint

These are not blockers for the current checkpoint:

- replacing all hand-written code with generated code immediately
- full Helix API parity before the manifest format is ready
- full EventSub parity in one giant change

The current implementation is already a usable package. The roadmap from here is
mainly about completing coverage and moving more of the SDK to manifest-driven
generation.
