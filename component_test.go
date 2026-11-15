package mx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

type stringerType struct{ s string }

func (s stringerType) String() string { return s.s }

type errorType struct{ msg string }

func (e errorType) Error() string { return e.msg }

type namedString string

// plainStruct implements neither Component, fmt.Stringer nor error, so it falls
// through to the fmt.Sprint default case.
type plainStruct struct{ X int }

func TestDefaultAsComponent_nilAndComponentPassthrough(t *testing.T) {
	if c := DefaultAsComponent(nil); c != nil {
		t.Errorf("DefaultAsComponent(nil) = %#v, want nil", c)
	}

	want := Text("kept")
	if got := DefaultAsComponent(want); got != want {
		t.Errorf("DefaultAsComponent(Component) = %#v, want same instance %#v", got, want)
	}
}

// Residual types (anything not matched by an explicit case) get their standard
// fmt.Sprint textual representation, rendered as an escaped Text node.
func TestDefaultAsComponent_stringifiesResidualTypes(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "string", in: "hello", want: "hello"},
		{name: "int", in: 42, want: "42"},
		{name: "negative int", in: -7, want: "-7"},
		{name: "uint", in: uint(7), want: "7"},
		{name: "bool", in: true, want: "true"},
		{name: "float", in: 3.5, want: "3.5"},
		// A named string type is not the builtin string, so it falls through
		// to pretty.Sprint, which backtick-quotes it.
		{name: "named string", in: namedString("nm"), want: "`nm`"},
		{name: "fmt.Stringer", in: stringerType{"S"}, want: "S"},
		{name: "error", in: errorType{"boom"}, want: "boom"},
		// pretty.Sprint tags the type and names fields, and dereferences the
		// pointer (no leading & to escape, unlike fmt.Sprint).
		{name: "struct", in: plainStruct{1}, want: "plainStruct{X:1}"},
		{name: "pointer to struct", in: &plainStruct{1}, want: "plainStruct{X:1}"},
		{name: "slice", in: []int{1, 2, 3}, want: "[1,2,3]"},
		{name: "map", in: map[string]int{"a": 1}, want: "{`a`:1}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := DefaultAsComponent(tt.in)
			if c == nil {
				t.Fatalf("DefaultAsComponent(%#v) = nil", tt.in)
			}
			if got := renderToString(t, c); got != tt.want {
				t.Errorf("render = %q, want %q", got, tt.want)
			}
		})
	}
}

// A stringified residual value must never be able to inject markup: the Writer
// escapes the Text node it is wrapped in. This holds for every markup target,
// since html and svg render through the same CheckedWriter / TextEscaper.
func TestDefaultAsComponent_residualTextIsEscaped(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "stringer with tags", in: stringerType{`<script>alert(1)</script>`}, want: `&lt;script&gt;alert(1)&lt;/script&gt;`},
		{name: "stringer with entities", in: stringerType{`a & b`}, want: `a &amp; b`},
		{name: "error with quotes", in: errorType{`he said "hi" & 'bye'`}, want: `he said &quot;hi&quot; &amp; &apos;bye&apos;`},
		{name: "struct with markup field", in: struct{ S string }{S: "<b>"}, want: "{S:`&lt;b&gt;`}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderToString(t, DefaultAsComponent(tt.in)); got != tt.want {
				t.Errorf("render = %q, want %q (markup not escaped)", got, tt.want)
			}
		})
	}
}

// Escaping a residual value inside an element must produce the same output as
// escaping the equivalent string, i.e. residual stringification is just text.
func TestDefaultAsComponent_residualMatchesEquivalentString(t *testing.T) {
	withResidual := renderToString(t, NewElement("div", stringerType{"<x>"}))
	withString := renderToString(t, NewElement("div", "<x>"))
	if withResidual != withString {
		t.Errorf("residual %q != string %q", withResidual, withString)
	}
}

// A component that fails to render, or panics, answers a 500 and is
// reported, a panic with the debug stack its error carries. A render the
// client canceled is not reported.
func TestComponentHTTPHandler_ReportsErrorsAndPanics(t *testing.T) {
	rep := recordReports(t)
	var renderErr error
	failing := ComponentHTTPHandler(ComponentFunc(func(context.Context, Writer) error { return renderErr }), DefaultWriterFactory, nil)
	renderErr = errors.New("template: pq: connection refused")
	if rec := serveReportRequest(failing, http.MethodGet, nil); rec.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", rec.Code)
	}
	rep.checkReported(t, true, renderErr)

	rep.reset()
	renderErr = fmt.Errorf("rendering: %w", context.Canceled)
	serveCancelledReportRequest(failing, http.MethodGet, nil)
	rep.checkReported(t, false)

	panicking := ComponentHTTPHandler(ComponentFunc(func(context.Context, Writer) error { panic("nil map in the sidebar") }), DefaultWriterFactory, nil)
	if rec := serveReportRequest(panicking, http.MethodGet, nil); rec.Code != http.StatusInternalServerError {
		t.Errorf("status %d, want 500", rec.Code)
	}
	checkReportedPanic(t, rep, "nil map in the sidebar")
}

// checkReportedPanic fails the test unless rep holds one report: the
// recovered panic value with the debug stack of the panic.
func checkReportedPanic(t *testing.T, rep *reports, value string) {
	t.Helper()
	texts := rep.texts()
	if len(texts) != 1 || !strings.HasPrefix(texts[0], value+"\n") || !strings.Contains(texts[0], "goroutine ") {
		t.Errorf("reported %q, want the panic %q with its stack", texts, value)
	}
}

// enumLikeAttrib is an Attrib whose String returns only its value, as the
// go-enum generated attribute types do.
type enumLikeAttrib string

func (a enumLikeAttrib) AttribName() string                          { return "dir" }
func (a enumLikeAttrib) AttribValue(context.Context) (string, error) { return string(a), nil }
func (a enumLikeAttrib) String() string                              { return string(a) }

// componentAttrib is both a Component and an Attrib.
type componentAttrib struct{}

func (componentAttrib) Render(ctx context.Context, w Writer) error {
	return Text("content").Render(ctx, w)
}
func (componentAttrib) AttribName() string                          { return "class" }
func (componentAttrib) AttribValue(context.Context) (string, error) { return "stray", nil }

// ctxAttribKey is the context key ctxAttrib reads its value from.
type ctxAttribKey struct{}

// ctxAttrib is an Attrib whose value comes from the render context.
type ctxAttrib struct{}

func (ctxAttrib) AttribName() string { return "data-user" }
func (ctxAttrib) AttribValue(ctx context.Context) (string, error) {
	value, _ := ctx.Value(ctxAttribKey{}).(string)
	return value, nil
}

// TestDefaultAsComponentAttrib guards that every Attrib passed as a child
// renders as name='value' text, escaped once, name included, not depending
// on whether its type has a String method or what that returns (an enum
// type's String returns only the value), and that a Component which is also
// an Attrib stays a Component.
func TestDefaultAsComponentAttrib(t *testing.T) {
	for _, tt := range []struct {
		name string
		c    any
		want string
	}{
		{"Attribute", NewAttrib("class", "x"), "class='x'"},
		{"ConstAttrib", ConstAttrib("type=text"), "type='text'"},
		{"String returns only the value", enumLikeAttrib("rtl"), "dir='rtl'"},
		{"no String method", UniqueID(), "id='"},
		{"escaped once", ConstAttrib("title=it's <b>\na"), "title='it&apos;s &lt;b>&#10;a'"},
		{"name escaped", NewAttrib("<img src=x onerror=alert(1)>", "v"), "&lt;img src=x onerror=alert(1)>='v'"},
		{"Component and Attrib", componentAttrib{}, "content"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := renderToString(t, DefaultAsComponent(tt.c))
			if !strings.HasPrefix(got, tt.want) {
				t.Errorf("got %q, want prefix %q", got, tt.want)
			}
		})
	}
}

// TestAsComponentsAttribNameNoInjection guards that an attribute whose name
// holds markup, passed as a child through AsComponents, cannot inject that
// markup: the name is escaped like the value.
func TestAsComponentsAttribNameNoInjection(t *testing.T) {
	got := renderToString(t, AsComponents(NewAttrib("<b>", "v")))
	if want := "&lt;b>='v'"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestDefaultAsComponentAttribShapes guards that every attribute shape
// AsAttribs accepts, not only a single Attrib, renders as the same escaped
// name='value' text when passed as a child: a slice is not pretty-printed and
// an Attribs map, whose String method escapes already, is not escaped twice.
// A nil Attrib among them is left out.
func TestDefaultAsComponentAttribShapes(t *testing.T) {
	for _, tt := range []struct {
		name string
		c    any
		want string
	}{
		{"[]Attrib", []Attrib{NewAttrib("title", "it's"), nil, NewAttrib("class", "x")}, "title='it&apos;s' class='x'"},
		{"[]Attribute", []Attribute{{Name: "title", Value: "it's"}}, "title='it&apos;s'"},
		{"Attribs map", Attribs{"title": "it's"}, "title='it&apos;s'"},
		{"map[string]string", map[string]string{"title": "it's"}, "title='it&apos;s'"},
		{"func() Attrib", func() Attrib { return NewAttrib("title", "it's") }, "title='it&apos;s'"},
		{"empty slice", []Attrib{}, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := renderToString(t, DefaultAsComponent(tt.c)); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestDefaultAsComponentAttribRender guards that an Attrib passed as a child
// is valued when it is rendered, as an attribute of an element is: with the
// render context, and with a failing value aborting the render with its
// error instead of being shown on the page.
func TestDefaultAsComponentAttribRender(t *testing.T) {
	t.Run("render context", func(t *testing.T) {
		var b strings.Builder
		ctx := context.WithValue(t.Context(), ctxAttribKey{}, "ann")
		err := DefaultAsComponent(ctxAttrib{}).Render(ctx, NewCheckedWriter(&b))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := b.String(), "data-user='ann'"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("failing value", func(t *testing.T) {
		var b strings.Builder
		err := AsComponents("a", ErrAttribf("class", "bad class")).Render(t.Context(), NewCheckedWriter(&b))
		if err == nil || !strings.Contains(err.Error(), "bad class") {
			t.Fatalf("Render error = %v, want the ErrAttrib error", err)
		}
		if strings.Contains(b.String(), "bad class") {
			t.Errorf("error text reached the page: %q", b.String())
		}
	})
}
