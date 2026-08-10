# Retiring `$value`, and the merge rule it hid

**Status:** Accepted (built). The `$value` substitution and the `SubmitField`
primitive are gone from the contract, the client and all four external module
repositories. Verified live for TMDB, which is the only one of the four linked
into the Platform binary; the other three are covered by their own tests and by
the shared corpus, not by a browser.

**Date:** 2026-07-25

## Context

[contracts#12](0012-fields-and-forms.md) built fields and forms and left one thing
behind: `SubmitField`, a primitive whose submit action carried the literal string
`$value`, which the client substituted with whatever had been typed. It was the
mechanism that predated scopes, it was documented as owed, and it survived four
separate attempts to remove it.

It survived for a structural reason that was never named. `submit` collected a
scope's values and merged them into the action's input **at the top level**, and
that is the wrong place for the one command that actually needed it. A module's
`configureModule` takes `moduleId` at the top of its input and the settings
document underneath, so a form that could only deliver values to the top could
not configure a module at all. `$value` could, because a string substitution goes
wherever the producer wrote it. So every attempt to delete the primitive ran into
a case the replacement genuinely could not express, and concluded the primitive
was still needed.

Nine call sites across four separately released module repositories were still on
it, and every one of them was a settings form.

## Decision

**`submit` carries a destination, `SubmitField` is deleted, and the merge rule is
reversed.**

- **`Submit(action, into)`.** `into` names a path in the action's input where the
  collected values land — `"settings"` for a module's `configureModule`, absent
  for the top level, which is what every existing call site wanted. This is the
  whole of what `$value` could do that scopes could not, expressed as a field
  rather than as a sentinel string a client greps for.
- **`SubmitField` leaves the primitive tier**, and the vocabulary goes to
  **3.0.0**. 26 primitives to 25. Removing a primitive is the direction that
  costs a client release, and it is taken here because the primitive's entire
  reason for existing was the gap the destination closes.
- **The scope wins over the action's input** for the names the form declares.
  This reverses [contracts#12](0012-fields-and-forms.md), which said the opposite and gave a reason that sounded
  right: a server that pinned a field is stating something a form must not
  overwrite. See below — this is the substantive part of the record.
- **Appending is the module's business, not the wire's.** Two of the nine sites
  add to a list — an addon, a custom catalog — and a merge sets a key; it cannot
  append. Rather than give the wire a list operation, the module carries a
  *pending-addition* field: `addAddon`, `addCatalogName`/`Query`/`Type`. The form
  writes one, the module folds it into the list on read and stores it back empty.
  The contract gains nothing and the module interprets its own settings, which it
  was already doing.
- **The merge is a corpus file.** Ten cases in `conformance/cases/submit.json`,
  and `mergeSubmit` is its own module so something can call it.

## The reversal, which is the part worth reading

[contracts#12](0012-fields-and-forms.md)'s rule was that collected values merge *under* the action's input. The
reasoning was that a producer who set a field explicitly meant it.

What that missed is that a module rewriting its settings has no way to set
*some* fields. `configureModule` replaces the whole document, so the action must
carry every field — including the one the form is editing. And it sends that one
**blank**, deliberately, because a form that rendered the stored value back would
make every save look like a no-op and would echo a stored API key into the page.

So the producer's explicit value, for exactly the field being edited, was always
the empty string. Under the old rule the empty string won. The form validated,
the `Invoke` went out, the Platform stored the document, and it answered 200. The
typed value was discarded somewhere between the button and the wire, and nothing
— no client error, no server error, no failing test, no log line — said so.

It was found by typing `GB` into the TMDB region field and noticing the badge
still read "Not set".

The corrected rule is that the scope wins for the names the form declares, and
the action supplies everything else. Pinning still works and is now explicit:
a producer that wants a field fixed leaves it out of the form's scope, where a
reader can see the intent rather than infer it from a merge order.

## Alternatives considered

**Keep `$value` and add the destination beside it.** *Rejected* — the whole
argument for the primitive was the gap, and keeping it once the gap is closed
leaves two ways to do the same thing, one of which is a sentinel string.

**Give the wire a list-append operation** so the add-addon form could append
directly. *Rejected* — it is the first instalment of an expression language, which
this thread has refused at every step. The pending-addition field is uglier in
the module and costs the contract nothing, which is the correct place for the
ugliness.

**Make the destination a general path** (`settings.catalogs[0].name`) rather than
a single key. *Rejected for now* — a path syntax is a small language, and one key
covers every call site that exists. Widening it later is additive.

**Leave the merge rule alone and have modules send only changed fields.**
*Rejected* — it is not a decision this contract can take. `configureModule` is
whole-document by the SDK's design, and a rule that requires every module author
to know that is a rule that will be got wrong.

**Fix the merge and skip the corpus case.** *Rejected*, and this is the one that
would have been tempting. The rule had been wrong since forms landed, and the
reason nobody caught it is that it lived as an expression inside a dispatcher,
reachable only by filling in a form. Fixing it without making it callable would
leave the next such rule exactly as unreachable.

## Consequences

- **The vocabulary is 3.0.0, 25 primitives**, and the declared gap is still
  empty: 25/25 primitives and 12/12 action kinds implemented.
- **The corpus is 76 cases across five files**, up from 66 across four. The new
  file is the first one added because a bug got through, rather than to describe
  behaviour already believed correct.
- **Five releases**, in order: contracts `v0.32.0` and `v0.33.0`, sdui-react
  `v0.15.0`, module-aiostreams `v0.5.0`, module-fanart-tv `v0.5.0`, module-tmdb
  `v0.8.0`, module-stremio-addons `v0.26.0`, and a Platform bump onto the last of
  them. The tmdb bump is load-bearing: a Platform still pinned to `v0.7.0` serves
  a settings screen containing a primitive the client no longer has.
- **TMDB's custom catalogs are two fields rather than one.** The old form took
  `name|query` in a single box and split on the pipe, which was a parser standing
  in for a second field. The old strings are still read.
- **Three of the four modules were not verified in a browser**, because
  aiostreams, fanart-tv and stremio-addons are not linked into the Platform
  binary this repository builds. Their tests pass and their forms are the same
  shape as TMDB's, which is evidence and not proof. Naming it because the thread
  has already produced one claim that a thing worked when the screen had never
  been opened.
- **What this does not close**: no Platform command produces a field-level
  rejection, so `submit`'s error path is still only exercised by client-side
  validators. That was true before this slice and remains true.
