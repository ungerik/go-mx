package pdf

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"testing/iotest"
)

// makePNG encodes a small solid-color PNG in memory for the image tests.
func makePNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			img.Set(x, y, color.RGBA{R: 0x33, G: 0x66, B: 0x99, A: 0xff})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

// renderToBytes is a small helper that renders components through a fresh A4
// renderer and returns the encoded PDF.
func renderToBytes(t *testing.T, comps ...any) []byte {
	t.Helper()
	doc := NewDocument("Test", comps...)
	data, err := doc.Bytes(context.Background())
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Fatalf("output is not a PDF, starts with %q", data[:min(8, len(data))])
	}
	return data
}

func TestDocumentMinimal(t *testing.T) {
	data := renderToBytes(t, Paragraph("Hello, PDF!"))
	if len(data) < 100 {
		t.Fatalf("PDF unexpectedly small: %d bytes", len(data))
	}
	if !bytes.HasSuffix(bytes.TrimSpace(data), []byte("%%EOF")) {
		t.Errorf("PDF should end with %%%%EOF")
	}
}

func TestTextAndDrawingPrimitives(t *testing.T) {
	renderToBytes(t,
		Font(Helvetica, StyleBold, 18),
		Text("flowing text"),
		NewLine(),
		Cell(40, 8, "a cell"),
		CellFormat(40, 8, "bordered", BorderFull, LnNewline, AlignCenter, AlignMiddle, false),
		MultiCell(0, 6, strings.Repeat("word ", 50), BorderNone, AlignLeft, false),
		TextAt(20, 100, "absolute"),
		Line(10, 110, 100, 110),
		Rect(10, 115, 30, 20, Stroke),
		RoundedRect(50, 115, 30, 20, 3, FillStroke),
		Circle(120, 125, 10, FillShape),
		Ellipse(160, 125, 15, 8, 0, Stroke),
		Polygon(FillStroke, Pt(10, 150), Pt(40, 150), Pt(25, 175)),
	)
}

func TestMultiPage(t *testing.T) {
	doc := NewDocument("Multi",
		Page(Paragraph("page one")),
		Page(Paragraph("page two")),
		PageFormat(Landscape, A5, Paragraph("page three landscape A5")),
	)
	r := doc.NewRenderer()
	if err := doc.Render(context.Background(), r); err != nil {
		t.Fatalf("render: %v", err)
	}
	if got := r.PageCount(); got != 3 {
		t.Errorf("PageCount = %d, want 3", got)
	}
}

func TestSaveRestoresState(t *testing.T) {
	r := NewRendererA4Portrait()
	r.AddPage()
	r.SetTextColor(0, 0, 0)
	r.SetLineWidth(0.2)
	r.SetLineCapStyle(string(CapButt))
	r.SetLineJoinStyle(string(JoinMiter))

	comp := Save(
		TextColor(Red),
		LineWidth(2),
		LineCap(CapRound),
		LineJoin(JoinBevel),
		Font(Times, StyleItalic, 30),
		Text("styled"),
	)
	if err := comp.Render(context.Background(), r); err != nil {
		t.Fatalf("render: %v", err)
	}

	if rr, g, b := r.GetTextColor(); rr != 0 || g != 0 || b != 0 {
		t.Errorf("text color not restored: %d,%d,%d", rr, g, b)
	}
	if w := r.GetLineWidth(); w != 0.2 {
		t.Errorf("line width not restored: %v", w)
	}
	if cap := r.GetLineCapStyle(); cap != string(CapButt) {
		t.Errorf("line cap not restored: %q", cap)
	}
	if join := r.GetLineJoinStyle(); join != string(JoinMiter) {
		t.Errorf("line join not restored: %q", join)
	}
	// GetFontFamily reports the family lower-cased, as fpdf stores it.
	if fam := r.GetFontFamily(); fam != "helvetica" {
		t.Errorf("font family not restored: %q", fam)
	}
	if sizePt, _ := r.GetFontSize(); sizePt != DefaultFontSize {
		t.Errorf("font size not restored: %v", sizePt)
	}
}

func TestConditionalAndForEach(t *testing.T) {
	items := []string{"one", "two", "three"}
	renderToBytes(t,
		If(true, Paragraph("shown")),
		If(false, Paragraph("hidden")).Else(Paragraph("fallback")),
		ForEach(items, func(s string) Component {
			return Paragraph(s)
		}),
	)
}

func TestColorHex(t *testing.T) {
	tests := []struct {
		in   string
		want Color
	}{
		{"#ffffff", White},
		{"000000", Black},
		{"#f00", Red},
		{"#00ff00", Color{0, 255, 0}},
		{"abc", Color{0xaa, 0xbb, 0xcc}},
	}
	for _, tt := range tests {
		got, err := Hex(tt.in)
		if err != nil {
			t.Errorf("Hex(%q) error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Hex(%q) = %+v, want %+v", tt.in, got, tt.want)
		}
	}
	if _, err := Hex("#xyz"); err == nil {
		t.Error("Hex(#xyz) should fail")
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	doc := NewDocument("X", Paragraph("never rendered"))
	if err := doc.Output(ctx, &bytes.Buffer{}); err == nil {
		t.Error("expected context cancellation error")
	}
}

func TestImageFromMemory(t *testing.T) {
	pngData := makePNG(t)

	// ImageBytes embeds an in-memory image with no filesystem access.
	withImage := renderToBytes(t,
		Paragraph("image below"),
		ImageBytes("logo", ImagePNG, pngData, 20, 40, 30, 30),
	)
	// The same document without the image should be meaningfully smaller,
	// confirming the image bytes were actually embedded.
	withoutImage := renderToBytes(t, Paragraph("image below"))
	if len(withImage) <= len(withoutImage) {
		t.Errorf("expected embedded image to enlarge the PDF: with=%d without=%d",
			len(withImage), len(withoutImage))
	}

	// ImageReader must accept any io.Reader.
	renderToBytes(t, ImageReader("logo2", ImagePNG, bytes.NewReader(pngData), 0, 0, 20, 20))
}

func TestUTF8FontReaderError(t *testing.T) {
	// A failing reader passed to the in-memory font loader must record the
	// error on the renderer instead of panicking.
	r := NewRendererA4Portrait()
	r.LoadUTF8FontReader("Custom", StyleRegular, iotest.ErrReader(errors.New("boom")))
	if r.Error() == nil {
		t.Error("expected the read error to be recorded on the renderer")
	}
}

func TestHeaderFooterErrorSurfaces(t *testing.T) {
	// A component error from a Header or Footer has no fpdf return path, so it
	// must be folded into the renderer error and surface from Render.
	headerErr := errors.New("header boom")
	doc := NewDocument("X", Paragraph("body"))
	doc.Header = ComponentFunc(func(context.Context, *Renderer) error { return headerErr })
	if _, err := doc.Bytes(context.Background()); err == nil {
		t.Error("expected the header error to surface from the document render")
	}

	footerErr := errors.New("footer boom")
	doc = NewDocument("X", Paragraph("body"))
	doc.Footer = ComponentFunc(func(context.Context, *Renderer) error { return footerErr })
	if _, err := doc.Bytes(context.Background()); err == nil {
		t.Error("expected the footer error to surface from the document render")
	}
}

func TestAutoFirstPage(t *testing.T) {
	// A drawing primitive with no explicit Page must still produce a page.
	r := NewRendererA4Portrait()
	if r.PageNo() != 0 {
		t.Fatalf("expected no page before render, got %d", r.PageNo())
	}
	if err := Paragraph("auto page").Render(context.Background(), r); err != nil {
		t.Fatalf("render: %v", err)
	}
	if r.PageNo() != 1 {
		t.Errorf("expected page 1 after first primitive, got %d", r.PageNo())
	}
}
