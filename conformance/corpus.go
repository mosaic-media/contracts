// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

package conformance

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
)

// The corpus: golden cases every client runs to prove it behaves like the
// others (contracts#17).
//
// The vocabulary fixture beside this says *what* the contract contains. It does
// not say how any of it behaves, and behaviour is where clients diverge — two
// implementations of "resolve a binding" agree about the easy cases and differ
// about the empty string, the missing path and the object that merely looks like
// a binding. Those are the cases here.
//
// It is the mechanism rather than the content that matters: this is how DivKit
// holds four platforms in parity, and it is what makes a second Mosaic client a
// tractable piece of work rather than an act of faith. A client that passes the
// corpus behaves like the one that exists; a client that does not knows exactly
// where it differs.

//go:embed cases/*.json
var caseFiles embed.FS

// Cases returns one corpus file's raw bytes, for a runner in any language that
// would rather decode it itself.
func Cases(name string) ([]byte, error) { return caseFiles.ReadFile("cases/" + name + ".json") }

// CaseFiles lists the corpus files present.
func CaseFiles() []string {
	entries, _ := fs.Glob(caseFiles, "cases/*.json")
	return entries
}

// ValidatorCase is one golden validator case.
type ValidatorCase struct {
	Name   string         `json:"name"`
	Rules  map[string]any `json:"rules"`
	Value  any            `json:"value"`
	Values map[string]any `json:"values"`
	Expect string         `json:"expect"`
}

// PredicateCases is the predicate corpus: one shared set of values, many cases.
type PredicateCases struct {
	Values map[string]any `json:"values"`
	Cases  []struct {
		Name      string         `json:"name"`
		Predicate map[string]any `json:"predicate"`
		Expect    bool           `json:"expect"`
	} `json:"cases"`
}

// BindingCases is the binding corpus: one shared scope, many prop bags.
type BindingCases struct {
	Scope map[string]any `json:"scope"`
	Cases []struct {
		Name   string         `json:"name"`
		Props  map[string]any `json:"props"`
		Expect map[string]any `json:"expect"`
	} `json:"cases"`
}

// LoadValidatorCases decodes the validator corpus.
func LoadValidatorCases() ([]ValidatorCase, error) {
	raw, err := Cases("validators")
	if err != nil {
		return nil, err
	}
	var doc struct {
		Cases []ValidatorCase `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("validator corpus: %w", err)
	}
	return doc.Cases, nil
}

// LoadPredicateCases decodes the predicate corpus.
func LoadPredicateCases() (PredicateCases, error) {
	raw, err := Cases("predicates")
	if err != nil {
		return PredicateCases{}, err
	}
	var doc PredicateCases
	if err := json.Unmarshal(raw, &doc); err != nil {
		return PredicateCases{}, fmt.Errorf("predicate corpus: %w", err)
	}
	return doc, nil
}

// LoadBindingCases decodes the binding corpus.
func LoadBindingCases() (BindingCases, error) {
	raw, err := Cases("bindings")
	if err != nil {
		return BindingCases{}, err
	}
	var doc BindingCases
	if err := json.Unmarshal(raw, &doc); err != nil {
		return BindingCases{}, fmt.Errorf("binding corpus: %w", err)
	}
	return doc, nil
}
