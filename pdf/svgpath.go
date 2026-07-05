package pdf

import (
	"math"
	"strconv"

	"github.com/domonda/go-errs"
)

// SVG path-data support for the best-effort SVG renderer (see svg.go): a
// scanner for SVG number syntax, the <path> d-attribute parser, and the
// shape-to-path helpers. All output goes through the pathSink interface so
// the parser can be tested without a Renderer.

// pathSink receives a parsed path as move/line/cubic segments. The Renderer
// implements it via rendererPathSink; tests use a recording implementation.
type pathSink interface {
	moveTo(x, y float64)
	lineTo(x, y float64)
	cubicTo(cx0, cy0, cx1, cy1, x, y float64)
	closePath()
}

// rendererPathSink adapts the Renderer's path construction methods to pathSink.
type rendererPathSink struct {
	r *Renderer
}

func (s rendererPathSink) moveTo(x, y float64) { s.r.MoveTo(x, y) }
func (s rendererPathSink) lineTo(x, y float64) { s.r.LineTo(x, y) }
func (s rendererPathSink) closePath()          { s.r.ClosePath() }
func (s rendererPathSink) cubicTo(cx0, cy0, cx1, cy1, x, y float64) {
	s.r.CurveBezierCubicTo(cx0, cy0, cx1, cy1, x, y)
}

// svgScanner reads numbers and flags from SVG number-list syntax, where
// values are separated by whitespace and/or a single comma and a sign or a
// leading dot can start a new number without a separator ("1.5.5-2" is the
// three numbers 1.5, 0.5 and -2).
type svgScanner struct {
	s string
	i int
}

func (sc *svgScanner) skipSeparators() {
	comma := false
	for sc.i < len(sc.s) {
		switch sc.s[sc.i] {
		case ' ', '\t', '\n', '\r', '\f':
			sc.i++
		case ',':
			if comma {
				return
			}
			comma = true
			sc.i++
		default:
			return
		}
	}
}

func (sc *svgScanner) atEnd() bool {
	sc.skipSeparators()
	return sc.i >= len(sc.s)
}

// hasNumber reports whether the next token can start a number.
func (sc *svgScanner) hasNumber() bool {
	sc.skipSeparators()
	if sc.i >= len(sc.s) {
		return false
	}
	switch c := sc.s[sc.i]; {
	case c >= '0' && c <= '9', c == '+', c == '-', c == '.':
		return true
	}
	return false
}

// number scans one SVG number: sign, integer digits, fraction, exponent.
func (sc *svgScanner) number() (float64, error) {
	sc.skipSeparators()
	start := sc.i
	i := sc.i
	digits := func() (n int) {
		for i < len(sc.s) && sc.s[i] >= '0' && sc.s[i] <= '9' {
			i++
			n++
		}
		return n
	}
	if i < len(sc.s) && (sc.s[i] == '+' || sc.s[i] == '-') {
		i++
	}
	intDigits := digits()
	fracDigits := 0
	if i < len(sc.s) && sc.s[i] == '.' {
		i++
		fracDigits = digits()
	}
	if intDigits == 0 && fracDigits == 0 {
		return 0, errs.Errorf("expected number at %q", sc.s[start:])
	}
	if i < len(sc.s) && (sc.s[i] == 'e' || sc.s[i] == 'E') {
		j := i + 1
		if j < len(sc.s) && (sc.s[j] == '+' || sc.s[j] == '-') {
			j++
		}
		expStart := j
		for j < len(sc.s) && sc.s[j] >= '0' && sc.s[j] <= '9' {
			j++
		}
		if j > expStart {
			i = j
		}
	}
	v, err := strconv.ParseFloat(sc.s[start:i], 64)
	if err != nil {
		return 0, errs.Errorf("invalid number %q", sc.s[start:i])
	}
	sc.i = i
	return v, nil
}

// flag scans an SVG arc flag, a single '0' or '1' that needs no separator
// from the following value.
func (sc *svgScanner) flag() (bool, error) {
	sc.skipSeparators()
	if sc.i < len(sc.s) {
		switch sc.s[sc.i] {
		case '0':
			sc.i++
			return false, nil
		case '1':
			sc.i++
			return true, nil
		}
	}
	return false, errs.Errorf("expected arc flag at %q", sc.s[sc.i:])
}

// parseSVGNumberList parses a whitespace/comma separated list of numbers, the
// value syntax of attributes like points and the transform function arguments.
func parseSVGNumberList(s string) ([]float64, error) {
	sc := svgScanner{s: s}
	var values []float64
	for !sc.atEnd() {
		v, err := sc.number()
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}

// renderSVGPathData parses SVG path data (the d attribute of <path>) and
// emits it into sink as move/line/cubic segments: quadratic Béziers are
// promoted to cubics and elliptical arcs are approximated by cubics.
func renderSVGPathData(d string, sink pathSink) error {
	sc := svgScanner{s: d}
	var (
		cmd          byte    // current command letter
		started      bool    // an initial moveto has been seen
		x, y         float64 // current point
		startX       float64 // subpath start for Z
		startY       float64
		ctrlX, ctrlY float64 // last cubic/quadratic control point for S/T
		quadX, quadY float64
	)
	for !sc.atEnd() {
		c := sc.s[sc.i]
		switch {
		case (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z'):
			cmd = c
			sc.i++
		case cmd == 0:
			return errs.Errorf("SVG path data must start with a moveto command: %q", d)
		case cmd == 'Z' || cmd == 'z':
			// After a closepath only a new command may follow. A coordinate
			// here has no command to consume it, so without this guard the
			// current position never advances and the loop spins forever.
			return errs.Errorf("SVG path data has a coordinate after closepath (Z): %q", sc.s[sc.i:])
		case !sc.hasNumber():
			return errs.Errorf("unexpected character %q in SVG path data", string(sc.s[sc.i]))
		}
		// A path must begin with a moveto.
		if !started && cmd != 'M' && cmd != 'm' {
			return errs.Errorf("SVG path data must start with a moveto command: %q", d)
		}
		if cmd == 'Z' || cmd == 'z' {
			sink.closePath()
			x, y = startX, startY
			ctrlX, ctrlY, quadX, quadY = x, y, x, y
			continue
		}
		rel := cmd >= 'a'
		var rx, ry float64 // relative offsets
		if rel {
			rx, ry = x, y
		}
		switch cmd {
		case 'M', 'm':
			mx, err := sc.number()
			if err != nil {
				return err
			}
			my, err := sc.number()
			if err != nil {
				return err
			}
			x, y = rx+mx, ry+my
			startX, startY = x, y
			sink.moveTo(x, y)
			started = true
			// Further coordinate pairs are implicit lineto commands.
			if cmd == 'M' {
				cmd = 'L'
			} else {
				cmd = 'l'
			}
		case 'L', 'l':
			lx, err := sc.number()
			if err != nil {
				return err
			}
			ly, err := sc.number()
			if err != nil {
				return err
			}
			x, y = rx+lx, ry+ly
			sink.lineTo(x, y)
		case 'H', 'h':
			hx, err := sc.number()
			if err != nil {
				return err
			}
			x = rx + hx
			sink.lineTo(x, y)
		case 'V', 'v':
			vy, err := sc.number()
			if err != nil {
				return err
			}
			y = ry + vy
			sink.lineTo(x, y)
		case 'C', 'c', 'S', 's':
			var c1x, c1y float64
			if cmd == 'C' || cmd == 'c' {
				var err error
				if c1x, err = sc.number(); err != nil {
					return err
				}
				if c1y, err = sc.number(); err != nil {
					return err
				}
				c1x += rx
				c1y += ry
			} else {
				// Reflect the previous cubic control point; after any other
				// command ctrlX/ctrlY equal the current point, so the
				// reflection degenerates to the current point as specified.
				c1x, c1y = 2*x-ctrlX, 2*y-ctrlY
			}
			c2x, err := sc.number()
			if err != nil {
				return err
			}
			c2y, err := sc.number()
			if err != nil {
				return err
			}
			ex, err := sc.number()
			if err != nil {
				return err
			}
			ey, err := sc.number()
			if err != nil {
				return err
			}
			c2x, c2y, ex, ey = rx+c2x, ry+c2y, rx+ex, ry+ey
			sink.cubicTo(c1x, c1y, c2x, c2y, ex, ey)
			x, y = ex, ey
			ctrlX, ctrlY = c2x, c2y
			quadX, quadY = x, y
			continue
		case 'Q', 'q', 'T', 't':
			var qx, qy float64
			if cmd == 'Q' || cmd == 'q' {
				var err error
				if qx, err = sc.number(); err != nil {
					return err
				}
				if qy, err = sc.number(); err != nil {
					return err
				}
				qx += rx
				qy += ry
			} else {
				qx, qy = 2*x-quadX, 2*y-quadY
			}
			ex, err := sc.number()
			if err != nil {
				return err
			}
			ey, err := sc.number()
			if err != nil {
				return err
			}
			ex, ey = rx+ex, ry+ey
			// Promote the quadratic to the equivalent cubic.
			sink.cubicTo(x+2*(qx-x)/3, y+2*(qy-y)/3, ex+2*(qx-ex)/3, ey+2*(qy-ey)/3, ex, ey)
			x, y = ex, ey
			quadX, quadY = qx, qy
			ctrlX, ctrlY = x, y
			continue
		case 'A', 'a':
			arx, err := sc.number()
			if err != nil {
				return err
			}
			ary, err := sc.number()
			if err != nil {
				return err
			}
			rot, err := sc.number()
			if err != nil {
				return err
			}
			largeArc, err := sc.flag()
			if err != nil {
				return err
			}
			sweep, err := sc.flag()
			if err != nil {
				return err
			}
			ex, err := sc.number()
			if err != nil {
				return err
			}
			ey, err := sc.number()
			if err != nil {
				return err
			}
			ex, ey = rx+ex, ry+ey
			svgArcToCubics(sink, x, y, arx, ary, rot, largeArc, sweep, ex, ey)
			x, y = ex, ey
		default:
			return errs.Errorf("invalid SVG path command %q", string(cmd))
		}
		ctrlX, ctrlY = x, y
		quadX, quadY = x, y
	}
	return nil
}

// svgArcToCubics converts an SVG elliptical arc from (x1, y1) to (x2, y2)
// into cubic Bézier segments of at most 90° each, following the
// endpoint-to-center conversion of the SVG spec (appendix B.2.4).
func svgArcToCubics(sink pathSink, x1, y1, rx, ry, rotDeg float64, largeArc, sweep bool, x2, y2 float64) {
	if x1 == x2 && y1 == y2 {
		return
	}
	if rx == 0 || ry == 0 {
		sink.lineTo(x2, y2)
		return
	}
	rx, ry = math.Abs(rx), math.Abs(ry)
	sinPhi, cosPhi := math.Sincos(rotDeg * math.Pi / 180)

	// Step 1: half the vector between the endpoints, in the ellipse's frame.
	dx, dy := (x1-x2)/2, (y1-y2)/2
	x1p := cosPhi*dx + sinPhi*dy
	y1p := -sinPhi*dx + cosPhi*dy

	// Correct out-of-range radii.
	lambda := x1p*x1p/(rx*rx) + y1p*y1p/(ry*ry)
	if lambda > 1 {
		s := math.Sqrt(lambda)
		rx *= s
		ry *= s
	}

	// Step 2: center in the ellipse's frame.
	num := rx*rx*ry*ry - rx*rx*y1p*y1p - ry*ry*x1p*x1p
	den := rx*rx*y1p*y1p + ry*ry*x1p*x1p
	co := math.Sqrt(math.Max(0, num/den))
	if largeArc == sweep {
		co = -co
	}
	cxp := co * rx * y1p / ry
	cyp := -co * ry * x1p / rx

	// Step 3: center in the original frame.
	cx := cosPhi*cxp - sinPhi*cyp + (x1+x2)/2
	cy := sinPhi*cxp + cosPhi*cyp + (y1+y2)/2

	// Step 4: start angle and sweep extent.
	angle := func(ux, uy, vx, vy float64) float64 {
		a := math.Atan2(ux*vy-uy*vx, ux*vx+uy*vy)
		return a
	}
	theta1 := angle(1, 0, (x1p-cxp)/rx, (y1p-cyp)/ry)
	dTheta := angle((x1p-cxp)/rx, (y1p-cyp)/ry, (-x1p-cxp)/rx, (-y1p-cyp)/ry)
	if !sweep && dTheta > 0 {
		dTheta -= 2 * math.Pi
	} else if sweep && dTheta < 0 {
		dTheta += 2 * math.Pi
	}

	// Approximate with one cubic per segment of at most 90°.
	segments := int(math.Ceil(math.Abs(dTheta) / (math.Pi / 2)))
	if segments == 0 {
		return
	}
	delta := dTheta / float64(segments)
	// Control-point distance for a circular arc segment of angle delta.
	alpha := 4.0 / 3.0 * math.Tan(delta/4)

	pointAndDeriv := func(theta float64) (px, py, dpx, dpy float64) {
		sinT, cosT := math.Sincos(theta)
		px = cx + rx*cosT*cosPhi - ry*sinT*sinPhi
		py = cy + rx*cosT*sinPhi + ry*sinT*cosPhi
		dpx = -rx*sinT*cosPhi - ry*cosT*sinPhi
		dpy = -rx*sinT*sinPhi + ry*cosT*cosPhi
		return px, py, dpx, dpy
	}

	px0, py0, dx0, dy0 := pointAndDeriv(theta1)
	for i := 1; i <= segments; i++ {
		theta := theta1 + delta*float64(i)
		px1, py1, dx1, dy1 := pointAndDeriv(theta)
		if i == segments {
			// Land exactly on the endpoint despite rounding.
			px1, py1 = x2, y2
		}
		sink.cubicTo(
			px0+alpha*dx0, py0+alpha*dy0,
			px1-alpha*dx1, py1-alpha*dy1,
			px1, py1,
		)
		px0, py0, dx0, dy0 = px1, py1, dx1, dy1
	}
}

// bezierCircleKappa is the control-point factor that makes four cubic Béziers
// approximate a circle.
const bezierCircleKappa = 0.5522847498307935

// svgEllipsePath emits a full ellipse as four cubic segments, clockwise from
// the rightmost point (the direction does not matter for fill or stroke).
func svgEllipsePath(sink pathSink, cx, cy, rx, ry float64) {
	kx, ky := rx*bezierCircleKappa, ry*bezierCircleKappa
	sink.moveTo(cx+rx, cy)
	sink.cubicTo(cx+rx, cy+ky, cx+kx, cy+ry, cx, cy+ry)
	sink.cubicTo(cx-kx, cy+ry, cx-rx, cy+ky, cx-rx, cy)
	sink.cubicTo(cx-rx, cy-ky, cx-kx, cy-ry, cx, cy-ry)
	sink.cubicTo(cx+kx, cy-ry, cx+rx, cy-ky, cx+rx, cy)
	sink.closePath()
}

// svgRoundedRectPath emits a rectangle with the corners rounded by the radii
// rx, ry (already clamped to the half extents by the caller).
func svgRoundedRectPath(sink pathSink, x, y, w, h, rx, ry float64) {
	if rx <= 0 || ry <= 0 {
		sink.moveTo(x, y)
		sink.lineTo(x+w, y)
		sink.lineTo(x+w, y+h)
		sink.lineTo(x, y+h)
		sink.closePath()
		return
	}
	kx, ky := rx*bezierCircleKappa, ry*bezierCircleKappa
	sink.moveTo(x+rx, y)
	sink.lineTo(x+w-rx, y)
	sink.cubicTo(x+w-rx+kx, y, x+w, y+ry-ky, x+w, y+ry)
	sink.lineTo(x+w, y+h-ry)
	sink.cubicTo(x+w, y+h-ry+ky, x+w-rx+kx, y+h, x+w-rx, y+h)
	sink.lineTo(x+rx, y+h)
	sink.cubicTo(x+rx-kx, y+h, x, y+h-ry+ky, x, y+h-ry)
	sink.lineTo(x, y+ry)
	sink.cubicTo(x, y+ry-ky, x+rx-kx, y, x+rx, y)
	sink.closePath()
}
