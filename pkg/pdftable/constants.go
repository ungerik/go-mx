//go:generate go -C ../../tools tool go-enum ../pkg/pdftable/$GOFILE

package pdftable

import (
	"fmt"
	"strings"
)

// Grid selects which table rules (lines) are stroked, as a combination of four
// independent parts: the outer table outline, the rule below the header row,
// the horizontal rules between body rows, and the vertical rules between
// columns. Like [pdf.Border], the named combinations double as concatenations
// in the canonical part order outer, header, rows, cols:
// GridOuter + GridRows == GridOuterRows. Rules are stroked with the renderer's
// current draw color and line width.
type Grid string //#enum

const (
	// GridNone strokes no rules at all.
	GridNone Grid = "" // no rules
	// GridOuter strokes the outer table outline (closed per page for tables spanning pages).
	GridOuter Grid = "O" // outer outline
	// GridHeader strokes the rule below the header row.
	GridHeader Grid = "H" // rule below the header
	// GridRows strokes the horizontal rules between body rows.
	GridRows Grid = "R" // rules between rows
	// GridCols strokes the vertical rules between columns.
	GridCols Grid = "C" // rules between columns
	// GridOuterHeader strokes the outline and the header rule.
	GridOuterHeader Grid = "OH" // outer + header
	// GridOuterRows strokes the outline and the rules between rows.
	GridOuterRows Grid = "OR" // outer + rows
	// GridOuterCols strokes the outline and the rules between columns.
	GridOuterCols Grid = "OC" // outer + cols
	// GridHeaderRows strokes the header rule and the rules between rows.
	GridHeaderRows Grid = "HR" // header + rows
	// GridHeaderCols strokes the header rule and the rules between columns.
	GridHeaderCols Grid = "HC" // header + cols
	// GridRowsCols strokes the rules between rows and between columns.
	GridRowsCols Grid = "RC" // rows + cols
	// GridOuterHeaderRows strokes everything except the rules between columns.
	GridOuterHeaderRows Grid = "OHR" // outer + header + rows
	// GridOuterHeaderCols strokes everything except the rules between rows.
	GridOuterHeaderCols Grid = "OHC" // outer + header + cols
	// GridOuterRowsCols strokes everything except the header rule.
	GridOuterRowsCols Grid = "ORC" // outer + rows + cols
	// GridHeaderRowsCols strokes all inner rules but no outline.
	GridHeaderRowsCols Grid = "HRC" // header + rows + cols
	// GridAll strokes every rule: the full grid.
	GridAll Grid = "OHRC" // all rules
)

// Valid indicates if g is any of the valid values for Grid
func (g Grid) Valid() bool {
	switch g {
	case
		GridNone,
		GridOuter,
		GridHeader,
		GridRows,
		GridCols,
		GridOuterHeader,
		GridOuterRows,
		GridOuterCols,
		GridHeaderRows,
		GridHeaderCols,
		GridRowsCols,
		GridOuterHeaderRows,
		GridOuterHeaderCols,
		GridOuterRowsCols,
		GridHeaderRowsCols,
		GridAll:
		return true
	}
	return false
}

// Validate returns an error if g is none of the valid values for Grid
func (g Grid) Validate() error {
	if !g.Valid() {
		return fmt.Errorf("invalid value %#v for type pdftable.Grid", g)
	}
	return nil
}

// Enums returns all valid values for Grid
func (Grid) Enums() []Grid {
	return []Grid{
		GridNone,
		GridOuter,
		GridHeader,
		GridRows,
		GridCols,
		GridOuterHeader,
		GridOuterRows,
		GridOuterCols,
		GridHeaderRows,
		GridHeaderCols,
		GridRowsCols,
		GridOuterHeaderRows,
		GridOuterHeaderCols,
		GridOuterRowsCols,
		GridHeaderRowsCols,
		GridAll,
	}
}

// EnumStrings returns all valid values for Grid as strings
func (Grid) EnumStrings() []string {
	return []string{
		"",
		"O",
		"H",
		"R",
		"C",
		"OH",
		"OR",
		"OC",
		"HR",
		"HC",
		"RC",
		"OHR",
		"OHC",
		"ORC",
		"HRC",
		"OHRC",
	}
}

// String implements the fmt.Stringer interface for Grid
func (g Grid) String() string {
	return string(g)
}

// Has reports whether g includes the given single part
// (GridOuter, GridHeader, GridRows or GridCols).
func (g Grid) Has(part Grid) bool {
	return part != GridNone && strings.Contains(string(g), string(part))
}
