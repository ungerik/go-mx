package pdf

import (
	"strings"
	"testing"

	"github.com/domonda/go-errs"
)

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
