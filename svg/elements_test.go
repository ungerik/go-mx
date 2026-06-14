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
				ViewBox("0 0 100 100"),
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
	got := Root(ViewBox("0 0 10 10"), Circle(CX("5"), CY("5"), R("4"))).String()
	want := `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 10'><circle cx='5' cy='5' r='4'></circle></svg>`
	require.Equal(t, want, got)
}

func TestVoidElement(t *testing.T) {
	got := VoidElement("rect", X("0"), Y("0"), Width("10"), Height("10")).String()
	want := `<rect x='0' y='0' width='10' height='10'/>`
	require.Equal(t, want, got)
}
