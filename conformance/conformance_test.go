// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package conformance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mosaic-media/contracts/conformance"
	"github.com/mosaic-media/contracts/definitions"
	"github.com/mosaic-media/contracts/sdui"
)

// The fixture and the Go registry are generated from the same spec, so these
// tests are not checking the generator against itself: they are checking that
// the two artefacts a consumer actually reads — the embedded JSON a non-Go
// client loads and the typed registry the Platform imports — say the same thing.
// A generator change that updated one and not the other would otherwise ship.

func load(t *testing.T) conformance.Fixture {
	t.Helper()
	f, err := conformance.Load()
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	return f
}

func TestFixtureMatchesRegistry(t *testing.T) {
	f := load(t)

	if f.Version != sdui.VocabularyVersion {
		t.Errorf("version: fixture %q, registry %q", f.Version, sdui.VocabularyVersion)
	}
	if got, want := len(f.Primitives), len(sdui.Primitives); got != want {
		t.Fatalf("primitive count: fixture %d, registry %d", got, want)
	}
	for i, p := range f.Primitives {
		r := sdui.Primitives[i]
		if p.Type != r.Type || p.Tier != r.Tier || p.Children != r.Children {
			t.Errorf("primitive %d: fixture %+v, registry %s/%s/%t", i, p, r.Type, r.Tier, r.Children)
		}
		if len(p.Props) != len(r.Props) {
			t.Errorf("primitive %s: fixture has %d props, registry %d", p.Type, len(p.Props), len(r.Props))
			continue
		}
		for j, key := range p.Props {
			if key != r.Props[j].Key {
				t.Errorf("primitive %s prop %d: fixture %q, registry %q", p.Type, j, key, r.Props[j].Key)
			}
		}
	}
	if len(f.Components) != len(sdui.Components) {
		t.Errorf("component count: fixture %d, registry %d", len(f.Components), len(sdui.Components))
	}
	if len(f.Actions) != len(sdui.ActionKinds) {
		t.Errorf("action count: fixture %d, registry %d", len(f.Actions), len(sdui.ActionKinds))
	}
	for i, k := range f.Actions {
		if i < len(sdui.ActionKinds) && k != sdui.ActionKinds[i].Kind {
			t.Errorf("action %d: fixture %q, registry %q", i, k, sdui.ActionKinds[i].Kind)
		}
	}
}

// Every primitive states why it is native. Growing this tier is the only change
// in the vocabulary that costs a client release (ADR 0024), and an entry with no
// justification is one that was added without anyone pricing it.
func TestEveryPrimitiveJustifiesItself(t *testing.T) {
	for _, p := range sdui.Primitives {
		if p.Native == "" {
			t.Errorf("primitive %s has no native justification", p.Type)
		}
		if p.Doc == "" {
			t.Errorf("primitive %s has no doc", p.Type)
		}
	}
}

// The two tiers are disjoint. A type in both is unresolvable for a client: it
// would have to choose between its own implementation and the served definition,
// and nothing in the contract says which wins.
func TestTiersDoNotOverlap(t *testing.T) {
	prim := map[string]bool{}
	for _, p := range sdui.Primitives {
		prim[p.Type] = true
	}
	for _, c := range sdui.Components {
		if prim[c] {
			t.Errorf("%s is both a primitive and a component", c)
		}
	}
}

// Every component the registry names is a definition file the Platform can
// serve. A component with no definition is authorable from Go and renders as an
// Unknown placeholder on every client — a hole in a screen that nothing reports.
func TestEveryComponentHasADefinition(t *testing.T) {
	raw, err := definitions.Library()
	if err != nil {
		t.Fatalf("library: %v", err)
	}
	var defs []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &defs); err != nil {
		t.Fatalf("decode library: %v", err)
	}
	have := map[string]bool{}
	for _, d := range defs {
		have[d.Name] = true
	}
	for _, c := range sdui.Components {
		if !have[c] {
			t.Errorf("component %s has no definition file", c)
		}
	}
	if len(defs) != len(sdui.Components) {
		t.Errorf("library has %d definitions, registry names %d components", len(defs), len(sdui.Components))
	}
}

// Every node type a definition's template expands into exists in the vocabulary.
// This is the check that would have caught a definition referring to a component
// that had been renamed, which expands to a node no client can render.
func TestDefinitionTemplatesReferenceKnownTypes(t *testing.T) {
	known := map[string]bool{}
	for _, p := range sdui.Primitives {
		known[p.Type] = true
	}
	for _, c := range sdui.Components {
		known[c] = true
	}

	files, err := filepath.Glob("../definitions/*.json")
	if err != nil || len(files) == 0 {
		t.Fatalf("no definitions found: %v", err)
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		var def struct {
			Name     string `json:"name"`
			Template any    `json:"template"`
		}
		if err := json.Unmarshal(data, &def); err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		for _, typ := range referencedTypes(def.Template) {
			if !known[typ] {
				t.Errorf("%s: template references unknown type %q", filepath.Base(f), typ)
			}
		}
	}
}

func referencedTypes(v any) []string {
	var out []string
	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			if t, ok := x["type"].(string); ok {
				out = append(out, t)
			}
			for _, val := range x {
				walk(val)
			}
		case []any:
			for _, val := range x {
				walk(val)
			}
		}
	}
	walk(v)
	return out
}
