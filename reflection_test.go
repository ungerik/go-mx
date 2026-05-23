package mx

import (
	"reflect"
	"testing"
)

type rfInner struct {
	X int    `attr:"x"`
	Y string `attr:"y"`
	z int    //nolint:unused // unexported, intentionally present to test that ReflectStructFields skips it
}

type rfOuter struct {
	A bool `attr:"a"`
	rfInner
	B *int `attr:"b"`
}

type rfPtrEmbed struct {
	A bool `attr:"a"`
	*rfInner
	B int `attr:"b"`
}

type rfDeep struct {
	rfOuter
	C float64 `attr:"c"`
}

// collect drives the iterator and returns the yielded field names + values in
// order, for compact assertions.
func collect(s any) (names []string, values []any) {
	for f, v := range ReflectStructFields(reflect.ValueOf(s)) {
		names = append(names, f.Name)
		values = append(values, v.Interface())
	}
	return
}

func TestReflectStructFields_Flat(t *testing.T) {
	type S struct {
		A int
		B string
		c bool // unexported — must be skipped
	}
	names, values := collect(S{A: 1, B: "two", c: true})
	wantNames := []string{"A", "B"}
	wantValues := []any{1, "two"}
	if !reflect.DeepEqual(names, wantNames) {
		t.Errorf("names = %v, want %v", names, wantNames)
	}
	if !reflect.DeepEqual(values, wantValues) {
		t.Errorf("values = %v, want %v", values, wantValues)
	}
}

func TestReflectStructFields_EmbeddedFlattens(t *testing.T) {
	i := 7
	o := rfOuter{A: true, rfInner: rfInner{X: 1, Y: "hi", z: 9}, B: &i}
	names, _ := collect(o)
	// Anonymous parent (rfInner) and unexported z must not appear; promoted
	// X, Y must appear in their declaration position between A and B.
	want := []string{"A", "X", "Y", "B"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("names = %v, want %v", names, want)
	}
}

func TestReflectStructFields_DeeplyNested(t *testing.T) {
	d := rfDeep{rfOuter: rfOuter{A: true, rfInner: rfInner{X: 1, Y: "hi"}}, C: 3.14}
	names, _ := collect(d)
	want := []string{"A", "X", "Y", "B", "C"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("names = %v, want %v", names, want)
	}
}

func TestReflectStructFields_PointerArgument(t *testing.T) {
	type S struct{ X int }
	s := &S{X: 42}
	names, values := collect(s)
	if !reflect.DeepEqual(names, []string{"X"}) || !reflect.DeepEqual(values, []any{42}) {
		t.Errorf("pointer arg should dereference; got names=%v values=%v", names, values)
	}
}

func TestReflectStructFields_NilEmbeddedPointerSkipped(t *testing.T) {
	// rfInner is the nil embedded *rfInner. Current code (pre-rewrite) panicked
	// here; the VisibleFields version skips the subtree silently.
	p := rfPtrEmbed{A: true, B: 42} // .rfInner is nil
	names, values := collect(p)
	want := []string{"A", "B"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("names = %v, want %v (X from nil *rfInner should be skipped)", names, want)
	}
	if !reflect.DeepEqual(values, []any{true, 42}) {
		t.Errorf("values = %v, want [true 42]", values)
	}
}

func TestReflectStructFields_FieldShadowing(t *testing.T) {
	// When two embeds both declare X, Go's visibility rules say neither is
	// promoted to the outer struct. reflect.VisibleFields enforces this; the
	// pre-rewrite recursion would have yielded both.
	type A struct{ X int }
	type B struct{ X int }
	type Both struct {
		A
		B
	}
	names, _ := collect(Both{})
	for _, n := range names {
		if n == "X" {
			t.Errorf("shadowed field X should not be yielded; got names = %v", names)
		}
	}
}

func TestReflectStructFields_EarlyBreak(t *testing.T) {
	type S struct {
		A, B, C int
	}
	seen := 0
	for range ReflectStructFields(reflect.ValueOf(S{1, 2, 3})) {
		seen++
		if seen == 2 {
			break
		}
	}
	if seen != 2 {
		t.Errorf("break should stop iteration at 2, got %d", seen)
	}
}

func TestReflectStructFields_PanicsOnNilPointer(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on nil pointer argument")
		}
	}()
	var s *struct{ X int }
	for range ReflectStructFields(reflect.ValueOf(s)) {
	}
}

func TestReflectStructFields_PanicsOnNonStruct(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic on non-struct argument")
		}
	}()
	for range ReflectStructFields(reflect.ValueOf(42)) {
	}
}
