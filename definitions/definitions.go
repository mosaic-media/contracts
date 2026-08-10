// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

// Package definitions is the standard SDUI component library as data — one JSON
// file per component, each a name, its default params and a template of
// primitives (contracts#2).
//
// **This is the only place a component is authored.** A component is data, not
// code: the Platform serves this library to every client over the session
// (contracts#4), and a client renders it by expanding the template against the
// primitives it implements natively. A component written as code in a client
// repository is renderable by that client alone, which is the whole thing the
// definition model exists to prevent — so it is a bug, not a shortcut.
//
// The package exists because Go cannot embed across a directory boundary: the
// accessor has to live beside the data it embeds. `Library` is what the Platform
// pushes on connect; the npm package exports the same files under
// `./definitions/*` for JavaScript clients and the storybook.
package definitions

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed *.json
var files embed.FS

// FS is the raw definition files, for a consumer that wants them one at a time.
var FS fs.FS = files

// Library returns the whole component library as the JSON array a client
// registers: `[{name, params, template}, …]`, ordered by file name so the bytes
// are stable between builds and a diff of what the Platform serves is readable.
func Library() ([]byte, error) {
	names, err := fs.Glob(files, "*.json")
	if err != nil {
		return nil, err
	}
	sort.Strings(names)

	// Decoded and re-encoded rather than concatenated: a definition that is not
	// valid JSON must fail here, at the seam that owns the library, rather than
	// arriving at a client as a malformed payload it can only discard.
	out := make([]json.RawMessage, 0, len(names))
	for _, name := range names {
		raw, err := files.ReadFile(name)
		if err != nil {
			return nil, err
		}
		var one json.RawMessage
		if err := json.Unmarshal(raw, &one); err != nil {
			return nil, fmt.Errorf("definition %s: %w", name, err)
		}
		out = append(out, one)
	}
	return json.Marshal(out)
}

// Count reports how many components the library carries.
func Count() int {
	names, _ := fs.Glob(files, "*.json")
	return len(names)
}
