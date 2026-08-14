# Bindings, scopes and forms, and no expression language

**Status:** Accepted (built). Consolidates the three input-handling records,
whose bodies this replaces and whose retired numbers stay retired — citing one
would either dangle or resolve to a later record that happens to hold it. Partly
superseded: the merge rule —
"the scope's values merge *under* the action's input" — was reversed by
[contracts#19](0019-retiring-value.md), which also removed the `SubmitField`
primitive this record left in place. The rest stands and is built.

**Date:** 2026-08-10

## Context

Three slices landed on **2026-07-25**, one after another, each opening by naming
what the last had left undone: bindable props, the scopes a binding reads, and
the fields and forms that write them. They are one decision about how a screen
carries user input, so they are recorded as one. The date above is the
consolidation's; the work itself is that single day, and no dated evidence
separates the three.

Before any of it, a prop value was a literal in every position, and there was
exactly one escape: `SubmitField` substituted the literal string `$value` into
its action before dispatching. That mechanism was specced nowhere, carried one
field, and substituted *everywhere the string appears* in the action —
`module-tmdb` documents working around precisely that when a custom catalog
needed both a name and a type in one action. A `TextInput` held what was typed in
the web client's own component state, where no other node could see it, so a tree
described one moment and nothing carried a value between two of them.

That is why no screen in Mosaic could have two inputs on it, and therefore why
onboarding, sign-in and the admin surface that would discharge the
more-than-one-user debt were all blocked behind the same gap. The first-boot
onboarding flow — a username, a password and a confirmation submitted together —
is the screen the whole thread was opened for, and there was no way to express
it.

The pressure to fix it comes with an obvious wrong answer attached. Every SDUI
framework that has tried this has been asked, immediately afterwards, for
formatting in the binding, then defaults, then conditionals, then arithmetic —
and Spotify's HubFramework, the most-cited component-driven SDUI of its
generation, was deprecated with a retrospective titled *"The Silver Bullet That
Wasn't"* for exactly that: over-generalising into a UI virtual machine.

So there are three questions and they are one question: where a prop reads
*from*; where it reads from something that *changes*; and the join between them —
an input that writes what it holds where other nodes can see it, and something
that collects those values and sends them.

## Decision

**A prop value may be a literal, or a binding. A `State` node declares the
variables a binding reads. A named input writes into the enclosing scope, and a
`Form` is a composition.**

### A binding names a path and nothing else

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
  ([platform#52](https://github.com/mosaic-media/platform/blob/main/docs/adr/0052-vocabulary-negotiation-and-deliberate-degradation.md))
  and for the same reason: a producer's own data with a field of that name must
  never be silently replaced by a lookup.
- **A malformed binding is refused, not passed through.** An empty path or a
  non-string path is an error at the boundary rather than a literal object that
  draws nothing — the failure shape this contract keeps having.
- **A node's `type` may not be bound.** A tree whose shape depends on data is a
  `ComponentDefinition`, and definitions are authored in the contract. Allowing
  it would put template expansion into the wire format.
- **Resolution is at render, against the screen's params, and — in this slice —
  only those.** The scopes below resolve *nearest-first above* the params, so what
  a binding means today does not change when they arrive; it gains places to look
  first. Resolving at render rather than once on arrival is deliberate: a
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

**The binding slice is purely additive.** Every literal stays a literal, nothing
existing changes meaning, and a client that ignored bindings entirely would
behave exactly as before. The vocabulary goes to 1.2.0, not 2.0.0.

**`$value` is not retired by it.** It is used at nine sites across four
separately released modules (`module-tmdb`, `module-fanart-tv`,
`module-aiostreams`, `module-stremio-addons`) and at none in the Platform.
Retiring it needs the form scope that replaces it, so it survives until then;
breaking it at the binding slice would leave four external repositories with no
replacement for two slices.

### A `State` node declares named variables and scopes them to its children

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
the client. At that point the conformance check read 26/27 primitives and 10/12
action kinds; `Form` and `submit` remained, and belong to forms.

### A named input writes into the enclosing scope, and a `Form` is a composition

- **Every input takes a `name`** — the field it reads and writes in the nearest
  enclosing `State` scope that declares it. `SearchBar` and `SubmitField` were
  the two without one, which is exactly why a screen could carry either of them
  once and no more.
- **An input *without* a name keeps the behaviour it had**: local state, invisible
  to everything else. That is deliberate rather than transitional — a search box
  inside a form is not one of the form's fields, and making every input join the
  nearest scope would put it there.
- **A name no scope declares falls back to local state** rather than making the
  control inert. An unusable control is a worse answer to a misconfigured screen
  than a control whose value nothing collects.
- **`Form` leaves the primitive tier.** It was declared native in
  [contracts#8](0008-one-generated-sdui-vocabulary.md) and implemented by nobody. It
  is a `State` scope, an `Outlet` and a button — three things that already exist
  — so it is `definitions/form.json`, data like every other composition. Nothing
  about it needs client code, which is the test for the tier.
- **`submit` carries the action it runs**, in the same field a `sequence` carries
  its steps, rather than the scope carrying a `submitAction` the client goes
  looking for. This is what keeps the composition honest: the scope stays a plain
  `State`, the button stays a plain `Pressable`, and the only thing that knows
  about submission is the action — which is data.
- **The scope's values merge *under* the action's input, not over it.** A server
  that pinned a field in the action it sent is stating something the form must
  not overwrite.
- **Only declared names that hold a value are collected.** An untouched variable
  with no initial value is absent from the payload rather than sent as an empty
  string, so a server can tell "left blank" from "cleared".
- **Nearest scope wins when collecting**, as it does everywhere else. A submit
  that collected an outer scope's value for a name an inner one redeclared would
  send something no control on the screen was showing.

**The vocabulary goes to 2.0.0.** A primitive left the tier. It is one no client
ever implemented, so nothing observable breaks and there was nothing to
deprecate — but the version is what a client reasons against, and quietly calling
a tier move additive is how a version stops meaning anything.

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

**Keep `Form` native and let it own the scope and the submit.** *Rejected* — it
is the special case this slice exists to avoid. A native `Form` means every
client implements form semantics identically, and the moment anyone wants a form
laid out differently the answer is a client release. As a definition, the layout
is data and the semantics are two things the vocabulary already has.

**Have `State` carry the `submitAction`, so `submit` needs no payload.**
*Rejected* — it makes `State` know about forms, which is precisely the
special-casing being removed. `State` would then have a prop meaningful only when
a descendant happens to be a submit button.

**Collect every input in the subtree rather than the scope's declared
variables.** *Rejected* — it makes the payload depend on what is rendered rather
than on what was declared, so a conditionally hidden field silently changes the
shape of the submission. Declared variables are a contract; rendered inputs are
a consequence.

**Retire `$value` at the forms slice, as the roadmap scheduled.** *Not done* —
see the consequences below. The replacement exists; the migration does not.

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
  [platform#52](https://github.com/mosaic-media/platform/blob/main/docs/adr/0052-vocabulary-negotiation-and-deliberate-degradation.md)
  has no equivalent for props; that gap is real and unaddressed.
- **A binding that resolves to nothing is silent**, and scopes make it more
  likely rather than less: a name misspelled in a binding now has two places it
  could have been looking. Still unaddressed.
- **A screen can carry more than one input's worth of state**, which is the block
  this thread was opened to remove.
- **`State` is a primitive that draws nothing**, which is a new shape in the
  tier — every other primitive renders something. The negotiation from
  [platform#52](https://github.com/mosaic-media/platform/blob/main/docs/adr/0052-vocabulary-negotiation-and-deliberate-degradation.md)
  handles it correctly by construction: a client that does not declare `State`
  has the whole subtree dropped, which is right, because the children would bind
  to nothing.
- **Nothing emitted a `State` when the scopes slice landed.** The Platform had
  the builders and did not use them; the capability was verified with a temporary
  emit against the running dev stack, not by shipping a screen. That was
  deliberate: the screens that want it are the ones forms unblock.
- **A screen can carry more than one input and submit them together.** Verified
  live: a two-field form on a running screen produced
  `{"username":"adam","password":"hunter2"}` on the wire, from two inputs that
  had never been able to coexist.
- **Verified live for bindings too**: the search screen emitting
  `title: {"$bind": "text"}` rendered the search term as its heading against the
  running dev stack, and the definition library still expanded correctly beneath
  it.
- **The client's declared gap is one action kind.** `26/26 primitives, 11/12
  action kinds` — `query` alone remains, and it belongs to partial region
  refresh.
- **`$value` still works and is still the only thing four modules use.** It is
  used at nine sites across `module-tmdb`, `module-fanart-tv`,
  `module-aiostreams` and `module-stremio-addons` — four separately released
  repositories, none of which is a Platform dependency. Retiring it is a
  migration of those four plus the removal here, and it is **owed**, not done.
  The roadmap scheduled the retirement for this slice; what landed is the
  replacement it needs.
- **A signature change broke one caller**, a Platform test using `ui.Submit()`
  with no argument, and a second test asserted a gap naming `Form` that stopped
  being a primitive. Both are the major bump being real rather than ceremonial —
  and the second is worth naming, because it had started passing by asserting
  nothing.
- **Field-level validation is not here.** The `validators` and `visibleWhen`
  props are declared on every input and enforced by nobody. That is the next
  slice, and until it lands a server can state a rule no client applies.
