// Package shadcntheme is the page head the go-mx demo sites share: the
// shadcn/ui new-york-v4 theme tokens compiled in the browser by the
// @tailwindcss/browser CDN build, and the light/dark switch that carries
// across page loads. The component gallery (cmd/shadcn-gallery) and the
// documentation site (tools/gen-docs-site) both build their <head> with
// [Head], so the two look the same without either forking the theme.
//
// It is internal because it is demo plumbing, not library API: an
// application ships compiled Tailwind CSS instead of the CDN build. The
// theme file itself, theme.css, is the one the shadcn tutorial tells readers
// to copy into their own project.
package shadcntheme

import (
	_ "embed"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// CSS is the shadcn/ui new-york-v4 globals.css, the Tailwind v4 input that
// [Head] puts into a <style type="text/tailwindcss"> block. See theme.css for
// the full note.
//
//go:embed theme.css
var CSS string

// TailwindBrowserURL is the @tailwindcss/browser 4.3.5 build that compiles
// the text/tailwindcss style block and the classes it finds in the live DOM
// at runtime, so the demo sites need no node toolchain. It is pinned to an
// exact version and loaded with [TailwindBrowserIntegrity], like htmx in the
// hx package, so neither a new release nor a changed file on the CDN alters
// the demo pages without a commit.
const TailwindBrowserURL = "https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4.3.5/dist/index.global.js"

// TailwindBrowserIntegrity is the Subresource Integrity hash of the file at
// [TailwindBrowserURL]; change the two together and run
//
//	MX_CHECK_CDN=1 go test ./internal/shadcntheme
//
// which checks the hash against the CDN.
const TailwindBrowserIntegrity = "sha384-t+qLIxS5zM/KF99/doWY+dxTsavdvgayaZt8A9ntLsUjEkiXzYl5M3Ij8zVFPL9j"

// InitScript applies the light/dark choice before the body renders, so the
// preference carries across full-page navigations with no flash of the wrong
// mode. It runs synchronously during head parsing. The choice is "light",
// "dark" or, when none is saved, the system's: then the page follows the
// operating system's prefers-color-scheme, also when it changes while the
// page is open. A ?theme=dark, ?theme=light or ?theme=system query param wins
// and is persisted (so it carries forward like using the toggle). It defines
// mxSetTheme(choice), which [ModeToggle] calls and [ToggleScript] uses
// indirectly through the saved value.
const InitScript = /*js*/ `try{var p=new URLSearchParams(location.search).get('theme');if(p==='dark'||p==='light')localStorage.theme=p;if(p==='system')localStorage.removeItem('theme');var m=matchMedia('(prefers-color-scheme: dark)');window.mxApplyTheme=function(){var t=localStorage.theme;document.documentElement.classList.toggle('dark',t==='dark'||(t!=='light'&&m.matches))};window.mxSetTheme=function(t){if(t==='system')localStorage.removeItem('theme');else localStorage.theme=t;mxApplyTheme()};m.addEventListener('change',mxApplyTheme);mxApplyTheme()}catch(e){}`

// ToggleScript is the onclick handler of a theme toggle button: it flips the
// dark class on the root element and saves the choice for [InitScript].
const ToggleScript = /*js*/ `localStorage.theme=document.documentElement.classList.toggle('dark')?'dark':'light'`

// Head returns the head elements of a demo page: [InitScript], a
// text/tailwindcss style block holding [CSS] followed by each of extraCSS
// (for example magicui.CSS, or page-specific @apply rules), and the
// [TailwindBrowserURL] script that compiles it, checked against
// [TailwindBrowserIntegrity].
func Head(extraCSS ...string) mx.Components {
	return mx.Components{
		html.ScriptJS(InitScript),
		html.Element("style", html.Type("text/tailwindcss"), html.Raw(TailwindInput(extraCSS...))),
		html.Script(
			html.Src(TailwindBrowserURL),
			html.Integrity(TailwindBrowserIntegrity),
			html.CrossOriginAnonymous,
		),
	}
}

// TailwindInput is the Tailwind CSS input [Head] embeds: [CSS] followed by
// each of extraCSS, separated by newlines.
func TailwindInput(extraCSS ...string) string {
	return strings.Join(append([]string{CSS}, extraCSS...), "\n")
}
