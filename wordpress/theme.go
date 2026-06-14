package wordpress

import (
	_ "embed"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// themeTokensCSS is shadcn/ui's new-york-v4 globals.css (the OKLCH color tokens
// and @theme mapping), compiled at runtime by the @tailwindcss/browser CDN
// build. Copied from cmd/shadcn-gallery/theme.css (shadcn/ui is MIT; see the
// repo's THIRD-PARTY-LICENSES.md).
//
//go:embed theme.css
var themeTokensCSS string

// tailwindCDN is the @tailwindcss/browser build that compiles themeTokensCSS and
// the shadcn utility classes against the live DOM, so the static output needs no
// node build step. It does require an internet connection when the page is
// viewed; the .wp-content prose layer below is plain CSS so the article body
// stays readable even when the CDN is unavailable.
const tailwindCDN = "https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"

// wpContentCSS styles the raw WordPress body HTML the content pipeline emits.
// Tailwind v4 Preflight resets every element's margins, list markers and
// heading sizes to nothing, so without this layer an imported article renders as
// run-together text. This is plain CSS (no Tailwind), scoped to .wp-content, and
// uses the shadcn theme tokens so it matches the rest of the page.
const wpContentCSS = /*css*/ `
.wp-content { line-height: 1.75; color: var(--foreground, #0a0a0a); overflow-wrap: break-word; }
.wp-content > * + * { margin-top: 1.25em; }
.wp-content h1, .wp-content h2, .wp-content h3,
.wp-content h4, .wp-content h5, .wp-content h6 {
  font-weight: 600; line-height: 1.25; margin-top: 2em; margin-bottom: 0.75em; }
.wp-content h2 { font-size: 1.5rem; }
.wp-content h3 { font-size: 1.25rem; }
.wp-content h4 { font-size: 1.125rem; }
.wp-content h5, .wp-content h6 { font-size: 1rem; }
.wp-content p { margin: 0; }
.wp-content a { color: var(--primary, #2563eb); text-decoration: underline; text-underline-offset: 2px; }
.wp-content ul, .wp-content ol { padding-left: 1.625em; margin: 0; }
.wp-content ul { list-style: disc; }
.wp-content ol { list-style: decimal; }
.wp-content li { margin-top: 0.375em; }
.wp-content li > ul, .wp-content li > ol { margin-top: 0.375em; }
.wp-content blockquote {
  border-left: 3px solid var(--border, #e5e5e5); padding-left: 1em;
  font-style: italic; color: var(--muted-foreground, #737373); }
.wp-content pre {
  background: var(--muted, #f5f5f5); padding: 1em; border-radius: 0.5rem;
  overflow-x: auto; font-size: 0.875em; line-height: 1.5; }
.wp-content code {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 0.875em; }
.wp-content :not(pre) > code {
  background: var(--muted, #f5f5f5); padding: 0.15em 0.35em; border-radius: 0.3rem; }
.wp-content img { max-width: 100%; height: auto; border-radius: 0.5rem; }
.wp-content figure { margin: 0; }
.wp-content figcaption {
  font-size: 0.875em; color: var(--muted-foreground, #737373); margin-top: 0.5em; text-align: center; }
.wp-content table { width: 100%; border-collapse: collapse; display: block; overflow-x: auto; font-size: 0.95em; }
.wp-content th, .wp-content td {
  border: 1px solid var(--border, #e5e5e5); padding: 0.5em 0.75em; text-align: left; }
.wp-content thead th { background: var(--muted, #f5f5f5); font-weight: 600; }
.wp-content hr { border: 0; border-top: 1px solid var(--border, #e5e5e5); margin: 2em 0; }
`

// HeadComponents returns the <head> contents the rendered pages need: the
// shadcn theme tokens, the Tailwind v4 CDN build that compiles them, and the
// .wp-content prose layer. Drop it into html.Document.HeadCustom when embedding
// views in your own page; [SiteDocument] installs it for you.
func HeadComponents() mx.Component {
	return mx.Components{
		html.Element("style", html.Type("text/tailwindcss"), html.Raw(themeTokensCSS)),
		html.Script(html.Src(tailwindCDN)),
		html.Element("style", html.Raw(wpContentCSS)),
	}
}

// ThemeCSS returns the plain (non-Tailwind) .wp-content prose CSS, for a caller
// embedding views in a page that already runs Tailwind and only needs the body
// typography layer.
func ThemeCSS() string { return wpContentCSS }
