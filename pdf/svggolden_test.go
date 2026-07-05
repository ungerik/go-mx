package pdf

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/svg"
)

// svgGoldenCell frames one SVG scene for visual judgement: a thin silver
// border marks the target box (so clipping and alignment are visible), the
// SVG is rendered into it, and a caption is printed below.
func svgGoldenCell(x, y float64, title string, root *mx.Element) Components {
	const size = 58
	return Components{
		Save(DrawColor(Silver), LineWidth(0.2), Rect(x, y, size, size, Stroke)),
		SVG(root, x, y, size, size),
		Save(Font(Helvetica, StyleRegular, 7), TextColor(Gray50), TextAt(x, y+size+4, title)),
	}
}

// svgGoldenDocument renders a gallery of SVG scenes covering the supported
// feature set, for visual comparison against the committed reference PDF.
func svgGoldenDocument() *Document {
	cells := []struct {
		title string
		root  *mx.Element
	}{
		{"rect, rounded corners, fill+stroke", svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			svg.Rect(svg.X(6), svg.Y(6), svg.Width(42), svg.Height(30), svg.Fill("#4f46e5")),
			svg.Rect(svg.X(54), svg.Y(6), svg.Width(40), svg.Height(30), svg.RX(10),
				svg.Fill("none"), svg.Stroke("crimson"), svg.StrokeWidth(3)),
			svg.Rect(svg.X(6), svg.Y(46), svg.Width(88), svg.Height(48), svg.RX(8), svg.RY(16),
				svg.Fill("gold"), svg.Stroke("black"), svg.StrokeWidth(2)),
		)},
		{"circle and ellipse", svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			svg.Circle(svg.CX(30), svg.CY(30), svg.R(22), svg.Fill("teal")),
			svg.Ellipse(svg.CX(65), svg.CY(70), svg.RX(30), svg.RY(18),
				svg.Fill("none"), svg.Stroke("darkorange"), svg.StrokeWidth(4)),
			svg.Circle(svg.CX(70), svg.CY(28), svg.R(16),
				svg.Fill("lavender"), svg.Stroke("indigo"), svg.StrokeWidth(2), svg.StrokeDashArray("6 3")),
		)},
		{"lines: width, caps, dashes", svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			svg.Line(svg.X1(10), svg.Y1(15), svg.X2(90), svg.Y2(15), svg.Stroke("black"), svg.StrokeWidth(2)),
			svg.Line(svg.X1(10), svg.Y1(35), svg.X2(90), svg.Y2(35),
				svg.Stroke("steelblue"), svg.StrokeWidth(8), svg.StrokeLineCapButt),
			svg.Line(svg.X1(10), svg.Y1(55), svg.X2(90), svg.Y2(55),
				svg.Stroke("steelblue"), svg.StrokeWidth(8), svg.StrokeLineCapRound),
			svg.Line(svg.X1(10), svg.Y1(75), svg.X2(90), svg.Y2(75),
				svg.Stroke("seagreen"), svg.StrokeWidth(3), svg.StrokeDashArray("8 4 2 4")),
			svg.Polyline(svg.Points(10, 95, 30, 85, 50, 95, 70, 85, 90, 95),
				svg.Fill("none"), svg.Stroke("firebrick"), svg.StrokeWidth(2), svg.StrokeLineJoinRound),
		)},
		{"fill-rule: nonzero vs evenodd star", svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			svg.G(
				svg.Transform("translate(2 25) scale(0.5)"),
				svg.Path(svg.D("M50 5 L79 96 L2 40 L98 40 L21 96 Z"), svg.Fill("royalblue")),
			),
			svg.G(
				svg.Transform("translate(50 25) scale(0.5)"),
				svg.Path(svg.D("M50 5 L79 96 L2 40 L98 40 L21 96 Z"),
					svg.Fill("royalblue"), svg.FillRuleEvenodd, svg.Stroke("navy")),
			),
		)},
		{"paths: cubics, arcs, quadratics", svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			svg.Path(
				svg.D("M30 45 C14 31 6 21 11 13 C16 6 26 8 30 16 C34 8 44 6 49 13 C54 21 46 31 30 45 Z"),
				svg.Fill("crimson"),
			),
			svg.Path(svg.D("M75 30 L75 8 A22 22 0 0 1 94 41 Z"),
				svg.Fill("darkorange"), svg.Stroke("saddlebrown")),
			svg.Path(svg.D("M10 75 Q30 55 50 75 T90 75"),
				svg.Fill("none"), svg.Stroke("purple"), svg.StrokeWidth(3)),
			svg.Path(svg.D("M10 88 A15 8 20 1 0 40 88"),
				svg.Fill("none"), svg.Stroke("teal"), svg.StrokeWidth(2)),
		)},
		{"colors and opacity", svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			svg.Circle(svg.CX(38), svg.CY(35), svg.R(24), svg.Fill("rgb(220, 40, 40)"), svg.FillOpacity(0.6)),
			svg.Circle(svg.CX(62), svg.CY(35), svg.R(24), svg.Fill("hsl(210, 80%, 50%)"), svg.FillOpacity(0.6)),
			svg.Circle(svg.CX(50), svg.CY(55), svg.R(24), svg.Fill("#22c55e80")),
			svg.G(
				svg.Color("darkorange"),
				svg.Rect(svg.X(8), svg.Y(78), svg.Width(38), svg.Height(16), svg.Fill("currentColor")),
			),
			svg.Rect(svg.X(54), svg.Y(78), svg.Width(38), svg.Height(16),
				svg.Style("fill: slateblue; stroke: black; stroke-width: 1.5"), svg.Opacity(0.75)),
		)},
		{"transforms: rotate, scale, skew", svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			svg.G(
				svg.Rect(svg.X(35), svg.Y(10), svg.Width(30), svg.Height(14), svg.Fill("silver")),
				svg.Rect(svg.X(35), svg.Y(10), svg.Width(30), svg.Height(14), svg.Fill("skyblue"),
					svg.Transform("rotate(15 50 17)"), svg.FillOpacity(0.8)),
				svg.Rect(svg.X(35), svg.Y(10), svg.Width(30), svg.Height(14), svg.Fill("steelblue"),
					svg.Transform("rotate(30 50 17)"), svg.FillOpacity(0.8)),
			),
			svg.G(
				svg.Transform("translate(14 40) scale(1.4)"),
				svg.Circle(svg.CX(12), svg.CY(12), svg.R(10), svg.Fill("mediumseagreen")),
				svg.Circle(svg.CX(34), svg.CY(12), svg.R(10), svg.Fill("mediumseagreen"), svg.Transform("scale(0.6)")),
			),
			svg.Rect(svg.X(20), svg.Y(78), svg.Width(40), svg.Height(14),
				svg.Fill("orchid"), svg.Transform("translate(30 0) skewX(-20)")),
		)},
		{"nested viewports: meet, slice, none", svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			// Each nested viewport maps the same tall 10x20 scene into a
			// wide 26x20 box with a different preserveAspectRatio.
			svg.SVG(svg.X(4), svg.Y(30), svg.Width(26), svg.Height(20), svg.ViewBox(0, 0, 10, 20),
				svg.PreserveAspectRatio("xMidYMid meet"), svgGoldenViewportScene()),
			svg.SVG(svg.X(37), svg.Y(30), svg.Width(26), svg.Height(20), svg.ViewBox(0, 0, 10, 20),
				svg.PreserveAspectRatio("xMidYMid slice"), svgGoldenViewportScene()),
			svg.SVG(svg.X(70), svg.Y(30), svg.Width(26), svg.Height(20), svg.ViewBox(0, 0, 10, 20),
				svg.PreserveAspectRatio("none"), svgGoldenViewportScene()),
			// The circle overflows the root viewBox on purpose: it must be
			// clipped at the cell border.
			svg.Circle(svg.CX(100), svg.CY(85), svg.R(12), svg.Fill("tomato")),
		)},
		{"text: anchors, tspan, fonts", svg.SVG(
			svg.ViewBox(0, 0, 100, 100),
			svg.Line(svg.X1(50), svg.Y1(2), svg.X2(50), svg.Y2(58), svg.Stroke("silver")),
			svg.Text(svg.X(50), svg.Y(14), svg.FontSize(10), svg.TextAnchorStart, "start"),
			svg.Text(svg.X(50), svg.Y(30), svg.FontSize(10), svg.TextAnchorMiddle, "middle"),
			svg.Text(svg.X(50), svg.Y(46), svg.FontSize(10), svg.TextAnchorEnd, "end"),
			svg.Text(svg.X(6), svg.Y(64), svg.FontSize(9),
				"mix: ",
				svg.TSpan(svg.Fill("crimson"), svg.FontWeight("bold"), "bold red"),
				svg.TSpan(svg.FontStyle("italic"), " italic"),
			),
			svg.Text(svg.X(6), svg.Y(78), svg.FontSize(9), svg.FontFamily("serif"), "serif family"),
			svg.Text(svg.X(6), svg.Y(92), svg.FontSize(9), svg.FontFamily("monospace"), svg.Fill("darkgreen"), "monospace family"),
		)},
	}

	body := Components{
		Font(Helvetica, StyleBold, 14),
		Paragraph("go-mx/pdf SVG rendering reference"),
		Font(Helvetica, StyleRegular, 9),
		Paragraph("Each cell renders an svg-package element tree via pdf.SVG into the silver box."),
	}
	positions := [][2]float64{
		{15, 30}, {80, 30}, {145, 30},
		{15, 100}, {80, 100}, {145, 100},
		{15, 170}, {80, 170}, {145, 170},
	}
	for i, cell := range cells {
		body = append(body, svgGoldenCell(positions[i][0], positions[i][1], cell.title, cell.root))
	}
	return &Document{
		Title: "go-mx pdf SVG golden",
		Body:  body,
	}
}

// svgGoldenViewportScene is the content of the nested-viewport cells: a
// framed 10x20 scene whose distortion or clipping shows how the viewBox was
// mapped.
func svgGoldenViewportScene() mx.Components {
	return mx.Components{
		svg.Rect(svg.Width(10), svg.Height(20), svg.Fill("lightyellow"), svg.Stroke("goldenrod"), svg.StrokeWidth(1)),
		svg.Circle(svg.CX(5), svg.CY(6), svg.R(4), svg.Fill("cadetblue")),
		svg.Rect(svg.X(2), svg.Y(12), svg.Width(6), svg.Height(6), svg.Fill("indianred")),
	}
}

// renderSVGGolden renders svgGoldenDocument byte-for-byte reproducibly, like
// renderGolden does for the base golden document.
func renderSVGGolden() ([]byte, error) {
	doc := svgGoldenDocument()
	r := doc.NewRenderer()
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	r.SetCreationDate(fixed)
	r.SetModificationDate(fixed)
	r.SetCompression(false)
	r.SetCatalogSort(true)
	if err := doc.Render(context.Background(), r); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// TestSVGGoldenDeterministic guards the golden comparison itself: two renders
// must be byte-identical, otherwise TestSVGGoldenPDF would be flaky.
func TestSVGGoldenDeterministic(t *testing.T) {
	a, err := renderSVGGolden()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	b, err := renderSVGGolden()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("SVG golden render is not deterministic across runs")
	}
}

// TestSVGGoldenPDF renders the SVG gallery and compares it against the
// committed reference PDF. Regenerate the reference after an intended change
// (and judge it visually) with:
//
//	go test ./pdf -run TestSVGGoldenPDF -update
func TestSVGGoldenPDF(t *testing.T) {
	got, err := renderSVGGolden()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	path := filepath.Join("testdata", "svg_reference.pdf")

	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s (%d bytes)", path, len(got))
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden reference (create it with `go test ./pdf -run TestSVGGoldenPDF -update`): %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("rendered PDF differs from %s (%d vs %d bytes); if the change is intended, regenerate with `go test ./pdf -run TestSVGGoldenPDF -update`",
			path, len(got), len(want))
	}
}
