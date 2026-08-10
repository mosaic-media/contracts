# State scopes

**Status:** Accepted (built)

**Date:** 2026-07-25

## Context

[contracts#10](0010-bindable-props.md) gave a prop somewhere to read *from*. It did
not give the vocabulary anywhere to read from that changes.

A tree describes one moment. Nothing carried a value between two of them: a
`TextInput` held what was typed in the web client's own component state, where no
other node could see it, and the only way to get a value back out was
`SubmitField` substituting the literal string `$value` into an action — one
field, everywhere it appeared in that action, specced nowhere. That is the whole
reason no screen in Mosaic can carry two inputs, and therefore why onboarding,
sign-in and the admin surface that would discharge the more-than-one-user debt
are all blocked behind the same gap.

## Decision

**A `State` node declares named variables and scopes them to its children.**

- **It is a primitive**, and it earns the tier on two counts rather than one: it
  holds values across renders, which no arrangement of static nodes can, and it
  *is* the scope boundary — which of two enclosing `State`s a name resolves to is
  a question only the renderer, walking its own tree, can answer.
- **Variables are declared, not inferred from what is written.** A client must
  know a name exists before anything writes to it, so a typo is a refusal rather
  than a new variable nobody reads; and it must know what type to coerce a
  written value to, because `setValue` carries its value as a string on the wire.
- **Three types — `string`, `number`, `boolean` — and the set is closed**, for
  the same reason the validator and predicate sets are closed. Each is a coercion
  every client must implement identically; one the server can name but a client
  coerces differently is a value that silently differs between platforms.
- **An absent initial value stays absent.** It is not a zero. A number with no
  initial value is unset, and inventing `0` for it is a slider that jumps to zero
  on first render — the visible form of conflating "unset" with "the zero value".
- **Resolution is nearest-first**: the innermost scope that *declares* the name,
  then outward, and finally the screen's params. Declares, not holds — a scope
  that merely happened to have a value would capture a name it was never given,
  and a declared variable holding nothing must stop the search rather than fall
  through to a screen param that happens to share its name.
- **`setValue` is handled where the control is, not where the provider is.** It
  is the one action whose meaning depends on the position it was emitted from,
  so the runtime a component sees composes its scope in. A `sequence` is walked
  there too, so a `setValue` nested inside one still resolves against the
  emitting component's scope rather than against nothing.
- **An undeclared write is refused**, with a message naming the field, rather
  than creating a variable nobody reads. A write to a name no enclosing scope
  declares is either a typo or a control outside the scope it thinks it is in,
  and both are worth hearing about.

This closes the `setValue` half of the gap between the declared vocabulary and
the client. The conformance check now reads 26/27 primitives and 10/12 action
kinds; `Form` and `submit` remain, and belong to forms.

## Alternatives considered

**One global state bag per screen, no scopes.** *Rejected* — it is simpler until
a screen has two of the same thing on it, which is the first screen anyone will
build with this: a list of module rows each with its own field. Two rows would
share one variable and overwrite each other, and the fix at that point is scopes,
retrofitted through every definition that had assumed a flat bag.

**Infer variables from the first `setValue` that names one.** *Rejected* — it
makes a typo indistinguishable from a new variable, which is exactly the silent
failure this thread exists to remove, and it leaves a client with no type to
coerce a written string to.

**Let `setValue` carry a typed value rather than a string.** *Rejected for now* —
it means a union on the wire that four client languages must agree about, to save
one declared type per variable. The declaration is needed anyway.

**Hold state in the primitives themselves, as `TextInput` already did, and add a
way to read a sibling's state.** *Rejected* — it is the same design that produced
`$value`, one indirection further on. Reading across a tree is what a scope is.

**Resolve bindings against a merged object of all scopes plus params.**
*Rejected* — flattening loses the order, so an outer scope's value would shadow
an inner one depending on merge direction. The lookup walks the chain instead.

## Consequences

- **A screen can now carry more than one input's worth of state**, which is the
  block this thread was opened to remove. It cannot yet *submit* them — that is
  forms, and it is next.
- **Nothing emits a `State` yet.** The Platform has the builders and does not use
  them; the capability was verified with a temporary emit against the running dev
  stack, not by shipping a screen. That is deliberate: the screens that want it
  are the ones forms unblock.
- **`State` is a primitive that draws nothing**, which is a new shape in the
  tier — every other primitive renders something. The negotiation from
  [platform#52](https://github.com/mosaic-media/platform/blob/main/docs/adr/0052-vocabulary-negotiation-and-deliberate-degradation.md) handles
  it correctly by construction: a client that does not declare `State` has the
  whole subtree dropped, which is right, because the children would bind to
  nothing.
- **A binding that resolves to nothing is still silent**, exactly as
  [contracts#10](0010-bindable-props.md) recorded. Scopes make it more likely rather
  than less: a name misspelled in a binding now has two places it could have been
  looking. Still unaddressed.
