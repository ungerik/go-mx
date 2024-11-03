package html

import (
	"context"
	"os"

	"github.com/ungerik/go-mx"
)

func ExampleDocument() {
	Document{
		Title: "Hello World",
		Body: mx.AsComponents(
			H1("Hello World"),
			Div(Class("content"), Lang("en"), ">>Simple HTML page<<"),
			mx.Raw("<p>Raw HTML without indentation</p>\n"),
			func() (children mx.Components) {
				for i := range 3 {
					if i%2 == 0 {
						children = append(children, Textf("Even number: %d", i), Br())
					}
				}
				return children
			},
		),
	}.Render(
		context.Background(),
		mx.NewCheckedWriter(os.Stdout).WithIndet("", "  "),
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
	// <p>Raw HTML without indentation</p>
	// Even number: 0  <br/>
	// Even number: 2  <br/>
	// </body>
	// </html>
}
