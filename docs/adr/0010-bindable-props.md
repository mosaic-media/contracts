# Bindable props, and no expression language

**Status:** Accepted (built)

**Date:** 2026-07-25

## Context

A prop value has been a literal in every position, and there was exactly one
escape from that: `SubmitField` substitutes the literal string `$value` into its
action before dispatching. That mechanism is specced nowhere, carries one field,
and substitutes *everywhere the string appears* in the action —
`module-tmdb` documents working around precisely that when a custom catalog
needed both a name and a type in one action.

It is why no screen in Mosaic can have two inputs on it, and therefore why
onboarding, sign-in and the admin surface that would discharge the
more-than-one-user debt are all blocked behind the same gap.

The pressure to fix it comes with an obvious wrong answer attached. Every SDUI
framework that has tried this has been asked, immediately afterwards, for
formatting in the binding, then defaults, then conditionals, then arithmetic —
and Spotify's HubFramework, the most-cited component-driven SDUI of its
generation, was deprecated with a retrospective titled *"The Silver Bullet That
Wasn't"* for exactly that: over-generalising into a UI virtual machine.

## Decision

**A prop value may be a literal, or a binding. A binding names a path and
nothing else.**

```json
{ "type": "Text", "props": { "text": { "$bind": "title" } } }
```

- **No formatting, no default, no expression, no operator.** Formatting stays
  server-side. That is what "presentation-ready data" has always meant here, and
  it is the honest reason this needs no evaluator — not an omission to be filled
  in later. A request for `{"$bind": "x", "format": "…"}` is a request for the
  virtual machine, and the answer is that the server formats it.
- **The marker is `$bind`, the spelling a definition template already uses**,
  because it means the same thing: resolve a path against a scope. They were two
  mechanisms that happened to share a spelling; the client now reads both through
  one function, so "is this a binding?" has one answer.
- **An object is a binding only with exactly that one key and a non-empty string
  path.** Carrying the marker beside anything else leaves a literal object. This
  is the same closed reading the action-kind check uses
  ([platform#52](https://github.com/mosaic-media/platform/blob/main/docs/adr/0052-vocabulary-negotiation-and-deliberate-degradation.md)) and for
  the same reason: a producer's own data with a field of that name must never be
  silently replaced by a lookup.
- **A malformed binding is refused, not passed through.** An empty path or a
  non-string path is an error at the boundary rather than a literal object that
  draws nothing — the failure shape this contract keeps having.
- **A node's `type` may not be bound.** A tree whose shape depends on data is a
  `ComponentDefinition`, and definitions are authored in the contract. Allowing
  it would put template expansion into the wire format.
- **Resolution is at render, against the screen's params, and only those.**
  Later scopes — state, form fields — resolve *nearest-first above* this one, so
  what a binding means today does not change when they arrive; it gains places to
  look first. Resolving at render rather than once on arrival is deliberate: a
  re-render with different params must re-resolve, and a tree stored
  half-resolved is one nobody can reason about once there is more than one scope.
- **An unresolved binding yields absence, not the empty string.** The prop is
  deleted and the component falls back exactly as if the server had not set it.
  `""` would make "no value" and "the empty value" indistinguishable, which is
  the shape of every silent-prop failure in this project's history.
- **Authoring is a generated `Bind*` helper per prop**, not
  `Prop("title", Bind(…))`. A prop set by string is the failure this contract
  keeps having — `ui.Subtitle` on a `Stack` drew nothing for a screen's whole
  life — and an untyped escape hatch for every bound prop would put them all back
  on that footing.

**This slice is purely additive.** Every literal stays a literal, nothing
existing changes meaning, and a client that ignored bindings entirely would
behave exactly as before. The vocabulary goes to 1.2.0, not 2.0.0.

**`$value` is not retired here.** It is used at nine sites across four
separately released modules (`module-tmdb`, `module-fanart-tv`,
`module-aiostreams`, `module-stremio-addons`) and at none in the Platform.
Retiring it needs the form scope that replaces it, so it survives until then;
breaking it now would leave four external repositories with no replacement for
two slices.

## Alternatives considered

**A small expression language — `{"$expr": "upper(title)"}`.** *Rejected*, and
this is the load-bearing rejection. It is the HubFramework failure by name: once
the wire can express computation, every client must implement the same evaluator
with the same semantics, and the conformance surface stops being a list of types
and becomes a language specification. Four client languages make that
prohibitive, and the server can already do it.

**Formatting options on the binding.** *Rejected* — the thin end of the same
wedge. Number and date formatting are locale-dependent, and a locale the server
knows and the client guesses is a bug waiting in every screen.

**Resolve once when the tree arrives, then store it resolved.** *Rejected* — it
is cheaper and it is wrong the moment there is more than one scope: a tree
resolved against screen params and then re-rendered inside a state scope would
carry stale values with no way to tell which. Resolving at render costs one
shallow scan per node, and returns the identical object when nothing is bound.

**Make bindings a typed protobuf field rather than a convention in the open
props bag.** *Rejected for now* — props is a `Struct`, so a binding is JSON
either way, and a typed field would mean every prop becomes a wrapper message.
That is the ergonomics redesign [contracts#6](0006-contracts-protobuf-workspace.md)
already deferred, and it is a bigger change than this one.

**Bind positional constructor arguments too.** *Not done* — a positional
argument is a typed Go/TypeScript parameter, and widening it to accept a binding
would make it `any`, losing the typing the authoring layer exists for. Every
positional key is also a sugar key, so `Section("", BindTitle("t"))` binds it;
the literal is overwritten.

## Consequences

- **Primitive props still have no typed helper, bound or literal.** The `Bind*`
  helpers cover the sugar set, which is what components and definitions bind.
  Primitive props go through `ui.Prop`, because four prop keys still carry two
  types across primitives (`value`, `style`, `name`, `size`) and one typed helper
  per key is impossible until that is resolved. That remains owed.
- **The template expander and the wire now share one reading of a binding.** The
  expander's own check was `"$bind" in v` — looser than the contract's — so an
  object carrying the marker beside other keys was already being reinterpreted
  there. It is now the strict rule in both places.
- **Bindings resolve against a scope the client owns.** The server cannot know
  what a binding will resolve to, which is the point, and also means a binding to
  a param the server did not send draws nothing. The unknown-type report from
  [platform#52](https://github.com/mosaic-media/platform/blob/main/docs/adr/0052-vocabulary-negotiation-and-deliberate-degradation.md) has no equivalent for props; that gap is real and unaddressed.
- **Verified live**: the search screen emitting `title: {"$bind": "text"}`
  rendered the search term as its heading against the running dev stack, and the
  definition library still expanded correctly beneath it.
