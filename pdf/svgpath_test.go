package pdf

import (
	"math"
	"strconv"
	"testing"
)

// pathString renders the recorded segments in a compact readable form for
// comparison in tests.
func pathString(p *recordedPath) string {
	s := ""
	num := func(f float64) string {
		// Round to 4 decimals so Bézier math comparisons are stable.
		return strconv.FormatFloat(math.Round(f*1e4)/1e4, 'f', -1, 64)
	}
	for _, seg := range p.segs {
		switch seg.op {
		case 'M':
			s += "M" + num(seg.coords[0]) + "," + num(seg.coords[1])
		case 'L':
			s += "L" + num(seg.coords[0]) + "," + num(seg.coords[1])
		case 'C':
			s += "C" + num(seg.coords[0]) + "," + num(seg.coords[1]) +
				" " + num(seg.coords[2]) + "," + num(seg.coords[3]) +
				" " + num(seg.coords[4]) + "," + num(seg.coords[5])
		case 'Z':
			s += "Z"
		}
	}
	return s
}

func TestRenderSVGPathData(t *testing.T) {
	tests := []struct {
		name string
		d    string
		want string
	}{
		{
			name: "absolute move and lines",
			d:    "M10 20 L30 40 H50 V60",
			want: "M10,20L30,40L50,40L50,60",
		},
		{
			name: "relative move and lines",
			d:    "m10 20 l5 5 h10 v-10",
			want: "M10,20L15,25L25,25L25,15",
		},
		{
			name: "implicit lineto after moveto",
			d:    "M0 0 10 0 10 10",
			want: "M0,0L10,0L10,10",
		},
		{
			name: "implicit relative lineto after relative moveto",
			d:    "m5 5 10 0 0 10",
			want: "M5,5L15,5L15,15",
		},
		{
			name: "close resets to subpath start",
			d:    "M10 10 L20 10 L20 20 Z L10 30",
			want: "M10,10L20,10L20,20ZL10,30",
		},
		{
			name: "comma separators and compact numbers",
			d:    "M1.5.5-2,3L4,5",
			want: "M1.5,0.5L-2,3L4,5",
		},
		{
			name: "cubic curve",
			d:    "M0 0 C10 0 20 10 20 20",
			want: "M0,0C10,0 20,10 20,20",
		},
		{
			name: "smooth cubic reflects previous control",
			d:    "M0 0 C10 0 20 10 20 20 S30 40 40 40",
			want: "M0,0C10,0 20,10 20,20C20,30 30,40 40,40",
		},
		{
			name: "smooth cubic without previous cubic uses current point",
			d:    "M10 10 S20 0 30 10",
			want: "M10,10C10,10 20,0 30,10",
		},
		{
			name: "quadratic promoted to cubic",
			d:    "M0 0 Q15 0 30 0",
			want: "M0,0C10,0 20,0 30,0",
		},
		{
			name: "smooth quadratic reflects previous control",
			d:    "M0 0 Q15 30 30 0 T60 0",
			want: "M0,0C10,20 20,20 30,0C40,-20 50,-20 60,0",
		},
		{
			name: "zero-radius arc degenerates to line",
			d:    "M0 0 A0 10 0 0 1 20 0",
			want: "M0,0L20,0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rec recordedPath
			err := renderSVGPathData(tt.d, &rec)
			if err != nil {
				t.Fatalf("renderSVGPathData(%q): %v", tt.d, err)
			}
			if got := pathString(&rec); got != tt.want {
				t.Errorf("renderSVGPathData(%q)\n got %s\nwant %s", tt.d, got, tt.want)
			}
		})
	}
}

func TestRenderSVGPathData_arc(t *testing.T) {
	// A half circle of radius 10 from (0,0) to (20,0), sweeping below the
	// x axis (SVG y grows downward, sweep flag 1 is the positive-angle
	// direction).
	var rec recordedPath
	err := renderSVGPathData("M0 0 A10 10 0 0 1 20 0", &rec)
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.segs) != 3 { // moveto + two 90° cubic segments
		t.Fatalf("expected 3 segments, got %d: %s", len(rec.segs), pathString(&rec))
	}
	for i, seg := range rec.segs[1:] {
		if seg.op != 'C' {
			t.Fatalf("segment %d: expected cubic, got %c", i+1, seg.op)
		}
	}
	end := rec.segs[len(rec.segs)-1]
	if end.coords[4] != 20 || end.coords[5] != 0 {
		t.Errorf("arc endpoint = (%g, %g), want (20, 0)", end.coords[4], end.coords[5])
	}
	// The arc's midpoint (10, -10) must be the junction of the two segments.
	mid := rec.segs[1]
	if math.Abs(mid.coords[4]-10) > 1e-9 || math.Abs(mid.coords[5]+10) > 1e-9 {
		t.Errorf("arc midpoint = (%g, %g), want (10, -10)", mid.coords[4], mid.coords[5])
	}
}

func TestRenderSVGPathData_errors(t *testing.T) {
	for _, d := range []string{
		"10 20 L30 40",           // must start with a command
		"M10",                    // missing coordinate
		"M10 20 L",               // missing coordinates
		"M10 20 X30 40",          // unknown command
		"M0 0 A10 10 0 2 1 20 0", // invalid arc flag
	} {
		var rec recordedPath
		if err := renderSVGPathData(d, &rec); err == nil {
			t.Errorf("renderSVGPathData(%q): expected error, got nil", d)
		}
	}
}

func TestParseSVGNumberList(t *testing.T) {
	got, err := parseSVGNumberList(" 1.5.5-2,3 ,4e1\n5")
	if err != nil {
		t.Fatal(err)
	}
	want := []float64{1.5, 0.5, -2, 3, 40, 5}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if _, err = parseSVGNumberList("1 2 x"); err == nil {
		t.Error("expected error for invalid number list")
	}
}
