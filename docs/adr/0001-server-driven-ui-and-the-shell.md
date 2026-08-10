# Server-Driven UI and the Shell

**Status:** Accepted; built. Two particulars have moved since — the Shell is a package in the `web` workspace rather than its own repository ([architecture#42](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0042-frontend-workspace.md)), and it reaches the Platform over Connect, not GraphQL, which was deleted outright ([architecture#61](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0061-one-client-transport.md)). Server-Driven UI itself is unchanged
**Date:** 2026-07-20

## Context

Everything built to date is reachable only through the GraphQL API. There is no human-facing surface — the Platform has never had one.

Two facts shape how to add one. First, Mosaic targets **more than one client**: a web client, and a native client (Flutter) for TV, desktop and mobile. Second, the Platform and its optional modules already own what content exists and what operations are possible — a module can introduce new content and new actions ([architecture#19](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0019-module-capability-and-invocation.md)–[architecture#21](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0021-module-settings.md)), and a client should not need a release to surface them.

Building each client as an independent application over the GraphQL API would ship navigation, screen composition and product logic once per client. The clients would drift, and every new module capability would require a change in each of them.

## Decision

**The Platform describes the interface; each client renders it.** Server-Driven UI (SDUI).

- The server sends a tree of typed nodes — an open **component vocabulary** (`UINode`: a `type`, `props`, `children`, named `slots`). Each client maps a node's `type` to its own component and renders the tree. It never receives markup or code, only a description.
- Behaviour travels as declarative **Action envelopes** — `navigate`, `invoke` (a mutation), `query`, `openOverlay`, `playPart`, `toast`, `sequence`, and so on. Actions are **data, not executable code**; the client interprets them.
- The vocabulary is **open**, the same stance [architecture#15](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0015-open-and-closed-vocabularies.md) takes for content types: a node type a client does not recognise renders a labelled placeholder rather than failing, so an older client tolerates a newer server or module.
- Errors surface through the Platform's seven fixed categories (`InvalidArgument`…`Internal`), so feedback is uniform regardless of client or transport.
- The contract is **technology-agnostic**. The node/action schema and the design tokens are the shared surface; a web client and a Flutter client implement the *same* contract. It is deliberately the intersection of what those render stacks can both express (see [contracts#2](0002-primitives-and-definitions.md)).

The first client is **[`mosaic-shell`](https://github.com/mosaic-media/mosaic-shell)** — its own first-party repository, React + TypeScript + Vite. It is a *client of the Platform over GraphQL*, not a Module (it is not compiled into the binary) and not part of the Platform. It is licensed **AGPL-3.0-only** as a first-party client ([architecture#22](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0022-licensing.md)).

The SDUI contract and the tokens are **not yet extracted**. They live in the Shell today and will be lifted to a neutral home when the second client (Flutter) makes them genuinely shared — the same *extract on the second consumer* discipline the SDK followed ([architecture#16](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0016-published-contract-surface.md)). Until then the HTML component implementations stay in the Shell, because they are a web binding no other client can reuse.

## Alternatives considered

**Server-rendered HTML (Go templates / HTMX).** Couples the interface to one render stack; a native Flutter client cannot consume HTML fragments. *Rejected* — it forecloses the multi-client goal.

**Independent clients over the GraphQL API.** Each reimplements navigation and screen composition; they drift, and a new module capability needs a matching change in every client. *Rejected.*

**A shared component library across clients.** Cannot cross the web/Flutter technology boundary *as code*. *Rejected* — the shared thing has to be data (the contract), not components.

## Consequences

- New content and module operations surface without a client release; the server composes screens.
- Two things become the real cross-client contract: the node/action schema and the design tokens. They must stay technology-neutral — no web-isms — which constrains what the vocabulary may contain. That constraint is the subject of [contracts#2](0002-primitives-and-definitions.md).
- The Platform must grow an SDUI surface (screens and their queries) it does not yet expose. Until it does, the Shell runs on mock SDUI payloads of exactly the shape the Platform will send.

## Implementation

`mosaic-shell` implements the runtime: a registry (`type` → component), a recursive renderer, an interpreter for every Action kind, and a token-driven skin the future Mosaic Design Language will replace by swapping tokens. Its GraphQL client normalises failures to the Platform's error categories. It renders the mock screens today. How its components are structured — the primitive/definition split that keeps the contract technology-agnostic — is [contracts#2](0002-primitives-and-definitions.md).
