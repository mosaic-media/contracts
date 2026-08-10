// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

// Package conformance publishes the vocabulary a conforming client must
// implement, as data.
//
// A client is conformant when the node types it renders natively are exactly the
// primitives listed here, and the action kinds it interprets are exactly these
// kinds. That sentence could not be written as a test before: the primitive tier
// existed only as TypeScript inside the React client, so "what should this client
// implement?" had no answer outside that client's own source, and a second client
// in Swift or Kotlin could not be written from the published contract at all.
//
// The fixture is generated from ui.spec.json and is shipped both ways — embedded
// here for a Go consumer, and exported from the npm package as
// `@mosaic-media/sdui/conformance` for a JavaScript one. A client in a language
// with neither reads the JSON directly; that is the point of it being data.
package conformance

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed vocabulary.json
var raw []byte

// Fixture is the declared vocabulary. Every field is a set a client is measured
// against, not a suggestion.
type Fixture struct {
	Comment string `json:"//"`
	// Version is the vocabulary version a client declares it implements.
	Version string `json:"version"`
	// Primitives are the node types a client renders itself.
	Primitives []Primitive `json:"primitives"`
	// Components are the node types the Platform serves as definitions. A client
	// implements none of them — one appearing in a client's registry is the
	// drift contracts#7 was written about.
	Components []string `json:"components"`
	Actions    []string `json:"actions"`
	Validators []string `json:"validators"`
	Predicates []string `json:"predicates"`
	Tones      []string `json:"tones"`
	Surfaces   []string `json:"surfaces"`
}

// Primitive is one native node type in the fixture: its name, why-tier, whether
// its children are rendered, and the prop keys it reads.
type Primitive struct {
	Type     string   `json:"type"`
	Tier     string   `json:"tier"`
	Children bool     `json:"children"`
	Props    []string `json:"props"`
}

// Load returns the embedded fixture.
func Load() (Fixture, error) {
	var f Fixture
	if err := json.Unmarshal(raw, &f); err != nil {
		return Fixture{}, fmt.Errorf("conformance fixture: %w", err)
	}
	return f, nil
}

// JSON returns the fixture's raw bytes, for a consumer that wants to hand them
// to a client's own test harness rather than decode them here.
func JSON() []byte { return raw }
