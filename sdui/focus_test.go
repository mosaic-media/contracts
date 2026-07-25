// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui_test

import (
	"testing"

	"github.com/mosaic-media/contracts/sdui"
)

func TestTheFourDirections(t *testing.T) {
	for _, d := range []string{"up", "down", "left", "right"} {
		if !sdui.KnownFocusDirection(d) {
			t.Errorf("%q refused", d)
		}
	}
	// No remote has a diagonal, and declaring one would make every client
	// implement a geometry nobody presses.
	for _, d := range []string{"upLeft", "forward", "next", ""} {
		if sdui.KnownFocusDirection(d) {
			t.Errorf("%q accepted", d)
		}
	}
	if len(sdui.FocusDirections) != 4 {
		t.Errorf("the direction set has %d members", len(sdui.FocusDirections))
	}
}

func TestNextFocusRefusesADirectionNothingResolves(t *testing.T) {
	out, unknown := sdui.NextFocus(map[string]string{"right": "card-2", "diagonal": "card-9"})
	if out != nil {
		t.Error("a map with an unresolvable direction was returned anyway")
	}
	if len(unknown) != 1 || unknown[0] != "diagonal" {
		t.Errorf("unknown = %v", unknown)
	}
}

func TestNextFocusBuildsTheOverride(t *testing.T) {
	out, unknown := sdui.NextFocus(map[string]string{"right": "card-2", "down": "rail-2", "up": ""})
	if unknown != nil {
		t.Fatalf("unknown = %v", unknown)
	}
	if out["right"] != "card-2" || out["down"] != "rail-2" {
		t.Errorf("out = %#v", out)
	}
	// A direction pointing at nothing is left out rather than carried as an
	// empty target: the client's geometry is a better answer than a dead end.
	if _, present := out["up"]; present {
		t.Error("an empty target was carried")
	}
}
