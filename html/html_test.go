package html_test

import (
	"context"
	"os"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

func ExampleHTML() {
	html.HTML{
		Title: "Hello World",
		Body: mx.Components{
			html.H1Text("Hello World"),
			html.N, // Raw "\n" for source readability
			html.Div{
				Class:    "content",
				Attribs:  html.Attribs{html.Lang("en")}, // Less common attribute
				Children: html.Text(">>Simple HTML page<<"),
			},
			html.Raw("\n<i>Raw</i> HTML"),
			mx.FuncComponent(func(ctx context.Context) (mx.Component, error) {
				var children mx.Components
				for i := range 3 {
					if i%2 == 0 {
						children = append(children, html.N, html.Textf("Even number: %d", i), html.BR)
					}
				}
				return children, nil
			}),
		},
	}.Render(context.Background(), os.Stdout)

	// Output:
	// <!DOCTYPE html>
	// <html>
	// <head>
	// <title>Hello World</title>
	// </head>
	// <body>
	// <h1>Hello World</h1>
	// <div class="content" lang="en">&gt;&gt;Simple HTML page&lt;&lt;</div>
	// <i>Raw</i> HTML
	// Even number: 0<br/>
	// Even number: 2<br/>
	// </body>
	// </html>
}
