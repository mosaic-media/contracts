# Claude Instructions — Mosaic SDUI

This repository is the **published Server-Driven-UI and session contract** shared
by the Platform, the Modules and every client. It is to the interface what
[`sdk`](https://github.com/mosaic-media/sdk) is to content: one language-neutral
surface a producer emits and a client renders, so nobody rewrites it per
language.

## This is generated. The SDK is not.

**Read this before adding a file.** Mosaic has two published contract
repositories and they are built in opposite ways, which is a reasonable thing to
get wrong:

| Repository | Form | Source of truth |
|---|---|---|
| **`sdui`** (this one) | **protobuf and JSON Schema, Go and TS generated** | `proto/**/*.proto` and `schema/sdui.schema.json` |
| **`sdk`** | **hand-written Go** | the `.go` files in `contracts/platform/v1/` |

[ADR 0044](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0044-contracts-protobuf-workspace.md)
made the SDUI and session contracts protobuf, and encoding option (b) landed —
the typed `mosaic.sdui.v1.UINode` rides the session envelope directly, with no
JSON step. The reason it is generated here and hand-written there: this is a wire
format consumed by four client languages, where codegen is exactly right, whereas
the SDK's job is Go interfaces a third party *implements* in its own process,
which protobuf cannot express.

**Never hand-edit a generated file.** Change the schema or the `.proto`,
regenerate, commit both. The drift guard (`scripts/check-generated.sh`)
regenerates and fails if the committed bindings are stale, and the conformance
tests validate what the authoring layer produces against the schema — so even the
ergonomic layer cannot drift from the contract.

**A known toolchain wrinkle:** the pinned quicktype emits `ActionKind` as a type
union rather than an enum, while `ts/ui.ts` uses `ActionKind.PlayPart` as a
*value*. Taking that regeneration breaks the `ui` layer. `git checkout --
ts/contract.gen.ts` after generating, and fix it properly rather than carrying
the workaround silently.

## Non-negotiable rules

- **The schema is the contract; the bindings are output.** A change that is not
  expressible in the schema is not a change to this repository.
- **Apache-2.0**, like the SDK — this is the permissive surface a third party
  builds against, unlike the Platform's AGPL and the web client's.
- **A component in the vocabulary is a *primitive* or a *definition*** (ADR 0024).
  Growing the primitive set is the only thing that requires a client release, so
  it is a decision, not a convenience — prefer a definition composed from what
  exists.
- **`ui.spec.json` is the vocabulary and all three tiers are generated from it**
  (ADR 0083). Primitives, components, actions, tones, surfaces, validators and
  predicates, roles and focus directions are declared there once; `go run
  ./tools/genui` emits **six** artefacts — the Go and TS constructors, the
  registry in each language, the client conformance fixture
  (`conformance/vocabulary.json`) and the published `REFERENCE.md`. All six are
  drift-guarded; none is hand-edited. **A primitive is added here, not in a
  client** — that tier drifted for the whole
  life of the project because ADR 0082's rule only ever covered components, and
  the lint gates exist to keep it from happening again. Every primitive states
  why it cannot be a definition, and lint refuses one that does not.
- **The spec, `schema/sdui.schema.json` and `proto/mosaic/sdui/v1/sdui.proto`
  must agree**, and `genui -lint` fails when they do not. Adding an action kind
  means adding it in all three; that is the point.
- **A change to behaviour is a change to `conformance/cases/*.json`** (ADR 0094).
  The corpus is golden inputs and outputs for the validators, the predicates,
  binding resolution and definition expansion, and it is run by *two*
  implementations — `go test ./conformance/...` here, and the client's own
  `scripts/check-conformance.mjs`. Changing what a validator says, what an
  unreadable predicate answers or how a template expands without touching the
  corpus means one implementation has quietly moved. Add the case in the same
  change; a corpus only one implementation executes is a test.
- **The closed sets are closed, and each is closed in the same direction**: the
  server may name only what every client can interpret. Validators, predicates,
  roles and focus directions. Widening one lets the server state something a
  client silently ignores, which fails open — the failure the whole vocabulary
  is arranged to prevent. Growing one is a decision with an ADR, not a spec
  edit.
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
  prop a definition binds, and `go run ./tools/genui -lint` fails when it does
  not — so a definition cannot be added without the Go and TS builders to author
  it. A prop with no helper is a prop the emit-side sets by string, which is how
  `ui.Subtitle` came to be set on a `Stack` that renders no subtitle, silently,
  for the whole life of a screen.
- **One prop key, one type, across every component.** `value` is a field's text,
  so a switch's state is `on`; `meta` is a hero's variadic line, so a card's
  provenance is `origin`. The generated helpers are typed, and a key that is a
  string in one component and a boolean in another cannot be.

## Versioning and release

Pre-1.0 on purpose. A change is a **minor** bump, tagged and pushed, with the
npm `version` and the git tag kept identical — CI checks that they match, so a
mismatch fails rather than shipping a package that lies about itself.

```bash
git tag v0.10.0 && git push origin main && git push origin v0.10.0
```

**Publishing to npm is not a thing anyone does by hand.** `.github/workflows/release.yml`
fires on `push: tags: ["v*"]` and publishes the package. Pushing the tag *is*
the publish — there is no `npm publish` step to run locally, no registry token
on any developer machine, and asking for one is asking for the wrong thing. If a
version needs to reach consumers, it needs a tag pushed and nothing else.

Consumers then bump: the Platform's `go.mod` require, and the `web` workspace's
`@mosaic-media/sdui` dependency. For local cross-repo work use a `replace` (Go)
or the workspace link (npm) temporarily — **neither may land in a commit.**

## Everything runs in the container, nothing runs on the host

**Do not run `go build`, `go test`, `scripts/generate.sh`, `npx` or `tsc`
directly on this machine.** This repository's gates run inside its test
container:

```bash
docker compose -f docker-compose.test.yml run --rm test
```

That runs the version check, the drift guard, `go build ./...`, `go test ./...`
and the TypeScript typecheck, in the order `.github/workflows/verify.yml` runs
them. Append `bash` for a shell in the same environment.

**This repository needs the container more than any other, because it is the
only one that needs two toolchains at once.** The schema is the contract and the
bindings are output, so the drift guard regenerates Go *and* TypeScript from
`schema/sdui.schema.json` and fails if the committed files moved — one command
needing `go`, `gofmt`, `node` and `npx` together. A host with three of the four
does not fail loudly; it produces a check that passes by not running. Two
specific ways that happens, both real:

- **A different generator version rewrites the bindings**, and the diff reads as
  a stale binding rather than as a different generator. `Dockerfile.test` pins
  what produces them.
- **`scripts/check-versions.mjs` catches its own git failure** and reports "no
  tags yet — nothing to check against", exit 0. Without a working git it passes
  by finding nothing, which is why the container's command runs `git rev-parse`
  first and why the image configures `safe.directory` for the bind mount.

Regenerating is the same command with the script named:
`docker compose -f docker-compose.test.yml run --rm test bash scripts/generate.sh`.

## Workflow

- Commit and push this repository **separately** from `platform` and `web`.
- **Commit author identity** must be `AdamNi-7080 <anicholls41@gmail.com>`.
- The test container green before pushing — it is the drift guard and the
  conformance tests as well as the build.
- Every exported type carries a doc comment saying *why*, not only what. This is
  a published contract read by people who cannot read the Platform's source.

## The roadmap and the decision records

These rules are identical in every Mosaic repository. They exist because the
state of the build and the reasons behind it are the two things that rot fastest
and report nothing when they do — no build fails, no test goes red.

### The roadmap is maintained, not consulted

**`docs/roadmap.md` in [`architecture`](https://github.com/mosaic-media/architecture)
is the single record of where the build is.** Read it before starting work, and
**update it in the same session as the change that dates it** — not in a
follow-up, which does not happen.

- **A slice that lands is marked landed, with what was left out.** "Built" with
  no qualifier is a claim that the whole slice shipped; if part of it did not,
  say which part and why in the same sentence.
- **Implementation that departs from the plan is recorded where it departed.**
  The roadmap is derived from the code, not from the intention that preceded it,
  and the surprises are the most valuable thing in it.
- **Do not restate the roadmap here.** A second copy of "what is built" in a
  `CLAUDE.md` is how the first copy goes stale unnoticed. This file carries how
  to work in *this* repository; the roadmap carries what has been done across all
  of them.
- **A capability with no client path is not done — it is
  [owed](https://github.com/mosaic-media/architecture/blob/main/docs/unreachable-capability.md).**
  If you delete or fail to build a client path to a working service, add its row
  to that register in the same change.

### Decision records are append-only

An ADR is an account of what was decided and why, at a time. It is evidence, not
documentation, and its value is that it was not edited afterwards.

- **Never rewrite a record's body to match what was built.** Not to correct it,
  not to annotate it, not to add "as built, this differs". That pattern turns a
  record into a running commentary and destroys the thing it is for.
- **State changes in the `**Status:**` line, and nowhere else.** That is where a
  record says it is built, built in part (naming the part), or superseded —
  wholly ("Superseded by ADR N") or partly ("Partly superseded: X was reversed by
  ADR N; the rest stands").
- **A changed decision needs a new record that supersedes it.** If the code
  deliberately does something a record decided against, that is a decision and it
  is written down as one, with its own Context / Decision / Alternatives /
  Consequences. Both records then stand: the old one keeps its reasoning, the new
  one carries the change.
- **An unbuilt decision is not a superseded one.** "We have not done this yet"
  belongs in the Status line and the roadmap. Only a genuine reversal earns a new
  record.
- **Records live only in `architecture/docs/adr/`**, numbered sequentially in
  kebab-case. Adding one means adding it to `nav:` in `mkdocs.yml`, and
  `mkdocs build --strict` must pass.

**If the code and a record disagree, say so rather than quietly picking one.** An
honest "this is unresolved" is worth more than a plausible reconciliation that
reads as settled.
