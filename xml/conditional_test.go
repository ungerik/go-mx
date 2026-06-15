package xml_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/xml"
)

func TestConditionalAndForEach(t *testing.T) {
	got := render(t, xml.Element("list",
		xml.If(true, xml.Element("a")),
		xml.If(false, xml.Element("b")),
		xml.ForEach([]string{"x", "y"}, func(s string) *mx.Element { return xml.Element("item", s) }),
	))
	require.Equal(t, `<list><a></a><item>x</item><item>y</item></list>`, got)
}

// ExampleForEach renders a catalog to stdout with indented formatting, using
// ForEach to emit one <book> per slice element and If to include the optional
// <in-stock/> marker only for books that are in stock. The go test framework
// compares the printed output against the Output comment below.
func ExampleForEach() {
	type Book struct {
		Title   string
		InStock bool
	}
	books := []Book{
		{Title: "The Go Programming Language", InStock: true},
		{Title: "Out of Print Classic", InStock: false},
	}

	catalog := xml.Element("catalog",
		xml.ForEach(books, func(b Book) *mx.Element {
			return xml.Element("book",
				xml.Element("title", b.Title),
				xml.If(b.InStock, xml.EmptyElement("in-stock")),
			)
		}),
	)

	w := mx.NewCheckedWriter(os.Stdout).WithIndent("", "  ")
	if err := catalog.Render(context.Background(), w); err != nil {
		panic(err)
	}

	// Output:
	// <catalog>
	//   <book>
	//     <title>The Go Programming Language</title>
	//     <in-stock/>
	//   </book>
	//   <book>
	//     <title>Out of Print Classic</title>
	//   </book>
	// </catalog>
}
