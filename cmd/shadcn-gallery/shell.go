package main

import (
	"context"
	_ "embed"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// themeCSS is the shadcn/ui new-york-v4 globals.css, injected into a
// <style type="text/tailwindcss"> block that the @tailwindcss/browser CDN build
// compiles at runtime. See theme.css for the full note.
//
//go:embed theme.css
var themeCSS string

// page builds the full HTML document for one gallery route: the shared <head>
// (Tailwind CDN + theme), then the sidebar-and-main layout around content.
func page(reg *Registry, currentSlug, title string, content mx.Component) *html.Document {
	doc := html.NewDocument("shadcn · go-mx — "+title,
		layout(reg, currentSlug, content),
	)
	doc.Meta = map[string]string{"viewport": "width=device-width, initial-scale=1"}
	doc.HeadCustom = head()
	return doc
}

// head wires up the two CDN dependencies, no build step:
//
//   - Tailwind v4 via @tailwindcss/browser: the theme tokens go in a
//     <style type="text/tailwindcss"> block, and the script compiles them plus
//     the classes it finds in the live DOM.
//   - Shiki for the Code tab (see shikiScript): a deferred ESM module that
//     highlights every <pre><code class="language-go"> with TextMate grammars —
//     so call sites, types and properties are colored, not just keywords and
//     strings the way a regex highlighter manages.
func head() mx.Component {
	return mx.Components{
		html.Script(mx.Raw(themeInitScript)),
		html.Element("style", html.Type("text/tailwindcss"), html.Raw(themeCSS)),
		html.Script(html.Src("https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4")),
		html.Script(html.Type("module"), mx.Raw(shikiScript)),
	}
}

// themeInitScript applies the light/dark choice before the body renders, so the
// preference carries across full-page navigations with no flash of the wrong
// mode. It runs synchronously during head parsing. Precedence: a ?theme=dark or
// ?theme=light query param wins and is persisted (so it carries forward like
// using the toggle); otherwise the value saved by the sidebar Theme toggle (see
// layout) is used.
const themeInitScript = /*js*/ `try{var p=new URLSearchParams(location.search).get('theme');if(p==='dark'||p==='light')localStorage.theme=p;if(localStorage.theme==='dark')document.documentElement.classList.add('dark')}catch(e){}`

// shikiScript highlights every Go code block in the browser with Shiki
// (https://shiki.style) loaded from a pinned ESM CDN, using the one-dark-pro
// theme. Module scripts are deferred, so the DOM is parsed when it runs. For
// each <pre><code class="language-go"> it asks Shiki for its own <pre> (which
// carries the theme's inline background and token colors), re-applies the
// layout classes, and swaps it in — covering Code tabs that start hidden.
const shikiScript = /*js*/ `
import { codeToHtml } from 'https://esm.sh/shiki@1';
(async function () {
  for (const code of document.querySelectorAll('pre > code.language-go')) {
    // theme id — browse all bundled themes at https://shiki.style/themes
    const out = await codeToHtml(code.textContent, { lang: 'go', theme: 'monokai' });
    const tpl = document.createElement('template');
    tpl.innerHTML = out;
    const pre = tpl.content.firstElementChild;
    pre.className += ' mt-2 overflow-x-auto rounded-md p-4 text-sm leading-relaxed';
    code.closest('pre').replaceWith(pre);
  }
})();
`

// layout is the two-column shell: a sticky sidebar listing every component and
// a main column holding the current page's content.
func layout(reg *Registry, currentSlug string, content mx.Component) mx.Component {
	return html.Div(html.Class("flex min-h-screen"),
		html.Aside(html.Class("w-64 shrink-0 border-r bg-sidebar text-sidebar-foreground"),
			html.Div(html.Class("sticky top-0 flex h-screen flex-col"),
				html.Div(html.Class("flex items-center justify-between border-b px-4 py-3"),
					html.A(html.HRef("/"), html.Class("text-sm font-semibold"), "shadcn · go-mx"),
					shadcn.Button(shadcn.ButtonOutline, shadcn.SizeSM,
						html.Attrib("onclick", "localStorage.theme=document.documentElement.classList.toggle('dark')?'dark':'light'"),
						"Theme"),
				),
				html.Nav(html.Class("flex-1 space-y-0.5 overflow-y-auto p-3"),
					sidebarLinks(reg, currentSlug),
				),
			),
		),
		html.Main(html.Class("min-w-0 flex-1"),
			html.Div(html.Class("mx-auto max-w-3xl px-8 py-10"),
				content,
			),
		),
	)
}

// sidebarLinks renders one nav link per component, highlighting the active one.
func sidebarLinks(reg *Registry, currentSlug string) mx.Component {
	return mx.ForEach(reg.Docs, func(d ComponentDoc) mx.Component {
		cls := "block rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground"
		if d.Slug == currentSlug {
			cls = "block rounded-md px-3 py-1.5 text-sm font-medium bg-accent text-accent-foreground"
		}
		return html.A(html.HRef("/components/"+d.Slug), html.Class(cls), d.Title)
	})
}

// indexContent is the landing page: a short intro plus a card grid linking to
// every component page, mirroring the shadcn/ui components index.
func indexContent(reg *Registry) mx.Component {
	return html.Div(
		html.H1(html.Class("text-3xl font-bold tracking-tight"), "Components"),
		html.P(html.Class("mt-2 text-lg text-muted-foreground"),
			"A go-mx rebuild of the shadcn/ui component docs — every preview is rendered server-side in Go, with its source shown alongside."),
		html.Div(html.Class("mt-8 grid gap-4 sm:grid-cols-2"),
			mx.ForEach(reg.Docs, func(d ComponentDoc) mx.Component {
				return html.A(html.HRef("/components/"+d.Slug),
					html.Class("rounded-lg border p-4 transition-colors hover:bg-accent"),
					html.Div(html.Class("font-medium"), d.Title),
					html.P(html.Class("mt-1 text-sm text-muted-foreground"), d.Description),
				)
			}),
		),
	)
}

// componentContent is one component's page: heading, description and every
// labeled preview.
func componentContent(d *ComponentDoc) mx.Component {
	blocks := make([]any, len(d.Examples))
	for i, ex := range d.Examples {
		blocks[i] = previewBlock(d.Slug, ex)
	}
	return html.Div(
		html.H1(html.Class("text-3xl font-bold tracking-tight"), d.Title),
		html.P(html.Class("mt-2 text-lg text-muted-foreground"), d.Description),
		html.Div(append([]any{html.Class("mt-10")}, blocks...)...),
	)
}

// previewBlock renders one example as a shadcn Tabs pair: a centered live
// preview and the Go source that produced it.
func previewBlock(slug string, ex Example) mx.Component {
	id := tabID(slug, ex.Name)
	return html.Section(html.Class("mb-10"),
		html.H3(html.Class("mb-3 text-sm font-medium text-muted-foreground"), ex.Name),
		shadcn.Tabs(id,
			shadcn.TabsList(
				shadcn.TabsTrigger(id, "preview", true, "Preview"),
				shadcn.TabsTrigger(id, "code", false, "Code"),
			),
			shadcn.TabsContent(id, "preview", true,
				html.Div(html.Class("mt-2 flex min-h-[220px] items-center justify-center rounded-md border p-10"),
					ex.Func(),
				),
			),
			shadcn.TabsContent(id, "code", false,
				codeBlock(ex.Source),
			),
		),
	)
}

// codeBlock renders the <pre><code> for an example's source and embeds it as
// raw HTML. The page is written with an indenting CheckedWriter that
// pretty-prints nested elements; inside a whitespace-preserving <pre> that
// would inject the ancestor indentation as leading spaces on the first source
// line. Rendering the block here with a non-indenting writer and inserting the
// result verbatim keeps the source exactly as written.
//
// The <pre> carries a dark placeholder background so the block looks right for
// the moment before Shiki paints; shikiScript then replaces this whole <pre>
// with Shiki's themed output (see head). The language-go class is both Shiki's
// selector and its language hint.
func codeBlock(source string) mx.Component {
	var buf strings.Builder
	block := html.Pre(
		html.Class("mt-2 overflow-x-auto rounded-md p-4 text-sm leading-relaxed bg-zinc-950 text-zinc-50"),
		html.Code(html.Class("language-go"), source),
	)
	if err := block.Render(context.Background(), mx.NewCheckedWriter(&buf)); err != nil {
		return html.Code(source) // fallback: never reached for static markup
	}
	return html.Raw(buf.String())
}

// tabID builds a stable, validateID-safe id for an example's Tabs instance,
// e.g. ("button", "Sizes") -> "tab-button-sizes".
func tabID(slug, name string) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		default:
			return '-'
		}
	}, name)
	return "tab-" + slug + "-" + safe
}
