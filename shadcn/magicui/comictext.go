package magicui

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

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
	const dotColor = "#EF4444"
	const backgroundColor = "#FACC15"
	fontSize := or(opts.FontSize, 5)
	// No caller value reaches these declarations except the number fontSize,
	// so they need no checkCSSValue (which would reject the quoted font
	// names).
	style := "font-size:" + num(fontSize) + "rem" +
		";font-family:'Bangers', 'Comic Sans MS', 'Impact', sans-serif" +
		";font-weight:900" +
		";-webkit-text-stroke:" + px(fontSize*0.35) + " #000000" +
		";transform:skewX(-10deg)" +
		";text-transform:uppercase" +
		";filter:drop-shadow(5px 5px 0px #000000) drop-shadow(3px 3px 0px " + dotColor + ")" +
		";background-color:" + backgroundColor +
		";background-image:radial-gradient(circle at 1px 1px, " + dotColor + " 1px, transparent 0)" +
		";background-size:8px 8px" +
		";background-clip:text" +
		";-webkit-background-clip:text" +
		";-webkit-text-fill-color:transparent"
	return finish(html.Div(attribsChildren...), "comic-text", "text-center select-none animate-comic-text-pop", style)
}
