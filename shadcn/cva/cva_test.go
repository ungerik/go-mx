// SPDX-License-Identifier: Apache-2.0
//
// Tests for the Go port of class-variance-authority (cva) v0.7.1 and of
// cva v1.0.0-beta.12's composes and defineConfig, Copyright 2022 Joe Bell,
// licensed under the Apache License, Version 2.0.
// See the LICENSE file in this directory and ../../THIRD-PARTY-LICENSES.md.

package cva

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/ungerik/go-mx/shadcn/twmerge"
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
			// cva would match the compound against the "" (a divergence, see
			// New): an empty prop counts as omitted for compounds too.
			name: "empty-string prop falls back to default for compound match",
			config: Config{
				Variants:         map[string]map[string]string{"shadow": {"sm": "shadow-sm", "md": "shadow-md"}},
				CompoundVariants: []Compound{{When: map[string][]string{"shadow": {"sm"}}, Class: "subtle"}},
				DefaultVariants:  map[string]string{"shadow": "sm"},
			},
			props: map[string]string{"shadow": ""},
			want:  "shadow-sm subtle",
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

// TestComposesEmptyPropResolvesOnce checks that an empty prop means the same
// in a composed config as in the composer: the default, including the default
// a composer retunes, for variants and compound variants alike.
func TestComposesEmptyPropResolvesOnce(t *testing.T) {
	shadow := Config{
		Variants: map[string]map[string]string{"shadow": {"sm": "shadow-sm", "md": "shadow-md"}},
		CompoundVariants: []Compound{
			{When: map[string][]string{"shadow": {"sm"}}, Class: "flat"},
			{When: map[string][]string{"shadow": {"md"}}, Class: "raised"},
		},
		DefaultVariants: map[string]string{"shadow": "sm"},
	}
	empty := map[string]string{"shadow": ""}
	if got, want := New(Config{Composes: []Config{shadow}})(empty), New(shadow)(empty); got != want {
		t.Errorf("composed = %q, standalone = %q", got, want)
	}
	card := New(Config{
		Composes:         []Config{shadow},
		CompoundVariants: []Compound{{When: map[string][]string{"shadow": {"md"}}, Class: "card-raised"}},
		DefaultVariants:  map[string]string{"shadow": "md"},
	})
	if got, want := card(empty), "shadow-md raised card-raised"; got != want {
		t.Errorf("retuned default = %q, want %q", got, want)
	}
}

// recorder is a concatenator that records the arguments of every call, the
// recording cx of cva v1.0.0-beta.12's "cva/config" tests. It returns the
// space join, so a resolver compiled with it returns what New's would.
type recorder struct {
	calls [][]string
}

func (r *recorder) cx(classes ...string) string {
	r.calls = append(r.calls, slices.Clone(classes))
	return strings.Join(classes, " ")
}

// TestDefineConfigForwardsToConcatenator follows "forwards inputs to the
// concatenator": the factory's resolvers and its CX both end in the
// concatenator, composed configs included.
func TestDefineConfigForwardsToConcatenator(t *testing.T) {
	join := func(classes ...string) string { return strings.Join(classes, "|") }
	f := DefineConfig(Options{CX: join})

	if got, want := f.CX("a", "b"), "a|b"; got != want {
		t.Errorf("CX = %q, want %q", got, want)
	}
	baseOnly := Config{Base: "base"}
	if got, want := f.New(baseOnly)(map[string]string{"class": "extra"}), "base|extra"; got != want {
		t.Errorf("base with class = %q, want %q", got, want)
	}
	if got, want := f.New(Config{Composes: []Config{baseOnly}, Base: "parent"})(nil), "base|parent"; got != want {
		t.Errorf("composed = %q, want %q", got, want)
	}
	button := f.New(Config{Base: "btn", Variants: map[string]map[string]string{"size": {"sm": "btn-sm"}}})
	if got, want := button(map[string]string{"size": "sm"}), "btn|btn-sm"; got != want {
		t.Errorf("variant = %q, want %q", got, want)
	}
}

// TestDefineConfigArgumentsVerbatim follows "receives assembled values
// verbatim, one argument each": a composed config resolves first, through
// the same concatenator, and its result is one argument of the composer's
// call, followed by base, variant, compound and "class".
func TestDefineConfigArgumentsVerbatim(t *testing.T) {
	var rec recorder
	f := DefineConfig(Options{CX: func(classes ...string) string {
		rec.cx(classes...)
		return "recorded"
	}})
	badge := f.New(Config{
		Composes:         []Config{{Base: "child"}},
		Base:             "badge badge--raised",
		Variants:         map[string]map[string]string{"tone": {"info": "bg-blue-500", "warn": "bg-yellow-500"}},
		CompoundVariants: []Compound{{When: map[string][]string{"tone": {"info"}}, Class: "compound-info"}},
		DefaultVariants:  map[string]string{"tone": "info"},
	})

	if got := badge(map[string]string{"class": "extra"}); got != "recorded" {
		t.Errorf("result = %q, want the concatenator's", got)
	}
	want := [][]string{
		{"child"},
		{"recorded", "badge badge--raised", "bg-blue-500", "compound-info", "extra"},
	}
	if !slices.EqualFunc(rec.calls, want, slices.Equal) {
		t.Errorf("calls = %q, want %q", rec.calls, want)
	}
}

// TestDefineConfigArgumentStreams follows the "base and class argument
// streams" cases of cva v1.0.0-beta.12's runtime semantics. Go has no
// undefined, null or className, so an empty string stands for all of them and
// is never passed on.
func TestDefineConfigArgumentStreams(t *testing.T) {
	button := Config{
		Base:             "button",
		Variants:         map[string]map[string]string{"intent": {"primary": "intent-primary"}},
		CompoundVariants: []Compound{{When: map[string][]string{"intent": {"primary"}}, Class: "compound-primary"}},
		DefaultVariants:  map[string]string{"intent": "primary"},
	}
	cases := []struct {
		name   string
		config Config
		calls  []map[string]string
		want   [][]string
	}{
		{"empty config passes no arguments", Config{}, []map[string]string{nil}, [][]string{{}}},
		{"base alone", Config{Base: "b"}, []map[string]string{nil}, [][]string{{"b"}}},
		{"base and class", Config{Base: "b"}, []map[string]string{{"class": "c"}}, [][]string{{"b", "c"}}},
		{"class alone", Config{}, []map[string]string{{"class": "c"}}, [][]string{{"c"}}},
		{"empty class is left out", Config{Base: "b"}, []map[string]string{{"class": ""}}, [][]string{{"b"}}},
		{"empty variant table", Config{Base: "b", Variants: map[string]map[string]string{}}, []map[string]string{{"class": "c"}}, [][]string{{"b", "c"}}},
		{"empty compound table", Config{Base: "b", CompoundVariants: []Compound{}}, []map[string]string{{"class": "c"}}, [][]string{{"b", "c"}}},
		{
			"nothing to emit passes no arguments",
			Config{Variants: map[string]map[string]string{"intent": {"primary": "intent-primary"}}},
			[]map[string]string{nil, {"class": "n"}},
			[][]string{{}, {"n"}},
		},
		{
			"compound class, then the caller's class", button,
			[]map[string]string{{"class": "caller-class"}},
			[][]string{{"button", "intent-primary", "compound-primary", "caller-class"}},
		},
		{
			"repeated calls stream the same", button,
			[]map[string]string{nil, {"class": "c"}, nil},
			[][]string{
				{"button", "intent-primary", "compound-primary"},
				{"button", "intent-primary", "compound-primary", "c"},
				{"button", "intent-primary", "compound-primary"},
			},
		},
		{
			"a prop equal to its default streams like omitting it", button,
			[]map[string]string{nil, {"intent": "primary"}},
			[][]string{
				{"button", "intent-primary", "compound-primary"},
				{"button", "intent-primary", "compound-primary"},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var rec recorder
			resolve := DefineConfig(Options{CX: rec.cx}).New(c.config)
			for _, props := range c.calls {
				resolve(props)
			}
			if !slices.EqualFunc(rec.calls, c.want, slices.Equal) {
				t.Errorf("calls = %q, want %q", rec.calls, c.want)
			}
		})
	}
}

// TestDefineConfigNeverPassesEmpty follows "a string-only concatenator never
// receives undefined": neither a resolver nor CX passes an empty string, so a
// concatenator need not filter.
func TestDefineConfigNeverPassesEmpty(t *testing.T) {
	f := DefineConfig(Options{CX: func(classes ...string) string {
		for _, cls := range classes {
			if cls == "" {
				t.Errorf("concatenator received an empty string in %q", classes)
			}
		}
		return strings.ToUpper(strings.Join(classes, " "))
	}})
	box := Config{Base: "box"}
	button := f.New(Config{
		Composes: []Config{box, {Base: ""}},
		Variants: map[string]map[string]string{"intent": {"primary": "primary"}, "size": {"sm": "sm"}},
	})
	if got, want := f.New(box)(nil), "BOX"; got != want {
		t.Errorf("box = %q, want %q", got, want)
	}
	if got, want := button(map[string]string{"intent": "primary"}), "BOX PRIMARY"; got != want {
		t.Errorf("button = %q, want %q", got, want)
	}
	if got, want := f.CX("a", "", "b"), "A B"; got != want {
		t.Errorf("CX = %q, want %q", got, want)
	}
}

// TestDefineConfigTailwindMerge follows the "tailwind-merge" cases: a
// concatenator that merges resolves conflicts across base, variants and the
// class override, which is the onComplete use the option replaced.
func TestDefineConfigTailwindMerge(t *testing.T) {
	tw := DefineConfig(Options{CX: func(classes ...string) string {
		return twmerge.Merge(strings.Join(classes, " "))
	}})
	button := tw.New(Config{
		Base: "font-semibold bg-gray-200 border rounded",
		Variants: map[string]map[string]string{"intent": {
			"primary":   "bg-blue-500 text-white border-transparent",
			"secondary": "bg-white text-gray-800 border-gray-400",
		}},
		DefaultVariants: map[string]string{"intent": "primary"},
	})
	if got, want := button(nil), "font-semibold border rounded bg-blue-500 text-white border-transparent"; got != want {
		t.Errorf("default = %q, want %q", got, want)
	}
	if got, want := button(map[string]string{"class": "bg-red-500"}), "font-semibold border rounded text-white border-transparent bg-red-500"; got != want {
		t.Errorf("class override = %q, want %q", got, want)
	}
	if got, want := tw.CX("bg-gray-200", "bg-blue-500"), "bg-blue-500"; got != want {
		t.Errorf("CX = %q, want %q", got, want)
	}

	// A composed config merges too, and its result is merged again with the
	// composer's classes.
	box := Config{Base: "p-4 bg-gray-100", Variants: map[string]map[string]string{"pad": {"none": "p-0"}}}
	card := tw.New(Config{
		Composes:        []Config{box},
		Base:            "font-semibold",
		Variants:        map[string]map[string]string{"intent": {"primary": "bg-blue-500", "secondary": "bg-white"}},
		DefaultVariants: map[string]string{"intent": "primary"},
	})
	if got, want := card(map[string]string{"pad": "none", "class": "text-white"}), "p-0 font-semibold bg-blue-500 text-white"; got != want {
		t.Errorf("composed = %q, want %q", got, want)
	}
}

// TestDefineConfigDefaultConcatenator checks that the zero Factory and a nil
// CX resolve exactly like New.
func TestDefineConfigDefaultConcatenator(t *testing.T) {
	config := Config{
		Composes:        []Config{{Base: "inner"}},
		Base:            "outer",
		Variants:        map[string]map[string]string{"size": {"sm": "h-8", "lg": "h-10"}},
		DefaultVariants: map[string]string{"size": "sm"},
	}
	props := map[string]string{"size": "lg", "class": "w-full"}
	want := New(config)(props)
	if got := (Factory{}).New(config)(props); got != want {
		t.Errorf("zero Factory = %q, want %q", got, want)
	}
	if got := DefineConfig(Options{}).New(config)(props); got != want {
		t.Errorf("nil CX = %q, want %q", got, want)
	}
	if got, want := (Factory{}).CX("a", "", "b"), "a b"; got != want {
		t.Errorf("zero Factory CX = %q, want %q", got, want)
	}
}
