package shadcn

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Kbd renders a keyboard key as a styled <kbd>. Class strings are
// reconstructed from shadcn/ui's post-July-2026 registry (kbd.tsx skeleton +
// the style-vega.css .cn-kbd rule) — see "Upstream restructure (July 2026)"
// in TODOS.md.
func Kbd(attribsChildren ...any) *mx.Element {
	return finish(html.Kbd(attribsChildren...), "kbd",
		"pointer-events-none inline-flex items-center justify-center select-none bg-muted text-muted-foreground in-data-[slot=tooltip-content]:bg-background/20 in-data-[slot=tooltip-content]:text-background dark:in-data-[slot=tooltip-content]:bg-background/10 h-5 w-fit min-w-5 gap-1 rounded-sm px-1 font-sans text-xs font-medium [&_svg:not([class*='size-'])]:size-3")
}

// KbdGroup groups several [Kbd] keys (e.g. a shortcut like ⌘ K). Like
// upstream it renders a <kbd> element itself, not a <div>.
func KbdGroup(attribsChildren ...any) *mx.Element {
	return finish(html.Kbd(attribsChildren...), "kbd-group",
		"inline-flex items-center gap-1")
}
