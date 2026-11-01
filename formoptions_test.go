package mx

import (
	"context"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestSelectPreface(t *testing.T) {
	opts := []NamedOption{{Name: "A", Value: "a"}}
	withEmpty := []NamedOption{{Value: ""}, {Name: "A", Value: "a"}}

	tests := []struct {
		name      string
		value     reflect.Value
		selected  string
		tag       FormTag
		context   bool
		options   []NamedOption
		want      OptionPreface
		wantShown string
	}{
		{name: "optional empty", selected: "", context: true, options: opts, want: OptionPrefaceEmpty},
		{name: "required empty", selected: "", tag: FormTag{Required: true}, context: true, options: opts, want: OptionPrefaceChoose},
		{name: "list already has empty", selected: "", context: true, options: withEmpty, want: OptionPrefaceNone},
		{name: "required list already has empty", selected: "", tag: FormTag{Required: true}, context: true, options: withEmpty, want: OptionPrefaceNone},
		{name: "optional listed value", selected: "a", context: true, options: opts, want: OptionPrefaceClear},
		{name: "required listed value", selected: "a", tag: FormTag{Required: true}, context: true, options: opts, want: OptionPrefaceNone},
		{name: "context out of list", selected: "z", context: true, options: opts, want: OptionPrefaceUnavailable},
		{name: "context out of list even when list has empty", selected: "z", context: true, options: withEmpty, want: OptionPrefaceUnavailable},
		{name: "listed value and list has empty", selected: "a", context: true, options: withEmpty, want: OptionPrefaceNone},
		{name: "sensitive out of list", selected: "z", tag: FormTag{Sensitive: true}, context: true, options: opts, want: OptionPrefaceUnavailable},
		{name: "static out of list", selected: "z", options: opts, want: OptionPrefaceClear},
		{name: "required static out of list", selected: "z", tag: FormTag{Required: true}, options: opts, want: OptionPrefaceUnavailable},
		// A required select rejects the blank, and a readonly one is never
		// saved, so neither placeholder says the value is removed.
		{name: "required context out of list", selected: "z", tag: FormTag{Required: true}, context: true, options: opts, want: OptionPrefaceUnavailable},
		{name: "readonly context out of list", selected: "z", tag: FormTag{Readonly: true}, context: true, options: opts, want: OptionPrefaceUnavailable},
		{name: "empty list with value", selected: "z", context: true, want: OptionPrefaceUnavailable},
		{name: "empty list required", selected: "", tag: FormTag{Required: true}, context: true, want: OptionPrefaceChoose},
		{name: "empty list optional blank", selected: "", context: true, want: OptionPrefaceEmpty},
		{name: "static empty required", selected: "", tag: FormTag{Required: true}, want: OptionPrefaceChoose},
		{name: "sensitive required empty", selected: "", tag: FormTag{Required: true, Sensitive: true}, context: true, options: opts, want: OptionPrefaceChoose},
		// An enum whose zero is none of its options holds no choice at its
		// zero: it shows the empty option, not an unavailable value.
		{name: "unlisted zero", value: reflect.ValueOf(0), selected: "0", options: opts, want: OptionPrefaceEmpty, wantShown: ""},
		{name: "required unlisted zero", value: reflect.ValueOf(0), selected: "0", tag: FormTag{Required: true}, context: true, options: opts, want: OptionPrefaceChoose, wantShown: ""},
		{name: "unlisted zero and list has empty", value: reflect.ValueOf(0), selected: "0", options: withEmpty, want: OptionPrefaceNone, wantShown: ""},
		{name: "listed zero", value: reflect.ValueOf(0), selected: "0", options: []NamedOption{{Value: "0"}, {Value: "1"}}, want: OptionPrefaceClear, wantShown: "0"},
		{name: "unlisted non-zero", value: reflect.ValueOf(7), selected: "7", context: true, options: opts, want: OptionPrefaceUnavailable, wantShown: "7"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, shown := SelectPreface(tt.value, tt.selected, tt.tag, tt.context, tt.options)
			if got != tt.want {
				t.Fatalf("preface = %v, want %v", got, tt.want)
			}
			wantShown := tt.wantShown
			if !tt.value.IsValid() {
				wantShown = tt.selected
			}
			if shown != wantShown {
				t.Fatalf("shown = %q, want %q", shown, wantShown)
			}
		})
	}
}

func TestEnumSetPreface(t *testing.T) {
	opts := []NamedOption{{Name: "A", Value: "a"}}
	tests := []struct {
		name    string
		members []string
		tag     FormTag
		context bool
		options []NamedOption
		want    OptionPreface
	}{
		{name: "all listed", members: []string{"a"}, context: true, options: opts, want: OptionPrefaceNone},
		{name: "missing member", members: []string{"a", "hidden"}, context: true, options: opts, want: OptionPrefaceValuesUnavailable},
		{name: "sensitive missing", members: []string{"hidden"}, tag: FormTag{Sensitive: true}, context: true, options: opts, want: OptionPrefaceNone},
		{name: "static missing", members: []string{"hidden"}, options: opts, want: OptionPrefaceNone},
		{name: "empty list with member", members: []string{"hidden"}, context: true, want: OptionPrefaceValuesUnavailable},
		{name: "empty list required", tag: FormTag{Required: true}, context: true, want: OptionPrefaceNone},
		{name: "empty list optional blank", context: true, want: OptionPrefaceNone},
		{name: "readonly missing", members: []string{"hidden"}, tag: FormTag{Readonly: true}, context: true, options: opts, want: OptionPrefaceValuesUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnumSetPreface(tt.members, tt.tag, tt.context, tt.options)
			if got != tt.want {
				t.Fatalf("preface = %v, want %v", got, tt.want)
			}
		})
	}
}

// Saving drops an unlisted value where the field is saved and a blank is
// accepted: not on a required select, which rejects the blank, nor on a
// readonly field, which is never saved.
func TestOptionPrefaceRemovesValue(t *testing.T) {
	for _, tc := range []struct {
		preface OptionPreface
		tag     FormTag
		want    bool
	}{
		{OptionPrefaceUnavailable, FormTag{}, true},
		{OptionPrefaceUnavailable, FormTag{Required: true}, false},
		{OptionPrefaceUnavailable, FormTag{Readonly: true}, false},
		{OptionPrefaceValuesUnavailable, FormTag{}, true},
		{OptionPrefaceValuesUnavailable, FormTag{Required: true}, true},
		{OptionPrefaceValuesUnavailable, FormTag{Readonly: true}, false},
		{OptionPrefaceClear, FormTag{}, false},
		{OptionPrefaceNone, FormTag{}, false},
	} {
		if got := tc.preface.RemovesValue(tc.tag); got != tc.want {
			t.Errorf("%v.RemovesValue(%+v) = %v, want %v", tc.preface, tc.tag, got, tc.want)
		}
	}
}

// Set members come out in the form option values use and sorted, so both
// renderers look a checkbox up by binary search; a nil set has none.
func TestSetMemberValues(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value any
		want  []string
	}{
		{"map", map[colorEnum]struct{}{"red": {}, "blue": {}}, []string{"blue", "red"}},
		{"slice", []int{3, 1, 2}, []string{"1", "2", "3"}},
		{"pointer to slice", &[]string{"b", "a"}, []string{"a", "b"}},
		{"nil pointer", (*[]string)(nil), nil},
		{"no set", "text", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SetMemberValues(reflect.ValueOf(tc.value)); !slices.Equal(got, tc.want) {
				t.Errorf("SetMemberValues = %q, want %q", got, tc.want)
			}
		})
	}
}

// A static option list is resolved when the form is built, so the hx layer
// can still enrich the option elements; a context list is collected while
// rendering, from the request context, and fails there.
func TestOptionComponents(t *testing.T) {
	render := func(opt NamedOption) Component { return Text(opt.Value + ";") }
	prepend := func(options []NamedOption) Component { return Text("first;") }

	static := OptionComponents(reflect.Value{}, FormTag{}, reflect.TypeFor[colorEnum](), render, prepend)
	children, ok := static.(Components)
	if !ok {
		t.Fatalf("static list is %T, want Components resolved at build time", static)
	}
	if got := renderToString(t, children); got != "first;red;green;blue;" {
		t.Errorf("static list rendered %q", got)
	}
	if got := renderToString(t, OptionComponents(reflect.Value{}, FormTag{}, reflect.TypeFor[colorEnum](), render, nil)); got != "red;green;blue;" {
		t.Errorf("without prepend rendered %q", got)
	}

	contextList := OptionComponents(reflect.Value{}, ParseFormTagString("options=test-partners"), reflect.TypeFor[string](), render, prepend)
	if _, ok := contextList.(Components); ok {
		t.Fatal("a context list must not be resolved at build time")
	}
	ctx := context.WithValue(context.Background(), tenantOptionsKey{}, []NamedOption{{Value: "acme-1"}})
	var buf strings.Builder
	if err := contextList.Render(ctx, NewCheckedWriter(&buf)); err != nil {
		t.Fatalf("render: %v", err)
	}
	if got := buf.String(); got != "first;acme-1;" {
		t.Errorf("context list rendered %q", got)
	}
	if err := contextList.Render(context.Background(), NewCheckedWriter(&buf)); err == nil {
		t.Error("a failing context list must fail the render")
	}
}
