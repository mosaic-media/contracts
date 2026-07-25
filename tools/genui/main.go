// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

// Command genui generates the SDUI vocabulary from the single spec
// ui.spec.json: the Go (ui/components.gen.go) and TypeScript (ts/ui.ts)
// ergonomic constructors, the machine-readable vocabulary registry
// (sdui/vocabulary.gen.go, ts/vocabulary.gen.ts) and the client conformance
// fixture (conformance/vocabulary.json).
//
// The spec has three tiers and one tool emits all three. `primitives` are the
// native vocabulary a client must implement (growing the set needs a client
// release — ADR 0024), `components` are definitions delivered as data
// (definitions/*.json), and `actions` are the behaviours a client interprets.
// Before this tool knew about primitives that tier existed only as TypeScript
// in one client, so a second client could not be written from the published
// contract at all; that is the drift these lint gates exist to keep closed.
//
// The ui layer sits on a small hand-written runtime (ui/element.go,
// ts/ui_runtime.ts); this tool emits only the per-type constructors, the typed
// sugar, the slot helpers and the action/tone re-exports — the mechanical part —
// so adding a primitive or a component is a spec edit, and Go/TS can never
// drift.
//
// Usage:
//
//	go run ./tools/genui         # regenerate the files, then lint
//	go run ./tools/genui -check  # verify the files are up to date (CI), then lint
//	go run ./tools/genui -lint   # lint the spec, definitions, schema and proto
//
// Lint enforces coverage in both directions:
//
//   - every definition has a component in the spec and every component has a
//     definition file;
//   - every prop a definition's template binds is exposed by some helper, and
//     every Outlet it declares has a slot helper;
//   - every node type a definition's template references exists in the
//     vocabulary (referenced-type existence);
//   - no type is both a primitive and a component (tier overlap);
//   - every primitive carries a `native` justification for why a definition
//     cannot express it (per-primitive justification);
//   - the spec's action kinds, tones and surfaces are exactly those in
//     schema/sdui.schema.json and proto/mosaic/sdui/v1/sdui.proto
//     (action-kind coverage — the three-vocabulary drift, mechanised).
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// ── spec model ───────────────────────────────────────────────────────────────

type spec struct {
	Version       string      `json:"version"`
	TypeSeparator string      `json:"typeSeparator"`
	BindingMarker string      `json:"bindingMarker"`
	Tones         []tone      `json:"tones"`
	Surfaces      []surface   `json:"surfaces"`
	Actions       []action    `json:"actions"`
	Validators    []validator `json:"validators"`
	Predicates    []predicate `json:"predicates"`
	Primitives    []primitive `json:"primitives"`
	Sugar         []sugar     `json:"sugar"`
	Slots         []slot      `json:"slots"`
	Components    []component `json:"components"`
}

type tone struct {
	Const string `json:"const"`
	Sdui  string `json:"sdui"`
	// TS names the member of the generated TypeScript Tone enum. It is stated
	// rather than derived from Value: the derivation is quicktype's, and a
	// generator guessing another generator's naming rule is a drift waiting to
	// happen.
	TS    string `json:"ts"`
	Value string `json:"value"`
}

type surface struct {
	Const string `json:"const"`
	Value string `json:"value"`
}

type validator struct {
	Name string `json:"name"`
	Arg  string `json:"arg"`
	Doc  string `json:"doc"`
}

type predicate struct {
	Name string `json:"name"`
	Doc  string `json:"doc"`
}

// primitive is one entry of the native tier. Native is load-bearing rather than
// decorative: it is the justification for why this cannot be a definition, and
// lint rejects a primitive without one — because the cost of the tier is a
// client release per addition, and an unjustified addition spends that silently.
type primitive struct {
	Func       string `json:"func"`
	Type       string `json:"type"`
	Tier       string `json:"tier"`
	Doc        string `json:"doc"`
	Native     string `json:"native"`
	Children   bool   `json:"children"`
	Positional []arg  `json:"positional"`
	Props      []prop `json:"props"`
}

// fn is the exported constructor name — the type name unless the spec overrides
// it, which it does only to break a collision with a sugar helper.
func (p primitive) fn() string {
	if p.Func != "" {
		return p.Func
	}
	return p.Type
}

type prop struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	Doc  string `json:"doc"`
}

type arg struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Field string `json:"field"`
	Key   string `json:"key"`
}

type action struct {
	Func     string `json:"func"`
	Sdui     string `json:"sdui"`
	Kind     string `json:"kind"`
	Enum     string `json:"enum"`
	Doc      string `json:"doc"`
	Args     []arg  `json:"args"`
	Optional []arg  `json:"optional"`
}

type sugar struct {
	Func string `json:"func"`
	Key  string `json:"key"`
	Type string `json:"type"`
	Doc  string `json:"doc"`
}

type slot struct {
	Func string `json:"func"`
	Name string `json:"name"`
	Doc  string `json:"doc"`
}

type component struct {
	Func       string `json:"func"`
	Type       string `json:"type"`
	Doc        string `json:"doc"`
	Positional []arg  `json:"positional"`
	Variadic   *bool  `json:"variadic"`
	Generic    bool   `json:"generic"`
}

func (c component) variadic() bool { return c.Variadic == nil || *c.Variadic }

func main() {
	var (
		root  = "."
		check bool
		lint  bool
	)
	for _, a := range os.Args[1:] {
		switch a {
		case "-check":
			check = true
		case "-lint":
			lint = true
		default:
			if v, ok := strings.CutPrefix(a, "-root="); ok {
				root = v
			} else {
				fatalf("unknown argument %q", a)
			}
		}
	}

	sp := loadSpec(filepath.Join(root, "ui.spec.json"))

	// The checks generation itself depends on, run before it. A type name is a
	// Go and TypeScript function name, so a malformed one does not reach the
	// lint at the end — it fails gofmt first and dumps the whole generated file
	// to stderr, which reads as a generator bug rather than as the spec error it
	// is.
	if errs := lintSpecSanity(sp); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "lint: %s\n", e)
		}
		os.Exit(1)
	}

	outputs := []struct {
		path string
		want []byte
	}{
		{filepath.Join(root, "ui", "components.gen.go"), genGo(sp)},
		{filepath.Join(root, "ts", "ui.ts"), genTS(sp)},
		{filepath.Join(root, "sdui", "vocabulary.gen.go"), genVocabularyGo(sp)},
		{filepath.Join(root, "ts", "vocabulary.gen.ts"), genVocabularyTS(sp)},
		{filepath.Join(root, "conformance", "vocabulary.json"), genFixture(sp)},
	}

	switch {
	case lint:
		// lint only
	case check:
		stale := false
		for _, f := range outputs {
			got, err := os.ReadFile(f.path)
			if err != nil || !bytes.Equal(normalize(got), normalize(f.want)) {
				fmt.Fprintf(os.Stderr, "stale: %s (run `go run ./tools/genui`)\n", f.path)
				stale = true
			}
		}
		if stale {
			os.Exit(1)
		}
		fmt.Println("genui: generated files are up to date")
	default:
		names := make([]string, 0, len(outputs))
		for _, f := range outputs {
			if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
				fatalf("mkdir %s: %v", filepath.Dir(f.path), err)
			}
			writeFile(f.path, f.want)
			names = append(names, f.path)
		}
		fmt.Printf("genui: wrote %s\n", strings.Join(names, ", "))
	}

	if errs := runLint(sp, root); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "lint: %s\n", e)
		}
		os.Exit(1)
	}
	fmt.Println("genui: lint ok")
}

// ── Go generation ────────────────────────────────────────────────────────────

func genGo(sp spec) []byte {
	var b strings.Builder
	b.WriteString("// Code generated by tools/genui from ui.spec.json. DO NOT EDIT.\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n")
	b.WriteString("// SPDX-FileCopyrightText: 2026 the Mosaic authors\n\n")
	b.WriteString("package ui\n\n")
	if len(sp.Tones) > 0 || len(sp.Actions) > 0 {
		b.WriteString("import \"github.com/mosaic-media/contracts/sdui\"\n\n")
	}

	b.WriteString("// ── primitives ─────────────────────────────────────────────────────────────\n")
	b.WriteString("// The native tier: node types a client implements itself. They are authored\n")
	b.WriteString("// here rather than through Component(\"Box\", …) because a type spelled as a\n")
	b.WriteString("// string is a type nothing checks — which is how seven primitives came to be\n")
	b.WriteString("// emitted by name with no entry in any contract.\n\n")
	for _, p := range sp.Primitives {
		if p.Doc != "" {
			fmt.Fprintf(&b, "// %s %s\n", p.fn(), dropLeadName(p.fn(), p.Doc))
		}
		fmt.Fprintf(&b, "// Native: %s\n", p.Native)
		// Always variadic, including for a leaf: the ...El tail carries props,
		// ids and slots as well as children, so a leaf that took none could not
		// be given a prop at all.
		params := []string{}
		for _, a := range p.Positional {
			params = append(params, a.Name+" "+goType(a.Type))
		}
		params = append(params, "els ...El")
		fmt.Fprintf(&b, "func %s(%s) *Element { return compose(%s, %s, els) }\n\n",
			p.fn(), strings.Join(params, ", "), strconv.Quote(p.Type), goBase(p.Positional))
	}

	b.WriteString("// ── components ─────────────────────────────────────────────────────────────\n\n")
	for _, c := range sp.Components {
		if c.Doc != "" {
			fmt.Fprintf(&b, "// %s %s\n", c.Func, dropLeadName(c.Func, c.Doc))
		}
		params := []string{}
		if c.Generic {
			params = append(params, "typ string")
		}
		for _, p := range c.Positional {
			params = append(params, p.Name+" "+goType(p.Type))
		}
		if c.variadic() {
			params = append(params, "els ...El")
		}
		typExpr := strconv.Quote(c.Type)
		if c.Generic {
			typExpr = "typ"
		}
		elsArg := "nil"
		if c.variadic() {
			elsArg = "els"
		}
		fmt.Fprintf(&b, "func %s(%s) *Element { return compose(%s, %s, %s) }\n\n",
			c.Func, strings.Join(params, ", "), typExpr, goBase(c.Positional), elsArg)
	}

	b.WriteString("// ── slots ──────────────────────────────────────────────────────────────────\n\n")
	for _, s := range sp.Slots {
		if s.Doc != "" {
			fmt.Fprintf(&b, "// %s %s\n", s.Func, dropLeadName(s.Func, s.Doc))
		}
		fmt.Fprintf(&b, "func %s(els ...El) El { return Slot(%s, els...) }\n\n", s.Func, strconv.Quote(s.Name))
	}

	b.WriteString("// ── sugar ──────────────────────────────────────────────────────────────────\n\n")
	for _, s := range sp.Sugar {
		if s.Doc != "" {
			fmt.Fprintf(&b, "// %s %s\n", s.Func, dropLeadName(s.Func, s.Doc))
		}
		fmt.Fprintf(&b, "func %s(%s) El { return Prop(%s, v) }\n\n", s.Func, goSugarParam(s.Type), strconv.Quote(s.Key))
	}

	b.WriteString("// ── bound sugar ────────────────────────────────────────────────────────────\n")
	b.WriteString("// The same props, set to a binding the client resolves where the node renders\n")
	b.WriteString("// rather than to a value decided now. One per helper, generated, because a\n")
	b.WriteString("// prop set by string is the failure this contract keeps having: ui.Subtitle on\n")
	b.WriteString("// a Stack drew nothing for a screen's whole life, and Prop(\"title\", Bind(…))\n")
	b.WriteString("// would put every bound prop back on that footing.\n\n")
	for _, s := range sp.Sugar {
		fmt.Fprintf(&b, "// Bind%s sets %s from the named path instead of from a value.\n", s.Func, strconv.Quote(s.Key))
		fmt.Fprintf(&b, "func Bind%s(path string) El { return Prop(%s, sdui.Bind(path)) }\n\n", s.Func, strconv.Quote(s.Key))
	}

	if len(sp.Tones) > 0 {
		b.WriteString("// Tone values (the open-bag string encoding), re-exported from the producer binding.\n")
		b.WriteString("const (\n")
		for _, t := range sp.Tones {
			fmt.Fprintf(&b, "\t%s = sdui.%s\n", t.Const, t.Sdui)
		}
		b.WriteString(")\n\n")
	}

	if len(sp.Actions) > 0 {
		b.WriteString("// Action constructors, re-exported from the producer binding; they ride the open\n")
		b.WriteString("// props bag as JSON (ADR 0044).\n")
		b.WriteString("var (\n")
		for _, a := range sp.Actions {
			fmt.Fprintf(&b, "\t%s = sdui.%s\n", a.Func, a.Sdui)
		}
		b.WriteString(")\n")
	}

	out, err := format.Source([]byte(b.String()))
	if err != nil {
		fmt.Fprintln(os.Stderr, b.String())
		fatalf("gofmt generated Go: %v", err)
	}
	return out
}

func goType(t string) string {
	switch t {
	case "string":
		return "string"
	case "int":
		return "int"
	case "number":
		return "float64"
	case "Action":
		return "Action"
	case "bool":
		return "bool"
	case "props":
		return "map[string]any"
	case "[]props":
		return "[]any"
	case "Tone", "Surface":
		// Both are string aliases in the Go binding (they ride the open props
		// bag as JSON), so the constants remain assignable.
		return "string"
	default:
		fatalf("unknown Go type %q", t)
		return ""
	}
}

// goSugarParam renders the single value parameter of a sugar helper, variadic
// for list types.
func goSugarParam(t string) string {
	if t == "[]string" {
		return "v ...string"
	}
	return "v " + goType(t)
}

func goBase(pos []arg) string {
	if len(pos) == 0 {
		return "nil"
	}
	parts := make([]string, len(pos))
	for i, p := range pos {
		parts[i] = fmt.Sprintf("%s: %s", strconv.Quote(p.Key), p.Name)
	}
	return "map[string]any{" + strings.Join(parts, ", ") + "}"
}

// ── TypeScript generation ────────────────────────────────────────────────────

func genTS(sp spec) []byte {
	var b strings.Builder
	b.WriteString("// Code generated by tools/genui from ui.spec.json. DO NOT EDIT.\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n")
	b.WriteString("// SPDX-FileCopyrightText: 2026 the Mosaic authors\n\n")
	b.WriteString("import { ActionKind, Tone, type Action, type Surface } from \"./contract.gen.js\";\n")
	b.WriteString("import { compose, Prop, Slot, type El, type Elish, type Element, type Props } from \"./ui_runtime.js\";\n")
	b.WriteString("import { bind } from \"./binding.js\";\n\n")
	b.WriteString("export { Group, ID, Prop, Slot, When, Element } from \"./ui_runtime.js\";\n")
	b.WriteString("export type { El, Elish, Props } from \"./ui_runtime.js\";\n")
	b.WriteString("export { ActionKind, Surface, Tone } from \"./contract.gen.js\";\n")
	b.WriteString("export { bind, isBinding, bindingPath, bindingMarker } from \"./binding.js\";\n")
	b.WriteString("export type { Binding } from \"./binding.js\";\n")
	b.WriteString("export type { Action, UINode } from \"./contract.gen.js\";\n\n")

	b.WriteString("// ── primitives ─────────────────────────────────────────────────────────────\n")
	b.WriteString("// The native tier: node types a client implements itself (ADR 0024).\n\n")
	for _, p := range sp.Primitives {
		fmt.Fprintf(&b, "/** %s\n *\n *  Native: %s */\n", p.Doc, p.Native)
		params := []string{}
		for _, a := range p.Positional {
			params = append(params, a.Name+": "+tsType(a.Type))
		}
		params = append(params, "...els: Elish[]")
		fmt.Fprintf(&b, "export function %s(%s): Element {\n  return compose(%s, %s, els);\n}\n\n",
			p.fn(), strings.Join(params, ", "), strconv.Quote(p.Type), tsBase(p.Positional))
	}

	b.WriteString("// ── components ─────────────────────────────────────────────────────────────\n\n")
	for _, c := range sp.Components {
		if c.Doc != "" {
			fmt.Fprintf(&b, "/** %s */\n", c.Doc)
		}
		params := []string{}
		if c.Generic {
			params = append(params, "typ: string")
		}
		for _, p := range c.Positional {
			params = append(params, p.Name+": "+tsType(p.Type))
		}
		if c.variadic() {
			params = append(params, "...els: Elish[]")
		}
		typExpr := strconv.Quote(c.Type)
		if c.Generic {
			typExpr = "typ"
		}
		elsArg := "[]"
		if c.variadic() {
			elsArg = "els"
		}
		fmt.Fprintf(&b, "export function %s(%s): Element {\n  return compose(%s, %s, %s);\n}\n\n",
			c.Func, strings.Join(params, ", "), typExpr, tsBase(c.Positional), elsArg)
	}

	b.WriteString("// ── slots ──────────────────────────────────────────────────────────────────\n\n")
	for _, s := range sp.Slots {
		if s.Doc != "" {
			fmt.Fprintf(&b, "/** %s */\n", s.Doc)
		}
		fmt.Fprintf(&b, "export function %s(...els: Elish[]): El {\n  return Slot(%s, ...els);\n}\n\n", s.Func, strconv.Quote(s.Name))
	}

	b.WriteString("// ── sugar ──────────────────────────────────────────────────────────────────\n\n")
	for _, s := range sp.Sugar {
		if s.Doc != "" {
			fmt.Fprintf(&b, "/** %s */\n", s.Doc)
		}
		fmt.Fprintf(&b, "export function %s(%s): El {\n  return Prop(%s, v);\n}\n\n", s.Func, tsSugarParam(s.Type), strconv.Quote(s.Key))
	}

	b.WriteString("// ── bound sugar ────────────────────────────────────────────────────────────\n")
	b.WriteString("// The same props, set to a binding the client resolves where the node renders.\n\n")
	for _, s := range sp.Sugar {
		fmt.Fprintf(&b, "/** Bind%s sets %s from the named path instead of from a value. */\n", s.Func, strconv.Quote(s.Key))
		fmt.Fprintf(&b, "export function Bind%s(path: string): El {\n  return Prop(%s, bind(path));\n}\n\n", s.Func, strconv.Quote(s.Key))
	}

	if len(sp.Tones) > 0 {
		b.WriteString("// Tone values, mirroring the Go Tone constants. They are the schema enum's\n")
		b.WriteString("// members rather than bare strings, so passing one where the contract wants a\n")
		b.WriteString("// Tone typechecks.\n")
		for _, t := range sp.Tones {
			fmt.Fprintf(&b, "export const %s = Tone.%s;\n", t.Const, t.TS)
		}
		b.WriteString("\n")
	}

	if len(sp.Actions) > 0 {
		b.WriteString("// ── actions ────────────────────────────────────────────────────────────────\n")
		b.WriteString("// Actions ride inside the open props bag as JSON (ADR 0044); each constructor\n")
		b.WriteString("// hides the discriminator and the optional fields.\n\n")
		for _, a := range sp.Actions {
			if a.Doc != "" {
				fmt.Fprintf(&b, "/** %s */\n", a.Doc)
			}
			params := make([]string, 0, len(a.Args)+len(a.Optional))
			for _, ar := range a.Args {
				params = append(params, ar.Name+": "+tsType(ar.Type))
			}
			for _, op := range a.Optional {
				params = append(params, op.Name+"?: "+tsType(op.Type))
			}
			fields := make([]string, 0, len(a.Args)+1)
			fields = append(fields, "kind: ActionKind."+a.Enum)
			for _, ar := range a.Args {
				fields = append(fields, tsField(ar.Field, ar.Name))
			}
			body := "{ " + strings.Join(fields, ", ")
			for _, op := range a.Optional {
				body += fmt.Sprintf(", ...(%s ? { %s } : {})", op.Name, tsField(op.Field, op.Name))
			}
			body += " }"
			fmt.Fprintf(&b, "export function %s(%s): Action {\n  return %s;\n}\n\n", a.Func, strings.Join(params, ", "), body)
		}
	}

	return []byte(strings.TrimRight(b.String(), "\n") + "\n")
}

func tsType(t string) string {
	switch t {
	case "string":
		return "string"
	case "int", "number":
		return "number"
	case "Action":
		return "Action"
	case "bool":
		return "boolean"
	case "props":
		return "Props"
	case "[]props":
		return "Props[]"
	case "Tone", "Surface":
		return t
	default:
		fatalf("unknown TS type %q", t)
		return ""
	}
}

func tsSugarParam(t string) string {
	if t == "[]string" {
		return "...v: string[]"
	}
	return "v: " + tsType(t)
}

func tsBase(pos []arg) string {
	if len(pos) == 0 {
		return "undefined"
	}
	parts := make([]string, len(pos))
	for i, p := range pos {
		parts[i] = tsField(p.Key, p.Name)
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// tsField renders an object entry, using shorthand when the key equals the name.
func tsField(key, name string) string {
	if key == name {
		return name
	}
	return strconv.Quote(key) + ": " + name
}

// ── vocabulary registry ──────────────────────────────────────────────────────

// genVocabularyGo emits the machine-readable registry: the node-type constants
// (which replaced a hand-written list that had drifted into naming three
// primitives as components and omitting every real one), the action/tone/surface
// enum values, and the primitive tier as data so a consumer can ask what the
// contract contains rather than reading a client's source.
func genVocabularyGo(sp spec) []byte {
	var b strings.Builder
	b.WriteString("// Code generated by tools/genui from ui.spec.json. DO NOT EDIT.\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n")
	b.WriteString("// SPDX-FileCopyrightText: 2026 the Mosaic authors\n\n")
	b.WriteString("package sdui\n\n")

	fmt.Fprintf(&b, "// VocabularyVersion is the version a client declares it implements. Additive\n")
	fmt.Fprintf(&b, "// growth is a minor bump; removing or changing the meaning of a primitive, a\n")
	fmt.Fprintf(&b, "// prop or an action is a major one.\n")
	fmt.Fprintf(&b, "const VocabularyVersion = %s\n\n", strconv.Quote(sp.Version))

	b.WriteString("// TypeSeparator divides a module's id from its own type name. Core types are\n")
	b.WriteString("// unprefixed and may never contain it; a module's are moduleId:type. Two\n")
	b.WriteString("// modules could otherwise both call a component StatChip, and one could call\n")
	b.WriteString("// it PosterCard and take the core component's place.\n")
	fmt.Fprintf(&b, "const TypeSeparator = %s\n\n", strconv.Quote(sp.TypeSeparator))

	b.WriteString("// BindingMarker is the single key that makes a prop value a binding rather\n")
	b.WriteString("// than a literal. It is spelled as a definition template's binding is,\n")
	b.WriteString("// because it means the same thing — resolve a path against a scope.\n")
	fmt.Fprintf(&b, "const BindingMarker = %s\n\n", strconv.Quote(sp.BindingMarker))

	b.WriteString("// Node type names — the primitive tier.\n")
	b.WriteString("const (\n")
	for _, p := range sp.Primitives {
		fmt.Fprintf(&b, "\tType%s = %s\n", p.Type, strconv.Quote(p.Type))
	}
	b.WriteString(")\n\n")

	b.WriteString("// Node type names — the component tier (definitions/*.json).\n")
	b.WriteString("const (\n")
	for _, c := range sp.Components {
		if c.Generic || c.Type == "" {
			continue
		}
		fmt.Fprintf(&b, "\tType%s = %s\n", c.Type, strconv.Quote(c.Type))
	}
	b.WriteString(")\n\n")

	b.WriteString("// Action kinds — the JSON discriminator values.\n")
	b.WriteString("const (\n")
	for _, a := range sp.Actions {
		fmt.Fprintf(&b, "\tKind%s = %s\n", a.Enum, strconv.Quote(a.Kind))
	}
	b.WriteString(")\n\n")

	b.WriteString("// Tones.\n")
	b.WriteString("const (\n")
	for _, t := range sp.Tones {
		fmt.Fprintf(&b, "\t%s = %s\n", t.Const, strconv.Quote(t.Value))
	}
	b.WriteString(")\n\n")

	b.WriteString("// Overlay surfaces.\n")
	b.WriteString("const (\n")
	for _, s := range sp.Surfaces {
		fmt.Fprintf(&b, "\t%s = %s\n", s.Const, strconv.Quote(s.Value))
	}
	b.WriteString(")\n\n")

	b.WriteString("// Primitives is the native tier as data: what a client must implement, and\n")
	b.WriteString("// for each one the reason it cannot be a definition.\n")
	b.WriteString("var Primitives = []PrimitiveSpec{\n")
	for _, p := range sp.Primitives {
		fmt.Fprintf(&b, "\t{Type: %s, Tier: %s, Doc: %s, Native: %s, Children: %t, Props: []PropSpec{\n",
			strconv.Quote(p.Type), strconv.Quote(p.Tier), strconv.Quote(p.Doc), strconv.Quote(p.Native), p.Children)
		for _, pr := range p.Props {
			fmt.Fprintf(&b, "\t\t{Key: %s, Type: %s, Doc: %s},\n",
				strconv.Quote(pr.Key), strconv.Quote(pr.Type), strconv.Quote(pr.Doc))
		}
		b.WriteString("\t}},\n")
	}
	b.WriteString("}\n\n")

	b.WriteString("// Components is the definition tier: the type names the Platform serves as\n")
	b.WriteString("// data. A client implements none of them.\n")
	b.WriteString("var Components = []string{\n")
	for _, c := range sp.Components {
		if c.Generic || c.Type == "" {
			continue
		}
		fmt.Fprintf(&b, "\t%s,\n", strconv.Quote(c.Type))
	}
	b.WriteString("}\n\n")

	b.WriteString("// ActionKinds is every behaviour a client interprets.\n")
	b.WriteString("var ActionKinds = []ActionSpec{\n")
	for _, a := range sp.Actions {
		fmt.Fprintf(&b, "\t{Kind: %s, Doc: %s},\n", strconv.Quote(a.Kind), strconv.Quote(a.Doc))
	}
	b.WriteString("}\n\n")

	b.WriteString("// Validators is the closed field-validation set. Closed on purpose: an open\n")
	b.WriteString("// set lets the server state a rule the client cannot enforce, which fails open.\n")
	b.WriteString("var Validators = []ValidatorSpec{\n")
	for _, v := range sp.Validators {
		fmt.Fprintf(&b, "\t{Name: %s, Arg: %s, Doc: %s},\n",
			strconv.Quote(v.Name), strconv.Quote(v.Arg), strconv.Quote(v.Doc))
	}
	b.WriteString("}\n\n")

	b.WriteString("// Predicates is the closed conditional set — the deliberate alternative to an\n")
	b.WriteString("// expression language and the evaluator VM one would need.\n")
	b.WriteString("var Predicates = []PredicateSpec{\n")
	for _, p := range sp.Predicates {
		fmt.Fprintf(&b, "\t{Name: %s, Doc: %s},\n", strconv.Quote(p.Name), strconv.Quote(p.Doc))
	}
	b.WriteString("}\n")

	out, err := format.Source([]byte(b.String()))
	if err != nil {
		fmt.Fprintln(os.Stderr, b.String())
		fatalf("gofmt generated vocabulary: %v", err)
	}
	return out
}

// genVocabularyTS emits the same registry for a JavaScript consumer — a client
// checking what it implements, or the storybook enumerating the vocabulary.
func genVocabularyTS(sp spec) []byte {
	var b strings.Builder
	b.WriteString("// Code generated by tools/genui from ui.spec.json. DO NOT EDIT.\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n")
	b.WriteString("// SPDX-FileCopyrightText: 2026 the Mosaic authors\n\n")
	b.WriteString("/** One prop a primitive reads. */\n")
	b.WriteString("export interface PropSpec {\n  key: string;\n  type: string;\n  doc: string;\n}\n\n")
	b.WriteString("/** One native node type, with the reason it cannot be a definition. */\n")
	b.WriteString("export interface PrimitiveSpec {\n  type: string;\n  tier: string;\n  doc: string;\n  native: string;\n  children: boolean;\n  props: PropSpec[];\n}\n\n")

	fmt.Fprintf(&b, "/** The vocabulary version a client declares it implements. */\nexport const vocabularyVersion = %s;\n\n", strconv.Quote(sp.Version))

	fmt.Fprintf(&b, "/** Divides a module's id from its own type name; core types never contain it. */\nexport const typeSeparator = %s;\n\n", strconv.Quote(sp.TypeSeparator))

	fmt.Fprintf(&b, "/** The single key that makes a prop value a binding rather than a literal. */\nexport const bindingMarker = %s;\n\n", strconv.Quote(sp.BindingMarker))

	b.WriteString("/** The native tier — what a client must implement (ADR 0024). */\n")
	b.WriteString("export const primitives: PrimitiveSpec[] = [\n")
	for _, p := range sp.Primitives {
		fmt.Fprintf(&b, "  {\n    type: %s,\n    tier: %s,\n    doc: %s,\n    native: %s,\n    children: %t,\n    props: [\n",
			strconv.Quote(p.Type), strconv.Quote(p.Tier), strconv.Quote(p.Doc), strconv.Quote(p.Native), p.Children)
		for _, pr := range p.Props {
			fmt.Fprintf(&b, "      { key: %s, type: %s, doc: %s },\n",
				strconv.Quote(pr.Key), strconv.Quote(pr.Type), strconv.Quote(pr.Doc))
		}
		b.WriteString("    ],\n  },\n")
	}
	b.WriteString("];\n\n")

	b.WriteString("/** The definition tier — served as data, implemented by nobody. */\n")
	b.WriteString("export const components: string[] = [\n")
	for _, c := range sp.Components {
		if c.Generic || c.Type == "" {
			continue
		}
		fmt.Fprintf(&b, "  %s,\n", strconv.Quote(c.Type))
	}
	b.WriteString("];\n\n")

	b.WriteString("/** Every behaviour a client interprets. */\n")
	b.WriteString("export const actionKinds: string[] = [\n")
	for _, a := range sp.Actions {
		fmt.Fprintf(&b, "  %s,\n", strconv.Quote(a.Kind))
	}
	b.WriteString("];\n\n")

	b.WriteString("/** The closed field-validation set. */\n")
	b.WriteString("export const validators: string[] = [\n")
	for _, v := range sp.Validators {
		fmt.Fprintf(&b, "  %s,\n", strconv.Quote(v.Name))
	}
	b.WriteString("];\n\n")

	b.WriteString("/** The closed conditional set — no expression language, on purpose. */\n")
	b.WriteString("export const predicates: string[] = [\n")
	for _, p := range sp.Predicates {
		fmt.Fprintf(&b, "  %s,\n", strconv.Quote(p.Name))
	}
	b.WriteString("];\n")
	return []byte(b.String())
}

// fixture is the client conformance artefact: the declared vocabulary as plain
// data, so a client in any language can assert that what it registers is exactly
// what the contract declares. It is published in the npm package and embedded in
// the Go module, because the drift it exists to catch is precisely a client that
// implements a set nobody compared against the contract.
type fixture struct {
	Comment       string        `json:"//"`
	Version       string        `json:"version"`
	TypeSeparator string        `json:"typeSeparator"`
	BindingMarker string        `json:"bindingMarker"`
	Primitives    []fixturePrim `json:"primitives"`
	Components    []string      `json:"components"`
	Actions       []string      `json:"actions"`
	Validators    []string      `json:"validators"`
	Predicates    []string      `json:"predicates"`
	Tones         []string      `json:"tones"`
	Surfaces      []string      `json:"surfaces"`
}

type fixturePrim struct {
	Type     string   `json:"type"`
	Tier     string   `json:"tier"`
	Children bool     `json:"children"`
	Props    []string `json:"props"`
}

func genFixture(sp spec) []byte {
	f := fixture{
		Comment: "Generated by tools/genui from ui.spec.json. DO NOT EDIT. The conformance " +
			"fixture: the vocabulary a conforming client must implement, as data. A client " +
			"test asserts that the types it registers and the action kinds it interprets are " +
			"exactly these — the check that could not exist while the primitive tier lived " +
			"only as one client's TypeScript.",
		Version:       sp.Version,
		TypeSeparator: sp.TypeSeparator,
		BindingMarker: sp.BindingMarker,
		Components:    []string{},
		Actions:       []string{},
		Validators:    []string{},
		Predicates:    []string{},
		Tones:         []string{},
		Surfaces:      []string{},
	}
	for _, p := range sp.Primitives {
		keys := make([]string, 0, len(p.Props))
		for _, pr := range p.Props {
			keys = append(keys, pr.Key)
		}
		f.Primitives = append(f.Primitives, fixturePrim{Type: p.Type, Tier: p.Tier, Children: p.Children, Props: keys})
	}
	for _, c := range sp.Components {
		if c.Generic || c.Type == "" {
			continue
		}
		f.Components = append(f.Components, c.Type)
	}
	for _, a := range sp.Actions {
		f.Actions = append(f.Actions, a.Kind)
	}
	for _, v := range sp.Validators {
		f.Validators = append(f.Validators, v.Name)
	}
	for _, p := range sp.Predicates {
		f.Predicates = append(f.Predicates, p.Name)
	}
	for _, t := range sp.Tones {
		f.Tones = append(f.Tones, t.Value)
	}
	for _, s := range sp.Surfaces {
		f.Surfaces = append(f.Surfaces, s.Value)
	}
	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		fatalf("marshal fixture: %v", err)
	}
	return append(out, '\n')
}

// ── lint ─────────────────────────────────────────────────────────────────────

// lintSpecSanity is the part of lint that needs nothing but the spec, and that
// generation itself depends on being true. It runs twice — once before
// generating, once as part of the full lint — because running it only at the end
// means a spec error surfaces as a generator crash.
func lintSpecSanity(sp spec) []string {
	var errs []string
	if sp.Version == "" {
		errs = append(errs, "the spec declares no version — a client cannot negotiate against an unnamed vocabulary")
	}
	if sp.BindingMarker == "" {
		errs = append(errs, "the spec declares no bindingMarker — nothing then distinguishes a bound prop from a literal object")
	}
	for _, sg := range sp.Sugar {
		if sg.Key == sp.BindingMarker {
			errs = append(errs, fmt.Sprintf("sugar %q sets a prop named %q, which is the binding marker — the two cannot share a key", sg.Func, sg.Key))
		}
	}
	if sp.TypeSeparator == "" {
		errs = append(errs, "the spec declares no typeSeparator — nothing then distinguishes a module's type from a core one")
		return errs
	}
	// The separator belongs to modules. A core type carrying it would be
	// indistinguishable from a module's own, so one `Foo:Bar` in the core
	// vocabulary makes every namespaced type ambiguous rather than just itself.
	for _, p := range sp.Primitives {
		if strings.Contains(p.Type, sp.TypeSeparator) {
			errs = append(errs, fmt.Sprintf("primitive %q contains the type separator %q, which is reserved for module namespacing", p.Type, sp.TypeSeparator))
		}
	}
	for _, c := range sp.Components {
		if c.Type != "" && strings.Contains(c.Type, sp.TypeSeparator) {
			errs = append(errs, fmt.Sprintf("component %q contains the type separator %q, which is reserved for module namespacing", c.Type, sp.TypeSeparator))
		}
	}
	return errs
}

func runLint(sp spec, root string) []string {
	defsDir := filepath.Join(root, "definitions")
	var errs []string

	errs = append(errs, lintSpecSanity(sp)...)

	// Integrity: no duplicate exported names.
	seen := map[string]string{}
	claim := func(name, kind string) {
		if prev, ok := seen[name]; ok {
			errs = append(errs, fmt.Sprintf("duplicate name %q (%s and %s)", name, prev, kind))
		}
		seen[name] = kind
	}
	for _, p := range sp.Primitives {
		claim(p.fn(), "primitive")
	}
	for _, c := range sp.Components {
		claim(c.Func, "component")
	}
	for _, s := range sp.Slots {
		claim(s.Func, "slot")
	}
	for _, s := range sp.Sugar {
		claim(s.Func, "sugar")
		claim("Bind"+s.Func, "bound sugar")
	}
	for _, a := range sp.Actions {
		claim(a.Func, "action")
	}
	for _, t := range sp.Tones {
		claim(t.Const, "tone")
	}

	// Authorable prop keys: any positional key across components, plus every
	// sugar key. A definition may bind any of these.
	authorable := map[string]bool{}
	for _, c := range sp.Components {
		for _, p := range c.Positional {
			authorable[p.Key] = true
		}
	}
	for _, s := range sp.Sugar {
		authorable[s.Key] = true
	}
	slotNames := map[string]bool{}
	for _, s := range sp.Slots {
		slotNames[s.Name] = true
	}
	byType := map[string]bool{}
	for _, c := range sp.Components {
		if c.Type != "" && !c.Generic {
			byType[c.Type] = true
		}
	}
	primType := map[string]bool{}
	for _, p := range sp.Primitives {
		primType[p.Type] = true
	}

	// Per-primitive justification. The tier costs a client release per addition
	// (ADR 0024), so an entry that does not say why a definition cannot express
	// it is an addition made without paying attention to what it costs.
	knownTiers := map[string]bool{
		"presentational": true, "interactive": true, "field": true,
		"computed": true, "player": true,
	}
	for _, p := range sp.Primitives {
		if strings.TrimSpace(p.Native) == "" {
			errs = append(errs, fmt.Sprintf("primitive %q has no `native` justification — say why a definition cannot express it", p.Type))
		}
		if !knownTiers[p.Tier] {
			errs = append(errs, fmt.Sprintf("primitive %q has unknown tier %q", p.Type, p.Tier))
		}
	}

	for _, t := range sp.Tones {
		if t.TS == "" || t.Sdui == "" {
			errs = append(errs, fmt.Sprintf("tone %q needs both `sdui` and `ts` names — the generated constants alias the two bindings' enums", t.Value))
		}
	}

	// Tier overlap. A type in both tiers is ambiguous to every client: it would
	// have to choose between its own implementation and the served definition,
	// and nothing says which wins.
	for _, p := range sp.Primitives {
		if byType[p.Type] {
			errs = append(errs, fmt.Sprintf("type %q is both a primitive and a component — a type belongs to exactly one tier", p.Type))
		}
	}

	// Action-kind, tone and surface coverage across the three places the
	// vocabulary is written down. This is the whole three-vocabularies fault
	// mechanised: the spec, the JSON Schema and the proto must agree.
	errs = append(errs, lintWireCoverage(sp, root)...)

	files, _ := filepath.Glob(filepath.Join(defsDir, "*.json"))
	sort.Strings(files)

	// Every component in the spec must have a definition file. The reverse
	// direction was already checked; without this one a component can be
	// authorable from Go and render as an Unknown placeholder on every client.
	defined := map[string]bool{}

	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", f, err))
			continue
		}
		var def struct {
			Name     string `json:"name"`
			Template any    `json:"template"`
			Fallback any    `json:"fallback"`
		}
		if err := json.Unmarshal(raw, &def); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", f, err))
			continue
		}
		base := filepath.Base(f)
		if !byType[def.Name] {
			errs = append(errs, fmt.Sprintf("%s: definition %q has no component in ui.spec.json — add one (a new component needs an authoring entry)", base, def.Name))
			continue
		}
		defined[def.Name] = true
		binds, aliases, outlets, refs := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
		collect(def.Template, binds, aliases, outlets, refs)

		// A fallback is held to the template's rules — it is a template, and the
		// client cannot tell which one it was sent — plus one of its own: it must
		// need strictly fewer primitives. A fallback that needs the same set
		// degrades nothing and would be served to a client that still cannot
		// draw it, which is worse than having none, because it reads as handled.
		if def.Fallback != nil {
			fbRefs := map[string]bool{}
			collect(def.Fallback, binds, aliases, outlets, fbRefs)
			tmplPrims := primitivesIn(refs, primType)
			fbPrims := primitivesIn(fbRefs, primType)
			for r := range fbRefs {
				refs[r] = true // held to the same referenced-type check below
			}
			fewer := false
			for p := range fbPrims {
				if !tmplPrims[p] {
					errs = append(errs, fmt.Sprintf("%s: fallback needs primitive %q that the template does not — a fallback may only use fewer", base, p))
				}
			}
			for p := range tmplPrims {
				if !fbPrims[p] {
					fewer = true
				}
			}
			if !fewer {
				errs = append(errs, fmt.Sprintf("%s: fallback needs the same primitives as the template, so it degrades nothing — remove it or simplify it", base))
			}
		}

		// Referenced-type existence. A template naming a type that is in neither
		// tier expands into a node every client renders as an Unknown
		// placeholder — a hole in a screen, reported by nothing.
		for r := range refs {
			if !primType[r] && !byType[r] {
				errs = append(errs, fmt.Sprintf("%s: template references type %q, which is neither a primitive nor a component", base, r))
			}
		}
		for b := range binds {
			// A $each alias and anything under it ("s", "s.label") is a loop
			// variable, not a component prop. Matching only the bare alias missed
			// every dotted path, which is the form a repeated node actually uses —
			// so any definition with an $each in it could not pass this lint, and
			// none of them lived here.
			if head, _, _ := strings.Cut(b, "."); aliases[head] {
				continue
			}
			// Runtime-injected bindings ($childCount, $slots) are supplied by the
			// expander from what the caller passed, so no helper sets them and
			// none should.
			if strings.HasPrefix(b, "$") {
				continue
			}
			if !authorable[b] {
				errs = append(errs, fmt.Sprintf("%s: template binds %q but no ui helper sets it (add it as a positional arg or sugar in ui.spec.json)", base, b))
			}
		}
		for o := range outlets {
			if !slotNames[o] {
				errs = append(errs, fmt.Sprintf("%s: template declares Outlet %q but no slot helper fills it", base, o))
			}
		}
	}

	for _, c := range sp.Components {
		if c.Generic || c.Type == "" {
			continue
		}
		if !defined[c.Type] {
			errs = append(errs, fmt.Sprintf("component %q has no definitions/*.json file — it is authorable from Go and renders as Unknown everywhere", c.Type))
		}
	}

	sort.Strings(errs)
	return errs
}

// primitivesIn narrows a set of referenced types to the primitive tier — the
// only tier a client can fail to draw, and therefore the only one a fallback is
// about.
func primitivesIn(refs, primType map[string]bool) map[string]bool {
	out := map[string]bool{}
	for r := range refs {
		if primType[r] {
			out[r] = true
		}
	}
	return out
}

// collect walks a definition template gathering $bind prop paths, $as loop
// aliases, Outlet slot names and every node type the template references.
func collect(v any, binds, aliases, outlets, refs map[string]bool) {
	switch x := v.(type) {
	case map[string]any:
		if b, ok := x["$bind"].(string); ok {
			binds[b] = true
		}
		if a, ok := x["$as"].(string); ok {
			aliases[a] = true
		}
		if t, ok := x["type"].(string); ok {
			refs[t] = true
			if t == "Outlet" {
				if props, ok := x["props"].(map[string]any); ok {
					if name, ok := props["name"].(string); ok {
						outlets[name] = true
					}
				}
			}
		}
		for _, val := range x {
			collect(val, binds, aliases, outlets, refs)
		}
	case []any:
		for _, val := range x {
			collect(val, binds, aliases, outlets, refs)
		}
	}
}

// ── wire coverage ────────────────────────────────────────────────────────────

var protoEnumRe = regexp.MustCompile(`(?m)^\s*ACTION_KIND_([A-Z0-9_]+)\s*=\s*\d+;`)

// lintWireCoverage reconciles the spec's enums with the two places the wire
// writes them down: schema/sdui.schema.json (what a JSON consumer validates
// against) and proto/mosaic/sdui/v1/sdui.proto (what the transport carries).
//
// This gate is the reason the spec exists as one file. Before it, the proto
// enumerated ten action kinds, the schema nine and the client nine — three
// numbers, none of them checked against another, and six of the wire kinds were
// unauthorable from Go. Nothing failed; the sets simply disagreed.
func lintWireCoverage(sp spec, root string) []string {
	var errs []string

	specKinds := make([]string, 0, len(sp.Actions))
	for _, a := range sp.Actions {
		specKinds = append(specKinds, a.Kind)
	}
	specTones := make([]string, 0, len(sp.Tones))
	for _, t := range sp.Tones {
		specTones = append(specTones, t.Value)
	}
	specSurfaces := make([]string, 0, len(sp.Surfaces))
	for _, s := range sp.Surfaces {
		specSurfaces = append(specSurfaces, s.Value)
	}

	schemaPath := filepath.Join(root, "schema", "sdui.schema.json")
	raw, err := os.ReadFile(schemaPath)
	if err != nil {
		return append(errs, fmt.Sprintf("read %s: %v", schemaPath, err))
	}
	var sch struct {
		Defs map[string]struct {
			Enum []string `json:"enum"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(raw, &sch); err != nil {
		return append(errs, fmt.Sprintf("parse %s: %v", schemaPath, err))
	}
	errs = append(errs, diffSet("schema ActionKind", specKinds, sch.Defs["ActionKind"].Enum)...)
	errs = append(errs, diffSet("schema Tone", specTones, sch.Defs["Tone"].Enum)...)
	errs = append(errs, diffSet("schema Surface", specSurfaces, sch.Defs["Surface"].Enum)...)

	protoPath := filepath.Join(root, "proto", "mosaic", "sdui", "v1", "sdui.proto")
	praw, err := os.ReadFile(protoPath)
	if err != nil {
		return append(errs, fmt.Sprintf("read %s: %v", protoPath, err))
	}
	var protoKinds []string
	for _, m := range protoEnumRe.FindAllStringSubmatch(string(praw), -1) {
		if m[1] == "UNSPECIFIED" {
			continue
		}
		protoKinds = append(protoKinds, screamingToCamel(m[1]))
	}
	errs = append(errs, diffSet("proto ActionKind", specKinds, protoKinds)...)
	return errs
}

// diffSet reports each way the two sets disagree, naming the direction — a kind
// the spec declares that the wire cannot carry is a different bug from one the
// wire carries that nothing can author.
func diffSet(what string, spec, wire []string) []string {
	var errs []string
	inWire := map[string]bool{}
	for _, w := range wire {
		inWire[w] = true
	}
	inSpec := map[string]bool{}
	for _, s := range spec {
		inSpec[s] = true
	}
	for _, s := range spec {
		if !inWire[s] {
			errs = append(errs, fmt.Sprintf("%s is missing %q, which ui.spec.json declares", what, s))
		}
	}
	for _, w := range wire {
		if !inSpec[w] {
			errs = append(errs, fmt.Sprintf("%s carries %q, which ui.spec.json does not declare — nothing can author it", what, w))
		}
	}
	return errs
}

// screamingToCamel turns a proto enum suffix (OPEN_URL) into the JSON
// discriminator the contract uses (openUrl).
func screamingToCamel(s string) string {
	parts := strings.Split(strings.ToLower(s), "_")
	for i := 1; i < len(parts); i++ {
		r := []rune(parts[i])
		r[0] = unicode.ToUpper(r[0])
		parts[i] = string(r)
	}
	return strings.Join(parts, "")
}

// ── helpers ──────────────────────────────────────────────────────────────────

func loadSpec(path string) spec {
	raw, err := os.ReadFile(path)
	if err != nil {
		fatalf("read spec: %v", err)
	}
	var sp spec
	if err := json.Unmarshal(raw, &sp); err != nil {
		fatalf("parse spec: %v", err)
	}
	return sp
}

func writeFile(path string, data []byte) {
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fatalf("write %s: %v", path, err)
	}
}

// dropLeadName lets a spec doc read as a full sentence in prose while GoDoc gets
// the required "Name …" lead: if the doc already starts with the func name, the
// name is not duplicated.
func dropLeadName(name, doc string) string {
	if rest, ok := strings.CutPrefix(doc, name+" "); ok {
		return rest
	}
	return doc
}

// normalize makes byte comparison newline-insensitive so a CRLF checkout does
// not read as stale against LF-generated output.
func normalize(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "genui: "+format+"\n", a...)
	os.Exit(1)
}
