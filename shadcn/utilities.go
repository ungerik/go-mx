package shadcn

import _ "embed"

// UtilitiesCSS holds the @utility rules of shadcn/ui's own stylesheet that
// its components use beyond Tailwind: the scroll-fade-* edge masks, which
// fade a scroll container's edges while there is more to scroll, and the
// shimmer text effect of a running or uploading item. They need @property
// registrations, @keyframes and scroll-driven animations, which no
// arbitrary-value class can express, so this package emits Tailwind classes
// only and substitutes them by default (see [UseUtilities]).
//
// To use them, add the CSS to the app's Tailwind input, after
// @import "tailwindcss", and set [UseUtilities].
//
//go:embed utilities.css
var UtilitiesCSS string

// UseUtilities makes the components emit upstream's classes from
// [UtilitiesCSS] instead of their Tailwind substitutes: a running
// [TraceLine]'s title and an uploading or processing [Attachment]'s title
// shimmer instead of pulsing with animate-pulse, and [AttachmentGroup],
// [CodeBlockContent] and [MessageScrollerViewport] fade their scrolled
// edges as upstream's do. Set it only together with adding UtilitiesCSS to
// the app's Tailwind input, where the classes are defined.
//
// Assign it during initialization, before the first component is built: it
// is read when a component is constructed, like [DefaultLabels].
var UseUtilities = false

// utilityClasses returns upstream, the classes of [UtilitiesCSS], when
// [UseUtilities] is set, else substitute.
func utilityClasses(upstream, substitute string) string {
	if UseUtilities {
		return upstream
	}
	return substitute
}
