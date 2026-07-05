package pdf

import (
	"strings"
	"testing"
)

// TestAddUTF8FontFromBytesMalformed feeds truncated and corrupt TrueType data
// into the font loader and requires a recorded error instead of a panic —
// font bytes may come from untrusted sources.
func TestAddUTF8FontFromBytesMalformed(t *testing.T) {
	cases := map[string][]byte{
		"empty":            {},
		"short magic":      {0x00, 0x01},
		"magic only":       {0x00, 0x01, 0x00, 0x00},
		"truncated header": {0x00, 0x01, 0x00, 0x00, 0x00, 0x02},
		"garbage tables":   append([]byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0, 0, 0, 0, 0, 0}, []byte(strings.Repeat("x", 40))...),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			r := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
			r.AddUTF8FontFromBytes("bad", "", data) // must not panic
			if r.Error() == nil {
				t.Error("malformed font did not set the renderer error")
			}
		})
	}
}

// TestUTF8CutFontMalformed requires the public subsetting utility to return an
// error instead of panicking on corrupt input.
func TestUTF8CutFontMalformed(t *testing.T) {
	if _, err := UTF8CutFont([]byte{0x00, 0x01, 0x00, 0x00, 0xFF}, "ab"); err == nil {
		t.Error("UTF8CutFont on truncated font returned no error")
	}
}
