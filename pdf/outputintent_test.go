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
