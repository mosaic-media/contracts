# The conformance corpus

**Status:** Accepted (built)

**Date:** 2026-07-25

## Context

[contracts#8](0008-one-generated-sdui-vocabulary.md) published the vocabulary as
data, so "what does this contract contain?" became a question with an answer a
client could check itself against. It says nothing about how any of it
*behaves*, and behaviour is where implementations diverge.

Two implementations of "resolve a binding" agree about the easy cases. They
differ about the empty string, the missing path, the object that merely looks
like a binding, and what a validator's message actually says. None of those
disagreements shows up as an error — they show up as a screen that behaves
differently on one platform, noticed by whoever is using that platform.

This is the mechanism DivKit uses to hold four clients in parity, and it is what
the whole thread has been building toward: a second Mosaic client is currently a
tractable piece of work or an act of faith depending entirely on whether this
exists.

## Decision

**Four files of golden cases, run by every implementation.**

- **Validators assert the message, not pass/fail.** A message is what a person
  reads, and two clients disagreeing about wording is two products. 19 cases,
  including the ordering — `required` beats `minLength` because it is the one a
  person can act on.
- **Predicates include every unreadable shape.** 24 cases, and the ones that
  matter most are the malformed ones: an unrecognised predicate must be `false`
  on every client, because `visibleWhen` showing a control it could not reason
  about is what puts an admin-only affordance on somebody's screen.
- **Bindings include the near-misses.** 11 cases: the marker beside another key,
  a near-miss spelling, an unresolved path. Getting these wrong silently replaces
  a producer's own data with a lookup.
- **Expansion states the definition rules as inputs and outputs.** 12 cases
  covering `$bind`, defaults, `Outlet`, named slots, `$if`, `$each` and `$match`.

**Trees are compared by meaning, not by encoding.** A runner drops empty props,
children and slots on both sides first. `{"type":"Box"}` and
`{"type":"Box","props":{},"children":[]}` are the same tree, and a corpus that
failed the second would be asserting a serialisation — failing a Swift client for
emitting an empty dictionary and teaching nobody anything. The rule is stated in
the corpus rather than left to each runner, because a comparison rule decided
per-language is a corpus that means something different in every language.

**Two runners, not one.** Go runs validators, predicates and bindings;
TypeScript runs all four. A corpus only one implementation executes is a test.

**`sdui.ResolveProps` and `expandOnce` exist because of this.** The Go binding
resolver was written so the binding corpus had a second implementation to run
against at all, and the expander was extracted out of the registered React
component so the expansion rules could be checked without rendering a screen.
Those rules — the ones the entire definition model rests on — were previously
reachable only by looking at output.

**Expansion is not run in Go.** This repository has no expander, and writing a
second one purely to run the corpus would be a second implementation to keep in
step: the drift the corpus exists to catch, manufactured in order to catch
itself. What Go checks is that the file stays well-formed, so it cannot rot
between the clients that do run it.

## Alternatives considered

**Assert behaviour in each repository's own tests.** *Rejected* — that is the
status quo, and it produces two test suites that pass while describing different
behaviour. The tests were not wrong; they were each right about their own
implementation.

**A single shared test runner.** *Rejected* — it would have to be written in a
language every client can host, which means either a second runtime dependency
per client or a bespoke DSL. Data plus a small runner per language is the
cheaper half of that trade and the one DivKit made.

**Byte-identical JSON comparison.** *Rejected* — see above. It asserts the
encoder, which is the part that legitimately differs.

**Generate the corpus from the implementations.** *Rejected*, firmly. A corpus
generated from an implementation asserts that the implementation does what it
does. The cases are written by hand and argued about in a diff, which is the only
form in which they mean anything.

## Consequences

- **Running the corpus against the client immediately found seven
  disagreements**, all the same encoding difference, which is what produced the
  comparison rule above. That is a corpus working on its first run.
- **59 of 66 cases passed before any correction**, which is the reassuring half:
  two implementations written weeks apart agreed about validators, predicates and
  bindings entirely.
- **`@mosaic-media/sdui-react`'s relative imports now carry explicit `.js`
  extensions.** The package was not loadable by Node — extensionless specifiers —
  which had already forced one workaround in the vocabulary check and blocked the
  corpus runner outright. Fixing it properly makes the published package
  Node-loadable, which it should always have been.
- **The corpus is published in the npm package** under
  `@mosaic-media/sdui/conformance/cases/*`, so a client in any language reads the
  same files rather than a copy.
- **A mistake worth recording**: the commit that added the comparison rule left
  the corpus as invalid JSON and was pushed anyway, because the command running
  the gate was chained to a `grep` and the push keyed off *grep's* exit status.
  The test caught it; the thing running the test did not. The gate worked and the
  harness around it did not, which is a more useful failure than the typo.
