// Code generated from schema/sdui.schema.json by quicktype. DO NOT EDIT.
package contract

// The single source of truth for the Mosaic Server-Driven-UI contract. Language bindings
// (Go, TypeScript, Dart) are GENERATED from this file — do not hand-edit them. The root
// object exists only so a generator reaches every top-level type; the useful types are in
// $defs.

// A declarative behaviour envelope. Data, never code — the client interprets the kind. Each
// kind uses a subset of the fields.
type Action struct {
	Actions []Action `json:"actions,omitempty"`
	// setValue: the name of the field written in the enclosing state scope. submit: where in
	// the action input the collected values merge — a path into it, or absent for its top level.
	Field    *string                `json:"field,omitempty"`
	Input    map[string]interface{} `json:"input,omitempty"`
	Kind     ActionKind             `json:"kind"`
	Message  *string                `json:"message,omitempty"`
	Mutation *string                `json:"mutation,omitempty"`
	Node     *UINode                `json:"node,omitempty"`
	NodeID   *string                `json:"nodeId,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
	PartID   *string                `json:"partId,omitempty"`
	Screen   *string                `json:"screen,omitempty"`
	Surface  *Surface               `json:"surface,omitempty"`
	Tone     *Tone                  `json:"tone,omitempty"`
	URL      *string                `json:"url,omitempty"`
	// setValue: the value written.
	Value *string `json:"value,omitempty"`
}

// One element of a server-driven UI tree. The `type` is an open vocabulary: a client that
// does not recognise a type renders a placeholder rather than failing.
//
// An alternative template for a client whose declared vocabulary lacks a primitive the main
// template needs. The server picks per session and sends one; the client never sees both.
// Optional — a definition without one is served unchanged to everyone.
type UINode struct {
	Children []UINode `json:"children,omitempty"`
	// A stable identity for this node. It is the React key and the target of a narrow region
	// update, and it is also the *analytics* identity: an onAppear action names what was seen
	// by carrying whatever the server put here, so an id that is a row index attributes nothing
	// and one that names the thing attributes everything. The server chooses it, because only
	// the server knows what the node is about.
	ID *string `json:"id,omitempty"`
	// Component-specific data. Open by design. Any value may be a literal or a Binding (see
	// #/$defs/Binding), which the client resolves where the node renders.
	Props map[string]interface{} `json:"props,omitempty"`
	Slots map[string][]UINode    `json:"slots,omitempty"`
	// Component discriminator, e.g. "PosterCard". Namespaced: a core type — every primitive and
	// every component in the published vocabulary — is unprefixed and never contains the type
	// separator; a module's own type is "moduleId:Type". Open and flat is not the same as open:
	// without the namespace two modules could both contribute a StatChip, and one could
	// contribute a PosterCard and take the core component's place in every client's registry.
	Type string `json:"type"`
}

// A style override applied below a viewport width — the vocabulary's one responsive
// capability, and what lets a layout adapt as DATA rather than through a client stylesheet.
// The override is a plain BoxStyle merged over the base: a field it does not mention keeps
// its base value, and null clears one (these travel as JSON, where an undefined member
// would vanish silently). One breakpoint, not a cascade.
type Responsive struct {
	// Viewport width in px, below which the override applies.
	Below float64  `json:"below"`
	Style BoxStyle `json:"style"`
}

// Layout and surface styling for a Box, in TOKENS ONLY — no literal colours, no raw pixels
// except explicit dimensions. This is the technology-agnostic seam: it is the intersection
// of what a flexbox client and a Flutter client can render identically, so a definition
// written once renders the same everywhere. A client implements this vocabulary natively;
// growing it is the only change that requires a client release (ADR 0024).
type BoxStyle struct {
	// Cross-axis alignment.
	Align *BoxStyleAlign `json:"align,omitempty"`
	// An aspect ratio as "w / h".
	AspectRatio *string     `json:"aspectRatio,omitempty"`
	Bg          *ColorToken `json:"bg,omitempty"`
	// A linear gradient between two token colours (or "transparent"), at an angle in degrees.
	BgGradient  *BgGradient `json:"bgGradient,omitempty"`
	Border      *bool       `json:"border,omitempty"`
	BorderColor *ColorToken `json:"borderColor,omitempty"`
	// Draw `border` on one edge only — a rule under a row, a marker down the side of the
	// selected item. Without it `border` draws all four.
	BorderSide *BorderSide `json:"borderSide,omitempty"`
	// The border's weight. Two steps, not a number: a hairline rule and a marker heavy enough
	// to read as a selection. Anything thicker is a filled box, which this vocabulary already
	// has.
	BorderWidth *BorderWidth `json:"borderWidth"`
	Bottom      *SpaceToken  `json:"bottom"`
	Color       *ColorToken  `json:"color,omitempty"`
	Direction   *Direction   `json:"direction,omitempty"`
	Flex        *float64     `json:"flex,omitempty"`
	Gap         *SpaceToken  `json:"gap"`
	// Render this surface in the acrylic material: translucent, blurred, and lit by the current
	// light source (the focused artwork, or the brand light when there is none). A client
	// without the material renders a plain translucent surface.
	Glass *bool `json:"glass,omitempty"`
	// A soft bloom cast over this box, screen-blended so it lights the artwork rather than
	// tinting it flat. "art" draws the palette sampled from the focused image (the same source
	// the acrylic material is lit by), so a hero glows in the colours of its own backdrop;
	// "brand" uses the accent pair, for a surface with no artwork to sample. A client with no
	// sampler renders "art" as "brand" rather than nothing.
	Glow *Glow `json:"glow,omitempty"`
	// Overlay the material texture — the grain, blotch and scuff the token set carries — so a
	// large flat surface reads as a material rather than as a fill. Soft-light blended over
	// whatever is beneath, and purely decorative: it never affects layout or hit-testing.
	Grain *bool `json:"grain,omitempty"`
	// grid: fixed track size for a rail, in px.
	GridAutoColumns *float64 `json:"gridAutoColumns,omitempty"`
	// grid: an explicit column track list, for the arrangements auto-fill cannot state — a
	// settings frame's nav/panel/aside, an episode row's fixed thumbnail beside a fluid title.
	// A track is a number of pixels (fixed), "auto" (sized to its content), or {"fill": n} (n
	// shares of what is left). Deliberately not a CSS grid-template string: these three are
	// what a flexbox client and a Flutter client can both lay out identically.
	GridColumns []GridColumnElement `json:"gridColumns,omitempty"`
	// grid: flow direction; "column" makes a horizontal rail.
	GridFlow *Direction `json:"gridFlow,omitempty"`
	// grid: responsive auto-fill columns of at least this width in px.
	GridMin *float64   `json:"gridMin,omitempty"`
	Grow    *bool      `json:"grow,omitempty"`
	Height  *Dimension `json:"height"`
	// Take the box out of the layout entirely. Its purpose is `responsive`: one payload carries
	// both a desktop and a phone arrangement, and each viewport drops the half it does not use.
	Hidden *bool `json:"hidden,omitempty"`
	// Mark this box as the region whose hover or focus reveals the `hoverReveal` boxes inside
	// it. Stated explicitly, and on the ancestor rather than inferred from interactivity,
	// because the thing revealed is frequently not inside the thing hovered: a rail's "see all"
	// appears when the *section* is approached, and a link that appeared only while pointing at
	// it could never be clicked.
	HoverGroup *bool `json:"hoverGroup,omitempty"`
	// Hide this box until its nearest `hoverGroup` ancestor is hovered or focused, then fade it
	// in. It is the card veil: the extra detail a tile shows on approach — time remaining, file
	// size, a play affordance — which must not be in the resting composition and must not cost
	// a second payload. An input model with no pointer reveals it on focus instead, and a
	// client with neither renders it always-visible rather than never: unreachable detail is
	// the worse failure.
	HoverReveal *bool `json:"hoverReveal,omitempty"`
	// Main-axis distribution.
	Justify *Justify `json:"justify,omitempty"`
	// DEPRECATED in favour of `responsive`. A named hook a client may map to a rule of its own;
	// it makes a layout depend on client CSS, which is what `responsive` exists to avoid.
	Kind *string `json:"kind,omitempty"`
	// Layout mode. "grid" enables the grid-* fields.
	Layout    *Layout     `json:"layout,omitempty"`
	Left      *SpaceToken `json:"left"`
	MaxWidth  *Dimension  `json:"maxWidth"`
	MinHeight *Dimension  `json:"minHeight"`
	MinWidth  *Dimension  `json:"minWidth"`
	Opacity   *float64    `json:"opacity,omitempty"`
	Overflow  *Overflow   `json:"overflow,omitempty"`
	OverflowX *Overflow   `json:"overflowX,omitempty"`
	OverflowY *Overflow   `json:"overflowY,omitempty"`
	// Pull this box up over what precedes it by one step of the spacing scale, so a content
	// sheet rides over the bottom of a full-bleed hero. A single direction and a token step,
	// rather than negative margins in four directions: this is the one case in the layout where
	// two siblings are meant to intersect, and naming it keeps it from becoming an
	// arbitrary-offset escape hatch. Pair with `z` to say which one is in front.
	Overlap  *SpaceToken  `json:"overlap"`
	P        *SpaceToken  `json:"p"`
	Pb       *SpaceToken  `json:"pb"`
	Pl       *SpaceToken  `json:"pl"`
	Position *Position    `json:"position,omitempty"`
	PR       *SpaceToken  `json:"pr"`
	Pt       *SpaceToken  `json:"pt"`
	Px       *SpaceToken  `json:"px"`
	Py       *SpaceToken  `json:"py"`
	Radius   *RadiusToken `json:"radius,omitempty"`
	// A style override applied below a viewport width — the vocabulary's one responsive
	// capability, and what lets a layout adapt as DATA rather than through a client stylesheet.
	// The override is a plain BoxStyle merged over the base: a field it does not mention keeps
	// its base value, and null clears one (these travel as JSON, where an undefined member
	// would vanish silently). One breakpoint, not a cascade.
	Responsive *Responsive `json:"responsive,omitempty"`
	Right      *SpaceToken `json:"right"`
	// A named legibility wash over artwork, so text laid on a backdrop stays readable whatever
	// the image behind it. NAMED rather than a gradient the server writes, for the same reason
	// colours are tokens: the recipe is a formula each client evaluates over its own tokens,
	// and one written as literal stops would be a second skin the Platform owns. "bottom"/"top"
	// fade an edge into the page; "leading" washes the text side (the direction follows the
	// reading direction, so it mirrors in RTL); "cinematic" is both — the full-bleed hero
	// treatment.
	Scrim  *Scrim       `json:"scrim,omitempty"`
	Shadow *ShadowToken `json:"shadow,omitempty"`
	// Scroll snapping axis, for carousels.
	Snap      *Snap       `json:"snap,omitempty"`
	SnapAlign *SnapAlign  `json:"snapAlign,omitempty"`
	Top       *SpaceToken `json:"top"`
	Width     *Dimension  `json:"width"`
	Wrap      *bool       `json:"wrap,omitempty"`
	// Portable stack order for chrome that overlays content — not an arbitrary index.
	Z *Z `json:"z,omitempty"`
}

// A linear gradient between two token colours (or "transparent"), at an angle in degrees.
type BgGradient struct {
	Angle *float64     `json:"angle,omitempty"`
	From  GradientStop `json:"from"`
	To    GradientStop `json:"to"`
}

type GridColumnClass struct {
	Fill float64 `json:"fill"`
}

// A component expressed as data: a name, default params, and a template of primitives.
// Clients register definitions and expand them; this is how a module contributes a
// component without shipping client code. A template node's props may hold binding objects
// ({"$bind":"path"} / {"$match":{…}}) and control keys ($if / $ifNot / $each / $as); a node
// of type "Outlet" renders the caller's children or a named slot.
type ComponentDefinition struct {
	// An alternative template for a client whose declared vocabulary lacks a primitive the main
	// template needs. The server picks per session and sends one; the client never sees both.
	// Optional — a definition without one is served unchanged to everyone.
	Fallback *UINode `json:"fallback,omitempty"`
	// The node type this definition provides.
	Name string `json:"name"`
	// Default param values, overridden by the caller's props.
	Params   map[string]interface{} `json:"params,omitempty"`
	Template UINode                 `json:"template"`
}

// Typography for a Text node, in tokens only.
type TextStyle struct {
	Align  *TextStyleAlign `json:"align,omitempty"`
	Color  *ColorToken     `json:"color,omitempty"`
	Italic *bool           `json:"italic,omitempty"`
	// Truncate after this many lines.
	LineClamp *int64 `json:"lineClamp,omitempty"`
	Mono      *bool  `json:"mono,omitempty"`
	// A legibility shadow behind the glyphs, for text laid directly over artwork with no scrim
	// beneath it — a title on a tile, a caption on a still. Boolean rather than a shadow spec:
	// the only question a payload gets to answer is whether the text is over an image, and how
	// that is drawn is the client's.
	Shadow *bool `json:"shadow,omitempty"`
	// Tabular figures, so numbers in a column line up.
	Tabular *bool `json:"tabular,omitempty"`
	// Letter-spacing: tight for display headings, wide for an eyebrow.
	Tracking  *Tracking    `json:"tracking,omitempty"`
	Transform *Transform   `json:"transform,omitempty"`
	Variant   *TextVariant `json:"variant,omitempty"`
	Weight    *FontWeight  `json:"weight,omitempty"`
}

// The behaviours a client interprets. Kept in step with ui.spec.json and
// proto/mosaic/sdui/v1/sdui.proto by `go run ./tools/genui -lint`, which fails when the
// three disagree — they had drifted to ten, nine and four before that gate existed.
type ActionKind string

const (
	ActionKindToast ActionKind = "toast"
	Back            ActionKind = "back"
	CloseOverlay    ActionKind = "closeOverlay"
	Invoke          ActionKind = "invoke"
	Navigate        ActionKind = "navigate"
	OpenOverlay     ActionKind = "openOverlay"
	OpenURL         ActionKind = "openUrl"
	PlayPart        ActionKind = "playPart"
	Query           ActionKind = "query"
	Sequence        ActionKind = "sequence"
	SetValue        ActionKind = "setValue"
	Submit          ActionKind = "submit"
)

type Surface string

const (
	Drawer Surface = "drawer"
	Modal  Surface = "modal"
	Sheet  Surface = "sheet"
)

type Tone string

const (
	Neutral     Tone = "neutral"
	ToneAccent  Tone = "accent"
	ToneDanger  Tone = "danger"
	ToneInfo    Tone = "info"
	ToneSuccess Tone = "success"
	ToneWarning Tone = "warning"
)

// Cross-axis alignment.
type BoxStyleAlign string

const (
	Baseline     BoxStyleAlign = "baseline"
	PurpleCenter BoxStyleAlign = "center"
	PurpleEnd    BoxStyleAlign = "end"
	PurpleStart  BoxStyleAlign = "start"
	Stretch      BoxStyleAlign = "stretch"
)

// A colour by ROLE, never a literal. The value behind it comes from the token set the
// Platform serves, so a re-skin changes no payload and no client.
type ColorToken string

const (
	ColorTokenAccent         ColorToken = "accent"
	ColorTokenAccentHover    ColorToken = "accent-hover"
	ColorTokenAccentQuiet    ColorToken = "accent-quiet"
	ColorTokenBg             ColorToken = "bg"
	ColorTokenBorder         ColorToken = "border"
	ColorTokenBorderStrong   ColorToken = "border-strong"
	ColorTokenDanger         ColorToken = "danger"
	ColorTokenDangerQuiet    ColorToken = "danger-quiet"
	ColorTokenInfo           ColorToken = "info"
	ColorTokenInfoQuiet      ColorToken = "info-quiet"
	ColorTokenRating         ColorToken = "rating"
	ColorTokenSuccess        ColorToken = "success"
	ColorTokenSuccessQuiet   ColorToken = "success-quiet"
	ColorTokenSurface        ColorToken = "surface"
	ColorTokenSurfaceOverlay ColorToken = "surface-overlay"
	ColorTokenSurfaceRaised  ColorToken = "surface-raised"
	ColorTokenText           ColorToken = "text"
	ColorTokenTextFaint      ColorToken = "text-faint"
	ColorTokenTextMuted      ColorToken = "text-muted"
	ColorTokenTextOnAccent   ColorToken = "text-on-accent"
	ColorTokenWarning        ColorToken = "warning"
	ColorTokenWarningQuiet   ColorToken = "warning-quiet"
)

// A gradient endpoint: a colour token or "transparent".
//
// A colour by ROLE, never a literal. The value behind it comes from the token set the
// Platform serves, so a re-skin changes no payload and no client.
type GradientStop string

const (
	GradientStopAccent         GradientStop = "accent"
	GradientStopAccentHover    GradientStop = "accent-hover"
	GradientStopAccentQuiet    GradientStop = "accent-quiet"
	GradientStopBg             GradientStop = "bg"
	GradientStopBorder         GradientStop = "border"
	GradientStopBorderStrong   GradientStop = "border-strong"
	GradientStopDanger         GradientStop = "danger"
	GradientStopDangerQuiet    GradientStop = "danger-quiet"
	GradientStopInfo           GradientStop = "info"
	GradientStopInfoQuiet      GradientStop = "info-quiet"
	GradientStopRating         GradientStop = "rating"
	GradientStopSuccess        GradientStop = "success"
	GradientStopSuccessQuiet   GradientStop = "success-quiet"
	GradientStopSurface        GradientStop = "surface"
	GradientStopSurfaceOverlay GradientStop = "surface-overlay"
	GradientStopSurfaceRaised  GradientStop = "surface-raised"
	GradientStopText           GradientStop = "text"
	GradientStopTextFaint      GradientStop = "text-faint"
	GradientStopTextMuted      GradientStop = "text-muted"
	GradientStopTextOnAccent   GradientStop = "text-on-accent"
	GradientStopWarning        GradientStop = "warning"
	GradientStopWarningQuiet   GradientStop = "warning-quiet"
	Transparent                GradientStop = "transparent"
)

// Draw `border` on one edge only — a rule under a row, a marker down the side of the
// selected item. Without it `border` draws all four.
type BorderSide string

const (
	BorderSideBottom BorderSide = "bottom"
	BorderSideTop    BorderSide = "top"
	Left             BorderSide = "left"
	Right            BorderSide = "right"
)

type SpaceTokenEnum string

const (
	Gutter SpaceTokenEnum = "gutter"
)

// grid: flow direction; "column" makes a horizontal rail.
type Direction string

const (
	Column Direction = "column"
	Row    Direction = "row"
)

// A soft bloom cast over this box, screen-blended so it lights the artwork rather than
// tinting it flat. "art" draws the palette sampled from the focused image (the same source
// the acrylic material is lit by), so a hero glows in the colours of its own backdrop;
// "brand" uses the accent pair, for a surface with no artwork to sample. A client with no
// sampler renders "art" as "brand" rather than nothing.
type Glow string

const (
	Art   Glow = "art"
	Brand Glow = "brand"
)

type GridColumnEnum string

const (
	GridColumnAuto GridColumnEnum = "auto"
)

// Main-axis distribution.
type Justify string

const (
	Around        Justify = "around"
	Between       Justify = "between"
	JustifyCenter Justify = "center"
	JustifyEnd    Justify = "end"
	JustifyStart  Justify = "start"
)

// Layout mode. "grid" enables the grid-* fields.
type Layout string

const (
	Flex Layout = "flex"
	Grid Layout = "grid"
)

type Overflow string

const (
	Hidden       Overflow = "hidden"
	OverflowAuto Overflow = "auto"
	Visible      Overflow = "visible"
)

type Position string

const (
	Absolute Position = "absolute"
	Fixed    Position = "fixed"
	Relative Position = "relative"
	Sticky   Position = "sticky"
)

// A corner radius from the scale.
type RadiusToken string

const (
	Pill          RadiusToken = "pill"
	RadiusTokenLg RadiusToken = "lg"
	RadiusTokenMd RadiusToken = "md"
	RadiusTokenSm RadiusToken = "sm"
	RadiusTokenXl RadiusToken = "xl"
)

// A named legibility wash over artwork, so text laid on a backdrop stays readable whatever
// the image behind it. NAMED rather than a gradient the server writes, for the same reason
// colours are tokens: the recipe is a formula each client evaluates over its own tokens,
// and one written as literal stops would be a second skin the Platform owns. "bottom"/"top"
// fade an edge into the page; "leading" washes the text side (the direction follows the
// reading direction, so it mirrors in RTL); "cinematic" is both — the full-bleed hero
// treatment.
type Scrim string

const (
	Cinematic   Scrim = "cinematic"
	Leading     Scrim = "leading"
	ScrimBottom Scrim = "bottom"
	ScrimTop    Scrim = "top"
)

// An elevation step.
type ShadowToken string

const (
	The1 ShadowToken = "1"
	The2 ShadowToken = "2"
	The3 ShadowToken = "3"
)

// Scroll snapping axis, for carousels.
type Snap string

const (
	X Snap = "x"
	Y Snap = "y"
)

type SnapAlign string

const (
	SnapAlignCenter SnapAlign = "center"
	SnapAlignStart  SnapAlign = "start"
)

// Portable stack order for chrome that overlays content — not an arbitrary index.
type Z string

const (
	Overlay Z = "overlay"
	Raised  Z = "raised"
	ZToast  Z = "toast"
)

type TextStyleAlign string

const (
	FluffyCenter TextStyleAlign = "center"
	FluffyEnd    TextStyleAlign = "end"
	FluffyStart  TextStyleAlign = "start"
)

// Letter-spacing: tight for display headings, wide for an eyebrow.
type Tracking string

const (
	Normal Tracking = "normal"
	Tight  Tracking = "tight"
	Wide   Tracking = "wide"
)

type Transform string

const (
	Capitalize Transform = "capitalize"
	None       Transform = "none"
	Uppercase  Transform = "uppercase"
)

// A step on the type scale.
type TextVariant string

const (
	TextVariantLg TextVariant = "lg"
	TextVariantMd TextVariant = "md"
	TextVariantSm TextVariant = "sm"
	TextVariantXl TextVariant = "xl"
	The2Xl        TextVariant = "2xl"
	The3Xl        TextVariant = "3xl"
	The4Xl        TextVariant = "4xl"
	Xs            TextVariant = "xs"
)

// A weight on the type scale.
type FontWeight string

const (
	Bold    FontWeight = "bold"
	Medium  FontWeight = "medium"
	Regular FontWeight = "regular"
)

// The border's weight. Two steps, not a number: a hairline rule and a marker heavy enough
// to read as a selection. Anything thicker is a filled box, which this vocabulary already
// has.
type BorderWidth struct {
	Double  *float64
	Integer *int64
}

// A step on the spacing scale (0–9), or "gutter" — the fluid page margin that clamps with
// the viewport, so page padding is responsive without a breakpoint.
//
// Pull this box up over what precedes it by one step of the spacing scale, so a content
// sheet rides over the bottom of a full-bleed hero. A single direction and a token step,
// rather than negative margins in four directions: this is the one case in the layout where
// two siblings are meant to intersect, and naming it keeps it from becoming an
// arbitrary-offset escape hatch. Pair with `z` to say which one is in front.
type SpaceToken struct {
	Enum    *SpaceTokenEnum
	Integer *int64
}

type GridColumnElement struct {
	Double          *float64
	Enum            *GridColumnEnum
	GridColumnClass *GridColumnClass
}

// A size: a number of pixels, "full" (100% of the parent), "screen" (the viewport in that
// axis — the one non-parent-relative size, for full-bleed surfaces and the app frame), "NN%
// screen" (a fraction of that same viewport axis), "auto", or a plain percentage string (of
// the parent). The viewport fraction is spelled out rather than left to a plain percentage
// because the two resolve against different things and the difference is invisible until it
// is catastrophic: a hero asking for 88% of a parent whose height is the page's own content
// grows to 88% of the whole scroll length.
type Dimension struct {
	Double *float64
	String *string
}
