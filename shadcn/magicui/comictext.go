package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// comicTextClasses are the lettering of a [ComicText]: red halftone dots
// (#EF4444) on yellow (#FACC15), a black outline of 0.35px per rem of font
// size, and a hard black and red drop shadow. One literal, so Tailwind finds
// every class in the source.
const comicTextClasses = "[--comic-text-font-size:5] text-[length:calc(var(--comic-text-font-size)*1rem)] " +
	"[font-family:Bangers,Comic_Sans_MS,Impact,sans-serif] font-black uppercase [transform:skewX(-10deg)] " +
	"[-webkit-text-stroke:calc(var(--comic-text-font-size)*0.35px)_#000000] " +
	"[filter:drop-shadow(5px_5px_0px_#000000)_drop-shadow(3px_3px_0px_#EF4444)] " +
	"bg-[#FACC15] bg-[radial-gradient(circle_at_1px_1px,#EF4444_1px,transparent_0)] bg-size-[8px_8px] bg-clip-text [-webkit-text-fill-color:transparent] " +
	"text-center select-none animate-comic-text-pop"

// ComicTextOptions configures a [ComicText]. The zero value is upstream's
// default size of 5rem.
type ComicTextOptions struct {
	// FontSize is the font size in rem, which also scales the outline;
	// 0 means 5.
	FontSize float64
}

// ComicText renders its children as comic-book lettering: skewed uppercase
// letters in a heavy black outline, filled with red halftone dots on yellow,
// with a hard drop shadow. The font stack starts with Bangers, which the app
// has to load itself for the intended look, and falls back to Comic Sans MS
// and Impact. It pops in once on render, scaling up from 0.8 with a slight
// rotation.
//
// Upstream accepts only a string child and animates the entrance with a
// motion spring; this port takes any children and uses a CSS animation
// with upstream's back-out cubic-bezier instead of the spring.
//
// Needs [CSS] in the app's Tailwind input.
func ComicText(opts ComicTextOptions, attribsChildren ...any) *mx.Element {
	// The lettering is comicTextClasses, with the size in
	// --comic-text-font-size, so a ComicText at its default size needs no
	// style attribute (see "Strict Content-Security-Policy" in the README).
	var style string
	if opts.FontSize != 0 {
		style = "--comic-text-font-size:" + num(opts.FontSize)
	}
	return finish(html.Div(attribsChildren...), "comic-text", comicTextClasses, style)
}
