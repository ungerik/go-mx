package shadcn

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
	"uuid"

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

func TestShadcn_UsesFieldSystem(t *testing.T) {
	// The reflected field renders the same Field markup a hand-written form
	// would: a Field root wrapping a FieldLabel and the control.
	type helped struct {
		Name string `form:"help=Your display name"`
	}
	out := renderShadcn(t, &helped{Name: "Alice"}, "Name", nil)
	for _, want := range []string{
		`data-slot="field"`, `role="group"`,
		`data-slot="field-label"`, `for="Name"`,
		// help text is a FieldDescription, not a bare <small>
		`data-slot="field-description"`, "Your display name",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// The pre-Field stack must be gone.
	if strings.Contains(out, "grid gap-1.5") || strings.Contains(out, "<small") {
		t.Errorf("pre-Field markup still present: %s", out)
	}
}

func TestShadcn_ErrorUsesFieldError(t *testing.T) {
	out := renderShadcn(t, &formSample{Name: "Alice"}, "Name",
		[]error{stringErr("bad value")})
	for _, want := range []string{
		// Field root flags the error state...
		`data-invalid="true"`,
		// ...and the message announces to screen readers.
		`data-slot="field-error"`, `role="alert"`, "bad value",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// The data-error field-path hook the html layer uses is preserved.
	if !strings.Contains(out, `data-error="Name"`) {
		t.Errorf("data-error path hook dropped: %s", out)
	}
}

func TestShadcn_BoolUsesField(t *testing.T) {
	out := renderShadcn(t, &formSample{Active: true}, "Active", nil)
	for _, want := range []string{`data-slot="field"`, `data-slot="field-label"`, `data-slot="switch"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
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

type shadcnPartnersKey struct{}

// Registry entries are process-global and duplicate registration
// panics, so tests register once in init() — never in test bodies,
// which would crash under `go test -count=2`.
func init() {
	mx.RegisterNamedOptions("test-shadcn-partners", func(ctx context.Context) ([]mx.NamedOption, error) {
		opts, _ := ctx.Value(shadcnPartnersKey{}).([]mx.NamedOption)
		return opts, nil
	})
	mx.RegisterNamedOptions("test-shadcn-tags", func(context.Context) ([]mx.NamedOption, error) {
		return []mx.NamedOption{
			{Name: "Tag A", Value: "a"},
			{Name: "Tag B", Value: "b"},
		}, nil
	})
}

func TestShadcn_SelectWithRegistryOptions(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-shadcn-partners"`
	}
	s := &draftForm{Partner: "p1"}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("Partner")
	beh := FieldDecider(mx.FieldPath("Partner"), f, v.Field(0))
	comp := beh.Render(mx.FieldPath("Partner"), f, v.Field(0), nil)

	ctx := context.WithValue(context.Background(), shadcnPartnersKey{},
		[]mx.NamedOption{{Name: "Partner One", Value: "p1"}})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, `data-slot="select"`) {
		t.Errorf("expected shadcn select: %q", out)
	}
	if !strings.Contains(out, "Partner One") || !strings.Contains(out, "selected") {
		t.Errorf("expected context option p1 rendered and selected: %q", out)
	}
}

// TestShadcn_SelectPlaceholderForOutOfListValue mirrors the html
// layer: a current value missing from the per-request option list
// shows as a disabled placeholder instead of the browser silently
// displaying the first option.
func TestShadcn_SelectPlaceholderForOutOfListValue(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-shadcn-partners"`
	}
	s := &draftForm{Partner: "gone"}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("Partner")
	beh := FieldDecider(mx.FieldPath("Partner"), f, v.Field(0))
	comp := beh.Render(mx.FieldPath("Partner"), f, v.Field(0), nil)

	ctx := context.WithValue(context.Background(), shadcnPartnersKey{},
		[]mx.NamedOption{{Name: "Partner One", Value: "p1"}})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, `<option value="" disabled="disabled" selected="selected">(current value not available)</option>`) {
		t.Errorf("out-of-list value must render as generic disabled placeholder: %q", out)
	}
	if strings.Contains(out, "gone") {
		t.Errorf("the out-of-list value must not be echoed into the markup: %q", out)
	}
	if strings.Contains(out, `value="p1" selected`) {
		t.Errorf("p1 must not be selected: %q", out)
	}

	// An empty current value is a different case: an optional select offers
	// a selectable empty option so the first real option is not submitted.
	s2 := &draftForm{}
	v2 := reflect.ValueOf(s2).Elem()
	comp2 := FieldDecider(mx.FieldPath("Partner"), f, v2.Field(0)).
		Render(mx.FieldPath("Partner"), f, v2.Field(0), nil)
	b.Reset()
	if err := comp2.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(b.String(), `disabled="disabled"`) || strings.Contains(b.String(), "not available") {
		t.Errorf("empty current value must not use the out-of-list placeholder: %q", b.String())
	}
	if !strings.Contains(b.String(), `<option value="" selected="selected">(none)</option>`) {
		t.Errorf("empty optional select must submit an empty option: %q", b.String())
	}
}

func TestShadcn_RequiredSelectEmptyValueChoosesNothing(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-shadcn-partners,required"`
	}
	s := &draftForm{}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("Partner")
	comp := FieldDecider(mx.FieldPath("Partner"), f, v.Field(0)).
		Render(mx.FieldPath("Partner"), f, v.Field(0), nil)
	ctx := context.WithValue(context.Background(), shadcnPartnersKey{},
		[]mx.NamedOption{{Name: "Partner One", Value: "p1"}})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, `<option value="" disabled="disabled" selected="selected">Choose…</option>`) {
		t.Errorf("required empty select must lead with Choose…: %q", out)
	}
	if strings.Contains(out, `value="p1" selected`) {
		t.Errorf("first real option must not be selected: %q", out)
	}
}

func TestShadcn_SelectEmptyOptionListRenders(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-shadcn-partners"`
	}
	s := &draftForm{Partner: "gone"}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("Partner")
	comp := FieldDecider(mx.FieldPath("Partner"), f, v.Field(0)).
		Render(mx.FieldPath("Partner"), f, v.Field(0), nil)
	ctx := context.WithValue(context.Background(), shadcnPartnersKey{}, []mx.NamedOption{})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("empty list must still render: %v", err)
	}
	out := b.String()
	if strings.Contains(out, "gone") {
		t.Errorf("stored value must not be echoed: %q", out)
	}
	if !strings.Contains(out, `(current value not available)`) {
		t.Errorf("empty list with a stored value must show the generic placeholder: %q", out)
	}
}

func TestShadcn_EnumSetAnnouncesUnlistedMembers(t *testing.T) {
	type tagForm struct {
		Tags []string `form:"options=test-shadcn-tags"`
	}
	s := &tagForm{Tags: []string{"kept", "b"}}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("Tags")
	comp := FieldDecider(mx.FieldPath("Tags"), f, v.Field(0)).
		Render(mx.FieldPath("Tags"), f, v.Field(0), nil)
	var b strings.Builder
	if err := comp.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if strings.Contains(out, "kept") {
		t.Errorf("unlisted member must not be echoed: %q", out)
	}
	if !strings.Contains(out, "(some current values are not available)") {
		t.Errorf("missing members must be announced: %q", out)
	}
	if strings.Count(out, `data-slot="checkbox"`) != 2 {
		t.Errorf("only listed options are checkboxes: %q", out)
	}
}

// Value: protects=a static shadcn enum select leads with Choose… when required and empty, and with an unselected (none) when an optional value is listed; fails_when=optionChildren skips the preface on lists that do not need context, so the browser submits the first option or the value cannot be cleared; why_new=shadcn preface tests use registry lists only and TestShadcn_EnumUsesShadcnSelect does not assert the leading option; seam=none; no_claim=html static selects are covered in html/formdecider_test.go
func TestShadcn_StaticSelectRendersPreface(t *testing.T) {
	out := renderShadcn(t, &formSample{Color: "green"}, "Color", nil)
	if !strings.Contains(out, `<option value="">(none)</option>`) {
		t.Errorf("optional listed value must offer an unselected (none): %q", out)
	}
	if strings.Contains(out, `value="" selected`) {
		t.Errorf("clear option must not be selected: %q", out)
	}
	if !strings.Contains(out, `value="green" selected`) {
		t.Errorf("listed value must stay selected: %q", out)
	}

	type requiredColor struct {
		Color colorEnum `form:"required"`
	}
	out = renderShadcn(t, &requiredColor{}, "Color", nil)
	if !strings.Contains(out, `<option value="" disabled="disabled" selected="selected">Choose…</option>`) {
		t.Errorf("required empty enum must lead with Choose…: %q", out)
	}
	if strings.Contains(out, `value="red" selected`) {
		t.Errorf("first enum value must not be selected: %q", out)
	}
}

// Value: protects=a shadcn enum set whose context option list is empty still renders and announces hidden members without echoing them; fails_when=an empty list fails the whole render or the note is skipped; why_new=TestShadcn_SelectEmptyOptionListRenders covers selects only; seam=none; no_claim=the preface decision itself is TestEnumSetPreface
func TestShadcn_EnumSetEmptyListWithMembersRenders(t *testing.T) {
	type tagForm struct {
		Tags []string `form:"options=test-shadcn-partners"`
	}
	s := &tagForm{Tags: []string{"kept"}}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("Tags")
	comp := FieldDecider(mx.FieldPath("Tags"), f, v.Field(0)).
		Render(mx.FieldPath("Tags"), f, v.Field(0), nil)
	ctx := context.WithValue(context.Background(), shadcnPartnersKey{}, []mx.NamedOption{})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("empty list must still render: %v", err)
	}
	out := b.String()
	if strings.Contains(out, "kept") {
		t.Errorf("unlisted member must not be echoed: %q", out)
	}
	if strings.Contains(out, `data-slot="checkbox"`) {
		t.Errorf("empty list must not render a checkbox: %q", out)
	}
	if !strings.Contains(out, "(some current values are not available)") {
		t.Errorf("missing members must be announced: %q", out)
	}
}

type shadeCode int

const shadeRed shadeCode = 1

func (c shadeCode) String() string {
	if c == shadeRed {
		return "red"
	}
	return fmt.Sprintf("%d", int(c))
}

func (shadeCode) Enums() []shadeCode { return []shadeCode{shadeRed} }

func TestShadcn_EnumSetStringerMatchesEnums(t *testing.T) {
	type form struct {
		Colors []shadeCode
	}
	s := &form{Colors: []shadeCode{shadeRed}}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("Colors")
	comp := FieldDecider(mx.FieldPath("Colors"), f, v.Field(0)).
		Render(mx.FieldPath("Colors"), f, v.Field(0), nil)
	var b strings.Builder
	if err := comp.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if strings.Contains(out, `value="1"`) || !strings.Contains(out, `value="red"`) {
		t.Fatalf("Stringer form must be the checkbox value: %s", out)
	}
	red := strings.Index(out, `value="red"`)
	start := strings.LastIndex(out[:red], "<input")
	end := strings.Index(out[red:], ">")
	if !strings.Contains(out[start:red+end], "checked") {
		t.Fatalf("stored red must be checked: %s", out[start:red+end])
	}
}

// marshaledID's text form is hex, not fmt's "[222 173 190 239]".
type marshaledID [4]byte

func (id marshaledID) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%02x%02x%02x%02x", id[0], id[1], id[2], id[3]), nil
}

// Value: protects=a shadcn enum-set checkbox is checked when the stored key's MarshalText equals the option value; fails_when=setMembers uses fmt.Sprint, so that checkbox renders unchecked and a later save drops the member; why_new=the html round-trip does not render shadcn checkboxes and shadcn enum-set tests use plain strings; seam=none; no_claim=parse and preserve stay in the html FieldDecider
func TestShadcn_EnumSetChecksMarshaledMember(t *testing.T) {
	type idForm struct {
		IDs map[marshaledID]struct{} `form:"options=test-shadcn-partners"`
	}
	listed := marshaledID{0xde, 0xad, 0xbe, 0xef}
	hidden := marshaledID{0x01, 0x02, 0x03, 0x04}
	s := &idForm{IDs: map[marshaledID]struct{}{listed: {}, hidden: {}}}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("IDs")
	comp := FieldDecider(mx.FieldPath("IDs"), f, v.Field(0)).
		Render(mx.FieldPath("IDs"), f, v.Field(0), nil)
	ctx := context.WithValue(context.Background(), shadcnPartnersKey{}, []mx.NamedOption{
		{Name: "Dead Beef", Value: "deadbeef"},
		{Name: "Other", Value: "11111111"},
	})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	i := strings.Index(out, `value="deadbeef"`)
	if i < 0 {
		t.Fatalf("missing deadbeef option: %s", out)
	}
	start := strings.LastIndex(out[:i], "<")
	end := strings.Index(out[i:], ">")
	tag := out[start : i+end]
	if !strings.Contains(tag, `checked="checked"`) {
		t.Fatalf("marshaled member must be checked: %s", tag)
	}
	if strings.Contains(out, "01020304") || strings.Contains(out, "[222 173 190 239]") {
		t.Fatalf("hidden or sprint form leaked: %s", out)
	}
}

func TestShadcn_EnumSetWithRegistryOptions(t *testing.T) {
	type tagForm struct {
		Tags []string `form:"options=test-shadcn-tags"`
	}
	s := &tagForm{Tags: []string{"b"}}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("Tags")
	beh := FieldDecider(mx.FieldPath("Tags"), f, v.Field(0))
	comp := beh.Render(mx.FieldPath("Tags"), f, v.Field(0), nil)

	var b strings.Builder
	if err := comp.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if strings.Count(out, `data-slot="checkbox"`) != 2 {
		t.Errorf("expected one shadcn checkbox per registry option: %q", out)
	}
	if !strings.Contains(out, "Tag A") || !strings.Contains(out, "Tag B") {
		t.Errorf("expected option labels: %q", out)
	}
	if !strings.Contains(out, "checked") {
		t.Errorf("expected current member b checked: %q", out)
	}
	if strings.Contains(out, "not available") {
		t.Errorf("listed members must not show the hidden-member note: %q", out)
	}
}

// The shadcn renderer keeps its own displayValue, so the zero uuid.UUID and
// the zero time.Time must render as no value here too: the select leads with
// Choose… and the inputs stay empty.
func TestShadcn_ZeroValueRendersAsNoValue(t *testing.T) {
	partner := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	ctx := context.WithValue(context.Background(), shadcnPartnersKey{},
		[]mx.NamedOption{{Name: "Partner One", Value: partner.String()}})
	render := func(t *testing.T, target any) string {
		t.Helper()
		v := reflect.ValueOf(target).Elem()
		field := v.Type().Field(0)
		path := mx.FieldPath(field.Name)
		comp := FieldDecider(path, field, v.Field(0)).Render(path, field, v.Field(0), nil)
		var b strings.Builder
		if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
			t.Fatalf("render: %v", err)
		}
		return b.String()
	}

	t.Run("required select", func(t *testing.T) {
		out := render(t, &struct {
			Partner uuid.UUID `form:"options=test-shadcn-partners,required"`
		}{})
		if !strings.Contains(out, `<option value="" disabled="disabled" selected="selected">Choose…</option>`) {
			t.Errorf("a required select without a value must lead with Choose…: %q", out)
		}
		if strings.Contains(out, "not available") {
			t.Errorf("the zero id is no value, not an unlisted one: %q", out)
		}
	})
	t.Run("text uuid", func(t *testing.T) {
		out := render(t, &struct {
			Code uuid.UUID
		}{})
		if strings.Contains(out, "00000000-0000") {
			t.Errorf("the zero id must render as an empty value: %q", out)
		}
	})
	t.Run("datetime", func(t *testing.T) {
		out := render(t, &struct {
			Due time.Time
		}{})
		if strings.Contains(out, "0001-01-01") {
			t.Errorf("the zero time must render as an empty value: %q", out)
		}
	})
}

// time.Time fields render as they did before form:"widget=date" learned
// string-kind fields, except that min and max now reach the date input as
// they do in the html renderer. Only the <input> start tag up to its class
// is compared; the class string is the shadcn Input's own.
func TestShadcn_TimeWidgets(t *testing.T) {
	type form struct {
		Date     time.Time `form:"widget=date,min=2026-01-01,max=2026-12-31,required"`
		DateTime time.Time `form:"widget=datetime"`
		Time     time.Time `form:"widget=time,readonly"`
		Default  time.Time
		Nil      *time.Time `form:"widget=date"`
		Zero     time.Time  `form:"widget=date"`
	}
	when := time.Date(2026, 7, 6, 9, 5, 7, 0, time.UTC)
	f := &form{Date: when, DateTime: when, Time: when, Default: when}
	renders := []struct{ field, want string }{
		{"Date", `<input data-slot="input" type="date" name="Date" id="Date" hx-trigger="change" value="2026-07-06" min="2026-01-01" max="2026-12-31" required="required" class="`},
		{"DateTime", `<input data-slot="input" type="datetime-local" name="DateTime" id="DateTime" hx-trigger="change" value="2026-07-06T09:05:07" class="`},
		{"Time", `<input data-slot="input" type="time" name="Time" id="Time" hx-trigger="change" value="09:05:07" readonly="readonly" class="`},
		{"Default", `<input data-slot="input" type="datetime-local" name="Default" id="Default" hx-trigger="change" value="2026-07-06T09:05:07" class="`},
		{"Nil", `<input data-slot="input" type="date" name="Nil" id="Nil" hx-trigger="change" class="`},
		{"Zero", `<input data-slot="input" type="date" name="Zero" id="Zero" hx-trigger="change" class="`},
	}
	for _, r := range renders {
		if got := inputStartTag(renderShadcn(t, f, r.field, nil)); got != r.want {
			t.Errorf("%s renders\n%s\nwant\n%s", r.field, got, r.want)
		}
	}
}

// testDate mimics a date type with an underlying string, like the
// date.Date of github.com/domonda/go-types, without depending on it.
// Format(time.DateOnly) returns a valid date normalized, "" for an empty
// or zero date, and an invalid value unchanged. Like date.Date it has no
// UnmarshalText, so binding sets the string itself.
type testDate string

func (d testDate) Format(layout string) string {
	for _, l := range []string{time.DateOnly, "02.01.2006"} {
		if t, err := time.Parse(l, string(d)); err == nil {
			if t.IsZero() {
				return ""
			}
			return t.Format(layout)
		}
	}
	return string(d)
}

// unmarshaledDate binds through UnmarshalText, which marks what it stores
// so a test can tell it ran.
type unmarshaledDate string

func (d *unmarshaledDate) UnmarshalText(text []byte) error {
	*d = unmarshaledDate("unmarshaled " + string(text))
	return nil
}

// dateStrings holds a date as text in each string-kind shape a
// form:"widget=date" field accepts.
type dateStrings struct {
	Str       string          `form:"widget=date"`
	Ptr       *string         `form:"widget=date"`
	Date      testDate        `form:"widget=date"`
	DatePtr   *testDate       `form:"widget=date"`
	Bounded   string          `form:"widget=date,min=2026-01-01,max=2026-12-31,required"`
	Unmarshal unmarshaledDate `form:"widget=date"`
	Localized string          `form:"widget=date,placeholder=JJJJ-MM-TT"`
}

// The shadcn renderer prefills a date held as text like the html one: the
// stored date in a date input, or the stored text in a text input when a
// date input cannot show it.
func TestShadcn_DateWidgetPrefillsStringKinds(t *testing.T) {
	tests := []struct {
		name  string
		form  dateStrings
		field string
		want  string
	}{
		// A Format method gives the date as yyyy-mm-dd.
		{"Format normalizes", dateStrings{Date: "06.07.2026"}, "Date",
			`<input data-slot="input" type="date" name="Date" id="Date" hx-trigger="change" value="2026-07-06" class="`},
		{"Format through pointer", dateStrings{DatePtr: new(testDate("06.07.2026"))}, "DatePtr",
			`<input data-slot="input" type="date" name="DatePtr" id="DatePtr" hx-trigger="change" value="2026-07-06" class="`},
		// Without Format, a yyyy-mm-dd string is shown as stored.
		{"string", dateStrings{Str: "2026-07-06"}, "Str",
			`<input data-slot="input" type="date" name="Str" id="Str" hx-trigger="change" value="2026-07-06" class="`},
		{"string pointer", dateStrings{Ptr: new("2026-07-06")}, "Ptr",
			`<input data-slot="input" type="date" name="Ptr" id="Ptr" hx-trigger="change" value="2026-07-06" class="`},
		// No date, no value attribute.
		{"empty string", dateStrings{}, "Str",
			`<input data-slot="input" type="date" name="Str" id="Str" hx-trigger="change" class="`},
		{"nil pointer", dateStrings{}, "Ptr",
			`<input data-slot="input" type="date" name="Ptr" id="Ptr" hx-trigger="change" class="`},
		{"Format zero date", dateStrings{Date: "0001-01-01"}, "Date",
			`<input data-slot="input" type="date" name="Date" id="Date" hx-trigger="change" class="`},
		// Anything else renders as a text input showing the stored text.
		{"non-canonical string", dateStrings{Str: "06.07.2026"}, "Str",
			`<input data-slot="input" type="text" name="Str" id="Str" hx-trigger="change" value="06.07.2026" placeholder="YYYY-MM-DD" aria-describedby="Str-format" aria-invalid="true" class="`},
		{"text through pointer", dateStrings{Ptr: new("next week")}, "Ptr",
			`<input data-slot="input" type="text" name="Ptr" id="Ptr" hx-trigger="change" value="next week" placeholder="YYYY-MM-DD" aria-describedby="Ptr-format" aria-invalid="true" class="`},
		{"Format returns it unchanged", dateStrings{Date: "garbage"}, "Date",
			`<input data-slot="input" type="text" name="Date" id="Date" hx-trigger="change" value="garbage" placeholder="YYYY-MM-DD" aria-describedby="Date-format" aria-invalid="true" class="`},
		// The date input carries the range, the text input only required.
		{"bounded date", dateStrings{Bounded: "2026-07-06"}, "Bounded",
			`<input data-slot="input" type="date" name="Bounded" id="Bounded" hx-trigger="change" value="2026-07-06" min="2026-01-01" max="2026-12-31" required="required" class="`},
		{"bounded text", dateStrings{Bounded: "06.07.2026"}, "Bounded",
			`<input data-slot="input" type="text" name="Bounded" id="Bounded" hx-trigger="change" value="06.07.2026" placeholder="YYYY-MM-DD" aria-describedby="Bounded-format" required="required" aria-invalid="true" class="`},
		// A placeholder from the tag, say a translated format, replaces the
		// placeholder; the hint names the format the parser reads.
		{"tag placeholder", dateStrings{Localized: "06.07.2026"}, "Localized",
			`<input data-slot="input" type="text" name="Localized" id="Localized" hx-trigger="change" value="06.07.2026" placeholder="JJJJ-MM-TT" aria-describedby="Localized-format" aria-invalid="true" class="`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := renderShadcn(t, &tt.form, tt.field, nil)
			if got := inputStartTag(out); got != tt.want {
				t.Errorf("rendered\n%s\nwant\n%s", got, tt.want)
			}
			// A text input is followed by the format to enter.
			hint := `id="` + tt.field + `-format"`
			// Every text input here holds text the parser rejects, and the
			// hint names the parser's format, also under a tag placeholder.
			if strings.Contains(tt.want, `type="text"`) {
				if !strings.Contains(out, hint) || !strings.Contains(out, ">Not in the expected format: YYYY-MM-DD</p>") {
					t.Errorf("the text input lacks the format hint: %s", out)
				}
			} else if strings.Contains(out, hint) {
				t.Errorf("a date input needs no format hint: %s", out)
			}
		})
	}
}

// shadcn binds through the html parser, so a date held as text binds the
// same way: an empty submission clears, a yyyy-mm-dd date is stored (via
// UnmarshalText when the type has one), and any other text is an error
// with the posted text kept in the field for the re-render.
func TestShadcn_DateWidgetBindsStringKinds(t *testing.T) {
	tests := []struct {
		name    string
		stored  dateStrings
		field   string
		raw     string
		want    dateStrings
		wantErr string
	}{
		{name: "empty clears pointer", stored: dateStrings{Ptr: new("2026-07-06")}, field: "Ptr", raw: "",
			want: dateStrings{}},
		{name: "string", field: "Str", raw: "2026-07-06",
			want: dateStrings{Str: "2026-07-06"}},
		{name: "replaces non-canonical date", stored: dateStrings{Date: "06.07.2026"}, field: "Date", raw: "2026-07-06",
			want: dateStrings{Date: "2026-07-06"}},
		{name: "UnmarshalText", field: "Unmarshal", raw: "2026-07-06",
			want: dateStrings{Unmarshal: "unmarshaled 2026-07-06"}},
		{name: "nil pointer gets the posted text", field: "Ptr", raw: "06.07.2026",
			want:    dateStrings{Ptr: new("06.07.2026")},
			wantErr: `"06.07.2026" is not a date in the format YYYY-MM-DD`},
		{name: "date gets the posted text", stored: dateStrings{Date: "2026-07-06"}, field: "Date", raw: "next week",
			want:    dateStrings{Date: "next week"},
			wantErr: `"next week" is not a date in the format YYYY-MM-DD`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stored
			values := url.Values{}
			values.Set(tt.field, tt.raw)
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			v := reflect.ValueOf(&got).Elem()
			field, _ := v.Type().FieldByName(tt.field)
			path := mx.FieldPath(tt.field)
			value := v.FieldByIndex(field.Index)
			err := FieldDecider(path, field, value).Parse(path, field, value, r)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bind %q: stored %+v, want %+v", tt.raw, got, tt.want)
			}
			switch {
			case tt.wantErr == "" && err != nil:
				t.Errorf("bind %q: %v", tt.raw, err)
			case tt.wantErr != "" && (err == nil || err.Error() != tt.wantErr):
				t.Errorf("bind %q: error %v, want %s", tt.raw, err, tt.wantErr)
			}
		})
	}
}

// A GET followed by a POST without edits stores exactly the dates the GET
// showed: a valid but non-canonical date comes back canonical, and a
// missing one stays missing.
func TestShadcn_DateStringRoundTrip(t *testing.T) {
	type record struct {
		Start  testDate  `form:"widget=date"`
		End    string    `form:"widget=date"`
		Review *testDate `form:"widget=date"`
		Signed *string   `form:"widget=date"`
	}
	var saved *record
	handler := mx.ReflectFormHandler(
		func(context.Context) (*record, error) {
			return &record{Start: "06.07.2026", End: "2026-08-05", Signed: new("2026-07-20")}, nil
		},
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/record", nil))
	body := rec.Body.String()
	shown := submittedValues(body)
	for _, f := range []struct{ name, value string }{
		{"Start", "2026-07-06"},
		{"End", "2026-08-05"},
		{"Review", ""},
		{"Signed", "2026-07-20"},
	} {
		if !strings.Contains(body, `type="date" name="`+f.name+`"`) {
			t.Errorf("GET renders %s without a date input: %s", f.name, body)
		}
		if got := shown.Get(f.name); got != f.value {
			t.Errorf("GET shows %s %q, want %q", f.name, got, f.value)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/record", strings.NewReader(shown.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("unchanged POST: status %d, want 303, body %s", rec.Code, rec.Body.String())
	}
	want := &record{Start: "2026-07-06", End: "2026-08-05", Signed: new("2026-07-20")}
	if !reflect.DeepEqual(saved, want) {
		t.Errorf("saved %+v, want %+v", saved, want)
	}
}

// inputStartTag returns the first <input> start tag of out up to and
// including ` class="`, or all of it when it has no class.
func inputStartTag(out string) string {
	start := strings.Index(out, "<input")
	if start < 0 {
		return ""
	}
	tag := out[start:]
	if end := strings.Index(tag, ` class="`); end >= 0 {
		return tag[:end+len(` class="`)]
	}
	if end := strings.Index(tag, ">"); end >= 0 {
		return tag[:end+1]
	}
	return tag
}

var (
	inputTagRe  = regexp.MustCompile(`<input((?:\s+[^\s=/>]+="[^"]*")*)\s*/?>`)
	attributeRe = regexp.MustCompile(`([^\s=]+)="([^"]*)"`)
)

// submittedValues returns what a browser submits for the <input> elements
// of body: the value of each named input, "" when it has none. An
// unchecked checkbox or radio button submits nothing.
func submittedValues(body string) url.Values {
	values := url.Values{}
	for _, input := range inputTagRe.FindAllStringSubmatch(body, -1) {
		attrs := map[string]string{}
		for _, a := range attributeRe.FindAllStringSubmatch(input[1], -1) {
			attrs[a[1]] = html.UnescapeString(a[2])
		}
		name, ok := attrs["name"]
		if !ok {
			continue
		}
		if typ := attrs["type"]; typ == "checkbox" || typ == "radio" {
			if _, checked := attrs["checked"]; !checked {
				continue
			}
		}
		values.Add(name, attrs["value"])
	}
	return values
}

// The shadcn renderer shows datetime and time text like the html one: the
// input's own form with zero seconds dropped, or a text input plus the
// format to enter.
func TestShadcn_TimeWidgetsPrefillStringKinds(t *testing.T) {
	type form struct {
		At    string `form:"widget=datetime"`
		Clock string `form:"widget=time"`
	}
	f := &form{At: "2026-07-06T09:05:00", Clock: "9.05"}
	if got, want := inputStartTag(renderShadcn(t, f, "At", nil)),
		`<input data-slot="input" type="datetime-local" name="At" id="At" hx-trigger="change" value="2026-07-06T09:05" class="`; got != want {
		t.Errorf("At renders\n%s\nwant\n%s", got, want)
	}
	out := renderShadcn(t, f, "Clock", nil)
	if got, want := inputStartTag(out),
		`<input data-slot="input" type="text" name="Clock" id="Clock" hx-trigger="change" value="9.05" placeholder="HH:MM" aria-describedby="Clock-format" aria-invalid="true" class="`; got != want {
		t.Errorf("Clock renders\n%s\nwant\n%s", got, want)
	}
	if !strings.Contains(out, ">Not in the expected format: HH:MM</p>") {
		t.Errorf("Clock lacks the format hint: %s", out)
	}
	if got, want := inputStartTag(renderShadcn(t, &form{Clock: "09:05:00"}, "Clock", nil)),
		`<input data-slot="input" type="time" name="Clock" id="Clock" hx-trigger="change" value="09:05" class="`; got != want {
		t.Errorf("valid Clock renders\n%s\nwant\n%s", got, want)
	}
}

// hintedDate declares itself a date, so its fields need no widget tag.
type hintedDate string

func (hintedDate) FormWidget() string { return "date" }

// A type whose FormWidget declares "date" renders a shadcn date input like
// a field tagged form:"widget=date", also behind a nil pointer.
func TestShadcn_HintedDateWidget(t *testing.T) {
	type form struct {
		Due    hintedDate
		DuePtr *hintedDate
	}
	f := &form{Due: "2026-07-06"}
	if got, want := inputStartTag(renderShadcn(t, f, "Due", nil)),
		`<input data-slot="input" type="date" name="Due" id="Due" hx-trigger="change" value="2026-07-06" class="`; got != want {
		t.Errorf("Due renders\n%s\nwant\n%s", got, want)
	}
	if got, want := inputStartTag(renderShadcn(t, f, "DuePtr", nil)),
		`<input data-slot="input" type="date" name="DuePtr" id="DuePtr" hx-trigger="change" class="`; got != want {
		t.Errorf("DuePtr renders\n%s\nwant\n%s", got, want)
	}
}

// consent declares the checkbox widget.
type consent bool

func (consent) FormWidget() string { return "checkbox" }

// A FormWidgetHint picks the bool control like a tag does: "checkbox"
// renders a Checkbox, not the default Switch.
func TestShadcn_HintedCheckbox(t *testing.T) {
	type form struct {
		Agree consent
	}
	out := renderShadcn(t, &form{}, "Agree", nil)
	if !strings.Contains(out, `data-slot="checkbox"`) || strings.Contains(out, `data-slot="switch"`) {
		t.Errorf("a checkbox hint must render a Checkbox: %s", out)
	}
}

// After a failed bind the date control gets the destructive border like
// every other shadcn control: aria-invalid on the input, data-invalid on
// the Field.
func TestShadcn_DateWidgetMarksErrors(t *testing.T) {
	out := renderShadcn(t, &dateStrings{Str: "06.07.2026"}, "Str",
		[]error{stringErr(`"06.07.2026" is not a date in the format YYYY-MM-DD`)})
	if got := inputStartTag(out); !strings.Contains(got, `aria-invalid="true"`) {
		t.Errorf("the input lacks aria-invalid: %s", got)
	}
	if !strings.Contains(out, `data-invalid="true"`) {
		t.Errorf("the Field lacks data-invalid: %s", out)
	}
	// A date the picker shows, such as one outside the range, is marked
	// invalid by the error alone.
	picker := inputStartTag(renderShadcn(t, &dateStrings{Bounded: "1900-01-01"}, "Bounded",
		[]error{stringErr(`"1900-01-01" is earlier than 2026-01-01`)}))
	if !strings.Contains(picker, `type="date"`) || !strings.Contains(picker, `aria-invalid="true"`) {
		t.Errorf("the date input lacks aria-invalid after an error: %s", picker)
	}
}

// emailAddress declares the email widget.
type emailAddress string

func (emailAddress) FormWidget() string { return "email" }

// shadcn reads a FormWidgetHint's input type from the tag DetectField
// fills, and an explicit widget tag still wins over the hint.
func TestShadcn_HintedInputType(t *testing.T) {
	type form struct {
		Contact emailAddress
		Code    emailAddress `form:"widget=text"`
	}
	f := &form{Contact: "a@example.com", Code: "x"}
	if got := inputStartTag(renderShadcn(t, f, "Contact", nil)); !strings.Contains(got, `type="email"`) {
		t.Errorf("the hint must render an email input: %s", got)
	}
	if got := inputStartTag(renderShadcn(t, f, "Code", nil)); !strings.Contains(got, `type="text"`) {
		t.Errorf("widget=text must win over the email hint: %s", got)
	}
}

// definedTime is defined as time.Time, without its methods.
type definedTime time.Time

// looseDay writes 2026-7-6, a spelling the date input cannot show, and
// reads it back.
type looseDay struct {
	Year, Month, Day int
}

func (d looseDay) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%d-%d-%d", d.Year, d.Month, d.Day), nil
}

func (d *looseDay) UnmarshalText(text []byte) error {
	t, err := time.Parse("2006-1-2", string(text))
	if err != nil {
		return err
	}
	*d = looseDay{t.Year(), int(t.Month()), t.Day()}
	return nil
}

// price is a struct with text methods, so a field of it is a catch-all
// text input.
type price struct {
	Cents int
}

func (p price) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%d.%02d", p.Cents/100, p.Cents%100), nil
}

func (p *price) UnmarshalText(text []byte) error {
	var units, cents int
	if _, err := fmt.Sscanf(string(text), "%d.%d", &units, &cents); err != nil {
		return err
	}
	*p = price{units*100 + cents}
	return nil
}

// boxedPrice embeds a pointer to a price; its promoted MarshalText panics on
// a nil one.
type boxedPrice struct {
	*price
}

// A catch-all value renders as a styled text input in a Field, like a
// string, with its errors marked, not as the unstyled html fallback.
func TestShadcn_CatchAllIsStyledText(t *testing.T) {
	type form struct {
		Total price
		Box   boxedPrice
	}
	f := &form{Total: price{1250}}
	if got := inputStartTag(renderShadcn(t, f, "Box", nil)); strings.Contains(got, "value=") {
		t.Errorf("a panicking MarshalText must render no value: %s", got)
	}
	out := renderShadcn(t, f, "Total", nil)
	if got, want := inputStartTag(out),
		`<input data-slot="input" type="text" name="Total" id="Total" hx-trigger="change" value="12.50" class="`; got != want {
		t.Errorf("Total renders\n%s\nwant\n%s", got, want)
	}
	if !strings.Contains(out, `data-slot="field"`) {
		t.Errorf("the input must sit in a Field: %s", out)
	}
	errOut := renderShadcn(t, f, "Total", []error{errors.New("bad price")})
	if !strings.Contains(inputStartTag(errOut), `aria-invalid="true"`) || !strings.Contains(errOut, "bad price") {
		t.Errorf("errors must mark the input and show: %s", errOut)
	}
}

// A readonly file input is disabled and never required.
func TestShadcn_ReadonlyFileDisabled(t *testing.T) {
	type form struct {
		Doc []byte `form:"widget=file,required,readonly"`
	}
	if got := inputStartTag(renderShadcn(t, &form{}, "Doc", nil)); !strings.Contains(got, `disabled="disabled"`) || strings.Contains(got, `required="required"`) {
		t.Errorf("a readonly file input renders %s", got)
	}
}

// A time shown as text is written as RFC 3339, which its parser reads,
// also for a type defined as time.Time that has no text methods of its own.
func TestShadcn_TextTimeIsRFC3339(t *testing.T) {
	type stamp time.Time
	type form struct {
		At    time.Time `form:"widget=text"`
		Stamp stamp     `form:"widget=text,readonly"`
	}
	at := time.Date(2026, 7, 6, 9, 5, 7, 0, time.FixedZone("CEST", 2*60*60))
	for _, name := range []string{"At", "Stamp"} {
		if got := inputStartTag(renderShadcn(t, &form{At: at, Stamp: stamp(at)}, name, nil)); !strings.Contains(got, `value="2026-07-06T09:05:07+02:00"`) {
			t.Errorf("%s renders %s", name, got)
		}
		// The zero time, as on a new record, shows an empty input.
		if got := inputStartTag(renderShadcn(t, &form{}, name, nil)); strings.Contains(got, "value=") {
			t.Errorf("zero %s renders %s", name, got)
		}
	}
}

// A sensitive number never shows its value; any other number does, or an
// unchanged save would post it empty.
func TestShadcn_SensitiveNumberHidesValue(t *testing.T) {
	type form struct {
		PIN int `form:"sensitive"`
		Qty int
	}
	if out := renderShadcn(t, &form{PIN: 1234}, "PIN", nil); strings.Contains(out, "1234") {
		t.Errorf("a sensitive number shows its value: %s", out)
	}
	if got := inputStartTag(renderShadcn(t, &form{Qty: 1234}, "Qty", nil)); !strings.Contains(got, `value="1234"`) {
		t.Errorf("a number renders %s", got)
	}
}

// A blank sensitive input keeps the stored value, so a sensitive pointer
// offers the clear control, as a nullable field does.
func TestShadcn_SensitivePointerHasClearControl(t *testing.T) {
	type form struct {
		Token *string    `form:"sensitive"`
		N     *int       `form:"sensitive"`
		At    *time.Time `form:"widget=date,sensitive"`
		Plain *string
		Flag  *bool `form:"sensitive"`
		// A select shows its stored option and clears with "(none)".
		Color *colorEnum `form:"sensitive"`
	}
	f := &form{Token: new("secret"), N: new(3), At: new(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)), Plain: new("p"), Flag: new(true), Color: new(colorEnum("red"))}
	for field, want := range map[string]bool{"Token": true, "N": true, "At": true, "Plain": false, "Flag": false, "Color": false} {
		if got := strings.Contains(renderShadcn(t, f, field, nil), mx.ClearSentinelName(mx.FieldPath(field))); got != want {
			t.Errorf("%s: clear control %t, want %t", field, got, want)
		}
	}
}

// A readonly nullable field is never written, so it offers no clear control.
func TestShadcn_ReadonlyNullableHasNoClearControl(t *testing.T) {
	type form struct {
		Locked nullableInt `form:"readonly"`
		Open   nullableInt
	}
	f := &form{Locked: nullableInt{v: 1}, Open: nullableInt{v: 2}}
	if out := renderShadcn(t, f, "Locked", nil); strings.Contains(out, mx.ClearSentinelName("Locked")) {
		t.Errorf("a readonly field offers a clear control: %s", out)
	}
	if out := renderShadcn(t, f, "Open", nil); !strings.Contains(out, mx.ClearSentinelName("Open")) {
		t.Errorf("a nullable field lacks its clear control: %s", out)
	}
}

// shadcn renders the wrapper kinds and the readonly text input like the
// html renderer: a defined time in the picker, a readonly fallback without
// the invalid look or the format hint.
func TestShadcn_DateKindsRender(t *testing.T) {
	type form struct {
		At      definedTime `form:"widget=datetime"`
		Old     string      `form:"widget=date,readonly"`
		Loose   looseDay    `form:"widget=date"`
		Bounded looseDay    `form:"widget=date,min=2026-01-01"`
		Seconds string      `form:"widget=datetime,min=2026-01-01T00:00"`
	}
	f := &form{
		At: definedTime(time.Date(2026, 7, 6, 9, 5, 7, 0, time.UTC)), Old: "06.07.2026",
		Loose: looseDay{2026, 7, 6}, Bounded: looseDay{2026, 7, 6}, Seconds: "2026-07-06T09:05:07",
	}
	// A date struct reads its own text back, so its text input is not
	// marked invalid, unless a bound it cannot be checked against is set.
	if out := renderShadcn(t, f, "Loose", nil); !strings.Contains(out, ">Expected format: YYYY-MM-DD</p>") {
		t.Errorf("Loose lacks the plain format hint: %s", out)
	}
	if got, want := inputStartTag(renderShadcn(t, f, "Loose", nil)),
		`<input data-slot="input" type="text" name="Loose" id="Loose" hx-trigger="change" value="2026-7-6" placeholder="YYYY-MM-DD" aria-describedby="Loose-format" class="`; got != want {
		t.Errorf("Loose renders\n%s\nwant\n%s", got, want)
	}
	bounded := renderShadcn(t, f, "Bounded", nil)
	if got := inputStartTag(bounded); !strings.Contains(got, `aria-invalid="true"`) || strings.Contains(got, "min=") {
		t.Errorf("Bounded renders %s, want an invalid text input without min", got)
	}
	// The Field root matches the control, so the label turns destructive too.
	if !strings.Contains(bounded, `data-invalid="true"`) {
		t.Errorf("Bounded's Field lacks data-invalid: %s", bounded)
	}
	if loose := renderShadcn(t, f, "Loose", nil); strings.Contains(loose, `data-invalid`) {
		t.Errorf("Loose is not invalid: %s", loose)
	}
	// With a min, a value with seconds gets step 1.
	if got, want := inputStartTag(renderShadcn(t, f, "Seconds", nil)),
		`<input data-slot="input" type="datetime-local" name="Seconds" id="Seconds" hx-trigger="change" value="2026-07-06T09:05:07" min="2026-01-01T00:00" step="1" class="`; got != want {
		t.Errorf("Seconds renders\n%s\nwant\n%s", got, want)
	}
	if got, want := inputStartTag(renderShadcn(t, f, "At", nil)),
		`<input data-slot="input" type="datetime-local" name="At" id="At" hx-trigger="change" value="2026-07-06T09:05:07" class="`; got != want {
		t.Errorf("At renders\n%s\nwant\n%s", got, want)
	}
	out := renderShadcn(t, f, "Old", nil)
	if got, want := inputStartTag(out),
		`<input data-slot="input" type="text" name="Old" id="Old" hx-trigger="change" value="06.07.2026" readonly="readonly" class="`; got != want {
		t.Errorf("Old renders\n%s\nwant\n%s", got, want)
	}
	if strings.Contains(out, "Expected format") || strings.Contains(out, "data-invalid") {
		t.Errorf("a readonly text input needs no format hint and is never marked invalid: %s", out)
	}
}
