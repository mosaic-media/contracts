// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 the Mosaic authors

/*
 * Bindable props, the TypeScript half.
 *
 * A prop value may be a literal, or a binding naming a path the client resolves
 * against the scopes in force where the node renders. A binding names a path and
 * nothing else — no formatting, no default, no expression, no operator.
 * Formatting stays server-side, which is what "presentation-ready data" means
 * and the honest reason none of this needs an evaluator.
 *
 * The marker is imported from the generated vocabulary rather than written here,
 * so a producer reading the Go package and a client reading this cannot disagree
 * about what makes a value a binding.
 */

import { bindingMarker } from "./vocabulary.gen.js";

export { bindingMarker };

/** A prop value that resolves at render rather than being sent whole. */
export type Binding = { [k: string]: string };

/** Name a path to resolve where the node renders. */
export function bind(path: string): Binding {
  return { [bindingMarker]: path };
}

/**
 * Is this prop value a binding?
 *
 * Exactly one key, and it is the marker, and its value is a non-empty string.
 * An object carrying the marker beside anything else is a literal object, not a
 * malformed binding — the same closed reading the action-kind check uses, so a
 * prop that happens to contain a field of this name is never silently
 * reinterpreted.
 */
export function isBinding(v: unknown): v is Binding {
  if (typeof v !== "object" || v === null || Array.isArray(v)) return false;
  const keys = Object.keys(v as Record<string, unknown>);
  if (keys.length !== 1 || keys[0] !== bindingMarker) return false;
  const path = (v as Record<string, unknown>)[bindingMarker];
  return typeof path === "string" && path.length > 0;
}

/** The path a binding names, or undefined if the value is not one. */
export function bindingPath(v: unknown): string | undefined {
  return isBinding(v) ? (v as Record<string, string>)[bindingMarker] : undefined;
}
