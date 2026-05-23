package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// scrollAreaClasses is the base class set for ScrollArea. shadcn's component
// nests a Radix Root, Viewport, ScrollBar and Corner; this port flattens them
// to a single overflow <div>, so the class string combines Root's relative,
// Viewport's outline/ring focus handling, and a small set of arbitrary
// utilities that style the native scrollbar (the native equivalent of Radix's
// rendered ScrollBar element).
const scrollAreaClasses = "relative overflow-auto focus-visible:ring-ring/50 rounded-[inherit] transition-[color,box-shadow] outline-none focus-visible:ring-[3px] focus-visible:outline-1 [scrollbar-width:thin] [scrollbar-color:var(--color-border)_transparent] [&::-webkit-scrollbar]:size-2.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-border"

// ScrollArea renders a shadcn/ui scroll area as a single <div> with overflow:
// auto and CSS-styled scrollbars. shadcn's Root/Viewport/ScrollBar/Corner
// structure collapses to one element here — a deliberate divergence, mirroring
// how AlertDialogOverlay is not ported (the native equivalent is a pseudo, not
// an element).
//
// shadcn's ScrollBar component is intentionally NOT exported by this package:
// in the native port the scrollbar is the ::-webkit-scrollbar pseudo-element
// (and Firefox's scrollbar-width / scrollbar-color), styled via the utility
// classes above. For finer control, override scrollbar styles in your own CSS.
func ScrollArea(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "scroll-area", scrollAreaClasses)
}
