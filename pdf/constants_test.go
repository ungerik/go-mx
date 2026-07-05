package pdf

import "testing"

func TestUnitNormalize(t *testing.T) {
	tests := []struct {
		in   Unit
		want Unit
		ok   bool
	}{
		{UnitMillimeter, UnitMillimeter, true},
		{Unit("MM"), UnitMillimeter, true},
		{Unit("point"), UnitPoint, true},
		{Unit("PT"), UnitPoint, true},
		{Unit("in"), UnitInch, true},
		{Unit("INCH"), UnitInch, true},
		{Unit("cm"), UnitCentimeter, true},
		{Unit("bogus"), "", false},
	}
	for _, tc := range tests {
		got, ok := tc.in.normalize()
		if ok != tc.ok || got != tc.want {
			t.Errorf("Unit(%q).normalize() = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestPageSizeNormalize(t *testing.T) {
	tests := []struct {
		in   PageSize
		want PageSize
		ok   bool
	}{
		{PageSizeA4, PageSizeA4, true},
		{PageSize("a4"), PageSizeA4, true},
		{PageSize("LETTER"), PageSizeLetter, true},
		{PageSize("legal"), PageSizeLegal, true},
		{PageSize("tabloid"), PageSizeTabloid, true},
		{PageSize("A3"), PageSizeA3, true},
		{PageSize("unknown"), "", false},
	}
	for _, tc := range tests {
		got, ok := tc.in.normalize()
		if ok != tc.ok || got != tc.want {
			t.Errorf("PageSize(%q).normalize() = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestPageSizeSizeType(t *testing.T) {
	sz, ok := PageSizeA4.SizeType()
	if !ok {
		t.Fatal("PageSizeA4.SizeType() ok = false")
	}
	if !floatEqual(sz.Wd, 595.28) || !floatEqual(sz.Ht, 841.89) {
		t.Errorf("PageSizeA4 in pt: got %v×%v, want ~595.28×841.89", sz.Wd, sz.Ht)
	}

	_, ok = PageSize("bogus").SizeType()
	if ok {
		t.Error("bogus page size should not have SizeType")
	}
}

func TestNewAcceptsLegacyUnitAndPageSizeStrings(t *testing.T) {
	r := New(OrientationPortrait, Unit("point"), PageSize("a4"), "")
	if r.Err() {
		t.Fatalf("New with legacy strings: %v", r.Error())
	}
	if r.unit != UnitPoint {
		t.Errorf("unit = %q, want %q", r.unit, UnitPoint)
	}
	w, h := r.GetPageSize()
	if !floatEqual(w, 595.28) || !floatEqual(h, 841.89) {
		t.Errorf("A4 in pt: got %v×%v, want ~595.28×841.89", w, h)
	}

	r = New(OrientationPortrait, Unit("in"), PageSize("letter"), "")
	if r.Err() {
		t.Fatalf("New with inch/letter: %v", r.Error())
	}
	if r.unit != UnitInch {
		t.Errorf("unit = %q, want %q", r.unit, UnitInch)
	}
}
