package pdf

import (
	"strings"
	"testing"

	"github.com/domonda/go-errs"
)

// TestUnicodeTranslatorEmbeddedMaps loads every embedded code-page map and
// checks a known translation, guarding the map files against the strict
// line-format handling of UnicodeTranslator (any malformed line is an error).
func TestUnicodeTranslatorEmbeddedMaps(t *testing.T) {
	for _, cp := range []string{"", "cp1250", "cp1252"} {
		r := New("portrait", "mm", "A4", "")
		translate := r.UnicodeTranslatorFromDescriptor(cp)
		if err := r.Error(); err != nil {
			t.Fatalf("load %q map: %v", cp, err)
		}
		if got := translate("abc"); got != "abc" {
			t.Errorf("%q map: ASCII %q changed to %q", cp, "abc", got)
		}
	}

	r := New("portrait", "mm", "A4", "")
	translate := r.UnicodeTranslatorFromDescriptor("") // cp1252
	if err := r.Error(); err != nil {
		t.Fatalf("load cp1252 map: %v", err)
	}
	if got := translate("€é"); got != "\x80\xe9" {
		t.Errorf("cp1252: translate(%q) = %x, want 80e9", "€é", got)
	}
}

// TestUnicodeTranslatorErrors checks that malformed map lines and reader
// failures surface as errors and yield the identity translator.
func TestUnicodeTranslatorErrors(t *testing.T) {
	translate, err := UnicodeTranslator(strings.NewReader("!80 U+20AC Euro\nnot a map line\n!82 U+201A quotesinglbase\n"))
	if err == nil {
		t.Fatal("malformed line did not return an error")
	}
	if got := translate("€"); got != "€" {
		t.Errorf("translator after error is not the identity: %q", got)
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
