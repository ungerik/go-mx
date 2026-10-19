package html

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	stdhtml "html"
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

type sampleNullable struct {
	v    string
	null bool
}

func (n sampleNullable) IsNull() bool { return n.null }
func (n *sampleNullable) SetNull()    { n.v = ""; n.null = true }

type allKindsForm struct {
	Name      string
	Age       int
	Active    bool
	Notes     string `form:"widget=textarea,help=Long description"`
	When      time.Time
	Color     colorEnum
	HiddenID  string `form:"hidden"`
	Optional  sampleNullable
	Sensitive string `form:"sensitive"`
	Required  string `form:"required,placeholder=Type here"`
}

func renderToString(t *testing.T, c mx.Component) string {
	t.Helper()
	if c == nil {
		return ""
	}
	var b strings.Builder
	w := mx.NewCheckedWriter(&b)
	if err := c.Render(context.Background(), w); err != nil {
		t.Fatalf("render: %v", err)
	}
	return b.String()
}

func renderFieldHTML(t *testing.T, target any, fieldName string) string {
	t.Helper()
	v := reflect.ValueOf(target).Elem()
	field, ok := v.Type().FieldByName(fieldName)
	if !ok {
		t.Fatalf("no field %q", fieldName)
	}
	beh := FieldDecider(mx.FieldPath(fieldName), field, v.FieldByName(fieldName))
	if beh.Render == nil {
		t.Fatalf("Render closure nil for %q", fieldName)
	}
	return renderToString(t, beh.Render(mx.FieldPath(fieldName), field, v.FieldByName(fieldName), nil))
}

func TestFieldDecider_StringInput(t *testing.T) {
	out := renderFieldHTML(t, &allKindsForm{Name: "Alice"}, "Name")
	if !strings.Contains(out, `type="text"`) ||
		!strings.Contains(out, `name="Name"`) ||
		!strings.Contains(out, `value="Alice"`) {
		t.Errorf("string input: %q", out)
	}
}

func TestFieldDecider_NumberInputWithMinMax(t *testing.T) {
	type withRange struct {
		N int `form:"min=4,max=6,step=1"`
	}
	out := renderFieldHTML(t, &withRange{N: 5}, "N")
	for _, want := range []string{`type="number"`, `min="4"`, `max="6"`, `step="1"`, `value="5"`} {
		if !strings.Contains(out, want) {
			t.Errorf("number input missing %q in %q", want, out)
		}
	}
}

func TestFieldDecider_CheckboxBool(t *testing.T) {
	out := renderFieldHTML(t, &allKindsForm{Active: true}, "Active")
	if !strings.Contains(out, `type="checkbox"`) {
		t.Errorf("expected checkbox: %q", out)
	}
	if !strings.Contains(out, `checked`) {
		t.Errorf("expected checked attribute: %q", out)
	}
}

func TestFieldDecider_TextareaWidget(t *testing.T) {
	out := renderFieldHTML(t, &allKindsForm{Notes: "hi"}, "Notes")
	if !strings.Contains(out, `<textarea`) {
		t.Errorf("expected textarea element: %q", out)
	}
	if !strings.Contains(out, "Long description") {
		t.Errorf("expected help text: %q", out)
	}
}

func TestFieldDecider_DateTime(t *testing.T) {
	when := time.Date(2026, 5, 25, 10, 30, 0, 0, time.UTC)
	out := renderFieldHTML(t, &allKindsForm{When: when}, "When")
	if !strings.Contains(out, `type="datetime-local"`) {
		t.Errorf("expected datetime-local: %q", out)
	}
	if !strings.Contains(out, "2026-05-25T10:30:00") {
		t.Errorf("expected ISO datetime in value: %q", out)
	}
}

func TestFieldDecider_EnumSelect(t *testing.T) {
	out := renderFieldHTML(t, &allKindsForm{Color: "green"}, "Color")
	if !strings.Contains(out, "<select") {
		t.Errorf("expected select element: %q", out)
	}
	for _, opt := range []string{"red", "green", "blue"} {
		if !strings.Contains(out, opt) {
			t.Errorf("missing option %q in %q", opt, out)
		}
	}
	if !strings.Contains(out, `value="green" selected`) {
		t.Errorf("green should be selected: %q", out)
	}
}

func TestFieldDecider_HiddenField(t *testing.T) {
	out := renderFieldHTML(t, &allKindsForm{HiddenID: "abc-123"}, "HiddenID")
	if !strings.Contains(out, `type="hidden"`) || !strings.Contains(out, "abc-123") {
		t.Errorf("hidden input wrong: %q", out)
	}
}

func TestFieldDecider_NullableEmitsClearSentinel(t *testing.T) {
	out := renderFieldHTML(t, &allKindsForm{Optional: sampleNullable{v: "x"}}, "Optional")
	wantClear := mx.ClearSentinelName("Optional")
	if !strings.Contains(out, wantClear) {
		t.Errorf("expected clear sentinel %q in %q", wantClear, out)
	}
}

func TestFieldDecider_SensitiveSuppressesValue(t *testing.T) {
	out := renderFieldHTML(t, &allKindsForm{Sensitive: "secret-123"}, "Sensitive")
	if strings.Contains(out, "secret-123") {
		t.Errorf("sensitive value leaked: %q", out)
	}
}

func TestFieldDecider_RequiredAndPlaceholder(t *testing.T) {
	out := renderFieldHTML(t, &allKindsForm{}, "Required")
	if !strings.Contains(out, `required`) {
		t.Errorf("expected required attribute: %q", out)
	}
	if !strings.Contains(out, `placeholder="Type here"`) {
		t.Errorf("expected placeholder: %q", out)
	}
}

func TestFieldDecider_AriaInvalidOnErrors(t *testing.T) {
	v := reflect.ValueOf(&allKindsForm{}).Elem()
	field, _ := v.Type().FieldByName("Name")
	beh := FieldDecider("Name", field, v.FieldByName("Name"))
	out := renderToString(t, beh.Render("Name", field, v.FieldByName("Name"), []error{errFake("bad")}))
	if !strings.Contains(out, `aria-invalid="true"`) {
		t.Errorf("expected aria-invalid on error: %q", out)
	}
	if !strings.Contains(out, "bad") {
		t.Errorf("error message missing: %q", out)
	}
	if !strings.Contains(out, `data-error="Name"`) {
		t.Errorf("expected error marker: %q", out)
	}
}

type errFake string

func (e errFake) Error() string { return string(e) }

func TestFieldDecider_ParsesStringNumberBoolTime(t *testing.T) {
	type form struct {
		Name   string
		Age    int
		Active bool
		When   time.Time
	}
	f := &form{}
	values := url.Values{}
	values.Set("Name", "Alice")
	values.Set("Age", "33")
	values.Set("Active", "on")
	values.Set("When", "2026-05-25T10:30:00")
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseForm(); err != nil {
		t.Fatalf("ParseForm: %v", err)
	}

	v := reflect.ValueOf(f).Elem()
	for _, fname := range []string{"Name", "Age", "Active", "When"} {
		field, _ := v.Type().FieldByName(fname)
		beh := FieldDecider(mx.FieldPath(fname), field, v.FieldByName(fname))
		if err := beh.Parse(mx.FieldPath(fname), field, v.FieldByName(fname), r); err != nil {
			t.Fatalf("parse %s: %v", fname, err)
		}
	}
	if f.Name != "Alice" || f.Age != 33 || !f.Active {
		t.Errorf("parse: %+v", f)
	}
	if f.When.Year() != 2026 || f.When.Month() != time.May || f.When.Day() != 25 {
		t.Errorf("parsed time wrong: %v", f.When)
	}
}

func TestFieldDecider_ParsesEnumSet(t *testing.T) {
	type form struct {
		Colors map[colorEnum]struct{}
	}
	f := &form{}
	values := url.Values{}
	values["Colors"] = []string{"red", "blue"}
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseForm(); err != nil {
		t.Fatal(err)
	}
	v := reflect.ValueOf(f).Elem()
	field, _ := v.Type().FieldByName("Colors")
	beh := FieldDecider("Colors", field, v.FieldByName("Colors"))
	if err := beh.Parse("Colors", field, v.FieldByName("Colors"), r); err != nil {
		t.Fatal(err)
	}
	if len(f.Colors) != 2 {
		t.Errorf("got %v, want 2 entries", f.Colors)
	}
	if _, ok := f.Colors["red"]; !ok {
		t.Errorf("missing red: %v", f.Colors)
	}
	if _, ok := f.Colors["blue"]; !ok {
		t.Errorf("missing blue: %v", f.Colors)
	}
}

// TestReflectFormHandler_EnumSetValidatesElements posts enum sets of the
// go-enum type [Dir] through a real handler: each element is checked with
// the generated Validate, so a tampered POST cannot bind a value that is
// none of the enum's values, while valid sets reach onSubmit.
func TestReflectFormHandler_EnumSetValidatesElements(t *testing.T) {
	type draft struct {
		Dirs   []Dir
		DirSet map[Dir]struct{}
	}
	var submitted *draft
	handler := mx.ReflectFormHandler(nil, func(_ context.Context, d *draft) error {
		submitted = d
		return nil
	}, FieldDecider)
	post := func(dirs, dirSet []string) *httptest.ResponseRecorder {
		values := url.Values{
			mx.PresentSentinelName("Dirs"):   {"1"},
			mx.PresentSentinelName("DirSet"): {"1"},
			"Dirs":                           dirs,
			"DirSet":                         dirSet,
		}
		req := httptest.NewRequest(http.MethodPost, "/draft", strings.NewReader(values.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		handler(rec, req)
		return rec
	}

	rec := post([]string{"ltr", "rtl"}, []string{"auto"})
	if rec.Code != http.StatusSeeOther || submitted == nil {
		t.Fatalf("valid sets: status = %d, submitted = %v, body: %s", rec.Code, submitted, rec.Body.String())
	}
	if len(submitted.Dirs) != 2 || len(submitted.DirSet) != 1 {
		t.Errorf("valid sets not bound: %+v", submitted)
	}

	for name, sets := range map[string][2][]string{
		"slice": {{"ltr", "sideways"}, {"auto"}},
		"map":   {{"ltr"}, {"auto", "sideways"}},
	} {
		submitted = nil
		rec := post(sets[0], sets[1])
		if rec.Code == http.StatusSeeOther || submitted != nil {
			t.Errorf("%s: an invalid element reached onSubmit, status = %d", name, rec.Code)
		}
		if body := rec.Body.String(); !strings.Contains(body, "sideways") || !strings.Contains(body, "html.Dir") {
			t.Errorf("%s: the Validate error of the invalid element is not rendered: %s", name, body)
		}
	}
}

func TestFieldDecider_EnumSetRendersCheckboxes(t *testing.T) {
	type form struct {
		Colors map[colorEnum]struct{}
	}
	f := &form{Colors: map[colorEnum]struct{}{"green": {}}}
	out := renderFieldHTML(t, f, "Colors")
	if !strings.Contains(out, `type="checkbox"`) {
		t.Errorf("expected checkboxes: %q", out)
	}
	if !strings.Contains(out, `value="green"`) || !strings.Contains(out, "checked") {
		t.Errorf("green should be checked: %q", out)
	}
}

func TestFieldDecider_SkipReturnsNil(t *testing.T) {
	type form struct {
		Ignored string `form:"-"`
	}
	v := reflect.ValueOf(&form{}).Elem()
	field, _ := v.Type().FieldByName("Ignored")
	beh := FieldDecider("Ignored", field, v.FieldByName("Ignored"))
	comp := beh.Render("Ignored", field, v.FieldByName("Ignored"), nil)
	if comp != nil {
		t.Errorf("skipped field should render nil, got %T", comp)
	}
}

type partnersCtxKey struct{}

// Registry entries are process-global and duplicate registration
// panics, so tests register once in init() — never in test bodies,
// which would crash under `go test -count=2`.
func init() {
	partnersFromCtx := func(ctx context.Context) ([]mx.NamedOption, error) {
		opts, ok := ctx.Value(partnersCtxKey{}).([]mx.NamedOption)
		if !ok {
			return nil, errors.New("no partners in context")
		}
		return opts, nil
	}
	mx.RegisterNamedOptions("test-html-partners", partnersFromCtx)
	mx.RegisterNamedOptions("test-html-handler-partners", partnersFromCtx)
	mx.RegisterNamedOptions("test-html-tags", func(context.Context) ([]mx.NamedOption, error) {
		return []mx.NamedOption{
			{Name: "Tag A", Value: "a"},
			{Name: "Tag B", Value: "b"},
		}, nil
	})
}

// TestFieldDecider_SelectWithRegistryOptions builds the component once
// and renders it under two different request contexts: the option list
// must follow the context, which is the point of the form:"options=…"
// registry (per-request dropdowns like tenant-scoped partner lists).
func TestFieldDecider_SelectWithRegistryOptions(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-html-partners"`
	}
	target := &draftForm{Partner: "p2"}
	v := reflect.ValueOf(target).Elem()
	field, _ := v.Type().FieldByName("Partner")
	beh := FieldDecider(mx.FieldPath("Partner"), field, v.Field(0))
	comp := beh.Render(mx.FieldPath("Partner"), field, v.Field(0), nil)

	renderWith := func(opts []mx.NamedOption) string {
		t.Helper()
		ctx := context.WithValue(context.Background(), partnersCtxKey{}, opts)
		var b strings.Builder
		if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
			t.Fatalf("render: %v", err)
		}
		return b.String()
	}

	out := renderWith([]mx.NamedOption{{Name: "Partner One", Value: "p1"}})
	if !strings.Contains(out, `<select`) || !strings.Contains(out, "Partner One") {
		t.Errorf("expected select with option from context: %q", out)
	}
	if strings.Contains(out, `value="p1" selected`) {
		t.Errorf("p1 must not be selected: %q", out)
	}
	if !strings.Contains(out, `value="" disabled="disabled" selected="selected"`) {
		t.Errorf("current value p2 is not in the list and must show as disabled placeholder: %q", out)
	}

	out = renderWith([]mx.NamedOption{{Name: "Partner Two", Value: "p2"}})
	if !strings.Contains(out, "Partner Two") {
		t.Errorf("same component must render second context's options: %q", out)
	}
	if !strings.Contains(out, "selected") {
		t.Errorf("p2 matches the field value and must be selected: %q", out)
	}
	if strings.Contains(out, "disabled") {
		t.Errorf("no placeholder when the current value is in the list: %q", out)
	}
}

// TestFieldDecider_SelectPlaceholderForOutOfListValue encodes the
// deferred follow-up from the per-request options ship review: when
// the current value is not in the per-request option list, the select
// must show it as a disabled placeholder instead of silently
// displaying (and on submit binding) the first option. Because the
// placeholder is disabled, browsers submit nothing for the select on
// an unchanged submit: the missing value clears a non-required field
// or fails the required check, forcing an explicit re-pick.
func TestFieldDecider_SelectPlaceholderForOutOfListValue(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-html-partners"`
	}
	render := func(partner string) string {
		t.Helper()
		target := &draftForm{Partner: partner}
		v := reflect.ValueOf(target).Elem()
		field, _ := v.Type().FieldByName("Partner")
		beh := FieldDecider(mx.FieldPath("Partner"), field, v.Field(0))
		comp := beh.Render(mx.FieldPath("Partner"), field, v.Field(0), nil)
		ctx := context.WithValue(context.Background(), partnersCtxKey{},
			[]mx.NamedOption{{Name: "Partner One", Value: "p1"}})
		var b strings.Builder
		if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
			t.Fatalf("render: %v", err)
		}
		return b.String()
	}

	out := render("gone")
	if !strings.Contains(out, `<option value="" disabled="disabled" selected="selected">(current value not available)</option>`) {
		t.Errorf("out-of-list value must render as generic disabled placeholder: %q", out)
	}
	if strings.Contains(out, "gone") {
		t.Errorf("the out-of-list value must not be echoed into the markup: %q", out)
	}

	out = render("")
	if strings.Contains(out, "disabled") || strings.Contains(out, "not available") {
		t.Errorf("empty current value must not use the out-of-list placeholder: %q", out)
	}
	if !strings.Contains(out, `<option value="" selected="selected">(none)</option>`) {
		t.Errorf("empty optional select must submit an empty option, not the first real one: %q", out)
	}

	// The out-of-list value is filtered by the option provider; it must
	// not appear in the markup in any form, escaped or not.
	out = render(`</option><script>x</script>`)
	if strings.Contains(out, "script") {
		t.Errorf("out-of-list value leaked into markup: %q", out)
	}
}

// TestFieldDecider_SelectPlaceholderRespectsSensitive: sensitive
// values never round-trip into markup (formtag.go contract), so a
// sensitive-tagged select renders no placeholder for an out-of-list
// value.
func TestFieldDecider_SelectPlaceholderRespectsSensitive(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-html-partners,sensitive"`
	}
	target := &draftForm{Partner: "secret-123"}
	v := reflect.ValueOf(target).Elem()
	field, _ := v.Type().FieldByName("Partner")
	beh := FieldDecider(mx.FieldPath("Partner"), field, v.Field(0))
	comp := beh.Render(mx.FieldPath("Partner"), field, v.Field(0), nil)
	ctx := context.WithValue(context.Background(), partnersCtxKey{},
		[]mx.NamedOption{{Name: "Partner One", Value: "p1"}})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if strings.Contains(out, "secret-123") {
		t.Errorf("sensitive value leaked into placeholder: %q", out)
	}
	if !strings.Contains(out, `<option value="" disabled="disabled" selected="selected">(current value not available)</option>`) {
		t.Errorf("sensitive out-of-list value must use the generic placeholder so the first option is not submitted: %q", out)
	}
}

// TestFieldDecider_SelectEmptyOptionListRenders: an empty context list is
// a legitimate "no options" answer. The page still renders, and a stored
// value shows the generic placeholder so the first option cannot be invented.
func TestFieldDecider_SelectEmptyOptionListRenders(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-html-partners"`
	}
	target := &draftForm{Partner: "gone"}
	v := reflect.ValueOf(target).Elem()
	field, _ := v.Type().FieldByName("Partner")
	beh := FieldDecider(mx.FieldPath("Partner"), field, v.Field(0))
	comp := beh.Render(mx.FieldPath("Partner"), field, v.Field(0), nil)
	ctx := context.WithValue(context.Background(), partnersCtxKey{}, []mx.NamedOption{})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("empty list must still render: %v", err)
	}
	out := b.String()
	if strings.Contains(out, "gone") {
		t.Errorf("stored value must not be echoed: %q", out)
	}
	if !strings.Contains(out, `<option value="" disabled="disabled" selected="selected">(current value not available)</option>`) {
		t.Errorf("empty list with a stored value must show the generic placeholder: %q", out)
	}
}

// TestFieldDecider_SelectEmptyListOptionalBlankStillRenders: an optional
// field with no current value and an empty list has nothing to lose, so
// it renders an empty option rather than a 500.
func TestFieldDecider_SelectEmptyListOptionalBlankStillRenders(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-html-partners"`
	}
	target := &draftForm{}
	v := reflect.ValueOf(target).Elem()
	field, _ := v.Type().FieldByName("Partner")
	beh := FieldDecider(mx.FieldPath("Partner"), field, v.Field(0))
	comp := beh.Render(mx.FieldPath("Partner"), field, v.Field(0), nil)
	ctx := context.WithValue(context.Background(), partnersCtxKey{}, []mx.NamedOption{})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(b.String(), `<option value="" selected="selected">(none)</option>`) {
		t.Errorf("optional blank field should render an empty option: %q", b.String())
	}
}

func TestFieldDecider_RequiredSelectEmptyValueChoosesNothing(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-html-partners,required"`
	}
	target := &draftForm{}
	v := reflect.ValueOf(target).Elem()
	field, _ := v.Type().FieldByName("Partner")
	beh := FieldDecider(mx.FieldPath("Partner"), field, v.Field(0))
	comp := beh.Render(mx.FieldPath("Partner"), field, v.Field(0), nil)
	ctx := context.WithValue(context.Background(), partnersCtxKey{},
		[]mx.NamedOption{{Name: "Partner One", Value: "p1"}})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, `required`) {
		t.Errorf("required attribute missing: %q", out)
	}
	if !strings.Contains(out, `<option value="" disabled="disabled" selected="selected">Choose…</option>`) {
		t.Errorf("required empty select must lead with a disabled empty option: %q", out)
	}
	if strings.Contains(out, `value="p1" selected`) {
		t.Errorf("first real option must not be selected: %q", out)
	}
}

func TestFieldDecider_RequiredSelectEmptyListRendersChoose(t *testing.T) {
	type draftForm struct {
		Partner string `form:"options=test-html-partners,required"`
	}
	target := &draftForm{}
	v := reflect.ValueOf(target).Elem()
	field, _ := v.Type().FieldByName("Partner")
	beh := FieldDecider(mx.FieldPath("Partner"), field, v.Field(0))
	comp := beh.Render(mx.FieldPath("Partner"), field, v.Field(0), nil)
	ctx := context.WithValue(context.Background(), partnersCtxKey{}, []mx.NamedOption{})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("required empty list must still render: %v", err)
	}
	if !strings.Contains(b.String(), `<option value="" disabled="disabled" selected="selected">Choose…</option>`) {
		t.Errorf("required empty list must lead with Choose…: %q", b.String())
	}
}

// TestFieldDecider_SelectStaticListNoPlaceholder pins the scope of the
// out-of-list placeholder: static enum lists stay lenient. An out-of-list
// value selects nothing. An optional select still offers "(none)", which
// is what the browser submits when nothing is marked selected.
func TestFieldDecider_SelectStaticListNoPlaceholder(t *testing.T) {
	out := renderFieldHTML(t, &allKindsForm{Color: "magenta"}, "Color")
	if strings.Contains(out, "disabled") {
		t.Errorf("static enum list must not render a disabled placeholder: %q", out)
	}
	if strings.Contains(out, "selected") {
		t.Errorf("out-of-list value on a static list must not select anything: %q", out)
	}
	noneAt := strings.Index(out, `value=""`)
	redAt := strings.Index(out, `value="red"`)
	if noneAt < 0 || redAt < 0 || noneAt > redAt {
		t.Errorf("empty option must come before the first enum value: %q", out)
	}
}

func TestFieldDecider_RequiredStaticSelectEmptyValueChoosesNothing(t *testing.T) {
	type form struct {
		Color colorEnum `form:"required"`
	}
	out := renderFieldHTML(t, &form{}, "Color")
	if !strings.Contains(out, `<option value="" disabled="disabled" selected="selected">Choose…</option>`) {
		t.Errorf("required empty enum must lead with Choose…: %q", out)
	}
	if strings.Contains(out, `value="red" selected`) {
		t.Errorf("first enum value must not be selected: %q", out)
	}
}

func TestFieldDecider_SelectWithUnregisteredOptionsName(t *testing.T) {
	type badForm struct {
		X string `form:"options=test-html-not-registered"`
	}
	target := &badForm{}
	v := reflect.ValueOf(target).Elem()
	field, _ := v.Type().FieldByName("X")
	beh := FieldDecider(mx.FieldPath("X"), field, v.Field(0))
	comp := beh.Render(mx.FieldPath("X"), field, v.Field(0), nil)

	var b strings.Builder
	err := comp.Render(context.Background(), mx.NewCheckedWriter(&b))
	if err == nil {
		t.Fatal("expected render error for unregistered options name")
	}
	if !strings.Contains(err.Error(), "test-html-not-registered") {
		t.Errorf("error should name the missing entry: %v", err)
	}
}

func TestFieldDecider_EnumSetWithRegistryOptions(t *testing.T) {
	type tagForm struct {
		Tags []string `form:"options=test-html-tags"`
	}
	target := &tagForm{Tags: []string{"b"}}
	v := reflect.ValueOf(target).Elem()
	field, _ := v.Type().FieldByName("Tags")
	beh := FieldDecider(mx.FieldPath("Tags"), field, v.Field(0))
	comp := beh.Render(mx.FieldPath("Tags"), field, v.Field(0), nil)

	var b strings.Builder
	if err := comp.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if strings.Count(out, `type="checkbox"`) != 2 {
		t.Errorf("expected one checkbox per registry option: %q", out)
	}
	if !strings.Contains(out, "Tag A") || !strings.Contains(out, "Tag B") {
		t.Errorf("expected option labels: %q", out)
	}
	if !strings.Contains(out, "checked") {
		t.Errorf("expected current member b to be checked: %q", out)
	}
	if strings.Contains(out, "not available") {
		t.Errorf("listed members must not show the hidden-member note: %q", out)
	}
}

// TestReflectFormHandler_RegistryOptionsGetRequestContext proves the
// whole point of the registry end to end: the HTTP request context —
// not anything on the field type — supplies the option list when a
// real handler renders the form.
func TestReflectFormHandler_RegistryOptionsGetRequestContext(t *testing.T) {
	type draft struct {
		Partner string `form:"options=test-html-handler-partners"`
	}
	handler := mx.ReflectFormHandler(nil, func(context.Context, *draft) error { return nil }, FieldDecider)

	req := httptest.NewRequest(http.MethodGet, "/draft", nil)
	req = req.WithContext(context.WithValue(req.Context(), partnersCtxKey{},
		[]mx.NamedOption{{Name: "Partner One", Value: "p1"}}))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, body: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<select") || !strings.Contains(body, "Partner One") {
		t.Errorf("expected select with request-context option in handler output: %q", body)
	}
}

// fakeID is a stand-in for uuid.UUID-shaped foreign key types: a [16]byte
// array with TextMarshaler/TextUnmarshaler — the motivating use case
// for form:"options=…" on fields whose type cannot implement a
// provider interface.
type fakeID [16]byte

func (id fakeID) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%x", id[:4])), nil
}

func (id *fakeID) UnmarshalText(text []byte) error {
	_, err := fmt.Sscanf(string(text), "%8x", id)
	return err
}

func TestFieldDecider_SelectWithRegistryOptions_IDType(t *testing.T) {
	type draft struct {
		Partner fakeID `form:"options=test-html-partners"`
	}
	target := &draft{Partner: fakeID{0xde, 0xad, 0xbe, 0xef}}
	v := reflect.ValueOf(target).Elem()
	field, _ := v.Type().FieldByName("Partner")

	kind, _ := mx.DetectField("Partner", field, v.Field(0))
	if kind != mx.FieldKindEnum {
		t.Fatalf("options-tagged ID type: kind = %s, want enum", kind)
	}

	beh := FieldDecider(mx.FieldPath("Partner"), field, v.Field(0))
	comp := beh.Render(mx.FieldPath("Partner"), field, v.Field(0), nil)

	ctx := context.WithValue(context.Background(), partnersCtxKey{}, []mx.NamedOption{
		{Name: "Other Partner", Value: "11223344"},
		{Name: "Dead Beef Inc", Value: "deadbeef"},
	})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if !strings.Contains(out, `value="deadbeef" selected`) {
		t.Errorf("marshaled ID must match the option Value for selection: %q", out)
	}
	if strings.Contains(out, `value="11223344" selected`) {
		t.Errorf("non-matching option must not be selected: %q", out)
	}
}

// TestFieldDecider_ParseRejectsOutOfListOption encodes D1 of the ship
// review: the per-request option list is a trust boundary, so POST
// must not bind values outside it (e.g. another tenant's partner ID).
func TestFieldDecider_ParseRejectsOutOfListOption(t *testing.T) {
	type draft struct {
		Partner string `form:"options=test-html-partners"`
	}
	f := &draft{Partner: "p1"}
	parse := func(submitted string) error {
		values := url.Values{}
		values.Set("Partner", submitted)
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r = r.WithContext(context.WithValue(r.Context(), partnersCtxKey{},
			[]mx.NamedOption{{Name: "Partner One", Value: "p1"}}))
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		v := reflect.ValueOf(f).Elem()
		field, _ := v.Type().FieldByName("Partner")
		beh := FieldDecider("Partner", field, v.Field(0))
		return beh.Parse("Partner", field, v.Field(0), r)
	}

	if err := parse("p1"); err != nil {
		t.Fatalf("in-list value must bind: %v", err)
	}
	if f.Partner != "p1" {
		t.Errorf("in-list value not bound: %q", f.Partner)
	}

	err := parse("p2") // not in this request's list
	if err == nil {
		t.Fatal("expected out-of-list value to be rejected")
	}
	if !strings.Contains(err.Error(), "allowed options") {
		t.Errorf("rejection should be a plain validation error: %v", err)
	}
	if f.Partner != "p1" {
		t.Errorf("rejected value must not be bound: %q", f.Partner)
	}

	// Empty submission means "no selection" — required-ness is a
	// separate concern, membership must not reject it.
	if err := parse(""); err != nil {
		t.Fatalf("empty value must pass membership validation: %v", err)
	}
}

func TestFieldDecider_ParseRejectsOutOfListEnumSet(t *testing.T) {
	type tagForm struct {
		Tags []string `form:"options=test-html-tags"` // registry offers a, b
	}
	f := &tagForm{}
	parse := func(submitted ...string) error {
		values := url.Values{"Tags": submitted}
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		v := reflect.ValueOf(f).Elem()
		field, _ := v.Type().FieldByName("Tags")
		beh := FieldDecider("Tags", field, v.Field(0))
		return beh.Parse("Tags", field, v.Field(0), r)
	}

	if err := parse("a", "b"); err != nil {
		t.Fatalf("in-list members must bind: %v", err)
	}
	if len(f.Tags) != 2 {
		t.Errorf("expected both members bound: %v", f.Tags)
	}

	err := parse("a", "admin") // "admin" was never offered
	if err == nil {
		t.Fatal("expected out-of-list member to be rejected")
	}
	if len(f.Tags) != 2 || f.Tags[0] != "a" || f.Tags[1] != "b" {
		t.Errorf("rejected submission must not modify the set: %v", f.Tags)
	}
}

// TestFieldDecider_EnumSetDropsMembersMissingFromOptionList: checkboxes
// only render the current list. A stored member that is absent from it is
// announced and not echoed, and a save does not keep it. The list is the
// trust boundary, so an uncheck of a listed member sticks even if that
// option disappears before POST. A value that was never stored and is not
// in the list is still rejected, and a rejected submit leaves the set as it was.
func TestFieldDecider_EnumSetDropsMembersMissingFromOptionList(t *testing.T) {
	type tagForm struct {
		Tags []string `form:"options=test-html-tags"` // registry offers a, b
	}

	target := &tagForm{Tags: []string{"kept", "b"}}
	v := reflect.ValueOf(target).Elem()
	field, _ := v.Type().FieldByName("Tags")
	beh := FieldDecider(mx.FieldPath("Tags"), field, v.Field(0))
	comp := beh.Render(mx.FieldPath("Tags"), field, v.Field(0), nil)
	var b strings.Builder
	if err := comp.Render(context.Background(), mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := b.String()
	if strings.Contains(out, "kept") {
		t.Errorf("unlisted member must not be echoed: %q", out)
	}
	if !strings.Contains(out, "(some current values are not available)") {
		t.Errorf("missing members must be announced without naming them: %q", out)
	}
	if strings.Count(out, `type="checkbox"`) != 2 {
		t.Errorf("only the listed options are checkboxes: %q", out)
	}

	parse := func(f *tagForm, submitted ...string) error {
		t.Helper()
		values := url.Values{"Tags": submitted}
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		pv := reflect.ValueOf(f).Elem()
		pf, _ := pv.Type().FieldByName("Tags")
		pb := FieldDecider("Tags", pf, pv.Field(0))
		return pb.Parse("Tags", pf, pv.Field(0), r)
	}

	f := &tagForm{Tags: []string{"kept", "b"}}
	if err := parse(f, "a"); err != nil {
		t.Fatalf("listed submission must bind: %v", err)
	}
	if len(f.Tags) != 1 || f.Tags[0] != "a" {
		t.Fatalf("got %v, want [a]", f.Tags)
	}

	f = &tagForm{Tags: []string{"kept", "a"}}
	if err := parse(f); err != nil {
		t.Fatalf("empty submission must bind: %v", err)
	}
	if len(f.Tags) != 0 {
		t.Fatalf("got %v, want no members", f.Tags)
	}

	f = &tagForm{Tags: []string{"kept"}}
	if err := parse(f, "admin"); err == nil {
		t.Fatal("expected a newly submitted unlisted member to be rejected")
	}
	if len(f.Tags) != 1 || f.Tags[0] != "kept" {
		t.Fatalf("rejected submission must not modify the set: %v", f.Tags)
	}
}

func TestFieldDecider_EnumSetEmptyListWithMembersRenders(t *testing.T) {
	// test-html-partners reads the option list from the request context,
	// so the test can hand it an empty slice.
	type ctxForm struct {
		Tags []string `form:"options=test-html-partners"`
	}
	target := &ctxForm{Tags: []string{"kept"}}
	v := reflect.ValueOf(target).Elem()
	field, _ := v.Type().FieldByName("Tags")
	beh := FieldDecider(mx.FieldPath("Tags"), field, v.Field(0))
	comp := beh.Render(mx.FieldPath("Tags"), field, v.Field(0), nil)
	ctx := context.WithValue(context.Background(), partnersCtxKey{}, []mx.NamedOption{})
	var b strings.Builder
	if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("empty list must still render: %v", err)
	}
	out := b.String()
	if strings.Contains(out, "kept") {
		t.Errorf("unlisted member must not be echoed: %q", out)
	}
	if strings.Contains(out, `type="checkbox"`) {
		t.Errorf("empty list must not render a checkbox: %q", out)
	}
	if !strings.Contains(out, "(some current values are not available)") {
		t.Errorf("missing members must be announced: %q", out)
	}
}

func TestFieldDecider_StaticEnumSetReplacesFromSubmission(t *testing.T) {
	type form struct {
		Colors []colorEnum
	}
	f := &form{Colors: []colorEnum{"magenta", "red"}}
	values := url.Values{"Colors": {"blue"}}
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseForm(); err != nil {
		t.Fatal(err)
	}
	v := reflect.ValueOf(f).Elem()
	field, _ := v.Type().FieldByName("Colors")
	beh := FieldDecider("Colors", field, v.Field(0))
	if err := beh.Parse("Colors", field, v.Field(0), r); err != nil {
		t.Fatal(err)
	}
	if len(f.Colors) != 1 || f.Colors[0] != "blue" {
		t.Fatalf("static list replaces the set from the submission, got %v", f.Colors)
	}
}

// marshaledID is a registry key whose text form is not fmt's default.
// Sprint of the array is "[222 173 190 239]"; MarshalText is "deadbeef".
type marshaledID [4]byte

func (id marshaledID) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(id[:])), nil
}

func (id *marshaledID) UnmarshalText(text []byte) error {
	raw, err := hex.DecodeString(string(text))
	if err != nil || len(raw) != len(id) {
		return fmt.Errorf("bad id %q", text)
	}
	copy(id[:], raw)
	return nil
}

func TestFieldDecider_ParseEmptyClearsMarshaledID(t *testing.T) {
	type draft struct {
		Partner marshaledID `form:"options=test-html-partners"`
	}
	f := &draft{Partner: marshaledID{0xde, 0xad, 0xbe, 0xef}}
	values := url.Values{}
	values.Set("Partner", "")
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = r.WithContext(context.WithValue(r.Context(), partnersCtxKey{}, []mx.NamedOption{
		{Name: "Dead Beef", Value: "deadbeef"},
	}))
	if err := r.ParseForm(); err != nil {
		t.Fatal(err)
	}
	v := reflect.ValueOf(f).Elem()
	field, _ := v.Type().FieldByName("Partner")
	beh := FieldDecider("Partner", field, v.Field(0))
	if err := beh.Parse("Partner", field, v.Field(0), r); err != nil {
		t.Fatalf("empty optional id must clear, not fail UnmarshalText: %v", err)
	}
	if f.Partner != (marshaledID{}) {
		t.Fatalf("got %v, want zero", f.Partner)
	}
}

type colorCode int

const (
	colorRed  colorCode = 1
	colorBlue colorCode = 2
)

func (c colorCode) String() string {
	switch c {
	case colorRed:
		return "red"
	case colorBlue:
		return "blue"
	default:
		return fmt.Sprintf("%d", int(c))
	}
}

func (colorCode) Enums() []colorCode { return []colorCode{colorRed, colorBlue} }

func (c *colorCode) UnmarshalText(text []byte) error {
	switch string(text) {
	case "red":
		*c = colorRed
	case "blue":
		*c = colorBlue
	default:
		return fmt.Errorf("bad color %q", text)
	}
	return nil
}

func TestFieldDecider_EnumSetStringerMatchesEnums(t *testing.T) {
	type form struct {
		Colors []colorCode
	}
	f := &form{Colors: []colorCode{colorRed}}
	out := renderFieldHTML(t, f, "Colors")
	if strings.Contains(out, `value="1"`) {
		t.Fatalf("numeric form must not be the option value: %s", out)
	}
	red := strings.Index(out, `value="red"`)
	if red < 0 {
		t.Fatalf("Stringer form must be the option value: %s", out)
	}
	start := strings.LastIndex(out[:red], "<input")
	end := strings.Index(out[red:], ">")
	tag := out[start : red+end]
	if !strings.Contains(tag, "checked") {
		t.Fatalf("stored red must be checked: %s", tag)
	}

	values := url.Values{"Colors": {"red"}}
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseForm(); err != nil {
		t.Fatal(err)
	}
	v := reflect.ValueOf(f).Elem()
	field, _ := v.Type().FieldByName("Colors")
	beh := FieldDecider("Colors", field, v.Field(0))
	if err := beh.Parse("Colors", field, v.Field(0), r); err != nil {
		t.Fatal(err)
	}
	if len(f.Colors) != 1 || f.Colors[0] != colorRed {
		t.Fatalf("got %v, want [red]", f.Colors)
	}
}

type prefixedID string

func (id prefixedID) MarshalText() ([]byte, error) {
	return []byte("id:" + string(id)), nil
}

func (id *prefixedID) UnmarshalText(text []byte) error {
	s := string(text)
	rest, ok := strings.CutPrefix(s, "id:")
	if !ok || rest == "" {
		return fmt.Errorf("bad id %q", text)
	}
	*id = prefixedID(rest)
	return nil
}

func TestFieldDecider_EnumSetNamedStringRoundTrip(t *testing.T) {
	opts := []mx.NamedOption{{Name: "Listed", Value: "id:listed"}}
	ctx := context.WithValue(context.Background(), partnersCtxKey{}, opts)

	check := func(t *testing.T, target any, parse func(submitted ...string) error, got func() []prefixedID) {
		t.Helper()
		v := reflect.ValueOf(target).Elem()
		field := v.Type().Field(0)
		beh := FieldDecider(mx.FieldPath(field.Name), field, v.Field(0))
		comp := beh.Render(mx.FieldPath(field.Name), field, v.Field(0), nil)
		var b strings.Builder
		if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(b.String(), `value="id:listed"`) {
			t.Fatalf("text form must be the checkbox value: %s", b.String())
		}
		if err := parse("id:listed"); err != nil {
			t.Fatal(err)
		}
		ids := got()
		if len(ids) != 1 || ids[0] != "listed" {
			t.Fatalf("got %q, want [listed]", ids)
		}
	}

	t.Run("slice", func(t *testing.T) {
		type form struct {
			IDs []prefixedID `form:"options=test-html-partners"`
		}
		f := &form{IDs: []prefixedID{"listed"}}
		check(t, f, func(submitted ...string) error {
			values := url.Values{"IDs": submitted}
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r = r.WithContext(ctx)
			if err := r.ParseForm(); err != nil {
				return err
			}
			v := reflect.ValueOf(f).Elem()
			field, _ := v.Type().FieldByName("IDs")
			beh := FieldDecider("IDs", field, v.Field(0))
			return beh.Parse("IDs", field, v.Field(0), r)
		}, func() []prefixedID { return f.IDs })
	})

	t.Run("map", func(t *testing.T) {
		type form struct {
			IDs map[prefixedID]struct{} `form:"options=test-html-partners"`
		}
		f := &form{IDs: map[prefixedID]struct{}{"listed": {}}}
		check(t, f, func(submitted ...string) error {
			values := url.Values{"IDs": submitted}
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r = r.WithContext(ctx)
			if err := r.ParseForm(); err != nil {
				return err
			}
			v := reflect.ValueOf(f).Elem()
			field, _ := v.Type().FieldByName("IDs")
			beh := FieldDecider("IDs", field, v.Field(0))
			return beh.Parse("IDs", field, v.Field(0), r)
		}, func() []prefixedID {
			out := make([]prefixedID, 0, len(f.IDs))
			for id := range f.IDs {
				out = append(out, id)
			}
			return out
		})
	})
}

func TestReflectFormHandler_RequiredMarshaledIDRejectsBlank(t *testing.T) {
	type draft struct {
		Partner marshaledID `form:"options=test-html-partners,required"`
	}
	handler := mx.ReflectFormHandler(nil, func(context.Context, *draft) error {
		t.Fatal("blank required id must not save")
		return nil
	}, FieldDecider)
	form := url.Values{}
	form.Set(mx.PresentSentinelName("Partner"), "1")
	form.Set("Partner", "")
	req := httptest.NewRequest(http.MethodPost, "/draft", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), partnersCtxKey{}, []mx.NamedOption{
		{Name: "Dead Beef", Value: "deadbeef"},
	}))
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200, body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "required") {
		t.Fatalf("required error missing: %s", rec.Body.String())
	}
}

// Value: protects=a text-marshaled enum-set member matches its option and stays removable, and a hidden member is dropped on save; fails_when=members are formatted with fmt.Sprint so the stored ID is neither checked nor removable, or a hidden member is written back; why_new=the ID select test is a single value and enum-set tests use plain strings; seam=none; no_claim=shadcn checkbox state is covered by TestShadcn_EnumSetChecksMarshaledMember
func TestFieldDecider_EnumSetMarshaledMemberRoundTrip(t *testing.T) {
	listed := marshaledID{0xde, 0xad, 0xbe, 0xef} // deadbeef
	hidden := marshaledID{0x01, 0x02, 0x03, 0x04} // 01020304
	opts := []mx.NamedOption{
		{Name: "Dead Beef", Value: "deadbeef"},
		{Name: "Other", Value: "11111111"},
	}
	ctx := context.WithValue(context.Background(), partnersCtxKey{}, opts)

	render := func(t *testing.T, target any) string {
		t.Helper()
		v := reflect.ValueOf(target).Elem()
		field := v.Type().Field(0)
		comp := FieldDecider("IDs", field, v.Field(0)).Render("IDs", field, v.Field(0), nil)
		var b strings.Builder
		if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
			t.Fatalf("render: %v", err)
		}
		return b.String()
	}
	checked := func(out, value string) bool {
		i := strings.Index(out, `value="`+value+`"`)
		if i < 0 {
			return false
		}
		start := strings.LastIndex(out[:i], "<")
		end := strings.Index(out[i:], ">")
		if start < 0 || end < 0 {
			return false
		}
		return strings.Contains(out[start:i+end], `checked="checked"`)
	}
	parse := func(t *testing.T, target any, submitted ...string) error {
		t.Helper()
		values := url.Values{"IDs": submitted}
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r = r.WithContext(ctx)
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		v := reflect.ValueOf(target).Elem()
		field := v.Type().Field(0)
		return FieldDecider("IDs", field, v.Field(0)).Parse("IDs", field, v.Field(0), r)
	}

	t.Run("slice", func(t *testing.T) {
		type form struct {
			IDs []marshaledID `form:"options=test-html-partners"`
		}
		out := render(t, &form{IDs: []marshaledID{listed, hidden}})
		if !checked(out, "deadbeef") {
			t.Fatalf("marshaled member must be checked: %s", out)
		}
		if strings.Contains(out, "01020304") || strings.Contains(out, "[222 173 190 239]") {
			t.Fatalf("hidden or sprint form leaked: %s", out)
		}
		if !strings.Contains(out, "(some current values are not available)") {
			t.Fatalf("hidden member must be announced: %s", out)
		}

		f := &form{IDs: []marshaledID{listed, hidden}}
		if err := parse(t, f, "deadbeef"); err != nil {
			t.Fatal(err)
		}
		if len(f.IDs) != 1 || f.IDs[0] != listed {
			t.Fatalf("got %x, want only the listed id", f.IDs)
		}

		f = &form{IDs: []marshaledID{listed, hidden}}
		if err := parse(t, f); err != nil {
			t.Fatal(err)
		}
		if len(f.IDs) != 0 {
			t.Fatalf("unchecked submission must drop every member, got %x", f.IDs)
		}

		f = &form{IDs: []marshaledID{hidden}}
		if err := parse(t, f, "99999999"); err == nil {
			t.Fatal("expected a new unlisted member to be rejected")
		}
		if len(f.IDs) != 1 || f.IDs[0] != hidden {
			t.Fatalf("rejected submit must not change the set: %x", f.IDs)
		}
	})

	t.Run("map", func(t *testing.T) {
		type form struct {
			IDs map[marshaledID]struct{} `form:"options=test-html-partners"`
		}
		out := render(t, &form{IDs: map[marshaledID]struct{}{listed: {}, hidden: {}}})
		if !checked(out, "deadbeef") {
			t.Fatalf("marshaled map key must be checked: %s", out)
		}
		if strings.Contains(out, "01020304") {
			t.Fatalf("hidden member leaked: %s", out)
		}

		f := &form{IDs: map[marshaledID]struct{}{listed: {}, hidden: {}}}
		if err := parse(t, f); err != nil {
			t.Fatal(err)
		}
		if len(f.IDs) != 0 {
			t.Fatalf("got %d members, want none", len(f.IDs))
		}
		if _, ok := f.IDs[hidden]; ok {
			t.Fatalf("hidden key must not be kept: %v", f.IDs)
		}
	})
}

// Value: protects=enum-set parse returns the options-provider error and leaves the stored set unchanged; fails_when=a CollectOptions failure is dropped and the submission is applied or rejected as a membership miss; why_new=membership tests use a healthy registry and render tests only cover provider failure on GET; seam=none; no_claim=an empty list with no provider error is the render-error tests
func TestFieldDecider_EnumSetParseReturnsProviderError(t *testing.T) {
	type tagForm struct {
		Tags []string `form:"options=test-html-partners"`
	}
	f := &tagForm{Tags: []string{"kept"}}
	values := url.Values{"Tags": {"p1"}}
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseForm(); err != nil {
		t.Fatal(err)
	}
	v := reflect.ValueOf(f).Elem()
	field, _ := v.Type().FieldByName("Tags")
	err := FieldDecider("Tags", field, v.Field(0)).Parse("Tags", field, v.Field(0), r)
	if err == nil || !strings.Contains(err.Error(), "no partners in context") {
		t.Fatalf("provider error = %v", err)
	}
	if len(f.Tags) != 1 || f.Tags[0] != "kept" {
		t.Fatalf("failed parse changed the set: %v", f.Tags)
	}
}

func TestReflectFormHandler_EnumSetDropsUnlistedMembers(t *testing.T) {
	type draft struct {
		Tags []string `form:"options=test-html-tags"`
	}
	var saved *draft
	handler := mx.ReflectFormHandler(
		func(context.Context) (*draft, error) {
			return &draft{Tags: []string{"kept", "b"}}, nil
		},
		func(_ context.Context, d *draft) error {
			cp := *d
			cp.Tags = append([]string{}, d.Tags...)
			saved = &cp
			return nil
		},
		FieldDecider,
	)

	get := httptest.NewRequest(http.MethodGet, "/tags", nil)
	rec := httptest.NewRecorder()
	handler(rec, get)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, body: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "kept") {
		t.Errorf("GET must not echo the unlisted member: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "(some current values are not available)") {
		t.Errorf("GET must say that some values are hidden: %s", rec.Body.String())
	}

	form := url.Values{}
	form.Set(mx.PresentSentinelName("Tags"), "1")
	form.Add("Tags", "a")
	post := httptest.NewRequest(http.MethodPost, "/tags", strings.NewReader(form.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	handler(rec, post)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST status = %d, body: %s", rec.Code, rec.Body.String())
	}
	if saved == nil || len(saved.Tags) != 1 || saved.Tags[0] != "a" {
		t.Fatalf("saved Tags = %v, want [a]", saved)
	}
}

type checkedEnum string

func (checkedEnum) EnumStrings() []string { return []string{"a", "b"} }

func (e checkedEnum) Validate() error {
	if e != "a" && e != "b" {
		return errors.New("invalid enum")
	}
	return nil
}

func TestReflectFormHandler_EnumSetDropsInvalidUnlistedMember(t *testing.T) {
	type draft struct {
		Tags []checkedEnum `form:"options=test-html-tags"`
	}
	var saved []checkedEnum
	handler := mx.ReflectFormHandler(
		func(context.Context) (*draft, error) {
			return &draft{Tags: []checkedEnum{"nope", "b"}}, nil
		},
		func(_ context.Context, d *draft) error {
			saved = append([]checkedEnum{}, d.Tags...)
			return nil
		},
		FieldDecider,
	)
	form := url.Values{}
	form.Set(mx.PresentSentinelName("Tags"), "1")
	form.Add("Tags", "a")
	post := httptest.NewRequest(http.MethodPost, "/tags", strings.NewReader(form.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler(rec, post)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("stale member must not block the save, status %d body %s", rec.Code, rec.Body.String())
	}
	if len(saved) != 1 || saved[0] != "a" {
		t.Fatalf("saved %v, want [a]", saved)
	}
}

func TestReflectFormHandler_EmptyOptionListStillRenders(t *testing.T) {
	type draft struct {
		Partner string `form:"options=test-html-handler-partners"`
	}
	handler := mx.ReflectFormHandler(
		func(context.Context) (*draft, error) {
			return &draft{Partner: "gone"}, nil
		},
		func(context.Context, *draft) error { return nil },
		FieldDecider,
	)
	req := httptest.NewRequest(http.MethodGet, "/draft", nil)
	req = req.WithContext(context.WithValue(req.Context(), partnersCtxKey{}, []mx.NamedOption{}))
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<form") {
		t.Fatalf("empty option list must still render the form: %q", body)
	}
	if strings.Contains(body, "gone") {
		t.Errorf("stored value must not be echoed: %q", body)
	}
	if !strings.Contains(body, "(current value not available)") {
		t.Errorf("stored value must show the generic placeholder: %q", body)
	}
}

// TestReflectFormHandler_ProviderErrorCleanInternalServerError encodes
// D2 of the ship review: a render-time options failure (typo'd registry
// name, provider DB error) must be a clean 500, never a truncated 200 —
// the handler buffers the render before writing.
func TestReflectFormHandler_ProviderErrorCleanInternalServerError(t *testing.T) {
	type draft struct {
		X string `form:"options=test-html-never-registered"`
	}
	handler := mx.ReflectFormHandler(nil, func(context.Context, *draft) error { return nil }, FieldDecider)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 — a streamed render would have sent a truncated 200", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<form") {
		t.Errorf("partial form HTML must not leak on render error: %q", rec.Body.String())
	}
}

// An empty submission stores the zero value of a non-nillable field
// (setScalar, setTime), so the required check has to read the zero
// uuid.UUID and the zero time.Time back as "no value". A length check alone
// counts a [16]byte array as always filled and lets a blank ID or date
// through as uuid.Nil() or 0001-01-01.
func TestReflectFormHandler_RequiredZeroValueRejectsBlank(t *testing.T) {
	t.Run("hidden uuid", func(t *testing.T) {
		type draft struct {
			Ref uuid.UUID `form:"hidden,required"`
		}
		postBlankRequired(t, "Ref", &draft{Ref: uuid.New()})
	})
	t.Run("text uuid", func(t *testing.T) {
		type draft struct {
			Code uuid.UUID `form:"required"`
		}
		postBlankRequired(t, "Code", &draft{Code: uuid.New()})
	})
	t.Run("datetime", func(t *testing.T) {
		type draft struct {
			Due time.Time `form:"required"`
		}
		postBlankRequired(t, "Due", &draft{Due: time.Date(2026, 10, 18, 9, 30, 0, 0, time.UTC)})
	})
}

// postBlankRequired posts an empty value for the required field name of a
// loaded draft and fails unless the handler re-renders the form instead of
// saving.
func postBlankRequired[T any](t *testing.T, name string, loaded *T) {
	t.Helper()
	saved := false
	handler := mx.ReflectFormHandler(
		func(context.Context) (*T, error) { return loaded, nil },
		func(context.Context, *T) error { saved = true; return nil },
		FieldDecider,
	)
	form := url.Values{}
	form.Set(mx.PresentSentinelName(mx.FieldPath(name)), "1")
	form.Set(name, "")
	req := httptest.NewRequest(http.MethodPost, "/draft", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler(rec, req)
	if saved || rec.Code != http.StatusOK {
		t.Errorf("blank required %s: status %d, saved %t; want the form re-rendered with 200", name, rec.Code, saved)
	}
}

// The zero uuid.UUID and the zero time.Time are what an empty submission
// stores, so they render as no value: a select leads with its empty-value
// option instead of the out-of-list placeholder, and inputs stay empty
// instead of showing 00000000-0000-0000-0000-000000000000 or 0001-01-01.
func TestFieldDecider_ZeroValueRendersAsNoValue(t *testing.T) {
	partner := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	ctx := context.WithValue(context.Background(), partnersCtxKey{},
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
			Partner uuid.UUID `form:"options=test-html-partners,required"`
		}{})
		if !strings.Contains(out, `<option value="" disabled="disabled" selected="selected">Choose…</option>`) {
			t.Errorf("a required select without a value must lead with Choose…: %q", out)
		}
		if strings.Contains(out, "not available") {
			t.Errorf("the zero id is no value, not an unlisted one: %q", out)
		}
	})
	t.Run("optional select", func(t *testing.T) {
		out := render(t, &struct {
			Partner uuid.UUID `form:"options=test-html-partners"`
		}{})
		if !strings.Contains(out, `<option value="" selected="selected">(none)</option>`) {
			t.Errorf("an optional select without a value must select (none): %q", out)
		}
	})
	t.Run("hidden uuid", func(t *testing.T) {
		out := render(t, &struct {
			Ref uuid.UUID `form:"hidden"`
		}{})
		if strings.Contains(out, "00000000-0000") {
			t.Errorf("the zero id must render as an empty value: %q", out)
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

// time.Time fields render and bind exactly as they did before
// form:"widget=date" learned string-kind fields.
func TestFieldDecider_TimeWidgetsUnchanged(t *testing.T) {
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
		{"Date", `<label for="Date">Date:</label><input type="date" name="Date" id="Date" value="2026-07-06" min="2026-01-01" max="2026-12-31" required="required"/>`},
		{"DateTime", `<label for="DateTime">DateTime:</label><input type="datetime-local" name="DateTime" id="DateTime" value="2026-07-06T09:05:07"/>`},
		{"Time", `<label for="Time">Time:</label><input type="time" name="Time" id="Time" value="09:05:07" readonly="readonly"/>`},
		{"Default", `<label for="Default">Default:</label><input type="datetime-local" name="Default" id="Default" value="2026-07-06T09:05:07"/>`},
		{"Nil", `<label for="Nil">Nil:</label><input type="date" name="Nil" id="Nil"/>`},
		{"Zero", `<label for="Zero">Zero:</label><input type="date" name="Zero" id="Zero"/>`},
	}
	for _, r := range renders {
		if got := renderFieldHTML(t, f, r.field); got != r.want {
			t.Errorf("%s renders\n%s\nwant\n%s", r.field, got, r.want)
		}
	}

	if err := parseFieldValue(t, f, "Nil", "2026-08-05"); err != nil {
		t.Fatalf("bind a date into *time.Time: %v", err)
	}
	if want := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC); f.Nil == nil || !f.Nil.Equal(want) {
		t.Errorf("Nil = %v, want %v", f.Nil, want)
	}
	err := parseFieldValue(t, f, "Zero", "06.07.2026")
	if want := `parsing time "06.07.2026" as "2006-01-02": cannot parse "06.07.2026" as "2006"`; err == nil || err.Error() != want {
		t.Errorf("bind 06.07.2026 into time.Time: error %v, want %s", err, want)
	}
	// The time widget reads a time of day without seconds too.
	if err := parseFieldValue(t, f, "Time", "09:05"); err != nil {
		t.Fatalf("bind 09:05 into a time widget: %v", err)
	}
	if f.Time.Hour() != 9 || f.Time.Minute() != 5 || f.Time.Second() != 0 {
		t.Errorf("Time = %v, want 09:05:00", f.Time)
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
	Strict    *strictDate     `form:"widget=date"`
	Localized string          `form:"widget=date,placeholder=TT.MM.JJJJ"`
}

// strictDate rejects dates before 2000 in UnmarshalText, so a test can see
// what binding does with a valid date the type itself refuses.
type strictDate string

func (d *strictDate) UnmarshalText(text []byte) error {
	if string(text) < "2000" {
		return errors.New("dates before 2000 are not accepted")
	}
	*d = strictDate(text)
	return nil
}

// A form:"widget=date" field that holds its date as text prefills the date
// input with the stored date. Stored text that a date input cannot show
// renders as a text input with that text, because a blank date input would
// clear the stored value on the next save.
func TestFieldDecider_DateWidgetPrefillsStringKinds(t *testing.T) {
	tests := []struct {
		name  string
		form  dateStrings
		field string
		want  string
	}{
		// A Format method gives the date as yyyy-mm-dd.
		{"Format normalizes", dateStrings{Date: "06.07.2026"}, "Date",
			`<input type="date" name="Date" id="Date" value="2026-07-06"/>`},
		{"Format keeps canonical", dateStrings{Date: "2026-07-06"}, "Date",
			`<input type="date" name="Date" id="Date" value="2026-07-06"/>`},
		{"Format through pointer", dateStrings{DatePtr: new(testDate("06.07.2026"))}, "DatePtr",
			`<input type="date" name="DatePtr" id="DatePtr" value="2026-07-06"/>`},
		// Without Format, a yyyy-mm-dd string is shown as stored.
		{"string", dateStrings{Str: "2026-07-06"}, "Str",
			`<input type="date" name="Str" id="Str" value="2026-07-06"/>`},
		{"string pointer", dateStrings{Ptr: new("2026-07-06")}, "Ptr",
			`<input type="date" name="Ptr" id="Ptr" value="2026-07-06"/>`},
		// No date, no value attribute.
		{"empty string", dateStrings{}, "Str",
			`<input type="date" name="Str" id="Str"/>`},
		{"nil pointer", dateStrings{}, "Ptr",
			`<input type="date" name="Ptr" id="Ptr"/>`},
		{"pointer to empty", dateStrings{Ptr: new("")}, "Ptr",
			`<input type="date" name="Ptr" id="Ptr"/>`},
		{"empty date", dateStrings{}, "Date",
			`<input type="date" name="Date" id="Date"/>`},
		{"Format zero date", dateStrings{Date: "0001-01-01"}, "Date",
			`<input type="date" name="Date" id="Date"/>`},
		// Anything else renders as a text input showing the stored text,
		// with the format to enter as placeholder and hint.
		{"non-canonical string", dateStrings{Str: "06.07.2026"}, "Str",
			`<input type="text" name="Str" id="Str" value="06.07.2026" placeholder="YYYY-MM-DD" aria-describedby="Str-format" aria-invalid="true"/>` +
				`<small class="form-help" id="Str-format">Expected format: YYYY-MM-DD</small>`},
		{"impossible date", dateStrings{Str: "2026-02-30"}, "Str",
			`<input type="text" name="Str" id="Str" value="2026-02-30" placeholder="YYYY-MM-DD" aria-describedby="Str-format" aria-invalid="true"/>`},
		{"text through pointer", dateStrings{Ptr: new("next week")}, "Ptr",
			`<input type="text" name="Ptr" id="Ptr" value="next week" placeholder="YYYY-MM-DD" aria-describedby="Ptr-format" aria-invalid="true"/>`},
		{"Format returns it unchanged", dateStrings{Date: "garbage"}, "Date",
			`<input type="text" name="Date" id="Date" value="garbage" placeholder="YYYY-MM-DD" aria-describedby="Date-format" aria-invalid="true"/>`},
		// The text input keeps required and drops the date range, which
		// does not apply to text.
		{"bounded date", dateStrings{Bounded: "2026-07-06"}, "Bounded",
			`<input type="date" name="Bounded" id="Bounded" value="2026-07-06" min="2026-01-01" max="2026-12-31" required="required"/>`},
		{"bounded text", dateStrings{Bounded: "06.07.2026"}, "Bounded",
			`<input type="text" name="Bounded" id="Bounded" value="06.07.2026" placeholder="YYYY-MM-DD" aria-describedby="Bounded-format" required="required" aria-invalid="true"/>`},
		// A placeholder from the tag, say a translated one, wins over the
		// format.
		{"tag placeholder", dateStrings{Localized: "06.07.2026"}, "Localized",
			`<input type="text" name="Localized" id="Localized" value="06.07.2026" placeholder="TT.MM.JJJJ" aria-describedby="Localized-format" aria-invalid="true"/>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := renderFieldHTML(t, &tt.form, tt.field)
			if !strings.Contains(out, tt.want) {
				t.Errorf("rendered\n%s\nwant the input\n%s", out, tt.want)
			}
		})
	}
}

// A submitted yyyy-mm-dd date binds into a string-kind field, through
// UnmarshalText when the type has one. An empty submission clears the
// field. Any other text is a field error, and the field holds the posted
// text so a re-render shows it; the handler saves nothing while the error
// stands.
func TestFieldDecider_DateWidgetBindsStringKinds(t *testing.T) {
	tests := []struct {
		name    string
		stored  dateStrings
		field   string
		raw     string
		want    dateStrings
		wantErr string
	}{
		// An empty submission stores the zero value.
		{name: "empty clears string", stored: dateStrings{Str: "2026-07-06"}, field: "Str", raw: "",
			want: dateStrings{}},
		{name: "empty clears pointer", stored: dateStrings{Ptr: new("2026-07-06")}, field: "Ptr", raw: "",
			want: dateStrings{}},
		{name: "empty clears date", stored: dateStrings{Date: "2026-07-06"}, field: "Date", raw: "",
			want: dateStrings{}},
		// A yyyy-mm-dd date is stored as submitted.
		{name: "string", field: "Str", raw: "2026-07-06",
			want: dateStrings{Str: "2026-07-06"}},
		{name: "allocates pointer", field: "Ptr", raw: "2026-07-06",
			want: dateStrings{Ptr: new("2026-07-06")}},
		{name: "replaces non-canonical date", stored: dateStrings{Date: "06.07.2026"}, field: "Date", raw: "2026-07-06",
			want: dateStrings{Date: "2026-07-06"}},
		{name: "date pointer", field: "DatePtr", raw: "2026-07-06",
			want: dateStrings{DatePtr: new(testDate("2026-07-06"))}},
		{name: "UnmarshalText", field: "Unmarshal", raw: "2026-07-06",
			want: dateStrings{Unmarshal: "unmarshaled 2026-07-06"}},
		// Any other text is an error, and the field holds what was posted.
		{name: "non-canonical text", stored: dateStrings{Str: "2026-07-06"}, field: "Str", raw: "06.07.2026",
			want:    dateStrings{Str: "06.07.2026"},
			wantErr: `"06.07.2026" is not a date in the format YYYY-MM-DD`},
		{name: "impossible date", stored: dateStrings{Date: "2026-07-06"}, field: "Date", raw: "2026-02-30",
			want:    dateStrings{Date: "2026-02-30"},
			wantErr: `"2026-02-30" is not a date in the format YYYY-MM-DD`},
		{name: "nil pointer gets the posted text", field: "Ptr", raw: "next week",
			want:    dateStrings{Ptr: new("next week")},
			wantErr: `"next week" is not a date in the format YYYY-MM-DD`},
		{name: "pointer gets the posted text", stored: dateStrings{Ptr: new("2026-07-06")}, field: "Ptr", raw: "next week",
			want:    dateStrings{Ptr: new("next week")},
			wantErr: `"next week" is not a date in the format YYYY-MM-DD`},
		{name: "UnmarshalText not called", field: "Unmarshal", raw: "06.07.2026",
			want:    dateStrings{Unmarshal: "06.07.2026"},
			wantErr: `"06.07.2026" is not a date in the format YYYY-MM-DD`},
		// A date the type's UnmarshalText refuses is an error too, and the
		// field holds the posted text like any other rejected submission.
		{name: "UnmarshalText accepts", field: "Strict", raw: "2026-07-06",
			want: dateStrings{Strict: new(strictDate("2026-07-06"))}},
		{name: "UnmarshalText error on nil", field: "Strict", raw: "1999-12-31",
			want:    dateStrings{Strict: new(strictDate("1999-12-31"))},
			wantErr: "dates before 2000 are not accepted"},
		{name: "UnmarshalText error on a date", stored: dateStrings{Strict: new(strictDate("2026-01-01"))}, field: "Strict", raw: "1999-12-31",
			want:    dateStrings{Strict: new(strictDate("1999-12-31"))},
			wantErr: "dates before 2000 are not accepted"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stored
			err := parseFieldValue(t, &got, tt.field, tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bind %q: stored %s, want %s", tt.raw, dumpDates(got), dumpDates(tt.want))
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

// A required date held as text rejects an empty submission like a
// required time.Time does.
func TestReflectFormHandler_RequiredDateStringRejectsBlank(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		type draft struct {
			Due string `form:"widget=date,required"`
		}
		postBlankRequired(t, "Due", &draft{Due: "2026-07-06"})
	})
	t.Run("date type", func(t *testing.T) {
		type draft struct {
			Due testDate `form:"widget=date,required"`
		}
		postBlankRequired(t, "Due", &draft{Due: "2026-07-06"})
	})
}

// A stored text that is not a date reaches the user as a text input, and
// posting it back unchanged shows the bind error at the field instead of
// saving, with the stored text still in the input.
func TestReflectFormHandler_DateStringErrorKeepsStoredText(t *testing.T) {
	type draft struct {
		Due string `form:"widget=date"`
	}
	saved := false
	handler := mx.ReflectFormHandler(
		func(context.Context) (*draft, error) { return &draft{Due: "06.07.2026"}, nil },
		func(context.Context, *draft) error { saved = true; return nil },
		FieldDecider,
	)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/draft", nil))
	shown := submittedValues(rec.Body.String())
	if got := shown.Get("Due"); got != "06.07.2026" {
		t.Fatalf("GET shows Due %q, want the stored text: %s", got, rec.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "/draft", strings.NewReader(shown.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	handler(rec, req)
	body := rec.Body.String()
	if saved || rec.Code != http.StatusOK {
		t.Fatalf("POST of a stored non-date: status %d, saved %t; want the form re-rendered with 200", rec.Code, saved)
	}
	if want := `<p class="form-error" data-error="Due">&quot;06.07.2026&quot; is not a date in the format YYYY-MM-DD</p>`; !strings.Contains(body, want) {
		t.Errorf("want the bind error at the field %s, got %s", want, body)
	}
	if want := `<input type="text" name="Due" id="Due" value="06.07.2026" placeholder="YYYY-MM-DD" aria-describedby="Due-format" aria-invalid="true"/>`; !strings.Contains(body, want) {
		t.Errorf("want the stored text re-rendered %s, got %s", want, body)
	}
}

// A GET followed by a POST without edits stores exactly the dates the GET
// showed: a valid but non-canonical date comes back canonical, and a
// missing one stays missing.
func TestReflectFormHandler_DateStringRoundTrip(t *testing.T) {
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

// parseFieldValue submits raw for fieldName and binds it into that field
// of target, a pointer to a struct, through FieldDecider's Parse.
func parseFieldValue(t *testing.T, target any, fieldName, raw string) error {
	t.Helper()
	values := url.Values{}
	values.Set(fieldName, raw)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := r.ParseForm(); err != nil {
		t.Fatal(err)
	}
	v := reflect.ValueOf(target).Elem()
	field, ok := v.Type().FieldByName(fieldName)
	if !ok {
		t.Fatalf("no field %q", fieldName)
	}
	path := mx.FieldPath(fieldName)
	value := v.FieldByIndex(field.Index)
	return FieldDecider(path, field, value).Parse(path, field, value, r)
}

// dumpDates prints d with its pointers followed, for test failures.
func dumpDates(d dateStrings) string {
	deref := func(s *string) string {
		if s == nil {
			return "nil"
		}
		return fmt.Sprintf("&%q", *s)
	}
	datePtr := "nil"
	if d.DatePtr != nil {
		datePtr = fmt.Sprintf("&%q", string(*d.DatePtr))
	}
	strict := "nil"
	if d.Strict != nil {
		strict = fmt.Sprintf("&%q", string(*d.Strict))
	}
	return fmt.Sprintf("{Str:%q Ptr:%s Date:%q DatePtr:%s Bounded:%q Unmarshal:%q Strict:%s Localized:%q}",
		d.Str, deref(d.Ptr), d.Date, datePtr, d.Bounded, d.Unmarshal, strict, d.Localized)
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
			attrs[a[1]] = stdhtml.UnescapeString(a[2])
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

// timeStrings holds a date with time, and a time of day, as text.
type timeStrings struct {
	At    string  `form:"widget=datetime"`
	AtPtr *string `form:"widget=datetime"`
	Clock string  `form:"widget=time"`
}

// The datetime and time widgets read and write text like the date widget.
// The text has no zone and carries seconds only when they are not zero,
// the form these inputs use themselves.
func TestFieldDecider_TimeWidgetsPrefillStringKinds(t *testing.T) {
	tests := []struct {
		name  string
		form  timeStrings
		field string
		want  string
	}{
		{"datetime minutes", timeStrings{At: "2026-07-06T09:05"}, "At",
			`<input type="datetime-local" name="At" id="At" value="2026-07-06T09:05"/>`},
		{"datetime drops zero seconds", timeStrings{At: "2026-07-06T09:05:00"}, "At",
			`<input type="datetime-local" name="At" id="At" value="2026-07-06T09:05"/>`},
		{"datetime seconds", timeStrings{At: "2026-07-06T09:05:07"}, "At",
			`<input type="datetime-local" name="At" id="At" value="2026-07-06T09:05:07"/>`},
		{"datetime nil pointer", timeStrings{}, "AtPtr",
			`<input type="datetime-local" name="AtPtr" id="AtPtr"/>`},
		{"datetime with zone", timeStrings{At: "2026-07-06T09:05:07Z"}, "At",
			`<input type="text" name="At" id="At" value="2026-07-06T09:05:07Z" placeholder="YYYY-MM-DDTHH:MM" aria-describedby="At-format" aria-invalid="true"/>` +
				`<small class="form-help" id="At-format">Expected format: YYYY-MM-DDTHH:MM</small>`},
		{"time minutes", timeStrings{Clock: "09:05"}, "Clock",
			`<input type="time" name="Clock" id="Clock" value="09:05"/>`},
		{"time drops zero seconds", timeStrings{Clock: "09:05:00"}, "Clock",
			`<input type="time" name="Clock" id="Clock" value="09:05"/>`},
		{"time other spelling", timeStrings{Clock: "9.05"}, "Clock",
			`<input type="text" name="Clock" id="Clock" value="9.05" placeholder="HH:MM" aria-describedby="Clock-format" aria-invalid="true"/>` +
				`<small class="form-help" id="Clock-format">Expected format: HH:MM</small>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := renderFieldHTML(t, &tt.form, tt.field)
			if !strings.Contains(out, tt.want) {
				t.Errorf("rendered\n%s\nwant\n%s", out, tt.want)
			}
		})
	}
}

// A submitted datetime or time binds as text with zero seconds dropped.
// Text in another form is a field error, and the field holds what was
// posted.
func TestFieldDecider_TimeWidgetsBindStringKinds(t *testing.T) {
	tests := []struct {
		name    string
		stored  timeStrings
		field   string
		raw     string
		want    timeStrings
		wantErr string
	}{
		{name: "datetime minutes", field: "At", raw: "2026-07-06T09:05",
			want: timeStrings{At: "2026-07-06T09:05"}},
		{name: "datetime zero seconds", field: "At", raw: "2026-07-06T09:05:00",
			want: timeStrings{At: "2026-07-06T09:05"}},
		{name: "datetime seconds", field: "At", raw: "2026-07-06T09:05:07",
			want: timeStrings{At: "2026-07-06T09:05:07"}},
		{name: "datetime allocates pointer", field: "AtPtr", raw: "2026-07-06T09:05",
			want: timeStrings{AtPtr: new("2026-07-06T09:05")}},
		{name: "datetime empty clears", stored: timeStrings{At: "2026-07-06T09:05"}, field: "At", raw: "",
			want: timeStrings{}},
		{name: "datetime with space", stored: timeStrings{At: "2026-07-06T09:05"}, field: "At", raw: "2026-07-06 09:05",
			want:    timeStrings{At: "2026-07-06 09:05"},
			wantErr: `"2026-07-06 09:05" is not a date and time in the format YYYY-MM-DDTHH:MM`},
		{name: "datetime nil pointer gets the posted text", field: "AtPtr", raw: "2026-07-06T09:05Z",
			want:    timeStrings{AtPtr: new("2026-07-06T09:05Z")},
			wantErr: `"2026-07-06T09:05Z" is not a date and time in the format YYYY-MM-DDTHH:MM`},
		{name: "time zero seconds", field: "Clock", raw: "09:05:00",
			want: timeStrings{Clock: "09:05"}},
		{name: "time one-digit hour", stored: timeStrings{Clock: "09:05"}, field: "Clock", raw: "9:05",
			want:    timeStrings{Clock: "9:05"},
			wantErr: `"9:05" is not a time in the format HH:MM`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stored
			err := parseFieldValue(t, &got, tt.field, tt.raw)
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

// A GET followed by an unchanged POST stores the datetime and time the GET
// showed, so zero seconds go and other seconds stay.
func TestReflectFormHandler_TimeStringRoundTrip(t *testing.T) {
	var saved *timeStrings
	handler := mx.ReflectFormHandler(
		func(context.Context) (*timeStrings, error) {
			return &timeStrings{At: "2026-07-06T09:05:00", AtPtr: new("2026-07-06T18:30"), Clock: "09:05:07"}, nil
		},
		func(_ context.Context, r *timeStrings) error { saved = r; return nil },
		FieldDecider,
	)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/times", nil))
	shown := submittedValues(rec.Body.String())
	req := httptest.NewRequest(http.MethodPost, "/times", strings.NewReader(shown.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("unchanged POST: status %d, want 303, body %s", rec.Code, rec.Body.String())
	}
	want := &timeStrings{At: "2026-07-06T09:05", AtPtr: new("2026-07-06T18:30"), Clock: "09:05:07"}
	if !reflect.DeepEqual(saved, want) {
		t.Errorf("saved %+v, want %+v", saved, want)
	}
}

// hintedDate declares itself a date, so its fields need no widget tag.
type hintedDate string

func (hintedDate) FormWidget() string { return "date" }

// A type whose FormWidget declares "date" renders and binds like a field
// tagged form:"widget=date", also behind a nil pointer.
func TestFieldDecider_HintedDateWidget(t *testing.T) {
	type form struct {
		Due     hintedDate
		DuePtr  *hintedDate
		Created hintedDate `form:"widget=text"`
	}
	f := &form{Due: "2026-07-06", Created: "2026-07-06"}
	for _, r := range []struct{ field, want string }{
		{"Due", `<input type="date" name="Due" id="Due" value="2026-07-06"/>`},
		{"DuePtr", `<input type="date" name="DuePtr" id="DuePtr"/>`},
		{"Created", `<input type="text" name="Created" id="Created" value="2026-07-06"/>`},
	} {
		if out := renderFieldHTML(t, f, r.field); !strings.Contains(out, r.want) {
			t.Errorf("%s renders\n%s\nwant\n%s", r.field, out, r.want)
		}
	}
	if err := parseFieldValue(t, f, "DuePtr", "2026-08-05"); err != nil {
		t.Fatalf("bind DuePtr: %v", err)
	}
	if f.DuePtr == nil || *f.DuePtr != "2026-08-05" {
		t.Errorf("DuePtr = %v, want 2026-08-05", f.DuePtr)
	}
	err := parseFieldValue(t, f, "Due", "06.07.2026")
	if want := `"06.07.2026" is not a date in the format YYYY-MM-DD`; err == nil || err.Error() != want {
		t.Errorf("bind 06.07.2026: error %v, want %s", err, want)
	}
}

// emailAddress declares the email widget.
type emailAddress string

func (emailAddress) FormWidget() string { return "email" }

// An explicit widget tag wins over the type's FormWidgetHint.
func TestFieldDecider_TagWidgetBeatsHint(t *testing.T) {
	type form struct {
		Contact emailAddress
		Code    emailAddress `form:"widget=text"`
	}
	f := &form{Contact: "a@example.com", Code: "x"}
	if out := renderFieldHTML(t, f, "Contact"); !strings.Contains(out, `type="email"`) {
		t.Errorf("the hint must render an email input: %s", out)
	}
	if out := renderFieldHTML(t, f, "Code"); !strings.Contains(out, `type="text"`) {
		t.Errorf("widget=text must win over the email hint: %s", out)
	}
}

// The rows of a repeatable field are bound from the submission alone, so a
// rejected date text has to stay in its row: an add-row click re-renders it
// in the text input, and an unchanged Save is rejected instead of storing "".
func TestReflectFormHandler_RepeatableDateTextSurvivesCommands(t *testing.T) {
	type line struct {
		Due string `form:"widget=date"`
	}
	type order struct {
		Lines []line `form:"repeatable"`
	}
	saved := false
	handler := mx.ReflectFormHandler(
		func(context.Context) (*order, error) { return &order{Lines: []line{{Due: "06.07.2026"}}}, nil },
		func(context.Context, *order) error { saved = true; return nil },
		FieldDecider,
	)
	post := func(values url.Values) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/order", strings.NewReader(values.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		handler(rec, req)
		return rec
	}
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/order", nil))
	shown := submittedValues(rec.Body.String())
	shown.Set("__cmd__", mx.AddRowCommand("Lines"))
	added := post(shown)
	if want := `type="text" name="Lines-0-Due" id="Lines-0-Due" value="06.07.2026"`; !strings.Contains(added.Body.String(), want) {
		t.Fatalf("after add-row the row must still show its text, want %s in %s", want, added.Body.String())
	}

	resubmit := submittedValues(added.Body.String())
	resubmit.Del("__cmd__")
	rec = post(resubmit)
	if saved || rec.Code != http.StatusOK {
		t.Errorf("unchanged Save of a row's non-date text: status %d, saved %t; want the form re-rendered with 200", rec.Code, saved)
	}
	if want := `is not a date in the format YYYY-MM-DD`; !strings.Contains(rec.Body.String(), want) {
		t.Errorf("want the row's bind error %q in %s", want, rec.Body.String())
	}
}

// civilDate is a date struct with a text form, the shape of civil.Date, that
// declares the date widget.
type civilDate struct {
	Year, Month, Day int
}

func (d civilDate) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%04d-%02d-%02d", d.Year, d.Month, d.Day), nil
}

func (d *civilDate) UnmarshalText(text []byte) error {
	t, err := time.Parse(time.DateOnly, string(text))
	if err != nil {
		return err
	}
	*d = civilDate{t.Year(), int(t.Month()), t.Day()}
	return nil
}

func (civilDate) FormWidget() string { return "date" }

// A date struct that declares the date widget shows its MarshalText in the
// date input, so an unchanged Save keeps it instead of storing the zero
// value. Its zero value renders no value.
func TestReflectFormHandler_DateStructRoundTrip(t *testing.T) {
	type record struct {
		Due   civilDate
		Empty civilDate
	}
	var saved *record
	handler := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{Due: civilDate{2026, 7, 6}}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/record", nil))
	body := rec.Body.String()
	for _, want := range []string{
		`<input type="date" name="Due" id="Due" value="2026-07-06"/>`,
		`<input type="date" name="Empty" id="Empty"/>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("GET lacks %s: %s", want, body)
		}
	}
	req := httptest.NewRequest(http.MethodPost, "/record", strings.NewReader(submittedValues(body).Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("unchanged POST: status %d, want 303, body %s", rec.Code, rec.Body.String())
	}
	if want := (record{Due: civilDate{2026, 7, 6}}); saved == nil || *saved != want {
		t.Errorf("saved %+v, want %+v", saved, want)
	}
}

// wrappedTime embeds time.Time, the shape of a null-time type.
type wrappedTime struct {
	time.Time
}

// definedTime is defined as time.Time, without its methods.
type definedTime time.Time

// Types that wrap time.Time render and bind like a time.Time: the embedded
// time takes the submitted value through its RFC 3339 UnmarshalText, a type
// defined as time.Time the converted value.
func TestFieldDecider_WrappedTimeWidgets(t *testing.T) {
	type form struct {
		Day     wrappedTime  `form:"widget=date"`
		At      wrappedTime  `form:"widget=datetime"`
		Clock   *wrappedTime `form:"widget=time"`
		Defined definedTime  `form:"widget=date"`
	}
	f := &form{Day: wrappedTime{time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)}}
	if out, want := renderFieldHTML(t, f, "Day"), `<input type="date" name="Day" id="Day" value="2026-07-06"/>`; !strings.Contains(out, want) {
		t.Errorf("Day renders\n%s\nwant\n%s", out, want)
	}
	binds := []struct {
		field, raw string
		got        func() time.Time
		want       time.Time
	}{
		{"Day", "2026-08-05", func() time.Time { return f.Day.Time }, time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)},
		{"At", "2026-08-05T10:30", func() time.Time { return f.At.Time }, time.Date(2026, 8, 5, 10, 30, 0, 0, time.UTC)},
		{"Clock", "09:05", func() time.Time { return f.Clock.Time }, time.Date(0, 1, 1, 9, 5, 0, 0, time.UTC)},
		{"Defined", "2026-08-05", func() time.Time { return time.Time(f.Defined) }, time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)},
	}
	for _, b := range binds {
		if err := parseFieldValue(t, f, b.field, b.raw); err != nil {
			t.Errorf("bind %s %q: %v", b.field, b.raw, err)
			continue
		}
		if got := b.got(); !got.Equal(b.want) {
			t.Errorf("bind %s %q: stored %v, want %v", b.field, b.raw, got, b.want)
		}
	}
}

// A field with min and max rejects a posted date outside the range on the
// server, shows the error, and marks the re-rendered date input invalid.
func TestReflectFormHandler_DateOutOfRangeRejected(t *testing.T) {
	type draft struct {
		Day string `form:"widget=date,min=2026-01-01,max=2026-12-31"`
	}
	saved := false
	handler := mx.ReflectFormHandler(
		func(context.Context) (*draft, error) { return &draft{Day: "2026-07-06"}, nil },
		func(context.Context, *draft) error { saved = true; return nil },
		FieldDecider,
	)
	form := url.Values{}
	form.Set(mx.PresentSentinelName("Day"), "1")
	form.Set("Day", "1900-01-01")
	req := httptest.NewRequest(http.MethodPost, "/draft", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler(rec, req)
	body := rec.Body.String()
	if saved || rec.Code != http.StatusOK {
		t.Fatalf("out-of-range POST: status %d, saved %t; want the form re-rendered with 200", rec.Code, saved)
	}
	for _, want := range []string{
		`&quot;1900-01-01&quot; is earlier than 2026-01-01`,
		`<input type="date" name="Day" id="Day" value="1900-01-01" min="2026-01-01" max="2026-12-31" aria-invalid="true"/>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("want %s in %s", want, body)
		}
	}
}
