package shadcn

// getDefaultConfig is a faithful port of tailwind-merge v3.6.0's
// getDefaultConfig (src/lib/default-config.ts), targeting Tailwind CSS v4.
//
// The class group ids, theme scales and conflict tables are transcribed
// verbatim from the upstream source. classGroups MUST stay in source order:
// the order in which groups are processed determines validator precedence
// inside the class map.

// g builds a class group entry whose group is a list of root-level
// definitions (plain classes, validators, or multi-key objects).
func g(id string, group classGroup) classGroupEntry {
	return classGroupEntry{id, group}
}

// grp builds the common `[{ key: items }]` class group entry.
func grp(id, key string, items classGroup) classGroupEntry {
	return classGroupEntry{id, classGroup{classObject{key: items}}}
}

func getDefaultConfig() *twConfig {
	// Theme getters for theme variable namespaces.
	themeColor := themeGetter{"color"}
	themeFont := themeGetter{"font"}
	themeText := themeGetter{"text"}
	themeFontWeight := themeGetter{"font-weight"}
	themeTracking := themeGetter{"tracking"}
	themeLeading := themeGetter{"leading"}
	themeBreakpoint := themeGetter{"breakpoint"}
	themeContainer := themeGetter{"container"}
	themeSpacing := themeGetter{"spacing"}
	themeRadius := themeGetter{"radius"}
	themeShadow := themeGetter{"shadow"}
	themeInsetShadow := themeGetter{"inset-shadow"}
	themeTextShadow := themeGetter{"text-shadow"}
	themeDropShadow := themeGetter{"drop-shadow"}
	themeBlur := themeGetter{"blur"}
	themePerspective := themeGetter{"perspective"}
	themeAspect := themeGetter{"aspect"}
	themeEase := themeGetter{"ease"}
	themeAnimate := themeGetter{"animate"}

	// Scale helpers. Each returns a fresh slice on every call, matching the
	// upstream behavior that guards against accidental shared mutation.
	scaleBreak := func() classGroup {
		return classGroup{"auto", "avoid", "all", "avoid-page", "page", "left", "right", "column"}
	}
	scalePosition := func() classGroup {
		return classGroup{
			"center", "top", "bottom", "left", "right",
			"top-left", "left-top", "top-right", "right-top",
			"bottom-right", "right-bottom", "bottom-left", "left-bottom",
		}
	}
	scalePositionWithArbitrary := func() classGroup {
		return append(scalePosition(), isArbitraryVariable, isArbitraryValue)
	}
	scaleOverflow := func() classGroup {
		return classGroup{"auto", "hidden", "clip", "visible", "scroll"}
	}
	scaleOverscroll := func() classGroup {
		return classGroup{"auto", "contain", "none"}
	}
	scaleUnambiguousSpacing := func() classGroup {
		return classGroup{isArbitraryVariable, isArbitraryValue, themeSpacing}
	}
	scaleInset := func() classGroup {
		return append(classGroup{isFraction, "full", "auto"}, scaleUnambiguousSpacing()...)
	}
	scaleGridTemplateColsRows := func() classGroup {
		return classGroup{isInteger, "none", "subgrid", isArbitraryVariable, isArbitraryValue}
	}
	scaleGridColRowStartAndEnd := func() classGroup {
		return classGroup{
			"auto",
			classObject{"span": classGroup{"full", isInteger, isArbitraryVariable, isArbitraryValue}},
			isInteger, isArbitraryVariable, isArbitraryValue,
		}
	}
	scaleGridColRowStartOrEnd := func() classGroup {
		return classGroup{isInteger, "auto", isArbitraryVariable, isArbitraryValue}
	}
	scaleGridAutoColsRows := func() classGroup {
		return classGroup{"auto", "min", "max", "fr", isArbitraryVariable, isArbitraryValue}
	}
	scaleAlignPrimaryAxis := func() classGroup {
		return classGroup{
			"start", "end", "center", "between", "around", "evenly",
			"stretch", "baseline", "center-safe", "end-safe",
		}
	}
	scaleAlignSecondaryAxis := func() classGroup {
		return classGroup{"start", "end", "center", "stretch", "center-safe", "end-safe"}
	}
	scaleMargin := func() classGroup {
		return append(classGroup{"auto"}, scaleUnambiguousSpacing()...)
	}
	scaleSizing := func() classGroup {
		return append(classGroup{
			isFraction, "auto", "full", "dvw", "dvh", "lvw", "lvh",
			"svw", "svh", "min", "max", "fit",
		}, scaleUnambiguousSpacing()...)
	}
	scaleSizingInline := func() classGroup {
		return append(classGroup{
			isFraction, "screen", "full", "dvw", "lvw", "svw", "min", "max", "fit",
		}, scaleUnambiguousSpacing()...)
	}
	scaleSizingBlock := func() classGroup {
		return append(classGroup{
			isFraction, "screen", "full", "lh", "dvh", "lvh", "svh", "min", "max", "fit",
		}, scaleUnambiguousSpacing()...)
	}
	scaleColor := func() classGroup {
		return classGroup{themeColor, isArbitraryVariable, isArbitraryValue}
	}
	scaleBgPosition := func() classGroup {
		return append(scalePosition(),
			isArbitraryVariablePosition, isArbitraryPosition,
			classObject{"position": classGroup{isArbitraryVariable, isArbitraryValue}},
		)
	}
	scaleBgRepeat := func() classGroup {
		return classGroup{
			"no-repeat",
			classObject{"repeat": classGroup{"", "x", "y", "space", "round"}},
		}
	}
	scaleBgSize := func() classGroup {
		return classGroup{
			"auto", "cover", "contain", isArbitraryVariableSize, isArbitrarySize,
			classObject{"size": classGroup{isArbitraryVariable, isArbitraryValue}},
		}
	}
	scaleGradientStopPosition := func() classGroup {
		return classGroup{isPercent, isArbitraryVariableLength, isArbitraryLength}
	}
	scaleRadius := func() classGroup {
		return classGroup{"", "none", "full", themeRadius, isArbitraryVariable, isArbitraryValue}
	}
	scaleBorderWidth := func() classGroup {
		return classGroup{"", isNumber, isArbitraryVariableLength, isArbitraryLength}
	}
	scaleLineStyle := func() classGroup {
		return classGroup{"solid", "dashed", "dotted", "double"}
	}
	scaleBlendMode := func() classGroup {
		return classGroup{
			"normal", "multiply", "screen", "overlay", "darken", "lighten",
			"color-dodge", "color-burn", "hard-light", "soft-light", "difference",
			"exclusion", "hue", "saturation", "color", "luminosity",
		}
	}
	scaleMaskImagePosition := func() classGroup {
		return classGroup{isNumber, isPercent, isArbitraryVariablePosition, isArbitraryPosition}
	}
	scaleBlur := func() classGroup {
		return classGroup{"", "none", themeBlur, isArbitraryVariable, isArbitraryValue}
	}
	scaleRotate := func() classGroup {
		return classGroup{"none", isNumber, isArbitraryVariable, isArbitraryValue}
	}
	scaleScale := func() classGroup {
		return classGroup{"none", isNumber, isArbitraryVariable, isArbitraryValue}
	}
	scaleSkew := func() classGroup {
		return classGroup{isNumber, isArbitraryVariable, isArbitraryValue}
	}
	scaleTranslate := func() classGroup {
		return append(classGroup{isFraction, "full"}, scaleUnambiguousSpacing()...)
	}

	theme := map[string]classGroup{
		"animate":      {"spin", "ping", "pulse", "bounce"},
		"aspect":       {"video"},
		"blur":         {isTshirtSize},
		"breakpoint":   {isTshirtSize},
		"color":        {isAny},
		"container":    {isTshirtSize},
		"drop-shadow":  {isTshirtSize},
		"ease":         {"in", "out", "in-out"},
		"font":         {isAnyNonArbitrary},
		"font-weight":  {"thin", "extralight", "light", "normal", "medium", "semibold", "bold", "extrabold", "black"},
		"inset-shadow": {isTshirtSize},
		"leading":      {"none", "tight", "snug", "normal", "relaxed", "loose"},
		"perspective":  {"dramatic", "near", "normal", "midrange", "distant", "none"},
		"radius":       {isTshirtSize},
		"shadow":       {isTshirtSize},
		"spacing":      {"px", isNumber},
		"text":         {isTshirtSize},
		"text-shadow":  {isTshirtSize},
		"tracking":     {"tighter", "tight", "normal", "wide", "wider", "widest"},
	}

	classGroups := []classGroupEntry{
		// --- Layout ---
		grp("aspect", "aspect", classGroup{"auto", "square", isFraction, isArbitraryValue, isArbitraryVariable, themeAspect}),
		g("container", classGroup{"container"}),
		grp("container-type", "@container", classGroup{"", "normal", "size", isArbitraryVariable, isArbitraryValue}),
		g("container-named", classGroup{isNamedContainerQuery}),
		grp("columns", "columns", classGroup{isNumber, isArbitraryValue, isArbitraryVariable, themeContainer}),
		grp("break-after", "break-after", scaleBreak()),
		grp("break-before", "break-before", scaleBreak()),
		grp("break-inside", "break-inside", classGroup{"auto", "avoid", "avoid-page", "avoid-column"}),
		grp("box-decoration", "box-decoration", classGroup{"slice", "clone"}),
		grp("box", "box", classGroup{"border", "content"}),
		g("display", classGroup{
			"block", "inline-block", "inline", "flex", "inline-flex", "table",
			"inline-table", "table-caption", "table-cell", "table-column",
			"table-column-group", "table-footer-group", "table-header-group",
			"table-row-group", "table-row", "flow-root", "grid", "inline-grid",
			"contents", "list-item", "hidden",
		}),
		g("sr", classGroup{"sr-only", "not-sr-only"}),
		grp("float", "float", classGroup{"right", "left", "none", "start", "end"}),
		grp("clear", "clear", classGroup{"left", "right", "both", "none", "start", "end"}),
		g("isolation", classGroup{"isolate", "isolation-auto"}),
		grp("object-fit", "object", classGroup{"contain", "cover", "fill", "none", "scale-down"}),
		grp("object-position", "object", scalePositionWithArbitrary()),
		grp("overflow", "overflow", scaleOverflow()),
		grp("overflow-x", "overflow-x", scaleOverflow()),
		grp("overflow-y", "overflow-y", scaleOverflow()),
		grp("overscroll", "overscroll", scaleOverscroll()),
		grp("overscroll-x", "overscroll-x", scaleOverscroll()),
		grp("overscroll-y", "overscroll-y", scaleOverscroll()),
		g("position", classGroup{"static", "fixed", "absolute", "relative", "sticky"}),
		grp("inset", "inset", scaleInset()),
		grp("inset-x", "inset-x", scaleInset()),
		grp("inset-y", "inset-y", scaleInset()),
		g("start", classGroup{classObject{"inset-s": scaleInset(), "start": scaleInset()}}),
		g("end", classGroup{classObject{"inset-e": scaleInset(), "end": scaleInset()}}),
		grp("inset-bs", "inset-bs", scaleInset()),
		grp("inset-be", "inset-be", scaleInset()),
		grp("top", "top", scaleInset()),
		grp("right", "right", scaleInset()),
		grp("bottom", "bottom", scaleInset()),
		grp("left", "left", scaleInset()),
		g("visibility", classGroup{"visible", "invisible", "collapse"}),
		grp("z", "z", classGroup{isInteger, "auto", isArbitraryVariable, isArbitraryValue}),

		// --- Flexbox and Grid ---
		grp("basis", "basis", append(classGroup{isFraction, "full", "auto", themeContainer}, scaleUnambiguousSpacing()...)),
		grp("flex-direction", "flex", classGroup{"row", "row-reverse", "col", "col-reverse"}),
		grp("flex-wrap", "flex", classGroup{"nowrap", "wrap", "wrap-reverse"}),
		grp("flex", "flex", classGroup{isNumber, isFraction, "auto", "initial", "none", isArbitraryValue}),
		grp("grow", "grow", classGroup{"", isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("shrink", "shrink", classGroup{"", isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("order", "order", classGroup{isInteger, "first", "last", "none", isArbitraryVariable, isArbitraryValue}),
		grp("grid-cols", "grid-cols", scaleGridTemplateColsRows()),
		grp("col-start-end", "col", scaleGridColRowStartAndEnd()),
		grp("col-start", "col-start", scaleGridColRowStartOrEnd()),
		grp("col-end", "col-end", scaleGridColRowStartOrEnd()),
		grp("grid-rows", "grid-rows", scaleGridTemplateColsRows()),
		grp("row-start-end", "row", scaleGridColRowStartAndEnd()),
		grp("row-start", "row-start", scaleGridColRowStartOrEnd()),
		grp("row-end", "row-end", scaleGridColRowStartOrEnd()),
		grp("grid-flow", "grid-flow", classGroup{"row", "col", "dense", "row-dense", "col-dense"}),
		grp("auto-cols", "auto-cols", scaleGridAutoColsRows()),
		grp("auto-rows", "auto-rows", scaleGridAutoColsRows()),
		grp("gap", "gap", scaleUnambiguousSpacing()),
		grp("gap-x", "gap-x", scaleUnambiguousSpacing()),
		grp("gap-y", "gap-y", scaleUnambiguousSpacing()),
		grp("justify-content", "justify", append(scaleAlignPrimaryAxis(), "normal")),
		grp("justify-items", "justify-items", append(scaleAlignSecondaryAxis(), "normal")),
		grp("justify-self", "justify-self", append(classGroup{"auto"}, scaleAlignSecondaryAxis()...)),
		grp("align-content", "content", append(classGroup{"normal"}, scaleAlignPrimaryAxis()...)),
		grp("align-items", "items", append(scaleAlignSecondaryAxis(), classObject{"baseline": classGroup{"", "last"}})),
		grp("align-self", "self", append(append(classGroup{"auto"}, scaleAlignSecondaryAxis()...), classObject{"baseline": classGroup{"", "last"}})),
		grp("place-content", "place-content", scaleAlignPrimaryAxis()),
		grp("place-items", "place-items", append(scaleAlignSecondaryAxis(), "baseline")),
		grp("place-self", "place-self", append(classGroup{"auto"}, scaleAlignSecondaryAxis()...)),

		// --- Spacing ---
		grp("p", "p", scaleUnambiguousSpacing()),
		grp("px", "px", scaleUnambiguousSpacing()),
		grp("py", "py", scaleUnambiguousSpacing()),
		grp("ps", "ps", scaleUnambiguousSpacing()),
		grp("pe", "pe", scaleUnambiguousSpacing()),
		grp("pbs", "pbs", scaleUnambiguousSpacing()),
		grp("pbe", "pbe", scaleUnambiguousSpacing()),
		grp("pt", "pt", scaleUnambiguousSpacing()),
		grp("pr", "pr", scaleUnambiguousSpacing()),
		grp("pb", "pb", scaleUnambiguousSpacing()),
		grp("pl", "pl", scaleUnambiguousSpacing()),
		grp("m", "m", scaleMargin()),
		grp("mx", "mx", scaleMargin()),
		grp("my", "my", scaleMargin()),
		grp("ms", "ms", scaleMargin()),
		grp("me", "me", scaleMargin()),
		grp("mbs", "mbs", scaleMargin()),
		grp("mbe", "mbe", scaleMargin()),
		grp("mt", "mt", scaleMargin()),
		grp("mr", "mr", scaleMargin()),
		grp("mb", "mb", scaleMargin()),
		grp("ml", "ml", scaleMargin()),
		grp("space-x", "space-x", scaleUnambiguousSpacing()),
		g("space-x-reverse", classGroup{"space-x-reverse"}),
		grp("space-y", "space-y", scaleUnambiguousSpacing()),
		g("space-y-reverse", classGroup{"space-y-reverse"}),

		// --- Sizing ---
		grp("size", "size", scaleSizing()),
		grp("inline-size", "inline", append(classGroup{"auto"}, scaleSizingInline()...)),
		grp("min-inline-size", "min-inline", append(classGroup{"auto"}, scaleSizingInline()...)),
		grp("max-inline-size", "max-inline", append(classGroup{"none"}, scaleSizingInline()...)),
		grp("block-size", "block", append(classGroup{"auto"}, scaleSizingBlock()...)),
		grp("min-block-size", "min-block", append(classGroup{"auto"}, scaleSizingBlock()...)),
		grp("max-block-size", "max-block", append(classGroup{"none"}, scaleSizingBlock()...)),
		grp("w", "w", append(classGroup{themeContainer, "screen"}, scaleSizing()...)),
		grp("min-w", "min-w", append(classGroup{themeContainer, "screen", "none"}, scaleSizing()...)),
		grp("max-w", "max-w", append(classGroup{
			themeContainer, "screen", "none", "prose",
			classObject{"screen": classGroup{themeBreakpoint}},
		}, scaleSizing()...)),
		grp("h", "h", append(classGroup{"screen", "lh"}, scaleSizing()...)),
		grp("min-h", "min-h", append(classGroup{"screen", "lh", "none"}, scaleSizing()...)),
		grp("max-h", "max-h", append(classGroup{"screen", "lh"}, scaleSizing()...)),

		// --- Typography ---
		grp("font-size", "text", classGroup{"base", themeText, isArbitraryVariableLength, isArbitraryLength}),
		g("font-smoothing", classGroup{"antialiased", "subpixel-antialiased"}),
		g("font-style", classGroup{"italic", "not-italic"}),
		grp("font-weight", "font", classGroup{themeFontWeight, isArbitraryVariableWeight, isArbitraryWeight}),
		grp("font-stretch", "font-stretch", classGroup{
			"ultra-condensed", "extra-condensed", "condensed", "semi-condensed",
			"normal", "semi-expanded", "expanded", "extra-expanded", "ultra-expanded",
			isPercent, isArbitraryValue,
		}),
		grp("font-family", "font", classGroup{isArbitraryVariableFamilyName, isArbitraryFamilyName, themeFont}),
		grp("font-features", "font-features", classGroup{isArbitraryValue}),
		g("fvn-normal", classGroup{"normal-nums"}),
		g("fvn-ordinal", classGroup{"ordinal"}),
		g("fvn-slashed-zero", classGroup{"slashed-zero"}),
		g("fvn-figure", classGroup{"lining-nums", "oldstyle-nums"}),
		g("fvn-spacing", classGroup{"proportional-nums", "tabular-nums"}),
		g("fvn-fraction", classGroup{"diagonal-fractions", "stacked-fractions"}),
		grp("tracking", "tracking", classGroup{themeTracking, isArbitraryVariable, isArbitraryValue}),
		grp("line-clamp", "line-clamp", classGroup{isNumber, "none", isArbitraryVariable, isArbitraryNumber}),
		grp("leading", "leading", append(classGroup{themeLeading}, scaleUnambiguousSpacing()...)),
		grp("list-image", "list-image", classGroup{"none", isArbitraryVariable, isArbitraryValue}),
		grp("list-style-position", "list", classGroup{"inside", "outside"}),
		grp("list-style-type", "list", classGroup{"disc", "decimal", "none", isArbitraryVariable, isArbitraryValue}),
		grp("text-alignment", "text", classGroup{"left", "center", "right", "justify", "start", "end"}),
		grp("placeholder-color", "placeholder", scaleColor()),
		grp("text-color", "text", scaleColor()),
		g("text-decoration", classGroup{"underline", "overline", "line-through", "no-underline"}),
		grp("text-decoration-style", "decoration", append(scaleLineStyle(), "wavy")),
		grp("text-decoration-thickness", "decoration", classGroup{isNumber, "from-font", "auto", isArbitraryVariable, isArbitraryLength}),
		grp("text-decoration-color", "decoration", scaleColor()),
		grp("underline-offset", "underline-offset", classGroup{isNumber, "auto", isArbitraryVariable, isArbitraryValue}),
		g("text-transform", classGroup{"uppercase", "lowercase", "capitalize", "normal-case"}),
		g("text-overflow", classGroup{"truncate", "text-ellipsis", "text-clip"}),
		grp("text-wrap", "text", classGroup{"wrap", "nowrap", "balance", "pretty"}),
		grp("indent", "indent", scaleUnambiguousSpacing()),
		grp("tab-size", "tab", classGroup{isInteger, isArbitraryVariable, isArbitraryValue}),
		grp("vertical-align", "align", classGroup{
			"baseline", "top", "middle", "bottom", "text-top", "text-bottom",
			"sub", "super", isArbitraryVariable, isArbitraryValue,
		}),
		grp("whitespace", "whitespace", classGroup{"normal", "nowrap", "pre", "pre-line", "pre-wrap", "break-spaces"}),
		grp("break", "break", classGroup{"normal", "words", "all", "keep"}),
		grp("wrap", "wrap", classGroup{"break-word", "anywhere", "normal"}),
		grp("hyphens", "hyphens", classGroup{"none", "manual", "auto"}),
		grp("content", "content", classGroup{"none", isArbitraryVariable, isArbitraryValue}),

		// --- Backgrounds ---
		grp("bg-attachment", "bg", classGroup{"fixed", "local", "scroll"}),
		grp("bg-clip", "bg-clip", classGroup{"border", "padding", "content", "text"}),
		grp("bg-origin", "bg-origin", classGroup{"border", "padding", "content"}),
		grp("bg-position", "bg", scaleBgPosition()),
		grp("bg-repeat", "bg", scaleBgRepeat()),
		grp("bg-size", "bg", scaleBgSize()),
		grp("bg-image", "bg", classGroup{
			"none",
			classObject{
				"linear": classGroup{
					classObject{"to": classGroup{"t", "tr", "r", "br", "b", "bl", "l", "tl"}},
					isInteger, isArbitraryVariable, isArbitraryValue,
				},
				"radial": classGroup{"", isArbitraryVariable, isArbitraryValue},
				"conic":  classGroup{isInteger, isArbitraryVariable, isArbitraryValue},
			},
			isArbitraryVariableImage, isArbitraryImage,
		}),
		grp("bg-color", "bg", scaleColor()),
		grp("gradient-from-pos", "from", scaleGradientStopPosition()),
		grp("gradient-via-pos", "via", scaleGradientStopPosition()),
		grp("gradient-to-pos", "to", scaleGradientStopPosition()),
		grp("gradient-from", "from", scaleColor()),
		grp("gradient-via", "via", scaleColor()),
		grp("gradient-to", "to", scaleColor()),

		// --- Borders ---
		grp("rounded", "rounded", scaleRadius()),
		grp("rounded-s", "rounded-s", scaleRadius()),
		grp("rounded-e", "rounded-e", scaleRadius()),
		grp("rounded-t", "rounded-t", scaleRadius()),
		grp("rounded-r", "rounded-r", scaleRadius()),
		grp("rounded-b", "rounded-b", scaleRadius()),
		grp("rounded-l", "rounded-l", scaleRadius()),
		grp("rounded-ss", "rounded-ss", scaleRadius()),
		grp("rounded-se", "rounded-se", scaleRadius()),
		grp("rounded-ee", "rounded-ee", scaleRadius()),
		grp("rounded-es", "rounded-es", scaleRadius()),
		grp("rounded-tl", "rounded-tl", scaleRadius()),
		grp("rounded-tr", "rounded-tr", scaleRadius()),
		grp("rounded-br", "rounded-br", scaleRadius()),
		grp("rounded-bl", "rounded-bl", scaleRadius()),
		grp("border-w", "border", scaleBorderWidth()),
		grp("border-w-x", "border-x", scaleBorderWidth()),
		grp("border-w-y", "border-y", scaleBorderWidth()),
		grp("border-w-s", "border-s", scaleBorderWidth()),
		grp("border-w-e", "border-e", scaleBorderWidth()),
		grp("border-w-bs", "border-bs", scaleBorderWidth()),
		grp("border-w-be", "border-be", scaleBorderWidth()),
		grp("border-w-t", "border-t", scaleBorderWidth()),
		grp("border-w-r", "border-r", scaleBorderWidth()),
		grp("border-w-b", "border-b", scaleBorderWidth()),
		grp("border-w-l", "border-l", scaleBorderWidth()),
		grp("divide-x", "divide-x", scaleBorderWidth()),
		g("divide-x-reverse", classGroup{"divide-x-reverse"}),
		grp("divide-y", "divide-y", scaleBorderWidth()),
		g("divide-y-reverse", classGroup{"divide-y-reverse"}),
		grp("border-style", "border", append(scaleLineStyle(), "hidden", "none")),
		grp("divide-style", "divide", append(scaleLineStyle(), "hidden", "none")),
		grp("border-color", "border", scaleColor()),
		grp("border-color-x", "border-x", scaleColor()),
		grp("border-color-y", "border-y", scaleColor()),
		grp("border-color-s", "border-s", scaleColor()),
		grp("border-color-e", "border-e", scaleColor()),
		grp("border-color-bs", "border-bs", scaleColor()),
		grp("border-color-be", "border-be", scaleColor()),
		grp("border-color-t", "border-t", scaleColor()),
		grp("border-color-r", "border-r", scaleColor()),
		grp("border-color-b", "border-b", scaleColor()),
		grp("border-color-l", "border-l", scaleColor()),
		grp("divide-color", "divide", scaleColor()),
		grp("outline-style", "outline", append(scaleLineStyle(), "none", "hidden")),
		grp("outline-offset", "outline-offset", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("outline-w", "outline", classGroup{"", isNumber, isArbitraryVariableLength, isArbitraryLength}),
		grp("outline-color", "outline", scaleColor()),

		// --- Effects ---
		grp("shadow", "shadow", classGroup{"", "none", themeShadow, isArbitraryVariableShadow, isArbitraryShadow}),
		grp("shadow-color", "shadow", scaleColor()),
		grp("inset-shadow", "inset-shadow", classGroup{"none", themeInsetShadow, isArbitraryVariableShadow, isArbitraryShadow}),
		grp("inset-shadow-color", "inset-shadow", scaleColor()),
		grp("ring-w", "ring", scaleBorderWidth()),
		g("ring-w-inset", classGroup{"ring-inset"}),
		grp("ring-color", "ring", scaleColor()),
		grp("ring-offset-w", "ring-offset", classGroup{isNumber, isArbitraryLength}),
		grp("ring-offset-color", "ring-offset", scaleColor()),
		grp("inset-ring-w", "inset-ring", scaleBorderWidth()),
		grp("inset-ring-color", "inset-ring", scaleColor()),
		grp("text-shadow", "text-shadow", classGroup{"none", themeTextShadow, isArbitraryVariableShadow, isArbitraryShadow}),
		grp("text-shadow-color", "text-shadow", scaleColor()),
		grp("opacity", "opacity", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("mix-blend", "mix-blend", append(scaleBlendMode(), "plus-darker", "plus-lighter")),
		grp("bg-blend", "bg-blend", scaleBlendMode()),
		g("mask-clip", classGroup{
			classObject{"mask-clip": classGroup{"border", "padding", "content", "fill", "stroke", "view"}},
			"mask-no-clip",
		}),
		grp("mask-composite", "mask", classGroup{"add", "subtract", "intersect", "exclude"}),
		grp("mask-image-linear-pos", "mask-linear", classGroup{isNumber}),
		grp("mask-image-linear-from-pos", "mask-linear-from", scaleMaskImagePosition()),
		grp("mask-image-linear-to-pos", "mask-linear-to", scaleMaskImagePosition()),
		grp("mask-image-linear-from-color", "mask-linear-from", scaleColor()),
		grp("mask-image-linear-to-color", "mask-linear-to", scaleColor()),
		grp("mask-image-t-from-pos", "mask-t-from", scaleMaskImagePosition()),
		grp("mask-image-t-to-pos", "mask-t-to", scaleMaskImagePosition()),
		grp("mask-image-t-from-color", "mask-t-from", scaleColor()),
		grp("mask-image-t-to-color", "mask-t-to", scaleColor()),
		grp("mask-image-r-from-pos", "mask-r-from", scaleMaskImagePosition()),
		grp("mask-image-r-to-pos", "mask-r-to", scaleMaskImagePosition()),
		grp("mask-image-r-from-color", "mask-r-from", scaleColor()),
		grp("mask-image-r-to-color", "mask-r-to", scaleColor()),
		grp("mask-image-b-from-pos", "mask-b-from", scaleMaskImagePosition()),
		grp("mask-image-b-to-pos", "mask-b-to", scaleMaskImagePosition()),
		grp("mask-image-b-from-color", "mask-b-from", scaleColor()),
		grp("mask-image-b-to-color", "mask-b-to", scaleColor()),
		grp("mask-image-l-from-pos", "mask-l-from", scaleMaskImagePosition()),
		grp("mask-image-l-to-pos", "mask-l-to", scaleMaskImagePosition()),
		grp("mask-image-l-from-color", "mask-l-from", scaleColor()),
		grp("mask-image-l-to-color", "mask-l-to", scaleColor()),
		grp("mask-image-x-from-pos", "mask-x-from", scaleMaskImagePosition()),
		grp("mask-image-x-to-pos", "mask-x-to", scaleMaskImagePosition()),
		grp("mask-image-x-from-color", "mask-x-from", scaleColor()),
		grp("mask-image-x-to-color", "mask-x-to", scaleColor()),
		grp("mask-image-y-from-pos", "mask-y-from", scaleMaskImagePosition()),
		grp("mask-image-y-to-pos", "mask-y-to", scaleMaskImagePosition()),
		grp("mask-image-y-from-color", "mask-y-from", scaleColor()),
		grp("mask-image-y-to-color", "mask-y-to", scaleColor()),
		grp("mask-image-radial", "mask-radial", classGroup{isArbitraryVariable, isArbitraryValue}),
		grp("mask-image-radial-from-pos", "mask-radial-from", scaleMaskImagePosition()),
		grp("mask-image-radial-to-pos", "mask-radial-to", scaleMaskImagePosition()),
		grp("mask-image-radial-from-color", "mask-radial-from", scaleColor()),
		grp("mask-image-radial-to-color", "mask-radial-to", scaleColor()),
		grp("mask-image-radial-shape", "mask-radial", classGroup{"circle", "ellipse"}),
		grp("mask-image-radial-size", "mask-radial", classGroup{
			classObject{"closest": classGroup{"side", "corner"}, "farthest": classGroup{"side", "corner"}},
		}),
		grp("mask-image-radial-pos", "mask-radial-at", scalePosition()),
		grp("mask-image-conic-pos", "mask-conic", classGroup{isNumber}),
		grp("mask-image-conic-from-pos", "mask-conic-from", scaleMaskImagePosition()),
		grp("mask-image-conic-to-pos", "mask-conic-to", scaleMaskImagePosition()),
		grp("mask-image-conic-from-color", "mask-conic-from", scaleColor()),
		grp("mask-image-conic-to-color", "mask-conic-to", scaleColor()),
		grp("mask-mode", "mask", classGroup{"alpha", "luminance", "match"}),
		grp("mask-origin", "mask-origin", classGroup{"border", "padding", "content", "fill", "stroke", "view"}),
		grp("mask-position", "mask", scaleBgPosition()),
		grp("mask-repeat", "mask", scaleBgRepeat()),
		grp("mask-size", "mask", scaleBgSize()),
		grp("mask-type", "mask-type", classGroup{"alpha", "luminance"}),
		grp("mask-image", "mask", classGroup{"none", isArbitraryVariable, isArbitraryValue}),

		// --- Filters ---
		grp("filter", "filter", classGroup{"", "none", isArbitraryVariable, isArbitraryValue}),
		grp("blur", "blur", scaleBlur()),
		grp("brightness", "brightness", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("contrast", "contrast", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("drop-shadow", "drop-shadow", classGroup{"", "none", themeDropShadow, isArbitraryVariableShadow, isArbitraryShadow}),
		grp("drop-shadow-color", "drop-shadow", scaleColor()),
		grp("grayscale", "grayscale", classGroup{"", isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("hue-rotate", "hue-rotate", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("invert", "invert", classGroup{"", isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("saturate", "saturate", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("sepia", "sepia", classGroup{"", isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("backdrop-filter", "backdrop-filter", classGroup{"", "none", isArbitraryVariable, isArbitraryValue}),
		grp("backdrop-blur", "backdrop-blur", scaleBlur()),
		grp("backdrop-brightness", "backdrop-brightness", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("backdrop-contrast", "backdrop-contrast", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("backdrop-grayscale", "backdrop-grayscale", classGroup{"", isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("backdrop-hue-rotate", "backdrop-hue-rotate", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("backdrop-invert", "backdrop-invert", classGroup{"", isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("backdrop-opacity", "backdrop-opacity", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("backdrop-saturate", "backdrop-saturate", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("backdrop-sepia", "backdrop-sepia", classGroup{"", isNumber, isArbitraryVariable, isArbitraryValue}),

		// --- Tables ---
		grp("border-collapse", "border", classGroup{"collapse", "separate"}),
		grp("border-spacing", "border-spacing", scaleUnambiguousSpacing()),
		grp("border-spacing-x", "border-spacing-x", scaleUnambiguousSpacing()),
		grp("border-spacing-y", "border-spacing-y", scaleUnambiguousSpacing()),
		grp("table-layout", "table", classGroup{"auto", "fixed"}),
		grp("caption", "caption", classGroup{"top", "bottom"}),

		// --- Transitions and Animation ---
		grp("transition", "transition", classGroup{
			"", "all", "colors", "opacity", "shadow", "transform", "none",
			isArbitraryVariable, isArbitraryValue,
		}),
		grp("transition-behavior", "transition", classGroup{"normal", "discrete"}),
		grp("duration", "duration", classGroup{isNumber, "initial", isArbitraryVariable, isArbitraryValue}),
		grp("ease", "ease", classGroup{"linear", "initial", themeEase, isArbitraryVariable, isArbitraryValue}),
		grp("delay", "delay", classGroup{isNumber, isArbitraryVariable, isArbitraryValue}),
		grp("animate", "animate", classGroup{"none", themeAnimate, isArbitraryVariable, isArbitraryValue}),

		// --- Transforms ---
		grp("backface", "backface", classGroup{"hidden", "visible"}),
		grp("perspective", "perspective", classGroup{themePerspective, isArbitraryVariable, isArbitraryValue}),
		grp("perspective-origin", "perspective-origin", scalePositionWithArbitrary()),
		grp("rotate", "rotate", scaleRotate()),
		grp("rotate-x", "rotate-x", scaleRotate()),
		grp("rotate-y", "rotate-y", scaleRotate()),
		grp("rotate-z", "rotate-z", scaleRotate()),
		grp("scale", "scale", scaleScale()),
		grp("scale-x", "scale-x", scaleScale()),
		grp("scale-y", "scale-y", scaleScale()),
		grp("scale-z", "scale-z", scaleScale()),
		g("scale-3d", classGroup{"scale-3d"}),
		grp("skew", "skew", scaleSkew()),
		grp("skew-x", "skew-x", scaleSkew()),
		grp("skew-y", "skew-y", scaleSkew()),
		grp("transform", "transform", classGroup{isArbitraryVariable, isArbitraryValue, "", "none", "gpu", "cpu"}),
		grp("transform-origin", "origin", scalePositionWithArbitrary()),
		grp("transform-style", "transform", classGroup{"3d", "flat"}),
		grp("translate", "translate", scaleTranslate()),
		grp("translate-x", "translate-x", scaleTranslate()),
		grp("translate-y", "translate-y", scaleTranslate()),
		grp("translate-z", "translate-z", scaleTranslate()),
		g("translate-none", classGroup{"translate-none"}),
		grp("zoom", "zoom", classGroup{isInteger, isArbitraryVariable, isArbitraryValue}),

		// --- Interactivity ---
		grp("accent", "accent", scaleColor()),
		grp("appearance", "appearance", classGroup{"none", "auto"}),
		grp("caret-color", "caret", scaleColor()),
		grp("color-scheme", "scheme", classGroup{"normal", "dark", "light", "light-dark", "only-dark", "only-light"}),
		grp("cursor", "cursor", classGroup{
			"auto", "default", "pointer", "wait", "text", "move", "help", "not-allowed",
			"none", "context-menu", "progress", "cell", "crosshair", "vertical-text",
			"alias", "copy", "no-drop", "grab", "grabbing", "all-scroll", "col-resize",
			"row-resize", "n-resize", "e-resize", "s-resize", "w-resize", "ne-resize",
			"nw-resize", "se-resize", "sw-resize", "ew-resize", "ns-resize", "nesw-resize",
			"nwse-resize", "zoom-in", "zoom-out", isArbitraryVariable, isArbitraryValue,
		}),
		grp("field-sizing", "field-sizing", classGroup{"fixed", "content"}),
		grp("pointer-events", "pointer-events", classGroup{"auto", "none"}),
		grp("resize", "resize", classGroup{"none", "", "y", "x"}),
		grp("scroll-behavior", "scroll", classGroup{"auto", "smooth"}),
		grp("scrollbar-thumb-color", "scrollbar-thumb", scaleColor()),
		grp("scrollbar-track-color", "scrollbar-track", scaleColor()),
		grp("scrollbar-gutter", "scrollbar-gutter", classGroup{"auto", "stable", "both"}),
		grp("scrollbar-w", "scrollbar", classGroup{"auto", "thin", "none"}),
		grp("scroll-m", "scroll-m", scaleUnambiguousSpacing()),
		grp("scroll-mx", "scroll-mx", scaleUnambiguousSpacing()),
		grp("scroll-my", "scroll-my", scaleUnambiguousSpacing()),
		grp("scroll-ms", "scroll-ms", scaleUnambiguousSpacing()),
		grp("scroll-me", "scroll-me", scaleUnambiguousSpacing()),
		grp("scroll-mbs", "scroll-mbs", scaleUnambiguousSpacing()),
		grp("scroll-mbe", "scroll-mbe", scaleUnambiguousSpacing()),
		grp("scroll-mt", "scroll-mt", scaleUnambiguousSpacing()),
		grp("scroll-mr", "scroll-mr", scaleUnambiguousSpacing()),
		grp("scroll-mb", "scroll-mb", scaleUnambiguousSpacing()),
		grp("scroll-ml", "scroll-ml", scaleUnambiguousSpacing()),
		grp("scroll-p", "scroll-p", scaleUnambiguousSpacing()),
		grp("scroll-px", "scroll-px", scaleUnambiguousSpacing()),
		grp("scroll-py", "scroll-py", scaleUnambiguousSpacing()),
		grp("scroll-ps", "scroll-ps", scaleUnambiguousSpacing()),
		grp("scroll-pe", "scroll-pe", scaleUnambiguousSpacing()),
		grp("scroll-pbs", "scroll-pbs", scaleUnambiguousSpacing()),
		grp("scroll-pbe", "scroll-pbe", scaleUnambiguousSpacing()),
		grp("scroll-pt", "scroll-pt", scaleUnambiguousSpacing()),
		grp("scroll-pr", "scroll-pr", scaleUnambiguousSpacing()),
		grp("scroll-pb", "scroll-pb", scaleUnambiguousSpacing()),
		grp("scroll-pl", "scroll-pl", scaleUnambiguousSpacing()),
		grp("snap-align", "snap", classGroup{"start", "end", "center", "align-none"}),
		grp("snap-stop", "snap", classGroup{"normal", "always"}),
		grp("snap-type", "snap", classGroup{"none", "x", "y", "both"}),
		grp("snap-strictness", "snap", classGroup{"mandatory", "proximity"}),
		grp("touch", "touch", classGroup{"auto", "none", "manipulation"}),
		grp("touch-x", "touch-pan", classGroup{"x", "left", "right"}),
		grp("touch-y", "touch-pan", classGroup{"y", "up", "down"}),
		g("touch-pz", classGroup{"touch-pinch-zoom"}),
		grp("select", "select", classGroup{"none", "text", "all", "auto"}),
		grp("will-change", "will-change", classGroup{"auto", "scroll", "contents", "transform", isArbitraryVariable, isArbitraryValue}),

		// --- SVG ---
		grp("fill", "fill", append(classGroup{"none"}, scaleColor()...)),
		grp("stroke-w", "stroke", classGroup{isNumber, isArbitraryVariableLength, isArbitraryLength, isArbitraryNumber}),
		grp("stroke", "stroke", append(classGroup{"none"}, scaleColor()...)),

		// --- Accessibility ---
		grp("forced-color-adjust", "forced-color-adjust", classGroup{"auto", "none"}),
	}

	return &twConfig{
		theme:       theme,
		classGroups: classGroups,
		conflictingClassGroups: map[string][]string{
			"container-named":  {"container-type"},
			"overflow":         {"overflow-x", "overflow-y"},
			"overscroll":       {"overscroll-x", "overscroll-y"},
			"inset":            {"inset-x", "inset-y", "inset-bs", "inset-be", "start", "end", "top", "right", "bottom", "left"},
			"inset-x":          {"right", "left"},
			"inset-y":          {"top", "bottom"},
			"flex":             {"basis", "grow", "shrink"},
			"gap":              {"gap-x", "gap-y"},
			"p":                {"px", "py", "ps", "pe", "pbs", "pbe", "pt", "pr", "pb", "pl"},
			"px":               {"pr", "pl"},
			"py":               {"pt", "pb"},
			"m":                {"mx", "my", "ms", "me", "mbs", "mbe", "mt", "mr", "mb", "ml"},
			"mx":               {"mr", "ml"},
			"my":               {"mt", "mb"},
			"size":             {"w", "h"},
			"font-size":        {"leading"},
			"fvn-normal":       {"fvn-ordinal", "fvn-slashed-zero", "fvn-figure", "fvn-spacing", "fvn-fraction"},
			"fvn-ordinal":      {"fvn-normal"},
			"fvn-slashed-zero": {"fvn-normal"},
			"fvn-figure":       {"fvn-normal"},
			"fvn-spacing":      {"fvn-normal"},
			"fvn-fraction":     {"fvn-normal"},
			"line-clamp":       {"display", "overflow"},
			"rounded": {
				"rounded-s", "rounded-e", "rounded-t", "rounded-r", "rounded-b", "rounded-l",
				"rounded-ss", "rounded-se", "rounded-ee", "rounded-es",
				"rounded-tl", "rounded-tr", "rounded-br", "rounded-bl",
			},
			"rounded-s":      {"rounded-ss", "rounded-es"},
			"rounded-e":      {"rounded-se", "rounded-ee"},
			"rounded-t":      {"rounded-tl", "rounded-tr"},
			"rounded-r":      {"rounded-tr", "rounded-br"},
			"rounded-b":      {"rounded-br", "rounded-bl"},
			"rounded-l":      {"rounded-tl", "rounded-bl"},
			"border-spacing": {"border-spacing-x", "border-spacing-y"},
			"border-w": {
				"border-w-x", "border-w-y", "border-w-s", "border-w-e", "border-w-bs", "border-w-be",
				"border-w-t", "border-w-r", "border-w-b", "border-w-l",
			},
			"border-w-x": {"border-w-r", "border-w-l"},
			"border-w-y": {"border-w-t", "border-w-b"},
			"border-color": {
				"border-color-x", "border-color-y", "border-color-s", "border-color-e",
				"border-color-bs", "border-color-be", "border-color-t", "border-color-r",
				"border-color-b", "border-color-l",
			},
			"border-color-x": {"border-color-r", "border-color-l"},
			"border-color-y": {"border-color-t", "border-color-b"},
			"translate":      {"translate-x", "translate-y", "translate-none"},
			"translate-none": {"translate", "translate-x", "translate-y", "translate-z"},
			"scroll-m": {
				"scroll-mx", "scroll-my", "scroll-ms", "scroll-me", "scroll-mbs", "scroll-mbe",
				"scroll-mt", "scroll-mr", "scroll-mb", "scroll-ml",
			},
			"scroll-mx": {"scroll-mr", "scroll-ml"},
			"scroll-my": {"scroll-mt", "scroll-mb"},
			"scroll-p": {
				"scroll-px", "scroll-py", "scroll-ps", "scroll-pe", "scroll-pbs", "scroll-pbe",
				"scroll-pt", "scroll-pr", "scroll-pb", "scroll-pl",
			},
			"scroll-px": {"scroll-pr", "scroll-pl"},
			"scroll-py": {"scroll-pt", "scroll-pb"},
			"touch":     {"touch-x", "touch-y", "touch-pz"},
			"touch-x":   {"touch"},
			"touch-y":   {"touch"},
			"touch-pz":  {"touch"},
		},
		conflictingClassGroupModifiers: map[string][]string{
			"font-size": {"leading"},
		},
		postfixLookupClassGroups: []string{"container-type"},
		orderSensitiveModifiers: []string{
			"*", "**", "after", "backdrop", "before", "details-content",
			"file", "first-letter", "first-line", "marker", "placeholder", "selection",
		},
	}
}
