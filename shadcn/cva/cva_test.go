// SPDX-License-Identifier: Apache-2.0
//
// Tests for the Go port of class-variance-authority (cva) v0.7.1 and of
// cva v1.0.0-beta.12's composes, Copyright 2022 Joe Bell, licensed under the
// Apache License, Version 2.0.
// See the LICENSE file in this directory and ../../THIRD-PARTY-LICENSES.md.

package cva

import (
	"maps"
	"testing"
)

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

// TestComposes follows the "cva — composes" cases of cva v1.0.0-beta.12's own
// test suite. Numeric variant values become strings because Go props are a
// string map, and a null class becomes "".
func TestComposes(t *testing.T) {
	box := Config{
		Variants:        map[string]map[string]string{"shadow": {"sm": "shadow-sm", "md": "shadow-md"}},
		DefaultVariants: map[string]string{"shadow": "sm"},
	}
	stack := Config{
		Variants:        map[string]map[string]string{"gap": {"unset": "", "1": "gap-1", "2": "gap-2", "3": "gap-3"}},
		DefaultVariants: map[string]string{"gap": "unset"},
	}
	pad := func(def string) Config {
		return Config{
			Variants:        map[string]map[string]string{"pad": {"sm": "pad-sm", "lg": "pad-lg"}},
			DefaultVariants: map[string]string{"pad": def},
		}
	}
	tone := Config{
		Variants:        map[string]map[string]string{"_tone": {"quiet": "tone-quiet", "loud": "tone-loud"}},
		DefaultVariants: map[string]string{"_tone": "quiet"},
	}

	cases := []struct {
		name   string
		config Config
		props  map[string]string
		want   string
	}{
		{"single component", Config{Composes: []Config{box}}, nil, "shadow-sm"},
		{"single component with class", Config{Composes: []Config{box}}, map[string]string{"class": "adhoc-class"}, "shadow-sm adhoc-class"},
		{"single component with variant", Config{Composes: []Config{box}}, map[string]string{"shadow": "md"}, "shadow-md"},
		{"merged", Config{Composes: []Config{box, stack}}, nil, "shadow-sm"},
		{"merged second variant", Config{Composes: []Config{box, stack}}, map[string]string{"gap": "2"}, "shadow-sm gap-2"},
		{
			"merged both variants and class", Config{Composes: []Config{box, stack}},
			map[string]string{"shadow": "md", "gap": "3", "class": "adhoc-class"}, "shadow-md gap-3 adhoc-class",
		},
		{
			"additional variants", Config{
				Composes:        []Config{box, stack},
				Variants:        map[string]map[string]string{"rounded": {"sm": "rounded-sm", "lg": "rounded-lg"}},
				DefaultVariants: map[string]string{"rounded": "sm"},
			},
			map[string]string{"shadow": "md", "gap": "2", "rounded": "lg"}, "shadow-md gap-2 rounded-lg",
		},
		{
			"locally added value of a composed variant", Config{
				Composes: []Config{box},
				Variants: map[string]map[string]string{"shadow": {"lg": "shadow-lg"}},
			},
			map[string]string{"shadow": "lg"}, "shadow-lg",
		},
		{
			// The composed default also selects the local class of the
			// redeclared variant, so no props equals selecting the default.
			"locally redeclared variant at the composed default", Config{
				Composes: []Config{box},
				Variants: map[string]map[string]string{"shadow": {"sm": "local-sm"}},
			},
			nil, "shadow-sm local-sm",
		},
		{
			"nested", Config{Base: "outer", Composes: []Config{{Base: "inner", Composes: []Config{
				{Base: "box", Variants: box.Variants, DefaultVariants: box.DefaultVariants},
			}}}},
			map[string]string{"shadow": "md"}, "box shadow-md inner outer",
		},
		{
			"composed config without variants", Config{Composes: []Config{
				{Base: "plain"},
				{Base: "styled", Variants: map[string]map[string]string{"shadow": {"sm": "shadow-sm"}}},
			}},
			map[string]string{"shadow": "sm"}, "plain styled shadow-sm",
		},
		{
			"composed-only variant", Config{Composes: []Config{tone}, Variants: pad("sm").Variants, DefaultVariants: pad("sm").DefaultVariants},
			map[string]string{"_tone": "loud"}, "tone-loud pad-sm",
		},
		{
			// The composer's default reaches the composed config too.
			"composer retunes a composed default", Config{
				Composes:        []Config{tone},
				Variants:        map[string]map[string]string{"_tone": {"loud": "loud-local"}},
				DefaultVariants: map[string]string{"_tone": "loud"},
			},
			nil, "tone-loud loud-local",
		},
		{"last composed default wins", Config{Composes: []Config{pad("sm"), pad("lg")}}, nil, "pad-lg pad-lg"},
		{
			"compound variant tests a composed variant", Config{
				Composes:         []Config{box},
				CompoundVariants: []Compound{{When: map[string][]string{"shadow": {"sm"}}, Class: "subtle"}},
			},
			nil, "shadow-sm subtle",
		},
		{"no composed configs", Config{Composes: []Config{}}, map[string]string{"class": "x"}, "x"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := New(c.config)(c.props); got != c.want {
				t.Errorf("New(...)(%v) = %q, want %q", c.props, got, c.want)
			}
		})
	}
}

// TestComposesClassOnce checks that the "class" override is withheld from the
// composed configs: each appends its props "class" itself, so passing it
// through would repeat the override once per config and bury it mid-string.
func TestComposesClassOnce(t *testing.T) {
	card := New(Config{Base: "c", Composes: []Config{{Base: "a"}, {Base: "b"}}})
	if got, want := card(map[string]string{"class": "x"}), "a b c x"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestComposesLeavesInputsUnchanged guards the copies made for the composed
// configs: neither the caller's props nor a config's defaults may change.
func TestComposesLeavesInputsUnchanged(t *testing.T) {
	props := map[string]string{"class": "x", "shadow": "md"}
	defaults := map[string]string{"shadow": "sm"}
	propsBefore, defaultsBefore := maps.Clone(props), maps.Clone(defaults)
	card := New(Config{Composes: []Config{{
		Variants:        map[string]map[string]string{"shadow": {"sm": "shadow-sm", "md": "shadow-md"}},
		DefaultVariants: defaults,
	}}})
	card(props)
	if !maps.Equal(props, propsBefore) {
		t.Errorf("props changed: got %v, want %v", props, propsBefore)
	}
	if !maps.Equal(defaults, defaultsBefore) {
		t.Errorf("composed defaults changed: got %v, want %v", defaults, defaultsBefore)
	}
	// The composer's own merged defaults are shared across calls, so a prop
	// of one call must not leak into the defaults of the next.
	if got, want := card(nil), "shadow-sm"; got != want {
		t.Errorf("call after an override = %q, want %q", got, want)
	}
}
