package html

import (
	"context"
	"fmt"
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

func TestValidLangTag(t *testing.T) {
	for _, tag := range []string{"de", "de-AT", "zh-Hant-TW", "x-default"} {
		if !ValidLangTag(tag) {
			t.Errorf("ValidLangTag(%q) = false, want true", tag)
		}
	}
	// Everything else could break the attribute or header it is put into.
	for _, tag := range []string{"", "en_US", "de-", "-de", `de" onload=x`, "toolonglang"} {
		if ValidLangTag(tag) {
			t.Errorf("ValidLangTag(%q) = true, want false", tag)
		}
	}
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
	// With and without HTMLAttribs, since the <html> tag is built on two
	// paths: a constant without them, the writer with them, which would
	// escape a bad tag into the markup instead of refusing it.
	for _, attribs := range [][]mx.Attrib{nil, {Class("no-js")}} {
		t.Run(fmt.Sprintf("invalid tag rejected with %d HTMLAttribs", len(attribs)), func(t *testing.T) {
			var b strings.Builder
			err := (&Document{Lang: `de" onload=x`, HTMLAttribs: attribs, Title: "T"}).
				Render(context.Background(), mx.NewCheckedWriter(&b))
			if err == nil {
				t.Fatalf("invalid Lang should error, got markup: %s", b.String())
			}
			if b.Len() != 0 {
				t.Errorf("invalid Lang must stop Render before any output: %s", b.String())
			}
		})
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

// TestDocumentZeroAttribsUnchanged pins that a Document without HTMLAttribs
// and BodyAttribs renders byte-for-byte as before those fields existed, with
// and without Lang and with every other field set.
func TestDocumentZeroAttribsUnchanged(t *testing.T) {
	for _, tt := range []struct {
		name string
		lang string
		root string
	}{
		{"without lang", "", "<html>"},
		{"with lang", "en", `<html lang="en">`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := renderDocument(t, &Document{
				Lang:         tt.lang,
				Title:        "T",
				Meta:         map[string]string{"description": "D"},
				MetaProperty: map[string]string{"og:title": "O"},
				Stylesheets:  []string{"/site.css"},
				Style:        "body{margin:0}",
				HeadCustom:   Script(Src("/site.js")),
				BodyChildren: P("Hello"),
			})
			want := "<!DOCTYPE html>\n" + tt.root +
				`<head><meta charset="UTF-8"/><title>T</title>` +
				`<meta name="description" content="D"/>` +
				`<meta property="og:title" content="O"/>` +
				`<link rel="stylesheet" href="/site.css"/>` +
				`<style>body{margin:0}</style>` +
				`<script src="/site.js"></script></head>` +
				"<body><p>Hello</p></body>\n</html>"
			if got != want {
				t.Errorf("render changed\ngot:  %q\nwant: %q", got, want)
			}
		})
	}
}

// TestDocumentBodyAttribs guards that BodyAttribs reach <body> through the
// normal element path, escaped like any attribute value.
func TestDocumentBodyAttribs(t *testing.T) {
	t.Run("class id and data", func(t *testing.T) {
		got := renderDocument(t, &Document{
			BodyAttribs: []mx.Attrib{
				Class("home page-template-page_front-page siteorigin-panels"),
				ID("top"),
				DataAttr("theme", "dark"),
			},
			BodyChildren: P("Hello"),
		})
		const want = `<body class="home page-template-page_front-page siteorigin-panels" id="top" data-theme="dark"><p>Hello</p></body>`
		if !strings.Contains(got, want) {
			t.Errorf("missing %s\ngot: %s", want, got)
		}
	})

	t.Run("value escaped", func(t *testing.T) {
		got := renderDocument(t, &Document{
			BodyAttribs: []mx.Attrib{Class(`a"><script>x`)},
		})
		const want = `<body class="a&quot;>&lt;script>x">`
		if !strings.Contains(got, want) {
			t.Errorf("missing %s\ngot: %s", want, got)
		}
		if strings.Contains(got, "<script>") {
			t.Errorf("class value broke out of the attribute: %s", got)
		}
	})
}

// contentWithAttribs is a Component that also provides attributes, so the
// element constructors would sort it into attributes rather than children.
type contentWithAttribs struct{}

func (contentWithAttribs) Render(ctx context.Context, w mx.Writer) error {
	return P("content").Render(ctx, w)
}

func (contentWithAttribs) Attribs() []mx.Attrib { return []mx.Attrib{Class("stray")} }

// TestDocumentBodyChildrenStayChildren guards that BodyChildren are always
// rendered as the children of <body>, even when they could convert to
// attributes, as an mx.AttribProvider can.
func TestDocumentBodyChildrenStayChildren(t *testing.T) {
	got := renderDocument(t, &Document{BodyChildren: contentWithAttribs{}})
	if !strings.Contains(got, "<body><p>content</p></body>") {
		t.Errorf("BodyChildren not rendered as children\ngot: %s", got)
	}
}

// TestNewDocumentSortsAttribs guards that NewDocument sorts its body
// arguments as mx.NewElement does: attributes onto <body> rather than
// rendered as escaped text inside it, everything else as children.
func TestNewDocumentSortsAttribs(t *testing.T) {
	t.Run("attribs and children", func(t *testing.T) {
		got := renderDocument(t, NewDocument("T", Class("dark"), Hidden, nil, P("Hello"), "text"))
		const want = `<body class="dark" hidden="hidden"><p>Hello</p>text</body>`
		if !strings.Contains(got, want) {
			t.Errorf("missing %s\ngot: %s", want, got)
		}
	})

	t.Run("component providing attribs stays content", func(t *testing.T) {
		// A Component is a child even when it also converts to attributes,
		// as Render treats BodyChildren.
		got := renderDocument(t, NewDocument("T", contentWithAttribs{}))
		const want = `<body><p>content</p></body>`
		if !strings.Contains(got, want) {
			t.Errorf("missing %s\ngot: %s", want, got)
		}
	})

	t.Run("non-component attrib provider", func(t *testing.T) {
		got := renderDocument(t, NewDocument("T", mx.Attribs{"id": "main"}, "text"))
		const want = `<body id="main">text</body>`
		if !strings.Contains(got, want) {
			t.Errorf("missing %s\ngot: %s", want, got)
		}
	})

	// BodyChildren stays nil without children, as for a Document literal,
	// so a renderer that fills in a missing body (web.PageTypeRenderer)
	// still does.
	t.Run("no body arguments", func(t *testing.T) {
		d := NewDocument("T")
		if d.BodyAttribs != nil {
			t.Errorf("BodyAttribs = %v, want nil", d.BodyAttribs)
		}
		if d.BodyChildren != nil {
			t.Errorf("BodyChildren = %#v, want nil", d.BodyChildren)
		}
		got := renderDocument(t, d)
		if !strings.Contains(got, "<body></body>") {
			t.Errorf("want an empty <body>\ngot: %s", got)
		}
	})

	t.Run("attribs only", func(t *testing.T) {
		d := NewDocument("T", Class("dark"), nil)
		if d.BodyChildren != nil {
			t.Errorf("BodyChildren = %#v, want nil", d.BodyChildren)
		}
		got := renderDocument(t, d)
		if !strings.Contains(got, `<body class="dark"></body>`) {
			t.Errorf("want an empty <body class=\"dark\">\ngot: %s", got)
		}
	})

	// A child the program-wide mx.AsComponent converts to nil is dropped,
	// as NewElement drops it, so a body of only such children leaves
	// BodyChildren nil and web.PageTypeRenderer still fills in the content.
	// Value: protects=NewDocument dropping a nil AsComponent result; fails_when=the result is appended unconditionally and BodyChildren becomes a non-nil mx.Components{nil}; why_new=no test overrides AsComponent for NewDocument; seam=none; no_claim=AsComponents or NewElement filtering
	t.Run("child AsComponent drops", func(t *testing.T) {
		base := mx.AsComponent
		t.Cleanup(func() { mx.AsComponent = base })
		mx.AsComponent = func(c any) mx.Component {
			if _, ok := c.(int); ok {
				return nil
			}
			return base(c)
		}

		d := NewDocument("T", Class("dark"), 42)
		if d.BodyChildren != nil {
			t.Errorf("BodyChildren = %#v, want nil", d.BodyChildren)
		}
		got := renderDocument(t, d)
		if !strings.Contains(got, `<body class="dark"></body>`) {
			t.Errorf("want an empty <body class=\"dark\">\ngot: %s", got)
		}
	})
}

// ctxLangAttrib is a lang attribute whose value depends on the render
// context, so NewDocument cannot read it.
type ctxLangAttrib struct{}

func (ctxLangAttrib) AttribName() string { return "lang" }

func (ctxLangAttrib) AttribValue(context.Context) (string, error) { return "de", nil }

// TestNewDocumentLang guards that a lang attribute among the body
// arguments of NewDocument sets the language of the page on <html>, not
// on <body>, and that one it cannot move fails Render instead of
// landing on <body> unnoticed.
func TestNewDocumentLang(t *testing.T) {
	for _, tt := range []struct {
		name string
		arg  any
	}{
		{"Lang", Lang("de")},
		{"upper case name", mx.NewAttrib("LANG", "de")},
		{"ConstAttrib", mx.ConstAttrib("lang=de")},
		{"inside a slice", []mx.Attrib{Class("dark"), Lang("de")}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDocument("T", tt.arg, P("Hallo"))
			if d.Lang != "de" {
				t.Errorf("Lang = %q, want %q", d.Lang, "de")
			}
			got := renderDocument(t, d)
			if !strings.Contains(got, `<html lang="de">`) {
				t.Errorf("missing <html lang=\"de\">\ngot: %s", got)
			}
			body := got[strings.Index(got, "<body"):]
			if strings.Contains(strings.ToLower(body), "lang=") {
				t.Errorf("lang also on <body>\ngot: %s", got)
			}
		})
	}

	t.Run("invalid tag", func(t *testing.T) {
		var b strings.Builder
		err := NewDocument("T", Lang(`de"><x`)).Render(context.Background(), mx.NewCheckedWriter(&b))
		if err == nil || b.Len() != 0 {
			t.Errorf("want an error before any markup, got %v and %q", err, b.String())
		}
	})

	for _, tt := range []struct {
		name    string
		args    []any
		wantErr string
	}{
		{"second lang", []any{Lang("de"), Lang("en")}, "more than one lang"},
		{"context-dependent value", []any{ctxLangAttrib{}}, "set Document.Lang instead"},
		{"ErrAttrib kept", []any{mx.ErrAttribf("lang", "bad lang")}, "bad lang"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			err := NewDocument("T", tt.args...).Render(context.Background(), mx.NewCheckedWriter(&b))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Render error = %v, want one containing %q", err, tt.wantErr)
			}
			if strings.Contains(b.String(), "<body lang") {
				t.Errorf("lang reached <body>: %s", b.String())
			}
		})
	}
}

// TestNewDocumentAttribErrorHint guards that an error about BodyAttribs
// tells a caller who only passed NewDocument arguments where the field
// got its value, and that an error about HTMLAttribs, which NewDocument
// never fills, does not.
func TestNewDocumentAttribErrorHint(t *testing.T) {
	const hint = "NewDocument puts its attribute arguments there"
	for _, tt := range []struct {
		name string
		args []any
	}{
		{"duplicate differing in case", []any{Class("a"), mx.NewAttrib("CLASS", "b")}},
		{"invalid name", []any{mx.NewAttrib(`x"y`, "v")}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			err := NewDocument("T", tt.args...).Render(t.Context(), mx.NewCheckedWriter(&b))
			if err == nil || !strings.Contains(err.Error(), hint) {
				t.Errorf("Render error = %v, want the NewDocument hint", err)
			}
		})
	}

	t.Run("HTMLAttribs without hint", func(t *testing.T) {
		d := NewDocument("T")
		d.HTMLAttribs = []mx.Attrib{Class("a"), mx.NewAttrib("CLASS", "b")}
		var b strings.Builder
		err := d.Render(t.Context(), mx.NewCheckedWriter(&b))
		if err == nil || strings.Contains(err.Error(), hint) {
			t.Errorf("Render error = %v, want one without the NewDocument hint", err)
		}
	})
}

// TestNewDocumentSkipsNilAttrib guards that a nil mx.Attrib among the
// attributes of a body argument is left out, as a nil argument is, rather
// than kept in BodyAttribs where Render would refuse it.
func TestNewDocumentSkipsNilAttrib(t *testing.T) {
	d := NewDocument("T", []mx.Attrib{nil, Class("dark"), nil}, "text")
	if len(d.BodyAttribs) != 1 {
		t.Fatalf("BodyAttribs = %#v, want only the class", d.BodyAttribs)
	}
	got := renderDocument(t, d)
	if !strings.Contains(got, `<body class="dark">text</body>`) {
		t.Errorf("missing <body class=\"dark\">text</body>\ngot: %s", got)
	}
}

// TestDocumentBodyAttribsRefused guards the BodyAttribs that Render
// refuses. A nil entry, an invalid name and a name differing only in case
// are caught before anything is written; a failing value only when <body>
// is written, as on any element.
func TestDocumentBodyAttribsRefused(t *testing.T) {
	for _, tt := range []struct {
		name    string
		attribs []mx.Attrib
		// early is set where Render refuses before writing anything.
		early bool
	}{
		{"nil entry", []mx.Attrib{Class("a"), nil}, true},
		{"duplicate in other case", []mx.Attrib{Class("a"), mx.NewAttrib("CLASS", "b")}, true},
		{"name breaking the tag", []mx.Attrib{mx.NewAttrib("x onload=alert(1)", "")}, true},
		{"value error", []mx.Attrib{mx.ErrAttribf("class", "bad class")}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			err := (&Document{BodyAttribs: tt.attribs}).Render(context.Background(), mx.NewCheckedWriter(&b))
			if err == nil {
				t.Fatalf("want an error, got markup: %s", b.String())
			}
			if tt.early && b.Len() != 0 {
				t.Errorf("markup written before the error: %s", b.String())
			}
			if strings.Contains(b.String(), "onload") {
				t.Errorf("invalid name reached the markup: %s", b.String())
			}
		})
	}
}

// TestDocumentHTMLAttribs guards the extra attributes of the <html> root,
// which is written as Raw so the element writer does not close it.
func TestDocumentHTMLAttribs(t *testing.T) {
	t.Run("lang plus class", func(t *testing.T) {
		got := renderDocument(t, &Document{
			Lang:        "en",
			HTMLAttribs: []mx.Attrib{Class("no-js"), DataAttr("theme", "dark")},
		})
		const want = "\n" + `<html lang="en" class="no-js" data-theme="dark"><head>`
		if !strings.Contains(got, want) {
			t.Errorf("missing %q\ngot: %s", want, got)
		}
	})

	t.Run("without lang", func(t *testing.T) {
		got := renderDocument(t, &Document{HTMLAttribs: []mx.Attrib{Class("no-js")}})
		if !strings.Contains(got, `<html class="no-js">`) {
			t.Errorf("missing class on bare root\ngot: %s", got)
		}
	})

	t.Run("value escaped", func(t *testing.T) {
		got := renderDocument(t, &Document{HTMLAttribs: []mx.Attrib{Class(`a"><script>x`)}})
		if !strings.Contains(got, `<html class="a&quot;>&lt;script>x">`) {
			t.Errorf("class value not escaped\ngot: %s", got)
		}
		if strings.Contains(got, "<script>") {
			t.Errorf("class value broke out of the attribute: %s", got)
		}
	})

	// The root's attributes are written by the same writer code as those
	// of <body>, so a value comes out byte-for-byte the same on both,
	// including the line breaks and tabs the writer turns into references.
	t.Run("same escaping as body", func(t *testing.T) {
		attr := DataAttr("v", "a\"<b>'&\r\n\tc")
		got := renderDocument(t, &Document{HTMLAttribs: []mx.Attrib{attr}, BodyAttribs: []mx.Attrib{attr}})
		const value = `data-v="a&quot;&lt;b>'&amp;&#13;&#10;&#9;c"`
		if !strings.Contains(got, "<html "+value+">") || !strings.Contains(got, "<body "+value+">") {
			t.Errorf("want %s on <html> and <body>\ngot: %s", value, got)
		}
	})

	// The root follows the writer's quote style and indentation as every
	// other element does; indentation only applies to the elements inside.
	t.Run("writer configuration", func(t *testing.T) {
		doc := &Document{
			Lang:        "en",
			HTMLAttribs: []mx.Attrib{Class("no-js")},
			BodyAttribs: []mx.Attrib{Class("home")},
		}
		var single strings.Builder
		err := doc.Render(context.Background(), mx.NewCheckedWriter(&single).WithSingleQuoteAttribs())
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(single.String(), `<html lang='en' class='no-js'>`) ||
			!strings.Contains(single.String(), `<body class='home'>`) {
			t.Errorf("single-quote writer not followed\ngot: %s", single.String())
		}
		var indented strings.Builder
		err = doc.Render(context.Background(), mx.NewCheckedWriter(&indented).WithIndent("", "  "))
		if err != nil {
			t.Fatal(err)
		}
		const want = "<!DOCTYPE html>\n<html lang=\"en\" class=\"no-js\">\n<head>"
		if !strings.HasPrefix(indented.String(), want) {
			t.Errorf("indented root\ngot:  %q\nwant prefix: %q", indented.String(), want)
		}
	})

	// A Writer that is not a *mx.CheckedWriter has no configuration to copy,
	// so the root's attributes are written by a default CheckedWriter: names
	// checked, values escaped, double quotes.
	t.Run("other Writer", func(t *testing.T) {
		var b strings.Builder
		w := wrappedWriter{mx.NewCheckedWriter(&b).WithSingleQuoteAttribs()}
		doc := &Document{Lang: "en", HTMLAttribs: []mx.Attrib{Class(`a"b`)}}
		if err := doc.Render(context.Background(), w); err != nil {
			t.Fatal(err)
		}
		const want = "\n" + `<html lang="en" class="a&quot;b"><head>`
		if !strings.Contains(b.String(), want) {
			t.Errorf("missing %q\ngot: %s", want, b.String())
		}
	})

	// Empty and boolean values render as the element path renders them.
	t.Run("empty and boolean values", func(t *testing.T) {
		got := renderDocument(t, &Document{HTMLAttribs: []mx.Attrib{DataAttr("empty", ""), Hidden}})
		const want = `<html data-empty="" hidden="hidden">`
		if !strings.Contains(got, want) {
			t.Errorf("missing %s\ngot: %s", want, got)
		}
		elem := renderDocument(t, &Document{BodyAttribs: []mx.Attrib{DataAttr("empty", ""), Hidden}})
		if !strings.Contains(elem, `<body data-empty="" hidden="hidden">`) {
			t.Errorf("<html> and <body> render empty and boolean values differently\ngot: %s", elem)
		}
	})

	// Every one of these must abort Render before markup is written.
	for _, tt := range []struct {
		name    string
		attribs []mx.Attrib
	}{
		{"lang refused", []mx.Attrib{Lang("de")}},
		{"upper-case lang refused", []mx.Attrib{mx.NewAttrib("LANG", "de")}},
		{"duplicate", []mx.Attrib{Class("a"), Class("b")}},
		{"duplicate in other case", []mx.Attrib{Class("a"), mx.NewAttrib("CLASS", "b")}},
		{"nil entry", []mx.Attrib{nil}},
		{"empty name", []mx.Attrib{mx.NewAttrib("", "x")}},
		{"name breaking the tag", []mx.Attrib{mx.NewAttrib("x onload=alert(1)", "")}},
		{"name closing the tag", []mx.Attrib{mx.NewAttrib("x><script>", "")}},
		{"value error", []mx.Attrib{mx.ErrAttribf("class", "bad class")}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			err := (&Document{Lang: "en", HTMLAttribs: tt.attribs}).
				Render(context.Background(), mx.NewCheckedWriter(&b))
			if err == nil {
				t.Fatalf("want an error, got markup: %s", b.String())
			}
			if b.Len() != 0 {
				t.Errorf("markup written before the error: %s", b.String())
			}
		})
	}
}

// wrappedWriter is an mx.Writer that is not a *mx.CheckedWriter, as an
// application's own Writer would be.
type wrappedWriter struct{ *mx.CheckedWriter }
