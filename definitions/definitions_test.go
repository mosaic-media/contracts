// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package definitions_test

import (
	"encoding/json"
	"testing"

	"github.com/mosaic-media/contracts/definitions"
)

// TestLibraryCarriesEveryDefinition — the library is what the Platform serves,
// so a definition that is on disk and not in it is a component no client can
// render, with nothing to say so.
func TestLibraryCarriesEveryDefinition(t *testing.T) {
	raw, err := definitions.Library()
	if err != nil {
		t.Fatalf("Library(): %v", err)
	}
	var lib []struct {
		Name     string          `json:"name"`
		Template json.RawMessage `json:"template"`
	}
	if err := json.Unmarshal(raw, &lib); err != nil {
		t.Fatalf("library is not a JSON array of definitions: %v", err)
	}
	if len(lib) != definitions.Count() {
		t.Fatalf("library has %d definitions, %d files on disk", len(lib), definitions.Count())
	}
	seen := map[string]bool{}
	for _, d := range lib {
		if d.Name == "" {
			t.Fatal("a definition has no name; nothing can register it")
		}
		if len(d.Template) == 0 {
			t.Fatalf("%s has no template", d.Name)
		}
		if seen[d.Name] {
			t.Fatalf("two definitions named %q — the second would silently replace the first", d.Name)
		}
		seen[d.Name] = true
	}
	// The frame and the card this library grew for, named outright: they are the
	// ones a settings screen cannot render without.
	for _, want := range []string{"Screen", "AppShell", "SettingsFrame", "ExtensionCard"} {
		if !seen[want] {
			t.Errorf("the library is missing %q", want)
		}
	}
}

// TestLibraryIsStable — the Platform embeds these bytes, so an unstable order
// would churn the payload (and its diff) for no reason.
func TestLibraryIsStable(t *testing.T) {
	a, err := definitions.Library()
	if err != nil {
		t.Fatal(err)
	}
	b, err := definitions.Library()
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Fatal("Library() is not byte-stable between calls")
	}
}
