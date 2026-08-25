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

// renderDocument renders d and returns the produced HTML, failing the test on
// a render error.
func renderDocument(t *testing.T, d *Document) string {
	t.Helper()
	var b strings.Builder
	if err := d.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("Document.Render: %v", err)
	}
	return b.String()
}

// TestDocumentLang guards the <html lang> attribute added for WCAG 3.1.1.
func TestDocumentLang(t *testing.T) {
	// The zero value must render exactly as before: a bare <html> tag with
	// no lang attribute at all.
	t.Run("empty omits attribute", func(t *testing.T) {
		got := renderDocument(t, &Document{Title: "T"})
		if !strings.Contains(got, "<html>") {
			t.Errorf("empty Lang should render bare <html>\ngot: %s", got)
		}
		if strings.Contains(got, "lang=") {
			t.Errorf("empty Lang must not emit a lang attribute\ngot: %s", got)
		}
	})

	// A valid tag is interpolated verbatim (the charset is injection-proof).
	for _, tt := range []struct {
		name string
		lang string
		want string
	}{
		{"simple subtag", "de", `<html lang="de">`},
		{"hyphenated subtag", "de-AT", `<html lang="de-AT">`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := renderDocument(t, &Document{Lang: tt.lang, Title: "T"})
			if !strings.Contains(got, tt.want) {
				t.Errorf("Lang %q: missing %q\ngot: %s", tt.lang, tt.want, got)
			}
		})
	}

	// An invalid tag must abort Render with an error rather than escape or
	// interpolate attribute-breaking characters into the markup.
	t.Run("invalid tag rejected", func(t *testing.T) {
		var b strings.Builder
		err := (&Document{Lang: `de" onload=x`, Title: "T"}).
			Render(context.Background(), mx.NewCheckedWriter(&b))
		if err == nil {
			t.Fatalf("invalid Lang should error, got markup: %s", b.String())
		}
		if strings.Contains(b.String(), "onload") {
			t.Errorf("invalid Lang must not reach the output: %s", b.String())
		}
	})
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
