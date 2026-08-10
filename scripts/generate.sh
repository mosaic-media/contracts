#!/usr/bin/env bash
# Generate the per-language contract bindings from the single source of truth,
# schema/sdui.schema.json. Do NOT hand-edit the generated files.
#
#   Go   -> sdui/contract/contract.gen.go   (package contract)
#   TS   -> ts/contract.gen.ts
#   Dart -> add when a Dart client lands (quicktype --lang dart)
#
# Run: scripts/generate.sh   (requires npx + gofmt on PATH)
set -euo pipefail
cd "$(dirname "$0")/.."

SCHEMA="schema/sdui.schema.json"
# quicktype is PINNED. Unpinned (`npx --yes quicktype`) fetches the latest,
# which makes the generator — and therefore the drift guard that regenerates
# and diffs — non-deterministic: a new quicktype release silently restyles the
# output, and the guard then fails on a "stale" binding that is only a different
# generator. 25.1.0 is the last release before 26.0.0, whose TypeScript renderer
# stopped annotating the open `props`/`params` maps; pinning here keeps those
# annotations (the type contract) and freezes the formatting. Bump deliberately,
# regenerating and reviewing the diff in the same change.
QT=(npx --yes quicktype@25.1.0 -s schema --top-level MosaicSDUI "$SCHEMA")

echo "generating Go -> sdui/contract/contract.gen.go"
"${QT[@]}" --lang go --package contract --just-types-and-package -o sdui/contract/contract.gen.go
# The consolidated schema has a wrapper root (so the generator reaches every
# top-level type). Strip that unused wrapper struct, then mark as generated.
sed -i '/^type MosaicSDUI struct {/,/^}/d' sdui/contract/contract.gen.go
sed -i '1s|^|// Code generated from schema/sdui.schema.json by quicktype. DO NOT EDIT.\n|' sdui/contract/contract.gen.go
gofmt -w sdui/contract/contract.gen.go

echo "generating TypeScript -> ts/contract.gen.ts"
"${QT[@]}" --lang typescript --just-types -o ts/contract.gen.ts
sed -i '/^export interface MosaicSDUI {/,/^}/d' ts/contract.gen.ts
sed -i '1s|^|// Code generated from schema/sdui.schema.json by quicktype. DO NOT EDIT.\n|' ts/contract.gen.ts

# The vocabulary — all three tiers — is generated from ui.spec.json by
# tools/genui: the authoring layer (ui/components.gen.go, ts/ui.ts), the
# machine-readable registry (sdui/vocabulary.gen.go, ts/vocabulary.gen.ts) and
# the client conformance fixture (conformance/vocabulary.json). It also lints
# the spec against definitions/, the schema and the proto.
echo "generating vocabulary -> ui/, ts/, sdui/vocabulary.gen.go, conformance/ (+ lint)"
go run ./tools/genui

# The component library as one importable module, for a JS consumer with no
# Platform to push it. The definitions are authored as definitions/*.json; this
# is an aggregate of them, as definitions.Library() is on the Go side.
echo "generating definition library -> ts/definitions.gen.ts"
node scripts/gen-definitions.mjs

# The design tokens as CSS custom properties, for a web client to apply before
# the session connects. tokens/tokens.json is the authored source and the
# Platform serves it; this is the bootstrap copy, never a second source.
echo "generating design tokens -> tokens/tokens.css"
node scripts/gen-tokens.mjs

# The protobuf bindings (contracts#6): Go + Connect stubs and TypeScript for the
# client-facing contracts, Go + gRPC stubs for the module wire (platform#39). Two
# templates because the two surfaces need different plugins — Connect for a
# browser-facing contract, gRPC for what go-plugin registers on a *grpc.Server —
# and the split is deliberate, not duplication (buf.gen.module.yaml says why).
#
# npm ci first, so protoc-gen-es resolves to the version package-lock.json pins
# rather than whatever `npx` would fetch — the same determinism quicktype's pin
# buys, for the TS half. buf and the Go plugins are pinned in Dockerfile.test.
if command -v buf >/dev/null 2>&1; then
  echo "generating protobuf -> gen/, gen-ts/"
  npm ci >/dev/null
  buf generate --template buf.gen.yaml \
    --path proto/mosaic/auth --path proto/mosaic/sdui --path proto/mosaic/session
  buf generate --template buf.gen.module.yaml --path proto/mosaic/module
else
  # buf is absent on a bare host; the container has it. Skipping here keeps the
  # script runnable for the quicktype/genui half without the protobuf toolchain,
  # while the drift guard — which always runs in the container — covers proto.
  echo "buf not found; skipping protobuf generation (the container gate covers it)"
fi

echo "done."
