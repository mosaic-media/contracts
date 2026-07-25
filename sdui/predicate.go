// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui

// The six predicates: conditions a client can evaluate against scope values.
//
// `visibleWhen` is the reason they exist, and the shape it did *not* take is the
// point. A bespoke "show this when that field is set" prop would have answered
// the first screen that needed it and nothing else, and the second screen would
// have earned a second prop. A closed predicate set answers all of them with one
// concept — and refuses the third thing, which is an expression language.
//
// There is no arithmetic, no comparison beyond equality, no function call and no
// path into anything but a named field. Each of the six is a line of code in
// every client language, which is the bar: a condition a client cannot evaluate
// identically is a screen that differs between platforms for reasons nobody can
// see.
//
// A predicate is a JSON object with exactly one shape:
//
//	{"field": "kind",  "equals": "series"}
//	{"field": "name",  "notEmpty": true}
//	{"field": "kind",  "oneOf": ["movie", "series"]}
//	{"not": <predicate>}
//	{"all": [<predicate>, …]}
//	{"any": [<predicate>, …]}

// Predicate is a condition over scope values.
type Predicate = map[string]any

// Evaluate answers a predicate against a set of values.
//
// An unrecognised shape is **false**, and the direction matters: a condition
// nothing understands must not reveal something. `visibleWhen` deciding to show
// a control because it could not read its own rule is the fail-open case, and it
// is the one that puts an admin-only affordance on a screen.
func Evaluate(p Predicate, values map[string]any) bool {
	if p == nil {
		return false
	}
	if inner, ok := p["not"]; ok {
		nested, ok := inner.(map[string]any)
		return ok && !Evaluate(nested, values)
	}
	if list, ok := p["all"]; ok {
		items, ok := list.([]any)
		if !ok {
			return false
		}
		for _, item := range items {
			nested, ok := item.(map[string]any)
			if !ok || !Evaluate(nested, values) {
				return false
			}
		}
		// An empty `all` is true, as a conjunction of nothing is — and it is
		// reachable, because a producer building a list of conditions may end up
		// with none.
		return true
	}
	if list, ok := p["any"]; ok {
		items, ok := list.([]any)
		if !ok {
			return false
		}
		for _, item := range items {
			if nested, ok := item.(map[string]any); ok && Evaluate(nested, values) {
				return true
			}
		}
		return false
	}

	field, ok := p["field"].(string)
	if !ok || field == "" {
		return false
	}
	value := values[field]
	if want, ok := p["equals"]; ok {
		return asText(value) == asText(want)
	}
	if want, ok := p["notEmpty"]; ok {
		set, _ := want.(bool)
		return set == (asText(value) != "")
	}
	if list, ok := p["oneOf"].([]any); ok {
		for _, item := range list {
			if asText(item) == asText(value) {
				return true
			}
		}
		return false
	}
	return false
}
