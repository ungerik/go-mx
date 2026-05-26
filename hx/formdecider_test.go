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
