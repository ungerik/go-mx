package pdf

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ungerik/go-mx/svg"
)

func TestResolveSVGBoxSize(t *testing.T) {
	tests := []struct {
		name         string
		attrs        map[string]string
		w, h         float64
		wantW, wantH float64
		wantErr      bool
	}{
		{name: "explicit box", attrs: nil, w: 100, h: 50, wantW: 100, wantH: 50},
		{
			name:  "height from width and viewBox aspect",
			attrs: map[string]string{"viewBox": "0 0 200 100"},
			w:     100, wantW: 100, wantH: 50,
		},
		{
			name:  "width from height and width/height attributes",
			attrs: map[string]string{"width": "40", "height": "20"},
			h:     10, wantW: 20, wantH: 10,
		},
		{
			name:  "intrinsic size",
			attrs: map[string]string{"width": "40", "height": "20"},
			wantW: 40, wantH: 20,
		},
		{
			name:  "intrinsic size from viewBox",
			attrs: map[string]string{"viewBox": "0 0 24 24"},
			wantW: 24, wantH: 24,
		},
		{
			name:  "percentage width is no intrinsic size",
			attrs: map[string]string{"width": "100%", "viewBox": "0 0 10 20"},
			wantW: 10, wantH: 20,
		},
		{name: "no size at all", attrs: nil, wantErr: true},
		{name: "only width given without aspect", attrs: nil, w: 100, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h, err := resolveSVGBoxSize(tt.attrs, tt.w, tt.h)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %g x %g", w, h)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if w != tt.wantW || h != tt.wantH {
				t.Errorf("got %g x %g, want %g x %g", w, h, tt.wantW, tt.wantH)
			}
		})
	}
}

// renderSVGToContent renders the component into an uncompressed one-page PDF
// and returns the whole PDF file as a string for content inspection.
func renderSVGToContent(t *testing.T, c Component) string {
	t.Helper()
	r := NewRendererA4Portrait()
	r.SetCompression(false)
	if err := c.Render(context.Background(), r); err != nil {
		t.Fatalf("render: %v", err)
	}
	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatalf("output: %v", err)
	}
	return buf.String()
}

func TestSVG_rendersShapes(t *testing.T) {
	content := renderSVGToContent(t, SVG(
		svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			svg.Rect(svg.X(10), svg.Y(10), svg.Width(30), svg.Height(20), svg.Fill("red")),
			svg.Circle(svg.CX(50), svg.CY(50), svg.R(10), svg.Fill("none"), svg.Stroke("#00f"), svg.StrokeWidth(2)),
			svg.Path(svg.D("M10 90 L90 90"), svg.Stroke("black")),
		),
		10, 10, 100, 100,
	))
	for _, want := range []string{
		"W n",                  // the viewport clip
		" cm",                  // the viewBox transform
		"1.000 0.000 0.000 rg", // red fill color
		"0.000 0.000 1.000 RG", // blue stroke color
		"\nf\n",                // fill op
		"\nS\n",                // stroke op
	} {
		if !strings.Contains(content, want) {
			t.Errorf("PDF content does not contain %q", want)
		}
	}
}

func TestSVG_fillRuleEvenOdd(t *testing.T) {
	content := renderSVGToContent(t, SVG(
		svg.SVG(
			svg.ViewBox(0, 0, 10, 10),
			svg.Path(svg.D("M0 0H10V10H0Z M2 2H8V8H2Z"), svg.FillRuleEvenodd),
			svg.Path(svg.D("M0 0H10V10H0Z M2 2H8V8H2Z"), svg.FillRuleEvenodd, svg.Stroke("black")),
		),
		0, 0, 10, 10,
	))
	if !strings.Contains(content, "f*") {
		t.Error("PDF content does not contain the even-odd fill operator f*")
	}
	if !strings.Contains(content, "B*") {
		t.Error("PDF content does not contain the even-odd fill+stroke operator B*")
	}
}

func TestSVG_dashedStroke(t *testing.T) {
	content := renderSVGToContent(t, SVG(
		svg.SVG(
			svg.ViewBox(0, 0, 10, 10),
			svg.Line(svg.X1(0), svg.Y1(5), svg.X2(10), svg.Y2(5),
				svg.Stroke("black"), svg.StrokeDashArray("2 1")),
		),
		0, 0, 10, 10,
	))
	if !strings.Contains(content, "] 0.00 d") {
		t.Error("PDF content does not contain a dash pattern")
	}
	// The dash state must be reset to solid after the shape.
	if !strings.Contains(content, "[] 0.00 d") {
		t.Error("PDF content does not reset the dash pattern to solid")
	}
}

func TestSVG_opacityUsesAlpha(t *testing.T) {
	content := renderSVGToContent(t, SVG(
		svg.SVG(
			svg.ViewBox(0, 0, 10, 10),
			svg.Rect(svg.Width(10), svg.Height(10), svg.Fill("black"), svg.FillOpacity(0.5)),
		),
		0, 0, 10, 10,
	))
	if !strings.Contains(content, " gs") {
		t.Error("PDF content does not select an alpha graphics state")
	}
}

func TestSVG_displayNoneRendersNothing(t *testing.T) {
	empty := renderSVGToContent(t, SVG(
		svg.SVG(svg.ViewBox(0, 0, 10, 10)),
		0, 0, 10, 10,
	))
	hidden := renderSVGToContent(t, SVG(
		svg.SVG(
			svg.ViewBox(0, 0, 10, 10),
			svg.Rect(svg.Width(10), svg.Height(10), svg.Display("none"), svg.Fill("red")),
			svg.G(svg.Display("none"), svg.Circle(svg.R(5), svg.Fill("blue"))),
		),
		0, 0, 10, 10,
	))
	if hidden != empty {
		t.Error("display:none content changed the PDF output")
	}
}

func TestSVG_skipsUnsupportedElements(t *testing.T) {
	// Unsupported elements must be skipped silently, not fail the render.
	content := renderSVGToContent(t, SVG(
		svg.SVG(
			svg.ViewBox(0, 0, 10, 10),
			svg.Defs(svg.LinearGradient(svg.ID("g"), svg.Stop(svg.Offset("0%")))),
			svg.Filter(svg.ID("f"), svg.FeGaussianBlur(svg.StdDeviation(2))),
			svg.Use(svg.Href("#nope")),
			svg.Title("ignored"),
			svg.Rect(svg.Width(10), svg.Height(10), svg.Fill("green")),
		),
		0, 0, 10, 10,
	))
	if !strings.Contains(content, "\nf\n") {
		t.Error("supported sibling of unsupported elements was not rendered")
	}
}

func TestSVG_errors(t *testing.T) {
	render := func(c Component) error {
		r := NewRendererA4Portrait()
		return c.Render(context.Background(), r)
	}
	tests := []struct {
		name string
		c    Component
	}{
		{name: "non-svg root", c: SVG(svg.Rect(svg.Width(10)), 0, 0, 10, 10)},
		{name: "bad fill color", c: SVG(svg.SVG(svg.ViewBox(0, 0, 9, 9), svg.Rect(svg.Width(9), svg.Height(9), svg.Fill("nope"))), 0, 0, 9, 9)},
		{name: "bad path data", c: SVG(svg.SVG(svg.ViewBox(0, 0, 9, 9), svg.Path(svg.D("M10 20 X"))), 0, 0, 9, 9)},
		{name: "bad transform", c: SVG(svg.SVG(svg.ViewBox(0, 0, 9, 9), svg.G(svg.Transform("bogus(1)"))), 0, 0, 9, 9)},
		{name: "negative rect size", c: SVG(svg.SVG(svg.ViewBox(0, 0, 9, 9), svg.Rect(svg.Width(-1), svg.Height(9))), 0, 0, 9, 9)},
		{name: "bad viewBox", c: SVG(svg.SVG(svg.Attrib("viewBox", "0 0 x 9")), 0, 0, 9, 9)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := render(tt.c); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}

	if err := render(SVG(nil, 0, 0, 10, 10)); err != nil {
		t.Errorf("nil root must render nothing, got error %v", err)
	}
}

func TestSVG_restoresGraphicsState(t *testing.T) {
	r := NewRendererA4Portrait()
	r.AddPage()
	r.SetFont(Courier, "B", 9)
	r.SetTextColor(1, 2, 3)
	r.SetFillColor(4, 5, 6)
	r.SetDrawColor(7, 8, 9)
	r.SetLineWidth(0.75)
	r.SetLineCapStyle("round")
	r.SetLineJoinStyle("bevel")
	r.SetXY(33, 44)

	err := SVG(
		svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			svg.Rect(svg.Width(50), svg.Height(50), svg.Fill("red"), svg.Stroke("blue"),
				svg.StrokeWidth(5), svg.StrokeLineCapSquare, svg.StrokeLineJoinRound,
				svg.StrokeDashArray("4 2"), svg.Opacity(0.5)),
			svg.Text(svg.X(10), svg.Y(10), svg.FontSize(30), svg.Fill("green"), "hello"),
		),
		10, 10, 100, 100,
	).Render(context.Background(), r)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if err = r.Error(); err != nil {
		t.Fatalf("renderer error: %v", err)
	}

	if family := r.GetFontFamily(); family != "courier" {
		t.Errorf("font family = %q, want courier", family)
	}
	if style := r.GetFontStyle(); style != "B" {
		t.Errorf("font style = %q, want B", style)
	}
	if pt, _ := r.GetFontSize(); pt != 9 {
		t.Errorf("font size = %g, want 9", pt)
	}
	if cr, cg, cb := r.GetTextColor(); cr != 1 || cg != 2 || cb != 3 {
		t.Errorf("text color = %d %d %d, want 1 2 3", cr, cg, cb)
	}
	if cr, cg, cb := r.GetFillColor(); cr != 4 || cg != 5 || cb != 6 {
		t.Errorf("fill color = %d %d %d, want 4 5 6", cr, cg, cb)
	}
	if cr, cg, cb := r.GetDrawColor(); cr != 7 || cg != 8 || cb != 9 {
		t.Errorf("draw color = %d %d %d, want 7 8 9", cr, cg, cb)
	}
	if lw := r.GetLineWidth(); lw != 0.75 {
		t.Errorf("line width = %g, want 0.75", lw)
	}
	if cap := r.GetLineCapStyle(); cap != "round" {
		t.Errorf("line cap = %q, want round", cap)
	}
	if join := r.GetLineJoinStyle(); join != "bevel" {
		t.Errorf("line join = %q, want bevel", join)
	}
	if x, y := r.GetXY(); x != 33 || y != 44 {
		t.Errorf("cursor = %g %g, want 33 44", x, y)
	}
	if alpha, _ := r.GetAlpha(); alpha != 1 {
		t.Errorf("alpha = %g, want 1", alpha)
	}
	if len(r.dashArray) != 0 {
		t.Errorf("dash array = %v, want solid", r.dashArray)
	}
}

func TestSVG_textAnchor(t *testing.T) {
	content := renderSVGToContent(t, SVG(
		svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			svg.Text(svg.X(50), svg.Y(20), svg.TextAnchorMiddle, "centered"),
			svg.Text(svg.X(50), svg.Y(40), "start"),
		),
		0, 0, 100, 100,
	))
	if !strings.Contains(content, "(centered)") && !strings.Contains(content, "centered") {
		t.Error("PDF content does not contain the text")
	}
}
