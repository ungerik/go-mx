package mx

import (
	"reflect"
	"testing"
	"time"
)

type walkInner struct {
	Color string
	Hue   int
}

type walkOuter struct {
	Name     string
	Branding walkInner `form:"section=Branding"`
	When     time.Time
	Hidden   string `form:"hidden"`
	Skipped  string `form:"-"`
}

type embeddedFields struct {
	Embedded string
}

type walkWithEmbed struct {
	embeddedFields
	Top string
}

func collectVisits(t *testing.T, target any) []FieldVisit {
	t.Helper()
	var visits []FieldVisit
	for v := range ReflectFormFields(target) {
		visits = append(visits, v)
	}
	return visits
}

func TestReflectFormFields_BasicWalk(t *testing.T) {
	visits := collectVisits(t, &walkOuter{})
	paths := pathSet(visits)
	want := []FieldPath{
		"Name",
		"Branding-Color",
		"Branding-Hue",
		"When",
		"Hidden",
		// "Skipped" omitted via form:"-"
	}
	for _, p := range want {
		if _, ok := paths[p]; !ok {
			t.Errorf("missing visit for path %q", p)
		}
	}
	if _, ok := paths["Skipped"]; ok {
		t.Errorf("Skipped should be omitted")
	}
}

func TestReflectFormFields_SectionAttached(t *testing.T) {
	for v := range ReflectFormFields(&walkOuter{}) {
		switch v.Path {
		case "Branding-Color", "Branding-Hue":
			if v.Section != "Branding" {
				t.Errorf("%s: section=%q, want Branding", v.Path, v.Section)
			}
		case "Name", "When", "Hidden":
			if v.Section != "" {
				t.Errorf("%s: section=%q, want empty", v.Path, v.Section)
			}
		}
	}
}

func TestReflectFormFields_AnonymousEmbedInlined(t *testing.T) {
	visits := collectVisits(t, &walkWithEmbed{})
	paths := pathSet(visits)
	if _, ok := paths["Embedded"]; !ok {
		t.Errorf("embedded field should appear at top level")
	}
	if _, ok := paths["Top"]; !ok {
		t.Errorf("top-level field missing")
	}
	if _, ok := paths["embeddedFields"]; ok {
		t.Errorf("anonymous embed itself should not be yielded")
	}
}

func TestReflectFormFields_TimeAsLeaf(t *testing.T) {
	visits := collectVisits(t, &walkOuter{})
	found := false
	for _, v := range visits {
		if v.Path == "When" {
			found = true
			if v.Kind != FieldKindDateTime {
				t.Errorf("When: kind=%s, want %s", v.Kind, FieldKindDateTime)
			}
		}
		if v.Path == "When-Year" || v.Path == "When-Month" {
			t.Errorf("time.Time should not be recursed into; got %q", v.Path)
		}
	}
	if !found {
		t.Errorf("When path missing")
	}
}

type depthOuter struct {
	Inner depthMiddle `form:"nested"`
}

type depthMiddle struct {
	Deeper depthLeaf `form:"nested"`
}

type depthLeaf struct {
	X string
}

func TestReflectFormFields_MaxDepthExceeded(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for nesting >1, got none")
		}
	}()
	for range ReflectFormFields(&depthOuter{}) {
		// never reached
	}
}

func TestReflectFormFields_NonStructPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for non-struct, got none")
		}
	}()
	for range ReflectFormFields(42) {
	}
}

func TestReflectFormFields_NilPointerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for nil pointer, got none")
		}
	}()
	var s *walkOuter
	for range ReflectFormFields(s) {
	}
}

func pathSet(visits []FieldVisit) map[FieldPath]reflect.Value {
	m := make(map[FieldPath]reflect.Value, len(visits))
	for _, v := range visits {
		m[v.Path] = v.Value
	}
	return m
}
