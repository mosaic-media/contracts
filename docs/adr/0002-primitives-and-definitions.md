# Primitives and definitions: no component holdouts

**Status:** Accepted
**Date:** 2026-07-20

## Context

Under Server-Driven UI ([contracts#1](0001-server-driven-ui-and-the-shell.md)) the contract must be technology-agnostic and renderable by every client, and a module should be able to contribute UI without shipping client code. If each visual component were hand-written per client, neither holds: a module could not add a component as data, and the web and Flutter clients would drift apart component by component.

But some widgets genuinely cannot be expressed as static data. A text field owns its value. Tabs own which panel is shown. A progress bar's fill is computed from a number. A skeleton shimmers on a keyframe. Pretending these are "just data" would require a reactive runtime inside the UI layer — interaction logic smuggled into a data format.

The question is therefore not "native or data" as a matter of taste, but a precise line: **what can a static tree of data express, and what genuinely cannot?**

## Decision

Two kinds of thing, split by expressibility. **There are no hand-coded component holdouts** — if something is a composition, it is data, containers included.

**1. Primitives** — the irreducible, client-implemented leaves. They are the only native UI code, and they *are* the technology-agnostic vocabulary each client implements:

- *Presentational* — `Box`, `Text`, `Image`, `Icon`, `Spacer`, plus `Fragment`/`Outlet`. These take a **token-only style vocabulary**: flexbox and grid, spacing / colour / radius / type tokens — deliberately the intersection of what web (flex/grid + CSS variables) and Flutter (`Row`/`Column`/`Container` + `ThemeData`) render identically. No raw pixels or hex, no `:hover`, no web-only CSS.
- *Interactive / stateful* — `Pressable`, the bare inputs (`TextInput`, `Switch`, `SelectInput`), `Tabs`, `Menu`, and selection controls. Each owns its own state.
- *Computed / animated* — `ProgressBar`, `Skeleton`.

**2. Definitions** — every *composition*, expressed as data. A `ComponentDefinition` is a name, default params, and a `template` — itself a tree of primitives. A small client-side expander turns a node of that type into the expanded tree. This is what a module ships to contribute a component; the same definition renders identically on any client. The template markers are minimal: `$bind` (with dot paths), `$match` (map an enum to a value), `$if` / `$ifNot`, `$each` (repeat over an array), `Outlet` (the caller's children or a named slot), and an injected `$childCount`.

**The boundary rule.** A thing must be a primitive **if and only if** it owns local state, its rendered output or emitted action couples to that state, or it is computed/animated. Everything else — including all containers (screen, section, grid, carousel, divider) and the presentational body of every media component — is a definition.

## Alternatives considered

**Hand-code components per client.** The holdout problem restated: modules cannot contribute UI as data, and clients drift. *Rejected.*

**A reactive state / binding runtime in the SDUI engine**, so even stateful widgets become "data". This pushes interaction behaviour into a data-soup layer and is a large runtime to build and to port to a second client. *Rejected* — encapsulate state in a bounded, named set of interactive primitives instead.

**Free-form per-node styling (arbitrary CSS).** Expressive on the web, but it does not port to Flutter and leaks web-isms into the shared contract. *Rejected* — styles are constrained to tokens.

## Consequences

- The native surface is **bounded and auditable**: only the primitive vocabulary is client code, and it is exactly the contract each client implements. Everything else is data the Platform or a module can author.
- The **Mosaic Design Language** becomes a token swap, not a component rewrite.
- The Shell reconstructs *its own entire component set* — containers included — from the primitives, which is the standing proof the vocabulary is expressive enough for a module to build any look from data.
- Known boundaries, each a candidate future vocabulary addition rather than a defect: interaction states (`:hover`) are the interactive primitives' job, not static styling; bindings do not compute (no substring or arithmetic — e.g. pagination takes server-supplied prev/next actions rather than deriving page ± 1); responsive layout uses flex-wrap rather than true breakpoints.

## Implementation

In `mosaic-shell`: `sdui/style.ts` is the token style vocabulary and its web translation; `components/primitives.tsx` and the interactive/computed leaves are the primitives; `sdui/template.tsx` is the definition type and expander; `components/definitions.ts` and `components/definitions.layout.ts` rebuild the Shell's own components from primitives; `mock/moduleComponents.ts` is a simulated module contributing components the same way. A Flutter client implements the same primitive vocabulary and the same expander, and renders every definition unchanged.
