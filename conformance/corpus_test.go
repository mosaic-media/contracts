// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package conformance_test

import (
	"encoding/json"
	"testing"

	"github.com/mosaic-media/contracts/conformance"
	"github.com/mosaic-media/contracts/sdui"
)

// The corpus, run against the Go implementation.
//
// A corpus that only one implementation executes is a test. These same files are
// run by the web client's own runner, which is the whole point: two
// implementations, one set of expected answers, and any disagreement shows up as
// a named failing case rather than as a screen behaving differently on one
// platform.

func TestValidatorCorpus(t *testing.T) {
	cases, err := conformance.LoadValidatorCases()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(cases) < 15 {
		t.Fatalf("the validator corpus has shrunk to %d cases", len(cases))
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var got string
			if e := sdui.ValidateField("f", c.Value, c.Rules, c.Values); e != nil {
				got = e.Message
			}
			if got != c.Expect {
				t.Errorf("got %q, want %q", got, c.Expect)
			}
		})
	}
}

func TestPredicateCorpus(t *testing.T) {
	doc, err := conformance.LoadPredicateCases()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(doc.Cases) < 20 {
		t.Fatalf("the predicate corpus has shrunk to %d cases", len(doc.Cases))
	}
	for _, c := range doc.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := sdui.Evaluate(c.Predicate, doc.Values); got != c.Expect {
				t.Errorf("got %v, want %v", got, c.Expect)
			}
		})
	}
}

func TestBindingCorpus(t *testing.T) {
	doc, err := conformance.LoadBindingCases()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(doc.Cases) < 8 {
		t.Fatalf("the binding corpus has shrunk to %d cases", len(doc.Cases))
	}
	for _, c := range doc.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got := sdui.ResolveProps(c.Props, doc.Scope)
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(c.Expect)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("got %s, want %s", gotJSON, wantJSON)
			}
		})
	}
}

// The expansion corpus is published and not executed here: this repository has
// no definition expander, and writing a second one in Go purely to run the
// corpus would be a second implementation to keep in step — which is the drift
// the corpus exists to catch, manufactured to catch itself.
//
// What is checked here is that the file is present and well-formed, so it cannot
// rot into invalid JSON unnoticed between the clients that do run it.
func TestExpansionCorpusIsWellFormed(t *testing.T) {
	raw, err := conformance.Cases("expansion")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var doc struct {
		Cases []struct {
			Name       string         `json:"name"`
			Definition map[string]any `json:"definition"`
			Node       map[string]any `json:"node"`
			Expect     map[string]any `json:"expect"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(doc.Cases) < 10 {
		t.Fatalf("the expansion corpus has shrunk to %d cases", len(doc.Cases))
	}
	for _, c := range doc.Cases {
		if c.Name == "" || c.Definition == nil || c.Expect == nil {
			t.Errorf("case %q is incomplete", c.Name)
		}
	}
}

// The submit corpus, likewise published and not executed here: the merge is a
// client dispatcher's behaviour and this repository has no dispatcher. Writing
// one in Go to run the corpus is the same manufactured drift as an expander.
//
// It is worth saying why this file exists at all, since the rule in it looks
// like a detail. The merge decided whose value wins when a form and its action
// both name a field, it got that backwards, and it lived inline in the
// dispatcher where nothing could call it. The symptom was a settings form that
// reported success and saved nothing (ADR 0096). It was found by filling the
// form in, which is the only way it could have been found.
func TestSubmitCorpusIsWellFormed(t *testing.T) {
	raw, err := conformance.Cases("submit")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var doc struct {
		Cases []struct {
			Name   string         `json:"name"`
			Action map[string]any `json:"action"`
			Values map[string]any `json:"values"`
			Into   string         `json:"into"`
			Expect map[string]any `json:"expect"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(doc.Cases) < 8 {
		t.Fatalf("the submit corpus has shrunk to %d cases", len(doc.Cases))
	}
	for _, c := range doc.Cases {
		if c.Name == "" || c.Action == nil || c.Expect == nil {
			t.Errorf("case %q is incomplete", c.Name)
		}
	}
}

// Every corpus file is reachable through the embed, so one added to the
// directory and forgotten in the loader is a failure rather than a file nobody
// runs.
func TestEveryCorpusFileIsAccountedFor(t *testing.T) {
	files := conformance.CaseFiles()
	if len(files) != 5 {
		t.Errorf("corpus files = %v; the runners cover five", files)
	}
}
