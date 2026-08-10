# The SDUI contract repository

**Status:** Accepted; the extraction stands. The repository has been renamed twice since — `mosaic-sdui` → `sdui` ([architecture#43](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0043-repository-naming-convention.md)) → `contracts` ([contracts#6](0006-contracts-protobuf-workspace.md), which also made the wire protobuf). The contract itself did not move with it: `mosaic.sdui.v1`, the `sdui` package and the npm name `@mosaic-media/sdui` are unchanged
**Date:** 2026-07-20

## Context

[contracts#1](0001-server-driven-ui-and-the-shell.md) said the SDUI contract and tokens would extract to a neutral home "when the second consumer makes them shared." That consumer arrived immediately, and in Go rather than later in Dart: to serve a real interface the Platform must **emit** screens, and a Module ([`mosaic-module-stremio`](https://github.com/mosaic-media/mosaic-module-stremio)) emits its own — two Go producers — while the Shell consumes the same contract in TypeScript. Hand-writing the contract in the Shell and copying it into the Platform would duplicate a moving target across languages, which is exactly what a shared contract exists to prevent.

So the contract must be extracted now, and a format chosen for it.

## Decision

Extract the SDUI contract into its own repository, **[`mosaic-sdui`](https://github.com/mosaic-media/mosaic-sdui)** — the interface counterpart of [`mosaic-sdk`](https://github.com/mosaic-media/mosaic-sdk). It holds three things:

- **The schema** — `UINode` (the open node tree), the `Action` envelope, and `ComponentDefinition`, as **one JSON Schema file that is the single source of truth**. The per-language bindings are **generated** from it (Go for producers, TypeScript for the Shell, Dart later) — never hand-written, so the contract cannot drift across languages. A drift guard fails on stale bindings, and conformance tests validate what the producer helpers emit against the schema.
- **The standard definition library** — `PosterCard`, `HeroBanner`, `Section`, … as `ComponentDefinition` *data*, not per-client code. A client registers them; a producer emits the type; every client renders it identically, the Module shipping **no** UI code. Only the primitives stay per-client native ([contracts#2](0002-primitives-and-definitions.md)).
- **The design tokens** — in W3C DTCG format, compiled to the Shell's CSS variables and (later) a Flutter theme.

**The contract is JSON Schema, not protobuf.** Three properties of SDUI decide it: the node vocabulary is *open* and its props are an untyped bag by design; the payload rides GraphQL as *JSON*; and the definition library and tokens are themselves *JSON data*. Protobuf's typed-message value collapses to `google.protobuf.Struct` exactly where the data lives, and it mismatches the JSON/GraphQL transport. JSON Schema fits all three and still generates types per language.

Licensed **Apache-2.0**, like the SDK — a contract surface must be permissive so a Module may build its UI against it under any licence (this extends [architecture#22](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0022-licensing.md)'s list to the SDUI contract).

## Alternatives considered

**Protobuf / gRPC.** Suits a *governed, typed* vocabulary — every component a registered message — delivered over gRPC. Coherent and arguably better-governed, but a different design: it closes the open vocabulary and changes the transport. *Rejected for now* as a larger change than the contract needs; recorded as the open question below.

**Keep the contract hand-written in the Shell and copy it into the Go producers.** Duplicates a moving target across three languages. *Rejected.*

**Fold the Go SDUI types into `mosaic-sdk`.** The SDK is a Go module; the SDUI contract is consumed by TypeScript and Dart as well, so it cannot live there. *Rejected* — its own polyglot repository.

## Consequences

- The Platform's emit-side and any UI-contributing Module build against one typed Go binding; the Shell — and a Dart client later — render against the same schema and the same definitions.
- A new shared surface to version. Until it is tagged, producers wire it with a `replace` directive, as the SDK did in local development.
- **Open question — the governed-vocabulary fork.** Whether the component vocabulary stays *open* (props an untyped bag, an unknown type rendering a placeholder) or becomes *governed* (each component a typed, registered contract). The `media_types` registry deferred in [architecture#15](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0015-open-and-closed-vocabularies.md) and the still-undecided module-manifest shape point toward governance; if that direction is taken, protobuf becomes the natural format and this decision is revisited. For now the vocabulary is open and the contract is JSON.

## Implementation

`mosaic-sdui` carries `schema/sdui.schema.json` (the source of truth), generated bindings (`sdui/contract/` Go, `ts/contract.gen.ts` TypeScript) with a hand-written ergonomic layer over the Go types (aliases, `Action` constructors, standard-component builders), `definitions/` (the library as data, seeded from the Shell), and `tokens/` (DTCG). `scripts/generate.sh` regenerates via quicktype; `scripts/check-generated.sh` is the drift guard; conformance tests hold the hand-written layer to the schema. The Go module builds and tests. The next slice wires the Shell to consume the generated bindings, definitions and tokens — retiring its local copies — and builds the Platform's emit-side against the Go binding.
