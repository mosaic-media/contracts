# Fields and forms

**Status:** Partly superseded: the merge rule — "the scope's values merge *under*
the action's input" — was reversed by [contracts#19](0019-retiring-value.md), which
also removed the `SubmitField` primitive this record left in place. The rest
stands and is built.

**Date:** 2026-07-25

## Context

This is the slice the whole SDUI thread was opened for. A first-boot onboarding
flow needs a username, a password and a confirmation submitted together, and
there was no way to express it: an input held its value in the client's own
component state where nothing else could read it, and the only route out was
`SubmitField` substituting the literal string `$value` into an action — one
field, everywhere the string appeared, specced nowhere.

[contracts#10](0010-bindable-props.md) gave a prop somewhere to read from and
[contracts#11](0011-state-scopes.md) gave the vocabulary somewhere to remember. What
was missing was the join: an input that writes what it holds where other nodes
can see it, and something that collects those values and sends them.

## Decision

**A named input writes into the enclosing scope. A `Form` is a composition.**

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

**Retire `$value` now, as the roadmap scheduled.** *Not done* — see below. The
replacement exists; the migration does not.

## Consequences

- **A screen can carry more than one input and submit them together.** Verified
  live: a two-field form on a running screen produced
  `{"username":"adam","password":"hunter2"}` on the wire, from two inputs that
  had never been able to coexist.
- **The client's declared gap is now one action kind.** `26/26 primitives, 11/12
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
