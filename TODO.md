# Checkpoint

This file is no longer a stale list of missing pieces from the original plan.
It is the current project status and roadmap as of 2026-04-16.

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
  - `Analytics`
  - `Bits`
  - `HypeTrain`
  - `Entitlements`
  - `Whispers`
  - `Teams`
  - `ContentClassificationLabels`
  - `Extensions`
  - `ChannelPoints`
  - `GuestStar`
  - `Polls`
  - `Predictions`
  - `Goals`
  - `Charity`
  - `Subscriptions`
  - `Clips`
  - `Videos`
  - `EventSub`
- Current Helix endpoint coverage now tracks the live Twitch reference very closely.
- The only obvious remaining Helix headings in the live reference are the deprecated
  stream-tags endpoints, which Twitch documents as legacy behavior.
- Notable currently implemented areas include:
  - `Users`: get users, update user, get authorization by user, get/block/unblock user blocks, get installed extensions, get active extensions, update active extensions
  - `Channels`: get channel info, get channel editors, get followed channels, get channel followers, start commercials, get ad schedules, snooze ads, get/add/remove VIPs, update channel info
  - `Chat`: get/update chat settings, get channel/global/user emotes, get emote sets, get channel/global badges, get chatters, get shared chat sessions, send announcements, send shoutouts, send chat messages, get/update user chat color, delete chat messages
  - `Moderation`: get/add/remove moderators, get moderated channels, get banned users, ban/unban users, check AutoMod status, manage held AutoMod messages, get/update AutoMod settings, warn users, manage suspicious-user states, get/resolve unban requests, get/add/remove blocked terms, get/update shield mode status
  - `Extensions`: get transactions, extension metadata, released extensions, live channels, hosted configuration, required configuration, JWT secrets, extension Bits products, extension chat, and PubSub messages
  - `ChannelPoints`: get/create/update/delete custom rewards, get redemptions, update redemption status
  - `GuestStar`: get/update channel settings, get/create/end sessions, manage invites, assign/update/delete slots, update slot settings
  - `Polls`, `Predictions`, `Goals`, `Charity`, `Subscriptions`, `Analytics`, `Bits`, `HypeTrain`, `Entitlements`, `Whispers`, `Teams`, and `ContentClassificationLabels` all have typed service coverage with tests
  - `Clips`: get clips, create live clips, create clips from VODs, get clip download URLs
  - `Videos`: get videos, delete videos
  - plus the earlier `Streams`, `Search`, `Games`, `Schedule`, `Markers`, `Raids`, and `EventSub` slices already in place

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

The remaining work is now mostly about maintenance, generator expansion, and API
polish rather than broad missing Helix coverage.

### 1. Keep Helix coverage aligned with the live Twitch reference

The SDK is close to reference parity for non-deprecated Helix endpoints.

Remaining work:

- periodically diff the live Twitch reference against the implemented surface
- decide deliberately whether to ignore or expose deprecated endpoints like stream tags
- smooth over naming mismatches where one SDK method currently spans multiple
  closely related reference pages
- continue tightening response models when Twitch documentation or examples are inconsistent

Recommended order:

1. keep tracking live reference additions and removals
2. avoid adding deprecated surface unless there is a strong compatibility reason
3. improve ergonomics only where it clearly helps SDK users

Notes:

- the current Helix roadmap is no longer "add the obvious missing service group"
- it is now more about staying current and improving usability

### 2. Keep EventSub coverage aligned with the live Twitch reference

The EventSub implementation is now much closer to the live documented surface
than it was earlier in the project.

Remaining work:

- periodically diff `internal/manifest/eventsub.json` against the live Twitch
  docs to catch newly added types or versions
- add newly documented EventSub subscription types and versions as they appear
- keep typed API helpers, manifest entries, generated registry, and tests in
  sync for any new EventSub additions
- continue hardening odd transport cases such as conduit and reconnect edge cases

Notes:

- the remaining EventSub work is now more likely to be incremental maintenance,
  newly published versions, or unusual transport/payload edge cases than large
  missing families

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
future feature additions and parity maintenance.

Remaining work:

- add tests alongside each new Helix endpoint
- add tests alongside each new EventSub family
- keep generator verification strict as manifest scope grows
- add golden or shape tests if generated Helix code is introduced
- keep doc-driven edge cases covered where Twitch returns atypical payload shapes
  such as string pagination or alternate field names

## Recommended roadmap

### Near term

1. Keep the SDK synced with live Twitch Helix and EventSub changes.
2. Keep additions small and review-friendly, one coherent family at a time.
3. Avoid adding deprecated Helix endpoints casually.

### Mid term

1. Normalize request/response model patterns across services.
2. Decide the manifest format needed for Helix code generation.
3. Reduce hand-written API shape quirks where the docs are inconsistent.

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
mainly about staying current, improving ergonomics, and moving more of the SDK
to manifest-driven generation.
