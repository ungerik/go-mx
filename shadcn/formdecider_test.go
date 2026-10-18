package shadcn

import (
	"context"
	"fmt"
	"reflect"
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
