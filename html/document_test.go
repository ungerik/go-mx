package html

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

// TestDocumentMetaPropertyOpenGraph guards that Document.MetaProperty renders
// valid Open Graph markup (<meta property="og:title" content="…">) rather than
// using the property key as the attribute name.
func TestDocumentMetaPropertyOpenGraph(t *testing.T) {
	var b strings.Builder
	err := (&Document{
		Title:        "T",
		MetaProperty: map[string]string{"og:title": "Hi"},
	}).Render(context.Background(), mx.NewCheckedWriter(&b))
	if err != nil {
		t.Fatal(err)
	}
	const want = `<meta property="og:title" content="Hi"/>`
	if !strings.Contains(b.String(), want) {
		t.Errorf("Document.MetaProperty render missing Open Graph markup\ngot:  %s\nwant substring: %s", b.String(), want)
	}
}

func ExampleNewDocument() {
	NewDocument("Hello World", // title
		// body:
		H1("Hello World"),
		Div(Class("content"), Lang("en"), ">>Simple HTML page<<"),
		mx.Newline,
		Raw("<p>Raw HTML</p>"),
		func() (children mx.Components) {
			for i := range 3 {
				if i%2 == 0 {
					children = append(children, mx.Newline, Textf("Even number: %d", i), Br())
				}
			}
			return children
		},
	).Render(
		context.Background(),
		mx.NewCheckedWriter(os.Stdout).WithIndent("", "  "),
	)

	// Output:
	// <!DOCTYPE html>
	// <html>
	// <head>
	//   <meta charset="UTF-8"/>
	//   <title>Hello World</title>
	// </head>
	// <body>
	//   <h1>Hello World</h1>
	//   <div class="content" lang="en">&gt;&gt;Simple HTML page&lt;&lt;</div>
	//   <p>Raw HTML</p>
	//   Even number: 0
	//   <br/>
	//   Even number: 2
	//   <br/>
	// </body>
	// </html>
}
