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
     * setValue: the name of the field written in the enclosing state scope.
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
 */
export interface UINode {
    children?: UINode[];
    id?:       string;
    /**
     * Component-specific data. Open by design.
     */
    props?: { [key: string]: any };
    slots?: { [key: string]: UINode[] };
    /**
     * Component discriminator, e.g. "PosterCard".
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
 * growing it is the only change that requires a client release (ADR 0024).
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
     * grid: fixed track size for a rail, in px.
     */
    gridAutoColumns?: number;
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
    p?:         SpaceTokenEnum | number;
    pb?:        SpaceTokenEnum | number;
    pl?:        SpaceTokenEnum | number;
    position?:  Position;
    pr?:        SpaceTokenEnum | number;
    pt?:        SpaceTokenEnum | number;
    px?:        SpaceTokenEnum | number;
    py?:        SpaceTokenEnum | number;
    radius?:    RadiusToken;
    /**
     * A style override applied below a viewport width — the vocabulary's one responsive
     * capability, and what lets a layout adapt as DATA rather than through a client stylesheet.
     * The override is a plain BoxStyle merged over the base: a field it does not mention keeps
     * its base value, and null clears one (these travel as JSON, where an undefined member
     * would vanish silently). One breakpoint, not a cascade.
     */
    responsive?: Responsive;
    right?:      SpaceTokenEnum | number;
    shadow?:     ShadowToken;
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
