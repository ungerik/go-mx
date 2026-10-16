package mx

import (
	"strings"
	"testing"
)

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
		wantErr  bool
	}{
		{name: "optional empty", selected: "", context: true, options: opts, want: OptionPrefaceEmpty},
		{name: "required empty", selected: "", tag: FormTag{Required: true}, context: true, options: opts, want: OptionPrefaceChoose},
		{name: "list already has empty", selected: "", context: true, options: withEmpty, want: OptionPrefaceNone},
		{name: "required list already has empty", selected: "", tag: FormTag{Required: true}, context: true, options: withEmpty, want: OptionPrefaceNone},
		{name: "optional listed value", selected: "a", context: true, options: opts, want: OptionPrefaceClear},
		{name: "required listed value", selected: "a", tag: FormTag{Required: true}, context: true, options: opts, want: OptionPrefaceNone},
		{name: "context out of list", selected: "z", context: true, options: opts, want: OptionPrefaceUnavailable},
		{name: "sensitive out of list", selected: "z", tag: FormTag{Sensitive: true}, context: true, options: opts, want: OptionPrefaceNone},
		{name: "static out of list", selected: "z", options: opts, want: OptionPrefaceClear},
		{name: "required static out of list", selected: "z", tag: FormTag{Required: true}, options: opts, want: OptionPrefaceNone},
		{name: "empty list with value", selected: "z", context: true, wantErr: true},
		{name: "empty list required", selected: "", tag: FormTag{Required: true}, context: true, wantErr: true},
		{name: "empty list optional blank", selected: "", context: true, want: OptionPrefaceEmpty},
		{name: "static empty required", selected: "", tag: FormTag{Required: true}, want: OptionPrefaceChoose},
		{name: "sensitive required empty", selected: "", tag: FormTag{Required: true, Sensitive: true}, context: true, options: opts, want: OptionPrefaceChoose},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectPreface(tt.selected, tt.tag, tt.context, tt.options)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "option list is empty") {
					t.Fatalf("got (%v, %v), want empty-list error", got, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("preface: %v", err)
			}
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
		wantErr bool
	}{
		{name: "all listed", members: []string{"a"}, context: true, options: opts, want: OptionPrefaceNone},
		{name: "missing member", members: []string{"a", "hidden"}, context: true, options: opts, want: OptionPrefaceValuesUnavailable},
		{name: "sensitive missing", members: []string{"hidden"}, tag: FormTag{Sensitive: true}, context: true, options: opts, want: OptionPrefaceNone},
		{name: "static missing", members: []string{"hidden"}, options: opts, want: OptionPrefaceNone},
		{name: "empty list with member", members: []string{"hidden"}, context: true, wantErr: true},
		{name: "empty list required", tag: FormTag{Required: true}, context: true, wantErr: true},
		{name: "empty list optional blank", context: true, want: OptionPrefaceNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnumSetPreface(tt.members, tt.tag, tt.context, tt.options)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "option list is empty") {
					t.Fatalf("got (%v, %v), want empty-list error", got, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("preface: %v", err)
			}
			if got != tt.want {
				t.Fatalf("preface = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPreserveUnlistedSetMembers(t *testing.T) {
	opts := []NamedOption{{Value: "a"}, {Value: "b"}}
	tests := []struct {
		name      string
		stored    []string
		submitted []string
		options   []NamedOption
		want      []string
	}{
		{name: "keep hidden drop unchecked", stored: []string{"hidden", "a"}, submitted: []string{"b"}, options: opts, want: []string{"b", "hidden"}},
		{name: "unchecked visible leaves only hidden", stored: []string{"hidden", "a"}, submitted: nil, options: []NamedOption{{Value: "a"}}, want: []string{"hidden"}},
		{name: "listed members follow the submission", stored: []string{"a", "b"}, submitted: []string{"a"}, options: opts, want: []string{"a"}},
		{name: "duplicate hidden stored once", stored: []string{"hidden", "hidden"}, submitted: nil, options: opts, want: []string{"hidden"}},
		{name: "empty list keeps stored", stored: []string{"hidden", "also"}, submitted: nil, want: []string{"hidden", "also"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PreserveUnlistedSetMembers(tt.stored, tt.submitted, tt.options)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestEmptyOptionListError(t *testing.T) {
	err := EmptyOptionListError("Partner", "partners")
	if !strings.Contains(err.Error(), `field Partner`) || !strings.Contains(err.Error(), `"partners"`) {
		t.Fatalf("provider error: %v", err)
	}
	err = EmptyOptionListError("Partner", "")
	if !strings.Contains(err.Error(), "field Partner: option list is empty") {
		t.Fatalf("type-provider error: %v", err)
	}
}
