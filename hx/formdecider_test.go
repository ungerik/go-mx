package hx

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

type sample struct {
	Name   string
	Active bool
	Notes  string `form:"widget=textarea"`
	Hidden string `form:"hidden"`
}

func renderField(t *testing.T, fieldName string) string {
	t.Helper()
	s := &sample{Name: "Alice", Active: true, Notes: "long"}
	v := reflect.ValueOf(s).Elem()
	f, ok := v.Type().FieldByName(fieldName)
	if !ok {
		t.Fatalf("no field %q", fieldName)
	}
	beh := FieldDecider(mx.FieldPath(fieldName), f, v.FieldByName(fieldName))
	if beh.Render == nil {
		t.Fatalf("nil Render")
	}
	comp := beh.Render(mx.FieldPath(fieldName), f, v.FieldByName(fieldName), nil)
	var b strings.Builder
	if err := comp.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	return b.String()
}

func TestFieldDecider_AddsHxTriggerOnInputs(t *testing.T) {
	out := renderField(t, "Name")
	if !strings.Contains(out, `hx-trigger="change"`) {
		t.Errorf("expected hx-trigger=change on text input: %q", out)
	}
}

func TestFieldDecider_AddsHxTriggerOnTextarea(t *testing.T) {
	out := renderField(t, "Notes")
	if !strings.Contains(out, `hx-trigger="change"`) {
		t.Errorf("expected hx-trigger=change on textarea: %q", out)
	}
}

func TestFieldDecider_AddsHxTriggerOnCheckbox(t *testing.T) {
	out := renderField(t, "Active")
	if !strings.Contains(out, `hx-trigger="change"`) {
		t.Errorf("expected hx-trigger=change on checkbox: %q", out)
	}
}

func TestFieldDecider_SkipsHiddenInputs(t *testing.T) {
	out := renderField(t, "Hidden")
	if strings.Contains(out, `hx-trigger`) {
		t.Errorf("hidden input should not get hx-trigger: %q", out)
	}
}

func TestFieldDecider_FallsThroughForUnknown(t *testing.T) {
	// hx should delegate to html for Parse — verify by parsing nothing.
	v := reflect.ValueOf(&sample{}).Elem()
	f, _ := v.Type().FieldByName("Name")
	beh := FieldDecider("Name", f, v.FieldByName("Name"))
	if beh.Parse == nil {
		t.Error("Parse closure should be inherited from html.FieldDecider")
	}
}

// Registry entries are process-global and duplicate registration
// panics, so tests register once in init() — never in test bodies,
// which would crash under `go test -count=2`.
func init() {
	mx.RegisterNamedOptions("test-hx-partners", func(context.Context) ([]mx.NamedOption, error) {
		return []mx.NamedOption{{Name: "Partner One", Value: "p1"}}, nil
	})
	mx.RegisterNamedOptions("test-hx-tags", func(context.Context) ([]mx.NamedOption, error) {
		return []mx.NamedOption{
			{Name: "Tag A", Value: "a"},
			{Name: "Tag B", Value: "b"},
		}, nil
	})
}

// TestFieldDecider_AddsHxTriggerOnRegistryOptionsSelect guards the
// render-time option collection design: the <select> element must stay
// a build-time element (only its options are deferred), otherwise this
// layer's attribute injection cannot reach it.
func TestFieldDecider_AddsHxTriggerOnRegistryOptionsSelect(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-hx-partners"`
	}
	s := &draftForm{}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("Partner")
	beh := FieldDecider(mx.FieldPath("Partner"), f, v.Field(0))
	comp := beh.Render(mx.FieldPath("Partner"), f, v.Field(0), nil)
	var b strings.Builder
	if err := comp.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, `<select`) || !strings.Contains(out, "Partner One") {
		t.Errorf("expected select with registry option: %q", out)
	}
	if !strings.Contains(out, `hx-trigger="change"`) {
		t.Errorf("expected hx-trigger=change on registry-options select: %q", out)
	}
}

// TestFieldDecider_RegistryEnumSetKnownLimitation pins a documented
// limitation of render-time option collection: enum-set checkboxes for
// registry-backed options are born inside the deferred ComponentFunc,
// so this layer's build-time attribute injection cannot reach them —
// they render correctly but without hx-trigger. (shadcn is unaffected:
// it wires hx.Trigger itself inside the render callback.)
func TestFieldDecider_RegistryEnumSetKnownLimitation(t *testing.T) {
	type tagForm struct {
		Tags []string `form:"options=test-hx-tags"`
	}
	s := &tagForm{}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("Tags")
	beh := FieldDecider(mx.FieldPath("Tags"), f, v.Field(0))
	comp := beh.Render(mx.FieldPath("Tags"), f, v.Field(0), nil)
	var b strings.Builder
	if err := comp.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if strings.Count(out, `type="checkbox"`) != 2 {
		t.Errorf("expected registry checkboxes to render: %q", out)
	}
	if strings.Contains(out, `hx-trigger`) {
		t.Errorf("limitation lifted? deferred enum-set checkboxes now carry hx-trigger — update the docs and this test: %q", out)
	}
}
