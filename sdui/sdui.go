// Package sdui is the Go binding of the Mosaic Server-Driven-UI contract — the
// producer side. The Platform and Modules build a tree of Nodes (generated
// protobuf UINodes) carrying Action envelopes; a client renders it.
//
// This is the faithful protobuf port (ADR 0044): the UINode tree is the typed
// mosaic.sdui.v1.UINode, so it rides the transport (RegionUpdate.ui_node) as a
// typed message. Actions and enums keep their JSON form inside the open props
// bag — props is a protobuf Struct, so anything in it is JSON-encoded regardless.
// A later ergonomics redesign may hoist actions and enums into typed fields;
// this port changes only the tree type, not the wire.
package sdui

import (
	sduiv1 "github.com/mosaic-media/contracts/gen/mosaic/sdui/v1"
)

// Node is a UI node — a pointer to the generated protobuf UINode (protobuf
// messages carry a do-not-copy marker, so producers pass them by pointer).
type Node = *sduiv1.UINode

// ComponentDefinition is a component expressed as data (ADR 0024).
type ComponentDefinition = *sduiv1.ComponentDefinition

// Props is a component's open property bag. It is JSON-encoded into the node's
// protobuf Struct when the node is built.
type Props = map[string]any

// ActionKind, Tone and Surface are the string enums that ride inside the open
// props bag. The generated protobuf enums exist for a future typed-action
// redesign; the faithful port keeps these as the JSON strings clients read.
type (
	ActionKind = string
	Tone       = string
	Surface    = string
)

// Action is a declarative behaviour envelope — data, never code. Faithful port:
// actions ride inside the open props bag as JSON, so this is a JSON-shaped struct
// rather than the (for now unused) protobuf Action message. Each kind uses a
// subset of the fields; the constructors in actions.go hide the pointer optionals.
type Action struct {
	Kind     ActionKind     `json:"kind"`
	Screen   *string        `json:"screen,omitempty"`
	Params   map[string]any `json:"params,omitempty"`
	URL      *string        `json:"url,omitempty"`
	Mutation *string        `json:"mutation,omitempty"`
	Input    map[string]any `json:"input,omitempty"`
	Surface  *Surface       `json:"surface,omitempty"`
	Node     map[string]any `json:"node,omitempty"`
	PartID   *string        `json:"partId,omitempty"`
	NodeID   *string        `json:"nodeId,omitempty"`
	Message  *string        `json:"message,omitempty"`
	Tone     *Tone          `json:"tone,omitempty"`
	Field    *string        `json:"field,omitempty"`
	Value    *string        `json:"value,omitempty"`
	Actions  []Action       `json:"actions,omitempty"`
}

// The action kind, tone, surface and node-type constants are generated into
// vocabulary.gen.go from ui.spec.json. They were hand-written here and in a
// components.go beside it, and both had drifted: the constant list named three
// primitives as components and omitted every real one, while the kinds it
// declared were one short of the proto's and three short of the vocabulary's.
