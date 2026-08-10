# One generated SDUI vocabulary

**Status:** Accepted (built)
**Date:** 2026-07-25

## Context

[contracts#2](0002-primitives-and-definitions.md) split the interface into two
tiers: **primitives**, native code each client implements, and **definitions**,
compositions expressed as data. [contracts#7](0007-components-are-authored-only-in-the-contract.md)
then closed the definition tier — a component is authored in the contract and
nowhere else, and `genui -lint` refuses one that is not.

Nothing was ever said about the other tier, and it drifted for the whole life of
the project behind a green build.

Trying to build a first-boot onboarding screen is what exposed it. That needs
several inputs submitted together, and the contract cannot express one. Looking
for the missing piece turned up something larger: **the contract does not
describe the vocabulary it claims to describe.** Three places wrote the
vocabulary down and no two agreed.

| | Primitives | Action kinds | Components |
|---|---|---|---|
| Wire (`proto`) | not enumerated | 10 | open string |
| Authoring spec (`ui.spec.json`) | **0** | 4 | 34 + one stray primitive |
| Client (`sdui-react`) | 25 | 9 | 0 — correct |

Nine cells; `genui -lint` reconciled one of them. Every specific defect found
while looking is an instance of that single fault:

- **The primitive tier existed only as TypeScript**, in one client's
  `components/index.ts`. A Swift or Kotlin client could not be written from the
  published contract at all, because the published contract does not say what a
  client must implement.
- **`sdui/components.go` was a hand-written constant list that had drifted.** It
  named `SearchBar`, `Slider` and `ProgressBar` as components when all three are
  primitives, and omitted every real primitive.
- **`Player` sat in the component list with no definition file** — authorable,
  and expanding to nothing on any client that took the library at its word.
- **The Platform reached past the typed layer for seven primitives**, as
  `ui.Component("SearchBar", …)` — the untyped escape hatch its own instructions
  warn against, used because there was nothing else to use.
- **Six of the ten wire action kinds were unauthorable from Go**, and one
  (`query`) had been removed from the schema by
  [architecture#61](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0061-one-client-transport.md) while surviving in the proto, fields
  and all.

The benchmark for what a complete answer looks like is deliberately external.
[DivKit](https://divkit.tech) generates every platform from one schema and holds
four clients in parity with a shared conformance corpus; that mechanism, not its
feature list, is the part worth copying.

## Decision

**`ui.spec.json` is the single source of the vocabulary, and all three tiers are
generated from it.**

- **The spec declares the whole vocabulary**: a `version`, the primitives (each
  with the props it reads and a `native` justification), the components, the
  action kinds, the tones, the overlay surfaces, the closed validator set and the
  closed predicate set.
- **`tools/genui` generates five artefacts** from it — the Go and TypeScript
  authoring constructors for *both* type tiers, the machine-readable registry in
  each language (`sdui/vocabulary.gen.go`, `ts/vocabulary.gen.ts`), and a client
  conformance fixture (`conformance/vocabulary.json`). The hand-written
  `sdui/components.go` is deleted; its constants are generated, which is why the
  drift is not repeatable rather than merely repaired.
- **Every primitive states why it is native.** Growing that tier is the only
  vocabulary change that costs a client release, so an entry that does not
  justify itself is an addition nobody priced. Lint refuses one.
- **Lint reconciles the three written-down vocabularies.** The spec's action
  kinds, tones and surfaces must be exactly those in `schema/sdui.schema.json`
  and `proto/mosaic/sdui/v1/sdui.proto`; a type a definition's template
  references must exist in one of the tiers; no type may be in both tiers; every
  component must have a definition file and every definition a component.
- **The conformance fixture is data, in both packages.** A client asserts that
  what it registers is exactly what the contract declares. That sentence could
  not previously be written as a test in any language.

**`query` returns to the contract as a different action.** [architecture#61](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0061-one-client-transport.md) removed a
kind of that name which carried a raw GraphQL string and a region to refresh
into; it was unimplementable with no endpoint to send it to. The kind declared
here names a screen the server already knows how to build, exactly as `navigate`
does, and differs from `navigate` only in not pushing history. The proto's field
numbers 7–9 are **reserved, not reused**, so an old message cannot be decoded as
the new action.

**Three kinds are declared and not yet interpreted by any client** — `query`,
`setValue` and `submit`, along with the `Form` primitive, the validators and the
predicates. They are in the vocabulary now because a vocabulary is versioned and
negotiated as a whole, and because declaring the shape before building against it
is what stops the next slice inventing a different one. The honest statement of
their status is that the conformance fixture lists them and no client implements
them; that is a visible gap rather than a silent one, which is the difference
this record is about.

## Alternatives considered

**Generate the client's registry too.** *Rejected for now* — attractive, and it
is where this ends up, but it is a change to `sdui-react` and this slice is
deliberately contract-only and behaviour-free. The fixture is the seam that makes
it possible without making it urgent.

**Keep the primitive tier hand-written and add a lint that reads the client's
TypeScript.** *Rejected* — it inverts the direction of authority. The contract
would then be derived from one client, which is what caused the drift; and it
generalises to nothing, since the second client is the whole point.

**Make the JSON Schema the vocabulary source and generate the spec from it.**
*Rejected* — the schema describes the *shape* of a node, not the set of node
types, and expressing "these 26 types are native, these 34 are data, and here is
why each of the 26 has to be" in JSON Schema means encoding it in `description`
strings. The schema stays the wire contract; the spec is the vocabulary; lint
holds them together.

**Declare only what is implemented today, and add the rest slice by slice.**
*Rejected* — the version is what a client negotiates against, and a version that
grows on every slice makes negotiation a moving target during the exact period it
is being built. The cost is that the vocabulary is briefly ahead of every client,
which the fixture makes measurable.

## Consequences

- **A second client becomes tractable.** What to implement is published data, and
  whether an implementation is complete is a test.
- **The untyped escape hatch stops being the only option for primitives.**
  `ui.Component("SearchBar", …)` still compiles — a module's own type needs it —
  but `ui.SearchBar(…)` now exists, and `ui.Player` is a primitive constructor
  rather than a component that had no definition behind it.
- **The vocabulary is ahead of the client, on purpose and visibly.** 26 declared
  primitives against 25 implemented, 12 action kinds against 9. Nothing emits the
  difference; the fixture is what will fail when a client claims conformance it
  does not have.
- **Four prop keys carry two types across primitives** — `value` is a string, a
  number and a boolean; `style` is a box style and a text style; `name` is an
  icon name and a field name; `size` is a token and a number. That violates the
  contract's own one-key-one-type rule, it is faithfully recorded here rather than
  quietly corrected, and fixing it renames props on the wire. It belongs to the
  slice that changes what a prop is, not to this one.
- **The drift guard is slower and covers more.** Five generated artefacts instead
  of two, and lint now reads the schema and the proto as well as the spec.
