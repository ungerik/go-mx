package pdf

import "github.com/domonda/go-errs"

// Color is an 8-bit-per-channel RGB color, the form fpdf uses for the draw,
// fill and text colors. fpdf has no notion of alpha in colors; use the
// renderer's SetAlpha for transparency.
type Color struct {
	R, G, B int
}

// RGB builds a [Color] from red, green and blue components in the range 0–255.
func RGB(r, g, b int) Color {
	return Color{R: r, G: g, B: b}
}

// Gray builds a gray [Color] with the given 0–255 value on all three channels.
func Gray(v int) Color {
	return Color{R: v, G: v, B: v}
}

// Hex parses a CSS-style hex color: "#rgb", "#rrggbb" or the same without the
// leading '#'. An invalid string returns black and a non-nil error, so callers
// that want strictness can check it; the State helpers ignore the error and
// fall back to black.
func Hex(s string) (Color, error) {
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	hexVal := func(b byte) (int, bool) {
		switch {
		case b >= '0' && b <= '9':
			return int(b - '0'), true
		case b >= 'a' && b <= 'f':
			return int(b-'a') + 10, true
		case b >= 'A' && b <= 'F':
			return int(b-'A') + 10, true
		}
		return 0, false
	}
	switch len(s) {
	case 3: // #rgb -> #rrggbb
		r, rok := hexVal(s[0])
		g, gok := hexVal(s[1])
		b, bok := hexVal(s[2])
		if rok && gok && bok {
			return Color{R: r * 17, G: g * 17, B: b * 17}, nil
		}
	case 6: // #rrggbb
		var c Color
		ok := true
		for i, p := range []*int{&c.R, &c.G, &c.B} {
			hi, hok := hexVal(s[i*2])
			lo, lok := hexVal(s[i*2+1])
			if !hok || !lok {
				ok = false
				break
			}
			*p = hi*16 + lo
		}
		if ok {
			return c, nil
		}
	}
	return Color{}, errs.Errorf("invalid hex color %q", s)
}

// MustHex is [Hex] that panics on an invalid string, for use with constant
// literals.
func MustHex(s string) Color {
	c, err := Hex(s)
	if err != nil {
		panic(err)
	}
	return c
}

// Common named colors, matching the basic CSS keyword palette.
var (
	Black   = Color{0, 0, 0}
	White   = Color{255, 255, 255}
	Red     = Color{255, 0, 0}
	Green   = Color{0, 128, 0}
	Blue    = Color{0, 0, 255}
	Yellow  = Color{255, 255, 0}
	Cyan    = Color{0, 255, 255}
	Magenta = Color{255, 0, 255}
	Gray50  = Color{128, 128, 128}
	Silver  = Color{192, 192, 192}
	Orange  = Color{255, 165, 0}
)

// colorMode selects between plain RGB and named spot color state on the
// renderer.
type colorMode int

const (
	colorModeRGB colorMode = iota
	colorModeSpot
)

// colorType holds the renderer's internal draw/fill/text color state.
type colorType struct {
	r, g, b    float64
	ir, ig, ib int
	mode       colorMode
	spotStr    string // name of current spot color
	gray       bool
	str        string
}
