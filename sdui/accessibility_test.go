// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui_test

import (
	"testing"

	"github.com/mosaic-media/contracts/sdui"
)

func TestTheRoleSetIsClosed(t *testing.T) {
	if !sdui.KnownRole("heading") || !sdui.KnownRole("navigation") {
		t.Error("a declared role was not recognised")
	}
	// The failure this closes is the least detectable one the contract can
	// produce: a role no client maps is a control invisible to a screen reader
	// and correct-looking to everyone else.
	if sdui.KnownRole("carousel") || sdui.KnownRole("") {
		t.Error("an undeclared role was accepted")
	}
	got := sdui.UnknownRoles("heading", "carousel", "banner", "list")
	if len(got) != 2 || got[0] != "banner" || got[1] != "carousel" {
		t.Errorf("UnknownRoles = %v", got)
	}
	if sdui.UnknownRoles("heading", "list") != nil {
		t.Error("known roles were reported as unknown")
	}
}

// The accessible label may not share a key with the visible one. A visible
// label and an accessible name are different things, and one key meaning both is
// how a control ends up announced as its own caption.
func TestTheAccessibleNameHasItsOwnKey(t *testing.T) {
	keys := map[string]bool{}
	for _, p := range sdui.Primitives {
		for _, pr := range p.Props {
			keys[pr.Key] = true
		}
	}
	if !keys["label"] {
		t.Skip("no primitive declares a visible label; the collision this guards cannot arise")
	}
	if keys["a11yLabel"] {
		t.Error("a11yLabel is a primitive prop as well as universal sugar; one of the two is unread")
	}
}

func TestHeadingLevelsAreAnOutline(t *testing.T) {
	for _, ok := range []int{1, 3, 6} {
		if !sdui.ValidHeadingLevel(ok) {
			t.Errorf("level %d refused", ok)
		}
	}
	// A level outside the outline is a mistake with a silent consequence: a
	// screen reader's heading navigation skips it.
	for _, bad := range []int{0, 7, -1} {
		if sdui.ValidHeadingLevel(bad) {
			t.Errorf("level %d accepted", bad)
		}
	}
}
