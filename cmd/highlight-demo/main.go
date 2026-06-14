// Command highlight-demo serves a single page that dogfoods the
// [github.com/ungerik/go-mx/highlight] package.
//
// It highlights a sample Go file two ways and shows both, side by side:
//
//   - the highlighted HTML of the sample, and
//   - the Go source (html.Pre/Code/Span calls) that highlight.GoSource
//     generates to build that same HTML — itself highlighted.
//
// The theme's CSS is emitted into the page <head>. Switch themes with the
// ?theme=dark query parameter.
//
//	go run ./cmd/highlight-demo
//
// Then browse to http://localhost:8080.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/highlight"
	"github.com/ungerik/go-mx/html"
)

// sample is the Go source highlighted on the demo page.
const sample = /*go*/ `package main

import "fmt"

// greet builds a greeting for name and prints it to stdout.
func greet(name string) {
	const prefix = "Hello"
	msg := fmt.Sprintf("%s, %s!", prefix, name)
	for i := 0; i < 3; i++ {
		fmt.Println(msg)
	}
}

func main() {
	greet("world")
}
`

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	http.HandleFunc("/", handle)

	log.Printf("highlight-demo listening on http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	theme := highlight.LightTheme
	other := "dark"
	if r.URL.Query().Get("theme") == "dark" {
		theme = highlight.DarkTheme
		other = "light"
	}

	// The generated Go source is plain text; highlight it too so the page
	// shows it the same way as the sample.
	generated := highlight.GoSource(sample)

	page := html.HTML(
		html.Lang("en"),
		html.Head(
			html.Meta(mx.NewAttrib("charset", "utf-8")),
			html.Meta(html.Name("viewport"), mx.NewAttrib("content", "width=device-width, initial-scale=1")),
			html.TitleElem("go-mx highlight demo"),
			theme.StyleElement(""),
			html.StyleElem( /*css*/ `
				body { font-family: ui-sans-serif, system-ui, sans-serif; margin: 2rem auto; max-width: 64rem; padding: 0 1rem; }
				h1 { font-size: 1.5rem; }
				h2 { font-size: 1.1rem; margin-top: 2rem; }
				a { color: #2563eb; }
				.cols { display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem; }
				@media (max-width: 48rem) { .cols { grid-template-columns: 1fr; } }
			`),
		),
		html.Body(
			html.H1("go-mx ", html.Code("highlight"), " demo"),
			html.P(
				"Theme: ", html.B(theme.Name), " · ",
				html.AHRef("?theme="+other, "switch to ", other),
			),
			html.DivClass("cols",
				html.Div(
					html.H2("Highlighted HTML"),
					highlight.Component(sample),
				),
				html.Div(
					html.H2("Generated ", html.Code("highlight.GoSource"), " output"),
					highlight.Component(generated),
				),
			),
		),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer := mx.NewCheckedWriter(w)
	if _, err := w.Write([]byte("<!DOCTYPE html>\n")); err != nil {
		return
	}
	if err := page.Render(r.Context(), writer); err != nil {
		log.Printf("render error: %v", err)
	}
}
