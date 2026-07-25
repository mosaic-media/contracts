// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui

// State scopes: named variables a subtree can bind to and write.
//
// A tree describes one moment. Something has to remember what was typed between
// two of them, and until now nothing did — a TextInput held its value in the
// client's own component state where nothing could read it, and the only way to
// get a value back out was SubmitField substituting the literal string "$value"
// into an action, one field at a time. That is why no screen in Mosaic can carry
// two inputs.
//
// A State node declares variables and scopes them to its children. A binding
// resolves against the nearest enclosing scope that declares the name, then
// outward, and finally against the screen's params — so a name means the closest
// thing to where it is used, which is the rule every language with lexical scope
// already has and the only one that behaves predictably when scopes nest.
//
// The variables are **declared**, not inferred from what gets written. A client
// must know a name exists before anything writes to it (so a typo is a refusal
// rather than a new variable nobody reads) and must know what type to coerce a
// written value to, because the wire carries setValue's value as a string.

// VarType is the closed set of types a state variable may have.
//
// Three, and no more, for the same reason the validator and predicate sets are
// closed: each one is a coercion every client must implement identically, and a
// type the server can name that a client coerces differently is a value that
// silently differs between platforms.
type VarType = string

const (
	VarString  VarType = "string"
	VarNumber  VarType = "number"
	VarBoolean VarType = "boolean"
)

// StateVar declares one variable in a scope.
type StateVar struct {
	Name string `json:"name"`
	Type string `json:"type"`
	// Value is the initial value. Its absence is not the same as a zero: a
	// string variable with no initial value is empty, and one initialised to ""
	// is also empty, but a *number* with no initial value is unset and a client
	// must not invent 0 for it — a slider that jumps to zero on first render is
	// the visible form of that mistake.
	Value any `json:"value,omitempty"`
}

// Var declares a variable for a State scope.
func Var(name string, typ VarType, initial any) map[string]any {
	v := map[string]any{"name": name, "type": typ}
	if initial != nil {
		v["value"] = initial
	}
	return v
}

// Vars is the props value a State node's `vars` carries.
func Vars(vars ...map[string]any) []any {
	out := make([]any, 0, len(vars))
	for _, v := range vars {
		out = append(out, v)
	}
	return out
}
