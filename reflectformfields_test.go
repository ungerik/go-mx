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

type walkMeta struct {
	Note string
}

type walkNilSection struct {
	Name string
	Meta *walkMeta `form:"nested"`
	Qty  int
}

// A nil pointer section is skipped, but the fields after it are walked.
func TestReflectFormFields_NilSectionKeepsSiblings(t *testing.T) {
	paths := pathSet(collectVisits(t, &walkNilSection{}))
	if _, ok := paths["Qty"]; !ok {
		t.Errorf("the field after a nil section is not walked: %v", paths)
	}
	if _, ok := paths["Meta-Note"]; ok {
		t.Errorf("a nil section's field is walked: %v", paths)
	}
}

type walkReadonlySection struct {
	Name string
	Meta walkMeta `form:"nested,readonly"`
}

// form:"readonly" on a section makes its fields readonly, for the walk and
// for a decider that detects the field again.
func TestReflectFormFields_ReadonlySection(t *testing.T) {
	for v := range ReflectFormFields(&walkReadonlySection{}) {
		_, tag := DetectField(v.Path, v.Field, v.Value)
		want := v.Path == "Meta-Note"
		if v.Tag.Readonly != want || tag.Readonly != want {
			t.Errorf("%s: visit readonly %t, detected readonly %t, want %t", v.Path, v.Tag.Readonly, tag.Readonly, want)
		}
	}
}

type walkLocked struct {
	Note   string `form:"widget=textarea,label=Notes"`
	Owner  string `form:"help=The owner's login email"`
	Hash   string `form:"-"`
	Open   string `form:"readonly=false"`
	Quoted string `form:"help='unterminated"`
}

type walkLockedSection struct {
	Name    string
	Account walkLocked `form:"section=Account,readonly"`
}

// Readonly passed down by a section keeps each field's own options, holds
// against a quote in them, and never brings a skipped field back.
func TestReflectFormFields_ReadonlySectionKeepsTags(t *testing.T) {
	got := map[FieldPath]FormTag{}
	for v := range ReflectFormFields(&walkLockedSection{}) {
		_, tag := DetectField(v.Path, v.Field, v.Value)
		got[v.Path] = tag
	}
	if _, ok := got["Account-Hash"]; ok {
		t.Errorf("a form:\"-\" field in a readonly section is walked: %v", got)
	}
	if tag := got["Account-Note"]; !tag.Readonly || tag.Widget != "textarea" || tag.Label != "Notes" {
		t.Errorf("Account-Note detected as %+v, want readonly with its own options", tag)
	}
	if tag := got["Account-Owner"]; !tag.Readonly {
		t.Errorf("an apostrophe in the help text cancels readonly: %+v", tag)
	}
	// Neither the field's own readonly=false nor an unterminated quote in
	// its options cancels the readonly the section passes down.
	for _, p := range []FieldPath{"Account-Open", "Account-Quoted"} {
		if tag := got[p]; !tag.Readonly {
			t.Errorf("%s detected as %+v, want readonly", p, tag)
		}
	}
}

type walkAudit struct {
	CreatedBy string
	IsAdmin   bool
}

type walkInternal struct {
	Role string
}

type walkEmbedTags struct {
	walkAudit    `form:"readonly"`
	walkInternal `form:"-"`
	Name         string
}

// The form tag of an embedded struct applies to the fields it promotes:
// form:"-" omits them, form:"readonly" locks them.
func TestReflectFormFields_EmbeddedStructTags(t *testing.T) {
	got := map[FieldPath]FieldVisit{}
	for v := range ReflectFormFields(&walkEmbedTags{}) {
		got[v.Path] = v
	}
	if _, ok := got["Role"]; ok {
		t.Errorf("a field of a form:\"-\" embed is walked: %v", got)
	}
	for _, p := range []FieldPath{"CreatedBy", "IsAdmin"} {
		v, ok := got[p]
		if !ok {
			t.Fatalf("%s is not walked", p)
		}
		if _, tag := DetectField(v.Path, v.Field, v.Value); !v.Tag.Readonly || !tag.Readonly {
			t.Errorf("%s of a readonly embed is not readonly", p)
		}
	}
	if got["Name"].Tag.Readonly {
		t.Errorf("Name is readonly")
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
