package pdf

import "github.com/go-pdf/fpdf"

type Renderer struct {
	*fpdf.Fpdf

	translate func(string) string
}

func NewRendererA4Portrait() *Renderer {
	p := fpdf.New(
		fpdf.OrientationPortrait,
		fpdf.UnitMillimeter,
		fpdf.PageSizeA4,
		"",
	)
	return &Renderer{
		Fpdf:      p,
		translate: p.UnicodeTranslatorFromDescriptor(""),
	}
}

func (w *Renderer) Str(s string) string {
	return w.translate(s)
}
