// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui_test

import (
	"testing"

	"github.com/mosaic-media/contracts/sdui"
)

func TestTheSixPredicates(t *testing.T) {
	values := map[string]any{"kind": "series", "name": "", "season": 2.0}
	cases := []struct {
		name string
		p    sdui.Predicate
		want bool
	}{
		{"equals", sdui.Predicate{"field": "kind", "equals": "series"}, true},
		{"equals false", sdui.Predicate{"field": "kind", "equals": "movie"}, false},
		{"equals across types", sdui.Predicate{"field": "season", "equals": "2"}, true},
		{"notEmpty on a blank", sdui.Predicate{"field": "name", "notEmpty": true}, false},
		{"notEmpty inverted", sdui.Predicate{"field": "name", "notEmpty": false}, true},
		{"oneOf", sdui.Predicate{"field": "kind", "oneOf": []any{"movie", "series"}}, true},
		{"not", sdui.Predicate{"not": map[string]any{"field": "kind", "equals": "movie"}}, true},
		{"all", sdui.Predicate{"all": []any{
			map[string]any{"field": "kind", "equals": "series"},
			map[string]any{"field": "season", "equals": "2"},
		}}, true},
		{"all with one false", sdui.Predicate{"all": []any{
			map[string]any{"field": "kind", "equals": "series"},
			map[string]any{"field": "name", "notEmpty": true},
		}}, false},
		{"any", sdui.Predicate{"any": []any{
			map[string]any{"field": "name", "notEmpty": true},
			map[string]any{"field": "kind", "equals": "series"},
		}}, true},
		// A conjunction of nothing is true, a disjunction of nothing is false.
		// Both are reachable: a producer building a condition list may end with none.
		{"empty all", sdui.Predicate{"all": []any{}}, true},
		{"empty any", sdui.Predicate{"any": []any{}}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := sdui.Evaluate(c.p, values); got != c.want {
				t.Errorf("Evaluate = %v, want %v", got, c.want)
			}
		})
	}
}

// An unreadable condition is false, never true. visibleWhen deciding to *show* a
// control because it could not understand its own rule is the fail-open case,
// and it is the one that puts an admin-only affordance on somebody's screen.
func TestAnUnreadablePredicateHidesRatherThanShows(t *testing.T) {
	for _, p := range []sdui.Predicate{
		nil,
		{},
		{"field": "kind"},                        // names a field, states no test
		{"field": "", "equals": "x"},             // no field
		{"gt": 3},                                // not one of the six
		{"all": "not a list"},                    // malformed
		{"not": "not a predicate"},               // malformed
		{"field": "kind", "matches": "^series$"}, // a validator's name, not a predicate's
	} {
		if sdui.Evaluate(p, map[string]any{"kind": "series"}) {
			t.Errorf("%#v evaluated true; an unreadable condition must not reveal anything", p)
		}
	}
}

func TestPredicatesAreExactlySix(t *testing.T) {
	if len(sdui.Predicates) != 6 {
		t.Errorf("the predicate set has %d members; each addition is a step toward the expression language this contract declines to have", len(sdui.Predicates))
	}
}
