// Copyright ©2025 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pdf

import (
	"errors"
	"io"
	"math"
	"strings"
	"testing"
	"time"
)

var floatEpsilon = math.Nextafter(1.0, 2.0) - 1.0

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) <= floatEpsilon
}

func TestGetAlpha(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetAlpha(0.17, BlendModeLuminosity)

	alpha, mode := pdf.GetAlpha()

	if got, want := alpha, 0.17; !floatEqual(got, want) {
		t.Errorf("invalid alpha value: got=%v, want=%v", got, want)
	}
	if got, want := mode, BlendModeLuminosity; got != want {
		t.Errorf("invalid blend mode: got=%v, want=%v", got, want)
	}
}

func TestGetAuthor(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetAuthor("John Doe", false)

	author := pdf.GetAuthor()

	if got, want := author, "John Doe"; got != want {
		t.Errorf("invalid author: got=%v, want=%v", got, want)
	}
}

func TestGetAutoPageBreak(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetAutoPageBreak(true, 10)

	autoPageBreak, margin := pdf.GetAutoPageBreak()

	if got, want := autoPageBreak, true; got != want {
		t.Errorf("invalid autoPageBreak: got=%v, want=%v", got, want)
	}
	if got, want := margin, 10.0; !floatEqual(got, want) {
		t.Errorf("invalid margin: got=%v, want=%v", got, want)
	}
}

func TestGetCatalogSort(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetCatalogSort(true)

	catalogSort := pdf.GetCatalogSort()

	if got, want := catalogSort, true; got != want {
		t.Errorf("invalid catalogSort: got=%v, want=%v", got, want)
	}

	pdf.SetCatalogSort(false)

	catalogSort = pdf.GetCatalogSort()

	if got, want := catalogSort, false; got != want {
		t.Errorf("invalid catalogSort: got=%v, want=%v", got, want)
	}
}

func TestGetCellMargin(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetCellMargin(6)

	cellMargin := pdf.GetCellMargin()

	if got, want := cellMargin, 6.0; !floatEqual(got, want) {
		t.Errorf("invalid cellMargin: got=%v, want=%v", got, want)
	}
}

func TestGetCompression(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetCompression(true)

	compression := pdf.GetCompression()

	if got, want := compression, true; got != want {
		t.Errorf("invalid compression: got=%v, want=%v", got, want)
	}

	pdf.SetCompression(false)

	compression = pdf.GetCompression()

	if got, want := compression, false; got != want {
		t.Errorf("invalid compression: got=%v, want=%v", got, want)
	}
}

func TestGetConversionRatio(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")

	conversionRatio := pdf.GetConversionRatio()

	if got, want := conversionRatio, 72.0/25.4; !floatEqual(got, want) {
		t.Errorf("invalid conversionRatio: got=%v, want=%v", got, want)
	}

	pdf = New(OrientationPortrait, UnitPoint, PageSizeA4, "")

	conversionRatio = pdf.GetConversionRatio()

	if got, want := conversionRatio, 1.0; !floatEqual(got, want) {
		t.Errorf("invalid conversionRatio: got=%v, want=%v", got, want)
	}
}

func TestGetCreationDate(t *testing.T) {
	setDate, _ := time.Parse(time.RFC3339, "2003-06-17T01:23:45Z")
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetCreationDate(setDate)

	creationDate := pdf.GetCreationDate()

	if got, want := creationDate, setDate; !got.Equal(want) {
		t.Errorf("invalid creationDate: got=%v, want=%v", got, want)
	}
}

func TestGetCreator(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetCreator("John Doe", false)

	creator := pdf.GetCreator()

	if got, want := creator, "John Doe"; got != want {
		t.Errorf("invalid creator: got=%v, want=%v", got, want)
	}
}

func TestGetDisplayMode(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetDisplayMode("real", "OneColumn")

	zoom, layout := pdf.GetDisplayMode()

	if got, want := zoom, "real"; got != want {
		t.Errorf("invalid zoom: got=%v, want=%v", got, want)
	}
	if got, want := layout, "OneColumn"; got != want {
		t.Errorf("invalid layout: got=%v, want=%v", got, want)
	}
}

func TestGetDrawColor(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetDrawColor(134, 26, 34)

	r, g, b := pdf.GetDrawColor()

	if got, want := r, 134; got != want {
		t.Errorf("invalid red component: got=%v, want=%v", got, want)
	}
	if got, want := g, 26; got != want {
		t.Errorf("invalid green component: got=%v, want=%v", got, want)
	}
	if got, want := b, 34; got != want {
		t.Errorf("invalid blue component: got=%v, want=%v", got, want)
	}
}

func TestGetDrawSpotColor(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.AddSpotColor("RAL 5018", 81, 0, 5, 48)
	pdf.SetDrawSpotColor("RAL 5018", 100)

	name, c, m, y, k := pdf.GetDrawSpotColor()

	if got, want := name, "RAL 5018"; got != want {
		t.Errorf("invalid spot color name: got=%v, want=%v", got, want)
	}
	if got, want := c, byte(81); got != want {
		t.Errorf("invalid cyan component: got=%v, want=%v", got, want)
	}
	if got, want := m, byte(0); got != want {
		t.Errorf("invalid magenta component: got=%v, want=%v", got, want)
	}
	if got, want := y, byte(5); got != want {
		t.Errorf("invalid yellow component: got=%v, want=%v", got, want)
	}
	if got, want := k, byte(48); got != want {
		t.Errorf("invalid black component: got=%v, want=%v", got, want)
	}
}

func TestGetFillColor(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetFillColor(255, 203, 0)

	r, g, b := pdf.GetFillColor()

	if got, want := r, 255; got != want {
		t.Errorf("invalid red component: got=%v, want=%v", got, want)
	}
	if got, want := g, 203; got != want {
		t.Errorf("invalid green component: got=%v, want=%v", got, want)
	}
	if got, want := b, 0; got != want {
		t.Errorf("invalid blue component: got=%v, want=%v", got, want)
	}
}

func TestGetFillSpotColor(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.AddSpotColor("RAL 5018", 81, 0, 5, 48)
	pdf.SetFillSpotColor("RAL 5018", 100)

	name, c, m, y, k := pdf.GetFillSpotColor()

	if got, want := name, "RAL 5018"; got != want {
		t.Errorf("invalid spot color name: got=%v, want=%v", got, want)
	}
	if got, want := c, byte(81); got != want {
		t.Errorf("invalid cyan component: got=%v, want=%v", got, want)
	}
	if got, want := m, byte(0); got != want {
		t.Errorf("invalid magenta component: got=%v, want=%v", got, want)
	}
	if got, want := y, byte(5); got != want {
		t.Errorf("invalid yellow component: got=%v, want=%v", got, want)
	}
	if got, want := k, byte(48); got != want {
		t.Errorf("invalid black component: got=%v, want=%v", got, want)
	}
}

func TestGetFontFamily(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetFont("Times", "", 12)

	fontFamily := pdf.GetFontFamily()

	if got, want := fontFamily, "times"; got != want {
		t.Errorf("invalid fontFamily: got=%v, want=%v", got, want)
	}
}

type testFontLoader struct {
	reader io.Reader
	err    error
}

func (tfl *testFontLoader) Open(name string) (io.Reader, error) {
	return tfl.reader, tfl.err
}

func TestGetFontLoader(t *testing.T) {
	testErr := errors.New("TestGetFontLoader error")
	tfl := &testFontLoader{
		reader: strings.NewReader("TestGetFontLoader reader"),
		err:    testErr,
	}

	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetFontLoader(tfl)

	fontLoader := pdf.GetFontLoader()

	reader, err := fontLoader.Open("test")

	if got := reader; got == nil {
		t.Fatalf("invalid reader: got=nil, want non-nil")
	}
	if got, want := err, testErr; got != want {
		t.Errorf("invalid error: got=%v, want=%v", got, want)
	}

	read, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("reading error: got=%v, want=nil", err)
	}
	if got, want := string(read[:]), "TestGetFontLoader reader"; got != want {
		t.Errorf("invalid reader content: got=%v, want=%v", got, want)
	}
}

func TestGetFontLocation(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetFontLocation("test-font-location")

	fontLocation := pdf.GetFontLocation()

	if got, want := fontLocation, "test-font-location"; got != want {
		t.Errorf("invalid fontLocation: got=%v, want=%v", got, want)
	}
}

func TestGetFontSize(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetFontSize(19)

	ptSize, _ := pdf.GetFontSize()

	if got, want := ptSize, 19.0; !floatEqual(got, want) {
		t.Errorf("invalid ptSize: got=%v, want=%v", got, want)
	}

	pdf.SetFontUnitSize(246)

	_, unitSize := pdf.GetFontSize()

	if got, want := unitSize, 246.0; !floatEqual(got, want) {
		t.Errorf("invalid unitSize: got=%v, want=%v", got, want)
	}
}

func TestGetFontStyle(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetFont("Arial", "BIUS", 12)

	fontStyle := pdf.GetFontStyle()

	if got, want := len(fontStyle), 4; got != want {
		t.Errorf("invalid fontStyle length: got=%v, want=%v", got, want)
	}
	if got, want := strings.Contains(fontStyle, "B"), true; got != want {
		t.Errorf("missing bold style: got=%v, want=%v", got, want)
	}
	if got, want := strings.Contains(fontStyle, "I"), true; got != want {
		t.Errorf("missing italic style: got=%v, want=%v", got, want)
	}
	if got, want := strings.Contains(fontStyle, "U"), true; got != want {
		t.Errorf("missing underline style: got=%v, want=%v", got, want)
	}
	if got, want := strings.Contains(fontStyle, "S"), true; got != want {
		t.Errorf("missing strikeout style: got=%v, want=%v", got, want)
	}
}

func TestGetJavascript(t *testing.T) {
	const want = `<script>console.log('fpdf is awesome')</script>`
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")

	if got, want := pdf.GetJavascript(), ""; got != want {
		t.Errorf("invalid javascript: got=%v, want=%v", got, want)
	}

	{
		want := ""
		pdf.SetJavascript(want)
		if got := pdf.GetJavascript(); got != want {
			t.Errorf("invalid javascript: got=%v, want=%v", got, want)
		}
	}

	pdf.SetJavascript(want)

	got := pdf.GetJavascript()
	if got != want {
		t.Errorf("invalid javascript: got=%v, want=%v", got, want)
	}
}

func TestGetKeywords(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetKeywords("test keywords", false)

	keywords := pdf.GetKeywords()

	if got, want := keywords, "test keywords"; got != want {
		t.Errorf("invalid keywords: got=%v, want=%v", got, want)
	}
}

func TestGetLang(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetLang("de-CH")

	lang := pdf.GetLang()

	if got, want := lang, "de-CH"; got != want {
		t.Errorf("invalid lang: got=%v, want=%v", got, want)
	}
}

func TestGetLineCapStyle(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetLineCapStyle("round")

	lineCapStyle := pdf.GetLineCapStyle()

	if got, want := lineCapStyle, "round"; got != want {
		t.Errorf("invalid lineCapStyle: got=%v, want=%v", got, want)
	}
}

func TestGetLineJoinStyle(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetLineJoinStyle("bevel")

	lineJoinStyle := pdf.GetLineJoinStyle()

	if got, want := lineJoinStyle, "bevel"; got != want {
		t.Errorf("invalid lineJoinStyle: got=%v, want=%v", got, want)
	}
}

func TestGetLineWidth(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetLineWidth(42)

	lineWidth := pdf.GetLineWidth()

	if got, want := lineWidth, 42.0; !floatEqual(got, want) {
		t.Errorf("invalid lineWidth: got=%v, want=%v", got, want)
	}
}

func TestGetMargins(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetMargins(17, 6, 3)
	pdf.SetAutoPageBreak(true, 3.14)

	left, top, right, bottom := pdf.GetMargins()

	if got, want := left, 17.0; !floatEqual(got, want) {
		t.Errorf("invalid left margin: got=%v, want=%v", got, want)
	}
	if got, want := top, 6.0; !floatEqual(got, want) {
		t.Errorf("invalid top margin: got=%v, want=%v", got, want)
	}
	if got, want := right, 3.0; !floatEqual(got, want) {
		t.Errorf("invalid right margin: got=%v, want=%v", got, want)
	}
	if got, want := bottom, 3.14; !floatEqual(got, want) {
		t.Errorf("invalid bottom margin: got=%v, want=%v", got, want)
	}
}

func TestGetModificationDate(t *testing.T) {
	setDate, _ := time.Parse(time.RFC3339, "9-08-02T09:54:32Z")
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetModificationDate(setDate)

	modificationDate := pdf.GetModificationDate()

	if got, want := modificationDate, setDate; !got.Equal(want) {
		t.Errorf("invalid modificationDate: got=%v, want=%v", got, want)
	}
}

func TestGetPageSize(t *testing.T) {
	pdf := New(OrientationPortrait, UnitPoint, PageSizeA4, "")

	pageWidth, pageHeight := pdf.GetPageSize()

	if got, want := pageWidth, 595.28; !floatEqual(got, want) {
		t.Errorf("invalid pageWidth: got=%v, want=%v", got, want)
	}
	if got, want := pageHeight, 841.89; !floatEqual(got, want) {
		t.Errorf("invalid pageHeight: got=%v, want=%v", got, want)
	}
}

func TestGetProducer(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetProducer("John Doe", false)

	producer := pdf.GetProducer()

	if got, want := producer, "John Doe"; got != want {
		t.Errorf("invalid producer: got=%v, want=%v", got, want)
	}
}

func TestGetSubject(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetSubject("test subject", false)

	subject := pdf.GetSubject()

	if got, want := subject, "test subject"; got != want {
		t.Errorf("invalid subject: got=%v, want=%v", got, want)
	}
}

func TestGetTextColor(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetTextColor(255, 203, 0)

	r, g, b := pdf.GetTextColor()

	if got, want := r, 255; got != want {
		t.Errorf("invalid red component: got=%v, want=%v", got, want)
	}
	if got, want := g, 203; got != want {
		t.Errorf("invalid green component: got=%v, want=%v", got, want)
	}
	if got, want := b, 0; got != want {
		t.Errorf("invalid blue component: got=%v, want=%v", got, want)
	}
}

func TestGetTextSpotColor(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.AddSpotColor("RAL 5018", 81, 0, 5, 48)
	pdf.SetTextSpotColor("RAL 5018", 100)

	name, c, m, y, k := pdf.GetTextSpotColor()

	if got, want := name, "RAL 5018"; got != want {
		t.Errorf("invalid spot color name: got=%v, want=%v", got, want)
	}
	if got, want := c, byte(81); got != want {
		t.Errorf("invalid cyan component: got=%v, want=%v", got, want)
	}
	if got, want := m, byte(0); got != want {
		t.Errorf("invalid magenta component: got=%v, want=%v", got, want)
	}
	if got, want := y, byte(5); got != want {
		t.Errorf("invalid yellow component: got=%v, want=%v", got, want)
	}
	if got, want := k, byte(48); got != want {
		t.Errorf("invalid black component: got=%v, want=%v", got, want)
	}
}

func TestGetTitle(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetTitle("test title", false)

	title := pdf.GetTitle()

	if got, want := title, "test title"; got != want {
		t.Errorf("invalid title: got=%v, want=%v", got, want)
	}
}

func TestGetUnderlineThickness(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetUnderlineThickness(17)

	underlineThickness := pdf.GetUnderlineThickness()

	if got, want := underlineThickness, 17.0; !floatEqual(got, want) {
		t.Errorf("invalid underlineThickness: got=%v, want=%v", got, want)
	}
}

func TestGetWordSpacing(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetWordSpacing(6)

	wordSpacing := pdf.GetWordSpacing()

	if got, want := wordSpacing, 6.0; !floatEqual(got, want) {
		t.Errorf("invalid wordSpacing: got=%v, want=%v", got, want)
	}
}

func TestGetX(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetX(17)

	x := pdf.GetX()

	if got, want := x, 17.0; !floatEqual(got, want) {
		t.Errorf("invalid x coordinate: got=%v, want=%v", got, want)
	}
}

func TestGetXmpMetadata(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetXmpMetadata([]byte("test xmp metadata"))

	xmpMetadata := pdf.GetXmpMetadata()

	if got, want := string(xmpMetadata[:]), "test xmp metadata"; got != want {
		t.Errorf("invalid xmpMetadata: got=%v, want=%v", got, want)
	}
}

func TestGetXY(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetXY(42, 4.13)

	x, y := pdf.GetXY()

	if got, want := x, 42.0; !floatEqual(got, want) {
		t.Errorf("invalid x coordinate: got=%v, want=%v", got, want)
	}
	if got, want := y, 4.13; !floatEqual(got, want) {
		t.Errorf("invalid y coordinate: got=%v, want=%v", got, want)
	}
}

func TestGetY(t *testing.T) {
	pdf := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	pdf.SetY(4.13)

	y := pdf.GetY()

	if got, want := y, 4.13; !floatEqual(got, want) {
		t.Errorf("invalid y coordinate: got=%v, want=%v", got, want)
	}
}
