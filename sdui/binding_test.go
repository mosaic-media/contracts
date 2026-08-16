// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui_test

import (
	"strings"
	"testing"

	"github.com/mosaic-media/contracts/sdui"
)

func TestBindNamesAPath(t *testing.T) {
	b := sdui.Bind("ref")
	if !sdui.IsBinding(b) {
		t.Fatalf("Bind did not produce a binding: %v", b)
	}
	path, ok := sdui.BindingPath(b)
	if !ok || path != "ref" {
		t.Errorf("BindingPath = %q %v", path, ok)
	}
	// A binding names a path and nothing else. If this map ever grows a second
	// key, something has added formatting or a default to the wire — which is
	// the step toward the evaluator this vocabulary declines to have.
	if len(b) != 1 {
		t.Errorf("a binding carries more than a path: %v", b)
	}
}

// The closed reading, and the reason for it: a props value that merely contains
// a field called $bind is the producer's own data. Reinterpreting it would
// silently replace real data with a lookup — the same false-positive risk the
// action-kind check is written to avoid.
func TestOnlyASingleMarkerKeyIsABinding(t *testing.T) {
	notBindings := []any{
		map[string]any{"$bind": "x", "format": "upper"}, // marker plus anything
		map[string]any{"$bind": ""},                     // no path
		map[string]any{"$bind": 3},                      // path is not a string
		map[string]any{"bind": "x"},                     // near-miss spelling
		map[string]any{},
		"$bind",
		nil,
		[]any{map[string]any{"$bind": "x"}},
	}
	for i, v := range notBindings {
		if sdui.IsBinding(v) {
			t.Errorf("case %d read as a binding: %#v", i, v)
		}
	}
}

func TestATypeMayNotBeBound(t *testing.T) {
	err := sdui.ValidateBindings(map[string]any{"type": sdui.Bind("kind")})
	if err == nil {
		t.Fatal("a bound type was accepted")
	}
	if !strings.Contains(err.Error(), "definition") {
		t.Errorf("the error does not say why: %v", err)
	}
}

// The near-misses that would otherwise be sent, read as literal objects, and
// draw nothing.
func TestAMalformedBindingIsRefusedRatherThanIgnored(t *testing.T) {
	for _, props := range []map[string]any{
		{"title": map[string]any{"$bind": ""}},
		{"title": map[string]any{"$bind": 7}},
		{"meta": []any{map[string]any{"$bind": ""}}},
		{"style": map[string]any{"inner": map[string]any{"$bind": ""}}},
	} {
		if err := sdui.ValidateBindings(props); err == nil {
			t.Errorf("a malformed binding was accepted: %v", props)
		}
	}
}

func TestOrdinaryPropsPassValidation(t *testing.T) {
	props := map[string]any{
		"title": "Akira",
		"meta":  []any{"1988", "movie"},
		"style": map[string]any{"gap": 3, "direction": "row"},
		// A producer's own data that happens to have a field called $bind
		// alongside others is data, and stays data.
		"filter": map[string]any{"$bind": "x", "and": "more"},
		"poster": sdui.Bind("posterUrl"),
	}
	if err := sdui.ValidateBindings(props); err != nil {
		t.Errorf("ordinary props were refused: %v", err)
	}
}

// The marker is the same string a definition template uses, on purpose: a prop
// binding and a template binding are one concept applied at two moments.
func TestTheMarkerIsTheTemplateBindingSpelling(t *testing.T) {
	if sdui.BindingMarker != "$bind" {
		t.Errorf("BindingMarker = %q; definitions/*.json use $bind", sdui.BindingMarker)
	}
}
