package svg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestElements(t *testing.T) {
	tests := []struct {
		name string
		elem interface{ String() string }
		want string
	}{
		{
			name: "circle",
			elem: Circle(CX("50"), CY("50"), R("40"), Fill("red")),
			want: `<circle cx='50' cy='50' r='40' fill='red'></circle>`,
		},
		{
			name: "numeric literal attribute values",
			elem: Circle(CX(50), CY(50), R(40), StrokeWidth(1.5), Fill("red")),
			want: `<circle cx='50' cy='50' r='40' stroke-width='1.5' fill='red'></circle>`,
		},
		{
			name: "small float renders as plain decimal, not scientific notation",
			elem: Circle(CX(50), StrokeWidth(0.00005), StrokeOpacity(float32(0.0001))),
			want: `<circle cx='50' stroke-width='0.00005' stroke-opacity='0.0001'></circle>`,
		},
		{
			name: "path",
			elem: Path(D("M10 10 H 90 V 90 H 10 Z"), Stroke("black"), Fill("none")),
			want: `<path d='M10 10 H 90 V 90 H 10 Z' stroke='black' fill='none'></path>`,
		},
		{
			name: "nested svg",
			elem: SVG(
				ViewBox(0, 0, 100, 100),
				Rect(X("0"), Y("0"), Width("100"), Height("100"), Fill("blue")),
			),
			want: `<svg viewBox='0 0 100 100'><rect x='0' y='0' width='100' height='100' fill='blue'></rect></svg>`,
		},
		{
			name: "camelCase element names preserved",
			elem: LinearGradient(ID("g"), GradientUnits("userSpaceOnUse")),
			want: `<linearGradient id='g' gradientUnits='userSpaceOnUse'></linearGradient>`,
		},
		{
			name: "points list renders space-separated numbers",
			elem: Polygon(Points(0, 0, 10, 0, 10, 10), Fill("lime")),
			want: `<polygon points='0 0 10 0 10 10' fill='lime'></polygon>`,
		},
		{
			name: "float64 filter constants and int integer attrs",
			elem: FeComposite(Operator("arithmetic"), K1(0.5), K2(0), K3(0), K4(1)),
			want: `<feComposite operator='arithmetic' k1='0.5' k2='0' k3='0' k4='1'></feComposite>`,
		},
		{
			name: "stdDeviation accepts one or two numbers",
			elem: FeGaussianBlur(StdDeviation(2, 3)),
			want: `<feGaussianBlur stdDeviation='2 3'></feGaussianBlur>`,
		},
		{
			name: "integer-only attrs use int",
			elem: FeTurbulence(NumOctaves(3), Seed(2), BaseFrequency(0.05)),
			want: `<feTurbulence numOctaves='3' seed='2' baseFrequency='0.05'></feTurbulence>`,
		},
		{
			name: "keyTimes renders a semicolon-separated list",
			elem: Animate(AttributeName("opacity"), KeyTimes(0, 0.5, 1)),
			want: `<animate attributeName='opacity' keyTimes='0;0.5;1'></animate>`,
		},
		{
			name: "keySplines groups four values per spline",
			elem: Animate(KeySplines(0, 0, 1, 1, 0.5, 0, 0.5, 1)),
			want: `<animate keySplines='0 0 1 1;0.5 0 0.5 1'></animate>`,
		},
		{
			name: "text element with content",
			elem: Text(X("10"), Y("20"), TextAnchor("middle"), "Hello"),
			want: `<text x='10' y='20' text-anchor='middle'>Hello</text>`,
		},
		{
			name: "title escapes content",
			elem: Title("A & B"),
			want: `<title>A &amp; B</title>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.elem.String())
		})
	}
}

func TestRoot(t *testing.T) {
	got := Root(ViewBox(0, 0, 10, 10), Circle(CX("5"), CY("5"), R("4"))).String()
	want := `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 10'><circle cx='5' cy='5' r='4'></circle></svg>`
	require.Equal(t, want, got)
}

func TestKeySplinesIncomplete(t *testing.T) {
	// A non-multiple-of-four value count must render a loud error into the
	// attribute value so the SVG fails rather than silently emitting an
	// invalid keySplines list.
	got := Animate(KeySplines(0, 0, 1)).String()
	want := "mx.Element.String: keySplines needs 4 values (x1 y1 x2 y2) per spline, got 3"
	require.Equal(t, want, got)
}

func TestEnumAttribs(t *testing.T) {
	// Enum constants are usable directly as element attributes, and a conversion
	// of a dynamic value works too (TextAnchor("middle")).
	got := Circle(CX(5), CY(5), R(4), FillRuleEvenodd, StrokeLineCapRound, TextAnchor("middle")).String()
	want := `<circle cx='5' cy='5' r='4' fill-rule='evenodd' stroke-linecap='round' text-anchor='middle'></circle>`
	require.Equal(t, want, got)

	// go-enum generated validation and value listing.
	require.True(t, FillRuleEvenodd.Valid())
	require.False(t, FillRule("bogus").Valid())
	require.Error(t, FillRule("bogus").Validate())
	require.Equal(t, []string{"userSpaceOnUse", "objectBoundingBox"}, GradientUnits("").EnumStrings())
}

func TestVoidElement(t *testing.T) {
	got := VoidElement("rect", X("0"), Y("0"), Width("10"), Height("10")).String()
	want := `<rect x='0' y='0' width='10' height='10'/>`
	require.Equal(t, want, got)
}

func TestJoinNums(t *testing.T) {
	tests := []struct {
		name   string
		sep    string
		values []float64
		want   string
	}{
		{name: "no values", sep: " ", values: nil, want: ""},
		{name: "single value (no separator)", sep: " ", values: []float64{1.5}, want: "1.5"},
		{name: "space separated", sep: " ", values: []float64{0, 0, 100, 100}, want: "0 0 100 100"},
		{name: "semicolon separated", sep: ";", values: []float64{0, 0.5, 1}, want: "0;0.5;1"},
		{
			name:   "small and large magnitudes render as plain decimals",
			sep:    " ",
			values: []float64{0.00005, 1000000, -2.5},
			want:   "0.00005 1000000 -2.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, joinNums(tt.sep, tt.values...))
		})
	}
}
