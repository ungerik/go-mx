package pdf

import (
	"strings"
	"testing"
	"time"

	"github.com/domonda/go-errs"
)

// pdfDate feeds the /CreationDate, /ModDate and attachment /Params dates that
// PDF/A validators compare against the XMP packet, so the UTC offset must be
// rendered exactly: a negative offset flips the sign but keeps positive
// hour/minute fields, and without a timezone the legacy short form is kept.
func TestPDFDate(t *testing.T) {
	tm := time.Date(2024, 5, 6, 7, 8, 9, 0, time.FixedZone("", -(5*3600+30*60)))
	if got, want := pdfDate(tm, true), "D:20240506070809-05'30'"; got != want {
		t.Errorf("pdfDate(-05:30) = %q, want %q", got, want)
	}
	if got, want := pdfDate(tm, false), "D:20240506070809"; got != want {
		t.Errorf("pdfDate without timezone = %q, want %q", got, want)
	}
	tm = time.Date(2024, 5, 6, 7, 8, 9, 0, time.FixedZone("", 2*3600))
	if got, want := pdfDate(tm, true), "D:20240506070809+02'00'"; got != want {
		t.Errorf("pdfDate(+02:00) = %q, want %q", got, want)
	}
}

// pdfDocEncode backs the filespec /F entry, which must stay single-byte:
// Latin-1 code points keep their byte value while anything else degrades to
// '?' — the exact Unicode name is carried by /UF, so a lossy but valid /F
// beats emitting raw UTF-8 bytes.
func TestPDFDocEncode(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{"invoice.xml", "invoice.xml"},     // ASCII fast path
		{"faktüra.xml", "fakt\xfcra.xml"},  // Latin-1 keeps its byte value
		{"price€.txt", "price?.txt"},       // above U+00FF
		{"faktüra\x09v2", "fakt\xfcra?v2"}, // control rune (tab)
		{"a\u0080b\u00a0c", "a?b?c"},       // U+0080 and U+00A0, the 0x7F–0xA0 gap
	} {
		if got := pdfDocEncode(tt.in); got != tt.want {
			t.Errorf("pdfDocEncode(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// pdfTextToUTF8 recovers the UTF-8 text of a metadata string stored on the
// renderer, whichever of the two setter encodings produced it, so applyXMP
// can carry a caller-set /Info value into the XMP packet. The encoding is
// declared, not sniffed: a Latin-1 value legitimately starting with the
// bytes "þÿ" (a fake BOM) must survive undamaged. Malformed values degrade
// predictably: an odd trailing byte after the BOM is dropped, and a BOM-only
// string is empty.
func TestPDFTextToUTF8(t *testing.T) {
	if got, want := pdfTextToUTF8(utf8toutf16("Prodücer 😀"), true), "Prodücer 😀"; got != want {
		t.Errorf("UTF-16BE round trip = %q, want %q", got, want)
	}
	if got, want := pdfTextToUTF8("Prod\xfccer", false), "Prodücer"; got != want {
		t.Errorf("Latin-1 decode = %q, want %q", got, want)
	}
	if got, want := pdfTextToUTF8("\xfe\xffLegacy", false), "þÿLegacy"; got != want {
		t.Errorf("Latin-1 with fake BOM = %q, want %q", got, want)
	}
	if got, want := pdfTextToUTF8("\xFE\xFF\x00A\x00", true), "A"; got != want {
		t.Errorf("odd-length UTF-16 body = %q, want %q", got, want)
	}
	if got := pdfTextToUTF8("\xFE\xFF", true); got != "" {
		t.Errorf("BOM-only string = %q, want empty", got)
	}
}

// TestUnicodeTranslatorEmbeddedMaps loads every embedded code-page map and
// checks a known translation.
func TestUnicodeTranslatorEmbeddedMaps(t *testing.T) {
	for _, cp := range []string{"", "cp1250", "cp1252"} {
		r := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
		translate := r.UnicodeTranslatorFromDescriptor(cp)
		if err := r.Error(); err != nil {
			t.Fatalf("load %q map: %v", cp, err)
		}
		if got := translate("abc"); got != "abc" {
			t.Errorf("%q map: ASCII %q changed to %q", cp, "abc", got)
		}
	}

	r := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	translate := r.UnicodeTranslatorFromDescriptor("") // cp1252
	if err := r.Error(); err != nil {
		t.Fatalf("load cp1252 map: %v", err)
	}
	if got := translate("€é"); got != "\x80\xe9" {
		t.Errorf("cp1252: translate(%q) = %x, want 80e9", "€é", got)
	}
}

// TestUnicodeTranslatorErrors checks that malformed map lines are skipped
// (matching the tolerance of the legacy engine towards comment/header lines
// in user-supplied .map files) while reader failures surface as errors and
// yield the identity translator.
func TestUnicodeTranslatorErrors(t *testing.T) {
	translate, err := UnicodeTranslator(strings.NewReader("!80 U+20AC Euro\nnot a map line\n!82 U+201A quotesinglbase\n"))
	if err != nil {
		t.Fatalf("malformed line was not skipped: %v", err)
	}
	if got := translate("€‚"); got != "\x80\x82" {
		t.Errorf("translate(%q) = %x, want 8082 (entries around the malformed line must load)", "€‚", got)
	}

	translate, err = UnicodeTranslator(failingReader{})
	if err == nil {
		t.Fatal("reader failure did not return an error")
	}
	if got := translate("€"); got != "€" {
		t.Errorf("translator after reader failure is not the identity: %q", got)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errs.New("failing reader")
}
