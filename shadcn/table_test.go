package shadcn

import (
	"strings"
	"testing"
)

func TestTable(t *testing.T) {
	out := render(t, Table(
		TableCaption("A list of invoices."),
		TableHeader(TableRow(TableHead("Invoice"))),
		TableBody(TableRow(TableCell("INV001"))),
		TableFooter(TableRow(TableCell("Total"))),
	))
	for _, want := range []string{
		`data-slot="table-container"`,
		`data-slot="table"`,
		`data-slot="table-caption"`,
		`data-slot="table-header"`,
		`data-slot="table-body"`,
		`data-slot="table-footer"`,
		`data-slot="table-row"`,
		`data-slot="table-head"`,
		`data-slot="table-cell"`,
		"overflow-x-auto",
		"caption-bottom",
		">Invoice<",
		">INV001<",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

// TestTableContainerWrapsTable guards the container/table nesting: the inner
// <table> must carry the caller class while the wrapper provides the overflow.
func TestTableContainerWrapsTable(t *testing.T) {
	out := render(t, Table())
	if !strings.HasPrefix(out, `<div data-slot="table-container"`) {
		t.Errorf("Table must render the container <div> first: %s", out)
	}
	if !strings.Contains(out, `<table data-slot="table"`) {
		t.Errorf("Table must nest a <table> inside the container: %s", out)
	}
}
