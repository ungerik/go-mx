package pdf

import (
	"strings"
	"testing"
)

func TestOutputIntentSubtypeEscaping(t *testing.T) {
	for _, tt := range []struct {
		name    string
		subtype OutputIntentSubtype
		want    string
	}{
		{"PDFX", OutputIntent_GTS_PDFX, "GTS_PDFX"},
		{"PDFA", OutputIntent_GTS_PDFA1, "GTS_PDFA1"},
		{"PDFE", OutputIntent_GTS_PDFE1, "GTS_PDFE1"},
		{"extension", "Vendor_Extension-1.0", "Vendor_Extension-1.0"},
		{"space", "Vendor Extension", "Vendor#20Extension"},
		{"delimiters", "Vendor/%()<>[]{}", "Vendor#2F#25#28#29#3C#3E#5B#5D#7B#7D"},
		{"hash", "Vendor#20Extension", "Vendor#2320Extension"},
		{"whitespace", "Vendor\t\n\r\fExtension", "Vendor#09#0A#0D#0CExtension"},
		{"non-ASCII", "Vendoré\x7f\xff", "Vendor#C3#A9#7F#FF"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
			r.AddPage()
			r.AddOutputIntent(OutputIntent{
				SubtypeIdent:              tt.subtype,
				OutputConditionIdentifier: "test profile",
				ICCProfile:                []byte("test profile data"),
			})
			data := string(renderBytes(t, r))
			want := "<< /Type /OutputIntent /S /" + tt.want + " /OutputConditionIdentifier"
			if !strings.Contains(data, want) {
				t.Errorf("output intent missing escaped subtype: want %q", want)
			}
		})
	}
}

// TestOutputIntentStringsEscaping checks that the strings of an output intent
// are written as PDF literal strings: a parenthesis or backslash is escaped
// instead of ending the string and injecting the rest as dictionary entries.
func TestOutputIntentStringsEscaping(t *testing.T) {
	r := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	r.AddPage()
	r.AddOutputIntent(OutputIntent{
		SubtypeIdent:              OutputIntent_GTS_PDFA1,
		OutputConditionIdentifier: `sRGB (IEC) \ v2`,
		Info:                      `Info) /Evil (x`,
		ICCProfile:                []byte("test profile data"),
	})
	data := string(renderBytes(t, r))
	for _, want := range []string{
		`/OutputConditionIdentifier (sRGB \(IEC\) \\ v2) `,
		`/Info (Info\) /Evil \(x) `,
	} {
		if !strings.Contains(data, want) {
			t.Errorf("output intent missing escaped string %q", want)
		}
	}
}

// TestOutputIntentStringsEncrypted checks that a protected document encrypts
// the strings of an output intent with the object number of the catalog that
// holds them, like every other string, so a reader decrypts the original.
func TestOutputIntentStringsEncrypted(t *testing.T) {
	r := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	r.SetProtection(0, "", "owner")
	r.AddPage()
	r.AddOutputIntent(OutputIntent{
		SubtypeIdent:              OutputIntent_GTS_PDFA1,
		OutputConditionIdentifier: "sRGB IEC61966-2.1",
		Info:                      "Info (with parentheses)",
		ICCProfile:                []byte("test profile data"),
	})
	data := string(renderBytes(t, r))

	_, n := objectHeader(t, data, "/Type /Catalog")
	for key, want := range map[string]string{
		"/OutputConditionIdentifier": "sRGB IEC61966-2.1",
		"/Info":                      "Info (with parentheses)",
	} {
		encrypted, ok := literalStringAfter(data, "/Type /OutputIntent ", key+" (")
		if !ok {
			t.Errorf("no %s string in the output intent", key)
			continue
		}
		if string(encrypted) == want {
			t.Errorf("%s is not encrypted", key)
			continue
		}
		if got := decryptString(t, r, n, encrypted); got != want {
			t.Errorf("decrypted %s = %q, want %q", key, got, want)
		}
	}
}
