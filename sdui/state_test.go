// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package sdui_test

import (
	"encoding/json"
	"testing"

	"github.com/mosaic-media/contracts/sdui"
)

func TestVarDeclaresNameAndType(t *testing.T) {
	v := sdui.Var("username", sdui.VarString, "")
	if v["name"] != "username" || v["type"] != "string" {
		t.Fatalf("Var = %#v", v)
	}
	if v["value"] != "" {
		t.Errorf("an explicit empty initial value was dropped: %#v", v)
	}
}

// An absent initial value must stay absent. A number with no initial value is
// unset, and a client that invented 0 for it would jump a slider to zero on
// first render — the visible form of conflating "unset" with "the zero value".
func TestAnAbsentInitialValueIsNotAZero(t *testing.T) {
	v := sdui.Var("volume", sdui.VarNumber, nil)
	if _, present := v["value"]; present {
		t.Errorf("a nil initial value became a key: %#v", v)
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != `{"name":"volume","type":"number"}` {
		t.Errorf("on the wire = %s", b)
	}
}

func TestVarsIsThePropsValue(t *testing.T) {
	vs := sdui.Vars(
		sdui.Var("username", sdui.VarString, ""),
		sdui.Var("remember", sdui.VarBoolean, true),
	)
	if len(vs) != 2 {
		t.Fatalf("Vars = %#v", vs)
	}
	// It must survive the props bag's JSON round trip unchanged, since that is
	// how it reaches a client.
	b, err := json.Marshal(vs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back []map[string]any
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back[1]["type"] != "boolean" || back[1]["value"] != true {
		t.Errorf("round trip lost a declaration: %#v", back[1])
	}
}

// The type set is closed, and these are its members. A fourth would be a
// coercion every client has to implement identically.
func TestTheVarTypeSetIsThree(t *testing.T) {
	if sdui.VarString != "string" || sdui.VarNumber != "number" || sdui.VarBoolean != "boolean" {
		t.Error("a VarType constant does not match its wire spelling")
	}
}

// Submit's destination is where the scope's values land in the action's input.
// Without one they could only ever land at the top level, which is why `$value`
// — a literal substituted anywhere in the action — outlived every attempt to
// replace it: a module's configureModule takes a whole settings document, and
// the field that was typed belongs inside it.
func TestSubmitCarriesItsDestination(t *testing.T) {
	a := sdui.Submit(sdui.Invoke("configureModule", map[string]any{"moduleId": "tmdb"}), "settings")
	if a.Kind != sdui.KindSubmit || len(a.Actions) != 1 {
		t.Fatalf("Submit = %+v", a)
	}
	if a.Field == nil || *a.Field != "settings" {
		t.Errorf("destination = %v", a.Field)
	}
	// An empty destination is absent rather than an empty string, so a client
	// reading it cannot mistake "the top level" for "a field called nothing".
	if b := sdui.Submit(sdui.Back(), ""); b.Field != nil {
		t.Errorf("an empty destination became a field: %v", b.Field)
	}
}

// SubmitField is gone. It existed only to carry `$value`, and a primitive that
// exists for a mechanism outlives the mechanism only by accident.
func TestSubmitFieldIsNoLongerAPrimitive(t *testing.T) {
	for _, p := range sdui.Primitives {
		if p.Type == "SubmitField" {
			t.Error("SubmitField is still declared; it exists only for the $value substitution")
		}
	}
}
