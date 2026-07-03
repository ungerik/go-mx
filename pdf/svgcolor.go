package pdf

import (
	"math"
	"strconv"
	"strings"

	"github.com/domonda/go-errs"
)

// SVG color and paint parsing for the best-effort SVG renderer (see svg.go).

// svgPaint is the parsed value of an SVG fill or stroke attribute. The alpha
// component (from #rrggbbaa, rgba() or hsla()) multiplies into the fill or
// stroke opacity.
type svgPaint struct {
	none         bool // "none", "transparent" or an unsupported url() reference
	currentColor bool // resolved against the CSS color property at use time
	color        Color
	alpha        float64
}

// parseSVGPaint parses an SVG paint value: none, currentColor, a color, or a
// url() reference. Referenced paint servers (gradients, patterns) are not
// supported: a url() renders as its fallback color if one is given after the
// reference, otherwise as no paint.
func parseSVGPaint(s string) (svgPaint, error) {
	s = strings.TrimSpace(s)
	if after, ok := strings.CutPrefix(s, "url("); ok {
		_, fallback, ok := strings.Cut(after, ")")
		if !ok {
			return svgPaint{}, errs.Errorf("invalid SVG paint %q", s)
		}
		fallback = strings.TrimSpace(fallback)
		if fallback == "" {
			return svgPaint{none: true, alpha: 1}, nil
		}
		return parseSVGPaint(fallback)
	}
	switch s {
	case "none", "transparent":
		return svgPaint{none: true, alpha: 1}, nil
	case "currentColor":
		return svgPaint{currentColor: true, alpha: 1}, nil
	case "context-fill", "context-stroke":
		return svgPaint{none: true, alpha: 1}, nil
	}
	c, alpha, err := parseSVGColor(s)
	if err != nil {
		return svgPaint{}, err
	}
	return svgPaint{color: c, alpha: alpha}, nil
}

// parseSVGColor parses a CSS color value: #hex (3, 4, 6 or 8 digits),
// rgb()/rgba(), hsl()/hsla(), or a color keyword.
func parseSVGColor(s string) (Color, float64, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "#") {
		return parseHexColor(s)
	}
	if fn, args, ok := colorFunction(s); ok {
		switch fn {
		case "rgb", "rgba":
			return parseRGBFunction(args)
		case "hsl", "hsla":
			return parseHSLFunction(args)
		}
		return Color{}, 0, errs.Errorf("unsupported SVG color function %q", fn)
	}
	if c, ok := svgColorKeywords[strings.ToLower(s)]; ok {
		return c, 1, nil
	}
	return Color{}, 0, errs.Errorf("invalid SVG color %q", s)
}

// parseHexColor parses #rgb, #rgba, #rrggbb and #rrggbbaa.
func parseHexColor(s string) (Color, float64, error) {
	hex := s[1:]
	switch len(hex) {
	case 3, 6:
		c, err := Hex(hex)
		return c, 1, err
	case 4:
		c, err := Hex(hex[:3])
		if err != nil {
			return Color{}, 0, err
		}
		a, err := strconv.ParseUint(hex[3:], 16, 8)
		if err != nil {
			return Color{}, 0, errs.Errorf("invalid hex color %q", s)
		}
		return c, float64(a*17) / 255, nil
	case 8:
		c, err := Hex(hex[:6])
		if err != nil {
			return Color{}, 0, err
		}
		a, err := strconv.ParseUint(hex[6:], 16, 8)
		if err != nil {
			return Color{}, 0, errs.Errorf("invalid hex color %q", s)
		}
		return c, float64(a) / 255, nil
	}
	return Color{}, 0, errs.Errorf("invalid hex color %q", s)
}

// colorFunction splits "name(args)" and reports whether s has that form.
func colorFunction(s string) (name string, args []string, ok bool) {
	open := strings.IndexByte(s, '(')
	if open < 0 || !strings.HasSuffix(s, ")") {
		return "", nil, false
	}
	name = strings.ToLower(strings.TrimSpace(s[:open]))
	inner := s[open+1 : len(s)-1]
	// Accept both comma-separated legacy syntax and the space-separated
	// modern syntax with an optional "/ alpha" component.
	inner = strings.ReplaceAll(inner, "/", " ")
	if strings.Contains(inner, ",") {
		args = strings.Split(inner, ",")
	} else {
		args = strings.Fields(inner)
	}
	for i := range args {
		args[i] = strings.TrimSpace(args[i])
	}
	return name, args, true
}

// parseColorChannel parses one rgb() channel: a number 0–255 or a percentage.
func parseColorChannel(s string) (int, error) {
	if p, ok := strings.CutSuffix(s, "%"); ok {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return 0, errs.Errorf("invalid color channel %q", s)
		}
		return clampByte(v / 100 * 255), nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errs.Errorf("invalid color channel %q", s)
	}
	return clampByte(v), nil
}

// parseAlphaValue parses an opacity or alpha value: a number or a percentage,
// clamped to [0, 1].
func parseAlphaValue(s string) (float64, error) {
	s = strings.TrimSpace(s)
	scale := 1.0
	if p, ok := strings.CutSuffix(s, "%"); ok {
		s = strings.TrimSpace(p)
		scale = 0.01
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errs.Errorf("invalid opacity value %q", s)
	}
	return math.Min(1, math.Max(0, v*scale)), nil
}

func parseRGBFunction(args []string) (Color, float64, error) {
	if len(args) != 3 && len(args) != 4 {
		return Color{}, 0, errs.Errorf("rgb() needs 3 or 4 arguments, got %d", len(args))
	}
	var c Color
	for i, p := range []*int{&c.R, &c.G, &c.B} {
		v, err := parseColorChannel(args[i])
		if err != nil {
			return Color{}, 0, err
		}
		*p = v
	}
	alpha := 1.0
	if len(args) == 4 {
		var err error
		alpha, err = parseAlphaValue(args[3])
		if err != nil {
			return Color{}, 0, err
		}
	}
	return c, alpha, nil
}

func parseHSLFunction(args []string) (Color, float64, error) {
	if len(args) != 3 && len(args) != 4 {
		return Color{}, 0, errs.Errorf("hsl() needs 3 or 4 arguments, got %d", len(args))
	}
	h, err := strconv.ParseFloat(strings.TrimSuffix(args[0], "deg"), 64)
	if err != nil {
		return Color{}, 0, errs.Errorf("invalid hsl() hue %q", args[0])
	}
	s, err := parseAlphaValue(strings.TrimSuffix(args[1], "%") + "%")
	if err != nil {
		return Color{}, 0, err
	}
	l, err := parseAlphaValue(strings.TrimSuffix(args[2], "%") + "%")
	if err != nil {
		return Color{}, 0, err
	}
	alpha := 1.0
	if len(args) == 4 {
		alpha, err = parseAlphaValue(args[3])
		if err != nil {
			return Color{}, 0, err
		}
	}
	return hslToColor(h, s, l), alpha, nil
}

// hslToColor converts hue (degrees), saturation and lightness (both 0–1) to
// an RGB Color.
func hslToColor(h, s, l float64) Color {
	h = math.Mod(math.Mod(h, 360)+360, 360)
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return Color{
		R: clampByte((r + m) * 255),
		G: clampByte((g + m) * 255),
		B: clampByte((b + m) * 255),
	}
}

func clampByte(v float64) int {
	return int(math.Min(255, math.Max(0, math.Round(v))))
}

// svgColorKeywords is the SVG/CSS color keyword palette.
var svgColorKeywords = map[string]Color{
	"aliceblue":            {240, 248, 255},
	"antiquewhite":         {250, 235, 215},
	"aqua":                 {0, 255, 255},
	"aquamarine":           {127, 255, 212},
	"azure":                {240, 255, 255},
	"beige":                {245, 245, 220},
	"bisque":               {255, 228, 196},
	"black":                {0, 0, 0},
	"blanchedalmond":       {255, 235, 205},
	"blue":                 {0, 0, 255},
	"blueviolet":           {138, 43, 226},
	"brown":                {165, 42, 42},
	"burlywood":            {222, 184, 135},
	"cadetblue":            {95, 158, 160},
	"chartreuse":           {127, 255, 0},
	"chocolate":            {210, 105, 30},
	"coral":                {255, 127, 80},
	"cornflowerblue":       {100, 149, 237},
	"cornsilk":             {255, 248, 220},
	"crimson":              {220, 20, 60},
	"cyan":                 {0, 255, 255},
	"darkblue":             {0, 0, 139},
	"darkcyan":             {0, 139, 139},
	"darkgoldenrod":        {184, 134, 11},
	"darkgray":             {169, 169, 169},
	"darkgreen":            {0, 100, 0},
	"darkgrey":             {169, 169, 169},
	"darkkhaki":            {189, 183, 107},
	"darkmagenta":          {139, 0, 139},
	"darkolivegreen":       {85, 107, 47},
	"darkorange":           {255, 140, 0},
	"darkorchid":           {153, 50, 204},
	"darkred":              {139, 0, 0},
	"darksalmon":           {233, 150, 122},
	"darkseagreen":         {143, 188, 143},
	"darkslateblue":        {72, 61, 139},
	"darkslategray":        {47, 79, 79},
	"darkslategrey":        {47, 79, 79},
	"darkturquoise":        {0, 206, 209},
	"darkviolet":           {148, 0, 211},
	"deeppink":             {255, 20, 147},
	"deepskyblue":          {0, 191, 255},
	"dimgray":              {105, 105, 105},
	"dimgrey":              {105, 105, 105},
	"dodgerblue":           {30, 144, 255},
	"firebrick":            {178, 34, 34},
	"floralwhite":          {255, 250, 240},
	"forestgreen":          {34, 139, 34},
	"fuchsia":              {255, 0, 255},
	"gainsboro":            {220, 220, 220},
	"ghostwhite":           {248, 248, 255},
	"gold":                 {255, 215, 0},
	"goldenrod":            {218, 165, 32},
	"gray":                 {128, 128, 128},
	"green":                {0, 128, 0},
	"greenyellow":          {173, 255, 47},
	"grey":                 {128, 128, 128},
	"honeydew":             {240, 255, 240},
	"hotpink":              {255, 105, 180},
	"indianred":            {205, 92, 92},
	"indigo":               {75, 0, 130},
	"ivory":                {255, 255, 240},
	"khaki":                {240, 230, 140},
	"lavender":             {230, 230, 250},
	"lavenderblush":        {255, 240, 245},
	"lawngreen":            {124, 252, 0},
	"lemonchiffon":         {255, 250, 205},
	"lightblue":            {173, 216, 230},
	"lightcoral":           {240, 128, 128},
	"lightcyan":            {224, 255, 255},
	"lightgoldenrodyellow": {250, 250, 210},
	"lightgray":            {211, 211, 211},
	"lightgreen":           {144, 238, 144},
	"lightgrey":            {211, 211, 211},
	"lightpink":            {255, 182, 193},
	"lightsalmon":          {255, 160, 122},
	"lightseagreen":        {32, 178, 170},
	"lightskyblue":         {135, 206, 250},
	"lightslategray":       {119, 136, 153},
	"lightslategrey":       {119, 136, 153},
	"lightsteelblue":       {176, 196, 222},
	"lightyellow":          {255, 255, 224},
	"lime":                 {0, 255, 0},
	"limegreen":            {50, 205, 50},
	"linen":                {250, 240, 230},
	"magenta":              {255, 0, 255},
	"maroon":               {128, 0, 0},
	"mediumaquamarine":     {102, 205, 170},
	"mediumblue":           {0, 0, 205},
	"mediumorchid":         {186, 85, 211},
	"mediumpurple":         {147, 112, 219},
	"mediumseagreen":       {60, 179, 113},
	"mediumslateblue":      {123, 104, 238},
	"mediumspringgreen":    {0, 250, 154},
	"mediumturquoise":      {72, 209, 204},
	"mediumvioletred":      {199, 21, 133},
	"midnightblue":         {25, 25, 112},
	"mintcream":            {245, 255, 250},
	"mistyrose":            {255, 228, 225},
	"moccasin":             {255, 228, 181},
	"navajowhite":          {255, 222, 173},
	"navy":                 {0, 0, 128},
	"oldlace":              {253, 245, 230},
	"olive":                {128, 128, 0},
	"olivedrab":            {107, 142, 35},
	"orange":               {255, 165, 0},
	"orangered":            {255, 69, 0},
	"orchid":               {218, 112, 214},
	"palegoldenrod":        {238, 232, 170},
	"palegreen":            {152, 251, 152},
	"paleturquoise":        {175, 238, 238},
	"palevioletred":        {219, 112, 147},
	"papayawhip":           {255, 239, 213},
	"peachpuff":            {255, 218, 185},
	"peru":                 {205, 133, 63},
	"pink":                 {255, 192, 203},
	"plum":                 {221, 160, 221},
	"powderblue":           {176, 224, 230},
	"purple":               {128, 0, 128},
	"rebeccapurple":        {102, 51, 153},
	"red":                  {255, 0, 0},
	"rosybrown":            {188, 143, 143},
	"royalblue":            {65, 105, 225},
	"saddlebrown":          {139, 69, 19},
	"salmon":               {250, 128, 114},
	"sandybrown":           {244, 164, 96},
	"seagreen":             {46, 139, 87},
	"seashell":             {255, 245, 238},
	"sienna":               {160, 82, 45},
	"silver":               {192, 192, 192},
	"skyblue":              {135, 206, 235},
	"slateblue":            {106, 90, 205},
	"slategray":            {112, 128, 144},
	"slategrey":            {112, 128, 144},
	"snow":                 {255, 250, 250},
	"springgreen":          {0, 255, 127},
	"steelblue":            {70, 130, 180},
	"tan":                  {210, 180, 140},
	"teal":                 {0, 128, 128},
	"thistle":              {216, 191, 216},
	"tomato":               {255, 99, 71},
	"turquoise":            {64, 224, 208},
	"violet":               {238, 130, 238},
	"wheat":                {245, 222, 179},
	"white":                {255, 255, 255},
	"whitesmoke":           {245, 245, 245},
	"yellow":               {255, 255, 0},
	"yellowgreen":          {154, 205, 50},
}
