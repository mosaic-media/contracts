# Claude Instructions — Mosaic contracts

This repository is the **published Server-Driven-UI and session contract** shared
by the Platform, the Modules and every client: one language-neutral surface a
producer emits and a client renders, so nobody rewrites it per language.

`README.md` describes the layout and how to use it. This file is how to *work*
here — the rules, and the reasons they exist.

## What this file may say

**A `CLAUDE.md` states rules, and facts about its own repository. It does not
state facts about another one — it links instead.** That is not tidiness. An
audit of all twelve of these files against their source found 74 stale claims:
none of roughly 180 rules was wrong, and 62 of the 74 were facts about somebody
else's repository. Ownership predicts nothing. A fact about this repository stays
true because the person who changes the code changes the sentence in the same
session; a fact about another one dies the moment they edit it and nothing here
goes red.

The same rule applies to facts this repository already publishes in a generated
artefact — counts, versions, what is built. Point at the artefact. A second copy
is a copy, and copies drift.

## This is generated. Hand-editing a binding is the bug.

**The schema is the contract; the bindings are output.** A change that is not
expressible in `schema/sdui.schema.json` or in `proto/**/*.proto` is not a change
to this repository. Change the source, regenerate, commit both.

The drift guard (`scripts/check-generated.sh`) regenerates everything and fails
if a committed file moved, and the conformance tests validate what the authoring
layer produces against the schema — so even the ergonomic layer cannot drift from
the contract. `gen/` and `gen-ts/` were outside the guard until a corrupted
`.pb.go` reached the Platform and failed at init: a textual rewrite of the module
path had mangled a length-prefixed descriptor, and every gate here passed. They
are covered now, and that is why the list in `check-generated.sh` is every
generated file rather than the interesting ones.

**The generators are pinned, and the pins are load-bearing.** An unpinned
generator restyles its own output, and the guard then fails on a "stale" binding
that is only a different generator. `scripts/generate.sh` and `Dockerfile.test`
carry the pins and each says what moving it costs — read that comment before
bumping one, and regenerate and review the diff in the same change.

**The other published contract repository is built the opposite way.** If you are
about to apply a rule you remember from [`sdk`](https://github.com/mosaic-media/sdk),
read that repository's own instructions and its own records first; this file
cannot tell you its state and must not try. One consequence is worth stating from
this side, because it has been argued the wrong way: **this repository's `go.mod`
requires protobuf, gRPC and Connect because generated bindings do not run without
their runtime.** That is the generator's output, not an implementation this
contract chooses on anyone's behalf. Do not strip it in another repository's
name.

## Non-negotiable rules

- **Apache-2.0**, the permissive surface a third party builds against, unlike the
  Platform's AGPL and the web client's
  ([architecture#1](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0001-licensing.md)).
- **A component in the vocabulary is a *primitive* or a *definition***
  ([contracts#2](docs/adr/0002-primitives-and-definitions.md)). Growing the
  primitive set is the only thing that requires a client release, so it is a
  decision, not a convenience — prefer a definition composed from what exists.
- **`ui.spec.json` is the vocabulary and all three tiers are generated from it**
  ([contracts#8](docs/adr/0008-one-generated-sdui-vocabulary.md)). Primitives,
  components, actions, tones, surfaces, validators and predicates, roles and
  focus directions are declared there once; `go run ./tools/genui` emits the Go
  and TS constructors, the registry in each language, the client conformance
  fixture and the published `REFERENCE.md` (the list is `outputs` in
  `tools/genui/main.go`). Every one is drift-guarded; none is hand-edited.
  **A primitive is added here, not in a client** — that tier drifted for the whole
  life of the project because
  [contracts#7](docs/adr/0007-components-are-authored-only-in-the-contract.md)'s
  rule only ever covered components. Every primitive states why it cannot be a
  definition, and lint refuses one that does not.
- **The spec, `schema/sdui.schema.json` and `proto/mosaic/sdui/v1/sdui.proto`
  must agree**, and `go run ./tools/genui -lint` fails when they do not. Adding
  an action kind means adding it in all three; that is the point.
- **A change to behaviour is a change to `conformance/cases/*.json`**
  ([contracts#17](docs/adr/0017-the-conformance-corpus.md)). The corpus is golden
  inputs and outputs for the validators, the predicates, binding resolution and
  definition expansion, and it is executed by *two* implementations: `go test
  ./conformance/...` here, and a second one in
  [`web`](https://github.com/mosaic-media/web). Changing what a validator says,
  what an unreadable predicate answers or how a template expands without touching
  the corpus means one implementation has quietly moved. Add the case in the same
  change; a corpus only one implementation executes is a test.
- **The closed sets are closed, and each is closed in the same direction**: the
  server may name only what every client can interpret. Validators, predicates,
  roles and focus directions. Widening one lets the server state something a
  client silently ignores, which fails open — the failure the whole vocabulary is
  arranged to prevent. Growing one is a decision with a record, not a spec edit.
- **Do not add a component because one screen wants it.** The Platform's
  emit-side is a consumer of this contract, not its owner.
- **`definitions/*.json` is the only place a component is authored. This is a
  hard rule and it has been broken before.** Every composition — `PosterCard`,
  `Section`, `AppShell`, `SettingsFrame` — is data here, one file per component,
  and the aggregates are generated: `definitions.Library()` for the Platform to
  serve, `ts/definitions.gen.ts` for a JS consumer with no Platform to ask.
  Neither aggregate is a place to add one.

  The rule exists because it was violated for most of this repository's life: ~30
  components lived as hand-written TypeScript inside the React client, the
  Platform served a dump of *that*, and this directory held four stale copies of
  components the client had since changed — three of the four had drifted, and
  nothing anywhere reported it. A component written in a client renders on that
  client and nowhere else, which is precisely what the definition model exists to
  prevent. If you are looking at a `.tsx` file and about to describe a layout in
  it, stop: the file you want is here.
- **The authoring layer must keep up.** `ui.spec.json` exposes a helper for every
  prop a definition binds, and `-lint` fails when it does not — so a definition
  cannot be added without the Go and TS builders to author it. A prop with no
  helper is a prop the emit-side sets by string, which is how `ui.Subtitle` came
  to be set on a `Stack` that renders no subtitle, silently, for the whole life
  of a screen.
- **One prop key, one type, across every component.** `value` is a field's text,
  so a switch's state is `on`; `meta` is a hero's variadic line, so a card's
  provenance is `origin`. The generated helpers are typed, and a key that is a
  string in one component and a boolean in another cannot be.

## Everything runs in the container, nothing runs on the host

**Do not run `go build`, `go test`, `scripts/generate.sh`, `npx`, `tsc` or the
record scripts directly on this machine.** This repository's gates run inside its
test container:

```bash
docker compose -f docker-compose.test.yml run --rm test
```

That runs the record-index check, the citation lint, the npm-version check, the
drift guard, the Go build and tests, and the TypeScript typecheck. Append `bash`
for a shell in the same environment. Regenerating is the same command with the
script named:

```bash
docker compose -f docker-compose.test.yml run --rm test bash scripts/generate.sh
```

`docker-compose.test.yml` is what you run and `.github/workflows/verify.yml` is
what refuses the push. **Keep the two in step** — a gate in one and not the other
is a gate that passes locally and fails on push, or worse, the reverse.

**This repository needs the container more than most, because it is the one that
needs several toolchains at once.** The drift guard regenerates Go *and*
TypeScript *and* the protobuf bindings, so one command needs `go`, `gofmt`,
`node`, `npx`, `buf` and now `python3` together. A host with most of them does not
fail loudly; it produces a check that passes by not running. Two specific ways,
both real:

- **A different generator version rewrites the bindings**, and the diff reads as
  a stale binding rather than as a different generator. `Dockerfile.test` pins
  what produces them.
- **`scripts/check-versions.mjs` catches its own git failure** and reports "no
  tags yet — nothing to check against", exit 0. Without a working git it passes
  by finding nothing, which is why the container's command runs `git rev-parse`
  first and why the image configures `safe.directory` for the bind mount.

## Versioning and release

Pre-1.0 on purpose. A change is a **minor** bump, with the npm `version` and the
git tag kept identical — CI checks both that `package.json` agrees with the
newest tag and, on the tag path, that it agrees with the tag being released, so a
mismatch fails rather than shipping a package that lies about itself.

**Publishing is a pushed tag and nothing else.** `.github/workflows/release.yml`
fires on `push: tags: ["v*"]`, re-runs the full gate against the tagged commit,
and publishes. There is no `npm publish` to run by hand and no registry token on
any developer machine; asking for one is asking for the wrong thing.

**Tag pushes are currently refused from this environment** — GitHub answers 403
on a tag ref while accepting branch pushes. Branch work lands normally; a release
does not. Do not describe tagging as a step you have completed, and do not tell
anyone a version is published when only its commit is on `main`. A Go consumer
can be moved onto an untagged commit as an ordinary pseudo-version `require` from
the proxy, which keeps the standing rule that **a `replace` must never land in a
commit** intact; an npm consumer has no equivalent and simply waits.

Consumers then bump: the Platform's `go.mod` require, and the `web` workspace's
`@mosaic-media/sdui` dependency. For local cross-repo work use a `replace` (Go)
or the workspace link (npm) temporarily — **neither may land in a commit.**

## Workflow

- Commit and push this repository **separately** from `platform` and `web`.
- **Commit author identity** must be `AdamNi-7080 <anicholls41@gmail.com>`.
- The test container green before pushing — it is the drift guard, the record
  checks and the conformance corpus as well as the build.
- Every exported type carries a doc comment saying *why*, not only what. This is
  a published contract read by people who cannot read the Platform's source.

<!-- shared-rules:begin -->
## Rules every Mosaic repository shares

*Generated. The source is `architecture/shared/repository-rules.md`; edit it there
and run `scripts/shared_rules.py --write` across the fleet. A copy edited in place
fails its repository's gate, which is the point: these rules were eleven
hand-kept copies in four variants, and the abridged ones had quietly dropped the
reasoning while keeping the rules — and in one case dropped a rule outright.*

### What this file may say

**A `CLAUDE.md` states rules, and facts about its own repository. It does not
state facts about another one — it links instead.**

An audit of all twelve of these files against their source found 74 stale claims.
None of roughly 180 rules was wrong; 62 of the 74 were facts about somebody
else's repository. Ownership predicts rot: a fact about this repository stays true
because whoever changes the code changes the sentence in the same session, and a
fact about another one dies the moment they edit it with nothing here going red.

The same applies to facts this repository already publishes in a generated
artefact — counts, versions, what is built. Point at the artefact.

### Decision records live with the code they govern

Each repository owns the records whose *mechanism* it holds — the spec file, the
lint gate, the conformance corpus, the composition root, the release workflow.
A decision can bind five repositories and still have exactly one steward.

- **`docs/adr/`**, numbered from 1 in every repository, with `docs/adr/README.md`
  a **generated** index. Read the index first; it is the bounded thing.
- **A record's heading carries no number.** The number lives in the filename and
  the index only, so a record's anchor survives being renumbered.
- **Cite a record as `repo#N`, and make it a link** — a relative path within a
  repository, an absolute URL across them, and the bare label only where no URL
  is possible, such as a code comment or a Dockerfile. The old `ADR NNNN`
  spelling is refused by a lint: once every repository numbers from 1, that form
  resolves quietly to a *different* record instead of dangling, and no tool in
  the fleet could detect it.
- **Cross-cutting records stay in [`architecture`](https://github.com/mosaic-media/architecture)** —
  the ones with no enforcing mechanism anywhere: licensing, repository naming and
  topology, the module tier model.

### Decision records are append-only

An ADR is an account of what was decided and why, at a time. It is evidence, not
documentation, and its value is that it was not edited afterwards.

- **Never rewrite a record's body** — not to correct it, not to annotate it, not
  to add "as built, this differs". That turns a record into a running commentary
  and destroys the thing it is for.
- **State changes go in the `**Status:**` line and nowhere else** — built, built
  in part (naming the part), or superseded, wholly or partly.
- **A changed decision earns a new record that supersedes it**, with its own
  Context / Decision / Alternatives / Consequences, and both records then point
  at each other through their Status lines. The old body stays exactly as it was.
- **An unbuilt decision is not a superseded one.** "Not done yet" belongs in the
  Status line and the roadmap; only a reversal earns a new record.

### The roadmap is maintained, not consulted

**`docs/roadmap.md` in [`architecture`](https://github.com/mosaic-media/architecture)
is the single record of where the build is, across every repository.** It stays
there because a milestone spans repositories by construction. Read it before
starting, and **update it in the same session as the change that dates it** — not
in a follow-up, which does not happen.

- A slice that lands is marked landed, **with what it left out named in the same
  sentence**. "Built" with no qualifier claims the whole slice shipped.
- Implementation that departed from its record is recorded where it departed.
  The surprises are the most valuable thing in it.
- **Do not restate the roadmap here.** A second copy of "what is built" in a
  `CLAUDE.md` is how the first copy goes stale unnoticed.
- A capability with no client path is not done — it is
  [owed](https://github.com/mosaic-media/architecture/blob/main/docs/unreachable-capability.md).

### Demonstrated, not asserted

**Say what you actually ran.** A skipped test is not a passed test, and "it should
work" is not evidence.

Each repository's container is the authority on its own gate, and the command is
in that repository's section below. It exists because the checks that matter fail
*soft*: a missing PostgreSQL skips storage tests and still prints `ok`, a missing
generator toolchain produces a drift guard that passes by not running. Where the
container cannot be run, running what you can on the host is better than running
nothing — **provided you report which checks ran and which did not.** Claiming a
gate passed when it was not executed is the one thing this rule exists to stop.

### Commit and push

- **Commit and push each repository separately.** They are siblings on disk and
  independent in git.
- **Commit author identity** must be `AdamNi-7080 <anicholls41@gmail.com>`. If git
  has no identity configured, set it repo-locally rather than globally.
- **Push once the change has been demonstrated working in this session.** Commit
  locally and say so otherwise. **Force-push always requires asking.**
<!-- shared-rules:end -->
