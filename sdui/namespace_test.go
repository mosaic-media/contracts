// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui_test

import (
	"strings"
	"testing"

	"github.com/mosaic-media/contracts/sdui"
)

func TestAModuleMayComposeFromTheCoreVocabulary(t *testing.T) {
	// The ordinary case, and the one a wrong rule would break: a settings screen
	// is a Screen of Text and TextField, all core.
	for _, typ := range []string{sdui.TypeScreen, sdui.TypeText, sdui.TypeTextField, sdui.TypeBox} {
		if err := sdui.ValidateModuleType("stremio", typ); err != nil {
			t.Errorf("core type %q refused: %v", typ, err)
		}
	}
}

func TestAModuleMayEmitItsOwnNamespace(t *testing.T) {
	if err := sdui.ValidateModuleType("stremio", "stremio:AddonRow"); err != nil {
		t.Errorf("own namespace refused: %v", err)
	}
	if got := sdui.NamespacedType("stremio", "AddonRow"); got != "stremio:AddonRow" {
		t.Errorf("NamespacedType = %q", got)
	}
}

// The first of the two collisions: an unprefixed name nobody owns. Two modules
// both calling it StatChip would collide in the client's registry, last writer
// winning, with no error anywhere.
func TestAnUnprefixedUnknownTypeIsRefused(t *testing.T) {
	err := sdui.ValidateModuleType("stremio", "StatChip")
	if err == nil {
		t.Fatal("an unprefixed unknown type was accepted")
	}
	if !strings.Contains(err.Error(), `"stremio:StatChip"`) {
		t.Errorf("the error does not say what to write instead: %v", err)
	}
}

// The second: a module naming a core component takes its place on every screen.
// Refused for the same reason as the first, with a different error, because the
// mistake is different.
func TestAModuleMayNotTakeACoreTypesName(t *testing.T) {
	// Emitting a PosterCard node is fine — that is composing.
	if err := sdui.ValidateModuleType("stremio", sdui.TypePosterCard); err != nil {
		t.Fatalf("emitting a core component was refused: %v", err)
	}
	// Claiming the name as its own is not, and cannot be spelled anyway: the
	// namespaced form is a different type, which is exactly the fix.
	if got := sdui.NamespacedType("stremio", sdui.TypePosterCard); sdui.IsCoreType(got) {
		t.Errorf("%q reads as a core type", got)
	}
}

func TestAModuleMayNotReachIntoAnotherModulesNamespace(t *testing.T) {
	err := sdui.ValidateModuleType("stremio", "aiostreams:StreamRow")
	if err == nil {
		t.Fatal("one module emitted another's type")
	}
	if !strings.Contains(err.Error(), "aiostreams") || !strings.Contains(err.Error(), "stremio") {
		t.Errorf("the error names neither module clearly: %v", err)
	}
}

func TestMalformedNamespacesAreRefused(t *testing.T) {
	for _, bad := range []string{"stremio:", "stremio:a:b", ""} {
		if err := sdui.ValidateModuleType("stremio", bad); err == nil {
			t.Errorf("%q was accepted", bad)
		}
	}
}

func TestSplitType(t *testing.T) {
	if id, local, ns := sdui.SplitType("stremio:AddonRow"); !ns || id != "stremio" || local != "AddonRow" {
		t.Errorf("SplitType = %q %q %v", id, local, ns)
	}
	if id, local, ns := sdui.SplitType("PosterCard"); ns || id != "" || local != "PosterCard" {
		t.Errorf("SplitType of a bare type = %q %q %v", id, local, ns)
	}
}

func TestValidateModuleID(t *testing.T) {
	if err := sdui.ValidateModuleID("stremio"); err != nil {
		t.Errorf("a plain id was refused: %v", err)
	}
	for _, bad := range []string{"", "a:b"} {
		if err := sdui.ValidateModuleID(bad); err == nil {
			t.Errorf("module id %q was accepted", bad)
		}
	}
}

// No core type may contain the separator, or every namespaced type becomes
// ambiguous rather than just that one. genui lints the spec for this; this
// checks the registry the lint produces, which is what callers actually read.
func TestNoCoreTypeContainsTheSeparator(t *testing.T) {
	for _, p := range sdui.Primitives {
		if strings.Contains(p.Type, sdui.TypeSeparator) {
			t.Errorf("primitive %q contains %q", p.Type, sdui.TypeSeparator)
		}
	}
	for _, c := range sdui.Components {
		if strings.Contains(c, sdui.TypeSeparator) {
			t.Errorf("component %q contains %q", c, sdui.TypeSeparator)
		}
	}
}
