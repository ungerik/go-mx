package mx

import (
	"errors"
	"reflect"
	"testing"
)

type fullChain struct {
	called string
}

func (f *fullChain) Normalize() []error  { f.called = "Normalize-slice"; return nil }
func (f *fullChain) NormalizeErr() error { panic("should not call") }

type singleNormalize struct {
	called string
}

func (s *singleNormalize) Normalize() error {
	s.called = "Normalize-err"
	return errors.New("normalized failure")
}

type valueValidate struct {
	called bool
	fail   bool
}

func (v valueValidate) Validate() error {
	if v.fail {
		return errors.New("validate failure")
	}
	return nil
}

type validBool struct {
	ok bool
}

func (v validBool) Valid() bool { return v.ok }

func TestRunValidationChain_Normalize_SlicePreferred(t *testing.T) {
	v := &fullChain{}
	errs := RunValidationChain(reflect.ValueOf(v).Elem())
	if len(errs) != 0 {
		t.Errorf("got errs=%v, want none (Normalize-slice returned nil)", errs)
	}
	if v.called != "Normalize-slice" {
		t.Errorf("called=%q, want Normalize-slice", v.called)
	}
}

func TestRunValidationChain_NormalizeErr(t *testing.T) {
	v := &singleNormalize{}
	errs := RunValidationChain(reflect.ValueOf(v).Elem())
	if len(errs) != 1 || errs[0].Error() != "normalized failure" {
		t.Errorf("got %v, want [normalized failure]", errs)
	}
	if v.called != "Normalize-err" {
		t.Errorf("called=%q, want Normalize-err", v.called)
	}
}

func TestRunValidationChain_Validate(t *testing.T) {
	pass := valueValidate{}
	errs := RunValidationChain(reflect.ValueOf(pass))
	if len(errs) != 0 {
		t.Errorf("pass: got %v, want none", errs)
	}

	addressable := valueValidate{fail: true}
	errs = RunValidationChain(reflect.ValueOf(&addressable).Elem())
	if len(errs) != 1 || errs[0].Error() != "validate failure" {
		t.Errorf("fail: got %v, want [validate failure]", errs)
	}
}

func TestRunValidationChain_Valid(t *testing.T) {
	pass := validBool{ok: true}
	if errs := RunValidationChain(reflect.ValueOf(pass)); len(errs) != 0 {
		t.Errorf("ok: got %v, want none", errs)
	}

	fail := validBool{}
	errs := RunValidationChain(reflect.ValueOf(fail))
	if len(errs) != 1 || errs[0].Error() != "invalid value" {
		t.Errorf("fail: got %v, want [invalid value]", errs)
	}
}

func TestRunValidationChain_NoMethod(t *testing.T) {
	type plain struct{ X int }
	if errs := RunValidationChain(reflect.ValueOf(plain{})); errs != nil {
		t.Errorf("plain struct: got %v, want nil", errs)
	}
}

func TestRunValidationChain_InvalidValue(t *testing.T) {
	if errs := RunValidationChain(reflect.Value{}); errs != nil {
		t.Errorf("zero Value: got %v, want nil", errs)
	}
}
