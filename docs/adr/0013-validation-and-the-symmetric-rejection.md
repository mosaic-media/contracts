# Validation, and the symmetric rejection

**Status:** Accepted (built)

**Date:** 2026-07-25

## Context

[contracts#20](0020-bindings-scopes-and-forms.md) made a screen able to carry several fields
and submit them together. It could not say that any of them was wrong.

The vocabulary has declared six validators and six predicates since it was
unified ([contracts#8](0008-one-generated-sdui-vocabulary.md)), and until now
**nothing enforced either**. A `validators` prop sat on every input and no client
read it. That is the closed-set discipline running in reverse: the set is closed
precisely so a server cannot state a rule a client fails to apply, and for four
slices the server could state all six and no client applied any.

The other half was the return journey. A server rejects things a client cannot
know — a username already taken, a key the upstream refused — and the only way it
could say so was a toast. A toast is a sentence floating next to a form; on a
form with four inputs it does not say which one, and it does not survive the
next re-render.

## Decision

**The six validators and the six predicates are enforced, and a rejection has
one shape whichever side produced it.**

- **Validators run on the client before anything is sent**, in a **fixed order**
  rather than object order — the same bad value must produce the same message on
  every platform and in every run, and `required` wins over `minLength` because
  "Required" is the message a person can act on.
- **The client is not a trust boundary.** The server enforces everything again.
  What the client's copy buys is the round trip: a field that is obviously wrong
  says so without one.
- **An unknown validator is refused, not ignored.** Ignoring it produces a field
  that accepts anything, silently, with the bad data surfacing somewhere else
  later. That is the fail-open case the closed set exists to prevent, and it is
  why the set may not be widened.
- **A pattern that will not compile does not fail the field.** It is the server's
  mistake, and failing the field blames the person typing for it. Patterns are
  RE2 on the Go side, so no input can make a validator backtrack.
- **An empty value is `required`'s business, not `pattern`'s** — otherwise every
  optional patterned field is effectively required.
- **Every registered field is validated on submit, including ones not on
  screen.** A conditionally hidden required field would otherwise pass by not
  being rendered. Rules are registered into the scope by the input that states
  them, so the rule stays next to the control and the check does not have to walk
  the tree.
- **Editing a field clears its rejection.** Leaving it up while the value changes
  underneath is a message about something that is no longer true.
- **`visibleWhen` is a predicate, not a prop of its own.** A "show this when that
  field is set" flag answers one screen and earns a second flag for the next.
- **An unreadable predicate is `false`.** This is the load-bearing direction:
  `visibleWhen` deciding to *show* a control because it could not understand its
  own rule is the fail-open case, and it is the one that puts an admin-only
  affordance on somebody's screen.
- **`FieldErrors` is symmetric.** A Platform error carries `Fields`, the session
  transport pushes them as `mosaic.session.v1.FieldErrors`, and each `State`
  scope adopts the names it declares. It is the same envelope the client's own
  validators fill, so a rejection from either side renders in the same place —
  otherwise a screen tells you where the problem is in two different ways.
- **A field name matching nothing is not dropped**; it stays a form-level
  message. A rejection nobody can see is worse than one in the wrong place.

## Alternatives considered

**Let the validator set be open, and have clients ignore what they do not
know.** *Rejected* — it is the fail-open case stated as a feature. A rule nobody
enforces is a field that accepts anything with no symptom at the point of
failure.

**Validate only on the server and push the results back.** *Rejected* — it makes
every typo a round trip, and it does not remove the need for the envelope, so it
costs the latency and saves nothing.

**Declare validators on the State variable rather than on the input.**
*Considered, and close.* The variable is where the contract already declares the
name and type, so it is the tidier home. The input won because a rule belongs
next to the control that states it — `minLength` on a password box reads as part
of that box — and registration into the scope recovers the collectability the
variable would have given for free. If a second client finds registration
awkward, this is the decision to revisit.

**Put the rejection in the `Invoke` reply rather than on the push lane.**
*Rejected* — [contracts#5](0005-cross-client-transport-two-lane-rpc.md) settled that
an intent's visible effect arrives on the push lane, and a rejection is a visible
effect. Answering in the Ack would make submission the one intent that works
differently.

**Reuse the toast for field errors, keyed by field.** *Rejected* — it is the
status quo with extra steps, and it keeps the message somewhere other than the
field it is about.

## Consequences

- **The client's declared gap is one action kind.** `26/26 primitives, 11/12
  action kinds, 6 validators, 6 predicates`. Only `query` remains.
- **The conformance gate now checks both closed sets in both directions.** A
  validator the contract declares and the client does not implement fails the
  build, and so does one the client implements that the contract does not
  declare. Both were checked by breaking them.
- **Verified live**: a form with `required`/`minLength` on one field and
  `matches` on another produced "Must be at least 4 characters" and "Does not
  match username", marked both fields, and sent **nothing** — zero Invokes.
  Correcting the values cleared the messages and the submission went, exactly
  once.
- **Almost nothing in the Platform produces a field rejection yet.**
  `contracts.RejectFields` exists and the transport routes it; no command calls
  it. The first will be whatever creates a user, which is the screen this whole
  thread was opened for.
- **`$value` still works and is still owed.** Nothing in this slice changed that
  — see [contracts#20](0020-bindings-scopes-and-forms.md).
