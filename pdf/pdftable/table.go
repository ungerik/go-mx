// Package pdftable renders data tables with the go-mx pdf package: measured
// columns, wrapped and aligned cell text, styled headers, grid rules, and
// automatic page breaks between rows with the header row repeated on every
// page.
//
// A [Table] is a retained structure, unlike the immediate-mode pdf primitives:
// Render first measures (resolves column widths against the page dimensions
// and wraps every cell to compute row heights) and then draws row by row,
// breaking to a new page whenever the next row does not fit. It implements
// [pdf.Component], so it composes with pdf.Document, Save, ForEach and the
// rest of the component model:
//
//	table := pdftable.New(
//	    []pdftable.Column{
//	        {Title: "Item", Weight: 1},
//	        {Title: "Qty", HeaderStyle: &pdftable.Style{HAlign: pdf.AlignRight}},
//	        {Title: "Price", Style: &pdftable.Style{HAlign: pdf.AlignRight}},
//	    },
//	    []string{"Golden Delicious", "12", "3.40"},
//	    []string{"Bananas", "3", "1.99"},
//	)
//	doc := pdf.NewDocument("Invoice", table)
package pdftable

import (
	"context"
	"fmt"

	pretty "github.com/domonda/go-pretty"

	"github.com/ungerik/go-mx/pdf"
)

var _ pdf.Component = (*Table)(nil)

// Table is a data table rendered as a pdf.Component. The zero value plus
// Columns and Rows is usable but strokes no grid and does not repeat the
// header row after page breaks; [New] applies the common defaults
// (GridAll, RepeatHeader).
type Table struct {
	// Columns defines the column count, widths and titles. Required.
	Columns []Column

	// Rows holds the body rows. Rows with fewer cells than columns are
	// padded with empty cells; more cells than columns is an error.
	Rows []Row

	// Width is the total table width in document units. Zero extends the
	// table from the cursor x position to the right margin. When all
	// columns are fixed or auto sized and demand less, the table is
	// narrower than Width; weighted columns absorb the leftover width.
	Width float64

	// Style holds the table-wide defaults for body cells. Unset fields
	// fall back to the renderer's state at render time (current font and
	// text color, left/middle alignment, the cell margin as padding).
	Style Style

	// HeaderStyle holds the defaults for header cells. Unset fields fall
	// back to Style with a bold font.
	HeaderStyle Style

	// Grid selects which rules are stroked, with the renderer's current
	// draw color and line width.
	Grid Grid

	// RepeatHeader redraws the header row at the top of every page the
	// table breaks onto.
	RepeatHeader bool
}

// Column defines one table column. The width is determined by the first set
// field: Width (fixed), Weight (share of the width left over after fixed and
// auto columns), or neither, which sizes the column to its widest cell
// content ("auto"), optionally capped by MaxWidth. When the auto columns
// demand more than the available width and no weighted columns compete, they
// are scaled down proportionally to fit.
type Column struct {
	// Title is the header cell text. A header row is rendered when any
	// column has a non-empty Title.
	Title string

	// Width > 0 fixes the column width in document units.
	Width float64

	// Weight > 0 gives the column a proportional share of the leftover
	// width (ignored when Width is set).
	Weight float64

	// MaxWidth caps the measured width of an auto column. Zero means no cap.
	MaxWidth float64

	// Style is merged over the table Style for this column's body cells.
	Style *Style

	// HeaderStyle is merged over the table HeaderStyle for this column's
	// header cell.
	HeaderStyle *Style
}

// Row is one body row of a table.
type Row struct {
	// Cells holds the row's cells, in column order.
	Cells []Cell

	// MinHeight is a lower bound for the row height in document units.
	MinHeight float64

	// Style is merged over the column and table styles for all cells of
	// this row (e.g. zebra striping via FillColor, or a bold totals row).
	Style *Style
}

// Cell is one table cell. Text cells wrap at the column width; a cell with a
// Draw callback renders custom content (an image, an SVG icon, raw vector
// graphics) instead of text.
type Cell struct {
	// Text is the cell content, word-wrapped to the column width.
	// Embedded "\n" force line breaks.
	Text string

	// Style is merged over the row, column and table styles.
	Style *Style

	// Draw, if set, replaces Text: it is called with the cell's inner
	// content box (inside the padding) after the background fill is
	// painted. Height provides the content height for row measurement.
	// Draw cells are not split across pages; when another cell forces the
	// row to split, the callback receives its box on the first fragment.
	Draw func(ctx context.Context, r *pdf.Renderer, x, y, w, h float64) error

	// Height is the content height in document units reserved for Draw.
	Height float64
}

// Style holds the visual properties of table cells. The zero value of every
// field means "inherit": cell styles are merged over the row, column and
// table styles, and whatever is still unset falls back to the renderer's
// state when the table renders. The pointer-typed fields are set with new:
// Style{FontStyle: new(pdf.StyleBoldItalic), TextColor: new(pdf.Red)}.
type Style struct {
	// FontFamily is the font family name; "" inherits.
	FontFamily string

	// FontStyle is the font style; nil inherits (pointer, because
	// pdf.StyleRegular is the zero value and must be settable).
	FontStyle *pdf.FontStyle

	// FontSize is the font size in points; 0 inherits.
	FontSize float64

	// TextColor is the text color; nil inherits.
	TextColor *pdf.Color

	// FillColor is the cell background color; nil inherits (default: no fill).
	FillColor *pdf.Color

	// HAlign is the horizontal text alignment; "" inherits (default left).
	// pdf.AlignJustify is not supported in tables and falls back to left.
	HAlign pdf.HAlign

	// VAlign is the vertical text alignment; "" inherits (default middle).
	// Rows split across pages are always drawn top-aligned.
	VAlign pdf.VAlign

	// Padding is the inner cell padding in document units on all four
	// sides; 0 inherits (default: the renderer's cell margin), a negative
	// value means no padding.
	Padding float64

	// LineHeight is the text line height in document units; 0 inherits
	// (default: the renderer's line height for the cell's font).
	LineHeight float64
}

// over returns s with every unset field inherited from base.
func (s Style) over(base Style) Style {
	if s.FontFamily == "" {
		s.FontFamily = base.FontFamily
	}
	if s.FontStyle == nil {
		s.FontStyle = base.FontStyle
	}
	if s.FontSize == 0 {
		s.FontSize = base.FontSize
	}
	if s.TextColor == nil {
		s.TextColor = base.TextColor
	}
	if s.FillColor == nil {
		s.FillColor = base.FillColor
	}
	if s.HAlign == "" {
		s.HAlign = base.HAlign
	}
	if s.VAlign == "" {
		s.VAlign = base.VAlign
	}
	if s.Padding == 0 {
		s.Padding = base.Padding
	}
	if s.LineHeight == 0 {
		s.LineHeight = base.LineHeight
	}
	return s
}

// overPtr is over for the optional *Style fields: nil inherits base unchanged.
func overPtr(s *Style, base Style) Style {
	if s == nil {
		return base
	}
	return s.over(base)
}

// pad resolves the effective padding: negative means none.
func (s Style) pad() float64 {
	return max(s.Padding, 0)
}

// fontStyle resolves the effective font style, defaulting to regular.
func (s Style) fontStyle() pdf.FontStyle {
	if s.FontStyle == nil {
		return pdf.StyleRegular
	}
	return *s.FontStyle
}

// applyFont selects the style's font on the renderer.
func (s Style) applyFont(r *pdf.Renderer) {
	r.SetFont(s.FontFamily, string(s.fontStyle()), s.FontSize)
}

// New creates a Table with the given columns and optional body rows of plain
// string cells, with the common defaults applied: the full grid and the
// header row repeated after page breaks.
func New(columns []Column, rows ...[]string) *Table {
	t := &Table{
		Columns:      columns,
		Rows:         make([]Row, 0, len(rows)),
		Grid:         GridAll,
		RepeatHeader: true,
	}
	for _, texts := range rows {
		cells := make([]Cell, len(texts))
		for i, text := range texts {
			cells[i] = Cell{Text: text}
		}
		t.Rows = append(t.Rows, Row{Cells: cells})
	}
	return t
}

// AddRow appends a body row, converting each value with [AsCell], and returns
// the table for chaining.
func (t *Table) AddRow(cells ...any) *Table {
	row := Row{Cells: make([]Cell, len(cells))}
	for i, cell := range cells {
		row.Cells[i] = AsCell(cell)
	}
	t.Rows = append(t.Rows, row)
	return t
}

// AsCell converts an arbitrary value into a [Cell], following the same
// philosophy as pdf.AsComponent: nil becomes an empty cell, a Cell passes
// through, a string becomes the cell text, error and fmt.Stringer render as
// their own text, and anything else falls back to its go-pretty
// representation.
func AsCell(value any) Cell {
	switch v := value.(type) {
	case nil:
		return Cell{}
	case Cell:
		return v
	case *Cell:
		if v == nil {
			return Cell{}
		}
		return *v
	case string:
		return Cell{Text: v}
	case error:
		return Cell{Text: v.Error()}
	case fmt.Stringer:
		return Cell{Text: v.String()}
	default:
		return Cell{Text: pretty.Sprint(v)}
	}
}

// hasHeader reports whether any column defines a header title.
func (t *Table) hasHeader() bool {
	for i := range t.Columns {
		if t.Columns[i].Title != "" {
			return true
		}
	}
	return false
}
