// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui

import (
	"fmt"
	"strings"
)

// Bindable props: a prop value may be a literal, or a binding naming a path the
// client resolves against the scopes in force where the node renders.
//
// A binding names a path and nothing else:
//
//	{"$bind": "ref"}
//
// There is no formatting, no default, no expression and no operator. Formatting
// stays server-side — that is what "presentation-ready data" means, and it is why
// no client needs an evaluator. Spotify's HubFramework is the cited warning:
// over-generalising a component-driven UI into a virtual machine, which each of
// those features is a step towards.
//
// The marker is spelled "$bind" because a definition template's binding is, and
// means the same thing — resolve a path against a scope. It is generated from the
// vocabulary as BindingMarker, so a client and a producer cannot disagree about
// what divides a binding from a literal.

// Binding is a prop value that resolves at render rather than being sent whole.
// It is a map rather than a struct because it rides the open props bag, which is
// JSON either way.
type Binding = map[string]any

// Bind names a path to resolve where the node renders.
//
// The screen's params are the outermost scope, so Bind("ref") is the value the
// screen was navigated with. State scopes and form fields add scopes above that
// one and resolve nearest-first, so what this returns does not change meaning as
// they arrive — it gains places to look first.
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
// allowed — anywhere a value goes, and nowhere a node's shape is decided.
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

// ResolveProps resolves every binding in a props bag against a scope, returning
// a new bag. It exists so the conformance corpus can be run against a Go
// implementation as well as a client's — a corpus only one implementation
// executes is a test, not a corpus.
//
// An unresolved binding removes the prop rather than emptying it, which is the
// rule a client follows for the same reason: the component then falls back
// exactly as if the server had not set it, and "" would make "no value" and
// "the empty value" indistinguishable.
func ResolveProps(props map[string]any, scope map[string]any) map[string]any {
	out := make(map[string]any, len(props))
	for key, val := range props {
		resolved, ok := resolveValue(val, scope)
		if !ok {
			continue
		}
		out[key] = resolved
	}
	return out
}

// resolveValue resolves one value. ok is false when a binding found nothing, so
// the caller can drop the key rather than store a nil.
func resolveValue(v any, scope map[string]any) (any, bool) {
	if path, isBinding := BindingPath(v); isBinding {
		found, ok := lookupPath(scope, path)
		return found, ok
	}
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			if resolved, ok := resolveValue(val, scope); ok {
				out[k] = resolved
			}
		}
		return out, true
	case []any:
		out := make([]any, 0, len(x))
		for _, val := range x {
			if resolved, ok := resolveValue(val, scope); ok {
				out = append(out, resolved)
			}
		}
		return out, true
	default:
		return v, true
	}
}

// lookupPath walks a dotted path through the scope.
func lookupPath(scope map[string]any, path string) (any, bool) {
	var cur any = scope
	for _, part := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[part]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}
