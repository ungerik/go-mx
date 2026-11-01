// SPDX-License-Identifier: Apache-2.0
//
// Tests for the Go port of cva v1.0.0-beta's compose, Copyright 2022 Joe Bell,
// licensed under the Apache License, Version 2.0. See the LICENSE file in this
// directory and ../../THIRD-PARTY-LICENSES.md.

package cva

import (
	"maps"
	"testing"
)

// TestCompose follows the "should merge into a single component" case of
// cva's own compose test, with the gap values as strings because Go props are
// a string map.
func TestCompose(t *testing.T) {
	box := New(Config{
		Variants: map[string]map[string]string{
			"shadow": {"sm": "shadow-sm", "md": "shadow-md"},
		},
		DefaultVariants: map[string]string{"shadow": "sm"},
	})
	stack := New(Config{
		Variants: map[string]map[string]string{
			"gap": {"unset": "", "1": "gap-1", "2": "gap-2", "3": "gap-3"},
		},
		DefaultVariants: map[string]string{"gap": "unset"},
	})
	card := Compose(box, stack)

	cases := []struct {
		name  string
		props map[string]string
		want  string
	}{
		{name: "nil props", props: nil, want: "shadow-sm"},
		{name: "defaults only", props: map[string]string{}, want: "shadow-sm"},
		{name: "class override", props: map[string]string{"class": "adhoc-class"}, want: "shadow-sm adhoc-class"},
		{name: "first part's variant", props: map[string]string{"shadow": "md"}, want: "shadow-md"},
		{name: "second part's variant", props: map[string]string{"gap": "2"}, want: "shadow-sm gap-2"},
		{
			name:  "both parts and class",
			props: map[string]string{"shadow": "md", "gap": "3", "class": "adhoc-class"},
			want:  "shadow-md gap-3 adhoc-class",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := card(c.props); got != c.want {
				t.Errorf("Compose(box, stack)(%v) = %q, want %q", c.props, got, c.want)
			}
		})
	}
}

// TestComposeClassOnce checks that the "class" override is withheld from the
// parts: each part appends its props "class" itself, so passing it through
// would repeat the override once per part and bury it before later parts.
func TestComposeClassOnce(t *testing.T) {
	a := New(Config{Base: "a"})
	b := New(Config{Base: "b"})
	if got, want := Compose(a, b)(map[string]string{"class": "x"}), "a b x"; got != want {
		t.Errorf("Compose(a, b) = %q, want %q", got, want)
	}
}

// TestComposeLeavesPropsUnchanged guards the copy Compose makes to withhold
// "class": deleting it from the caller's map would change the caller's data.
func TestComposeLeavesPropsUnchanged(t *testing.T) {
	props := map[string]string{"class": "x", "shadow": "md"}
	before := maps.Clone(props)
	Compose(New(Config{Base: "a"}))(props)
	if !maps.Equal(props, before) {
		t.Errorf("Compose modified props: got %v, want %v", props, before)
	}
}

// TestComposeEdgeCases covers the shapes with nothing to join.
func TestComposeEdgeCases(t *testing.T) {
	if got := Compose()(map[string]string{"class": "x"}); got != "x" {
		t.Errorf("Compose() with class = %q, want %q", got, "x")
	}
	if got := Compose()(nil); got != "" {
		t.Errorf("Compose()(nil) = %q, want empty", got)
	}
	// A part that resolves to nothing adds no stray separator.
	empty := New(Config{})
	if got := Compose(empty, New(Config{Base: "b"}), empty)(nil); got != "b" {
		t.Errorf("Compose(empty, b, empty) = %q, want %q", got, "b")
	}
}

// TestComposeNested checks that a composed resolver is itself a part, so a
// family can be built up in layers.
func TestComposeNested(t *testing.T) {
	inner := Compose(New(Config{Base: "a"}), New(Config{Base: "b"}))
	outer := Compose(inner, New(Config{Base: "c"}))
	if got, want := outer(map[string]string{"class": "x"}), "a b c x"; got != want {
		t.Errorf("nested Compose = %q, want %q", got, want)
	}
}
