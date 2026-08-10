# Cross-client transport: protobuf and two lanes

**Status:** Accepted — built on both sides. **Partly superseded:** the
"split by audience" answer to the GraphQL question below was reversed by
[platform#37](https://github.com/mosaic-media/platform/blob/main/docs/adr/0037-one-client-transport.md), which retired GraphQL entirely; the
rest of this record stands.
**Date:** 2026-07-21

## Context

[platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md) chose one persistent, bidirectional
**WebSocket** for the live session, justified by two real needs — streaming
input (search-as-you-type) and server push — and it explicitly assumed the
transport would stay technology-agnostic so "a Flutter client speaks the same
session." That assumption is now a firm requirement, and it changes the
calculus enough to revisit the decision.

**The client set is no longer web-first.** Mosaic is designed to be rendered by
several first-class clients: the React Shell today, and Flutter, Jetpack
Compose, and native iOS (Swift) clients as peers. The SDUI contract already
exists to make this possible ([contracts#1](0001-server-driven-ui-and-the-shell.md),
[contracts#3](0003-sdui-contract-repository.md)). The transport underneath it must
be judged by the same standard: *most efficient while working uniformly across
all four clients*, not merely "works in a browser."

Three facts reframe the choice:

- **One-socket-carries-everything is the exception, not the norm.** The industry
  pattern for server-driven media UIs is **two lanes**: the UI is *pulled* over
  request/response (Netflix's Falcor, Apple's TVMLKit documents, YouTube's
  continuation responses), and a *narrow push channel* is reserved for the
  genuinely server-initiated (Spotify Connect device sync, Twitch chat/PubSub).
  [platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md)'s single bidirectional socket carrying navigation, input, actions and
  push is the Twitch-shaped exception — justified there by live interactivity
  being the product, which for a media *library* it is not.
- **SSE was the wrong second guess.** [platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md) considered "HTTP intents + SSE for
  push" and set it aside because SSE is one-way. That reasoning holds, but the
  deeper problem surfaces only with native clients: SSE is a web technology.
  `EventSource` is native in the browser (with caveats — no custom headers, an
  HTTP/1.1 connection-count cost) but **absent on Flutter and iOS**, where it
  means hand-writing `text/event-stream` parsing per platform. SSE fails the
  cross-client bar.
- **The contract is generated, and native codegen prefers protobuf.** The SDUI
  and platform contracts are schema-first, generated into per-language bindings
  ([platform#12](https://github.com/mosaic-media/platform/blob/main/docs/adr/0012-published-contract-surface.md),
  [contracts#3](0003-sdui-contract-repository.md)). Extending that to Swift, Kotlin
  and Dart is far cleaner from protobuf than from the current JSON-Schema, and
  binary framing is materially cheaper on the wire — which mobile clients feel as
  data and battery.

The question this ADR answers: **what transport is the most efficient one that
is uniform across web, Flutter, Jetpack Compose and native iOS — and how does it
carry unsolicited server push and region-level (not whole-screen) updates?**

## Decision

**Adopt a typed, two-lane RPC transport — Connect/gRPC over protobuf — for the
first-party client session. Client intents travel as unary calls; server push
travels as one server-streaming call. This supersedes [platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md)'s bespoke
WebSocket as the client transport and folds [supervisor#4](https://github.com/mosaic-media/supervisor/blob/main/docs/adr/0004-supervisor-driven-live-handover.md)'s handover into stream
resume.**

The two lanes map onto exactly the two gRPC shapes that are supported uniformly
on *every* target client, browser included:

- **Lane 1 — intents (unary).** `Navigate`, `Invoke`, `SubmitInput` (a
  search-as-you-type keystroke), `Attach`. Request/response. Each resolves to
  the same application command or query the HTTP path would; the command boundary
  ([platform#12](https://github.com/mosaic-media/platform/blob/main/docs/adr/0012-published-contract-surface.md)) and the caller model
  ([platform#13](https://github.com/mosaic-media/platform/blob/main/docs/adr/0013-how-a-capability-acts.md)) are unchanged — this is a
  transport, not a second application layer.
- **Lane 2 — push (server-streaming).** One long-lived `Subscribe` stream per
  session over which the server pushes region updates, toasts, shell mutations
  and unsolicited domain events. This is the channel HTTP request/response cannot
  provide and the reason a live channel exists at all.

Deliberately, the model needs **no bidirectional streaming**. The one gRPC
capability browsers lack (client- and bidi-streaming over grpc-web/Connect) is
the one this design does not use: unary + server-streaming are supported on web
(Connect / grpc-web), Flutter (`grpc`), Compose (grpc-kotlin) and Swift
(grpc-swift) alike. Over HTTP/2 both lanes **multiplex onto a single
connection**, so the two-lane logical split costs one physical connection, not
two — mobile-friendly, and without the hand-rolled framing a raw socket needs.

Supporting decisions:

- **The region-update protocol is contract-owned and transport-agnostic.** The
  `replace / append / prepend / remove / patch` operations on a named region
  ([platform#19](https://github.com/mosaic-media/platform/blob/main/docs/adr/0019-sdui-emit-side.md), [platform#21](https://github.com/mosaic-media/platform/blob/main/docs/adr/0021-server-owned-app-shell.md))
  live in the shared SDUI contract, not in this transport. A Compose composable,
  a SwiftUI view and a React component apply the same ops. This keeps the wire a
  swappable detail: the region vocabulary does not change if the transport does.
- **Server push is first-class and unsolicited.** A finished import, a config
  change, a library edit from another device, or a Supervisor handover reach the
  client on the `Subscribe` stream without a prior request — the requirement
  request/response cannot meet.
- **Search-as-you-type still coalesces server-side.** `SubmitInput` calls land
  and the Platform debounces them against session state before sourcing, so a
  fast typist does not fan out a request per keystroke to the upstream addons.
  The backstop of [platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md) is kept; only its mechanism moves (see Consequences).
- **The contract framing consolidates on protobuf.** Running GraphQL for the pull
  lane and protobuf for the push lane would mean two client stacks per platform —
  the worst outcome. The first-party client path is protobuf on both lanes.
  GraphQL is retained only where a flexible public query surface earns it
  (external/third-party integrations and tooling), not for the first-party
  clients (see "The GraphQL question").

### The UINode encoding question (sub-decision, called out deliberately)

The transport carries `UINode` subtrees in each region update. Two encodings are
possible, and this ADR names the choice rather than smuggling it:

- **(a) Keep JSON-Schema `UINode`, carried as bytes inside the protobuf
  envelope.** Smallest change; mosaic-sdui stays the single source of truth. But
  the largest, most frequent payload — the UI tree itself — is still JSON that
  native clients parse by hand, so the codegen and wire wins are partial.
- **(b) Generate `UINode` as protobuf too.** Full binary efficiency and native
  codegen for the biggest payload. But it is a real change to the SDUI contract
  repository ([contracts#3](0003-sdui-contract-repository.md)) and its generation
  pipeline.

**Recommendation:** start at (a) so the transport can land without reopening the
SDUI contract, and treat (b) as a follow-on once a second native client is real
and the codegen/​wire pressure is measured rather than assumed. The `.proto`
below is written so the envelope is identical either way.

## Alternatives considered

**Keep the [platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md) WebSocket, upgrade the payload to protobuf.** *Considered,
and the runner-up.* WebSocket is the only transport that is natively bidi on all
four clients, it keeps the clean single-ordered-reader concurrency model, and it
tolerates the dumbest reverse proxy (HTTP/1.1 upgrade). It is the right pick for
"smallest safe step." It is rejected here as the *foundation* because it
entrenches the non-standard socket-carries-all shape, hand-rolls request/reply
correlation, framing, resume and versioning that Connect gives for free, and
does not, by itself, deliver the typed cross-client codegen that is the point.

**HTTP intents + SSE for push.** *Rejected.* One-way (needs a second up-channel
anyway) and, decisively, not natively available on Flutter or iOS — it fails the
cross-client bar this ADR is built around.

**GraphQL subscriptions (graphql-ws).** *Rejected.* It reintroduces the WS
coupling with none of the codegen benefit, the live protocol is region-update
operations rather than subscription-result semantics, and there are no
subscriptions in the schema today to build on.

**WebTransport (HTTP/3 / QUIC).** *Rejected for now.* True bidi with datagrams
and no head-of-line blocking is the eventual successor to WebSocket, but proxy,
CDN and client-library support is too thin in 2026 for a self-hosted product
behind arbitrary infrastructure. Revisit.

**HTTP long-poll.** *Rejected.* Emulated push with reconnect churn; a fallback,
never a target.

## Consequences

The client session becomes a typed, two-lane RPC surface generated from one
`.proto` into every client language — the cross-client, wire-efficient
foundation the four-client goal needs. Honest costs, all acceptable but real:

1. **The concurrency surface grows, and needs the right pattern.** [platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md)'s
   single ordered read loop is replaced by concurrent unary calls running
   alongside the push stream, all touching per-session state — the same
   read/write hazard that produced the `current`-route race, now with more
   writers. The discipline is a **per-session outbound mailbox**: unary handlers
   never touch the stream directly; they enqueue onto a per-session channel that
   a single sender goroutine drains to `stream.Send` (gRPC `Send` is not
   goroutine-safe). Session state is owned by that one goroutine or guarded
   explicitly. This is well-trodden, arguably cleaner than the timer race it
   replaces — but it is real engineering, not free.
2. **Input coalescing moves from inline to session state.** `SubmitInput` calls
   arrive concurrently rather than in one ordered stream; the debounce becomes
   session-keyed state whose result is delivered on the push lane. The backstop
   is preserved; the mechanism is more moving parts.
3. **HTTP/2 is strongly preferred.** The one-connection multiplexing of the two
   lanes depends on it. Connect degrades to HTTP/1.1 (unary as normal requests,
   the stream as a chunked response), which works but may split the lanes across
   connections behind a proxy that will not carry HTTP/2. The deployment surface
   ([supervisor#1](https://github.com/mosaic-media/supervisor/blob/main/docs/adr/0001-supervisor-as-host-manager.md)) should account for this.
4. **Stream resume replaces the [supervisor#4](https://github.com/mosaic-media/supervisor/blob/main/docs/adr/0004-supervisor-driven-live-handover.md) handover.** The `Subscribe` stream
   carries a monotonic sequence; on reconnect the client presents the last
   sequence it saw and the server replays or rebuilds. This subsumes [supervisor#4](https://github.com/mosaic-media/supervisor/blob/main/docs/adr/0004-supervisor-driven-live-handover.md)'s
   bespoke going-away/reconnect dance and gives resumability across a rolling
   upgrade — but it is new resume/replay logic to build and bound.
5. **A second contract framing enters the codebase.** protobuf/Connect sits
   alongside the JSON-Schema SDUI contract and any retained GraphQL. This is
   consistent with the schema-first ethos ([platform#12](https://github.com/mosaic-media/platform/blob/main/docs/adr/0012-published-contract-surface.md))
   but is another generator, another set of generated bindings, and a versioning
   surface of its own.
6. **This supersedes an Accepted ADR.** [platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md)'s
   WebSocket transport is replaced and [supervisor#4](https://github.com/mosaic-media/supervisor/blob/main/docs/adr/0004-supervisor-driven-live-handover.md)
   is revised on acceptance. The *goals* of [platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md) — streaming input, server
   push, region updates, a technology-agnostic client — are unchanged and fully
   met; only the wire changes.

### The GraphQL question

> **Superseded by [platform#37](https://github.com/mosaic-media/platform/blob/main/docs/adr/0037-one-client-transport.md).** The recommendation
> below — split by audience — was taken, and did not hold: the retained surface
> attracted no caller, the Shell used one operation of it, and three commands
> ended up implemented twice. [platform#37](https://github.com/mosaic-media/platform/blob/main/docs/adr/0037-one-client-transport.md) takes the *consolidate* option this
> record deferred as "the largest migration"; it turned out to be small, because
> almost nothing was using what had been kept. The reasoning below is preserved
> as the record of why the split looked right at the time.

The Platform serves a GraphQL HTTP API today, and [platform#22](https://github.com/mosaic-media/platform/blob/main/docs/adr/0022-live-session-websocket.md)'s socket reuses its
mutation resolvers. Three coexistence options:

- **Consolidate.** Move the first-party queries and mutations to Connect unary
  and retire the GraphQL transport. One contract surface, cleanest long term;
  the largest migration.
- **Split by audience (recommended).** Connect two-lane is the *first-party
  client* surface; GraphQL is kept only where a flexible, ad-hoc query surface
  genuinely earns it — external integrations, tooling, exploration. A clear
  boundary, and it stops GraphQL from being on the hot client path without a
  disruptive removal.
- **GraphQL pull + protobuf push.** *Rejected* — two client stacks per platform,
  the worst of both.

## Implementation implications

A new Connect/gRPC transport surface on the Platform, generated from a session
`.proto`, alongside (then gradually instead of) the GraphQL HTTP API. Illustrative
shape:

```proto
syntax = "proto3";
package mosaic.session.v1;

service SessionService {
  // Lane 1 — intents (unary). Each resolves to the same application
  // command/query the HTTP path would (the command boundary, [platform#12](https://github.com/mosaic-media/platform/blob/main/docs/adr/0012-published-contract-surface.md)).
  rpc Attach(AttachRequest) returns (Ack);          // (re)bind this call to a session
  rpc Navigate(NavigateRequest) returns (Ack);
  rpc Invoke(InvokeRequest) returns (Ack);          // an action → a mutation
  rpc SubmitInput(InputRequest) returns (Ack);      // one search-as-you-type keystroke

  // Lane 2 — push (server-streaming). One long-lived stream per session; the
  // server pushes region updates, toasts, shell mutations and unsolicited
  // events. resume_cursor replays what a reconnecting client missed
  // (superseding [supervisor#4](https://github.com/mosaic-media/supervisor/blob/main/docs/adr/0004-supervisor-driven-live-handover.md)'s handover).
  rpc Subscribe(SubscribeRequest) returns (stream ServerMessage);
}

message SubscribeRequest {
  string session = 1;         // opaque session ref ([platform#13](https://github.com/mosaic-media/platform/blob/main/docs/adr/0013-how-a-capability-acts.md)); auth is on this call
  uint64 resume_cursor = 2;   // last ServerMessage.seq seen; 0 on a fresh connect
}

message ServerMessage {
  uint64 seq = 1;             // monotonic per session; the resume cursor
  oneof body {
    RegionUpdate region = 2;
    ShellUpdate  shell  = 3;
    Toast        toast  = 4;
    Event        event  = 5;  // unsolicited: import finished, config changed, …
  }
}

message RegionUpdate {
  string region = 1;          // a named slot ([platform#19](https://github.com/mosaic-media/platform/blob/main/docs/adr/0019-sdui-emit-side.md) / 0031)
  Op     op     = 2;
  bytes  ui_node = 3;         // a UINode subtree in the SDUI contract encoding
                              // (JSON today; see "The UINode encoding question")
  enum Op { REPLACE = 0; APPEND = 1; PREPEND = 2; REMOVE = 3; PATCH = 4; }
}
```

The Platform side needs: a session manager holding per-connection live state
behind the **outbound-mailbox** pattern above; intent → command routing that
reuses the application services unchanged; the server-stream sender; and stream
resume keyed on `seq`. Clients generate typed stubs per platform (Connect-Web /
`grpc` / grpc-kotlin / grpc-swift) and apply region updates to their native
render trees. Sequence after the region-update op-set is settled in the SDUI
contract, so the wire and the vocabulary are not in flight at once. Start with
UINode encoding option (a); measure before committing to (b).
