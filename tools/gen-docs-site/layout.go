package main

import (
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/highlight"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/internal/shadcntheme"
	"github.com/ungerik/go-mx/shadcn"
)

// navItem is one link of the site navigation.
type navItem struct {
	// Title is the link text.
	Title string
	// File is the linked site file, relative to the site root.
	File string
}

// navSection is one titled group of the site navigation.
type navSection struct {
	Title string
	Items []navItem
}

// buildNav returns the navigation of the site: a section per Page.Section in
// the order of pages (see sortPages), with the Examples section after the
// documentation and before the Repository section.
func buildNav(pages []*Page, examples navSection) []navSection {
	var nav []navSection
	index := make(map[string]int)
	for _, p := range pages {
		if p.Section == repoSection && len(examples.Items) > 0 {
			if _, ok := index[examples.Title]; !ok {
				index[examples.Title] = len(nav)
				nav = append(nav, examples)
			}
		}
		i, ok := index[p.Section]
		if !ok {
			i = len(nav)
			index[p.Section] = i
			nav = append(nav, navSection{Title: p.Section})
		}
		nav[i].Items = append(nav[i].Items, navItem{Title: p.NavTitle, File: p.File()})
	}
	if _, ok := index[examples.Title]; !ok && len(examples.Items) > 0 {
		nav = append(nav, examples)
	}
	return nav
}

// docsCSS styles the rendered Markdown, the role the Tailwind typography
// plugin plays elsewhere, which the browser build cannot load. It is
// Tailwind input, compiled with the theme, so it uses the shadcn tokens and
// follows the dark mode.
const docsCSS = /*css*/ `
/* The dark mode, halfway from shadcn's neutral greys toward One Monokai's
   blue-grey backgrounds (the theme of the code blocks): page #0a0a0a and
   #282c34, sidebar and floating surfaces #171717 and #21252b, hovered and
   current items #262626 and #2f333d. theme.css stays shadcn's own. */
.dark {
  --background: #191b1f;
  --card: #1c1e21;
  --popover: #1c1e21;
  --sidebar: #1c1e21;
  --secondary: #2a2c32;
  --muted: #2a2c32;
  --accent: #2a2c32;
  --sidebar-accent: #2a2c32;
  --primary-foreground: #1c1e21;
}
.prose-mx { @apply text-base leading-7; }
.prose-mx > :first-child { @apply mt-0; }
.prose-mx h1 { @apply mt-0 mb-6 scroll-m-20 text-3xl font-bold tracking-tight text-balance; }
.prose-mx h2 { @apply mt-12 mb-4 scroll-m-20 border-b pb-2 text-2xl font-semibold tracking-tight; }
.prose-mx h3 { @apply mt-8 mb-3 scroll-m-20 text-xl font-semibold tracking-tight; }
.prose-mx h4 { @apply mt-6 mb-2 scroll-m-20 text-lg font-semibold; }
.prose-mx h5, .prose-mx h6 { @apply mt-6 mb-2 scroll-m-20 font-semibold; }
.prose-mx p { @apply my-4; }
.prose-mx a { @apply font-medium underline underline-offset-4 decoration-muted-foreground/50 hover:decoration-foreground; }
.prose-mx ul { @apply my-4 ml-6 list-disc; }
.prose-mx ol { @apply my-4 ml-6 list-decimal; }
.prose-mx li { @apply mt-1.5; }
.prose-mx li > ul, .prose-mx li > ol { @apply my-1.5; }
.prose-mx li:has(> input[type=checkbox]) { @apply list-none -ml-6; }
.prose-mx input[type=checkbox] { @apply mr-2 align-middle; }
.prose-mx blockquote { @apply my-6 border-l-2 pl-6 text-muted-foreground; }
.prose-mx :not(pre) > code { @apply rounded bg-muted px-[0.3rem] py-[0.15rem] font-mono text-[0.85em]; }
.prose-mx pre { @apply my-6 overflow-x-auto text-sm leading-relaxed; }
.prose-mx table { @apply my-6 block w-full overflow-x-auto text-sm; }
.prose-mx th { @apply border bg-muted/50 px-3 py-2 text-left font-semibold; }
.prose-mx td { @apply border px-3 py-2 align-top; }
.prose-mx hr { @apply my-10; }
.prose-mx img { @apply inline max-w-full; }
.prose-mx details { @apply my-4 rounded-md border px-4 py-2; }
.prose-mx summary { @apply cursor-pointer font-medium; }
.prose-mx dt { @apply mt-4 font-semibold; }
.prose-mx dd { @apply ml-6; }
.prose-mx sup { @apply text-xs; }
/* Footnote references in brackets, as GitHub shows them, so adjacent ones
   read [5][2][6] instead of 526. */
.prose-mx .footnote-ref { @apply no-underline; }
.prose-mx .footnote-ref::before { content: "["; }
.prose-mx .footnote-ref::after { content: "]"; }
.prose-mx .footnotes { @apply mt-12 text-sm text-muted-foreground; }
`

// shell renders a page of the site in file (relative to the site root): the
// head shared with the gallery (shadcntheme.Head) plus the prose styles and
// the light and dark highlight themes of the code blocks, and the shadcn
// Sidebar navigation around content. The document is returned unrendered, so
// the caller can add the page metadata.
func shell(nav []navSection, file, title string, content mx.Component) *html.Document {
	if title != "go-mx" {
		title += " · go-mx"
	}
	doc := html.NewDocument(title,
		shadcn.SidebarProvider(
			// The site's own styling over shadcn's: the sidebar stays in
			// view while the page scrolls, and collapses out of sight
			// rather than to an icon strip, since the navigation has no
			// icons. Below lg it is hidden like upstream's, whose mobile
			// Sheet the port does not have.
			shadcn.Sidebar(html.Class("sticky top-0 hidden h-svh self-start lg:flex group-data-[state=collapsed]/sidebar-wrapper:w-0 group-data-[state=collapsed]/sidebar-wrapper:border-r-0"),
				shadcn.SidebarHeader(html.Class("gap-0 p-0"), siteHeader(file, "theme-menu")),
				shadcn.SidebarContent(
					html.Nav(html.Attrib("aria-label", "Documentation"), sidebarNav(nav, file)),
				),
			),
			// The window scrolls, not the inset, so the sticky sidebar and
			// fragment links work as on any page.
			shadcn.SidebarInset(html.Class("min-w-0 overflow-visible"),
				// Below lg the navigation folds into a disclosure at the top,
				// no script needed.
				html.DivClass("border-b bg-sidebar text-sidebar-foreground lg:hidden",
					siteHeader(file, "theme-menu-small"),
					html.Details(html.Class("px-4 pb-3"),
						html.Summary(html.Class("cursor-pointer text-sm font-medium"), "Menu"),
						html.Nav(html.Class("mt-2"), html.Attrib("aria-label", "Documentation"), navLinks(nav, file)),
					),
				),
				html.DivClass("hidden px-4 pt-3 lg:flex", shadcn.SidebarTrigger()),
				// The column grows with the window and is capped only on
				// wide screens, where longer lines would get hard to read.
				html.DivClass("mx-auto w-full max-w-6xl px-6 py-10 lg:px-10 lg:pt-4",
					content,
				),
			),
		),
	)
	doc.Lang = "en"
	doc.Meta = map[string]string{"viewport": "width=device-width, initial-scale=1"}
	doc.HeadCustom = mx.Components{
		shadcntheme.Head(docsCSS),
		html.StyleElem(codeThemeCSS()),
	}
	return doc
}

// codeThemeCSS is the stylesheet of the code blocks: the colors of
// highlight.LightTheme, and those of highlight.OneMonokaiTheme below the .dark
// root the theme toggle sets, so code follows the page's mode.
func codeThemeCSS() string {
	var dark strings.Builder
	for line := range strings.Lines(highlight.OneMonokaiTheme.CSS("")) {
		// Every rule of the theme starts at the beginning of a line with
		// its selector, pre.hl or a .hl-… token class.
		if strings.HasPrefix(line, ".") || strings.HasPrefix(line, "pre.") {
			dark.WriteString(".dark ")
		}
		dark.WriteString(line)
	}
	return highlight.LightTheme.CSS("") + dark.String()
}

// siteHeader is the site title linking home, a link to the repository, and
// the light/dark/system mode toggle, whose menu gets the id themeMenuID: the
// header is in the page twice, in the sidebar and above the content on small
// screens.
func siteHeader(file, themeMenuID string) mx.Component {
	return html.DivClass("flex items-center justify-between gap-2 border-b px-4 py-3",
		html.AHRef(relativeURL(file, "index.html"), html.Class("text-sm font-semibold"), "go-mx"),
		html.DivClass("flex items-center gap-1",
			html.AHRef(repoURL, html.Class(shadcn.ButtonClasses(shadcn.ButtonGhost, shadcn.SizeSM)), "GitHub"),
			shadcntheme.ModeToggle(themeMenuID),
		),
	)
}

// sidebarNav renders the sections of nav as shadcn sidebar groups, the link
// to file marked as the current page. The group labels are restyled to stand
// out over the links, which are muted until hovered or current.
func sidebarNav(nav []navSection, file string) mx.Component {
	return mx.ForEach(nav, func(s navSection) mx.Component {
		return shadcn.SidebarGroup(
			shadcn.SidebarGroupLabel(html.Class("text-sm font-semibold text-sidebar-foreground"), s.Title),
			shadcn.SidebarGroupContent(
				shadcn.SidebarMenu(html.Class("gap-0.5"),
					mx.ForEach(s.Items, func(item navItem) mx.Component {
						link := []any{
							html.HRef(relativeURL(file, item.File)),
							html.Class("py-1.5 text-muted-foreground data-[active=true]:text-sidebar-accent-foreground"),
						}
						if item.File == file {
							link = append(link, html.DataAttr("active", "true"), html.Attrib("aria-current", "page"))
						}
						return shadcn.SidebarMenuItem(shadcn.SidebarMenuButton(append(link, item.Title)...))
					}),
				),
			),
		)
	})
}

// navLinks renders the sections of nav as plain links for the menu of small
// screens, the link to file marked as the current page.
func navLinks(nav []navSection, file string) mx.Component {
	return mx.ForEach(nav, func(s navSection) mx.Component {
		return mx.Components{
			html.DivClass("px-3 pt-5 pb-1.5 text-sm font-semibold text-foreground", s.Title),
			mx.ForEach(s.Items, func(item navItem) mx.Component {
				if item.File == file {
					return html.AHRef(relativeURL(file, item.File),
						html.Class("block rounded-md bg-accent px-3 py-1.5 text-sm font-medium text-accent-foreground"),
						html.Attrib("aria-current", "page"),
						item.Title)
				}
				return html.AHRef(relativeURL(file, item.File),
					html.Class("block rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground"),
					item.Title)
			}),
		}
	})
}

// markdownContent is the main column of a Markdown page: the rendered body,
// headed by the title when the Markdown has no level-one heading of its own,
// and a footer linking the page's source on GitHub and its Markdown
// alternate.
func markdownContent(page *Page, body, markdownFile string) mx.Component {
	return mx.Components{
		html.Article(html.Class("prose-mx"),
			mx.If(!page.hasH1, html.H1(page.Title)),
			html.Raw(body),
		),
		html.Footer(html.Class("mt-16 flex flex-wrap gap-x-6 gap-y-2 border-t pt-6 text-sm text-muted-foreground"),
			html.AHRef(repoURL+"blob/main/"+page.Source, html.Class("hover:text-foreground"), "View ", html.Code(page.Source), " on GitHub"),
			html.AHRef(relativeURL(page.File(), markdownFile), html.Class("hover:text-foreground"), html.Attrib("type", "text/markdown"), "Markdown version"),
		),
	}
}
