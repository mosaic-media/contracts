# Accessibility in the contract

**Status:** Accepted (built)

**Date:** 2026-07-25

## Context

Nothing in the SDUI vocabulary could say what a node *is*. A `Box` with a border
and a heading-sized `Text` inside it is a section to a sighted reader and an
undifferentiated div to everything else, and the client had no information to do
better with — so it inferred, or it did nothing.

This slice is in the thread's ordering for a reason that has nothing to do with
urgency. The ordering principle is **retrofit hostility**, and accessibility
props are the clearest case of it: an optional prop added later is additive, but
a vocabulary that never had a role model cannot grow one without revisiting every
definition — and a client that has spent a year inferring roles from node types
has to be talked out of its inferences before it can be given facts.

## Decision

**Role, accessible name, heading level and live-region politeness are props any
node may carry.**

- **The role set is closed**, like the validators and the predicates, and for a
  sharper version of the same reason. A role is not passed through: the web maps
  it to an ARIA attribute, Compose to a semantics property, SwiftUI to a trait. A
  role the server can name and a client maps to nothing produces a control that
  is **invisible to a screen reader and looks perfectly correct to everybody
  else** — the least detectable failure this contract can produce, because the
  people who would notice are the ones least likely to be asked.
- **Nineteen roles, chosen for what this vocabulary contains** rather than
  transcribed from the ARIA specification. A transcription would be a list nobody
  could implement completely, which is an open set wearing a closed set's
  clothes.
- **An unrecognised role is dropped, not passed through.** Passing it to the DOM
  would set an invalid ARIA role, which browsers treat as *no* role — the same
  outcome, reached silently, with the contract's closed set bypassed.
- **The accessible name is `a11yLabel`, not `label`.** A visible label and an
  accessible name are different things; one key meaning both is how a control
  ends up announced as its own caption. This is the one-key-one-type rule doing
  real work rather than being cited.
- **Heading level is stated, not inferred from text size.** Size is a design
  decision and outline depth is a structural one, and a screen reader navigates
  the second. 1–6, because a level outside the outline has exactly one symptom:
  heading navigation skips it.
- **An icon with no accessible name is `aria-hidden`.** Declaring it decoration
  is the useful statement — a screen reader announcing "image" for every glyph in
  a row of controls is worse than one announcing none of them.

## Alternatives considered

**Infer roles from node types in each client.** *Rejected* — it is what happens
by default and it is wrong in the way that matters: the inference lives in the
client, so two clients infer differently and neither is checkable against
anything. It also cannot distinguish a `Box` that is a section from one that is
a spacer, which is the case that actually needs the answer.

**Let `role` be an open string, passed through to ARIA.** *Rejected* — it works
on exactly one platform. Every other client has a mapping table, and an open set
guarantees holes in it that nobody can enumerate.

**Transcribe the full ARIA role set.** *Rejected* — a closed set nobody
implements completely is an open set with a longer list, and the conformance
check would either fail permanently or be weakened to meaninglessness.

**Wait until a screen needs it.** *Rejected* on the thread's ordering principle.
This is the retrofit-hostile case: the cost of adding it now is four optional
props nothing sets, and the cost of adding it later is every definition plus a
client's accumulated inferences.

## Consequences

- **Nothing sets these props yet.** The contract declares them, the client maps
  them, and no screen carries one. That is the honest state, and it is the same
  shape as the lifecycle triggers: the vocabulary is ready before the emit-side
  is, which is the point of ordering by retrofit cost rather than by demand.
- **The conformance check covers roles** in both directions, alongside the
  validators and predicates. A role the contract declares and the client drops
  fails the build — which matters more here than elsewhere, because the symptom
  in production is silence.
- **Verified live**: a `role="search"` with an accessible name, a `role="heading"`
  with `aria-level="2"`, and a `role="status"` with `aria-live="polite"` all
  reached the accessibility tree; an undeclared `carousel` role was **dropped**
  while its node still rendered; and six nameless icons were marked
  `aria-hidden`.
- **Only `Box`, `Text` and `Icon` apply the props so far.** They are what every
  definition is built from, so most trees can be described — but a `Pressable`
  or an input cannot yet carry an accessible name distinct from its visible one.
  That is a gap, it is small, and it is owed.
- **Focus is not here.** Focusability, focus groups and initial focus are the
  next slice, and they are the half of accessibility that a keyboard user needs
  most. Declaring roles without focus is half the job, stated as half.
