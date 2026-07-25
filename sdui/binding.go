// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui

import "fmt"

// Bindable props: a prop value may be a literal, or a binding naming a path the
// client resolves against the scopes in force where the node renders.
//
// Until now a prop was a literal in every position, and there was exactly one
// escape from that: SubmitField substituted the literal string "$value" into its
// action before dispatching. That mechanism is specced nowhere, carries one
// field, and substitutes *everywhere it appears* in the action — which is why a
// module wanting to submit two fields of one object cannot, and works around it.
// It is the reason no screen in Mosaic can have two inputs on it.
//
// A binding is deliberately the smallest thing that fixes that. It names a path
// and nothing else:
//
//	{"$bind": "ref"}
//
// There is no formatting, no default, no expression and no operator. Formatting
// stays server-side — that is what "presentation-ready data" means, and it is the
// honest reason this needs no evaluator. Spotify's HubFramework is the cited
// warning: the failure was over-generalising a component-driven UI into a virtual
// machine, and every one of those features is a step along that path.
//
// The marker is spelled "$bind" because a definition template already uses it,
// with the same meaning — resolve a path against a scope. The two were different
// mechanisms that happened to share a spelling; they are now one concept applied
// at two moments, and the marker is generated from the vocabulary so a client and
// a producer cannot disagree about it.

// Binding is a prop value that resolves at render rather than being sent whole.
// It is a map rather than a struct because it rides the open props bag, which is
// JSON either way.
type Binding = map[string]any

// Bind names a path to resolve where the node renders.
//
// In this slice the only scope is the screen's params, so `Bind("ref")` is the
// value the screen was navigated with. State scopes and form fields add scopes
// *above* that one and resolve nearest-first, so what this returns does not
// change meaning when they arrive — it gains places to look first.
func Bind(path string) Binding { return Binding{BindingMarker: path} }

// IsBinding reports whether a prop value is a binding.
//
// Exactly one key, and it is the marker. An object carrying the marker beside
// anything else is a literal object, not a malformed binding — the same closed
// reading the action-kind check uses, so a prop that happens to contain a field
// called "$bind" is never silently reinterpreted as one.
func IsBinding(v any) bool {
	m, ok := v.(map[string]any)
	if !ok || len(m) != 1 {
		return false
	}
	path, ok := m[BindingMarker].(string)
	return ok && path != ""
}

// BindingPath returns the path a binding names, and whether the value was one.
func BindingPath(v any) (string, bool) {
	if !IsBinding(v) {
		return "", false
	}
	path, _ := v.(map[string]any)[BindingMarker].(string)
	return path, true
}

// ValidateBindings checks that a props bag uses bindings only where they are
// allowed — which is anywhere a *value* goes, and nowhere a node's shape is
// decided.
//
// A binding is a value, never a type. A tree whose shape depends on data is a
// template, and templates are definitions authored in the contract; allowing
// `{"type": {"$bind": …}}` would make every emitted tree a template and put the
// expansion rules in the wire format.
func ValidateBindings(props map[string]any) error {
	for key, val := range props {
		if key == "type" && IsBinding(val) {
			return fmt.Errorf("a node's type may not be bound — a tree whose shape depends on data is a definition, not a screen")
		}
		if err := validateBindingValue(val); err != nil {
			return fmt.Errorf("prop %q: %w", key, err)
		}
	}
	return nil
}

// validateBindingValue rejects the near-misses that would otherwise be read as
// literals and silently do nothing: the marker present with an empty path, or
// present with a non-string one.
func validateBindingValue(v any) error {
	switch x := v.(type) {
	case map[string]any:
		if raw, present := x[BindingMarker]; present && len(x) == 1 {
			path, ok := raw.(string)
			if !ok {
				return fmt.Errorf("a binding's path must be a string")
			}
			if path == "" {
				return fmt.Errorf("a binding names no path")
			}
			return nil
		}
		for _, val := range x {
			if err := validateBindingValue(val); err != nil {
				return err
			}
		}
	case []any:
		for _, val := range x {
			if err := validateBindingValue(val); err != nil {
				return err
			}
		}
	}
	return nil
}
