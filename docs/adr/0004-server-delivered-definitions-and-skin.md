# Definitions and the skin are server-delivered data

**Status:** Accepted; **both tiers now built**. **Partly superseded: the
bundled-fallback clause was reversed by [contracts#7](0007-components-are-authored-only-in-the-contract.md)
— the client now bundles no definitions at all, and components are authored only
in the contract; the rest stands.** The session pushes the component-definition
library on connect, before the shell, and now the **design tokens** with it: the
set lives in the contract (`tokens/tokens.json`), the Platform serves it, and a
client applies it over a bootstrap copy synced from the same source — a token
changed in the contract reaches a running client with no client build, which was
this record's whole claim and was unbuilt for its first year. The style
*vocabulary* the tokens are spent through (`BoxStyle`/`TextStyle`) is likewise
specified in the schema now, so a second client has something to implement.
Negotiation and versioning of the UI bundle remain untested.
**Date:** 2026-07-21

## Context

The component model ([contracts#2](0002-primitives-and-definitions.md)) draws a
clean line: a small set of **primitives** is native code each client implements,
and everything else is a **definition** — a composition expressed as a primitive
tree, *data, not code*. The skin is likewise **tokens** — DTCG values
([contracts#3](0003-sdui-contract-repository.md)) — not code. The stated payoff is
that the interface changes without shipping a client.

The current implementation does not deliver that payoff. The React runtime
([web#1](https://github.com/mosaic-media/web/blob/main/docs/adr/0001-react-sdui-runtime.md)) **bundles** its definitions
(`definitions.ts`, `definitions.layout.ts` as compiled objects) and its tokens
(CSS) into the `@mosaic-media/sdui-react` package. So a new definition — a new
card layout, a new row — *or* a re-skin requires a package bump, i.e. a client
release.

On web that gap is invisible: redeploying the static Shell is free. For the
planned **Flutter client it is the whole problem** — every definition tweak or
theme change becomes an app-store review, which is exactly what SDUI exists to
avoid. [web#2](https://github.com/mosaic-media/web/blob/main/docs/adr/0002-server-owned-navigation.md) removed *navigation* from the
client; the remaining things still baked into the client are **definitions and
tokens**. This ADR moves them out.

## Decision

**The client bundles only the native vocabulary. Definitions and the active token
set (the skin) are server-delivered data — fetched at boot, cached, versioned,
and negotiated against the client's native vocabulary. Nothing composed from the
vocabulary is baked into the client.**

Three tiers, with a hard line between the first and the rest:

- **Native vocabulary (baked, versioned by the client release).** The
  primitives, the action kinds the dispatcher interprets, and the style/render
  capabilities the primitives can draw. This is the irreducible native contract
  ([contracts#2](0002-primitives-and-definitions.md)); growing it is the *only*
  thing that needs a client release.
- **UI library (server-delivered data, cached).** The definition library and the
  active token set — semi-static, changes with product and design but not per
  interaction.
- **Live surface (pushed per interaction).** Screens, content and navigation
  ([platform#19](https://github.com/mosaic-media/platform/blob/main/docs/adr/0019-sdui-emit-side.md), [platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md),
  [web#2](https://github.com/mosaic-media/web/blob/main/docs/adr/0002-server-owned-navigation.md)).

The UI-library tier works like this:

- **The Platform serves the UI bundle** — the standard definition library
  ([contracts#3](0003-sdui-contract-repository.md)) merged with any definitions the
  enabled modules contribute ([sdk#4](https://github.com/mosaic-media/sdk/blob/main/docs/adr/0004-module-contributed-settings-ui.md)
  established that modules ship SDUI as data; this extends it from settings
  screens to reusable definitions), plus the active token set — as a single
  **content-addressed, versioned** payload.
- **The client fetches it at boot and caches it by version.** It re-fetches only
  when the version changes; the live-session handshake ([platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md))
  advertises the current bundle version, so a change propagates without a client
  release and without re-downloading on every connect.
- **Delivery is negotiated against the client's native-vocabulary version.** The
  client advertises which primitives / action kinds / capabilities it implements
  (its runtime-capability version). The Platform serves the newest definition/
  token set compatible with it. So a new definition that uses only existing
  primitives reaches already-deployed clients as data; a definition that needs a
  *new* primitive is gated to clients that have shipped it. This is the seam that
  keeps "server can add UI" honest without breaking old clients.
- **Unknown references degrade, never crash.** A definition (or a stray screen
  node) referencing a primitive the client lacks renders a typed fallback, not an
  error — graceful degradation becomes load-bearing, since the server can now
  serve definitions a given client may not fully support.
- **Definitions stay declarative data, never executable code.** A definition is a
  template (`$bind`/`$each`/`$if`/`$match`/`Outlet`), interpreted by the renderer.
  The client never evaluates server-supplied code — this is a security boundary,
  not only an architectural one.

The result: a Flutter (or web) client is releasable **only** to grow the native
vocabulary. New card layouts, new rows, a full re-skin, a seasonal theme, a
per-deployment palette — all server data, no release.

## Alternatives considered

**Keep bundling definitions and tokens (status quo).** *Rejected.* Every
definition or skin change is a client release — a store review on Flutter — which
negates SDUI's premise the moment a native client exists.

**Push the entire library over the live session on every connect.** *Rejected.*
The library is large and static relative to the per-interaction surface; pushing
it each connect wastes bandwidth and startup time. Content-addressed fetch with
caching and a version advertised on connect is the efficient shape, and it lets
an unchanged library cost nothing on reconnect.

**Ship definitions as code the client dynamically loads and executes.**
*Rejected.* Definitions are declarative templates; they need no code execution,
and executing server-supplied code would be a serious security regression and is
impossible under app-store rules for native clients anyway. Data interpreted by a
fixed renderer is the whole point.

**Server-side render to pixels; the client shows images.** *Rejected.* It throws
away native performance, real interaction, accessibility and offline behaviour —
it is not server-*driven* UI, it is remote screenshots. The primitive/definition
split exists precisely to keep rendering native while making *composition* data.

## Consequences

- **The client is releasable only to grow the native vocabulary** — a new
  primitive, action kind, or render capability. Everything else (definitions,
  skin, screens, content, navigation) flows from the server. This is the
  guarantee that makes SDUI real for Flutter, not just web, and it is what
  [web#2](https://github.com/mosaic-media/web/blob/main/docs/adr/0002-server-owned-navigation.md) and this ADR together deliver.
- **Skinning becomes a server operation** — re-theme, per-deployment palettes,
  seasonal skins, even experiments, without a client release. Token *values* are
  data; only a new token *capability* (a render effect a primitive can't draw) is
  native code.
- **New machinery:** a served, versioned UI-bundle (definitions + tokens), client
  fetch-and-cache, and capability negotiation on connect. The runtime stops
  bundling its definitions and CSS and consumes them as data instead.
- **The client's native-vocabulary version becomes a first-class contract** —
  advertised and negotiated — so the server can roll UI forward to old clients
  safely and gate primitive-dependent UI to updated ones.
- **Graceful degradation is now essential,** not optional: unknown primitive →
  typed fallback.
- **Modules ship reusable UI as data,** extending
  [sdk#4](https://github.com/mosaic-media/sdk/blob/main/docs/adr/0004-module-contributed-settings-ui.md) — a module's definitions
  merge into the served library, gated by the same capability negotiation.
- **Honest limits:** the native vocabulary still versions with the client, and a
  definition that needs a new primitive still waits on the release that adds it —
  the gain is that everything *within* the vocabulary is free. And the vocabulary
  must be designed to be *expressive and stable*: if primitives are too granular
  or too many, "new definition without release" narrows. Keeping the primitive
  set small and composable ([contracts#2](0002-primitives-and-definitions.md)) is
  what makes this ADR pay off.

## Implementation implications

Proposed; a slice of its own, sequenced with [web#2](https://github.com/mosaic-media/web/blob/main/docs/adr/0002-server-owned-navigation.md)
so the client contract is settled before the Flutter client is built against it.

- **Platform:** compose and serve the UI bundle for this deployment — the
  `mosaic-sdui` standard definitions + enabled modules' definitions + the active
  token set — content-addressed and versioned; advertise the version on the live
  handshake; accept the client's runtime-capability version and serve a
  compatible set.
- **Runtime (`mosaic-sdui-react`, and the future Flutter renderer):** fetch the
  bundle at boot, cache by version, register definitions and apply tokens from
  data instead of from the package; advertise the native-vocabulary version;
  render a typed fallback for unknown primitives. The package ships the
  primitives, the renderer, the token *application* mechanism and the negotiation
  — not the definition library or specific token values.
- **`mosaic-sdui`:** already the canonical home of the definitions (JSON) and
  DTCG tokens; this makes them a *delivered artefact* rather than a compile-time
  dependency of the runtime.
