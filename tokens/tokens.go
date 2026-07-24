// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

// Package tokens is the design token set as data — every literal value the
// interface is built from, in one place, for every client.
//
// It is the skin half of what the SDUI contract carries: [definitions] say what
// a screen is made of, these say what it looks like. The Platform serves both
// over the session, so a re-skin reaches a running client without a client
// release — which is the whole point of ADR 0040's UI-library tier and the half
// of it that had never been built. The client's own stylesheet held the values
// and this file held a stale copy of some of them.
//
// What is not here: anything DERIVED from a token — a colour mixed from the
// accent, a gradient composed of two tokens, the acrylic material's layers.
// Those are formulas, and a formula belongs to the client that evaluates it.
package tokens

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed tokens.json
var raw []byte

// Set returns the token document as served: `{base, themes:{dark,light}}`, each
// leaf a DTCG `{$value,$type}`. Bytes are returned rather than a typed tree
// because a client applies them verbatim; the Platform is a courier here, not an
// interpreter.
func Set() []byte { return raw }

// Validate reports whether the embedded set parses and carries the groups a
// client needs. It runs in the Platform's start-up path: a token set that a
// client cannot apply leaves the whole interface unstyled, which is a failure
// worth catching at boot rather than in a browser.
func Validate() error {
	var doc struct {
		Base   map[string]json.RawMessage `json:"base"`
		Themes struct {
			Dark  map[string]json.RawMessage `json:"dark"`
			Light map[string]json.RawMessage `json:"light"`
		} `json:"themes"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("token set is not valid JSON: %w", err)
	}
	if len(doc.Base) == 0 {
		return fmt.Errorf("token set carries no base scale")
	}
	if len(doc.Themes.Dark) == 0 {
		return fmt.Errorf("token set carries no dark theme")
	}
	return nil
}
