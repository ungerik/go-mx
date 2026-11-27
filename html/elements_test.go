package html

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ungerik/go-mx"
)

func TestAHRef(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"text only", AHRef("https://example.com", "Example").String(), `<a href='https://example.com'>Example</a>`},
		{
			"attrib then text",
			AHRef("https://example.com", TargetBlank, "Example").String(),
			`<a href='https://example.com' target='_blank'>Example</a>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}

func TestImg(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"src first", Img("/a.png", Alt("A")).String(), `<img src='/a.png' alt='A'/>`},
		// An empty src leaves the source to attribs, here a srcset.
		{"empty src renders no src", Img("", SrcSet("/a.png 1x, /a@2x.png 2x")).String(), `<img srcset='/a.png 1x, /a@2x.png 2x'/>`},
		{"empty src without attribs", Img("").String(), `<img/>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}

func TestRequiredAttribConstructors(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Track", Track("/en.vtt", SrcLang("en")).String(), `<track src='/en.vtt' srclang='en'/>`},
		{"Track empty src", Track("").String(), `<track/>`},
		{"OptGroup", OptGroup("Fruits", Option("Apple")).String(), `<optgroup label='Fruits'><option>Apple</option></optgroup>`},
		{"OptGroup empty label", OptGroup("", LabelAttr("Fruits")).String(), `<optgroup label='Fruits'></optgroup>`},
		{"Data", Data("398", "Mini Ketchup").String(), `<data value='398'>Mini Ketchup</data>`},
		{"Data empty value", Data("", "x").String(), `<data>x</data>`},
		{"Map", Map("floors", Area(ShapeDefault)).String(), `<map name='floors'><area shape='default'/></map>`},
		{"Map empty name", Map("").String(), `<map></map>`},
		{"BDo", BDo(DirRTL, "Hello").String(), `<bdo dir='rtl'>Hello</bdo>`},
		{"BDo empty dir", BDo("", DirLTR, "Hello").String(), `<bdo dir='ltr'>Hello</bdo>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}

func TestBDoRejectsDirAuto(t *testing.T) {
	// <bdo> overrides the direction, so it needs an explicit one, whether
	// DirAuto comes as the dir parameter or among the attributes.
	for _, bdo := range []*mx.Element{BDo(DirAuto, "Hello"), BDo("", DirAuto, "Hello")} {
		err := bdo.Render(t.Context(), mx.NewCheckedWriter(io.Discard))
		require.Error(t, err)
	}
}

func TestButtonTypeConstructors(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"submit", SubmitButton(Text("Save")).String(), `<button type='submit'>Save</button>`},
		{"reset", ResetButton(Text("Reset")).String(), `<button type='reset'>Reset</button>`},
		{"button", ButtonButton(Text("Toggle")).String(), `<button type='button'>Toggle</button>`},
		{"bare defaults to submit", Button(Text("Save")).String(), `<button>Save</button>`},
		{
			"attribs and children mix",
			SubmitButton(Class("primary"), ID("save"), Text("Save")).String(),
			`<button type='submit' class='primary' id='save'>Save</button>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}

func TestOLTypeConstructors(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"bare has no type", OL(LI("First")).String(), `<ol><li>First</li></ol>`},
		{"decimal", OLDecimal(LI("First")).String(), `<ol type='1'><li>First</li></ol>`},
		{"lower alpha", OLLowerAlpha(LI("First")).String(), `<ol type='a'><li>First</li></ol>`},
		{"upper alpha", OLUpperAlpha(LI("First")).String(), `<ol type='A'><li>First</li></ol>`},
		{"lower roman", OLLowerRoman(LI("First")).String(), `<ol type='i'><li>First</li></ol>`},
		{"upper roman", OLUpperRoman(LI("First")).String(), `<ol type='I'><li>First</li></ol>`},
		{
			"type with start, reversed and item children",
			OLUpperRoman(Start("3"), Reversed, LI("Third"), LI("Second")).String(),
			`<ol type='I' start='3' reversed='reversed'><li>Third</li><li>Second</li></ol>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}
