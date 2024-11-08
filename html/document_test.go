package html

import (
	"context"
	"os"

	"github.com/ungerik/go-mx"
)

func ExampleDocument() {
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
