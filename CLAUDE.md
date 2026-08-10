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
else's repository. Ownership predicts rot. A fact about this repository stays
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
  ([architecture#22](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0022-licensing.md)).
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

## The roadmap and the decision records

These rules are identical in every Mosaic repository. They exist because the
state of the build and the reasons behind it are the two things that rot fastest
and report nothing when they do — no build fails, no test goes red.

### The roadmap is maintained, not consulted

**[`architecture`'s roadmap](https://github.com/mosaic-media/architecture/blob/main/docs/roadmap.md)
is the single record of where the build is**, for every repository. Read it
before starting work, and **update it in the same session as the change that
dates it** — not in a follow-up, which does not happen.

- **A slice that lands is marked landed, with what was left out.** "Built" with
  no qualifier is a claim that the whole slice shipped; if part of it did not,
  say which part and why in the same sentence.
- **Implementation that departs from the plan is recorded where it departed.**
  The roadmap is derived from the code, not from the intention that preceded it,
  and the surprises are the most valuable thing in it.
- **Do not restate the roadmap here.** A second copy of "what is built" in a
  `CLAUDE.md` is how the first copy goes stale unnoticed.
- **A capability with no client path is not done — it is
  [owed](https://github.com/mosaic-media/architecture/blob/main/docs/unreachable-capability.md).**
  If you delete or fail to build a client path to a working service, add its row
  to that register in the same change.

### This repository owns its decision records

**`docs/adr/` holds the decisions this repository owns**, and `docs/adr/README.md`
is the **generated** index — read it rather than listing the directory, and never
hand-edit it. Records live in the repository whose mechanism enforces them: the
SDUI decisions are here because `ui.spec.json`, `tools/genui` and the conformance
corpus are here, and a record whose rule is checked three directories away is a
record nobody's gate defends. `architecture` keeps only the genuinely
cross-cutting ones.

A new record is a file in `docs/adr/`, numbered sequentially, in kebab-case, in
the standard Context / Decision / Alternatives / Consequences form — then
regenerate the index. **The heading carries no number**: the number lives in the
filename and in the index, so a record's anchor survives being renumbered, and
there is no third copy for a renumbering to miss.

`scripts/adr_index.py` and `scripts/adr_lint.py` are **vendored copies**; the
source of truth is `architecture/scripts/`. Fix a bug there and re-vendor —
editing the copy here is the drift these scripts exist to catch.

### How a record is cited

**`contracts#7`, written as a link.** Same repository:
`[contracts#7](docs/adr/0007-components-are-authored-only-in-the-contract.md)`.
Another repository: the label plus the full GitHub URL, as
[architecture#16](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0016-published-contract-surface.md)
is written here.

**The bare form is refused by lint**, and the reason is that it fails *open*
rather than dangling. Now that every repository numbers from 1, a stale bare
citation does not 404 — it resolves, quietly, to this repository's own record of
that number, which is a different decision. No red test, and the citation still
reads as valid. That is strictly worse than a dead link.

The lint also refuses a citation that names a record its repository does not
have, and a cross-repository citation in Markdown that is not a link — a reader
in another repository has no other way to reach it. Its ceiling for the old bare
form is a **ratchet**: it is set to the count still outstanding fleet-wide and
**only ever goes down**. The number lives in `docker-compose.test.yml` and
`.github/workflows/verify.yml`. Lower it when you convert citations; never raise
it to make a change fit.

### Decision records are append-only

A record is an account of what was decided and why, at a time. It is evidence,
not documentation, and its value is that it was not edited afterwards.

- **Never rewrite a record's body to match what was built.** Not to correct it,
  not to annotate it, not to add "as built, this differs". That pattern turns a
  record into a running commentary and destroys the thing it is for.
- **State changes go in the `**Status:**` line and nowhere else.** That is where
  a record says it is built, built in part (naming the part), or superseded —
  wholly ("Superseded by contracts#19") or partly ("Partly superseded: X was
  reversed by contracts#19; the rest stands").
- **A changed decision needs a new record that supersedes it.** If the code
  deliberately does something a record decided against, that is a decision and it
  is written down as one, with its own Context / Decision / Alternatives /
  Consequences. Both records then stand: the old one keeps its reasoning, the new
  one carries the change, and each points at the other through its Status line.
- **An unbuilt decision is not a superseded one.** "We have not done this yet"
  belongs in the Status line and the roadmap. Only a genuine reversal earns a new
  record.

**If the code and a record disagree, say so rather than quietly picking one.** An
honest "this is unresolved" is worth more than a plausible reconciliation that
reads as settled.
