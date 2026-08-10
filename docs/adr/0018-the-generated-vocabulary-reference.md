# The generated vocabulary reference

**Status:** Accepted (built)

**Date:** 2026-07-25

## Context

The vocabulary is now one spec generating five artefacts, negotiated on the wire,
enforced by lint, and held in parity by a conformance corpus. It is still
unreadable by anyone who has not read its source.

That is not a cosmetic gap. Most of why DivKit and Beagle are usable by people
who did not write them is that both publish a reference: a document that says
what the vocabulary contains, what each part means, and what a client must
implement. A contract nobody can read without reading its implementation is a
contract with one implementer, which is precisely the state this whole thread was
opened to leave.

## Decision

**`REFERENCE.md` is generated from `ui.spec.json`, and drift-guarded like every
other artefact.**

- **Generated, not written.** A hand-written reference is a fourth place the
  vocabulary is written down, and this repository has already demonstrated what
  becomes of those: three vocabularies disagreeing behind a green build was the
  fault the thread opened on. The reference cannot disagree with the bindings,
  the fixture or the corpus, because all of them come from one file.
- **It carries the things only the spec knows.** Every primitive's `native`
  justification — the reason a definition cannot express it — and the tier table,
  which states the thing an implementer most needs and no type signature
  conveys: a component costs nothing, a primitive costs a release on every
  platform at once.
- **It is drift-guarded.** `scripts/check-generated.sh` regenerates and diffs,
  so a spec change without a regenerated reference fails the container. Verified
  by changing the version and watching `REFERENCE.md` appear in the failure.
- **It ends by saying how to prove a client implements it**, pointing at the
  fixture and the corpus. A reference that describes a contract without saying
  how to check yourself against it is half of what the reader needs.
- **It lives in `contracts`, not in `architecture`.** The architecture
  repository is three documents plus a register, deliberately, and a fourth
  needs a reason that survives being asked why it does not belong in one of the
  existing ones. This belongs beside the thing it describes and ships in the npm
  package with it.

## Alternatives considered

**Write it by hand, in prose.** *Rejected* — it would be better written, and it
would be wrong within two slices. Everything in it is already stated in the spec;
restating it by hand creates the exact drift the thread spent eleven slices
removing.

**Publish it on the architecture site.** *Rejected* — the architecture
repository's own rule is three documents and a register, and this is neither
architecture nor a decision. It also versions with the contract rather than with
the prose, which is an argument for it living where the contract does.

**Generate HTML with anchors and cross-links, like DivKit's.** *Not now* —
Markdown renders on GitHub, in the npm package, and in any editor, and the
content is the part that matters. A site is a presentation decision that can be
taken later without changing what is generated.

**Include the definition templates.** *Rejected* — they are data in
`definitions/*.json`, already published in the package, and a reference that
inlined them would be a second copy of thirty-five files that must stay in step.

## Consequences

- **The contract is now readable by someone who has never seen it**: 525 lines,
  every primitive with its justification, every component, every action, and the
  four closed sets with the reason each is closed.
- **The vocabulary now generates six artefacts from one file** — two authoring
  layers, two registries, the conformance fixture and the reference — and lint
  reconciles the spec against the JSON Schema and the proto besides. That is the
  state [contracts#8](0008-one-generated-sdui-vocabulary.md) set out to reach, arrived at through eleven more slices than it
  described.
- **It will go stale in exactly one way**, and it is worth naming: the prose
  *around* the generated tables is written by hand inside the generator. If the
  vocabulary grows a concept the generator does not know to describe — as it did
  five times in this thread — the tables will include it and the prose will not
  mention it. There is no gate for that, and there cannot easily be one.
- **This closes the SDUI thread's thirteen slices.** What the thread does not
  close is enumerated in the roadmap: `$value` is still live in four external
  modules, no Platform command produces a field rejection, no screen emits a
  lifecycle trigger or an accessibility or focus prop, and primitive props still
  have no typed authoring helpers. Every one of those is a stated position rather
  than an oversight.
