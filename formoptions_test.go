package mx

import "testing"

func TestSelectPreface(t *testing.T) {
	opts := []NamedOption{{Name: "A", Value: "a"}}
	withEmpty := []NamedOption{{Value: ""}, {Name: "A", Value: "a"}}

	tests := []struct {
		name     string
		selected string
		tag      FormTag
		context  bool
		options  []NamedOption
		want     OptionPreface
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
		{name: "empty list with value", selected: "z", context: true, want: OptionPrefaceUnavailable},
		{name: "empty list required", selected: "", tag: FormTag{Required: true}, context: true, want: OptionPrefaceChoose},
		{name: "empty list optional blank", selected: "", context: true, want: OptionPrefaceEmpty},
		{name: "static empty required", selected: "", tag: FormTag{Required: true}, want: OptionPrefaceChoose},
		{name: "sensitive required empty", selected: "", tag: FormTag{Required: true, Sensitive: true}, context: true, options: opts, want: OptionPrefaceChoose},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectPreface(tt.selected, tt.tag, tt.context, tt.options)
			if got != tt.want {
				t.Fatalf("preface = %v, want %v", got, tt.want)
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
