package mx

import (
	"context"
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

func TestDeciderFromContext_NilContext(t *testing.T) {
	// Intentional nil to verify the nil-safety guard. Using a typed
	// variable rather than the literal nil so static analyzers do not
	// flag the call (SA1012); the runtime check is the same.
	var ctx context.Context
	d := DeciderFromContext(ctx)
	if d == nil {
		t.Errorf("expected non-nil decider for nil context")
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
