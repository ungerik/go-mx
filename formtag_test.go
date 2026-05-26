package mx

import "testing"

func TestParseFormTagString(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want FormTag
	}{
		{name: "empty", raw: "", want: FormTag{}},
		{name: "dash skips", raw: "-", want: FormTag{Skip: true}},
		{name: "widget skip", raw: "widget=skip", want: FormTag{Widget: "skip", Skip: true}},
		{name: "required bare", raw: "required", want: FormTag{Required: true}},
		{name: "required true", raw: "required=true", want: FormTag{Required: true}},
		{name: "required false", raw: "required=false", want: FormTag{}},
		{name: "numeric", raw: "min=4,max=6,step=0.5",
			want: FormTag{Min: "4", Max: "6", Step: "0.5"}},
		{name: "widget+placeholder", raw: "widget=textarea,placeholder=Notes",
			want: FormTag{Widget: "textarea", Placeholder: "Notes"}},
		{name: "section infers nested", raw: "section=Accounting",
			want: FormTag{Section: "Accounting", Nested: true}},
		{name: "hidden", raw: "hidden", want: FormTag{Hidden: true}},
		{name: "readonly", raw: "readonly", want: FormTag{Readonly: true}},
		{name: "sensitive", raw: "sensitive", want: FormTag{Sensitive: true}},
		{name: "quoted label with comma",
			raw:  "label='Owner, Operator',help=\"a, b\"",
			want: FormTag{Label: "Owner, Operator", Help: "a, b"}},
		{name: "options", raw: "options=ISO4217",
			want: FormTag{Options: "ISO4217"}},
		{name: "repeatable", raw: "repeatable",
			want: FormTag{Repeatable: true}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseFormTagString(c.raw)
			if got != c.want {
				t.Errorf("ParseFormTagString(%q) =\n  %+v\nwant\n  %+v", c.raw, got, c.want)
			}
		})
	}
}
