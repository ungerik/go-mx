package html

// TableView is an interface implemented by
// types with table like data
// to enable reading (viewing) the data
// in a uniform table like way.
//
// The design of this package assumes that
// the contents of a TableView are first read
// into memory and then wrapped as TableView,
// so the TableView methods don't need a
// context parameter and error result.
type TableView interface {
	// Title of the View
	Title() string
	// Columns returns the column names
	// which can be empty strings.
	Columns() []string
	// Numrows returns the number of rows
	NumRows() int
	// Cell returns the empty interface value of the cell at the given row and column.
	// If row and col are out of bounds then nil is returned.
	Cell(row, col int) any
}
