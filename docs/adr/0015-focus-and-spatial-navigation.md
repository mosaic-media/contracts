# Focus and spatial navigation

**Status:** Accepted (built)

**Date:** 2026-07-25

## Context

[contracts#14](0014-accessibility-in-the-contract.md) put roles, names and heading
levels in the contract and said plainly that it was half the job. This is the
other half, and it is the half a keyboard user needs most.

A web client can nearly get away without a focus model. The browser has a tab
order and a focus ring, and a keyboard user gets *something* — usually wrong, but
something. **A television cannot.** There is no tab key on a remote; there is a
direction pad, and every press has to resolve to exactly one node or the viewer
is stuck on a screen with no way off it.

That is why this slice exists at all. It is the precondition for the second
client the transport was chosen for ([contracts#5](0005-cross-client-transport-two-lane-rpc.md)),
and it is the item on the list that genuinely cannot be added afterwards: a
vocabulary with no focus model has every client inventing one, and no two of them
agreeing about which node a press should land on.

## Decision

**The mechanism is the client's. The four things geometry cannot know are the
contract's.**

Given a focused node and a direction, finding the nearest focusable node that way
is geometry, and it depends on where things actually ended up on screen — which
only the renderer knows. What the contract carries is:

- **`focusable`** — which nodes are worth landing on. Stated, not inferred from
  interactivity: a card that navigates is focusable, and so is a heading a viewer
  needs to land on to read, and only the emit-side knows which.
- **`focusGroup`** — which sets behave as one stop. A rail of forty cards is
  **one** stop in the tab order, not forty; a keyboard user pressing tab should
  pass the rail, not traverse it. Which of the forty is the stop is the one they
  last used, which is the roving tabindex pattern.
- **`initialFocus`** — where focus starts. At most one wins, and it is the first
  to ask: a screen naming two is stating a preference it does not have, and
  moving focus twice on arrival is worse than honouring either.
- **`nextFocus`** — the specific cases where the geometric answer is wrong. An
  override, not the mechanism.

**Four directions, closed, no diagonals.** No remote has them, and declaring them
would make every client implement a geometry nobody presses. `NextFocus` refuses
an undeclared direction rather than dropping it, because a nextFocus naming
something nothing resolves is a dead end — and *"focus went nowhere when I
pressed right"* is the least reportable bug there is, since the person hitting it
cannot get anywhere to report it from.

**The geometry weights the primary axis.** Nearest is centre-to-centre among
elements actually *in* that direction, with the across-axis distance multiplied
by four. Without the weighting, pressing right in a grid jumps diagonally to
whatever is nearest in a straight line, which reads as focus moving at random.
The number is arbitrary and stated: high enough that a neighbour in the same row
always beats one a row away, low enough that a slightly offset item is still
reachable.

**An override naming a node that is not on screen falls through to the
geometry** rather than trapping focus. A dead end on a remote is the one outcome
with no way out of it.

## Alternatives considered

**Rely on the browser's tab order and DOM order.** *Rejected* — it is not a
model, it is an accident of markup, and it does not exist at all on the platform
that needs one. It also cannot express "this rail is one stop", which is the
single most common thing a media interface needs to say.

**Put the geometry in the contract — ship coordinates or an explicit graph.**
*Rejected* — the server does not know where anything ended up. A layout that
depends on viewport width (which `style.responsive` explicitly supports) has no
server-side geometry to ship, and an explicit graph would have to be regenerated
whenever anything moved.

**Make every interactive node focusable automatically.** *Rejected* — it is
right often enough to be tempting and wrong in the cases that matter. A decorative
`Pressable` wrapping a whole card region becomes a stop nobody wants, and a
heading that a TV viewer must land on to read never becomes one.

**Allow diagonals for completeness.** *Rejected* — completeness against what? No
input device produces them, so every client would implement a geometry that is
never exercised, which is how a feature becomes wrong without anyone noticing.

## Consequences

- **Two silent no-ops were found by the browser and by nothing else**, and both
  are the same shape as the `display: contents` bug in
  [web#6](https://github.com/mosaic-media/web/blob/main/docs/adr/0006-lifecycle-triggers-and-the-absent-telemetry-lane.md):
  1. The focus host was resolved **during render**, before React attached the
     ref, so every effect saw `null` and never re-ran. The hook compiled, the
     build was green, and nothing was ever focusable.
  2. A `nextFocus` override resolved to the wrapper carrying the node id — which
     is `display: contents` and cannot take focus — so the override was a no-op
     that left the viewer exactly where they were. **The dead end the override
     exists to prevent, produced by the override itself.**

  Three of these in three slices is a pattern worth naming: a wrapper that
  deliberately has no layout presence is invisible to every browser API that
  works on boxes, and every use of it needs the same correction.
- **Verified live**: initial focus landed on the declared node; arrow keys walked
  the group geometrically; the `nextFocus` override on the third member jumped
  back to the first rather than continuing to the fourth; and a group of four
  had exactly **one** tab stop.
- **Nothing emits a focus prop.** Same as roles and triggers: the vocabulary is
  ahead of the emit-side by design, because this is the retrofit-hostile end of
  the list.
- **The web is the wrong platform to prove this on, and that is the point.** It
  is the one where the model matters least, so implementing it here is what
  turns the contract's focus model into something other than a guess about a
  client nobody has written yet.
