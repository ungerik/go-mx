package pdf

import (
	"bytes"
	"testing"
	"time"
)

// newDeterministicRenderer returns a renderer with pinned dates and disabled
// compression so its output is byte-comparable across runs.
func newDeterministicRenderer() *Renderer {
	r := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	r.SetCreationDate(fixed)
	r.SetModificationDate(fixed)
	r.SetCompression(false)
	r.SetCatalogSort(true)
	return r
}

func renderBytes(t *testing.T, r *Renderer) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := r.Output(&buf); err != nil {
		t.Fatalf("output: %v", err)
	}
	return buf.Bytes()
}

// TestConstructorsInstallTranslator guards that every constructor installs the
// cp1252 translator, so text components render non-ASCII cp1252 text the same
// way regardless of which constructor created the renderer.
func TestConstructorsInstallTranslator(t *testing.T) {
	for name, r := range map[string]*Renderer{
		"New":         New(OrientationPortrait, UnitMillimeter, PageSizeA4, ""),
		"NewCustom":   NewCustom(&InitType{}),
		"NewRenderer": NewRenderer(OrientationPortrait, UnitMillimeter, PageSizeA4),
	} {
		if got := r.Str("é€"); got != "\xe9\x80" {
			t.Errorf("%s: Str(%q) = %x, want e980 (cp1252 translator not installed)", name, "é€", got)
		}
	}
}

// TestGetStringSymbolWidthCwBoundary guards the character-width lookup against
// the off-by-one where a rune's code point equals len(Cw): it must fall
// through to MissingWidth instead of indexing out of range.
func TestGetStringSymbolWidthCwBoundary(t *testing.T) {
	r := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	r.isCurrentUTF8 = true
	r.currentFont = fontDefType{
		Cw:   make([]int, 'A'+1), // rune 'B' == len(Cw)
		Desc: FontDescType{MissingWidth: 111},
	}
	r.currentFont.Cw['A'] = 222
	if got := r.GetStringSymbolWidth("B"); got != 111 {
		t.Errorf("GetStringSymbolWidth(rune == len(Cw)) = %d, want MissingWidth 111", got)
	}
	if got := r.GetStringSymbolWidth("A"); got != 222 {
		t.Errorf("GetStringSymbolWidth(in range) = %d, want 222", got)
	}
}

// TestSpotColorOutputDeterministic renders a document with several spot colors
// repeatedly and requires byte-identical output: the spot color objects and
// resource dictionary must not depend on map iteration order.
func TestSpotColorOutputDeterministic(t *testing.T) {
	render := func() []byte {
		r := newDeterministicRenderer()
		r.AddSpotColor("PANTONE 145 CVC", 0, 42, 100, 25)
		r.AddSpotColor("PANTONE 300 CVC", 100, 43, 0, 0)
		r.AddSpotColor("PANTONE 871 CVC", 30, 40, 70, 15)
		r.AddPage()
		r.SetFillSpotColor("PANTONE 300 CVC", 100)
		r.Rect(20, 20, 50, 20, "F")
		return renderBytes(t, r)
	}
	want := render()
	for range 10 {
		if got := render(); !bytes.Equal(got, want) {
			t.Fatal("spot color output differs between runs")
		}
	}
}

// TestPageBoxOutputDeterministic renders a page with multiple page boxes
// repeatedly and requires byte-identical output.
func TestPageBoxOutputDeterministic(t *testing.T) {
	render := func() []byte {
		r := newDeterministicRenderer()
		r.SetPageBox("trim", 5, 5, 200, 287)
		r.SetPageBox("bleed", 2, 2, 206, 293)
		r.SetPageBox("crop", 0, 0, 210, 297)
		r.SetFont("Helvetica", "", 12)
		r.AddPage()
		r.Cell(50, 8, "boxes")
		return renderBytes(t, r)
	}
	want := render()
	for range 10 {
		if got := render(); !bytes.Equal(got, want) {
			t.Fatal("page box output differs between runs")
		}
	}
}

// TestReplaceAliasesInteracting guards that aliases are replaced longest
// first, so an alias containing another alias as a prefix is never corrupted
// by the shorter one's replacement, independent of map iteration order.
func TestReplaceAliasesInteracting(t *testing.T) {
	render := func() []byte {
		r := newDeterministicRenderer()
		r.AddPage()
		r.RegisterAlias("{n}", "1")
		r.RegisterAlias("{n}2", "2")
		r.Text(20, 20, "{n}2 {n}")
		return renderBytes(t, r)
	}
	want := render()
	if !bytes.Contains(want, []byte("(2 1)")) {
		t.Errorf("aliases not replaced longest-first: output does not contain %q", "(2 1)")
	}
	for range 10 {
		if got := render(); !bytes.Equal(got, want) {
			t.Fatal("alias replacement differs between runs")
		}
	}
}
