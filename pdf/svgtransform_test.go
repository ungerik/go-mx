package pdf

import (
	"math"
	"testing"
)

// applySVGMatrix maps a point through the matrix like an SVG renderer.
func applySVGMatrix(m svgMatrix, x, y float64) (float64, float64) {
	return m.a*x + m.c*y + m.e, m.b*x + m.d*y + m.f
}

func TestParseSVGTransform(t *testing.T) {
	tests := []struct {
		name         string
		transform    string
		x, y         float64 // input point
		wantX, wantY float64
	}{
		{name: "translate", transform: "translate(10 20)", x: 1, y: 2, wantX: 11, wantY: 22},
		{name: "translate single value", transform: "translate(10)", x: 1, y: 2, wantX: 11, wantY: 2},
		{name: "scale", transform: "scale(2 3)", x: 1, y: 2, wantX: 2, wantY: 6},
		{name: "uniform scale", transform: "scale(2)", x: 1, y: 2, wantX: 2, wantY: 4},
		{name: "rotate 90", transform: "rotate(90)", x: 1, y: 0, wantX: 0, wantY: 1},
		{name: "rotate about center", transform: "rotate(180 10 10)", x: 0, y: 0, wantX: 20, wantY: 20},
		{name: "skewX 45", transform: "skewX(45)", x: 0, y: 1, wantX: 1, wantY: 1},
		{name: "skewY 45", transform: "skewY(45)", x: 1, y: 0, wantX: 1, wantY: 1},
		{name: "matrix", transform: "matrix(1 0 0 1 5 6)", x: 1, y: 1, wantX: 6, wantY: 7},
		{
			// The list composes left to right: translate first, then the
			// translated content is scaled from the (already translated)
			// origin — i.e. p' = translate(scale(p)).
			name: "list composition", transform: "translate(10 0) scale(2)",
			x: 1, y: 1, wantX: 12, wantY: 2,
		},
		{
			name: "comma separated arguments", transform: "translate(10,20),scale(2,2)",
			x: 1, y: 1, wantX: 12, wantY: 22,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := parseSVGTransform(tt.transform)
			if err != nil {
				t.Fatalf("parseSVGTransform(%q): %v", tt.transform, err)
			}
			gotX, gotY := applySVGMatrix(m, tt.x, tt.y)
			if math.Abs(gotX-tt.wantX) > 1e-9 || math.Abs(gotY-tt.wantY) > 1e-9 {
				t.Errorf("parseSVGTransform(%q) maps (%g, %g) to (%g, %g), want (%g, %g)",
					tt.transform, tt.x, tt.y, gotX, gotY, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestParseSVGTransform_errors(t *testing.T) {
	for _, s := range []string{
		"rotate(45",       // missing closing parenthesis
		"translate 10 20", // missing parentheses
		"scale(1 2 3)",    // wrong argument count
		"frobnicate(1)",   // unknown function
		"matrix(1 2 3)",   // wrong argument count
		"translate(a)",    // not a number
	} {
		if _, err := parseSVGTransform(s); err == nil {
			t.Errorf("parseSVGTransform(%q): expected error, got nil", s)
		}
	}
}

// TestTransformSVGConjugation verifies that the SVG-space to PDF-space matrix
// conjugation in transformSVG is consistent: the conjugate of a product must
// equal the product of the conjugates, which is what makes nested SVG
// transforms compose correctly as nested PDF cm operators.
func TestTransformSVGConjugation(t *testing.T) {
	conj := func(m svgMatrix, h, k float64) TransformMatrix {
		return TransformMatrix{
			A: m.a, B: -m.b, C: -m.c, D: m.d,
			E: (m.c*h + m.e) * k,
			F: (h - m.d*h - m.f) * k,
		}
	}
	mulPDF := func(m, n TransformMatrix) TransformMatrix {
		// PDF matrices are row vectors: [a b; c d; e f] with p' = p·M.
		return TransformMatrix{
			A: n.A*m.A + n.B*m.C,
			B: n.A*m.B + n.B*m.D,
			C: n.C*m.A + n.D*m.C,
			D: n.C*m.B + n.D*m.D,
			E: n.E*m.A + n.F*m.C + m.E,
			F: n.E*m.B + n.F*m.D + m.F,
		}
	}
	const h, k = 297.0, 72.0 / 25.4
	a := svgTranslate(10, 20).mul(svgRotate(30))
	b := svgScale(2, 0.5).mul(svgTranslate(-3, 7))

	left := conj(a.mul(b), h, k)
	right := mulPDF(conj(a, h, k), conj(b, h, k))
	for i, pair := range [][2]float64{
		{left.A, right.A}, {left.B, right.B}, {left.C, right.C},
		{left.D, right.D}, {left.E, right.E}, {left.F, right.F},
	} {
		if math.Abs(pair[0]-pair[1]) > 1e-9 {
			t.Fatalf("component %d: conj(a·b) = %v but conj(a)·conj(b) = %v", i, left, right)
		}
	}
}
