// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// The six validators, and why there are only six.
//
// A validator is a rule stated by the server and enforced by the client. That
// arrangement fails open by construction if the set is open: a server naming a
// rule a client does not implement gets a field that accepts anything, silently,
// and the only symptom is bad data arriving later. So the set is closed, small,
// and implementable identically in every client language — no back-references,
// no lookahead, nothing whose semantics differ between regex engines.
//
// The server still enforces everything, because a client is not a trust
// boundary. What the client's copy buys is the round trip: a field that is
// obviously wrong says so before anything is sent. When the server rejects
// something the client could not have known — a username already taken — it
// answers in the same shape (FieldErrors), so a rejection from either side
// renders in the same place.

// FieldError is one field's rejection, in the shape both sides produce.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// RuleSet is the validators attached to one field, as they ride the props bag:
// the validator's name to its argument. It is not called Validators because
// that is the generated registry of the six the contract declares — this is one
// field's selection from them.
type RuleSet = map[string]any

// ValidateField applies a field's validators to its value, returning the first
// failure. Order is fixed rather than map order, so the same bad value produces
// the same message on every platform and in every run — a message that varies
// between runs is one a user cannot describe and a test cannot assert.
//
// `values` is the rest of the scope, which `matches` needs and nothing else
// reads.
func ValidateField(field string, value any, rules RuleSet, values map[string]any) *FieldError {
	for _, name := range validatorOrder {
		arg, present := rules[name]
		if !present {
			continue
		}
		if msg := applyValidator(name, arg, value, values); msg != "" {
			return &FieldError{Field: field, Message: msg}
		}
	}
	return nil
}

// validatorOrder is the order rules are applied in — cheapest and most
// fundamental first, so "this is required" wins over "this is too short", which
// is the message a person can act on.
var validatorOrder = []string{"required", "minLength", "maxLength", "pattern", "matches", "oneOf"}

// KnownValidator reports whether a name is one of the six the contract declares.
//
// A rule outside the set is refused rather than ignored, and that direction is
// the whole reason the set is closed: ignoring an unknown rule produces a field
// that accepts anything, silently, with the bad data arriving somewhere else
// later. Refusing it produces a complaint at the boundary, addressed to whoever
// wrote the rule.
func KnownValidator(name string) bool {
	for _, v := range Validators {
		if v.Name == name {
			return true
		}
	}
	return false
}

// ValidateRuleSet checks that every rule named in a field's set is one of the
// six, returning the offenders sorted.
func ValidateRuleSet(rules RuleSet) []string {
	var unknown []string
	for name := range rules {
		if !KnownValidator(name) {
			unknown = append(unknown, name)
		}
	}
	sort.Strings(unknown)
	return unknown
}

func applyValidator(name string, arg, value any, values map[string]any) string {
	s := asText(value)
	switch name {
	case "required":
		if want, _ := arg.(bool); want && strings.TrimSpace(s) == "" {
			return "Required"
		}
	case "minLength":
		if n, ok := asInt(arg); ok && len([]rune(s)) < n {
			return fmt.Sprintf("Must be at least %d characters", n)
		}
	case "maxLength":
		if n, ok := asInt(arg); ok && len([]rune(s)) > n {
			return fmt.Sprintf("Must be at most %d characters", n)
		}
	case "pattern":
		pat, _ := arg.(string)
		if pat == "" || s == "" {
			return ""
		}
		// RE2 (Go's regexp) rather than PCRE, so no input can make this
		// backtrack — a validator a user can hang their own client with is
		// worse than no validator.
		re, err := regexp.Compile(pat)
		if err != nil {
			// A pattern that does not compile is the server's mistake, and
			// failing the field would blame the user for it.
			return ""
		}
		if !re.MatchString(s) {
			return "Not in the expected format"
		}
	case "matches":
		other, _ := arg.(string)
		if other == "" {
			return ""
		}
		if s != asText(values[other]) {
			return "Does not match " + other
		}
	case "oneOf":
		list, ok := arg.([]any)
		if !ok || len(list) == 0 {
			return ""
		}
		for _, item := range list {
			if asText(item) == s {
				return ""
			}
		}
		return "Not an allowed value"
	}
	return ""
}

func asText(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", x)
	}
}

func asInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case float64:
		return int(x), true
	default:
		return 0, false
	}
}
