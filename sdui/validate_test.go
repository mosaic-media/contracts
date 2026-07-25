// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui_test

import (
	"strings"
	"testing"

	"github.com/mosaic-media/contracts/sdui"
)

func msg(e *sdui.FieldError) string {
	if e == nil {
		return ""
	}
	return e.Message
}

func TestTheSixValidators(t *testing.T) {
	cases := []struct {
		name   string
		rules  sdui.RuleSet
		value  any
		values map[string]any
		want   string // "" means it passes
	}{
		{"required rejects blank", sdui.RuleSet{"required": true}, "", nil, "Required"},
		{"required rejects whitespace", sdui.RuleSet{"required": true}, "   ", nil, "Required"},
		{"required accepts a value", sdui.RuleSet{"required": true}, "a", nil, ""},
		{"required false accepts blank", sdui.RuleSet{"required": false}, "", nil, ""},

		{"minLength counts characters", sdui.RuleSet{"minLength": 3.0}, "ab", nil, "at least 3"},
		{"minLength counts runes not bytes", sdui.RuleSet{"minLength": 3.0}, "héllo", nil, ""},
		{"maxLength", sdui.RuleSet{"maxLength": 2.0}, "abc", nil, "at most 2"},

		{"pattern rejects", sdui.RuleSet{"pattern": "^[a-z]+$"}, "Ab1", nil, "expected format"},
		{"pattern accepts", sdui.RuleSet{"pattern": "^[a-z]+$"}, "abc", nil, ""},
		// An empty value is `required`'s business, not the pattern's; otherwise
		// every optional patterned field would be effectively required.
		{"pattern ignores an empty value", sdui.RuleSet{"pattern": "^[a-z]+$"}, "", nil, ""},
		// A pattern that does not compile is the server's mistake; failing the
		// field would blame the person typing for it.
		{"an uncompilable pattern does not blame the user", sdui.RuleSet{"pattern": "([a-z"}, "abc", nil, ""},

		{"matches", sdui.RuleSet{"matches": "password"}, "b", map[string]any{"password": "a"}, "Does not match"},
		{"matches passes", sdui.RuleSet{"matches": "password"}, "a", map[string]any{"password": "a"}, ""},

		{"oneOf rejects", sdui.RuleSet{"oneOf": []any{"a", "b"}}, "c", nil, "allowed value"},
		{"oneOf accepts", sdui.RuleSet{"oneOf": []any{"a", "b"}}, "b", nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := msg(sdui.ValidateField("f", c.value, c.rules, c.values))
			if c.want == "" && got != "" {
				t.Errorf("expected to pass, got %q", got)
			}
			if c.want != "" && !strings.Contains(got, c.want) {
				t.Errorf("got %q, want something containing %q", got, c.want)
			}
		})
	}
}

// The order is fixed rather than map order. A message that varies between runs
// is one a user cannot describe and a test cannot assert — and "Required" is the
// one a person can act on, so it wins over "too short".
func TestTheFirstFailureIsTheMostFundamental(t *testing.T) {
	rules := sdui.RuleSet{"required": true, "minLength": 5.0, "pattern": "^x"}
	for i := 0; i < 20; i++ {
		if got := msg(sdui.ValidateField("f", "", rules, nil)); got != "Required" {
			t.Fatalf("run %d gave %q", i, got)
		}
	}
}

// The set is closed, and an unknown rule is refused rather than ignored.
// Ignoring it produces a field that accepts anything, silently — the fail-open
// case the closed set exists to prevent.
func TestAnUnknownValidatorIsRefused(t *testing.T) {
	unknown := sdui.ValidateRuleSet(sdui.RuleSet{"required": true, "luhn": true, "creditCard": true})
	if len(unknown) != 2 || unknown[0] != "creditCard" || unknown[1] != "luhn" {
		t.Errorf("ValidateRuleSet = %v", unknown)
	}
	if sdui.ValidateRuleSet(sdui.RuleSet{"required": true, "oneOf": []any{"a"}}) != nil {
		t.Error("a set of known rules was reported as unknown")
	}
	for _, v := range sdui.Validators {
		if !sdui.KnownValidator(v.Name) {
			t.Errorf("the registry declares %q but KnownValidator refuses it", v.Name)
		}
	}
}

func TestValidatorsAreExactlySix(t *testing.T) {
	if len(sdui.Validators) != 6 {
		t.Errorf("the validator set has %d members; widening it lets the server state a rule a client cannot enforce", len(sdui.Validators))
	}
}
