# Module types are namespaced

**Status:** Accepted (built)

**Date:** 2026-07-25

## Context

The SDUI node `type` is an **open** vocabulary, and deliberately so: it is what
lets a module introduce a component the Platform never shipped, without a client
release ([architecture#15](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0015-open-and-closed-vocabularies.md),
[contracts#2](0002-primitives-and-definitions.md)).

Open and **flat** are different properties, and only the first was ever decided.
The vocabulary was both.

A client resolves a node's type through a map. Registration is last-writer-wins —
it has to be, because the Platform pushes the authoritative definition library
over whatever a client booted with ([contracts#4](0004-server-delivered-definitions-and-skin.md)).
So a flat vocabulary has two collisions in it and neither produces an error
anywhere:

- **Two modules can mean different things by the same name.** Both contribute a
  `StatChip`; the second to register replaces the first. Every screen the first
  module drew now renders the second module's component with the first module's
  props.
- **A module can take a core component's name.** A module contributing a
  `PosterCard` replaces the core one — not on its own screens, on *every* screen
  in the product.

Neither is hypothetical in the sense that matters: nothing prevents them, nothing
detects them, and the symptom is a component that looks wrong rather than a
failure that reports itself. The client-side simulation of a module's
contribution had already invented its own separator (`module.StatChip`) for a
convention that did not exist, which is the same drift in miniature — a client
deciding a contract question because the contract was silent.

[contracts#8](0008-one-generated-sdui-vocabulary.md) is what makes this fixable now:
until the vocabulary was published as data, "is this a core type?" had no answer
a producer could compute.

## Decision

**A core type is unprefixed. A module's type is `moduleId:type`.**

- **The separator is declared in the vocabulary** (`typeSeparator` in
  `ui.spec.json`) and generated into both bindings and the conformance fixture,
  so a producer reading the Go package and a client reading the published JSON
  cannot disagree about what divides the two halves.
- **No core type may contain it.** Lint refuses one, and runs that check *before*
  generation rather than after — a type name is a function name in both
  bindings, so a malformed one failed gofmt first and dumped the generated file
  to stderr, which reads as a generator bug rather than the spec error it is.
- **The rule is enforced at the `ModuleSettingsUI` boundary**, which is the one
  place a module's own tree crosses into the Platform today. Every node in the
  tree is checked, not the root: a bad type two levels down is precisely the kind
  that survives being looked at.
- **Three answers, and the middle one is the point.** A **core** type is fine —
  composing a settings screen out of the standard vocabulary is what a module is
  *for*, and a rule that refused `Text` would be useless. The module's **own**
  namespace is fine. Anything else is refused, and the error distinguishes an
  unprefixed name that would collide from a reach into another module's
  namespace, because those are different mistakes and deserve different
  sentences.
- **A module id that cannot carry a namespace is refused** before its tree is
  read: `a:b` as an id would make `a:b:Row` ambiguous.
- **Props are not walked.** They are an open bag, and a module's own data
  carrying a field called `type` is data. Only `children` and `slots` are node
  positions, so only they are traversed.

## Alternatives considered

**Reserve a prefix for core instead (`mosaic:PosterCard`, modules unprefixed).**
*Rejected* — it renames every type in the contract, in every definition template,
every client registry and every emit-side call, to protect the smaller of the two
populations. The core vocabulary is closed and reviewed; the open half is the one
that needs the namespace.

**A dot, as the client's simulation had invented.** *Rejected* — a dot is
already meaningful in the binding-path syntax a definition template uses
(`{"$bind": "s.label"}`), and a type and a path appearing side by side in the
same JSON should not share a separator. A colon appears in neither and reads as a
namespace in every language a client might be written in.

**Enforce at registration on the client instead.** *Rejected* — it inverts the
direction of authority, and it puts the check in the one place that cannot
report: the client would have to decide what to do about a collision it detects,
and every client would have to decide the same way.

**Validate only the root node.** *Rejected* — it is the cheap version of the
check and it fails at exactly the depth where mistakes hide.

**Do nothing until a module actually contributes a component.** *Rejected* — the
retrofit cost is the argument. Namespacing a vocabulary after modules have
shipped types in it means renaming other people's components, and this thread is
ordered by what cannot be retrofitted. Enforcing it while every shipped module
uses only core types costs nothing and closes the hole permanently.

## Consequences

- **Every shipped module already passes.** All five modules that contribute a
  settings screen compose it entirely from core types, so this is a rule that
  binds the future rather than a migration.
- **The definition-contribution path inherits the rule for free.** There is no
  way for a module to contribute a `ComponentDefinition` over the wire yet; when
  there is, the name it may register is already decided and already validated.
- **A module cannot pass through another module's component**, even
  deliberately, and even if it knows the other is installed. That is a real
  restriction. It is the right one: a module that could would be depending on
  another module's UI with nothing declaring the dependency.
- **The client's simulated module components were renamed** from
  `module.StatChip` to `demo:StatChip`. They are a simulation of a module's
  contribution, so leaving them on an invented separator after specifying a real
  one would be the drift this record is about.
