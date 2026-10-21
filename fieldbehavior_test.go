package mx

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestFieldErrors_Error(t *testing.T) {
	fe := FieldErrors{
		"Branding-PrimaryColor": errors.New("must be hex"),
		"AccountingEmail":       errors.New("invalid address"),
	}
	got := fe.Error()
	// sorted by path, joined by newline (errors.Join semantics)
	want := "invalid address\nmust be hex"
	if got != want {
		t.Errorf("FieldErrors.Error() =\n%q\nwant\n%q", got, want)
	}
}

// KeepsStoredValue is the one rule the form handler and both renderers
// share: only a sensitive value of a kind with a "keep" blank that is
// neither zero nor null is kept. sql.NullString{String: "x"} is null but
// not zero, so nothing is kept, the clear control renders ticked and a
// blank is parsed.
func TestKeepsStoredValue(t *testing.T) {
	sensitive := FormTag{Sensitive: true}
	for _, tc := range []struct {
		name  string
		tag   FormTag
		kind  FieldKind
		value any
		want  bool
	}{
		{"stored text", sensitive, FieldKindString, "secret", true},
		{"stored number", sensitive, FieldKindNumber, 7, true},
		{"empty text", sensitive, FieldKindString, "", false},
		{"null but not zero", sensitive, FieldKindString, sql.NullString{String: "x"}, false},
		{"valid null type", sensitive, FieldKindString, sql.NullString{String: "x", Valid: true}, true},
		{"nil pointer", sensitive, FieldKindNumber, (*int)(nil), false},
		{"not sensitive", FormTag{}, FieldKindString, "secret", false},
		{"checkbox has its own blank", sensitive, FieldKindBool, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := KeepsStoredValue(tc.tag, tc.kind, reflect.ValueOf(tc.value)); got != tc.want {
				t.Errorf("KeepsStoredValue = %v, want %v", got, tc.want)
			}
		})
	}
	if KeepsStoredValue(sensitive, FieldKindString, reflect.Value{}) {
		t.Error("an invalid value keeps nothing")
	}
}

func TestFieldErrors_EmptyIsEmptyString(t *testing.T) {
	if got := (FieldErrors{}).Error(); got != "" {
		t.Errorf("empty FieldErrors.Error() = %q, want \"\"", got)
	}
}

func TestFieldErrors_UnwrapAllowsErrorsIs(t *testing.T) {
	sentinel := errors.New("specific")
	fe := FieldErrors{
		"A": sentinel,
		"B": errors.New("other"),
	}
	if !errors.Is(fe, sentinel) {
		t.Errorf("errors.Is should find sentinel via Unwrap()")
	}
}

// A nil entry is no error, so Unwrap leaves it out: a caller ranging over
// the parts never gets a nil error.
func TestFieldErrors_UnwrapDropsNil(t *testing.T) {
	e := errors.New("bad")
	got := FieldErrors{"A": nil, "B": e}.Unwrap()
	if len(got) != 1 || got[0] != e {
		t.Errorf("Unwrap() = %v, want only the non-nil entry", got)
	}
}

// A render probe is recognized as one, and nothing else is: not a caller's
// error with the same text, not the probe joined with another error.
func TestIsRenderProbe(t *testing.T) {
	if !IsRenderProbe(renderProbe()) {
		t.Error("the handler's probe is not recognized")
	}
	for _, errs := range [][]error{nil, {}, {errors.New("mx: render probe")}, append(renderProbe(), errors.New("x"))} {
		if IsRenderProbe(errs) {
			t.Errorf("%v must not be a render probe", errs)
		}
	}
}

func TestDeciderFromContext_Unconfigured(t *testing.T) {
	d := DeciderFromContext(context.Background())
	if d == nil {
		t.Fatal("expected non-nil decider")
	}
	beh := d("X", reflect.StructField{Name: "X"}, reflect.Value{})
	if beh.Render == nil || beh.Parse == nil || beh.Validate == nil {
		t.Fatalf("expected all three closures set, got %+v", beh)
	}
	err := beh.Parse("X", reflect.StructField{Name: "X"}, reflect.Value{}, nil)
	if err == nil {
		t.Errorf("Parse should return a clear error when unconfigured")
	}
	if !strings.Contains(err.Error(), "Middleware") {
		t.Errorf("error %q should mention Middleware", err)
	}
}

func TestSetField_ResolvesNamedField(t *testing.T) {
	type inner struct {
		Color string
	}
	type outer struct {
		Branding inner
	}
	o := &outer{}
	v, f, err := SetField(reflect.ValueOf(o), "Branding-Color")
	if err != nil {
		t.Fatalf("SetField: %v", err)
	}
	if f.Name != "Color" {
		t.Errorf("got field %q, want Color", f.Name)
	}
	if !v.CanSet() {
		t.Errorf("returned value should be settable")
	}
	v.SetString("blue")
	if o.Branding.Color != "blue" {
		t.Errorf("write through path failed: got %q", o.Branding.Color)
	}
}

func TestSetField_RejectsMissing(t *testing.T) {
	type s struct{ A string }
	_, _, err := SetField(reflect.ValueOf(&s{}), "Missing")
	if err == nil {
		t.Errorf("expected error for missing field")
	}
}
