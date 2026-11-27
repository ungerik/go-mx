package html

import (
	"io"
	"strings"
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

// TestClassConstructors covers the class shortcuts of the layout, list,
// code, form, button, label and table elements: the class comes first,
// followed by the other attributes and the children in the order passed.
func TestClassConstructors(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"main", MainClass("mx-auto", "Content").String(), `<main class='mx-auto'>Content</main>`},
		{"header", HeaderClass("sticky top-0", "Title").String(), `<header class='sticky top-0'>Title</header>`},
		{"footer", FooterClass("border-t", "Footer").String(), `<footer class='border-t'>Footer</footer>`},
		{"nav", NavClass("flex", "Links").String(), `<nav class='flex'>Links</nav>`},
		{"section", SectionClass("mb-10", "Part").String(), `<section class='mb-10'>Part</section>`},
		{"article", ArticleClass("prose", "Text").String(), `<article class='prose'>Text</article>`},
		{"aside", AsideClass("w-64", "Sidebar").String(), `<aside class='w-64'>Sidebar</aside>`},
		{"ul", ULClass("list-none", LI("Item")).String(), `<ul class='list-none'><li>Item</li></ul>`},
		{"ol", OLClass("list-decimal", LI("Item")).String(), `<ol class='list-decimal'><li>Item</li></ol>`},
		{"li", LIClass("mt-3", "Item").String(), `<li class='mt-3'>Item</li>`},
		{"pre", PreClass("hl", "x := 1").String(), `<pre class='hl'>x := 1</pre>`},
		{"code", CodeClass("language-go", "x").String(), `<code class='language-go'>x</code>`},
		{"form", FormClass("space-y-4", ID("login")).String(), `<form class='space-y-4' id='login'></form>`},
		{"button", ButtonClass("btn", "Save").String(), `<button class='btn'>Save</button>`},
		{"label", LabelClass("text-sm", For("email"), "Email").String(), `<label class='text-sm' for='email'>Email</label>`},
		{"table", TableClass("w-full", TR(TD("1"))).String(), `<table class='w-full'><tr><td>1</td></tr></table>`},
		{"tr", TRClass("border-b", TD("1")).String(), `<tr class='border-b'><td>1</td></tr>`},
		{"th", THClass("text-left", "Name").String(), `<th class='text-left'>Name</th>`},
		{"td", TDClass("p-2", "1").String(), `<td class='p-2'>1</td>`},
		{
			"attribs and children after the class",
			NavClass("mt-2", Attrib("aria-label", "Documentation"), A("Home")).String(),
			`<nav class='mt-2' aria-label='Documentation'><a>Home</a></nav>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.got)
		})
	}
}

// TestRequiredAttribConstructors covers the constructors that take the
// attribute their element requires as the first parameter: the parameter is
// the element's only such attribute, a value the spec does not allow makes
// Render return an error, and the same attribute passed again among the
// attributes makes Render fail as a duplicate.
func TestRequiredAttribConstructors(t *testing.T) {
	valid := []struct {
		name string
		elem *mx.Element
		want string
	}{
		{"Img", Img("/a.png", Alt("A")), `<img src="/a.png" alt="A"/>`},
		{"Img with SrcSet", Img("/a.png", SrcSet("/a.png 1x, /a@2x.png 2x")), `<img src="/a.png" srcset="/a.png 1x, /a@2x.png 2x"/>`},
		{"Track", Track("/en.vtt", SrcLang("en")), `<track src="/en.vtt" srclang="en"/>`},
		{"OptGroup", OptGroup("Fruits", Option("Apple")), `<optgroup label="Fruits"><option>Apple</option></optgroup>`},
		// The spec allows an empty label and an empty value, so they render.
		{"OptGroup empty label", OptGroup("", Option("Apple")), `<optgroup label=""><option>Apple</option></optgroup>`},
		{"Data", Data("398", "Mini Ketchup"), `<data value="398">Mini Ketchup</data>`},
		{"Data empty value", Data("", "x"), `<data value="">x</data>`},
		{"Map", Map("floors", Area(ShapeDefault)), `<map name="floors"><area shape="default"/></map>`},
		{"BDo ltr", BDo(DirLTR, "Hello"), `<bdo dir="ltr">Hello</bdo>`},
		{"BDo rtl", BDo(DirRTL, "Hello"), `<bdo dir="rtl">Hello</bdo>`},
	}
	for _, tt := range valid {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			require.NoError(t, tt.elem.Render(t.Context(), mx.NewCheckedWriter(&b)))
			require.Equal(t, tt.want, b.String())
		})
	}

	invalid := []struct {
		name    string
		elem    *mx.Element
		wantErr string
	}{
		{"Img empty src", Img(""), "img src must not be empty"},
		// A srcset does not replace the required src.
		{"Img empty src with SrcSet", Img("", SrcSet("/a.png 1x")), "img src must not be empty"},
		{"Track empty src", Track(""), "track src must not be empty"},
		{"Map empty name", Map(""), `map name must be non-empty without ASCII whitespace, not ""`},
		{"Map name with space", Map("floor plan"), `not "floor plan"`},
		{"Map name with tab", Map("floor\tplan"), `not "floor\tplan"`},
		{"Map name with newline", Map("floor\n"), `not "floor\n"`},
		{"Map name with form feed", Map("\ffloor"), `not "\ffloor"`},
		{"Map name with carriage return", Map("floor\r"), `not "floor\r"`},
		// <bdo> overrides the direction, so it needs an explicit one.
		{"BDo auto", BDo(DirAuto, "Hello"), `bdo dir must be "ltr" or "rtl", not "auto"`},
		{"BDo empty", BDo("", "Hello"), `not ""`},
		{"BDo unknown", BDo("up", "Hello"), `not "up"`},
		// The parameter is the only source of the attribute: the same
		// attribute among the others is a duplicate. An invalid parameter
		// fails with its own error before the duplicate is reached.
		{"Img duplicate src", Img("/a.png", Src("/b.png")), "duplicate attribute"},
		{"Track duplicate src", Track("/en.vtt", Src("/de.vtt")), "duplicate attribute"},
		{"OptGroup duplicate label", OptGroup("Fruits", LabelAttr("Obst")), "duplicate attribute"},
		{"OptGroup empty and label", OptGroup("", LabelAttr("Fruits")), "duplicate attribute"},
		{"Data duplicate value", Data("1", mx.NewAttrib("value", "2"), "x"), "duplicate attribute"},
		{"Map duplicate name", Map("a", Name("b")), "duplicate attribute"},
		{"BDo duplicate dir", BDo(DirLTR, DirRTL, "Hello"), "duplicate attribute"},
		{"BDo empty and dir", BDo("", DirLTR, "Hello"), `not ""`},
		// The forms a typed DirAuto check would miss.
		{"BDo dir auto as Attribute", BDo(DirLTR, mx.NewAttrib("dir", "auto"), "Hello"), "duplicate attribute"},
		{"BDo dir auto as ConstAttrib", BDo(DirLTR, mx.ConstAttrib("dir=auto"), "Hello"), "duplicate attribute"},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.elem.Render(t.Context(), mx.NewCheckedWriter(io.Discard))
			require.ErrorContains(t, err, tt.wantErr)
		})
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
