// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui

// The vocabulary registry's types. The data is generated into vocabulary.gen.go
// from ui.spec.json; these are the shapes it fills.
//
// The registry exists because the vocabulary was previously knowable only by
// reading a client's source. The primitive tier — the set a client must
// implement natively — lived as TypeScript in the React client alone, so a Swift
// or Kotlin client could not be written from the published contract at all, and
// nothing anywhere could report that the three written-down vocabularies (this
// module, the JSON Schema and the proto) had stopped agreeing. Publishing the
// vocabulary as data is what makes "does this client implement the contract?" a
// question with an answer; conformance/vocabulary.json is the same data for a
// consumer that is not Go.

// PropSpec is one property a primitive reads.
type PropSpec struct {
	// Key is the name in the node's open props bag.
	Key string
	// Type names the value's shape. It is the contract's own type language
	// (string, number, boxStyle, enum:cover|contain, array:option), not any one
	// language's — the same string has to mean something in Go, TypeScript,
	// Swift and Kotlin.
	Type string
	Doc  string
}

// PrimitiveSpec is one native node type: something a client implements itself
// because a tree of other nodes genuinely cannot express it.
type PrimitiveSpec struct {
	Type string
	// Tier groups primitives by why they are native — presentational,
	// interactive, field, computed, player.
	Tier string
	Doc  string
	// Native is the justification for the entry: what a definition cannot do.
	// It is required, and lint refuses a primitive without one, because growing
	// this set is the only change in the vocabulary that costs a client release
	// (ADR 0024). An addition nobody had to justify is one nobody priced.
	Native string
	// Children reports whether the client renders this node's child list. A leaf
	// that is handed children drops them silently, which is why the contract
	// states it rather than leaving each client to decide.
	Children bool
	Props    []PropSpec
}

// ActionSpec is one behaviour a client interprets.
type ActionSpec struct {
	Kind string
	Doc  string
}

// ValidatorSpec is one member of the closed field-validation set.
type ValidatorSpec struct {
	Name string
	// Arg names the type of the validator's argument (boolean, number, string,
	// array:string).
	Arg string
	Doc string
}

// PredicateSpec is one member of the closed conditional set — what a client can
// evaluate against form state without an expression language, and therefore
// without the evaluator VM that deprecated Spotify's HubFramework.
type PredicateSpec struct {
	Name string
	Doc  string
}
