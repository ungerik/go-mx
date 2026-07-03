package pdf

import (
	"testing"
)

func TestParseSVGColor(t *testing.T) {
	tests := []struct {
		in        string
		want      Color
		wantAlpha float64
	}{
		{in: "#f00", want: Color{255, 0, 0}, wantAlpha: 1},
		{in: "#ff0000", want: Color{255, 0, 0}, wantAlpha: 1},
		{in: "#ff000080", want: Color{255, 0, 0}, wantAlpha: 128.0 / 255},
		{in: "#f008", want: Color{255, 0, 0}, wantAlpha: 136.0 / 255},
		{in: "rgb(1, 2, 3)", want: Color{1, 2, 3}, wantAlpha: 1},
		{in: "rgb(100%, 0%, 50%)", want: Color{255, 0, 128}, wantAlpha: 1},
		{in: "rgba(10, 20, 30, 0.5)", want: Color{10, 20, 30}, wantAlpha: 0.5},
		{in: "rgb(10 20 30 / 25%)", want: Color{10, 20, 30}, wantAlpha: 0.25},
		{in: "hsl(120, 100%, 50%)", want: Color{0, 255, 0}, wantAlpha: 1},
		{in: "hsl(0, 100%, 50%)", want: Color{255, 0, 0}, wantAlpha: 1},
		{in: "hsla(240, 100%, 50%, 0.5)", want: Color{0, 0, 255}, wantAlpha: 0.5},
		{in: "tomato", want: Color{255, 99, 71}, wantAlpha: 1},
		{in: "RebeccaPurple", want: Color{102, 51, 153}, wantAlpha: 1},
		{in: " grey ", want: Color{128, 128, 128}, wantAlpha: 1},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			c, alpha, err := parseSVGColor(tt.in)
			if err != nil {
				t.Fatalf("parseSVGColor(%q): %v", tt.in, err)
			}
			if c != tt.want {
				t.Errorf("parseSVGColor(%q) = %v, want %v", tt.in, c, tt.want)
			}
			if diff := alpha - tt.wantAlpha; diff > 1e-9 || diff < -1e-9 {
				t.Errorf("parseSVGColor(%q) alpha = %g, want %g", tt.in, alpha, tt.wantAlpha)
			}
		})
	}

	for _, in := range []string{"", "#12345", "#xyz", "rgb(1,2)", "hsl(a,b,c)", "notacolor", "url(#gradient)"} {
		if _, _, err := parseSVGColor(in); err == nil {
			t.Errorf("parseSVGColor(%q): expected error, got nil", in)
		}
	}
}

func TestParseSVGPaint(t *testing.T) {
	if p, err := parseSVGPaint("none"); err != nil || !p.none {
		t.Errorf("parseSVGPaint(none) = %+v, %v", p, err)
	}
	if p, err := parseSVGPaint("transparent"); err != nil || !p.none {
		t.Errorf("parseSVGPaint(transparent) = %+v, %v", p, err)
	}
	if p, err := parseSVGPaint("currentColor"); err != nil || !p.currentColor {
		t.Errorf("parseSVGPaint(currentColor) = %+v, %v", p, err)
	}
	// Unsupported paint servers degrade to no paint, or to the fallback color.
	if p, err := parseSVGPaint("url(#gradient)"); err != nil || !p.none {
		t.Errorf("parseSVGPaint(url) = %+v, %v", p, err)
	}
	p, err := parseSVGPaint("url(#gradient) red")
	if err != nil || p.none || p.color != (Color{255, 0, 0}) {
		t.Errorf("parseSVGPaint(url with fallback) = %+v, %v", p, err)
	}
	if _, err = parseSVGPaint("url(#gradient"); err == nil {
		t.Error("parseSVGPaint(unclosed url): expected error, got nil")
	}
	if _, err = parseSVGPaint("bogus"); err == nil {
		t.Error("parseSVGPaint(bogus): expected error, got nil")
	}
}
