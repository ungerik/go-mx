package cva

import "testing"

// TestNew exercises the resolver against cases derived from
// class-variance-authority v0.7.1's own test suite.
func TestNew(t *testing.T) {
	cases := []struct {
		name   string
		config Config
		props  map[string]string
		want   string
	}{
		{
			name:   "empty config",
			config: Config{},
			props:  map[string]string{},
			want:   "",
		},
		{
			name:   "base only",
			config: Config{Base: "button font-semibold"},
			props:  map[string]string{},
			want:   "button font-semibold",
		},
		{
			name: "variant without default",
			config: Config{Variants: map[string]map[string]string{
				"intent": {"primary": "bg-blue-500", "secondary": "bg-white"},
			}},
			props: map[string]string{"intent": "primary"},
			want:  "bg-blue-500",
		},
		{
			name: "variant with default, prop omitted",
			config: Config{
				Base:            "button",
				Variants:        map[string]map[string]string{"intent": {"primary": "bg-blue-500"}},
				DefaultVariants: map[string]string{"intent": "primary"},
			},
			props: map[string]string{},
			want:  "button bg-blue-500",
		},
		{
			name: "boolean variant",
			config: Config{Variants: map[string]map[string]string{
				"disabled": {"true": "opacity-50", "false": "cursor-pointer"},
			}},
			props: map[string]string{"disabled": "true"},
			want:  "opacity-50",
		},
		{
			name: "unset variant value contributes nothing",
			config: Config{Variants: map[string]map[string]string{
				"size": {"unset": "", "small": "text-sm"},
			}},
			props: map[string]string{"size": "unset"},
			want:  "",
		},
		{
			name: "missing variant, no default, contributes nothing",
			config: Config{
				Base:     "btn",
				Variants: map[string]map[string]string{"intent": {"primary": "bg-blue"}},
			},
			props: map[string]string{},
			want:  "btn",
		},
		{
			name: "empty-string prop falls back to default",
			config: Config{
				Variants:        map[string]map[string]string{"intent": {"primary": "bg-blue"}},
				DefaultVariants: map[string]string{"intent": "primary"},
			},
			props: map[string]string{"intent": ""},
			want:  "bg-blue",
		},
		{
			name: "compound variant, scalar condition",
			config: Config{
				Variants: map[string]map[string]string{
					"intent": {"primary": "blue", "secondary": "white"},
					"size":   {"small": "sm", "medium": "md"},
				},
				CompoundVariants: []Compound{{
					When:  map[string][]string{"intent": {"primary"}, "size": {"medium"}},
					Class: "primary-medium",
				}},
			},
			props: map[string]string{"intent": "primary", "size": "medium"},
			want:  "blue md primary-medium",
		},
		{
			name: "compound variant, array condition",
			config: Config{
				Variants: map[string]map[string]string{
					"intent": {"primary": "blue", "secondary": "white", "danger": "red"},
				},
				CompoundVariants: []Compound{{
					When:  map[string][]string{"intent": {"secondary", "danger"}},
					Class: "border-red",
				}},
			},
			props: map[string]string{"intent": "danger"},
			want:  "red border-red",
		},
		{
			name: "compound variant does not match",
			config: Config{
				Variants: map[string]map[string]string{
					"intent": {"primary": "blue", "secondary": "white"},
				},
				CompoundVariants: []Compound{{
					When:  map[string][]string{"intent": {"primary"}},
					Class: "only-primary",
				}},
			},
			props: map[string]string{"intent": "secondary"},
			want:  "white",
		},
		{
			name: "class override appended last",
			config: Config{
				Base:            "button",
				Variants:        map[string]map[string]string{"intent": {"primary": "bg-blue"}},
				DefaultVariants: map[string]string{"intent": "primary"},
			},
			props: map[string]string{"class": "custom-class"},
			want:  "button bg-blue custom-class",
		},
		{
			name: "defaults feed compound match",
			config: Config{
				Variants: map[string]map[string]string{
					"intent": {"primary": "blue", "secondary": "white"},
					"size":   {"medium": "md"},
				},
				CompoundVariants: []Compound{{
					When:  map[string][]string{"intent": {"primary"}, "size": {"medium"}},
					Class: "pm",
				}},
				DefaultVariants: map[string]string{"intent": "primary", "size": "medium"},
			},
			props: map[string]string{},
			want:  "blue md pm",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := New(c.config)(c.props); got != c.want {
				t.Errorf("New(...)(%v) = %q, want %q", c.props, got, c.want)
			}
		})
	}
}
