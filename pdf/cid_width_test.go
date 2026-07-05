package pdf

import (
	"strings"
	"testing"
)

func cidFontWidthArray(font *fontDefType, lastRune int) string {
	r := New(OrientationPortrait, UnitMillimeter, PageSizeA4, "")
	r.generateCIDFontMap(font, lastRune)
	return normalizeWOutput(r.buffer.String())
}

func normalizeWOutput(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}

func TestCIDWidthRangeEntryCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		widths   []int
		interval bool
		want     int
	}{
		{name: "empty", want: 0},
		{name: "widths only", widths: []int{500, 600}, want: 2},
		{name: "widths and interval", widths: []int{700, 700}, interval: true, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &cidWidthRange{widths: tt.widths, interval: tt.interval}
			if got := r.entryCount(); got != tt.want {
				t.Fatalf("entryCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMergeCIDWidthRanges(t *testing.T) {
	t.Parallel()

	t.Run("both nil", func(t *testing.T) {
		t.Parallel()
		got := mergeCIDWidthRanges(nil, nil)
		if got == nil || len(got.widths) != 0 || got.interval {
			t.Fatalf("merge(nil,nil) = %#v, want empty range", got)
		}
	})

	t.Run("append widths and copy interval", func(t *testing.T) {
		t.Parallel()
		a := &cidWidthRange{widths: []int{500, 600}}
		b := &cidWidthRange{widths: []int{700, 700}, interval: true}
		got := mergeCIDWidthRanges(a, b)
		want := &cidWidthRange{widths: []int{500, 600, 700, 700}, interval: true}
		if got.interval != want.interval || !slicesEqual(got.widths, want.widths) {
			t.Fatalf("merge = %#v, want %#v", got, want)
		}
	})

	t.Run("keep existing interval", func(t *testing.T) {
		t.Parallel()
		a := &cidWidthRange{widths: []int{500}, interval: true}
		b := &cidWidthRange{widths: []int{600}, interval: true}
		got := mergeCIDWidthRanges(a, b)
		if !got.interval || !slicesEqual(got.widths, []int{500, 600}) {
			t.Fatalf("merge = %#v, want widths [500 600] with interval", got)
		}
	})

	t.Run("clone when other side nil", func(t *testing.T) {
		t.Parallel()
		a := &cidWidthRange{widths: []int{500, 600}, interval: true}
		got := mergeCIDWidthRanges(a, nil)
		if got.interval != a.interval || !slicesEqual(got.widths, a.widths) {
			t.Fatalf("merge(a,nil) = %#v, want clone of a", got)
		}
		got.widths[0] = 999
		if a.widths[0] == 999 {
			t.Fatal("merge should clone widths, not alias backing array")
		}
	})
}

func TestGenerateCIDFontMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		widths   map[int]int
		lastRune int
		used     map[int]int
		want     string
		notWant  string
	}{
		{
			name: "interval snapshot keeps separate range",
			widths: map[int]int{
				1: 500,
				2: 600,
				3: 700,
				4: 700,
				6: 800,
			},
			lastRune: 6,
			want:     "/W [ 1 [ 500 600 700 700 ] 6 6 800 ]",
			notWant:  "700 700 800",
		},
		{
			name: "single width run compresses",
			widths: map[int]int{
				1: 500,
				2: 500,
				3: 500,
			},
			lastRune: 3,
			want:     "/W [ 1 3 500 ]",
		},
		{
			name: "skips zero width cids",
			widths: map[int]int{
				1: 500,
				3: 600,
			},
			lastRune: 3,
			want:     "/W [ 1 1 500 3 3 600 ]",
		},
		{
			name: "skips unused high cids without usedRunes entry",
			widths: map[int]int{
				256: 500,
				258: 600,
			},
			lastRune: 258,
			used: map[int]int{
				256: 1,
				258: 1,
			},
			want: "/W [ 256 256 500 258 258 600 ]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cw := make([]int, tt.lastRune+1)
			for cid, width := range tt.widths {
				cw[cid] = width
			}
			font := &fontDefType{
				Cw:        cw,
				usedRunes: tt.used,
			}

			got := cidFontWidthArray(font, tt.lastRune)
			if got != tt.want {
				t.Fatalf("generateCIDFontMap output = %q, want %q", got, tt.want)
			}
			if tt.notWant != "" && strings.Contains(got, tt.notWant) {
				t.Fatalf("generateCIDFontMap output = %q, must not contain %q", got, tt.notWant)
			}
		})
	}
}

func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
