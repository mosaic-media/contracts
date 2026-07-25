// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui

import "sort"

// Accessibility in the contract, rather than in each client's judgement.
//
// The props are optional and nothing has to set them, which is exactly why they
// are declared *now*: an optional prop added later is additive, but a vocabulary
// that never had a role model cannot grow one without revisiting every
// definition — and a client that has been inferring roles from types for a year
// has to be talked out of its inferences first.
//
// A role is the piece that has to be a closed set. It is not passed through: a
// web client maps it to an ARIA attribute, a Compose client to a semantics
// property, a SwiftUI client to a trait. A role the server can name and a client
// maps to nothing produces a control that is invisible to a screen reader and
// looks perfectly correct to everybody else — which is the least detectable
// failure this contract can produce, because the people who would notice are the
// ones least likely to be asked.
//
// The accessible label is deliberately not `label`. A visible label and an
// accessible name are different things, and one key meaning both is how a
// control ends up announced as its own caption.

// KnownRole reports whether a role is one the contract declares.
func KnownRole(role string) bool {
	for _, r := range Roles {
		if r == role {
			return true
		}
	}
	return false
}

// UnknownRoles returns the roles in a set that the contract does not declare,
// sorted. A producer checks against this rather than hoping.
func UnknownRoles(roles ...string) []string {
	var out []string
	for _, r := range roles {
		if r != "" && !KnownRole(r) {
			out = append(out, r)
		}
	}
	sort.Strings(out)
	return out
}

// Live region politeness. Anything else is not announced — including the empty
// string, which is the common case and must stay silent.
const (
	LivePolite    = "polite"
	LiveAssertive = "assertive"
)

// ValidHeadingLevel reports whether a heading depth is one an outline can carry.
//
// 1–6 because that is what every platform's outline model has, and because a
// level outside it is a mistake with a silent consequence: a screen reader's
// heading navigation simply skips it.
func ValidHeadingLevel(level int) bool { return level >= 1 && level <= 6 }
