# pdftable

[![Go Reference](https://pkg.go.dev/badge/github.com/ungerik/go-mx/pkg/pdftable.svg)](https://pkg.go.dev/github.com/ungerik/go-mx/pkg/pdftable)

Data tables for the go-mx [`pdf`](../../pdf) package: measured columns,
wrapped and aligned cell text, styled header rows, grid rules, and automatic
page breaks between rows with the header repeated on every page.

```go
import (
    "github.com/ungerik/go-mx/pdf"
    "github.com/ungerik/go-mx/pkg/pdftable"
)

table := pdftable.New(
    []pdftable.Column{
        {Title: "Item", Weight: 1},
        {Title: "Qty", Style: &pdftable.Style{HAlign: pdf.AlignRight}},
        {Title: "Price", Width: 22, Style: &pdftable.Style{HAlign: pdf.AlignRight}},
    },
    []string{"Golden Delicious", "12", "3.40"},
    []string{"Bananas", "3", "1.99"},
)
table.AddRow("Cherries", 5, "7.25") // any value converts like pdf children

doc := pdf.NewDocument("Invoice", table)
err := doc.OutputFile(ctx, "invoice.pdf")
```

`Table` implements `pdf.Component`, so it composes with `pdf.Document`,
`ForEach`, headers/footers and the rest of the component model.

## Why a retained structure

Every other `pdf` primitive draws immediately at the cursor. A table cannot:
column widths must be consistent across all rows and pages, and a row must
not be torn apart by fpdf's automatic mid-cell page break. So `Render` runs
in two passes — **measure** (resolve column widths against the page
dimensions, wrap every cell, compute row heights) and **draw** (paint row by
row, breaking between rows). fpdf's automatic page break is suspended while
the table draws and restored afterwards, as are the font, the text and fill
colors and the cell margin.

## Column sizing

The first set field of a `Column` decides its sizing mode:

| Field        | Mode                                     |
| ------------ | ---------------------------------------- |
| `Width > 0`  | fixed width in document units            |
| `Weight > 0` | proportional share of the leftover width |
| neither      | auto: the widest cell content, capped by `MaxWidth` |

The table spans from the cursor x to `Table.Width` or the right margin.
Fixed columns take their width, auto columns their measured content width,
and weighted columns share what is left. When auto columns demand more than
the available width they are scaled down proportionally to fit (like HTML
table layout) — unless weighted columns compete for the same space, which is
a deferred error instead of collapsing them to zero width. Fixed widths
exceeding the table width are also an error.

## Styles

`Style` covers font (family/style/size), text and fill color, horizontal and
vertical alignment, padding and line height. Every zero field means
*inherit*; the cascade is **cell → row → column → table → renderer state**:

```go
table.Style = pdftable.Style{FontSize: 10}              // whole table
table.HeaderStyle = pdftable.Style{FillColor: &pdf.Silver} // header row (bold by default)
row.Style = &pdftable.Style{FillColor: &pdf.Color{R: 235, G: 240, B: 248}} // zebra stripe
```

The pointer-typed fields (`FontStyle`, `TextColor`, `FillColor`) distinguish
"unset" from a deliberate zero value and are set with `new`:
`Style{FontStyle: new(pdf.StyleBoldItalic)}`. Padding inherits the
renderer's cell margin by default; a negative `Padding` means none.
`pdf.AlignJustify` is not supported in tables and falls back to left.

## Page breaks

Before each row the remaining page height is checked: a row that does not
fit moves to the next page whole. Page breaks run the document's `Footer`
and `Header` components as usual, and the table's header row is redrawn when
`RepeatHeader` is set (the `New` default). A header is never left orphaned
at the bottom of a page — it moves along with the first row. Only a row
taller than a whole page is split, between text lines, drawn as top-aligned
fragments.

The measurement helpers this builds on are exported on the renderer for
general use: `Renderer.ContentWidth`, `ContentHeight` and `RemainingHeight`,
plus the existing `GetStringWidth`, `SplitText` and `LineHeight`.

## Grid

`Grid` combines four independent rule sets — `GridOuter`, `GridHeader`,
`GridRows`, `GridCols` — with named combinations up to `GridAll` (the `New`
default), concatenating like `pdf.Border`: `GridOuter + GridRows ==
GridOuterRows`. Rules are stroked with the renderer's current draw color and
line width, so `pdf.DrawColor`/`pdf.LineWidth` before the table style the
grid. A table spanning pages closes and reopens its outline on each page.

## Non-text cells

A cell with a `Draw` callback renders custom content instead of text — an
image, an SVG icon, raw vector graphics — inside the cell's padded content
box, with `Height` reserving its vertical space:

```go
pdftable.Cell{
    Height: 6,
    Draw: func(ctx context.Context, r *pdf.Renderer, x, y, w, h float64) error {
        return pdf.SVG(icon, x, y, min(w, h), min(w, h)).Render(ctx, r)
    },
}
```

## Non-goals for v1

Deliberately out of scope for this first version:

- **No `colspan`/`rowspan`** — every row has one cell per column.
- **No nested tables** — a `Table` is not a valid cell content.
- **No per-cell border overrides beyond `Grid`** — rules are a table-level
  choice (the `pdf.Border` enum stays what it is: the per-box border of
  `pdf.CellFormat`).
- **No arbitrary `pdf.Component` children in cells** — components are not
  measurable; the `Draw` callback is the escape hatch.
- **No generic `TableOf[T]`** (column definitions with `func(T) string`
  extractors, the table analog of `pdf.ForEach`) yet.

All of these are documented limitations, not accidents — and `colspan` plus
`TableOf` are the natural phase 2.

Smaller behavioral limitations: `Draw` cells are not split across pages, and
rows split across pages lose per-row `MinHeight` leftovers and are always
top-aligned.

The committed visual reference is `testdata/table_reference.pdf`, rendered
by `TestGoldenPDF` (regenerate with
`go test ./pkg/pdftable -run TestGoldenPDF -update`).
