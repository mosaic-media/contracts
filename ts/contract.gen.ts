// Code generated from schema/sdui.schema.json by quicktype. DO NOT EDIT.
/**
 * The single source of truth for the Mosaic Server-Driven-UI contract. Language bindings
 * (Go, TypeScript, Dart) are GENERATED from this file — do not hand-edit them. The root
 * object exists only so a generator reaches every top-level type; the useful types are in
 * $defs.
 */

/**
 * A declarative behaviour envelope. Data, never code — the client interprets the kind. Each
 * kind uses a subset of the fields.
 */
export interface Action {
    actions?: Action[];
    /**
     * setValue: the name of the field written in the enclosing state scope. submit: where in
     * the action input the collected values merge — a path into it, or absent for its top level.
     */
    field?:    string;
    input?:    { [key: string]: any };
    kind:      ActionKind;
    message?:  string;
    mutation?: string;
    node?:     UINode;
    nodeId?:   string;
    params?:   { [key: string]: any };
    partId?:   string;
    screen?:   string;
    surface?:  Surface;
    tone?:     Tone;
    url?:      string;
    /**
     * setValue: the value written.
     */
    value?: string;
}

/**
 * The behaviours a client interprets. Kept in step with ui.spec.json and
 * proto/mosaic/sdui/v1/sdui.proto by `go run ./tools/genui -lint`, which fails when the
 * three disagree — they had drifted to ten, nine and four before that gate existed.
 */
export enum ActionKind {
    Back = "back",
    CloseOverlay = "closeOverlay",
    Invoke = "invoke",
    Navigate = "navigate",
    OpenOverlay = "openOverlay",
    OpenURL = "openUrl",
    PlayPart = "playPart",
    Query = "query",
    Sequence = "sequence",
    SetValue = "setValue",
    Submit = "submit",
    Toast = "toast",
}

/**
 * One element of a server-driven UI tree. The `type` is an open vocabulary: a client that
 * does not recognise a type renders a placeholder rather than failing.
 *
 * An alternative template for a client whose declared vocabulary lacks a primitive the main
 * template needs. The server picks per session and sends one; the client never sees both.
 * Optional — a definition without one is served unchanged to everyone.
 */
export interface UINode {
    children?: UINode[];
    /**
     * A stable identity for this node. It is the React key and the target of a narrow region
     * update, and it is also the *analytics* identity: an onAppear action names what was seen
     * by carrying whatever the server put here, so an id that is a row index attributes nothing
     * and one that names the thing attributes everything. The server chooses it, because only
     * the server knows what the node is about.
     */
    id?: string;
    /**
     * Component-specific data. Open by design. Any value may be a literal or a Binding (see
     * #/$defs/Binding), which the client resolves where the node renders.
     */
    props?: { [key: string]: any };
    slots?: { [key: string]: UINode[] };
    /**
     * Component discriminator, e.g. "PosterCard". Namespaced: a core type — every primitive and
     * every component in the published vocabulary — is unprefixed and never contains the type
     * separator; a module's own type is "moduleId:Type". Open and flat is not the same as open:
     * without the namespace two modules could both contribute a StatChip, and one could
     * contribute a PosterCard and take the core component's place in every client's registry.
     */
    type: string;
}

export enum Surface {
    Drawer = "drawer",
    Modal = "modal",
    Sheet = "sheet",
}

export enum Tone {
    Accent = "accent",
    Danger = "danger",
    Info = "info",
    Neutral = "neutral",
    Success = "success",
    Warning = "warning",
}

/**
 * A style override applied below a viewport width — the vocabulary's one responsive
 * capability, and what lets a layout adapt as DATA rather than through a client stylesheet.
 * The override is a plain BoxStyle merged over the base: a field it does not mention keeps
 * its base value, and null clears one (these travel as JSON, where an undefined member
 * would vanish silently). One breakpoint, not a cascade.
 */
export interface Responsive {
    /**
     * Viewport width in px, below which the override applies.
     */
    below: number;
    style: BoxStyle;
}

/**
 * Layout and surface styling for a Box, in TOKENS ONLY — no literal colours, no raw pixels
 * except explicit dimensions. This is the technology-agnostic seam: it is the intersection
 * of what a flexbox client and a Flutter client can render identically, so a definition
 * written once renders the same everywhere. A client implements this vocabulary natively;
 * growing it is the only change that requires a client release (contracts#2).
 */
export interface BoxStyle {
    /**
     * Cross-axis alignment.
     */
    align?: BoxStyleAlign;
    /**
     * An aspect ratio as "w / h".
     */
    aspectRatio?: string;
    bg?:          ColorToken;
    /**
     * A linear gradient between two token colours (or "transparent"), at an angle in degrees.
     */
    bgGradient?:  BgGradient;
    border?:      boolean;
    borderColor?: ColorToken;
    /**
     * Draw `border` on one edge only — a rule under a row, a marker down the side of the
     * selected item. Without it `border` draws all four.
     */
    borderSide?: BorderSide;
    /**
     * The border's weight. Two steps, not a number: a hairline rule and a marker heavy enough
     * to read as a selection. Anything thicker is a filled box, which this vocabulary already
     * has.
     */
    borderWidth?: number | number;
    bottom?:      SpaceTokenEnum | number;
    color?:       ColorToken;
    direction?:   Direction;
    flex?:        number;
    gap?:         SpaceTokenEnum | number;
    /**
     * Render this surface in the acrylic material: translucent, blurred, and lit by the current
     * light source (the focused artwork, or the brand light when there is none). A client
     * without the material renders a plain translucent surface.
     */
    glass?: boolean;
    /**
     * A soft bloom cast over this box, screen-blended so it lights the artwork rather than
     * tinting it flat. "art" draws the palette sampled from the focused image (the same source
     * the acrylic material is lit by), so a hero glows in the colours of its own backdrop;
     * "brand" uses the accent pair, for a surface with no artwork to sample. A client with no
     * sampler renders "art" as "brand" rather than nothing.
     */
    glow?: Glow;
    /**
     * Overlay the material texture — the grain, blotch and scuff the token set carries — so a
     * large flat surface reads as a material rather than as a fill. Soft-light blended over
     * whatever is beneath, and purely decorative: it never affects layout or hit-testing.
     */
    grain?: boolean;
    /**
     * grid: fixed track size for a rail, in px.
     */
    gridAutoColumns?: number;
    /**
     * grid: an explicit column track list, for the arrangements auto-fill cannot state — a
     * settings frame's nav/panel/aside, an episode row's fixed thumbnail beside a fluid title.
     * A track is a number of pixels (fixed), "auto" (sized to its content), or {"fill": n} (n
     * shares of what is left). Deliberately not a CSS grid-template string: these three are
     * what a flexbox client and a Flutter client can both lay out identically.
     */
    gridColumns?: Array<GridColumnClass | number | GridColumnEnum>;
    /**
     * grid: flow direction; "column" makes a horizontal rail.
     */
    gridFlow?: Direction;
    /**
     * grid: responsive auto-fill columns of at least this width in px.
     */
    gridMin?: number;
    grow?:    boolean;
    height?:  number | string;
    /**
     * Take the box out of the layout entirely. Its purpose is `responsive`: one payload carries
     * both a desktop and a phone arrangement, and each viewport drops the half it does not use.
     */
    hidden?: boolean;
    /**
     * Mark this box as the region whose hover or focus reveals the `hoverReveal` boxes inside
     * it. Stated explicitly, and on the ancestor rather than inferred from interactivity,
     * because the thing revealed is frequently not inside the thing hovered: a rail's "see all"
     * appears when the *section* is approached, and a link that appeared only while pointing at
     * it could never be clicked.
     */
    hoverGroup?: boolean;
    /**
     * Hide this box until its nearest `hoverGroup` ancestor is hovered or focused, then fade it
     * in. It is the card veil: the extra detail a tile shows on approach — time remaining, file
     * size, a play affordance — which must not be in the resting composition and must not cost
     * a second payload. An input model with no pointer reveals it on focus instead, and a
     * client with neither renders it always-visible rather than never: unreachable detail is
     * the worse failure.
     */
    hoverReveal?: boolean;
    /**
     * Main-axis distribution.
     */
    justify?: Justify;
    /**
     * DEPRECATED in favour of `responsive`. A named hook a client may map to a rule of its own;
     * it makes a layout depend on client CSS, which is what `responsive` exists to avoid.
     */
    kind?: string;
    /**
     * Layout mode. "grid" enables the grid-* fields.
     */
    layout?:    Layout;
    left?:      SpaceTokenEnum | number;
    maxWidth?:  number | string;
    minHeight?: number | string;
    minWidth?:  number | string;
    opacity?:   number;
    overflow?:  Overflow;
    overflowX?: Overflow;
    overflowY?: Overflow;
    /**
     * Pull this box up over what precedes it by one step of the spacing scale, so a content
     * sheet rides over the bottom of a full-bleed hero. A single direction and a token step,
     * rather than negative margins in four directions: this is the one case in the layout where
     * two siblings are meant to intersect, and naming it keeps it from becoming an
     * arbitrary-offset escape hatch. Pair with `z` to say which one is in front.
     */
    overlap?:  SpaceTokenEnum | number;
    p?:        SpaceTokenEnum | number;
    pb?:       SpaceTokenEnum | number;
    pl?:       SpaceTokenEnum | number;
    position?: Position;
    pr?:       SpaceTokenEnum | number;
    pt?:       SpaceTokenEnum | number;
    px?:       SpaceTokenEnum | number;
    py?:       SpaceTokenEnum | number;
    radius?:   RadiusToken;
    /**
     * A style override applied below a viewport width — the vocabulary's one responsive
     * capability, and what lets a layout adapt as DATA rather than through a client stylesheet.
     * The override is a plain BoxStyle merged over the base: a field it does not mention keeps
     * its base value, and null clears one (these travel as JSON, where an undefined member
     * would vanish silently). One breakpoint, not a cascade.
     */
    responsive?: Responsive;
    right?:      SpaceTokenEnum | number;
    /**
     * A named legibility wash over artwork, so text laid on a backdrop stays readable whatever
     * the image behind it. NAMED rather than a gradient the server writes, for the same reason
     * colours are tokens: the recipe is a formula each client evaluates over its own tokens,
     * and one written as literal stops would be a second skin the Platform owns. "bottom"/"top"
     * fade an edge into the page; "leading" washes the text side (the direction follows the
     * reading direction, so it mirrors in RTL); "cinematic" is both — the full-bleed hero
     * treatment.
     */
    scrim?:  Scrim;
    shadow?: ShadowToken;
    /**
     * Scroll snapping axis, for carousels.
     */
    snap?:      Snap;
    snapAlign?: SnapAlign;
    top?:       SpaceTokenEnum | number;
    width?:     number | string;
    wrap?:      boolean;
    /**
     * Portable stack order for chrome that overlays content — not an arbitrary index.
     */
    z?: Z;
}

/**
 * Cross-axis alignment.
 */
export enum BoxStyleAlign {
    Baseline = "baseline",
    Center = "center",
    End = "end",
    Start = "start",
    Stretch = "stretch",
}

/**
 * A colour by ROLE, never a literal. The value behind it comes from the token set the
 * Platform serves, so a re-skin changes no payload and no client.
 */
export enum ColorToken {
    Accent = "accent",
    AccentHover = "accent-hover",
    AccentQuiet = "accent-quiet",
    Bg = "bg",
    Border = "border",
    BorderStrong = "border-strong",
    Danger = "danger",
    DangerQuiet = "danger-quiet",
    Info = "info",
    InfoQuiet = "info-quiet",
    Rating = "rating",
    Success = "success",
    SuccessQuiet = "success-quiet",
    Surface = "surface",
    SurfaceOverlay = "surface-overlay",
    SurfaceRaised = "surface-raised",
    Text = "text",
    TextFaint = "text-faint",
    TextMuted = "text-muted",
    TextOnAccent = "text-on-accent",
    Warning = "warning",
    WarningQuiet = "warning-quiet",
}

/**
 * A linear gradient between two token colours (or "transparent"), at an angle in degrees.
 */
export interface BgGradient {
    angle?: number;
    from:   GradientStop;
    to:     GradientStop;
}

/**
 * A gradient endpoint: a colour token or "transparent".
 *
 * A colour by ROLE, never a literal. The value behind it comes from the token set the
 * Platform serves, so a re-skin changes no payload and no client.
 */
export enum GradientStop {
    Accent = "accent",
    AccentHover = "accent-hover",
    AccentQuiet = "accent-quiet",
    Bg = "bg",
    Border = "border",
    BorderStrong = "border-strong",
    Danger = "danger",
    DangerQuiet = "danger-quiet",
    Info = "info",
    InfoQuiet = "info-quiet",
    Rating = "rating",
    Success = "success",
    SuccessQuiet = "success-quiet",
    Surface = "surface",
    SurfaceOverlay = "surface-overlay",
    SurfaceRaised = "surface-raised",
    Text = "text",
    TextFaint = "text-faint",
    TextMuted = "text-muted",
    TextOnAccent = "text-on-accent",
    Transparent = "transparent",
    Warning = "warning",
    WarningQuiet = "warning-quiet",
}

/**
 * Draw `border` on one edge only — a rule under a row, a marker down the side of the
 * selected item. Without it `border` draws all four.
 */
export enum BorderSide {
    Bottom = "bottom",
    Left = "left",
    Right = "right",
    Top = "top",
}

export enum SpaceTokenEnum {
    Gutter = "gutter",
}

/**
 * grid: flow direction; "column" makes a horizontal rail.
 */
export enum Direction {
    Column = "column",
    Row = "row",
}

/**
 * A soft bloom cast over this box, screen-blended so it lights the artwork rather than
 * tinting it flat. "art" draws the palette sampled from the focused image (the same source
 * the acrylic material is lit by), so a hero glows in the colours of its own backdrop;
 * "brand" uses the accent pair, for a surface with no artwork to sample. A client with no
 * sampler renders "art" as "brand" rather than nothing.
 */
export enum Glow {
    Art = "art",
    Brand = "brand",
}

export interface GridColumnClass {
    fill: number;
}

export enum GridColumnEnum {
    Auto = "auto",
}

/**
 * Main-axis distribution.
 */
export enum Justify {
    Around = "around",
    Between = "between",
    Center = "center",
    End = "end",
    Start = "start",
}

/**
 * Layout mode. "grid" enables the grid-* fields.
 */
export enum Layout {
    Flex = "flex",
    Grid = "grid",
}

export enum Overflow {
    Auto = "auto",
    Hidden = "hidden",
    Visible = "visible",
}

export enum Position {
    Absolute = "absolute",
    Fixed = "fixed",
    Relative = "relative",
    Sticky = "sticky",
}

/**
 * A corner radius from the scale.
 */
export enum RadiusToken {
    Lg = "lg",
    Md = "md",
    Pill = "pill",
    Sm = "sm",
    Xl = "xl",
}

/**
 * A named legibility wash over artwork, so text laid on a backdrop stays readable whatever
 * the image behind it. NAMED rather than a gradient the server writes, for the same reason
 * colours are tokens: the recipe is a formula each client evaluates over its own tokens,
 * and one written as literal stops would be a second skin the Platform owns. "bottom"/"top"
 * fade an edge into the page; "leading" washes the text side (the direction follows the
 * reading direction, so it mirrors in RTL); "cinematic" is both — the full-bleed hero
 * treatment.
 */
export enum Scrim {
    Bottom = "bottom",
    Cinematic = "cinematic",
    Leading = "leading",
    Top = "top",
}

/**
 * An elevation step.
 */
export enum ShadowToken {
    The1 = "1",
    The2 = "2",
    The3 = "3",
}

/**
 * Scroll snapping axis, for carousels.
 */
export enum Snap {
    X = "x",
    Y = "y",
}

export enum SnapAlign {
    Center = "center",
    Start = "start",
}

/**
 * Portable stack order for chrome that overlays content — not an arbitrary index.
 */
export enum Z {
    Overlay = "overlay",
    Raised = "raised",
    Toast = "toast",
}

/**
 * A component expressed as data: a name, default params, and a template of primitives.
 * Clients register definitions and expand them; this is how a module contributes a
 * component without shipping client code. A template node's props may hold binding objects
 * ({"$bind":"path"} / {"$match":{…}}) and control keys ($if / $ifNot / $each / $as); a node
 * of type "Outlet" renders the caller's children or a named slot.
 */
export interface ComponentDefinition {
    /**
     * An alternative template for a client whose declared vocabulary lacks a primitive the main
     * template needs. The server picks per session and sends one; the client never sees both.
     * Optional — a definition without one is served unchanged to everyone.
     */
    fallback?: UINode;
    /**
     * The node type this definition provides.
     */
    name: string;
    /**
     * Default param values, overridden by the caller's props.
     */
    params?:  { [key: string]: any };
    template: UINode;
}

/**
 * Typography for a Text node, in tokens only.
 */
export interface TextStyle {
    align?:  TextStyleAlign;
    color?:  ColorToken;
    italic?: boolean;
    /**
     * Truncate after this many lines.
     */
    lineClamp?: number;
    mono?:      boolean;
    /**
     * A legibility shadow behind the glyphs, for text laid directly over artwork with no scrim
     * beneath it — a title on a tile, a caption on a still. Boolean rather than a shadow spec:
     * the only question a payload gets to answer is whether the text is over an image, and how
     * that is drawn is the client's.
     */
    shadow?: boolean;
    /**
     * Tabular figures, so numbers in a column line up.
     */
    tabular?: boolean;
    /**
     * Letter-spacing: tight for display headings, wide for an eyebrow.
     */
    tracking?:  Tracking;
    transform?: Transform;
    variant?:   TextVariant;
    weight?:    FontWeight;
}

export enum TextStyleAlign {
    Center = "center",
    End = "end",
    Start = "start",
}

/**
 * Letter-spacing: tight for display headings, wide for an eyebrow.
 */
export enum Tracking {
    Normal = "normal",
    Tight = "tight",
    Wide = "wide",
}

export enum Transform {
    Capitalize = "capitalize",
    None = "none",
    Uppercase = "uppercase",
}

/**
 * A step on the type scale.
 */
export enum TextVariant {
    Lg = "lg",
    Md = "md",
    Sm = "sm",
    The2Xl = "2xl",
    The3Xl = "3xl",
    The4Xl = "4xl",
    Xl = "xl",
    Xs = "xs",
}

/**
 * A weight on the type scale.
 */
export enum FontWeight {
    Bold = "bold",
    Medium = "medium",
    Regular = "regular",
}
