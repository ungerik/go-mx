---
title: The public knowledge layer
---

# The public knowledge layer

A page is read by more than browsers. Search engines want a canonical URL and
the page's translations; AI agents do better with Markdown than with layout and
scripts, and look for an `llms.txt` index or an MCP endpoint; rich results need
schema.org structured data. go-mx derives all of that from the same
[`web.Site`](https://pkg.go.dev/github.com/ungerik/go-mx/web) and its pages, so
every surface publishes the same facts.

**One rule decides what is public:** `Page.Indexable()` — published, not
scheduled, not marked `NoIndex`. The sitemap, the `llms.txt` helper and the MCP
tools list only pages it is true for. Everything here is generic; whatever
needs to know what a site sells (prices, offerings, opening hours) is the
site's own code, added through the extension points below.

## The pieces

| Piece | What it gives a page or a site |
|-------|--------------------------------|
| `Page.Alternates` | Translations and other formats of a page. `Site.AddPageMetadata` renders them as `<link>` elements, `Site.PageHeaders` as `Link` headers — from one list, so the two cannot disagree. A page with translations names its own language too (`Page.Lang`, else `Site.Lang`), and `Site.DefaultLang` adds the `x-default` link. |
| Markdown alternates | `Site.MarkdownAlternate` turns a page with Markdown content into a standalone Markdown document: a header with title, description, canonical URL, language, last update and translations, your own `Sections`, then the page's source. `MarkdownPath` gives its URL (`/en/faq` → `/en/faq.md`), `Site.ServeMarkdown` serves it, and `AcceptsMarkdown` tells a client that asked for Markdown from a browser. |
| `web.LLMsTxt` | An [llms.txt](https://llmstxt.org) index, validated before it is written. `Site.LLMsTxt` builds one listing the Markdown alternates of all public pages, grouped by `Page.Type`. |
| `html.JSONLD` | A `<script type="application/ld+json">` that is safe for untrusted values. |
| `web/schemaorg` | Typed schema.org values for it: `Organization`, `Person`, `WebPage`, `FAQPage`, `Event`, `Offer`, … Prices are built from integer cents with `schemaorg.Price`. |
| `mcp` module | The public pages as read-only MCP tools (`search_pages`, `list_pages`, `get_page`) on an MCP server you own, plus `GuardHandler` for the endpoint. See the [mcp README](https://github.com/ungerik/go-mx/blob/main/mcp/README.md). |

## Wiring a site

The site's pages carry their language and alternates:

```go
page := &web.Page{
    Path:  "/en/faq",
    Lang:  "en",
    Title: "FAQ",
    Alternates: []web.Alternate{
        {HrefLang: "de", Path: "/de/faq"},
        {Type: "text/markdown", Path: "/en/faq.md"},
    },
    // ... Published, ContentType: mx.ContentTypeMarkdown, Content: source
}
```

A `GlobPageSource` reads `lang` from the front matter; alternates are up to the
source, since only it knows which pages translate each other.

The `Site` registers robots.txt and the sitemap itself. llms.txt, the Markdown
alternates and the MCP endpoint are publishing decisions, so they are wired
explicitly:

```go
site := &web.Site{
    BaseURL:     "https://example.com",
    Title:       "Example",
    Description: "What the site is about.",
    Lang:        "en",
    DefaultLang: "en",
    Sources:     []web.PageSource{pages},
}
mux := http.NewServeMux()
if err := site.RegisterRoutes(mux); err != nil {
    log.Fatal(err)
}

mux.HandleFunc("GET "+web.LLMsTxtPath, func(w http.ResponseWriter, r *http.Request) {
    llms, err := site.LLMsTxt(r.Context(), web.LLMsTxtOptions{})
    if err != nil {
        http.Error(w, "llms.txt unavailable", http.StatusInternalServerError)
        return
    }
    llms.HandleHTTP(w, r)
})

mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
    page := findPage(r.URL.Path) // the site's own lookup, by page or .md path
    if page == nil || !page.Indexable() {
        http.NotFound(w, r)
        return
    }
    // The same page as Markdown, at its .md URL or for a client asking for it.
    if strings.HasSuffix(r.URL.Path, ".md") || web.AcceptsMarkdown(r) {
        site.ServeMarkdown(w, r, page, revision(page), page.LastUpdated)
        return
    }

    header, err := site.PageHeaders(page)
    if err != nil {
        http.Error(w, "page unavailable", http.StatusInternalServerError)
        return
    }
    for _, link := range header.Values("Link") {
        w.Header().Add("Link", link)
    }
    doc, err := site.RenderPage(r.Context(), page)
    if err != nil {
        http.Error(w, "page unavailable", http.StatusInternalServerError)
        return
    }
    org := schemaorg.NewOrganization(site.Title)
    org.URL = site.BaseURL
    doc.HeadCustom = mx.Components{doc.HeadCustom, html.JSONLD(org)}
    doc.HandleHTTP(w, r)
})
```

`site.RenderPage` adds the canonical link and the alternates to the head;
`PageHeaders` puts the same links on the response. `ServeMarkdown` marks its
response with a canonical `Link` back to the HTML page, so search engines
index the page, not its Markdown copy, and with `Vary: Accept`, so a cache
does not hand the Markdown to a browser.

The MCP endpoint goes on the same mux; the
[mcp README](https://github.com/ungerik/go-mx/blob/main/mcp/README.md) shows the
wiring and how to add the site's own tools next to the page tools.
