package pdf

import (
	"fmt"
	"math"
	"strings"
)

// SVG affine transform support for the best-effort SVG renderer (see svg.go):
// the svgMatrix type, the transform-list parser, and the conversion from SVG
// user space (top-left origin, y down, document units) to the PDF coordinate
// space used by Renderer.Transform (bottom-left origin, y up, scaled by k).

// svgMatrix is an affine transform in SVG notation: a point (x, y) maps to
// (a*x + c*y + e, b*x + d*y + f), i.e. the matrix [a c e; b d f; 0 0 1].
type svgMatrix struct {
	a, b, c, d, e, f float64
}

var svgIdentity = svgMatrix{a: 1, d: 1}

// mul returns the product m·n, the transform that applies n first and then m —
// the composition order of an SVG transform list read left to right.
func (m svgMatrix) mul(n svgMatrix) svgMatrix {
	return svgMatrix{
		a: m.a*n.a + m.c*n.b,
		b: m.b*n.a + m.d*n.b,
		c: m.a*n.c + m.c*n.d,
		d: m.b*n.c + m.d*n.d,
		e: m.a*n.e + m.c*n.f + m.e,
		f: m.b*n.e + m.d*n.f + m.f,
	}
}

func svgTranslate(tx, ty float64) svgMatrix {
	return svgMatrix{a: 1, d: 1, e: tx, f: ty}
}

func svgScale(sx, sy float64) svgMatrix {
	return svgMatrix{a: sx, d: sy}
}

func svgRotate(deg float64) svgMatrix {
	sin, cos := math.Sincos(deg * math.Pi / 180)
	return svgMatrix{a: cos, b: sin, c: -sin, d: cos}
}

// parseSVGTransform parses an SVG transform list such as
// "translate(10 20) rotate(45) scale(2)" into a single matrix, composing the
// functions left to right. Supported functions: matrix, translate, scale,
// rotate (with optional center), skewX and skewY.
func parseSVGTransform(s string) (svgMatrix, error) {
	m := svgIdentity
	rest := strings.TrimSpace(s)
	for rest != "" {
		open := strings.IndexByte(rest, '(')
		if open < 0 {
			return m, fmt.Errorf("invalid SVG transform %q", s)
		}
		name := strings.TrimSpace(rest[:open])
		closing := strings.IndexByte(rest[open:], ')')
		if closing < 0 {
			return m, fmt.Errorf("invalid SVG transform %q", s)
		}
		args, err := parseSVGNumberList(rest[open+1 : open+closing])
		if err != nil {
			return m, fmt.Errorf("invalid SVG transform %q: %w", s, err)
		}
		rest = strings.TrimSpace(rest[open+closing+1:])
		rest = strings.TrimSpace(strings.TrimPrefix(rest, ","))

		var fn svgMatrix
		switch {
		case name == "matrix" && len(args) == 6:
			fn = svgMatrix{args[0], args[1], args[2], args[3], args[4], args[5]}
		case name == "translate" && len(args) == 1:
			fn = svgTranslate(args[0], 0)
		case name == "translate" && len(args) == 2:
			fn = svgTranslate(args[0], args[1])
		case name == "scale" && len(args) == 1:
			fn = svgScale(args[0], args[0])
		case name == "scale" && len(args) == 2:
			fn = svgScale(args[0], args[1])
		case name == "rotate" && len(args) == 1:
			fn = svgRotate(args[0])
		case name == "rotate" && len(args) == 3:
			fn = svgTranslate(args[1], args[2]).
				mul(svgRotate(args[0])).
				mul(svgTranslate(-args[1], -args[2]))
		case name == "skewX" && len(args) == 1:
			fn = svgMatrix{a: 1, d: 1, c: math.Tan(args[0] * math.Pi / 180)}
		case name == "skewY" && len(args) == 1:
			fn = svgMatrix{a: 1, d: 1, b: math.Tan(args[0] * math.Pi / 180)}
		default:
			return m, fmt.Errorf("invalid SVG transform function %q with %d arguments", name, len(args))
		}
		m = m.mul(fn)
	}
	return m, nil
}

// transformSVG applies m, given in SVG user space (top-left origin, y down,
// document units), as a PDF transform. Renderer methods convert a point
// (x, y) in document units to the PDF device point S(x, y) = (x*k, (h-y)*k),
// so the equivalent device-space matrix is the conjugate S·m·S⁻¹; because
// conjugation distributes over products, nested calls compose exactly like
// nested SVG transforms. Must be called inside TransformBegin/TransformEnd.
func (r *Renderer) transformSVG(m svgMatrix) {
	r.Transform(TransformMatrix{
		A: m.a,
		B: -m.b,
		C: -m.c,
		D: m.d,
		E: (m.c*r.h + m.e) * r.k,
		F: (r.h - m.d*r.h - m.f) * r.k,
	})
}
