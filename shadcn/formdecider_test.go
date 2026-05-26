package shadcn

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ungerik/go-mx"
)

type colorEnum string

func (colorEnum) EnumStrings() []string { return []string{"red", "green", "blue"} }

type nullableInt struct {
	v    int
	null bool
}

func (n nullableInt) IsNull() bool { return n.null }
func (n *nullableInt) SetNull()    { n.v = 0; n.null = true }

type formSample struct {
	Name     string
	Active   bool
	Notes    string `form:"widget=textarea"`
	When     time.Time
	Color    colorEnum
	Features map[colorEnum]struct{}
	Optional nullableInt
	Required string `form:"required"`
}

func renderShadcn(t *testing.T, target any, fieldName string, errs []error) string {
	t.Helper()
	v := reflect.ValueOf(target).Elem()
	f, ok := v.Type().FieldByName(fieldName)
	if !ok {
		t.Fatalf("no field %q", fieldName)
	}
	beh := FieldDecider(mx.FieldPath(fieldName), f, v.FieldByName(fieldName))
	if beh.Render == nil {
		t.Fatalf("nil render")
	}
	comp := beh.Render(mx.FieldPath(fieldName), f, v.FieldByName(fieldName), errs)
	if comp == nil {
		return ""
	}
	var b strings.Builder
	if err := comp.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	return b.String()
}

func TestShadcn_StringInputUsesShadcnInput(t *testing.T) {
	out := renderShadcn(t, &formSample{Name: "Alice"}, "Name", nil)
	if !strings.Contains(out, `data-slot="input"`) {
		t.Errorf("expected shadcn input slot: %q", out)
	}
	if !strings.Contains(out, "border-input") {
		t.Errorf("expected shadcn input classes: %q", out)
	}
	if !strings.Contains(out, `value="Alice"`) {
		t.Errorf("expected current value: %q", out)
	}
}

func TestShadcn_BoolUsesSwitch(t *testing.T) {
	out := renderShadcn(t, &formSample{Active: true}, "Active", nil)
	if !strings.Contains(out, `data-slot="switch"`) {
		t.Errorf("expected switch slot: %q", out)
	}
	if !strings.Contains(out, `role="switch"`) {
		t.Errorf("expected role=switch: %q", out)
	}
}

func TestShadcn_TextareaUsesShadcnTextarea(t *testing.T) {
	out := renderShadcn(t, &formSample{Notes: "long"}, "Notes", nil)
	if !strings.Contains(out, `data-slot="textarea"`) {
		t.Errorf("expected shadcn textarea: %q", out)
	}
}

func TestShadcn_EnumUsesShadcnSelect(t *testing.T) {
	out := renderShadcn(t, &formSample{Color: "green"}, "Color", nil)
	if !strings.Contains(out, `data-slot="select"`) {
		t.Errorf("expected shadcn select: %q", out)
	}
	if !strings.Contains(out, "green") {
		t.Errorf("expected enum value rendered: %q", out)
	}
}

func TestShadcn_EnumSetUsesCheckboxGrid(t *testing.T) {
	out := renderShadcn(t, &formSample{
		Features: map[colorEnum]struct{}{"red": {}, "blue": {}},
	}, "Features", nil)
	if !strings.Contains(out, `data-slot="checkbox"`) {
		t.Errorf("expected shadcn checkbox: %q", out)
	}
	if strings.Count(out, `data-slot="checkbox"`) < 3 {
		t.Errorf("expected one checkbox per enum value (≥3): %q", out)
	}
	if !strings.Contains(out, "grid-cols-2") {
		t.Errorf("expected grid layout: %q", out)
	}
}

func TestShadcn_DateTime(t *testing.T) {
	out := renderShadcn(t, &formSample{When: time.Date(2026, 5, 25, 10, 30, 0, 0, time.UTC)}, "When", nil)
	if !strings.Contains(out, `type="datetime-local"`) {
		t.Errorf("expected datetime-local: %q", out)
	}
	if !strings.Contains(out, "2026-05-25T10:30:00") {
		t.Errorf("expected date value: %q", out)
	}
}

func TestShadcn_AriaInvalidOnErrors(t *testing.T) {
	out := renderShadcn(t, &formSample{Name: "Alice"}, "Name",
		[]error{stringErr("bad value")})
	if !strings.Contains(out, `aria-invalid="true"`) {
		t.Errorf("expected aria-invalid on error: %q", out)
	}
	if !strings.Contains(out, "bad value") {
		t.Errorf("error not rendered: %q", out)
	}
}

func TestShadcn_NullableEmitsClearControl(t *testing.T) {
	out := renderShadcn(t, &formSample{Optional: nullableInt{v: 7}}, "Optional", nil)
	if !strings.Contains(out, mx.ClearSentinelName("Optional")) {
		t.Errorf("expected clear sentinel: %q", out)
	}
}

func TestShadcn_RequiredHasNoClearSentinel(t *testing.T) {
	out := renderShadcn(t, &formSample{Required: "x"}, "Required", nil)
	if strings.Contains(out, mx.ClearSentinelName("Required")) {
		t.Errorf("required field should not be clearable: %q", out)
	}
	if !strings.Contains(out, `required`) {
		t.Errorf("expected required attribute: %q", out)
	}
}

func TestShadcn_FallthroughForHidden(t *testing.T) {
	type withHidden struct {
		HID string `form:"hidden"`
	}
	out := renderShadcn(t, &withHidden{HID: "abc"}, "HID", nil)
	// hidden is delegated to html — should still render
	if !strings.Contains(out, `type="hidden"`) || !strings.Contains(out, "abc") {
		t.Errorf("hidden fallthrough broken: %q", out)
	}
}

func TestShadcn_AddsHxTrigger(t *testing.T) {
	out := renderShadcn(t, &formSample{Name: "x"}, "Name", nil)
	if !strings.Contains(out, `hx-trigger="change"`) {
		t.Errorf("expected hx-trigger=change on input: %q", out)
	}
}

func TestSectionCard_WrapsChildren(t *testing.T) {
	c := SectionCard("Accounting", mx.Text("inside"))
	var b strings.Builder
	if err := c.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if !strings.Contains(out, `data-slot="card"`) {
		t.Errorf("expected shadcn card: %q", out)
	}
	if !strings.Contains(out, "Accounting") {
		t.Errorf("missing title: %q", out)
	}
	if !strings.Contains(out, "inside") {
		t.Errorf("missing child: %q", out)
	}
}

type stringErr string

func (s stringErr) Error() string { return string(s) }
