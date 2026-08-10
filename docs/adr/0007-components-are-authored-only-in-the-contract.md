# Components are authored only in the contract, and clients bundle none

**Status:** Accepted (built)
**Date:** 2026-07-24

## Context

[contracts#2](0002-primitives-and-definitions.md) drew the line the whole interface
rests on: a small set of **primitives** is native code each client implements,
and every composition is a **definition** — a primitive tree, *data, not code*.
[contracts#4](0004-server-delivered-definitions-and-skin.md) then moved the
definition library out of the client and had the Platform serve it, so a new
component costs no client release.

The implementation did neither.

- **Components were authored in the web renderer.** About thirty definitions
  lived as hand-written TypeScript in `sdui-react`'s `definitions.ts` and
  `definitions.layout.ts`. The "server-delivered" library the Platform embedded
  was produced by *building that package and dumping its objects to JSON*
  (`scripts/dump-definitions.mjs`). So the source of the design system was a
  client, and the Platform served a copy of it.
- **The published contract carried a second, stale library.** `mosaic-sdui`'s
  `definitions/` held four components — `Badge`, `Section`, `PosterCard`,
  `HeroBanner` — schema-conformance-tested and lint-checked against the authoring
  spec, and **three of the four had drifted** from the copies that actually
  rendered: the contract's `HeroBanner` had none of the variant/art-light/gradient
  work, its `PosterCard` no glass or border, its `Section` a different gap. No
  test failed, because nothing compared them.
- **The authoring layer had fallen behind by more than half.** Nineteen of the
  thirty-four components had no entry in `ui.spec.json` at all, so the Platform
  authored them as `ui.Component("MediaTile", ui.Prop("title", …))` — strings
  into an open props bag. `ui.Subtitle` set on a `Stack`, which has no subtitle
  in its template, drew nothing for the life of a screen and reported nothing.
- **The lint could not have caught up.** `genui -lint` treated a `$each` alias's
  dotted paths (`s.label`) and runtime-injected bindings (`$childCount`) as
  unauthorable props, so *any* definition using a repeat or a child count failed
  it — which is a large share of them, and a standing reason not to move them
  into the contract.

The cost is not stylistic. A component written in a client renders on that client
and nowhere else: the Flutter client [contracts#4](0004-server-delivered-definitions-and-skin.md)
exists for would have had to reimplement thirty components from a TypeScript file
that was not the contract, and keep them in step with a copy that had already
proven it drifts silently.

## Decision

**A component is authored exactly once, as data, in the contract. Clients bundle
no definitions at all.**

- **`mosaic-sdui`'s `definitions/*.json` is the only place a component is
  written** — one file per component. All thirty-four now live there.
- **The aggregates are generated, never authored**: `definitions.Library()` (Go
  embed) is what the Platform serves; `ts/definitions.gen.ts` is for a JavaScript
  consumer with no Platform to ask (the storybook). Both are guarded by the
  repository's existing drift check.
- **The client bundles nothing.** `sdui-react` registers the native vocabulary —
  primitives, action kinds, the style translation, the registry, the renderer,
  the expander — and no components. **This reverses [contracts#4](0004-server-delivered-definitions-and-skin.md)'s fallback clause**
  ("the client falls back to its bundled definitions if the payload is
  malformed"): a bundled fallback is a second copy of every component, and a
  second copy is exactly what drifted. A client with no library renders the
  `Unknown` placeholder if the payload fails, which is a visible, diagnosable
  failure rather than a silently different interface.
- **The authoring layer must cover the library, enforced.** `ui.spec.json` now
  exposes a helper for every prop every definition binds, and the lint — fixed to
  understand `$each` aliases and runtime-injected bindings — fails when it does
  not. The Platform authors screens with generated, typed builders.
- **One prop key means one type across every component.** A switch's state is
  `on` because `value` is a field's text; a card's provenance is `origin` because
  `meta` is a hero's variadic line. Typed helpers make the collision a compile
  error instead of a silent coercion.

## Alternatives considered

**Keep authoring in the client, delete the contract's copy.** *Rejected.* It
removes the drift but keeps the design system inside one client, which is the
larger problem: a second client cannot render what it cannot see, and "server-
delivered" would remain a dump of a renderer's build output.

**Keep a generated fallback in the client, built from the contract.** *Considered
and rejected by the same reasoning that motivates this record.* It would not
drift, being generated — but it is still a copy shipped in a client release, and
it preserves the idea that a client may hold components. The failure it protects
against (a malformed payload) is better made loud.

**Move only the new components and migrate the rest later.** *Rejected.* That is
the two-library state this record exists to end; a partial move leaves exactly the
condition under which the four copies went stale.

## Consequences

- **A component is now a contract change**, with the release that implies: tag
  `mosaic-sdui`, bump the Platform's `go.mod` and the web workspace's npm
  dependency. That is a real cost and it is the correct one — it is the same cost
  as any other change to a published contract, and it buys a component that every
  client renders identically.
- **A new screen costs no client release**, which is what [contracts#4](0004-server-delivered-definitions-and-skin.md) promised and
  this makes true: the Platform composes from the library and serves it.
- **`sdui-react` becomes smaller and its job becomes describable in one line** —
  the native vocabulary and the machinery to expand data into it. Removing the
  bundled library is a breaking change for the package (`0.2.0`).
- **A malformed definitions payload now degrades visibly.** There is no bundled
  set to fall back to, so components render as `Unknown`. This is deliberate: the
  alternative is a client quietly rendering a different interface from the one the
  Platform serves, which is the failure mode that hid the drift for months.
- **The storybook consumes the contract like any other client** — it imports the
  generated library rather than holding a copy.
- **The style vocabulary is the remaining gap**, and it is named here rather than
  fixed: `BoxStyle`/`TextStyle` still exist only in the web package, so the
  fields a definition's template uses (including `responsive` and `hidden`) have
  no client-agnostic spec. A Flutter client can read every definition and still
  not know what `radius: "md"` or `responsive: {below: 720}` obliges it to do.
  Specifying it in the contract is the next slice.
