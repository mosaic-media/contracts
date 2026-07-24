// Code generated from schema/sdui.schema.json by quicktype. DO NOT EDIT.
package contract

// The single source of truth for the Mosaic Server-Driven-UI contract. Language bindings
// (Go, TypeScript, Dart) are GENERATED from this file — do not hand-edit them. The root
// object exists only so a generator reaches every top-level type; the useful types are in
// $defs.

// A declarative behaviour envelope. Data, never code — the client interprets the kind. Each
// kind uses a subset of the fields.
type Action struct {
	Actions  []Action               `json:"actions,omitempty"`
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
}

// One element of a server-driven UI tree. The `type` is an open vocabulary: a client that
// does not recognise a type renders a placeholder rather than failing.
type UINode struct {
	Children []UINode `json:"children,omitempty"`
	ID       *string  `json:"id,omitempty"`
	// Component-specific data. Open by design.
	Props map[string]interface{} `json:"props,omitempty"`
	Slots map[string][]UINode    `json:"slots,omitempty"`
	// Component discriminator, e.g. "PosterCard".
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
	Bottom      *SpaceToken `json:"bottom"`
	Color       *ColorToken `json:"color,omitempty"`
	Direction   *Direction  `json:"direction,omitempty"`
	Flex        *float64    `json:"flex,omitempty"`
	Gap         *SpaceToken `json:"gap"`
	// Render this surface in the acrylic material: translucent, blurred, and lit by the current
	// light source (the focused artwork, or the brand light when there is none). A client
	// without the material renders a plain translucent surface.
	Glass *bool `json:"glass,omitempty"`
	// grid: fixed track size for a rail, in px.
	GridAutoColumns *float64 `json:"gridAutoColumns,omitempty"`
	// grid: flow direction; "column" makes a horizontal rail.
	GridFlow *Direction `json:"gridFlow,omitempty"`
	// grid: responsive auto-fill columns of at least this width in px.
	GridMin *float64   `json:"gridMin,omitempty"`
	Grow    *bool      `json:"grow,omitempty"`
	Height  *Dimension `json:"height"`
	// Take the box out of the layout entirely. Its purpose is `responsive`: one payload carries
	// both a desktop and a phone arrangement, and each viewport drops the half it does not use.
	Hidden *bool `json:"hidden,omitempty"`
	// Main-axis distribution.
	Justify *Justify `json:"justify,omitempty"`
	// DEPRECATED in favour of `responsive`. A named hook a client may map to a rule of its own;
	// it makes a layout depend on client CSS, which is what `responsive` exists to avoid.
	Kind *string `json:"kind,omitempty"`
	// Layout mode. "grid" enables the grid-* fields.
	Layout    *Layout      `json:"layout,omitempty"`
	Left      *SpaceToken  `json:"left"`
	MaxWidth  *Dimension   `json:"maxWidth"`
	MinHeight *Dimension   `json:"minHeight"`
	MinWidth  *Dimension   `json:"minWidth"`
	Opacity   *float64     `json:"opacity,omitempty"`
	Overflow  *Overflow    `json:"overflow,omitempty"`
	OverflowX *Overflow    `json:"overflowX,omitempty"`
	OverflowY *Overflow    `json:"overflowY,omitempty"`
	P         *SpaceToken  `json:"p"`
	Pb        *SpaceToken  `json:"pb"`
	Pl        *SpaceToken  `json:"pl"`
	Position  *Position    `json:"position,omitempty"`
	PR        *SpaceToken  `json:"pr"`
	Pt        *SpaceToken  `json:"pt"`
	Px        *SpaceToken  `json:"px"`
	Py        *SpaceToken  `json:"py"`
	Radius    *RadiusToken `json:"radius,omitempty"`
	// A style override applied below a viewport width — the vocabulary's one responsive
	// capability, and what lets a layout adapt as DATA rather than through a client stylesheet.
	// The override is a plain BoxStyle merged over the base: a field it does not mention keeps
	// its base value, and null clears one (these travel as JSON, where an undefined member
	// would vanish silently). One breakpoint, not a cascade.
	Responsive *Responsive  `json:"responsive,omitempty"`
	Right      *SpaceToken  `json:"right"`
	Shadow     *ShadowToken `json:"shadow,omitempty"`
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

// A component expressed as data: a name, default params, and a template of primitives.
// Clients register definitions and expand them; this is how a module contributes a
// component without shipping client code. A template node's props may hold binding objects
// ({"$bind":"path"} / {"$match":{…}}) and control keys ($if / $ifNot / $each / $as); a node
// of type "Outlet" renders the caller's children or a named slot.
type ComponentDefinition struct {
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
	// Tabular figures, so numbers in a column line up.
	Tabular *bool `json:"tabular,omitempty"`
	// Letter-spacing: tight for display headings, wide for an eyebrow.
	Tracking  *Tracking    `json:"tracking,omitempty"`
	Transform *Transform   `json:"transform,omitempty"`
	Variant   *TextVariant `json:"variant,omitempty"`
	Weight    *FontWeight  `json:"weight,omitempty"`
}

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
	Sequence        ActionKind = "sequence"
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
	Auto    Overflow = "auto"
	Hidden  Overflow = "hidden"
	Visible Overflow = "visible"
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

// A step on the spacing scale (0–9), or "gutter" — the fluid page margin that clamps with
// the viewport, so page padding is responsive without a breakpoint.
type SpaceToken struct {
	Enum    *SpaceTokenEnum
	Integer *int64
}

// A size: a number of pixels, "full" (100% of the parent), "screen" (the viewport in that
// axis — the one non-parent-relative size, for full-bleed surfaces and the app frame),
// "auto", or a percentage string.
type Dimension struct {
	Double *float64
	String *string
}
