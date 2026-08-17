# Claude Instructions — contracts

Fleet-wide conventions — commits, decision records, citation form, the roadmap —
are in [`architecture`](https://github.com/mosaic-media/architecture/blob/main/CLAUDE.md).
This file is what is specific to `contracts`.

This repository is the published Server-Driven-UI and session contract: one
language-neutral surface a producer (the Platform, a module) emits and a client
renders. `README.md` describes the layout and how to consume it. Apache-2.0, so a
third party may build UI against it under any licence.

## Everything is generated, from two files

**`schema/sdui.schema.json` and `ui.spec.json` are the sources. Hand-editing a
generated file is the bug.** `scripts/generate.sh` is the whole pipeline:

| Source | Generator | Output |
|---|---|---|
| `schema/sdui.schema.json` | quicktype | `sdui/contract/contract.gen.go`, `ts/contract.gen.ts` |
| `ui.spec.json` | `tools/genui` | `ui/components.gen.go`, `ts/ui.ts`, `sdui/vocabulary.gen.go`, `ts/vocabulary.gen.ts`, `conformance/vocabulary.json`, `REFERENCE.md` |
| `definitions/*.json` | `scripts/gen-definitions.mjs` | `ts/definitions.gen.ts` |
| `tokens/tokens.json` | `scripts/gen-tokens.mjs` | `tokens/tokens.css` |
| `proto/mosaic/**` | `buf`, two templates | `gen/`, `gen-ts/` |

`scripts/check-generated.sh` regenerates and `git diff`s exactly the paths in its
`GENERATED` array. **A new `tools/genui` output goes in two places** — `outputs`
in `tools/genui/main.go` *and* that array; one missing there is regenerated and
then never compared.

**That array is every generated file rather than the interesting ones, and the
protobuf bindings are why.** A binding is regenerated, never edited — least of
all textually: a `.pb.go` carries a length-prefixed descriptor, so a rewrite
that reads as correct in a diff corrupts it, and the failure lands at init
inside a consumer rather than in any gate here.

**The generator pins are load-bearing.** quicktype is pinned in
`scripts/generate.sh`, `buf` and the three protoc plugins in `Dockerfile.test`,
`protoc-gen-es` by `package-lock.json` through `npm ci`. An unpinned generator
restyles its own output and the guard then fails on a "stale" binding that is
only a different generator. Each pin carries a comment saying what moving it
costs — read it, then bump and regenerate in one change.

## What `go run ./tools/genui -lint` refuses

Adding a component or a primitive is an edit to `ui.spec.json`. The lint refuses:

- a `definitions/*.json` with no component in the spec, and a component with no
  definition file;
- a prop a definition's template binds that no `ui` helper sets, and an `Outlet`
  it declares that no slot helper fills;
- a template or fallback referencing a type in neither tier;
- a fallback needing a primitive its template does not, or the same set (it then
  degrades nothing);
- a type declared as both a primitive and a component;
- a primitive with no `native` justification for why a definition cannot express
  it — growing that tier is the one change costing every client a release
  ([contracts#2](docs/adr/0002-primitives-and-definitions.md));
- an action kind, tone or surface that is not identical in `ui.spec.json`,
  `schema/sdui.schema.json` and `proto/mosaic/sdui/v1/sdui.proto`;
- a core type containing the spec's `typeSeparator`, reserved for module
  namespacing.

A prop with no helper is one the emit side sets by string, and the props bag
accepts whatever it is spelled — so it draws nothing, silently. Every prop also
gets a generated `Bind*` helper for that reason: a bound value never needs
`Prop("title", …)` either.

## One prop key, one type, across every component

`value` is a field's text, so a switch's state is `on`; `meta` is a hero's
variadic line, so a card's provenance is `origin`. The helpers are typed and the
bag is not — a key that is a string in one component and a boolean in another is
read as text in the second, where every non-empty string is on. Each split is
justified in that prop's `doc` in `ui.spec.json`; write one there when adding
another.

## A component is authored only in `definitions/*.json`

One JSON file per component, and both aggregates are generated —
`definitions.Library()` for the Platform to serve, `ts/definitions.gen.ts` for a
JS consumer with no Platform to ask. Neither aggregate is a place to add one, and
neither is a client: a component written as code renders on that client alone,
which is what the definition model exists to prevent.
[`web`](https://github.com/mosaic-media/web/blob/main/CLAUDE.md) carries the same
rule from the consuming side.

## The closed sets are closed in one direction

Validators, predicates, roles and focus directions are exhaustive: the server may
name only what every client can interpret. Widening one lets the server state
something a client silently ignores, which fails open — the reasoning is at the
top of `sdui/validate.go` and `sdui/predicate.go`. Growing one is a decision with
a record, not a spec edit.

## A behaviour change is a change to `conformance/cases/*.json`

The corpus is golden inputs and outputs, run by two implementations — `go test
./conformance/...` here, and a second runner in
[`web`](https://github.com/mosaic-media/web). Changing what a validator says,
what an unreadable predicate answers, how a template expands or whose value wins
in a submit merge, without touching the corpus, means one implementation has
quietly moved.

`corpus_test.go` executes the validator, predicate and binding files and checks
only that the expansion and submit files are well-formed: this repository has no
expander and no dispatcher, so **those two are checked in a client or nowhere.**
`TestEveryCorpusFileIsAccountedFor` fails when a file is added to the directory
and to no runner.

## The gate

```bash
docker compose -f docker-compose.test.yml run --rm test
```

That runs, in order: `git rev-parse`, `check-versions.mjs`, the record index
check, the citation lint, `check-generated.sh`, `go build ./...`, `go test ./...`
and `tsc --noEmit --strict` over the generated TypeScript. Append `bash` for a
shell in the same environment, or name a step — `bash scripts/generate.sh`.

**The leading `git rev-parse` is not ceremony, and neither is the image's
`safe.directory`.** `check-versions.mjs` catches its own git failure and reports
"no tags yet — nothing to check against", exit 0; `check-generated.sh`'s whole
verdict is a `git diff`, where dubious ownership reads as a stale binding. A git
this container cannot use turns both checks into passes.

**This repository needs several toolchains in one process tree** — `go`, `gofmt`,
`node`, `npx`, `buf`, the protoc plugins and `python3` — which is why it is the
one with a `Dockerfile.test` of its own. A host with most of them does not fail
loudly; it produces a check that passes by not running.

**`.github/workflows/verify.yml` restates those steps on a bare runner rather
than running the compose service, and it installs no `buf` or protoc plugins.**
`scripts/generate.sh` skips protobuf generation when `buf` is absent, so `gen/`
and `gen-ts/` are drift-checked in the container and not in CI. Run the container
before pushing; a green tick does not cover the protobuf bindings.

## Versioning

Pre-1.0, minor bumps. `package.json`'s `version` and the git tag must agree:
`check-versions.mjs` compares it against the newest tag, `release.yml`
additionally against the tag being released. **Publishing is a pushed tag and
nothing else** — `release.yml` reuses `verify.yml` against the tagged commit and
then publishes to npm; Go needs no step, the proxy serves the tag. There is no
`npm publish` to run by hand. Consumers then bump their own dependency, and **a
`replace` (Go) or a workspace link (npm) may never land in a commit.**

## Records and vendored tooling

[`docs/adr/README.md`](docs/adr/README.md) is the generated index — read it
rather than counting files, and do not hand-edit it or restate a status here.
`scripts/adr_index.py` and `scripts/adr_lint.py` are vendored copies whose source
is `architecture/scripts/`; the gate runs them, so a stale index or an
unresolvable citation refuses a push. **Do not edit either copy here** — change
them there and re-vendor.

Every exported type carries a doc comment saying *why*: this is a published
contract read by people who cannot read the Platform's source.
