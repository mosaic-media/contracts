// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui

import (
	"fmt"
	"strings"
)

// Type namespacing: core types are unprefixed, a module's are `moduleId:type`.
//
// The vocabulary's `type` is an open string — deliberately, since that is what
// lets a module introduce a component the Platform never shipped (platform#11,
// contracts#2). Open and flat is a different thing: without a namespace two
// modules could both contribute a `StatChip`, and a module could contribute a
// `PosterCard` and take the place of the core component on every screen in the
// product. Neither collision produces an error anywhere — the client's registry
// is a map, and the last writer wins.
//
// A namespace fixes both by construction rather than by policing: a module's
// type carries the id of the module that owns it, so two modules cannot name the
// same type, and no module can name a core one. The check is cheap, mechanical,
// and belongs at the boundary a module's tree crosses.
//
// The separator is generated from the vocabulary spec (TypeSeparator, in
// vocabulary.gen.go) rather than written here, so a client reading the published
// conformance fixture and a producer reading this package cannot disagree about
// what divides the two halves.

// coreTypes is every type the contract itself defines — both tiers. A module may
// use any of them; what it may not do is claim to be one.
var coreTypes = func() map[string]bool {
	m := make(map[string]bool, len(Primitives)+len(Components))
	for _, p := range Primitives {
		m[p.Type] = true
	}
	for _, c := range Components {
		m[c] = true
	}
	return m
}()

// IsCoreType reports whether a node type is part of the published vocabulary.
func IsCoreType(name string) bool { return coreTypes[name] }

// NamespacedType is the type name a module's component must carry.
func NamespacedType(moduleID, name string) string {
	return moduleID + TypeSeparator + name
}

// SplitType divides a namespaced type into the module that owns it and its local
// name. namespaced is false for a bare type, which is a claim to be core.
func SplitType(t string) (moduleID, local string, namespaced bool) {
	id, rest, found := strings.Cut(t, TypeSeparator)
	if !found {
		return "", t, false
	}
	return id, rest, true
}

// ValidateModuleType reports whether a module may emit a node of this type.
//
// Three answers:
//
//   - a core type is fine — composing a settings screen out of the standard
//     vocabulary is exactly what a module is meant to do;
//   - `<moduleID>:<name>` is fine — the module's own namespace;
//   - anything else is refused, whether it is an unprefixed name that would
//     shadow or collide, or another module's namespace.
//
// The error distinguishes the last case from the first: reaching into another
// module's namespace is a different mistake from an unprefixed name.
func ValidateModuleType(moduleID, t string) error {
	if t == "" {
		return fmt.Errorf("a node has no type")
	}
	owner, local, namespaced := SplitType(t)
	if !namespaced {
		if IsCoreType(t) {
			return nil
		}
		return fmt.Errorf("type %q is neither a core type nor namespaced — a module's own type must be %q",
			t, NamespacedType(moduleID, t))
	}
	if owner != moduleID {
		return fmt.Errorf("type %q belongs to module %q, not to %q — a module may only emit its own types",
			t, owner, moduleID)
	}
	if local == "" {
		return fmt.Errorf("type %q has an empty local name", t)
	}
	if strings.Contains(local, TypeSeparator) {
		return fmt.Errorf("type %q has more than one %q — a namespace is one level", t, TypeSeparator)
	}
	return nil
}

// ValidateModuleID reports whether a module id can carry a namespace at all. An
// id containing the separator would make `a:b:C` ambiguous, and an empty one
// would produce a type indistinguishable from a core name with a stray colon.
func ValidateModuleID(moduleID string) error {
	if moduleID == "" {
		return fmt.Errorf("a module id is required to namespace its types")
	}
	if strings.Contains(moduleID, TypeSeparator) {
		return fmt.Errorf("module id %q contains %q, which divides an id from a type name", moduleID, TypeSeparator)
	}
	return nil
}
