package wordpress

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// ImportReportView renders the import diagnostics as a shadcn page — the
// in-browser surface of the wedge, linking each finding to the posts it came
// from. Render it after the content pages so the findings are complete.
func (v *Views) ImportReportView() mx.Component {
	r := v.rep
	sections := []any{html.Class("mx-auto max-w-3xl"),
		html.H1(html.Class("text-3xl font-bold tracking-tight"), "Import report"),
		html.P(html.Class("mt-2 text-muted-foreground"), countsLine(r.Counts)),
	}
	sections = appendComp(sections, v.findingSection("Unknown shortcodes",
		"Plugin shortcodes can’t run; the delimiters were stripped and the inner content kept.",
		r.UnknownShortcodes), "mt-10")
	sections = appendComp(sections, v.findingSection("Plugin blocks",
		"Non-core Gutenberg blocks were rendered as their raw inner HTML, not specially handled.",
		r.UnsupportedBlocks), "mt-10")
	sections = appendComp(sections, v.findingSection("Removed markup",
		"Disallowed elements and attributes the sanitizer dropped.",
		r.DroppedHTML), "mt-10")
	sections = appendComp(sections, v.findingSection("Blocked URLs",
		"URLs with dangerous schemes (javascript:, data:text/html, …) that were removed.",
		r.BlockedURLs), "mt-10")
	if len(r.SkippedItems) > 0 {
		sections = append(sections, html.Div(html.Class("mt-10"), v.skippedSection(r.SkippedItems)))
	}
	if reportClean(r) {
		sections = append(sections, html.Div(html.Class("mt-10"),
			shadcn.Alert(shadcn.AlertDefault,
				shadcn.AlertTitle("Clean import"),
				shadcn.AlertDescription("No content issues were found."))))
	}
	return html.Div(sections...)
}

func (v *Views) findingSection(title, desc string, fs []Finding) mx.Component {
	if len(fs) == 0 {
		return nil
	}
	rows := make([]any, 0, len(fs)+1)
	rows = append(rows, html.Class("mt-3 space-y-2"))
	for _, f := range fs {
		rows = append(rows, html.LI(html.Class("flex flex-wrap items-center gap-2"),
			shadcn.Badge(shadcn.BadgeSecondary, fmt.Sprintf("×%d", f.Count)),
			html.Span(html.Class("font-mono text-sm"), f.Name),
			html.Span(html.Class("text-xs text-muted-foreground"), f.Disposition),
			v.findingPosts(f.PostIDs),
		))
	}
	return html.Section(
		html.H2(html.Class("text-xl font-semibold"), title),
		html.P(html.Class("text-sm text-muted-foreground"), desc),
		html.UL(rows...),
	)
}

func (v *Views) findingPosts(ids []int64) mx.Component {
	if len(ids) == 0 {
		return nil
	}
	links := make([]any, 0, len(ids)+1)
	links = append(links, html.Class("flex flex-wrap gap-2 text-xs"))
	for _, id := range ids {
		label := fmt.Sprintf("#%d", id)
		if r := v.pl.post[id]; r != "" {
			links = append(links, html.A(html.HRef(r), html.Class("underline text-muted-foreground"), label))
		} else if r := v.pl.page[id]; r != "" {
			links = append(links, html.A(html.HRef(r), html.Class("underline text-muted-foreground"), label))
		} else {
			links = append(links, html.Span(html.Class("text-muted-foreground"), label))
		}
	}
	return html.Span(links...)
}

func (v *Views) skippedSection(items []SkippedItem) mx.Component {
	rows := make([]any, 0, len(items)+1)
	rows = append(rows, html.Class("mt-3 space-y-1 text-sm"))
	for _, it := range items {
		label := it.Title
		if label == "" {
			label = fmt.Sprintf("#%d", it.PostID)
		}
		rows = append(rows, html.LI(
			shadcn.Badge(shadcn.BadgeOutline, it.PostType),
			html.Span(html.Class("ml-2"), label),
			html.Span(html.Class("ml-2 text-muted-foreground"), it.Reason),
		))
	}
	return html.Section(
		html.H2(html.Class("text-xl font-semibold"), "Skipped items"),
		html.UL(rows...),
	)
}

func countsLine(c Counts) string {
	return fmt.Sprintf("%d posts · %d pages · %d categories · %d tags · %d authors · %d attachments · %d comments",
		c.Posts, c.Pages, c.Categories, c.Tags, c.Authors, c.Attachments, c.Comments)
}

func reportClean(r *Report) bool {
	return len(r.UnknownShortcodes) == 0 && len(r.UnsupportedBlocks) == 0 &&
		len(r.DroppedHTML) == 0 && len(r.BlockedURLs) == 0 && len(r.SkippedItems) == 0
}
