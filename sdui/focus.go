// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui

import "sort"

// Focus lives in the contract rather than in each client's judgement.
//
// A web client can mostly get away without it: the browser has a tab order and
// a focus ring, and a keyboard user gets something. A television cannot. There is
// no tab key on a remote; there is a direction pad, and every press has to
// resolve to exactly one node or the viewer is stuck. A vocabulary without a
// focus model has every client inventing one and no two agreeing, which is why
// it cannot be added afterwards.
//
// The mechanism is geometric and belongs to the client: given the focused node
// and a direction, find the nearest focusable node that way. What the contract
// carries is the four things geometry cannot know — which nodes are worth
// landing on, which sets of them behave as one stop, where focus should start,
// and the specific cases where the geometric answer is wrong.

// FocusDirection is one of the four a nextFocus override may name.
type FocusDirection = string

const (
	FocusUp    FocusDirection = "up"
	FocusDown  FocusDirection = "down"
	FocusLeft  FocusDirection = "left"
	FocusRight FocusDirection = "right"
)

// KnownFocusDirection reports whether a direction is one of the four.
//
// Diagonals are absent deliberately: no remote has them, and declaring them
// would make every client implement a geometry nobody presses.
func KnownFocusDirection(d string) bool {
	for _, k := range FocusDirections {
		if k == d {
			return true
		}
	}
	return false
}

// NextFocus builds the override map from direction to target node id, refusing
// any direction the contract does not declare.
//
// It refuses rather than drops, because a nextFocus naming a direction nothing
// resolves is a dead end on a remote control, and "focus went nowhere when I
// pressed right" is the least reportable bug there is — the person hitting it
// cannot get anywhere to report it from.
func NextFocus(targets map[string]string) (map[string]any, []string) {
	var unknown []string
	out := map[string]any{}
	for dir, id := range targets {
		if !KnownFocusDirection(dir) {
			unknown = append(unknown, dir)
			continue
		}
		if id != "" {
			out[dir] = id
		}
	}
	sort.Strings(unknown)
	if len(unknown) > 0 {
		return nil, unknown
	}
	return out, nil
}
