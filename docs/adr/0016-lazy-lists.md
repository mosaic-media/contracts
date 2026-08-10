# Lazy lists

**Status:** Accepted (built). The catalog-paging consequence it recorded as owed
is **half discharged**: SDK `v0.23.0` added `CatalogItemsResponse.HasMore` and
the Platform carries it through to the catalog screen, but every shipped catalog
provider is on `v0.22.0` and sets nothing, so `hasMore` is always false and the
screen does not page. Built and unreachable until a provider opts in.

**Date:** 2026-07-25

## Context

Every list in Mosaic has been the whole list. Search returned whatever the
providers gave and rendered all of it; a catalog page rendered one page and had
no way to say there was another. The vocabulary had `Pagination` — prev/next
buttons a screen draws — but nothing that says *this is a page of something
longer, and here is how to get the rest*.

## Decision

**A list declares `hasMore`, and carries the `loadMore` action that fetches the
next page.**

- **`hasMore` is the server's statement, never the client's inference.** A page
  that happens to be full is not evidence there is another one, and a client
  guessing from the count asks for a page that does not exist.
- **The Platform earns the statement by asking for one more than it renders.**
  That extra result is the only honest evidence of a further page. It is never
  rendered — it is a question, not a result.
- **`loadMore` carries whatever cursor the server needs**, because only the
  server knows what "next" means for a given list: a skip count here, a token
  elsewhere.
- **The action is `query`, not `navigate`.** They differ only in the history
  entry, and a further page is not somewhere the back button should return to. A
  viewer who scrolled through five pages should get one press back to where they
  came from, not five. This is what closes the last entry in the client's
  declared gap: `query` was the one kind still unimplemented, and lazy loading is
  what it was for.
- **The trigger is the list's end approaching**, with a 600px margin so the
  request is usually in flight before the viewer arrives. Stated rather than
  tuned per list, so two lists behave the same.
- **One request per page, guarded by the child count.** Until the list is longer
  than it was when the request went out, the page has not arrived — and asking
  again would repeat the request at whatever rate the observer fires, which is
  how a lazy list becomes a request loop.
- **The page size is the Platform's number.** A client cannot know what an
  upstream costs: every search result is a row the providers were fanned out for
  and the library was checked against, and the page size is the only lever on
  that fan-out.

## Alternatives considered

**Infer more-availability from a full page.** *Rejected* — it is wrong exactly
at the boundary, which is the only place it is ever consulted, and the symptom is
a request for a page that does not exist on every list whose length happens to be
a multiple of the page size.

**A `Pagination` control instead — prev/next buttons.** *Rejected as the
default*, not removed. It already exists and is right for a screen where a person
wants to jump about. It is wrong for a rail or a grid, where the interaction is
scrolling and a button at the bottom is a wall.

**`navigate` for the load action, with the client suppressing history.**
*Rejected* — it puts the decision at every call site and gets it wrong once. The
difference between the two is precisely the history entry, so it is a different
kind.

**Let the client decide the page size.** *Rejected* — it is the client that
knows how much fits on screen and the server that knows what a page costs, and
the cost dominates: a page here is a provider fan-out, not a database range scan.

## Consequences

- **The client's declared gap is now empty**: 26/26 primitives, 12/12 action
  kinds, and every closed set matched. Every entry left the list the right way —
  `setValue` when state landed, `submit` when forms did, `query` here, and `Form`
  by turning out not to be a primitive at all.
- **This departs from the roadmap's stated mechanism, and the departure is the
  interesting part.** S10 said "over the `APPEND` region op that already exists".
  `APPEND` appends to the *region's* top-level node list, and a screen is that
  region's single node — so an appended grid renders *after* the screen's frame
  rather than inside its existing grid. Targeting a node *within* a screen is the
  region-host plumbing the roadmap itself already lists as deferred. So the page
  arrives as a `REPLACE` of the whole screen with a longer list. The transport is
  unchanged, which was the other half of the claim; the op is not the one named.
- **A replaced screen re-renders everything above the new page.** It is correct
  and it is more work than appending; React's reconciliation keeps the scroll
  position and the already-rendered cards, so it is not visible today. It will
  matter when a list is long enough that re-rendering it costs a frame, and that
  is when the region-host plumbing earns itself.
- **Search is paged for the first time.** 24 per page. Verified live: the grid
  grew 49 → 61 as pages loaded, stopped when the extra-result probe came back
  short, and the history length did not change across any of it.
- **Nothing else uses it.** The catalog screen has a `Skip` its query already
  supports and does not page, because its provider reports no more-availability
  and the Platform cannot ask for one extra through the SDK's
  `CatalogItemsRequest`. Closing that is an additive SDK field, and it is owed.
