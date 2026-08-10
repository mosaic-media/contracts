# The SDUI and session contracts are protobuf in one workspace

**Status:** Accepted (built) — the `proto` workspace, generated Go and
TypeScript are built; encoding option (b) landed, with the typed
`mosaic.sdui.v1.UINode` riding the session envelope directly and no JSON step;
and the `sdui` → `contracts` rename landed on 2026-07-24 at `contracts`
`v0.12.0`, with every Go importer moved to the new module path. The npm name
`@mosaic-media/sdui` is unchanged, as this record specifies.
**Date:** 2026-07-21

## Context

[contracts#3](0003-sdui-contract-repository.md) made the SDUI contract JSON Schema,
and recorded the protobuf alternative as an open question to revisit *if the
vocabulary became governed*. That is **not** the trigger firing here — the
vocabulary stays open. Two other forces do.

- **The transport is now protobuf.** [contracts#5](0005-cross-client-transport-two-lane-rpc.md)
  replaced the WebSocket with a Connect/protobuf `SessionService`. [contracts#3](0003-sdui-contract-repository.md)'s
  first reason for JSON — "the payload rides GraphQL as JSON" — no longer holds:
  the payload now rides a protobuf envelope, and a JSON `UINode` carried as
  opaque bytes inside it is the mismatch, not the fit.
- **Native clients codegen from protobuf, not JSON Schema.** [contracts#5](0005-cross-client-transport-two-lane-rpc.md) exists to
  serve web + Flutter + Compose + Swift from one contract. Typed Swift/Kotlin/Dart
  clients generate cleanly from `.proto`; from JSON Schema they do not.

[contracts#5](0005-cross-client-transport-two-lane-rpc.md) named this the "UINode encoding" sub-decision and shipped **option (a)**
— `ui_node` as JSON bytes — as the smaller step. In practice the Platform push
went further wrong than option (a) intended: the session `.proto` and its
generated Go landed **inside** `platform`, under `internal/gen/` and licensed
**AGPL**. `internal/` means no other module can import it and AGPL means no client
should — so the contract as pushed cannot be consumed by any client at all,
defeating the four-client purpose. The contract must be a **published, permissive**
artifact, and — chosen now — a **protobuf** one, so the wire is typed end to end
(**option (b)**).

[contracts#3](0003-sdui-contract-repository.md)'s other two reasons for JSON resolve rather than block:

- *"props are an untyped bag"* — they stay that way. protobuf types the **tree
  skeleton, actions and region operations**; the open props and attributes
  ([platform#11](https://github.com/mosaic-media/platform/blob/main/docs/adr/0011-open-and-closed-vocabularies.md)) ride `google.protobuf.Struct`.
  The Struct collapse is real but confined to the leaves that are open *by design*.
- *"definitions and tokens are JSON data"* — tokens **stay** DTCG JSON (below);
  definitions stay data, carried via canonical protobuf-JSON.

## Decision

**The SDUI contract migrates from JSON Schema to protobuf. The session transport
contract joins it as a second module in one buf workspace. The `sdui` repository
is renamed `contracts` and publishes two packages; both are Apache-2.0.**

- **One repo, `contracts`** (renamed from `sdui`) — a **buf workspace** with two
  proto modules:
  - `proto/mosaic/sdui/v1/*` — `UINode`, `Action`, `ComponentDefinition`, the
    region operations ([platform#19](https://github.com/mosaic-media/platform/blob/main/docs/adr/0019-sdui-emit-side.md) / [platform#21](https://github.com/mosaic-media/platform/blob/main/docs/adr/0021-server-owned-app-shell.md)).
  - `proto/mosaic/session/v1/*` — the `SessionService` two lanes ([contracts#5](0005-cross-client-transport-two-lane-rpc.md)),
    **importing** the sdui module: `RegionUpdate.ui_node` is a typed
    `mosaic.sdui.v1.UINode`, not bytes (**option (b)**).
- **Two published packages**, independently consumable: `@mosaic-media/sdui` and
  `@mosaic-media/transport` (npm), with Go bindings under
  `github.com/mosaic-media/contracts/gen/…` (a **published** path, not `internal/`).
  Generated for Go and TypeScript now; Swift/Kotlin/Dart when those clients land.
  The one-way dependency (transport → sdui) means an sdui-only consumer pulls only
  the sdui package.
- **Both Apache-2.0**, matching the SDK and the SDUI contract before it
  ([architecture#1](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0001-licensing.md)). The session contract **relicenses** from the
  AGPL it was mistakenly pushed under.
- **Carve-outs — what does *not* become protobuf:**
  - **Design tokens stay DTCG JSON.** They are a design-tool interchange format
    (Figma, Style Dictionary); protobuf buys nothing and fights the ecosystem.
    Delivered as data ([contracts#4](0004-server-delivered-definitions-and-skin.md)).
  - **Open props / attributes are `google.protobuf.Struct`.** The vocabulary stays
    open ([platform#11](https://github.com/mosaic-media/platform/blob/main/docs/adr/0011-open-and-closed-vocabularies.md)); this is **not** the
    governance fork [contracts#3](0003-sdui-contract-repository.md) named — no per-component typed messages.
  - **The definition library stays data**, carried and authored via canonical
    protobuf-JSON (`protojson`); clients still register definitions, emit types,
    and ship no UI code ([contracts#2](0002-primitives-and-definitions.md)).
- **[contracts#5](0005-cross-client-transport-two-lane-rpc.md) resolves to option
  (b)**, and its session contract moves out of `platform` into this workspace.
  Platform keeps the transport *implementation* (`internal/transport/session/*`)
  and imports the published bindings.

## Alternatives considered

**Keep JSON Schema ([contracts#3](0003-sdui-contract-repository.md) status quo).** *Rejected now.* Its two live reasons
have flipped: the transport is protobuf, and native clients are real. The third
(governed vocabulary) was never the point and still isn't — props stay open.

**Option (a): `UINode` as JSON bytes in the protobuf envelope.** *Rejected.* It
leaves the largest, most frequent payload untyped — hand-parsed JSON on every
native client — and half-types the wire. As pushed it also stranded the contract
in `platform/internal/gen` (un-importable, AGPL). Option (b) types the whole wire.

**Two repos (`sdui`, `transport`) with buf cross-repo deps.** *Rejected.* The
schemas are coupled — `transport.proto` imports `sdui.proto` — so two repos means
BSR/git buf-dependency management for a dependency that genuinely exists. One
workspace resolves the import locally with one breaking-change gate, and matches
[web#3](https://github.com/mosaic-media/web/blob/main/docs/adr/0003-frontend-workspace.md)'s rule: split a repo only to enforce a
boundary, and there is none between two coupled Apache contracts.

**Govern the vocabulary (every component a typed message).** The [contracts#3](0003-sdui-contract-repository.md) fork.
*Not taken* — this migration is about cross-client codegen and a typed transport,
not closing the open vocabulary.

## Consequences

- **A breaking change to a published contract with several consumers** — the
  Platform emit-side, the `web` runtime (`@mosaic-media/sdui-react`), and the
  storybook. Pre-1.0, so a clean cut, not a dual-format transition.
- **One buf/protoc toolchain replaces quicktype/JSON-Schema generation.** The
  drift guard, conformance tests and `generate.sh` are re-expressed for buf.
- **The typing win is the skeleton, actions and region-ops.** Open props/attributes
  remain `Struct`; no generated per-component prop types (that is the governance
  fork, still open).
- **The storybook's "render beside its `UINode`" view goes through `protojson`**,
  since binary protobuf is not human-readable.
- **Tokens are untouched** (DTCG JSON).
- **The `sdui` → `contracts` rename** is a GitHub rename (redirects) plus a Go
  module-path change `github.com/mosaic-media/sdui` → `…/contracts` — coordinated
  with its importers and retagged, exactly like the [architecture#2](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0002-repository-naming-convention.md)
  renames. The npm name `@mosaic-media/sdui` is unaffected (scope is decoupled
  from repo name).
- **This amends [contracts#3](0003-sdui-contract-repository.md)** (format and repo
  identity), **resolves [contracts#5](0005-cross-client-transport-two-lane-rpc.md)'s
  UINode encoding to (b)** and relocates its contract, and is consistent with
  [web#3](https://github.com/mosaic-media/web/blob/main/docs/adr/0003-frontend-workspace.md).

## Implementation implications

**`contracts`** (renamed from `sdui`): a buf workspace. `proto/mosaic/sdui/v1/`
carries the content messages; `proto/mosaic/session/v1/session.proto` carries the
`SessionService`, importing sdui so `ui_node` is a typed `UINode`. `buf.gen.yaml`
emits Go and TS into published `gen/` paths (Swift/Kotlin/Dart later); `definitions/`
becomes protojson data; `tokens/` (DTCG) is unchanged. Apache-2.0 across both
modules.

**`platform`**: delete the in-repo `proto/`, `internal/gen/` and buf config; add a
dependency on `github.com/mosaic-media/contracts`; rewire `internal/transport/session/*`
from `…/platform/internal/gen/…` to `…/contracts/gen/…`; rebuild the emit-side
against the protobuf sdui Go binding. The transport *implementation* stays.

**`web`**: `@mosaic-media/sdui-react` consumes the generated TS bindings; the
storybook reads UINode via protojson.

**Sequence:** migrate the sdui module in `contracts` → add the session module
importing it → regenerate Go + TS → rewire `platform` (emit-side + transport) →
update `web` (runtime + storybook) → retire the JSON-Schema/quicktype pipeline. It
can proceed independently of finishing [contracts#5](0005-cross-client-transport-two-lane-rpc.md)'s client port, since it changes the
contract and generation, not the transport's runtime shape.
