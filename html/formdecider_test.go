package html

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	stdhtml "html"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
	_ "time/tzdata"
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
	mx.RegisterNamedOptions("test-html-partners-message", func(context.Context) ([]mx.NamedOption, error) {
		return nil, fmt.Errorf("partner lookup: %w", mx.FormMessage("The partner list is unavailable, try again later."))
	})
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

// A save submits nothing for the disabled placeholder: an optional select
// is cleared, so its caption says so, while a required one rejects the
// blank and a readonly one is never saved.
func TestFieldDecider_SelectPlaceholderSaysRemoval(t *testing.T) {
	type draftForm struct {
		Optional string `form:"options=test-html-partners"`
		Required string `form:"options=test-html-partners,required"`
		Readonly string `form:"options=test-html-partners,readonly"`
	}
	target := &draftForm{Optional: "gone", Required: "gone", Readonly: "gone"}
	v := reflect.ValueOf(target).Elem()
	ctx := context.WithValue(context.Background(), partnersCtxKey{},
		[]mx.NamedOption{{Name: "Partner One", Value: "p1"}})
	for name, removed := range map[string]bool{"Optional": true, "Required": false, "Readonly": false} {
		field, _ := v.Type().FieldByName(name)
		value := v.FieldByName(name)
		comp := FieldDecider(mx.FieldPath(name), field, value).Render(mx.FieldPath(name), field, value, nil)
		var b strings.Builder
		if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
			t.Fatalf("%s: render: %v", name, err)
		}
		out := b.String()
		if !strings.Contains(out, `selected="selected">(current value not available)</option>`) {
			t.Errorf("%s: want the short caption in %q", name, out)
		}
		if got := strings.Contains(out, `<p class="form-help" id="`+name+`-note">The current value is removed when you save.</p>`); got != removed {
			t.Errorf("%s: removal note shown %t, want %t: %q", name, got, removed, out)
		}
		// The select is described by the note, so a screen reader hears it.
		if got := strings.Contains(out, `aria-describedby="`+name+`-note"`); got != removed {
			t.Errorf("%s: described by the removal note %t, want %t: %q", name, got, removed, out)
		}
	}
}

// A set's field label names its group, as no control has the path as its
// id.
func TestFieldDecider_EnumSetLabelNamesGroup(t *testing.T) {
	type tagForm struct {
		Tags []string `form:"options=test-html-tags"`
	}
	out := renderFieldHTML(t, &tagForm{}, "Tags")
	if !strings.Contains(out, `<label id="Tags-label">Tags:</label><div role="group" aria-labelledby="Tags-label" aria-describedby="Tags-note">`) {
		t.Errorf("want the group named by its label: %s", out)
	}
}

// The placeholder keeps its promise on save: the disabled option submits
// nothing, so an optional select is cleared, and a readonly one keeps its
// value.
func TestReflectFormHandler_PlaceholderSaveRemovesValue(t *testing.T) {
	type draft struct {
		Partner  string `form:"options=test-html-partners"`
		Readonly string `form:"options=test-html-partners,readonly"`
	}
	var saved *draft
	inner := mx.ReflectFormHandler(
		func(context.Context) (*draft, error) { return &draft{Partner: "gone", Readonly: "gone"}, nil },
		func(_ context.Context, d *draft) error { saved = d; return nil },
		FieldDecider,
	)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inner(w, r.WithContext(context.WithValue(r.Context(), partnersCtxKey{}, []mx.NamedOption{{Name: "Partner One", Value: "p1"}})))
	})
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec := postValues(t, h, submittedValues(get.Body.String())); rec.Code != http.StatusSeeOther || saved == nil || saved.Partner != "" || saved.Readonly != "gone" {
		t.Errorf("status %d, saved %+v; want Partner cleared and Readonly kept", rec.Code, saved)
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
	if !strings.Contains(out, "(some current values are not available and are removed when you save)") {
		t.Errorf("missing members must be announced without naming them: %q", out)
	}
	if strings.Count(out, `type="checkbox"`) != 2 {
		t.Errorf("only the listed options are checkboxes: %q", out)
	}
	// A readonly set is never saved and keeps them, so it shows no note.
	type readonlyForm struct {
		Tags []string `form:"options=test-html-tags,readonly"`
	}
	if out := renderFieldHTML(t, &readonlyForm{Tags: []string{"kept", "b"}}, "Tags"); !strings.Contains(out, ">(some current values are not available)<") || strings.Contains(out, "removed") {
		t.Errorf("a readonly set says some values are not available, but not that they are removed: %q", out)
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
	if !strings.Contains(out, "(some current values are not available and are removed when you save)") {
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
		if !strings.Contains(out, "(some current values are not available and are removed when you save)") {
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
	// The provider's text stays out of the page (see
	// TestFieldDecider_OptionListParseErrorNotShown); the failure does not.
	err := FieldDecider("Tags", field, v.Field(0)).Parse("Tags", field, v.Field(0), r)
	if err == nil || err.Error() != mx.FormErrorMessage {
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
	if !strings.Contains(rec.Body.String(), "(some current values are not available and are removed when you save)") {
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
	if !strings.Contains(body, "(current value not available)") || !strings.Contains(body, "The current value is removed when you save.") {
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
	Localized string          `form:"widget=date,placeholder=JJJJ-MM-TT"`
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
				`<small class="form-help" id="Str-format">Not in the expected format: YYYY-MM-DD</small>`},
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
		// A placeholder from the tag, say a translated format, replaces the
		// placeholder; the hint names the format the parser reads.
		{"tag placeholder", dateStrings{Localized: "06.07.2026"}, "Localized",
			`<input type="text" name="Localized" id="Localized" value="06.07.2026" placeholder="JJJJ-MM-TT" aria-describedby="Localized-format" aria-invalid="true"/><small class="form-help" id="Localized-format">Not in the expected format: YYYY-MM-DD</small>`},
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
				`<small class="form-help" id="At-format">Not in the expected format: YYYY-MM-DDTHH:MM</small>`},
		{"time minutes", timeStrings{Clock: "09:05"}, "Clock",
			`<input type="time" name="Clock" id="Clock" value="09:05"/>`},
		{"time drops zero seconds", timeStrings{Clock: "09:05:00"}, "Clock",
			`<input type="time" name="Clock" id="Clock" value="09:05"/>`},
		{"time other spelling", timeStrings{Clock: "9.05"}, "Clock",
			`<input type="text" name="Clock" id="Clock" value="9.05" placeholder="HH:MM" aria-describedby="Clock-format" aria-invalid="true"/>` +
				`<small class="form-help" id="Clock-format">Not in the expected format: HH:MM</small>`},
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

// civTime is a time-of-day struct shaped like civil.Time: its text always
// carries seconds.
type civTime struct {
	H, M, S int
}

func (c civTime) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%02d:%02d:%02d", c.H, c.M, c.S), nil
}

func (c *civTime) UnmarshalText(text []byte) error {
	t, err := time.Parse(time.TimeOnly, string(text))
	if err != nil {
		return err
	}
	*c = civTime{t.Hour(), t.Minute(), t.Second()}
	return nil
}

// civDateTime is a date-and-time struct shaped like civil.DateTime.
type civDateTime struct {
	T time.Time
}

func (c civDateTime) MarshalText() ([]byte, error) {
	return []byte(c.T.Format("2006-01-02T15:04:05")), nil
}

func (c *civDateTime) UnmarshalText(text []byte) error {
	t, err := time.Parse("2006-01-02T15:04:05", string(text))
	if err != nil {
		return err
	}
	c.T = t
	return nil
}

// A time-of-day or date-and-time struct whose text needs seconds survives a
// GET and an unchanged POST: the inputs leave zero seconds out, so the
// parser hands UnmarshalText the value with seconds too.
func TestReflectFormHandler_TimeStructRoundTrip(t *testing.T) {
	type record struct {
		Start civTime     `form:"widget=time"`
		At    civDateTime `form:"widget=datetime"`
	}
	stored := record{Start: civTime{9, 5, 0}, At: civDateTime{time.Date(2026, 7, 6, 9, 5, 0, 0, time.UTC)}}
	var saved *record
	handler := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { r := stored; return &r, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/record", nil))
	shown := submittedValues(rec.Body.String())
	if shown.Get("Start") != "09:05" || shown.Get("At") != "2026-07-06T09:05" {
		t.Fatalf("GET shows Start %q and At %q, want 09:05 and 2026-07-06T09:05", shown.Get("Start"), shown.Get("At"))
	}
	req := httptest.NewRequest(http.MethodPost, "/record", strings.NewReader(shown.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("unchanged POST: status %d, want 303, body %s", rec.Code, rec.Body.String())
	}
	if saved == nil || *saved != stored {
		t.Errorf("saved %+v, want %+v", saved, stored)
	}
}

// nullTime embeds a time next to its own validity flag and reports itself
// null without it, but has no UnmarshalText of its own: the promoted one
// sets only the time.
type nullTime struct {
	time.Time
	Valid bool
}

func (n nullTime) IsNull() bool { return !n.Valid }

// sqlNullTime is a nullable time like null.Time: it embeds sql.NullTime
// and its own RFC 3339 text methods set the validity flag too.
type sqlNullTime struct {
	sql.NullTime
}

func (n sqlNullTime) MarshalText() ([]byte, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time.MarshalText()
}

func (n *sqlNullTime) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*n = sqlNullTime{}
		return nil
	}
	if err := n.Time.UnmarshalText(text); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// A nullable time with its own RFC 3339 text methods round-trips through a
// date input, at the top level and in a repeatable row after an add-row
// click, where it is bound into a fresh struct.
func TestReflectFormHandler_NullableTimeRoundTrip(t *testing.T) {
	type line struct {
		Due sqlNullTime `form:"widget=date"`
	}
	type order struct {
		Due   sqlNullTime `form:"widget=date"`
		Lines []line      `form:"repeatable"`
	}
	july := sqlNullTime{sql.NullTime{Time: time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC), Valid: true}}
	var saved *order
	handler := mx.ReflectFormHandler(
		func(context.Context) (*order, error) { return &order{Due: july, Lines: []line{{Due: july}}}, nil },
		func(_ context.Context, r *order) error { saved = r; return nil },
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
	if shown.Get("Due") != "2026-07-06" || shown.Get("Lines-0-Due") != "2026-07-06" {
		t.Fatalf("GET shows %q and %q, want 2026-07-06", shown.Get("Due"), shown.Get("Lines-0-Due"))
	}
	shown.Set("__cmd__", mx.AddRowCommand("Lines"))
	added := submittedValues(post(shown).Body.String())
	added.Del("__cmd__")
	if rec = post(added); rec.Code != http.StatusSeeOther {
		t.Fatalf("unchanged Save: status %d, body %s", rec.Code, rec.Body.String())
	}
	if saved == nil || !saved.Due.Valid || !saved.Due.Time.Equal(july.Time) || len(saved.Lines) != 2 || !saved.Lines[0].Due.Valid || !saved.Lines[0].Due.Time.Equal(july.Time) {
		t.Errorf("saved %+v, want the date kept and valid", saved)
	}
}

// ptrWrappedTime embeds a pointer to a time; its promoted methods panic on
// a nil one.
type ptrWrappedTime struct {
	*time.Time
}

// compactDate stores 20260706 but writes and reads ISO text.
type compactDate string

func (d compactDate) MarshalText() ([]byte, error) {
	t, err := time.Parse("20060102", string(d))
	if err != nil {
		return []byte(d), nil
	}
	return []byte(t.Format(time.DateOnly)), nil
}

func (d *compactDate) UnmarshalText(text []byte) error {
	t, err := time.Parse(time.DateOnly, string(text))
	if err != nil {
		return err
	}
	*d = compactDate(t.Format("20060102"))
	return nil
}

// sqlTime stores SQL-style text with a space, the README's recipe for such
// a column: Format writes the input's layout, UnmarshalText its own.
type sqlTime string

func (s sqlTime) Format(layout string) string {
	t, err := time.Parse(time.DateTime, string(s))
	if err != nil {
		return string(s)
	}
	return t.Format(layout)
}

func (s *sqlTime) UnmarshalText(text []byte) error {
	for _, layout := range []string{"2006-01-02T15:04:05", "2006-01-02T15:04"} {
		if t, err := time.Parse(layout, string(text)); err == nil {
			*s = sqlTime(t.Format(time.DateTime))
			return nil
		}
	}
	return fmt.Errorf("bad time %q", text)
}

// The struct and wrapper kinds bind like a time.Time: an empty submission
// stores the zero value, an invalid one is an error that keeps the value,
// a struct holding more than a time is never half set, and a nil embedded
// pointer is an error, not a panic.
func TestFieldDecider_DateStructBinds(t *testing.T) {
	type form struct {
		Date    civilDate      `form:"widget=date"`
		DateAt  civilDate      `form:"widget=datetime"`
		Defined definedTime    `form:"widget=date"`
		Wrapped wrappedTime    `form:"widget=date"`
		Null    nullTime       `form:"widget=date"`
		Ptr     ptrWrappedTime `form:"widget=date"`
		WrapPtr *wrappedTime   `form:"widget=date"`
		Compact compactDate    `form:"widget=date"`
		SQL     sqlTime        `form:"widget=datetime"`
		NullAt  nullTime       `form:"widget=datetime"`
		NilText nilTextDate    `form:"widget=date"`
	}
	july := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	fresh := func() *form {
		return &form{
			Date: civilDate{2026, 7, 6}, DateAt: civilDate{2026, 7, 6}, Defined: definedTime(july),
			Wrapped: wrappedTime{july}, WrapPtr: &wrappedTime{july},
		}
	}
	tests := []struct {
		name    string
		field   string
		raw     string
		wantErr bool
		check   func(f *form) bool
	}{
		{"struct empty clears", "Date", "", false, func(f *form) bool { return f.Date == civilDate{} }},
		{"struct invalid keeps", "Date", "06.07.2026", true, func(f *form) bool { return f.Date == civilDate{2026, 7, 6} }},
		{"struct in another form", "DateAt", "2026-08-05T10:30", true, func(f *form) bool { return f.DateAt == civilDate{2026, 7, 6} }},
		{"defined empty clears", "Defined", "", false, func(f *form) bool { return time.Time(f.Defined).IsZero() }},
		{"defined invalid keeps", "Defined", "06.07.2026", true, func(f *form) bool { return time.Time(f.Defined).Equal(july) }},
		{"wrapped invalid keeps", "Wrapped", "06.07.2026", true, func(f *form) bool { return f.Wrapped.Equal(july) }},
		{"wrapped pointer empty clears", "WrapPtr", "", false, func(f *form) bool { return f.WrapPtr == nil }},
		{"wrapped pointer invalid keeps", "WrapPtr", "06.07.2026", true, func(f *form) bool { return f.WrapPtr != nil && f.WrapPtr.Equal(july) }},
		{"time with a flag is not half set", "Null", "2026-08-05", true, func(f *form) bool { return !f.Null.Valid && f.Null.IsZero() }},
		// Nor by RFC 3339 posted directly, which the promoted UnmarshalText reads.
		{"time with a flag not half set by RFC 3339", "Null", "2026-08-05T00:00:00Z", true, func(f *form) bool { return !f.Null.Valid && f.Null.IsZero() }},
		{"datetime with a flag not half set by RFC 3339", "NullAt", "2026-08-05T10:30:00Z", true, func(f *form) bool { return !f.NullAt.Valid && f.NullAt.IsZero() }},
		// A promoted UnmarshalText on a nil embedded pointer panics; that is
		// an error, not a failed request.
		{"nil embedded text pointer", "NilText", "2026-08-05", true, func(f *form) bool { return f.NilText.civilDate == nil }},
		// A nil embedded time pointer gets a time to take the date in.
		{"nil embedded pointer takes a date", "Ptr", "2026-08-05", false, func(f *form) bool {
			return f.Ptr.Time != nil && f.Ptr.Equal(time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC))
		}},
		{"string type with UnmarshalText", "Compact", "2026-08-01", false, func(f *form) bool { return f.Compact == "20260801" }},
		{"SQL-style text", "SQL", "2026-08-05T10:30", false, func(f *form) bool { return f.SQL == "2026-08-05 10:30:00" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fresh()
			err := parseFieldValue(t, f, tt.field, tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("bind %s %q: error %v, want error %t", tt.field, tt.raw, err, tt.wantErr)
			}
			if !tt.check(f) {
				t.Errorf("bind %s %q: stored %+v", tt.field, tt.raw, f)
			}
		})
	}
}

// A nil embedded time pointer renders an empty input instead of panicking,
// and the SQL-style recipe shows its text in the input's form.
func TestFieldDecider_DateStructRenders(t *testing.T) {
	type form struct {
		Ptr ptrWrappedTime `form:"widget=date"`
		SQL sqlTime        `form:"widget=datetime"`
		At  string         `form:"widget=datetime,min=2026-01-01T00:00"`
		Old string         `form:"widget=date,readonly"`
	}
	f := &form{SQL: "2026-07-06 09:05:00", At: "2026-07-06T09:05:07", Old: "06.07.2026"}
	for _, r := range []struct{ field, want string }{
		{"Ptr", `<input type="date" name="Ptr" id="Ptr"/>`},
		{"SQL", `<input type="datetime-local" name="SQL" id="SQL" value="2026-07-06T09:05"/>`},
		// With a min, a value with seconds gets step 1 so the browser
		// accepts it unchanged.
		{"At", `<input type="datetime-local" name="At" id="At" value="2026-07-06T09:05:07" min="2026-01-01T00:00" step="1"/>`},
		// A readonly field is never saved: its text input asks for nothing.
		{"Old", `<input type="text" name="Old" id="Old" value="06.07.2026" readonly="readonly"/>`},
	} {
		out := renderFieldHTML(t, f, r.field)
		if !strings.Contains(out, r.want) {
			t.Errorf("%s renders\n%s\nwant\n%s", r.field, out, r.want)
		}
		if r.field == "Old" && strings.Contains(out, "Expected format") {
			t.Errorf("a readonly text input needs no format hint: %s", out)
		}
	}
}

// looseDate writes 2026-7-6, a spelling the date input cannot show, and
// reads both that and 2026-07-06.
type looseDate struct {
	Year, Month, Day int
}

func (d looseDate) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%d-%d-%d", d.Year, d.Month, d.Day), nil
}

func (d *looseDate) UnmarshalText(text []byte) error {
	for _, layout := range []string{"2006-1-2", time.DateOnly} {
		if t, err := time.Parse(layout, string(text)); err == nil {
			*d = looseDate{t.Year(), int(t.Month()), t.Day()}
			return nil
		}
	}
	return fmt.Errorf("bad date %q", text)
}

func (looseDate) FormWidget() string { return "date" }

// A date struct whose own text the picker cannot show reads that text back
// itself, so its text input is not marked invalid, and an unchanged save
// keeps the value: at the top level and in a repeatable row after an
// add-row click, where the row is bound from the submission alone.
func TestReflectFormHandler_DateStructOwnText(t *testing.T) {
	type line struct {
		Due looseDate
	}
	type order struct {
		Due   looseDate
		Lines []line `form:"repeatable"`
	}
	stored := order{Due: looseDate{2026, 7, 6}, Lines: []line{{Due: looseDate{2026, 8, 5}}}}
	var saved *order
	handler := mx.ReflectFormHandler(
		func(context.Context) (*order, error) {
			r := stored
			r.Lines = append([]line(nil), stored.Lines...)
			return &r, nil
		},
		func(_ context.Context, r *order) error { saved = r; return nil },
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
	body := rec.Body.String()
	if want := `<input type="text" name="Due" id="Due" value="2026-7-6" placeholder="YYYY-MM-DD" aria-describedby="Due-format"/><small class="form-help" id="Due-format">Expected format: YYYY-MM-DD</small>`; !strings.Contains(body, want) {
		t.Fatalf("GET lacks %s: %s", want, body)
	}
	shown := submittedValues(body)
	shown.Set("__cmd__", mx.AddRowCommand("Lines"))
	added := post(shown)
	if want := `value="2026-8-5"`; !strings.Contains(added.Body.String(), want) {
		t.Fatalf("after add-row the row must keep its text %s: %s", want, added.Body.String())
	}
	resubmit := submittedValues(added.Body.String())
	resubmit.Del("__cmd__")
	rec = post(resubmit)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("unchanged Save: status %d, want 303, body %s", rec.Code, rec.Body.String())
	}
	if saved == nil || saved.Due != stored.Due || len(saved.Lines) != 2 || saved.Lines[0].Due != stored.Lines[0].Due {
		t.Errorf("saved %+v, want Due %v and row 0 %v kept", saved, stored.Due, stored.Lines[0].Due)
	}
}

// The range is checked on the server for every kind, not only for text: a
// time.Time, a type defined as it, a wrapper and a date struct reject a
// posted date outside min/max. The posted date is bound even so, as
// rejected text is, so the re-render shows it; the handler never saves a
// record while a field has an error.
func TestFieldDecider_DateRangeOnServer(t *testing.T) {
	type form struct {
		Time    time.Time   `form:"widget=date,min=2026-01-01,max=2026-12-31"`
		Defined definedTime `form:"widget=date,min=2026-01-01"`
		Wrapped wrappedTime `form:"widget=date,max=2026-12-31"`
		Struct  civilDate   `form:"widget=date,min=2026-01-01"`
	}
	july := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	f := &form{Time: july, Defined: definedTime(july), Wrapped: wrappedTime{july}, Struct: civilDate{2026, 7, 6}}
	for _, b := range []struct{ field, raw, wantErr string }{
		{"Time", "1900-01-01", `"1900-01-01" is earlier than 2026-01-01`},
		{"Time", "2027-01-01", `"2027-01-01" is later than 2026-12-31`},
		{"Defined", "1900-01-01", `"1900-01-01" is earlier than 2026-01-01`},
		{"Wrapped", "2027-01-01", `"2027-01-01" is later than 2026-12-31`},
		{"Struct", "1900-01-01", `"1900-01-01" is earlier than 2026-01-01`},
	} {
		if err := parseFieldValue(t, f, b.field, b.raw); err == nil || err.Error() != b.wantErr {
			t.Errorf("bind %s %q: error %v, want %s", b.field, b.raw, err, b.wantErr)
		}
	}
	jan2027 := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	jan1900 := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	if !f.Time.Equal(jan2027) || !time.Time(f.Defined).Equal(jan1900) || !f.Wrapped.Equal(jan2027) || f.Struct != (civilDate{1900, 1, 1}) {
		t.Errorf("a date outside the range must still be bound: %+v", f)
	}
}

// A stored date outside a min that was tightened later survives an add-row
// click in a repeatable row, where the row is bound from the submission
// alone and the command discards bind errors. The unchanged Save is then
// rejected rather than storing the zero value for a row nobody touched.
func TestReflectFormHandler_RepeatableOutOfRangeDateSurvivesCommands(t *testing.T) {
	type line struct {
		Due   time.Time `form:"widget=date,min=2026-01-01"`
		Civil civilDate `form:"widget=date,min=2026-01-01"`
	}
	type order struct {
		Lines []line `form:"repeatable"`
	}
	saved := false
	handler := mx.ReflectFormHandler(
		func(context.Context) (*order, error) {
			return &order{Lines: []line{{Due: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), Civil: civilDate{2025, 6, 1}}}}, nil
		},
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
	added := submittedValues(post(shown).Body.String())
	for _, name := range []string{"Lines-0-Due", "Lines-0-Civil"} {
		if got := added.Get(name); got != "2025-06-01" {
			t.Errorf("after add-row %s = %q, want the stored 2025-06-01", name, got)
		}
	}
	added.Del("__cmd__")
	rec = post(added)
	if rec.Code != http.StatusOK || saved {
		t.Errorf("unchanged Save: status %d, saved %t; want 200 and not saved", rec.Code, saved)
	}
	if want := "earlier than 2026-01-01"; !strings.Contains(rec.Body.String(), want) {
		t.Errorf("Save must name the range: lacks %s", want)
	}
}

// germanDate reads 6.7.2026 as well as 2026-07-06 and writes ISO text.
type germanDate struct {
	Year, Month, Day int
}

func (d germanDate) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%04d-%02d-%02d", d.Year, d.Month, d.Day), nil
}

func (d *germanDate) UnmarshalText(text []byte) error {
	for _, layout := range []string{"2.1.2006", time.DateOnly} {
		if t, err := time.Parse(layout, string(text)); err == nil {
			*d = germanDate{t.Year(), int(t.Month()), t.Day()}
			return nil
		}
	}
	return fmt.Errorf("bad date %q", text)
}

// A date struct's own spelling, which no input layout reads, is range
// checked on the value as the input shows it. Where the input can show that
// value only as text, as looseDate's 2026-7-6, the range cannot be checked,
// so the spelling is rejected while a bound is set, and the text input
// rendering such a value is marked invalid.
func TestFieldDecider_DateStructOwnTextRange(t *testing.T) {
	type form struct {
		Loose  looseDate  `form:"widget=date,min=2026-01-01"`
		German germanDate `form:"widget=date,min=2026-01-01"`
	}
	f := &form{}
	for _, b := range []struct{ field, raw, wantErr string }{
		{"Loose", "2025-1-1", `"2025-1-1" is not a date in the format YYYY-MM-DD`},
		{"Loose", "2026-7-6", `"2026-7-6" is not a date in the format YYYY-MM-DD`},
		{"Loose", "2025-01-01", `"2025-01-01" is earlier than 2026-01-01`},
		{"Loose", "2026-07-06", ""},
		{"German", "6.7.2025", `"2025-07-06" is earlier than 2026-01-01`},
		{"German", "6.7.2026", ""},
	} {
		err := parseFieldValue(t, f, b.field, b.raw)
		if got := fmt.Sprint(err); (b.wantErr == "" && err != nil) || (b.wantErr != "" && got != b.wantErr) {
			t.Errorf("bind %s %q: error %v, want %q", b.field, b.raw, err, b.wantErr)
		}
	}
	if f.Loose != (looseDate{2026, 7, 6}) || f.German != (germanDate{2026, 7, 6}) {
		t.Errorf("valid dates must be bound: %+v", f)
	}
	if out := renderFieldHTML(t, f, "Loose"); !strings.Contains(out, `type="text"`) || !strings.Contains(out, `aria-invalid="true"`) {
		t.Errorf("own text that cannot be range checked must render as an invalid text input: %s", out)
	}
}

// The bounds have no zone, so the wall clock of a zoned submission is
// compared, as the form will show it: on a new record the posted one, else
// in the stored time's zone. A crafted RFC 3339 offset cannot move a date
// across min or max. A time range may wrap past midnight, as a browser's
// time input allows.
func TestFieldDecider_DateRangeWallClockAndOvernight(t *testing.T) {
	type form struct {
		At      time.Time   `form:"min=2024-01-01T00:00,max=2024-12-31T23:59"`
		Start   time.Time   `form:"widget=time,min=22:00,max=06:00"`
		End     string      `form:"widget=time,min=22:00,max=06:00"`
		Wrapped wrappedTime `form:"widget=date,max=2026-12-31"`
		Opaque  opaqueDate  `form:"widget=date,min=2026-01-01"`
	}
	for _, b := range []struct{ field, raw, wantErr string }{
		{"At", "2023-12-31T13:00:00-12:00", `"2023-12-31T13:00:00-12:00" is earlier than 2024-01-01T00:00`},
		{"At", "2025-01-01T09:00:00+14:00", `"2025-01-01T09:00:00+14:00" is later than 2024-12-31T23:59`},
		{"Start", "23:30", ""},
		{"Start", "01:00", ""},
		{"Start", "12:00", `"12:00" is later than 06:00 and earlier than 22:00`},
		{"End", "22:00", ""},
		{"End", "06:00", ""},
		{"End", "06:01", `"06:01" is later than 06:00 and earlier than 22:00`},
		// RFC 3339 caps offsets below 24 hours; a larger one cannot be
		// marshaled again, so it is no value.
		{"At", "2024-06-01T00:00:00+24:00", `"2024-06-01T00:00:00+24:00" is not a date and time in the format YYYY-MM-DDTHH:MM`},
		{"At", "2024-06-01T00:00:00-24:00", `"2024-06-01T00:00:00-24:00" is not a date and time in the format YYYY-MM-DDTHH:MM`},
		{"At", "2024-06-01T12:00:00+23:59", ""},
		// A wrapped time only takes what the date input sends, never its
		// own RFC 3339, so a crafted offset cannot slip past max.
		{"Wrapped", "2099-01-01T00:00:00+24:00", `parsing time "2099-01-01T00:00:00+24:00": extra text: "T00:00:00+24:00"`},
		// A value whose MarshalText fails cannot be checked against min.
		{"Opaque", "sometime", `"sometime" is not a date in the format YYYY-MM-DD`},
	} {
		f := &form{} // each case binds into a new record
		err := parseFieldValue(t, f, b.field, b.raw)
		if got := fmt.Sprint(err); (b.wantErr == "" && err != nil) || (b.wantErr != "" && got != b.wantErr) {
			t.Errorf("bind %s %q: error %v, want %q", b.field, b.raw, err, b.wantErr)
		}
	}
	// With a stored time, a posted offset is converted into its zone, and
	// that wall clock is what the form shows and what is checked: 11:00 at
	// -12:00 is 23:00 in the stored UTC, past max.
	type stored struct {
		At time.Time `form:"max=2026-01-01T12:00"`
	}
	f := &stored{At: time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC)}
	if err := parseFieldValue(t, f, "At", "2026-01-01T11:00:00-12:00"); fmt.Sprint(err) != `"2026-01-01T11:00:00-12:00" is later than 2026-01-01T12:00` {
		t.Errorf("a posted offset converted past max: error %v, stored %v", err, f.At)
	}
}

// A rebuilt repeatable row starts as a copy of its loaded row, so a typo
// followed by an add-row click keeps the stored date instead of blanking
// the row, and a flagged or pointer-embedded time binds in a row as it
// does at the top level.
func TestReflectFormHandler_RepeatableRowKeepsStoredDates(t *testing.T) {
	type line struct {
		Loose looseDate
		Time  time.Time      `form:"widget=date"`
		Null  nullTime       `form:"widget=date"`
		Ptr   ptrWrappedTime `form:"widget=date"`
	}
	type sheet struct {
		Lines []line `form:"repeatable"`
	}
	july := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	stored := func() line {
		return line{Loose: looseDate{2026, 8, 5}, Time: july, Null: nullTime{july, true}, Ptr: ptrWrappedTime{new(july)}}
	}
	var saved *sheet
	handler := mx.ReflectFormHandler(
		func(context.Context) (*sheet, error) { return &sheet{Lines: []line{stored()}}, nil },
		func(_ context.Context, s *sheet) error { saved = s; return nil },
		FieldDecider,
	)
	post := func(values url.Values) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/s", strings.NewReader(values.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		handler(rec, req)
		return rec
	}
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/s", nil))
	shown := submittedValues(rec.Body.String())
	shown.Set("Lines-0-Loose", "2026-8-5x")
	shown.Set("Lines-0-Time", "2026-13-45")
	shown.Set("__cmd__", mx.AddRowCommand("Lines"))
	added := submittedValues(post(shown).Body.String())
	for name, want := range map[string]string{"Lines-0-Loose": "2026-8-5", "Lines-0-Time": "2026-07-06"} {
		if got := added.Get(name); got != want {
			t.Errorf("after typo and add-row %s = %q, want the stored %q", name, got, want)
		}
	}
	added.Del("__cmd__")
	if rec = post(added); rec.Code != http.StatusSeeOther {
		t.Fatalf("unchanged Save: status %d, body %s", rec.Code, rec.Body.String())
	}
	got := saved.Lines[0]
	if got.Loose != (looseDate{2026, 8, 5}) || !got.Time.Equal(july) || !got.Null.Valid || !got.Null.Equal(july) || got.Ptr.Time == nil || !got.Ptr.Equal(july) {
		t.Errorf("saved row %+v, want the stored dates kept", got)
	}
}

// Numbers are checked against form:"min=…,max=…" on the server too, not
// only by the browser. A bound that is no number is ignored.
func TestFieldDecider_NumberRangeOnServer(t *testing.T) {
	type form struct {
		N int     `form:"min=1,max=10"`
		F float64 `form:"max=2.5"`
		P *int    `form:"min=0"`
		X int     `form:"max=lots"`
		B int64   `form:"max=9007199254740992"`
		G float64
		M float64 `form:"min=1"`
		H int     `form:"max=0x10,min=1/2"`
		R int     `form:"min=1,required"`
		O int     `form:"min=1"`
		D float64 `form:"min=0.3,max=99.9"`
		S float32 `form:"max=0.1"`
		E float64 `form:"min=1e400"`
		T float32 `form:"min=1e39"`
		I float64 `form:"max=+5,min=Inf"`
		Z int     `form:"sensitive,min=1000"`
		Y int     `form:"sensitive,max=9999"`
		U uint    `form:"min=1,max=10"`
		Q float32 `form:"min=0.7"`
	}
	f := &form{}
	for _, b := range []struct{ field, raw, wantErr string }{
		{"N", "9999", `"9999" is greater than 10`},
		{"N", "0", `"0" is less than 1`},
		{"N", "5", ""},
		{"N", "1", ""},
		{"N", "10", ""},
		// No bound holds NaN, so a non-finite number is never accepted.
		{"F", "NaN", `"NaN" is not a finite number`},
		{"G", "Inf", `"Inf" is not a finite number`},
		// Compared exactly, not as float64.
		{"B", "9007199254740993", `"9007199254740993" is greater than 9007199254740992`},
		// The stored value is compared, whatever text the parser read.
		{"M", "1e-1000001", `"1e-1000001" is less than 1`},
		{"G", strings.Repeat("1", 65), "too long for a number"},
		// A bound a browser would not read is no bound.
		{"H", "99", ""},
		{"F", "3", `"3" is greater than 2.5`},
		{"P", "", ""},
		{"P", "-1", `"-1" is less than 0`},
		{"X", "9999", ""},
		// A blank stores 0 in a number that is no pointer: a required one
		// rejects it, and the range is checked on it; a nil pointer is not.
		{"R", "", "required"},
		{"O", "", `"0" is less than 1`},
		// A sensitive number's range error names the bound, not the value.
		{"Z", "5", "must be at least 1000"},
		{"Y", "123456", "must be at most 9999"},
		// A float compares at its own precision, as a browser compares
		// doubles, so a value on its decimal bound is in range.
		{"D", "99.9", ""},
		{"D", "0.3", ""},
		{"D", "99.91", `"99.91" is greater than 99.9`},
		{"S", "0.1", ""},
		// A bound past the range of a double is no bound, as in a browser.
		{"E", "5", ""},
		{"I", "99", ""},
		{"T", "1", `"1" is less than 1e39`},
		// An unsigned number is held to its range like a signed one.
		{"U", "11", `"11" is greater than 10`},
		{"U", "0", `"0" is less than 1`},
		{"U", "10", ""},
		// A float32 min is read at float32 precision too: float32(0.7)
		// lies below the decimal 0.7, yet on its bound.
		{"Q", "0.7", ""},
		{"Q", "0.69", `"0.69" is less than 0.7`},
	} {
		err := parseFieldValue(t, f, b.field, b.raw)
		if got := fmt.Sprint(err); (b.wantErr == "" && err != nil) || (b.wantErr != "" && got != b.wantErr) {
			t.Errorf("bind %s %q: error %v, want %q", b.field, b.raw, err, b.wantErr)
		}
	}
	if f.N != 10 || f.P == nil || *f.P != -1 || f.H != 99 {
		t.Errorf("numbers must be bound, in range or not: %+v", f)
	}
}

// Text, number and bool fields bind through a pointer into a new pointee,
// never writing a pointee the field shares with a stored record.
func TestFieldDecider_ScalarBindKeepsSharedPointees(t *testing.T) {
	type form struct {
		S *string
		N *int
		B *bool
		C *[]colorEnum
	}
	storedS, storedN, storedB, storedC := "kept", 7, true, []colorEnum{"red"}
	f := &form{S: &storedS, N: &storedN, B: &storedB, C: &storedC}
	for field, raw := range map[string]string{"S": "posted", "N": "8", "B": "", "C": "blue"} {
		if err := parseFieldValue(t, f, field, raw); err != nil {
			t.Fatalf("bind %s: %v", field, err)
		}
	}
	if storedS != "kept" || storedN != 7 || !storedB || !reflect.DeepEqual(storedC, []colorEnum{"red"}) {
		t.Errorf("stored pointees changed: %q %d %t %v", storedS, storedN, storedB, storedC)
	}
	if *f.S != "posted" || *f.N != 8 || *f.B || !reflect.DeepEqual(*f.C, []colorEnum{"blue"}) {
		t.Errorf("the form must hold the bound values: %q %d %t %v", *f.S, *f.N, *f.B, *f.C)
	}
}

// nilTextDate embeds a pointer to a date struct; its promoted methods panic
// on a nil one.
type nilTextDate struct {
	*civilDate
}

// A half-set refusal names the type rather than an internal message.
func TestFieldDecider_HalfSetErrorNamesType(t *testing.T) {
	type form struct {
		At nullTime `form:"widget=datetime"`
	}
	err := parseFieldValue(t, &form{}, "At", "2026-08-05T10:30:00Z")
	if want := `"2026-08-05T10:30:00Z" cannot be stored as html.nullTime`; fmt.Sprint(err) != want {
		t.Errorf("error %v, want %s", err, want)
	}
}

// A type whose UnmarshalText zeroes itself on an error, as netip.Addr's
// does, keeps the stored value when a submission is rejected, at the top
// level and in a repeatable row through a typo and an add-row click.
func TestReflectFormHandler_RejectedTextKeepsStoredValue(t *testing.T) {
	type line struct {
		Addr netip.Addr
	}
	type record struct {
		Addr  netip.Addr
		Lines []line `form:"repeatable"`
	}
	stored := netip.MustParseAddr("10.0.0.1")
	if f := (&record{Addr: stored}); parseFieldValue(t, f, "Addr", "10.0.0.x") == nil || f.Addr != stored {
		t.Errorf("a rejected address changed the field: %v", f.Addr)
	}
	var saved *record
	handler := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{Addr: stored, Lines: []line{{stored}}}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	post := func(values url.Values) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/r", strings.NewReader(values.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		handler(rec, req)
		return rec
	}
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/r", nil))
	shown := submittedValues(rec.Body.String())
	shown.Set("Lines-0-Addr", "10.0.0.x")
	shown.Set("__cmd__", mx.AddRowCommand("Lines"))
	added := submittedValues(post(shown).Body.String())
	if got := added.Get("Lines-0-Addr"); got != "10.0.0.1" {
		t.Errorf("after a typo and add-row the row shows %q, want the stored 10.0.0.1", got)
	}
	added.Del("__cmd__")
	if rec = post(added); rec.Code != http.StatusSeeOther || saved == nil || saved.Lines[0].Addr != stored {
		t.Errorf("unchanged Save: status %d, saved %+v", rec.Code, saved)
	}
}

// A readonly repeatable field renders each cell readonly and offers no row
// buttons; a readonly file input is disabled and never required.
func TestFieldDecider_ReadonlyRowsAndFiles(t *testing.T) {
	type line struct {
		Desc string
	}
	type record struct {
		Lines []line `form:"repeatable,readonly"`
		Doc   []byte `form:"widget=file,required,readonly"`
	}
	handler := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{Lines: []line{{"a"}}}, nil },
		func(context.Context, *record) error { return nil },
		FieldDecider,
	)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/r", nil))
	body := rec.Body.String()
	if want := `<input type="text" name="Lines-0-Desc" id="Lines-0-Desc" value="a" readonly="readonly"/>`; !strings.Contains(body, want) {
		t.Errorf("GET lacks %s: %s", want, body)
	}
	if strings.Contains(body, "__cmd__") {
		t.Errorf("a readonly repeatable field offers row buttons: %s", body)
	}
	if want := `<input type="file" name="Doc" id="Doc" disabled="disabled"/>`; !strings.Contains(body, want) {
		t.Errorf("GET lacks %s: %s", want, body)
	}
}

// A sensitive number never shows its value, and an error a form shows
// carries no call stack.
func TestFieldDecider_SensitiveNumberAndPlainErrors(t *testing.T) {
	type form struct {
		PIN   int `form:"sensitive"`
		Plain struct{ N int }
	}
	f := &form{PIN: 1234}
	if out := renderFieldHTML(t, f, "PIN"); strings.Contains(out, "1234") {
		t.Errorf("a sensitive number shows its value: %s", out)
	}
	err := parseFieldValue(t, f, "Plain", "x")
	if want := `cannot parse value into struct { N int }: implement encoding.TextUnmarshaler or database/sql.Scanner, or use a form:"widget=..." override`; fmt.Sprint(err) != want {
		t.Errorf("error %q, want %q", fmt.Sprint(err), want)
	}
}

// A hidden time, such as an UpdatedAt kept for optimistic locking,
// round-trips unchanged, with its zone and fraction.
func TestReflectFormHandler_HiddenTimeRoundTrip(t *testing.T) {
	type record struct {
		UpdatedAt time.Time   `form:"hidden"`
		Defined   definedTime `form:"hidden"`
		Day       time.Time   `form:"hidden,widget=date,min=2030-01-01"`
		Ptr       *time.Time  `form:"hidden"`
		Nil       *time.Time  `form:"hidden"`
	}
	stamp := time.Date(2026, 7, 6, 9, 5, 7, 123456789, time.FixedZone("CEST", 2*60*60))
	var saved *record
	handler := mx.ReflectFormHandler(
		func(context.Context) (*record, error) {
			return &record{UpdatedAt: stamp, Defined: definedTime(stamp), Day: stamp, Ptr: new(stamp)}, nil
		},
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/r", nil))
	req := httptest.NewRequest(http.MethodPost, "/r", strings.NewReader(submittedValues(rec.Body.String()).Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusSeeOther || saved == nil || !saved.UpdatedAt.Equal(stamp) || !time.Time(saved.Defined).Equal(stamp) ||
		!saved.Day.Equal(stamp) || saved.Ptr == nil || !saved.Ptr.Equal(stamp) || saved.Nil != nil {
		t.Errorf("unchanged POST: status %d, saved %+v, body %s", rec.Code, saved, rec.Body.String())
	}
}

// A repeatable-tagged slice of times is skipped: never rendered or bound,
// so it neither blocks the save nor loses its stored times.
func TestReflectFormHandler_RepeatableTimesSkipped(t *testing.T) {
	type record struct {
		Name string
		Days []definedTime `form:"repeatable"`
	}
	day := definedTime(time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC))
	var saved *record
	handler := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{Name: "n", Days: []definedTime{day}}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/r", nil))
	if strings.Contains(rec.Body.String(), "Days") {
		t.Errorf("the skipped field renders: %s", rec.Body.String())
	}
	req := httptest.NewRequest(http.MethodPost, "/r", strings.NewReader(submittedValues(rec.Body.String()).Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusSeeOther || saved == nil || len(saved.Days) != 1 || saved.Days[0] != day {
		t.Errorf("unchanged POST: status %d, saved %+v", rec.Code, saved)
	}
}

// A blank sensitive field keeps a stored value, but with none stored a
// required one is still rejected, and unchecking every box of a
// sensitive checkbox set clears it rather than keeping the stored members.
func TestReflectFormHandler_SensitiveBlankEdges(t *testing.T) {
	type signup struct {
		Password string `form:"required,sensitive"`
	}
	saved := false
	h := mx.ReflectFormHandler(nil, func(context.Context, *signup) error { saved = true; return nil }, FieldDecider)
	if rec := postValues(t, h, url.Values{mx.PresentSentinelName("Password"): {"1"}, "Password": {""}}); rec.Code != http.StatusOK || saved {
		t.Errorf("a blank required sensitive field with nothing stored: status %d, saved %t", rec.Code, saved)
	}

	type record struct {
		Colors map[checkedColor]struct{} `form:"sensitive"`
	}
	var kept *record
	sets := mx.ReflectFormHandler(
		func(context.Context) (*record, error) {
			return &record{Colors: map[checkedColor]struct{}{"red": {}}}, nil
		},
		func(_ context.Context, r *record) error { kept = r; return nil },
		FieldDecider,
	)
	if rec := postValues(t, sets, url.Values{mx.PresentSentinelName("Colors"): {"1"}}); rec.Code != http.StatusSeeOther || kept == nil || len(kept.Colors) != 0 {
		t.Errorf("unchecking every box of a sensitive set: status %d, saved %+v", rec.Code, kept)
	}
}

// A sensitive date never shows its value, so an unchanged save, which
// posts it blank, keeps the stored date.
func TestReflectFormHandler_SensitiveDateKeepsStored(t *testing.T) {
	type record struct {
		Born string    `form:"widget=date,sensitive"`
		At   time.Time `form:"widget=date,sensitive"`
	}
	at := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{"2000-01-02", at}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	if body := get.Body.String(); strings.Contains(body, "2000-01-02") || strings.Contains(body, "2026-07-06") {
		t.Errorf("a sensitive date shows its value: %s", body)
	}
	if rec := postValues(t, h, submittedValues(get.Body.String())); rec.Code != http.StatusSeeOther || saved == nil || saved.Born != "2000-01-02" || !saved.At.Equal(at) {
		t.Errorf("unchanged save: status %d, saved %+v", rec.Code, saved)
	}
}

// A select shows its stored option, so choosing "(none)" on a sensitive
// one clears it rather than keeping the stored value.
func TestReflectFormHandler_SensitiveSelectNoneClears(t *testing.T) {
	type record struct {
		Color colorEnum `form:"sensitive"`
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{Color: "green"}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	if rec := postValues(t, h, url.Values{mx.PresentSentinelName("Color"): {"1"}, "Color": {""}}); rec.Code != http.StatusSeeOther || saved == nil || saved.Color != "" {
		t.Errorf("status %d, saved %+v; want (none) to clear", rec.Code, saved)
	}
}

// A blank sensitive input keeps the stored value, so a sensitive pointer
// gets the clear control to be set back to nil.
func TestReflectFormHandler_SensitivePointerClears(t *testing.T) {
	type record struct {
		Token *string        `form:"sensitive"`
		At    *time.Time     `form:"widget=date,sensitive"`
		N     *int           `form:"sensitive"`
		Note  *string        `form:"sensitive,widget=textarea"`
		Code  *sensitiveCode `form:"sensitive"`
		Plain *string
		Flag  *bool      `form:"sensitive"`
		Color *colorEnum `form:"sensitive"`
	}
	stored := func() *record {
		return &record{Token: new("secret"), At: new(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)), N: new(3), Note: new("n"), Code: &sensitiveCode{"c"}, Plain: new("p"), Flag: new(true), Color: new(colorEnum("red"))}
	}
	for field, want := range map[string]bool{"Token": true, "At": true, "N": true, "Note": true, "Code": true, "Plain": false, "Flag": false, "Color": false} {
		if got := strings.Contains(renderFieldHTML(t, stored(), field), mx.ClearSentinelName(mx.FieldPath(field))); got != want {
			t.Errorf("%s: clear control %t, want %t", field, got, want)
		}
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return stored(), nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	vals := url.Values{}
	for _, f := range []string{"Token", "At"} {
		vals.Set(mx.PresentSentinelName(mx.FieldPath(f)), "1")
		vals.Set(mx.ClearSentinelName(mx.FieldPath(f)), "1")
	}
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil || saved.Token != nil || saved.At != nil {
		t.Errorf("status %d, saved %+v; want both cleared", rec.Code, saved)
	}
}

// zonedStamp wraps a time.Time.
type zonedStamp struct{ time.Time }

// The input shows a stored time's wall clock without its zone, so an
// unchanged save keeps the stored time, also of a wrapper, and a new record
// reads UTC.
func TestReflectFormHandler_ZonedTimeKeepsInstant(t *testing.T) {
	type record struct {
		At    time.Time      `form:"widget=datetime"`
		Day   time.Time      `form:"widget=date"`
		Stamp zonedStamp     `form:"widget=datetime"`
		Ptr   ptrWrappedTime `form:"widget=datetime"`
		Null  sqlNullTime    `form:"widget=datetime"`
	}
	cest := time.FixedZone("CEST", 2*60*60)
	edt := time.FixedZone("EDT", -4*60*60)
	at := time.Date(2026, 6, 14, 10, 30, 0, 0, cest)
	day := time.Date(2026, 6, 14, 0, 0, 0, 0, edt)
	for name, loaded := range map[string]record{
		"zoned": {At: at, Day: day, Stamp: zonedStamp{at}, Ptr: ptrWrappedTime{new(at)}, Null: sqlNullTime{sql.NullTime{Time: at, Valid: true}}},
		"new":   {},
	} {
		var saved *record
		h := mx.ReflectFormHandler(
			func(context.Context) (*record, error) { r := loaded; return &r, nil },
			func(_ context.Context, r *record) error { saved = r; return nil },
			FieldDecider,
		)
		get := httptest.NewRecorder()
		h(get, httptest.NewRequest(http.MethodGet, "/", nil))
		vals := submittedValues(get.Body.String())
		if name == "new" {
			vals.Set("At", "2026-06-14T10:30")
		}
		rec := postValues(t, h, vals)
		if rec.Code != http.StatusSeeOther || saved == nil {
			t.Fatalf("%s: status %d, body %s", name, rec.Code, rec.Body.String())
		}
		switch name {
		case "zoned":
			if !saved.At.Equal(at) || !saved.Day.Equal(day) || !saved.Stamp.Equal(at) || saved.Ptr.Time == nil || !saved.Ptr.Equal(at) || !saved.Null.Time.Equal(at) {
				t.Errorf("zoned: saved %+v, want the instants of %v and %v", saved, at, day)
			}
		case "new":
			if want := time.Date(2026, 6, 14, 10, 30, 0, 0, time.UTC); !saved.At.Equal(want) {
				t.Errorf("new: saved %v, want %v", saved.At, want)
			}
		}
	}
}

// prefixedNumber is a number whose text carries a prefix.
type prefixedNumber int

func (n prefixedNumber) MarshalText() ([]byte, error) { return fmt.Appendf(nil, "id_%d", int(n)), nil }

func (n *prefixedNumber) UnmarshalText(text []byte) error {
	_, err := fmt.Sscanf(string(text), "id_%d", (*int)(n))
	return err
}

// utcStamp embeds a time but writes it in UTC, so the input shows the UTC
// wall clock whatever the stored location.
type utcStamp struct{ time.Time }

func (s utcStamp) MarshalText() ([]byte, error) {
	return []byte(s.Time.UTC().Format(time.RFC3339)), nil
}

func (s *utcStamp) UnmarshalText(text []byte) error {
	t, err := time.Parse(time.RFC3339, string(text))
	s.Time = t
	return err
}

// A changed time is read in the zone the input showed it in: a date at the
// first instant of its day, also where the clocks jump over midnight; a time
// on the stored date, in the stored zone; a wrapper's wall clock in the zone
// its text was shown in. An unchanged save keeps the stored time as it is,
// also in the hour repeated when the clocks go back.
func TestReflectFormHandler_ShownZone(t *testing.T) {
	// The test binary embeds the zone database (time/tzdata), so these load
	// on any machine.
	santiago, err1 := time.LoadLocation("America/Santiago")
	berlin, err2 := time.LoadLocation("Europe/Berlin")
	if err1 != nil || err2 != nil {
		t.Fatal(err1, err2)
	}
	type record struct {
		Day    time.Time   `form:"widget=date"`
		Clock  time.Time   `form:"widget=time"`
		First  time.Time   `form:"widget=datetime"`
		Second time.Time   `form:"widget=datetime"`
		Exact  time.Time   `form:"widget=datetime"`
		UTC    utcStamp    `form:"widget=datetime"`
		Null   sqlNullTime `form:"widget=datetime"`
	}
	// 02:30 on 2026-10-25 happens twice in Berlin, first CEST, then CET.
	first := time.Date(2026, 10, 25, 0, 30, 0, 0, time.UTC).In(berlin)
	second := time.Date(2026, 10, 25, 1, 30, 0, 0, time.UTC).In(berlin)
	loaded := record{
		Day:    time.Date(2026, 1, 15, 0, 0, 0, 0, santiago),
		Clock:  time.Date(2026, 6, 14, 9, 0, 0, 0, berlin),
		First:  first,
		Second: second,
		Exact:  time.Date(2026, 6, 14, 9, 0, 0, 500_000_000, berlin), // the input drops the fraction
		UTC:    utcStamp{time.Date(2026, 6, 14, 10, 0, 0, 0, berlin)},
		Null:   sqlNullTime{sql.NullTime{Time: time.Date(2026, 6, 14, 10, 0, 0, 0, time.FixedZone("CEST", 2*60*60)), Valid: true}},
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { r := loaded; return &r, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	vals := submittedValues(get.Body.String())
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil ||
		!saved.First.Equal(first) || !saved.Second.Equal(second) || !saved.Exact.Equal(loaded.Exact) || !saved.UTC.Equal(loaded.UTC.Time) {
		t.Fatalf("unchanged save: status %d, saved %+v", rec.Code, saved)
	}
	vals.Set("Day", "2026-09-06") // Santiago skips 00:00 that day
	vals.Set("Clock", "10:15")
	vals.Set("UTC", "2026-06-14T09:00")
	vals.Set("Null", "2026-06-14T09:00")
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil {
		t.Fatalf("changed save: status %d, body %s", rec.Code, rec.Body.String())
	}
	if got := saved.Day.In(santiago).Format(time.DateOnly); got != "2026-09-06" {
		t.Errorf("Day saved as %v, want 2026-09-06", saved.Day)
	}
	if want := time.Date(2026, 6, 14, 10, 15, 0, 0, berlin); !saved.Clock.Equal(want) {
		t.Errorf("Clock saved as %v, want %v", saved.Clock, want)
	}
	if want := time.Date(2026, 6, 14, 9, 0, 0, 0, time.UTC); !saved.UTC.Equal(want) {
		t.Errorf("UTC saved as %v, want %v", saved.UTC.Time, want)
	}
	if want := time.Date(2026, 6, 14, 7, 0, 0, 0, time.UTC); !saved.Null.Time.Equal(want) {
		t.Errorf("Null saved as %v, want 09:00 in the zone it was shown in, %v", saved.Null.Time, want)
	}
}

// A changed date or time in a named zone is read at that zone's offset for
// the new date, for a time.Time and a wrapper alike; a time widget keeps the
// stored date and zone when unchanged, and a clock in its DST gap moves
// forward. A typed sensitive time is read in the stored zone too.
func TestReflectFormHandler_NamedZoneAcrossDST(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	// A wrapper's text has only an offset, which time.Parse turns into the
	// named zone only when it matches time.Local; a southern zone matches
	// neither a European machine nor a UTC one.
	sydney, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		At     time.Time  `form:"widget=datetime"`
		Stamp  zonedStamp `form:"widget=datetime"`
		Far    zonedStamp `form:"widget=datetime"`
		Day    zonedStamp `form:"widget=date"`
		Clock  time.Time  `form:"widget=time"`
		Gap    time.Time  `form:"widget=time"`
		Secret time.Time  `form:"widget=datetime,sensitive"`
	}
	june := time.Date(2026, 6, 14, 10, 0, 0, 0, berlin)
	loaded := record{
		At: june, Stamp: zonedStamp{june}, Far: zonedStamp{time.Date(2026, 6, 14, 10, 0, 0, 0, sydney)},
		Day:   zonedStamp{time.Date(2026, 6, 14, 0, 0, 0, 0, berlin)},
		Clock: time.Date(2026, 6, 14, 9, 0, 0, 0, berlin), Gap: time.Date(2026, 3, 29, 9, 0, 0, 0, berlin), Secret: june,
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { r := loaded; return &r, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	vals := submittedValues(get.Body.String())
	vals.Set("At", "2026-12-14T10:00")
	vals.Set("Stamp", "2026-12-14T10:00")
	vals.Set("Far", "2026-12-14T10:00")
	vals.Set("Day", "2026-12-14")
	vals.Set("Gap", "02:30")
	vals.Set("Secret", "2026-06-14T09:00")
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
	}
	december := time.Date(2026, 12, 14, 10, 0, 0, 0, berlin)
	if !saved.At.Equal(december) || !saved.Stamp.Equal(december) {
		t.Errorf("At %v, Stamp %v, want %v", saved.At, saved.Stamp.Time, december)
	}
	if want := time.Date(2026, 12, 14, 10, 0, 0, 0, sydney); !saved.Far.Equal(want) {
		t.Errorf("Far saved as %v, want %v (summer time in Sydney)", saved.Far.Time, want)
	}
	if got := saved.Day.In(berlin).Format(time.DateOnly); got != "2026-12-14" {
		t.Errorf("Day saved as %v, want 2026-12-14 in Berlin", saved.Day.Time)
	}
	if !saved.Clock.Equal(loaded.Clock) || saved.Clock.Location() != berlin {
		t.Errorf("unchanged Clock saved as %v, want %v", saved.Clock, loaded.Clock)
	}
	if want := time.Date(2026, 3, 29, 3, 30, 0, 0, berlin); !saved.Gap.Equal(want) {
		t.Errorf("Gap saved as %v, want %v", saved.Gap, want)
	}
	if want := time.Date(2026, 6, 14, 9, 0, 0, 0, berlin); !saved.Secret.Equal(want) {
		t.Errorf("Secret saved as %v, want %v", saved.Secret, want)
	}
}

// A nil hidden *bool stays nil, and a file cell a loaded row renders in a
// section behind a pointer makes the form multipart.
func TestReflectFormHandler_HiddenNilBoolAndRowFile(t *testing.T) {
	type record struct {
		Flag *bool `form:"hidden"`
		Name string
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{Name: "x"}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec := postValues(t, h, submittedValues(get.Body.String())); rec.Code != http.StatusSeeOther || saved == nil || saved.Flag != nil {
		t.Errorf("status %d, saved %+v; want Flag nil", rec.Code, saved)
	}
	type attach struct {
		Doc []byte `form:"widget=file"`
	}
	type line struct {
		Desc   string
		Attach *attach `form:"section=Files"`
	}
	type rows struct {
		Lines []line `form:"repeatable"`
	}
	rh := mx.ReflectFormHandler(
		func(context.Context) (*rows, error) { return &rows{Lines: []line{{Desc: "a", Attach: &attach{}}}}, nil },
		func(context.Context, *rows) error { return nil },
		FieldDecider,
	)
	rget := httptest.NewRecorder()
	rh(rget, httptest.NewRequest(http.MethodGet, "/", nil))
	if body := rget.Body.String(); !strings.Contains(body, `type="file"`) || !strings.Contains(body, `enctype="multipart/form-data"`) {
		t.Errorf("a row's file input needs a multipart form: %s", body)
	}
}

// A database/sql type has no text methods: it renders the driver value it
// holds and is read back with Scan, hidden or as a catch-all text input.
// An unchanged submission keeps the stored value exactly, so a valid empty
// string stays valid and a time keeps its named zone.
func TestReflectFormHandler_SQLNullRoundTrip(t *testing.T) {
	type record struct {
		ID    sql.NullInt64   `form:"hidden"`
		Note  sql.NullString  `form:"hidden"`
		Gone  sql.NullString  `form:"hidden"`
		At    sql.NullTime    `form:"hidden"`
		Flag  sql.NullBool    `form:"hidden"`
		Rate  sql.NullFloat64 `form:"hidden"`
		Count sql.NullInt64
	}
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	stored := record{
		ID:   sql.NullInt64{Int64: 42, Valid: true},
		Note: sql.NullString{Valid: true},
		// The first 02:30 of the hour Berlin's clocks repeat when they go back.
		At:    sql.NullTime{Time: time.Date(2026, 10, 25, 0, 30, 0, 0, time.UTC).In(berlin), Valid: true},
		Flag:  sql.NullBool{Bool: true, Valid: true},
		Rate:  sql.NullFloat64{Float64: 1.5, Valid: true},
		Count: sql.NullInt64{Int64: 7, Valid: true},
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { r := stored; return &r, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	body := get.Body.String()
	for _, want := range []string{
		`name="ID" value="42"`,
		`name="At" value="2026-10-25T02:30:00+02:00"`,
		`name="Flag" value="true"`,
		`name="Rate" value="1.5"`,
		`name="Count" id="Count" value="7"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("want %s in %s", want, body)
		}
	}

	vals := submittedValues(body)
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil {
		t.Fatalf("unchanged: status %d: %s", rec.Code, rec.Body)
	}
	if saved.ID != stored.ID || saved.Note != stored.Note || saved.Gone != stored.Gone ||
		saved.Flag != stored.Flag || saved.Rate != stored.Rate || saved.Count != stored.Count {
		t.Errorf("unchanged submission changed the record: %+v", saved)
	}
	// Read back from its text, the time would be in a fixed +02:00 zone.
	if saved.At.Time.Location() != berlin || !saved.At.Time.Equal(stored.At.Time) {
		t.Errorf("At = %v, want %v in Europe/Berlin", saved.At.Time, stored.At.Time)
	}

	saved = nil
	vals.Set("ID", "43")
	vals.Set("Gone", "back")
	vals.Set("At", "2026-01-02T03:04:05Z")
	vals.Set("Flag", "false")
	vals.Set("Count", "")
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil {
		t.Fatalf("changed: status %d: %s", rec.Code, rec.Body)
	}
	if saved.ID != (sql.NullInt64{Int64: 43, Valid: true}) || saved.Gone != (sql.NullString{String: "back", Valid: true}) ||
		saved.Flag != (sql.NullBool{Valid: true}) || saved.Count.Valid {
		t.Errorf("changed submission: %+v", saved)
	}
	if want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC); !saved.At.Valid || !saved.At.Time.Equal(want) {
		t.Errorf("At = %+v, want %v", saved.At, want)
	}

	saved = nil
	vals.Set("ID", "abc")
	if rec := postValues(t, h, vals); rec.Code == http.StatusSeeOther || saved != nil {
		t.Errorf("text that does not scan must not save: status %d", rec.Code)
	}
}

// centsMoney stores cents but reads and shows major units with its own
// methods, so its driver value is not its text.
type centsMoney struct{ cents int64 }

func (m centsMoney) Value() (driver.Value, error) { return m.cents, nil }
func (m *centsMoney) Scan(src any) error          { m.cents, _ = src.(int64); return nil }
func (m centsMoney) String() string               { return fmt.Sprintf("%d.%02d", m.cents/100, m.cents%100) }

func (m *centsMoney) UnmarshalText(text []byte) error {
	major, minor, _ := strings.Cut(string(text), ".")
	units, err := strconv.ParseInt(major+minor, 10, 64)
	m.cents = units
	return err
}

// jsonDoc is a JSON column type whose Scan takes []byte only, as many do.
type jsonDoc struct {
	A string
	B int
}

func (d jsonDoc) Value() (driver.Value, error) {
	return json.Marshal(struct {
		A string
		B int
	}(d))
}

func (d *jsonDoc) Scan(src any) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}
	return json.Unmarshal(b, (*struct {
		A string
		B int
	})(d))
}

// prefixedCode embeds sql.NullString but writes and reads its own text.
type prefixedCode struct{ sql.NullString }

func (c prefixedCode) MarshalText() ([]byte, error) {
	if !c.Valid {
		return nil, nil
	}
	return []byte("code:" + c.String), nil
}

func (c *prefixedCode) UnmarshalText(text []byte) error {
	v, ok := strings.CutPrefix(string(text), "code:")
	if !ok {
		return errors.New("missing code: prefix")
	}
	c.String, c.Valid = v, true
	return nil
}

// A database/sql number renders as a text input without min or max, so the
// server is all that holds it to its range, also behind a pointer and for
// a float.
func TestFieldDecider_ScannedRangePointerAndFloat(t *testing.T) {
	type form struct {
		Ptr  *sql.NullInt64  `form:"min=0,max=10"`
		Rate sql.NullFloat64 `form:"min=0,max=1.5"`
	}
	f := &form{}
	for _, b := range []struct{ field, raw, wantErr string }{
		{"Ptr", "11", `"11" is greater than 10`},
		{"Ptr", "-1", `"-1" is less than 0`},
		{"Ptr", "7", ""},
		{"Rate", "1.6", `"1.6" is greater than 1.5`},
		{"Rate", "-0.1", `"-0.1" is less than 0`},
		{"Rate", "1.5", ""},
	} {
		err := parseFieldValue(t, f, b.field, b.raw)
		if got := fmt.Sprint(err); (b.wantErr == "" && err != nil) || (b.wantErr != "" && got != b.wantErr) {
			t.Errorf("bind %s %q: error %v, want %q", b.field, b.raw, err, b.wantErr)
		}
	}
}

// A database/sql type is held to the rules of what it holds: a number to
// those of a number field, text it must scan as []byte to that, and an
// error names what to enter. A type with its own text methods or String,
// whose driver value is not its text, keeps using them.
func TestReflectFormHandler_SQLNullRules(t *testing.T) {
	type record struct {
		Price centsMoney `form:"hidden"`
		Rate  sql.NullFloat64
		Count sql.NullInt64 `form:"min=0,max=10"`
		Doc   jsonDoc
		At    sql.NullTime
		Flag  sql.NullBool
		Code  prefixedCode `form:"hidden"`
	}
	stored := record{
		Price: centsMoney{1250},
		Rate:  sql.NullFloat64{Float64: 1.5, Valid: true},
		Count: sql.NullInt64{Int64: 3, Valid: true},
		Doc:   jsonDoc{A: "x", B: 1},
		At:    sql.NullTime{Time: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), Valid: true},
		Flag:  sql.NullBool{Bool: true, Valid: true},
		Code:  prefixedCode{sql.NullString{String: "x", Valid: true}},
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { r := stored; return &r, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	body := get.Body.String()
	for _, want := range []string{`name="Price" value="12.50"`, `name="Code" value="code:x"`, `name="Doc" id="Doc" value="{&quot;A&quot;:&quot;x&quot;,&quot;B&quot;:1}"`} {
		if !strings.Contains(body, want) {
			t.Errorf("want %s in %s", want, body)
		}
	}
	base := submittedValues(body)
	if rec := postValues(t, h, base); rec.Code != http.StatusSeeOther || saved == nil || *saved != stored {
		t.Fatalf("unchanged save: status %d, saved %+v", rec.Code, saved)
	}
	post := func(field, raw string) (*httptest.ResponseRecorder, *record) {
		t.Helper()
		saved = nil
		vals := url.Values{}
		for k, v := range base {
			vals[k] = v
		}
		vals.Set(field, raw)
		return postValues(t, h, vals), saved
	}
	for _, c := range []struct{ field, raw, msg string }{
		{"Rate", "NaN", `&quot;NaN&quot; is not a finite number`},
		{"Rate", "-Inf", `&quot;-Inf&quot; is not a finite number`},
		{"Rate", strings.Repeat("1", maxNumberLen+1), "too long for a number"},
		{"Count", "50", "10"},
		{"Count", "abc", `&quot;abc&quot; is not a whole number`},
		{"At", "nope", "is not a time in RFC 3339 format"},
		{"Flag", "yes", `&quot;yes&quot; is not true or false`},
	} {
		if rec, got := post(c.field, c.raw); rec.Code == http.StatusSeeOther || got != nil || !strings.Contains(rec.Body.String(), c.msg) {
			t.Errorf("%s=%q: status %d, saved %v; want rejected with %q in %s", c.field, c.raw, rec.Code, got != nil, c.msg, rec.Body.String())
		}
	}
	if rec, got := post("Count", "7"); rec.Code != http.StatusSeeOther || got.Count != (sql.NullInt64{Int64: 7, Valid: true}) {
		t.Errorf("Count=7: status %d", rec.Code)
	}
	if rec, got := post("Doc", `{"A":"y","B":2}`); rec.Code != http.StatusSeeOther || got.Doc != (jsonDoc{A: "y", B: 2}) {
		t.Errorf("Doc edit: status %d, %s", rec.Code, rec.Body.String())
	}
	if rec, got := post("Price", "13.75"); rec.Code != http.StatusSeeOther || got.Price.cents != 1375 {
		t.Errorf("Price edit: status %d", rec.Code)
	}
	if rec, got := post("Code", "code:y"); rec.Code != http.StatusSeeOther || got.Code.String != "y" {
		t.Errorf("Code edit: status %d", rec.Code)
	}
}

// A sensitive database/sql value keeps what is stored on a blank save and
// is emptied by its clear box, also a valid empty string, which shows as
// blank as well.
func TestReflectFormHandler_SensitiveScannerCells(t *testing.T) {
	type record struct {
		N sql.NullInt64  `form:"sensitive"`
		S sql.NullString `form:"sensitive"`
	}
	stored := record{N: sql.NullInt64{Int64: 42, Valid: true}, S: sql.NullString{Valid: true}}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { r := stored; return &r, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	vals := submittedValues(get.Body.String())
	vals.Del(mx.ClearSentinelName("N"))
	vals.Del(mx.ClearSentinelName("S"))
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil || *saved != stored {
		t.Fatalf("blank save: status %d, saved %+v", rec.Code, saved)
	}
	vals.Set(mx.ClearSentinelName("N"), "1")
	vals.Set(mx.ClearSentinelName("S"), "1")
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved.N.Valid || saved.S.Valid {
		t.Errorf("ticked save: status %d, saved %+v; want both null", rec.Code, saved)
	}
}

// A blank required *sql.NullString is "required", not a crash: its nil
// pointer is null without calling Value through it.
func TestReflectFormHandler_RequiredNilScannerPointer(t *testing.T) {
	type record struct {
		Name *sql.NullString `form:"required"`
	}
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{}, nil },
		func(context.Context, *record) error { return nil },
		FieldDecider,
	)
	vals := url.Values{mx.PresentSentinelName("Name"): {"1"}, "Name": {""}}
	if rec := postValues(t, h, vals); rec.Code == http.StatusSeeOther || !strings.Contains(rec.Body.String(), "required") {
		t.Errorf("status %d: %s", rec.Code, rec.Body.String())
	}
}

// secStamp writes its time in whole seconds, as a database timestamp type
// may, while holding a fraction.
type secStamp struct{ time.Time }

func (s secStamp) MarshalText() ([]byte, error) { return []byte(s.Format(time.RFC3339)), nil }

func (s *secStamp) UnmarshalText(text []byte) error {
	t, err := time.Parse(time.RFC3339, string(text))
	s.Time = t
	return err
}

// privStamp keeps its time in an unexported field.
type privStamp struct{ t time.Time }

func (s privStamp) MarshalText() ([]byte, error) { return []byte(s.t.Format(time.RFC3339)), nil }

func (s *privStamp) UnmarshalText(text []byte) error {
	t, err := time.Parse(time.RFC3339, string(text))
	s.t = t
	return err
}

// A wrapper's named zone is found through pointers and nested wrappers,
// and through text that drops the wrapped time's fraction, so a changed
// date takes that date's offset. A wrapper whose time sits in an
// unexported field renders and saves with its text's offset.
func TestReflectFormHandler_NamedZoneNestedWrappers(t *testing.T) {
	sydney, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		Null sqlNullTime    `form:"widget=datetime"`
		Ptr  ptrWrappedTime `form:"widget=datetime"`
		Sec  secStamp       `form:"widget=datetime"`
		Priv privStamp      `form:"widget=datetime"`
	}
	june := time.Date(2026, 6, 14, 10, 0, 0, 0, sydney)
	loaded := record{
		Null: sqlNullTime{sql.NullTime{Time: june, Valid: true}},
		Ptr:  ptrWrappedTime{new(june)},
		Sec:  secStamp{june.Add(123456 * time.Microsecond)},
		Priv: privStamp{june},
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) {
			r := loaded
			r.Ptr = ptrWrappedTime{new(*loaded.Ptr.Time)}
			return &r, nil
		},
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	if get.Code != http.StatusOK {
		t.Fatalf("GET status %d: %s", get.Code, get.Body)
	}
	vals := submittedValues(get.Body.String())
	for _, name := range []string{"Null", "Ptr", "Sec"} {
		vals.Set(name, "2026-12-14T10:00")
	}
	vals.Set("Priv", "2026-06-14T11:00")
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
	}
	december := time.Date(2026, 12, 14, 10, 0, 0, 0, sydney) // summer time, +11:00
	if !saved.Null.Time.Equal(december) || saved.Ptr.Time == nil || !saved.Ptr.Time.Equal(december) || !saved.Sec.Equal(december) {
		t.Errorf("Null %v, Ptr %v, Sec %v; want %v", saved.Null.Time, saved.Ptr.Time, saved.Sec.Time, december)
	}
	if want := time.Date(2026, 6, 14, 11, 0, 0, 0, sydney); !saved.Priv.t.Equal(want) {
		t.Errorf("Priv %v, want %v", saved.Priv.t, want)
	}
}

// A time input's range compares the clock, also on a record whose stored
// time puts a posted time on its date.
func TestReflectFormHandler_TimeRangeOnStoredRecord(t *testing.T) {
	type record struct {
		From  time.Time `form:"widget=time,min=09:00"`
		Shift time.Time `form:"widget=time,min=09:00,max=17:00"`
		Night time.Time `form:"widget=time,min=22:00,max=06:00"`
	}
	stored := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) {
			return &record{stored, stored, stored.Add(11 * time.Hour)}, nil
		},
		func(context.Context, *record) error { return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	base := submittedValues(get.Body.String())
	for _, c := range []struct {
		field, posted string
		ok            bool
	}{
		{"From", "08:00", false},
		{"From", "10:00", true},
		{"Shift", "11:00", true},
		{"Shift", "18:00", false},
		{"Night", "23:00", true},
		{"Night", "12:00", false},
	} {
		vals := url.Values{}
		for k, v := range base {
			vals[k] = v
		}
		vals.Set(c.field, c.posted)
		rec := postValues(t, h, vals)
		if got := rec.Code == http.StatusSeeOther; got != c.ok {
			t.Errorf("%s=%s: status %d, want accepted %t", c.field, c.posted, rec.Code, c.ok)
		}
	}
}

// A time posted with an offset keeps its instant through a re-render, and a
// zoned row time keeps its instant through an add-row click.
func TestReflectFormHandler_ZonesThroughRerender(t *testing.T) {
	type line struct {
		At time.Time `form:"widget=datetime"`
	}
	type record struct {
		At    time.Time `form:"widget=datetime"`
		Lines []line    `form:"repeatable"`
	}
	cest := time.FixedZone("CEST", 2*60*60)
	stored := time.Date(2026, 6, 14, 10, 30, 0, 0, cest)
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{At: stored, Lines: []line{{stored}}}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	vals := submittedValues(get.Body.String())
	vals.Set("At", "2026-06-14T10:30:00Z")
	vals.Set("__cmd__", mx.AddRowCommand("Lines"))
	added := postValues(t, h, vals)
	want := time.Date(2026, 6, 14, 10, 30, 0, 0, time.UTC)
	if rec := postValues(t, h, submittedValues(added.Body.String())); rec.Code != http.StatusSeeOther || saved == nil || !saved.At.Equal(want) || !saved.Lines[0].At.Equal(stored) {
		t.Errorf("status %d, saved %+v, want At %v and the row at %v", rec.Code, saved, want, stored)
	}
}

// A posted time with an explicit offset keeps its instant, shown in the
// stored time's location.
func TestReflectFormHandler_PostedOffsetWins(t *testing.T) {
	type record struct {
		At time.Time `form:"widget=datetime"`
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) {
			return &record{At: time.Date(2026, 6, 14, 10, 30, 0, 0, time.FixedZone("CEST", 2*60*60))}, nil
		},
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	want := time.Date(2026, 6, 14, 10, 30, 0, 0, time.UTC)
	if rec := postValues(t, h, url.Values{mx.PresentSentinelName("At"): {"1"}, "At": {"2026-06-14T10:30:00Z"}}); rec.Code != http.StatusSeeOther || saved == nil || !saved.At.Equal(want) {
		t.Errorf("status %d, saved %+v, want %v", rec.Code, saved, want)
	}
}

// A hidden number, such as a per-row key, round-trips as a number, and one
// with its own text form round-trips through it.
func TestReflectFormHandler_HiddenNumberRoundTrip(t *testing.T) {
	type line struct {
		ID   int `form:"hidden"`
		Desc string
	}
	type record struct {
		Version int            `form:"hidden"`
		Ref     prefixedNumber `form:"hidden"`
		Lines   []line         `form:"repeatable"`
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) {
			return &record{Version: 7, Ref: 5, Lines: []line{{42, "a"}}}, nil
		},
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(get.Body.String(), `value="id_5"`) {
		t.Errorf("a hidden number with its own text form renders %s", get.Body.String())
	}
	if rec := postValues(t, h, submittedValues(get.Body.String())); rec.Code != http.StatusSeeOther || saved == nil || saved.Version != 7 || saved.Ref != 5 || saved.Lines[0].ID != 42 {
		t.Errorf("status %d, saved %+v, body %s", rec.Code, saved, rec.Body.String())
	}
}

// A hidden bool round-trips as a checkbox value, at the top level and in a
// row through an add-row click.
func TestReflectFormHandler_HiddenBoolRoundTrip(t *testing.T) {
	type line struct {
		Done bool `form:"hidden"`
		Desc string
	}
	type record struct {
		Confirmed bool   `form:"hidden"`
		Lines     []line `form:"repeatable"`
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) {
			return &record{Confirmed: true, Lines: []line{{true, "a"}}}, nil
		},
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	vals := submittedValues(get.Body.String())
	vals.Set("__cmd__", mx.AddRowCommand("Lines"))
	added := postValues(t, h, vals)
	if rec := postValues(t, h, submittedValues(added.Body.String())); rec.Code != http.StatusSeeOther || saved == nil || !saved.Confirmed || !saved.Lines[0].Done || saved.Lines[1].Done {
		t.Errorf("status %d, saved %+v, body %s", rec.Code, saved, rec.Body.String())
	}
}

// A hidden number that is no number fails the save with a form-level
// message, and a row's hidden key follows its row through a remove-row
// click.
func TestReflectFormHandler_HiddenNumberEdges(t *testing.T) {
	type line struct {
		ID   int `form:"hidden"`
		Desc string
	}
	type record struct {
		Version int    `form:"hidden"`
		Lines   []line `form:"repeatable"`
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) {
			return &record{Version: 7, Lines: []line{{41, "a"}, {42, "b"}}}, nil
		},
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	tampered := submittedValues(get.Body.String())
	tampered.Set("Version", "7x")
	if rec := postValues(t, h, tampered); rec.Code != http.StatusOK || saved != nil || !strings.Contains(rec.Body.String(), `<p role="alert">`) {
		t.Errorf("a hidden number that is no number: status %d, saved %+v", rec.Code, saved)
	}
	vals := submittedValues(get.Body.String())
	vals.Set("__cmd__", mx.DeleteRowCommand("Lines-0"))
	removed := postValues(t, h, vals)
	if rec := postValues(t, h, submittedValues(removed.Body.String())); rec.Code != http.StatusSeeOther || saved == nil || len(saved.Lines) != 1 || saved.Lines[0].ID != 42 {
		t.Errorf("the row key must follow its row: status %d, saved %+v", rec.Code, saved)
	}
}

// With nothing stored, a blank sensitive number has nothing to keep, so a
// required one is rejected and the range is checked, also in an added row.
func TestReflectFormHandler_SensitiveNumberNothingStored(t *testing.T) {
	type line struct {
		PIN int `form:"sensitive,min=1000"`
	}
	type record struct {
		PIN   int    `form:"sensitive,required"`
		Lines []line `form:"repeatable"`
	}
	saved := false
	h := mx.ReflectFormHandler(nil, func(context.Context, *record) error { saved = true; return nil }, FieldDecider)
	top := url.Values{mx.PresentSentinelName("PIN"): {"1"}, "PIN": {""}}
	row := url.Values{
		mx.PresentSentinelName("Lines"):       {"1"},
		mx.RowSentinelName("Lines-0"):         {"new"},
		mx.PresentSentinelName("Lines-0-PIN"): {"1"},
		"Lines-0-PIN":                         {""},
		mx.PresentSentinelName("PIN"):         {"1"},
		"PIN":                                 {"1234"},
	}
	for name, vals := range map[string]url.Values{"required": top, "row range": row} {
		if rec := postValues(t, h, vals); rec.Code != http.StatusOK || saved {
			t.Errorf("%s: status %d, saved %t", name, rec.Code, saved)
		}
	}

	// The same as the browser posts it: the clear box of an empty optional
	// input is ticked, and a tick parses the blank, so the range applies,
	// also in a row added with add-row; text typed over the tick wins.
	type optional struct {
		PIN   int    `form:"sensitive,min=1000"`
		Lines []line `form:"repeatable"`
	}
	var kept *optional
	oh := mx.ReflectFormHandler(nil, func(_ context.Context, r *optional) error { kept = r; return nil }, FieldDecider)
	get := httptest.NewRecorder()
	oh(get, httptest.NewRequest(http.MethodGet, "/", nil))
	vals := submittedValues(get.Body.String())
	if vals.Get(mx.ClearSentinelName("PIN")) != "1" {
		t.Fatalf("the clear box of an empty sensitive number must be ticked: %v", vals)
	}
	vals.Set("__cmd__", mx.AddRowCommand("Lines"))
	added := submittedValues(postValues(t, oh, vals).Body.String())
	if rec := postValues(t, oh, added); rec.Code != http.StatusOK || kept != nil {
		t.Errorf("a ticked blank below min saved: status %d, saved %+v", rec.Code, kept)
	}
	added.Set("PIN", "5")
	added.Set("Lines-0-PIN", "1500")
	if rec := postValues(t, oh, added); rec.Code != http.StatusOK || kept != nil {
		t.Errorf("5 over the tick: status %d, saved %+v", rec.Code, kept)
	}
	added.Set("PIN", "1500")
	if rec := postValues(t, oh, added); rec.Code != http.StatusSeeOther || kept == nil || kept.PIN != 1500 || kept.Lines[0].PIN != 1500 {
		t.Errorf("1500 over the tick: status %d, saved %+v", rec.Code, kept)
	}
}

// A sensitive row cell keeps, replaces or clears only its own row's
// stored value.
func TestReflectFormHandler_SensitiveRowCellActions(t *testing.T) {
	type line struct {
		Token string `form:"sensitive"`
	}
	type record struct {
		Lines []line `form:"repeatable"`
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{Lines: []line{{"a"}, {"b"}}}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	for _, c := range []struct {
		name, typed string
		tick        bool
		want        string
	}{{"blank", "", false, "a"}, {"typed", "new", false, "new"}, {"tick", "", true, ""}, {"typed+tick", "new", true, "new"}} {
		saved = nil
		vals := submittedValues(get.Body.String())
		vals.Set("Lines-0-Token", c.typed)
		if c.tick {
			vals.Set(mx.ClearSentinelName("Lines-0-Token"), "1")
		}
		if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil || saved.Lines[0].Token != c.want || saved.Lines[1].Token != "b" {
			t.Errorf("%s: status %d, saved %+v; want row 0 %q and row 1 kept", c.name, rec.Code, saved, c.want)
		}
	}
}

// A stored sensitive value renders no value, and a blank keeps it, so the
// input does not demand one from the browser while a value is stored; with
// nothing stored it does, and the server still rejects the blank.
func TestReflectFormHandler_RequiredSensitiveStored(t *testing.T) {
	type record struct {
		Key string `form:"required,sensitive"`
		PW  string `form:"required,widget=password"`
	}
	stored := &record{Key: "sk_1", PW: "hunter2"}
	for _, field := range []string{"Key", "PW"} {
		if got := inputTag(t, renderFieldHTML(t, stored, field), field); strings.Contains(got, "required") {
			t.Errorf("%s: a stored sensitive input demands a value: %s", field, got)
		}
		if got := inputTag(t, renderFieldHTML(t, &record{}, field), field); !strings.Contains(got, "required") {
			t.Errorf("%s: with nothing stored the input must be required: %s", field, got)
		}
	}
	if out := renderFieldHTML(t, stored, "Key"); strings.Contains(out, mx.ClearSentinelName("Key")) {
		t.Errorf("a required sensitive field offers a clear control: %s", out)
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { r := *stored; return &r, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec := postValues(t, h, submittedValues(get.Body.String())); rec.Code != http.StatusSeeOther || saved == nil || *saved != *stored {
		t.Errorf("unchanged save: status %d, saved %+v", rec.Code, saved)
	}
}

// inputTag returns the <input> or <textarea> start tag named name in out.
func inputTag(t *testing.T, out, name string) string {
	t.Helper()
	i := strings.Index(out, `name="`+name+`"`)
	if i < 0 {
		t.Fatalf("no input %s in %s", name, out)
	}
	start := strings.LastIndex(out[:i], "<")
	return out[start : i+strings.Index(out[i:], ">")+1]
}

// An optional sensitive input keeps its stored value when blank, so the
// clear control empties it, also when it is no pointer; it is ticked while
// the value is empty, so a clear survives a re-render, and text typed into
// the input wins over the tick.
func TestReflectFormHandler_SensitiveClear(t *testing.T) {
	type line struct{ Desc string }
	type record struct {
		Secret string  `form:"sensitive"`
		Token  *string `form:"sensitive"`
		Lines  []line  `form:"repeatable"`
	}
	box := func(out, name string) string {
		i := strings.Index(out, `name="`+mx.ClearSentinelName(mx.FieldPath(name))+`"`)
		if i < 0 {
			return ""
		}
		start := strings.LastIndex(out[:i], "<")
		return out[start : i+strings.Index(out[i:], ">")+1]
	}
	stored := func() *record { return &record{Secret: "s", Token: new("t")} }
	if out := renderFieldHTML(t, stored(), "Secret"); box(out, "Secret") == "" || strings.Contains(box(out, "Secret"), "checked") || !strings.Contains(box(out, "Secret"), `aria-label="clear Secret"`) {
		t.Errorf("a stored optional sensitive string needs an unticked clear control named after it: %s", out)
	}
	if out := renderFieldHTML(t, &record{}, "Secret"); !strings.Contains(box(out, "Secret"), "checked") {
		t.Errorf("an empty sensitive string's clear control must be ticked: %s", out)
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return stored(), nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))

	// Clear both, then click add-row: the re-render keeps the clear.
	vals := submittedValues(get.Body.String())
	vals.Set(mx.ClearSentinelName("Secret"), "1")
	vals.Set(mx.ClearSentinelName("Token"), "1")
	vals.Set("__cmd__", mx.AddRowCommand("Lines"))
	added := postValues(t, h, vals)
	if rec := postValues(t, h, submittedValues(added.Body.String())); rec.Code != http.StatusSeeOther || saved == nil || saved.Secret != "" || saved.Token != nil {
		t.Errorf("a clear across a re-render: status %d, saved %+v", rec.Code, saved)
	}

	// Text typed into the input wins over the tick.
	vals = submittedValues(get.Body.String())
	vals.Set(mx.ClearSentinelName("Secret"), "1")
	vals.Set("Secret", "new")
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil || saved.Secret != "new" || saved.Token == nil || *saved.Token != "t" {
		t.Errorf("typed text with the tick: status %d, saved %+v", rec.Code, saved)
	}

	// It wins on a nullable input too, where the clear itself would
	// succeed and drop the typed text.
	saved = nil
	vals = submittedValues(get.Body.String())
	vals.Set(mx.ClearSentinelName("Token"), "1")
	vals.Set("Token", "new")
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil || saved.Token == nil || *saved.Token != "new" {
		t.Errorf("typed text with the tick on a pointer: status %d, saved %+v", rec.Code, saved)
	}

	// With nothing stored the tick is rendered set, so a first value the
	// user types is always posted with it and must be saved.
	saved = nil
	empty := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	first := httptest.NewRecorder()
	empty(first, httptest.NewRequest(http.MethodGet, "/", nil))
	vals = submittedValues(first.Body.String())
	if vals.Get(mx.ClearSentinelName("Token")) == "" {
		t.Fatalf("an empty sensitive pointer must render its clear control ticked: %s", first.Body.String())
	}
	vals.Set("Token", "first")
	if rec := postValues(t, empty, vals); rec.Code != http.StatusSeeOther || saved == nil || saved.Token == nil || *saved.Token != "first" {
		t.Errorf("a first value typed next to the preset tick: status %d, saved %+v", rec.Code, saved)
	}
}

// An upload larger than the handler's MaxMemory is a field error rather
// than read into memory; a smaller one binds.
func TestReflectFormHandler_UploadLimit(t *testing.T) {
	type record struct {
		Doc []byte `form:"widget=file"`
	}
	var saved *record
	h := mx.ReflectFormHandlerWith(mx.ReflectFormConfig{MaxMemory: 64},
		func(context.Context) (*record, error) { return &record{}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	for _, c := range []struct {
		size int
		want int
	}{{100, http.StatusOK}, {10, http.StatusSeeOther}} {
		saved = nil
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		_ = mw.WriteField(mx.PresentSentinelName("Doc"), "1")
		part, _ := mw.CreateFormFile("Doc", "doc.bin")
		_, _ = part.Write(bytes.Repeat([]byte("xy"), c.size)[:c.size])
		_ = mw.Close()
		req := httptest.NewRequest(http.MethodPost, "/", &body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		rec := httptest.NewRecorder()
		h(rec, req)
		if rec.Code != c.want || (c.want == http.StatusSeeOther && (saved == nil || !bytes.Equal(saved.Doc, bytes.Repeat([]byte("xy"), c.size)[:c.size]))) {
			t.Errorf("%d-byte upload: status %d, saved %v, body %s", c.size, rec.Code, saved, rec.Body.String())
		}
	}
}

// A file field of type string takes the uploaded file's name, and one of
// another type is a field error rather than a silent no-op.
func TestReflectFormHandler_UploadDestinations(t *testing.T) {
	post := func(h http.HandlerFunc) *httptest.ResponseRecorder {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		_ = mw.WriteField(mx.PresentSentinelName("Doc"), "1")
		part, _ := mw.CreateFormFile("Doc", "report.pdf")
		_, _ = part.Write([]byte("%PDF"))
		_ = mw.Close()
		req := httptest.NewRequest(http.MethodPost, "/", &body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		rec := httptest.NewRecorder()
		h(rec, req)
		return rec
	}
	type named struct {
		Doc string `form:"widget=file"`
	}
	var saved *named
	if rec := post(mx.ReflectFormHandler(func(context.Context) (*named, error) { return &named{}, nil },
		func(_ context.Context, r *named) error { saved = r; return nil }, FieldDecider)); rec.Code != http.StatusSeeOther || saved == nil || saved.Doc != "report.pdf" {
		t.Errorf("string file field: status %d, saved %+v", rec.Code, saved)
	}
	type numbered struct {
		Doc int `form:"widget=file"`
	}
	if rec := post(mx.ReflectFormHandler(func(context.Context) (*numbered, error) { return &numbered{}, nil },
		func(context.Context, *numbered) error { return nil }, FieldDecider)); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "file upload requires") {
		t.Errorf("int file field: status %d, body %s", rec.Code, rec.Body.String())
	}
}

// keyCode rejects text without its prefix, wrapping a FormMessage in an
// error that quotes the text.
type keyCode struct{ v string }

func (c keyCode) MarshalText() ([]byte, error) { return []byte(c.v), nil }

func (c *keyCode) UnmarshalText(b []byte) error {
	if !strings.HasPrefix(string(b), "sk_") {
		return fmt.Errorf("%q: %w", b, mx.FormMessage("must start with sk_"))
	}
	c.v = string(b)
	return nil
}

// A sensitive field shows only the FormMessage of a parse error, not the
// error that wraps it and quotes the posted value.
func TestReflectFormHandler_SensitiveWrappedMessage(t *testing.T) {
	type record struct {
		Key keyCode `form:"sensitive"`
	}
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{}, nil },
		func(context.Context, *record) error { return nil },
		FieldDecider,
	)
	body := postValues(t, h, url.Values{mx.PresentSentinelName("Key"): {"1"}, "Key": {"leaked-SECRET"}}).Body.String()
	if !strings.Contains(body, ">must start with sk_<") || strings.Contains(body, "SECRET") {
		t.Errorf("body %s, want only the message", body)
	}
}

// sensitiveCode is a struct the catch-all binds through its text methods.
type sensitiveCode struct{ v string }

func (c sensitiveCode) MarshalText() ([]byte, error)  { return []byte(c.v), nil }
func (c *sensitiveCode) UnmarshalText(b []byte) error { c.v = string(b); return nil }

// A password widget is sensitive by itself: it never shows the stored
// password, and a blank submission keeps it.
func TestReflectFormHandler_PasswordWidgetIsSensitive(t *testing.T) {
	type record struct {
		PW string `form:"widget=password"`
	}
	if out := renderFieldHTML(t, &record{PW: "hunter2"}, "PW"); strings.Contains(out, "hunter2") {
		t.Errorf("a password widget shows its value: %s", out)
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{PW: "hunter2"}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	if rec := postValues(t, h, url.Values{mx.PresentSentinelName("PW"): {"1"}, "PW": {""}}); rec.Code != http.StatusSeeOther || saved == nil || saved.PW != "hunter2" {
		t.Errorf("status %d, saved %+v; want the stored password kept", rec.Code, saved)
	}
}

// postValues posts vals to h as a urlencoded form.
func postValues(t *testing.T, h http.HandlerFunc, vals url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

// checkedColor is an enum that validates itself.
type checkedColor string

func (checkedColor) EnumStrings() []string { return []string{"red", "green"} }

func (c checkedColor) Validate() error {
	if c != "red" && c != "green" {
		return errors.New("not a color")
	}
	return nil
}

// prio is an int enum whose zero is no valid value.
type prio int

func (prio) Enums() []prio { return []prio{1, 2} }

func (p prio) Validate() error {
	if p != 1 && p != 2 {
		return errors.New("invalid prio")
	}
	return nil
}

// A repeatable row validates like a top-level field: a checkbox set is
// validated element by element, and the "(none)" of an optional int enum
// saves, there and at the top level.
func TestReflectFormHandler_RowEnumValidation(t *testing.T) {
	type line struct {
		Colors map[checkedColor]struct{}
		Prio   prio
	}
	type record struct {
		Prio  prio
		Lines []line `form:"repeatable"`
	}
	var saved *record
	handler := mx.ReflectFormHandler(nil, func(_ context.Context, r *record) error { saved = r; return nil }, FieldDecider)
	post := func(colors ...string) *httptest.ResponseRecorder {
		form := url.Values{}
		form.Set(mx.PresentSentinelName("Prio"), "1")
		form.Set("Prio", "")
		form.Set(mx.PresentSentinelName("Lines"), "1")
		form.Set(mx.RowSentinelName("Lines-0"), "new")
		form.Set(mx.PresentSentinelName("Lines-0-Prio"), "1")
		form.Set("Lines-0-Prio", "")
		form.Set(mx.PresentSentinelName("Lines-0-Colors"), "1")
		for _, c := range colors {
			form.Add("Lines-0-Colors", c)
		}
		req := httptest.NewRequest(http.MethodPost, "/r", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		handler(rec, req)
		return rec
	}
	if rec := post("red"); rec.Code != http.StatusSeeOther || saved == nil {
		t.Fatalf("blank optional int enums: status %d, body %s", rec.Code, rec.Body.String())
	}
	saved = nil
	if rec := post("red", "blue"); rec.Code != http.StatusOK || saved != nil || !strings.Contains(rec.Body.String(), "not a color") {
		t.Errorf("an invalid set member in a row: status %d, saved %+v, body %s", rec.Code, saved, rec.Body.String())
	}
}

// A required int enum's blank "(none)" is rejected, at the top level and in
// a row. A blank binds the zero, which the handler's required check counts
// as a number, not as empty, and the validation chain skips for a blank
// enum, so the parser's required check is what keeps a 0 that is no
// constant from being saved.
func TestReflectFormHandler_RequiredIntEnumRejectsBlank(t *testing.T) {
	type line struct {
		Color colorCode `form:"required"`
	}
	type record struct {
		Color colorCode `form:"required"`
		Lines []line    `form:"repeatable"`
	}
	for _, c := range []struct {
		name, top, row string
		wantSaved      bool
	}{{"both chosen", "red", "blue", true}, {"blank top", "", "blue", false}, {"blank row", "red", "", false}} {
		var saved *record
		h := mx.ReflectFormHandler(nil, func(_ context.Context, r *record) error { saved = r; return nil }, FieldDecider)
		vals := url.Values{}
		vals.Set(mx.PresentSentinelName("Color"), "1")
		vals.Set("Color", c.top)
		vals.Set(mx.PresentSentinelName("Lines"), "1")
		vals.Set(mx.RowSentinelName("Lines-0"), "new")
		vals.Set(mx.PresentSentinelName("Lines-0-Color"), "1")
		vals.Set("Lines-0-Color", c.row)
		rec := postValues(t, h, vals)
		if c.wantSaved {
			if rec.Code != http.StatusSeeOther || saved == nil || saved.Color != colorRed || len(saved.Lines) != 1 || saved.Lines[0].Color != colorBlue {
				t.Errorf("%s: status %d, saved %+v, body %s", c.name, rec.Code, saved, rec.Body.String())
			}
			continue
		}
		if rec.Code != http.StatusOK || saved != nil || !strings.Contains(rec.Body.String(), ">required<") {
			t.Errorf("%s: a blank required int enum must not save: status %d, saved %+v, body %s", c.name, rec.Code, saved, rec.Body.String())
		}
	}
}

// levelCode is an int enum whose first constant is its zero value, as the
// usual iota pattern gives: the zero is one of its values, not a blank.
type levelCode int

const (
	levelLow levelCode = iota
	levelHigh
)

func (levelCode) Enums() []levelCode { return []levelCode{levelLow, levelHigh} }

func (l levelCode) MarshalText() ([]byte, error) {
	switch l {
	case levelLow:
		return []byte("low"), nil
	case levelHigh:
		return []byte("high"), nil
	}
	return nil, fmt.Errorf("bad level %d", int(l))
}

func (l *levelCode) UnmarshalText(text []byte) error {
	switch string(text) {
	case "low":
		*l = levelLow
	case "high":
		*l = levelHigh
	default:
		return fmt.Errorf("bad level %q", text)
	}
	return nil
}

// A required int enum whose zero value is one of its constants saves that
// constant when it is chosen, at the top level and in a row: only a blank
// select is "required", not the zero a chosen option binds.
func TestReflectFormHandler_RequiredZeroEnumConstantSaves(t *testing.T) {
	type line struct {
		Level levelCode `form:"required"`
	}
	type record struct {
		Level levelCode `form:"required"`
		Lines []line    `form:"repeatable"`
	}
	for _, c := range []struct {
		name, top, row string
		wantSaved      bool
	}{{"zero chosen", "low", "low", true}, {"blank top", "", "low", false}, {"blank row", "low", "", false}} {
		var saved *record
		h := mx.ReflectFormHandler(
			func(context.Context) (*record, error) { return &record{Level: levelHigh}, nil },
			func(_ context.Context, r *record) error { saved = r; return nil },
			FieldDecider,
		)
		vals := url.Values{}
		vals.Set(mx.PresentSentinelName("Level"), "1")
		vals.Set("Level", c.top)
		vals.Set(mx.PresentSentinelName("Lines"), "1")
		vals.Set(mx.RowSentinelName("Lines-0"), "new")
		vals.Set(mx.PresentSentinelName("Lines-0-Level"), "1")
		vals.Set("Lines-0-Level", c.row)
		rec := postValues(t, h, vals)
		if c.wantSaved {
			if rec.Code != http.StatusSeeOther || saved == nil || saved.Level != levelLow || len(saved.Lines) != 1 || saved.Lines[0].Level != levelLow {
				t.Errorf("%s: the zero constant must save: status %d, saved %+v, body %s", c.name, rec.Code, saved, rec.Body.String())
			}
			continue
		}
		if rec.Code != http.StatusOK || saved != nil || !strings.Contains(rec.Body.String(), ">required<") {
			t.Errorf("%s: a blank required enum must not save: status %d, saved %+v, body %s", c.name, rec.Code, saved, rec.Body.String())
		}
	}
}

// After a submit, the field error says the text is not in the format, so
// the hint only names the format.
func TestFieldDecider_FormatHintAfterError(t *testing.T) {
	type form struct {
		Due string `form:"widget=date"`
	}
	f := &form{Due: "06.07.2026"}
	v := reflect.ValueOf(f).Elem()
	field, _ := v.Type().FieldByName("Due")
	beh := FieldDecider("Due", field, v.Field(0))
	out := renderToString(t, beh.Render("Due", field, v.Field(0), []error{errors.New(`"06.07.2026" is not a date in the format YYYY-MM-DD`)}))
	if !strings.Contains(out, ">Expected format: YYYY-MM-DD<") || strings.Contains(out, "Not in the expected format") {
		t.Errorf("hint after an error: %s", out)
	}
}

// A time shown as text round-trips: it is written as RFC 3339, the text
// its parser reads back, zone and fraction included.
func TestReflectFormHandler_TextTimeRoundTrip(t *testing.T) {
	type record struct {
		At   time.Time `form:"widget=text"`
		Seen time.Time `form:"hidden"`
	}
	stamp := time.Date(2026, 7, 6, 9, 5, 7, 123456789, time.FixedZone("CEST", 2*60*60))
	loaded := record{At: stamp, Seen: stamp}
	var saved *record
	handler := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { r := loaded; return &r, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/r", nil))
	req := httptest.NewRequest(http.MethodPost, "/r", strings.NewReader(submittedValues(rec.Body.String()).Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusSeeOther || saved == nil || !saved.At.Equal(stamp) || !saved.Seen.Equal(stamp) {
		t.Errorf("unchanged POST: status %d, saved %+v, body %s", rec.Code, saved, rec.Body.String())
	}

	// The zero time, as on a new record, shows empty inputs.
	loaded = record{}
	rec = httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/r", nil))
	if body := rec.Body.String(); strings.Contains(body, "0001-01-01") {
		t.Errorf("the zero time renders a value: %s", body)
	}
}

// opaqueDate reads any text but cannot write it back.
type opaqueDate struct {
	text string
}

func (d opaqueDate) MarshalText() ([]byte, error) { return nil, errors.New("opaque") }

func (d *opaqueDate) UnmarshalText(text []byte) error {
	d.text = string(text)
	return nil
}

// secondsClock holds a time of day as text that always carries seconds,
// like civil.Time's.
type secondsClock string

func (c *secondsClock) UnmarshalText(text []byte) error {
	t, err := time.Parse(time.TimeOnly, string(text))
	if err != nil {
		return err
	}
	*c = secondsClock(t.Format(time.TimeOnly))
	return nil
}

// secondsStamp is a date and time held as text whose UnmarshalText needs
// seconds.
type secondsStamp string

func (s *secondsStamp) UnmarshalText(text []byte) error {
	t, err := time.Parse("2006-01-02T15:04:05", string(text))
	if err != nil {
		return err
	}
	*s = secondsStamp(t.Format("2006-01-02T15:04:05"))
	return nil
}

// panickyText's UnmarshalText panics, as a promoted method of a nil
// embedded pointer does.
type panickyText string

func (*panickyText) UnmarshalText([]byte) error { panic("nil receiver") }

// A time held as text whose UnmarshalText needs seconds takes a whole
// minute from the time or datetime input, which leaves zero seconds out.
// An UnmarshalText that panics rejects the value instead of failing the
// request, and the posted text stays bound.
func TestFieldDecider_TimeTextNeedsSeconds(t *testing.T) {
	type form struct {
		Start secondsClock `form:"widget=time"`
		At    secondsStamp `form:"widget=datetime"`
		Bad   panickyText  `form:"widget=date"`
	}
	f := &form{}
	for _, b := range []struct{ raw, want string }{{"09:05", "09:05:00"}, {"09:05:07", "09:05:07"}} {
		if err := parseFieldValue(t, f, "Start", b.raw); err != nil || string(f.Start) != b.want {
			t.Errorf("bind %q: error %v, stored %q, want %q", b.raw, err, f.Start, b.want)
		}
	}
	if err := parseFieldValue(t, f, "At", "2026-06-14T09:05"); err != nil || string(f.At) != "2026-06-14T09:05:00" {
		t.Errorf("bind datetime: error %v, stored %q", err, f.At)
	}
	if err := parseFieldValue(t, f, "Bad", "2026-06-14"); err == nil || string(f.Bad) != "2026-06-14" {
		t.Errorf("bind into a panicking UnmarshalText: error %v, stored %q", err, f.Bad)
	}
}

// dateOnly embeds a time but writes and reads only its date.
type dateOnly struct {
	time.Time
}

func (d dateOnly) MarshalText() ([]byte, error) { return []byte(d.Format(time.DateOnly)), nil }

func (d *dateOnly) UnmarshalText(text []byte) error {
	t, err := time.Parse(time.DateOnly, string(text))
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

// A struct embedding time.Time with text methods of its own stays a text
// input through them, as before this change, instead of a datetime picker
// whose values it cannot read.
func TestReflectFormHandler_DateOnlyWrapperIsText(t *testing.T) {
	type record struct {
		Day dateOnly
	}
	var saved *record
	handler := mx.ReflectFormHandler(
		func(context.Context) (*record, error) {
			return &record{Day: dateOnly{time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)}}, nil
		},
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/r", nil))
	body := rec.Body.String()
	if want := `<input type="text" name="Day" id="Day" value="2026-07-06"/>`; !strings.Contains(body, want) {
		t.Fatalf("GET lacks %s: %s", want, body)
	}
	req := httptest.NewRequest(http.MethodPost, "/r", strings.NewReader(submittedValues(body).Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusSeeOther || saved == nil || !saved.Day.Equal(time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("unchanged POST: status %d, saved %+v", rec.Code, saved)
	}
}

// A readonly nullable field is never written, so it offers no clear control.
func TestFieldDecider_ReadonlyNullableHasNoClearControl(t *testing.T) {
	type form struct {
		Locked sampleNullable `form:"readonly"`
		Open   sampleNullable
	}
	f := &form{Locked: sampleNullable{v: "x"}, Open: sampleNullable{v: "y"}}
	if out := renderFieldHTML(t, f, "Locked"); strings.Contains(out, mx.ClearSentinelName("Locked")) {
		t.Errorf("a readonly field offers a clear control: %s", out)
	}
	if out := renderFieldHTML(t, f, "Open"); !strings.Contains(out, mx.ClearSentinelName("Open")) {
		t.Errorf("a nullable field lacks its clear control: %s", out)
	}
}

// dayNumber is an int-kind date, counting days since 1970-01-01, with ISO
// text methods.
type dayNumber int32

func (d dayNumber) MarshalText() ([]byte, error) {
	return []byte(time.Unix(int64(d)*86400, 0).UTC().Format(time.DateOnly)), nil
}

func (d *dayNumber) UnmarshalText(text []byte) error {
	t, err := time.Parse(time.DateOnly, string(text))
	if err != nil {
		return err
	}
	*d = dayNumber(t.Unix() / 86400)
	return nil
}

// A date type that is no struct binds through its UnmarshalText like a date
// struct does, so an unchanged save keeps the value it rendered. A type with
// neither is an error that names the type, without a call stack.
func TestReflectFormHandler_IntKindDateRoundTrip(t *testing.T) {
	type form struct {
		Day   dayNumber       `form:"widget=date"`
		Plain struct{ N int } `form:"widget=date"`
	}
	july := dayNumber(time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC).Unix() / 86400)
	f := &form{Day: july}
	if out := renderFieldHTML(t, f, "Day"); !strings.Contains(out, `value="2026-07-06"`) {
		t.Fatalf("Day renders %s", out)
	}
	f.Day = 0
	if err := parseFieldValue(t, f, "Day", "2026-07-06"); err != nil || f.Day != july {
		t.Errorf("bind Day: error %v, stored %d, want %d", err, f.Day, july)
	}
	if err := parseFieldValue(t, f, "Plain", "2026-07-06"); fmt.Sprint(err) != "cannot assign time to struct { N int }" {
		t.Errorf("bind Plain: error %v", err)
	}
}

// Binding never writes through a pointer the field shares with a stored
// record, as a shallow-copying load leaves it: rejected or bound, the value
// goes into new pointees, and the stored record changes only by onSubmit.
func TestFieldDecider_DateBindKeepsSharedPointees(t *testing.T) {
	type form struct {
		Text  *string        `form:"widget=date"`
		Time  *time.Time     `form:"widget=date,min=2026-01-01"`
		Ptr   ptrWrappedTime `form:"widget=date"`
		Named optDate        `form:"widget=date"`
	}
	july := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	storedText, storedTime, storedPtr, storedNamed := "2026-07-06", july, july, july
	f := &form{Text: &storedText, Time: &storedTime, Ptr: ptrWrappedTime{&storedPtr}, Named: optDate{&storedNamed}}
	for _, b := range []struct {
		field, raw string
		wantErr    bool
	}{
		{"Text", "garbage", true},
		{"Text", "2026-08-05", false},
		{"Time", "2025-01-01", true},
		{"Time", "2026-08-05", false},
		{"Ptr", "06.07.2026", true},
		{"Ptr", "2026-08-05", false},
		{"Named", "2026-08-05", false},
	} {
		if err := parseFieldValue(t, f, b.field, b.raw); (err != nil) != b.wantErr {
			t.Errorf("bind %s %q: error %v, want error %t", b.field, b.raw, err, b.wantErr)
		}
	}
	if storedText != "2026-07-06" || !storedTime.Equal(july) || !storedPtr.Equal(july) || !storedNamed.Equal(july) {
		t.Errorf("stored pointees changed: %q %v %v %v", storedText, storedTime, storedPtr, storedNamed)
	}
	aug := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)
	if *f.Text != "2026-08-05" || !f.Time.Equal(aug) || f.Ptr.Time == nil || !f.Ptr.Equal(aug) || f.Named.At == nil || !f.Named.At.Equal(aug) {
		t.Errorf("the form must hold the bound values: %q %v %v %v", *f.Text, f.Time, f.Ptr.Time, f.Named.At)
	}
}

// optDate is a date struct holding its time in a named pointer field, which
// its UnmarshalText writes through.
type optDate struct{ At *time.Time }

func (o optDate) MarshalText() ([]byte, error) {
	if o.At == nil {
		return nil, nil
	}
	return []byte(o.At.Format(time.DateOnly)), nil
}

func (o *optDate) UnmarshalText(b []byte) error {
	t, err := time.Parse(time.DateOnly, string(b))
	if err != nil {
		return err
	}
	if o.At == nil {
		o.At = new(time.Time)
	}
	*o.At = t
	return nil
}

// A sensitive field's errors name the format and the bound but never echo
// the posted value into the page.
func TestFieldDecider_SensitiveDateErrors(t *testing.T) {
	type form struct {
		Text    string      `form:"widget=date,sensitive,max=2010-12-31"`
		Time    time.Time   `form:"widget=date,sensitive"`
		Civil   civilDate   `form:"widget=date,sensitive"`
		Wrapped wrappedTime `form:"widget=date,sensitive"`
		Strict  strictDate  `form:"widget=date,sensitive"`
		German  germanDate  `form:"widget=date,sensitive,min=2026-01-01"`
	}
	f := &form{}
	for _, b := range []struct{ field, raw, wantErr string }{
		{"Text", "02.01.1980", "not a date in the format YYYY-MM-DD"},
		{"Text", "2015-03-04", "later than 2010-12-31"},
		{"Time", "02.01.1980", "not a date in the format YYYY-MM-DD"},
		// The types' own UnmarshalText errors would quote the value.
		{"Civil", "06.07.1980", "not a date in the format YYYY-MM-DD"},
		{"Wrapped", "06.07.1980", "not a date in the format YYYY-MM-DD"},
		{"Strict", "1999-12-31", "not a date in the format YYYY-MM-DD"},
		// The range applies to a sensitive value in the type's own spelling.
		{"German", "6.7.2025", "earlier than 2026-01-01"},
	} {
		err := parseFieldValue(t, f, b.field, b.raw)
		if err == nil || err.Error() != b.wantErr {
			t.Errorf("bind %s %q: error %v, want %s", b.field, b.raw, err, b.wantErr)
		}
	}
}

// money is a registered single-value struct that is no time: it is a text
// input through MarshalText and UnmarshalText, not a date or time field.
type money struct {
	Cents    int
	Currency string
}

func (m money) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%d.%02d %s", m.Cents/100, m.Cents%100, m.Currency), nil
}

func (m *money) UnmarshalText(text []byte) error {
	var units, cents int
	var currency string
	if _, err := fmt.Sscanf(string(text), "%d.%d %s", &units, &cents, &currency); err != nil {
		return err
	}
	*m = money{units*100 + cents, currency}
	return nil
}

// boxedMoney embeds a pointer to money; its promoted UnmarshalText panics
// on a nil one once the text parses.
type boxedMoney struct {
	*money
}

// The text and number binders guard the type's own methods: a panic is a
// field error instead of a failed request, and a sensitive field's error
// does not quote the posted value, as UnmarshalText and strconv errors do.
func TestFieldDecider_ScalarBindGuards(t *testing.T) {
	type form struct {
		Box   boxedMoney
		Count int      `form:"sensitive"`
		Total money    `form:"sensitive"`
		Tag   string   `form:"options=test-html-tags,sensitive"`
		Tags  []string `form:"options=test-html-tags,sensitive"`
	}
	f := &form{}
	for _, b := range []struct{ field, raw, wantErr string }{
		{"Box", "1.00 EUR", `"1.00 EUR" cannot be stored as html.boxedMoney`},
		{"Count", "secret", "not a valid value"},
		{"Total", "secret", "not a valid value"},
		// The option list's error would quote the posted value.
		{"Tag", "secret", "not a valid value"},
		{"Tags", "secret", "not a valid value"},
	} {
		if err := parseFieldValue(t, f, b.field, b.raw); fmt.Sprint(err) != b.wantErr {
			t.Errorf("bind %s %q: error %v, want %s", b.field, b.raw, err, b.wantErr)
		}
	}
	// Rendering guards the promoted MarshalText of the nil pointer too.
	if out := renderFieldHTML(t, f, "Box"); !strings.Contains(out, `<input type="text" name="Box" id="Box"/>`) {
		t.Errorf("Box renders %s", out)
	}
}

func TestReflectFormHandler_RegisteredMoneyIsText(t *testing.T) {
	mx.SingleValueTypes.Register(reflect.TypeFor[money]())
	type invoice struct {
		Total money
	}
	var saved *invoice
	handler := mx.ReflectFormHandler(
		func(context.Context) (*invoice, error) { return &invoice{Total: money{1250, "EUR"}}, nil },
		func(_ context.Context, r *invoice) error { saved = r; return nil },
		FieldDecider,
	)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/invoice", nil))
	body := rec.Body.String()
	if want := `<input type="text" name="Total" id="Total" value="12.50 EUR"/>`; !strings.Contains(body, want) {
		t.Fatalf("GET lacks %s: %s", want, body)
	}
	req := httptest.NewRequest(http.MethodPost, "/invoice", strings.NewReader(submittedValues(body).Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusSeeOther || saved == nil || saved.Total != (money{1250, "EUR"}) {
		t.Errorf("unchanged POST: status %d, saved %+v; want 303 and 12.50 EUR kept", rec.Code, saved)
	}
}

// Text no date input sends, past any input's length, is rejected without
// being echoed: neither the error nor the re-rendered input repeats it, as
// %q would grow a control byte to four. A database/sql number's error
// leaves such text out too.
func TestReflectFormHandler_LongDateTextNotEchoed(t *testing.T) {
	type record struct {
		Due   string    `form:"widget=date"`
		At    time.Time `form:"widget=date"`
		Count sql.NullInt64
	}
	long := strings.Repeat("\x01", 200)
	for _, field := range []string{"Due", "At", "Count"} {
		var saved *record
		h := mx.ReflectFormHandler(
			func(context.Context) (*record, error) { return &record{Due: "2026-07-06"}, nil },
			func(_ context.Context, r *record) error { saved = r; return nil },
			FieldDecider,
		)
		vals := url.Values{}
		vals.Set(mx.PresentSentinelName(mx.FieldPath(field)), "1")
		vals.Set(field, long)
		rec := postValues(t, h, vals)
		body := rec.Body.String()
		if rec.Code != http.StatusOK || saved != nil {
			t.Errorf("%s: status %d, saved %+v; want a rejection", field, rec.Code, saved)
		}
		if strings.Contains(body, `\x01\x01`) || strings.Contains(body, "\x01\x01") || strings.Contains(body, "&#x1;&#x1;") || strings.Contains(body, "&#1;&#1;") {
			t.Errorf("%s: the long text was echoed: %s", field, body)
		}
	}
	// The stored text stays in the input after the rejection.
	type textOnly struct {
		Due string `form:"widget=date"`
	}
	f := &textOnly{Due: "2026-07-06"}
	if err := parseFieldValue(t, f, "Due", long); err == nil || f.Due != "2026-07-06" {
		t.Errorf("a long date text: error %v, Due %q; want an error and the stored date kept", err, f.Due)
	}
}

// A context option list that fails while a POST is parsed puts no internal
// error into the page: like an onSubmit error, its text is shown only when
// mx.RevealInternalServerErrors is set.
func TestFieldDecider_OptionListParseErrorNotShown(t *testing.T) {
	type form struct {
		Partner string   `form:"options=test-html-partners"`
		Tags    []string `form:"options=test-html-partners"`
	}
	// The partners list reads the request context, which has none here.
	for _, field := range []string{"Partner", "Tags"} {
		err := parseFieldValue(t, &form{}, field, "a")
		if err == nil || err.Error() != mx.FormErrorMessage {
			t.Errorf("%s: error %v, want the generic %q", field, err, mx.FormErrorMessage)
		}
	}
	// A list that fails with a mx.FormMessage, also wrapped, means its text
	// for the user, so the field shows it instead of the generic message.
	type messageForm struct {
		Partner string   `form:"options=test-html-partners-message"`
		Tags    []string `form:"options=test-html-partners-message"`
	}
	for _, field := range []string{"Partner", "Tags"} {
		err := parseFieldValue(t, &messageForm{}, field, "a")
		if err == nil || err.Error() != "The partner list is unavailable, try again later." {
			t.Errorf("%s: error %v, want the list's FormMessage", field, err)
		}
	}
	prev := mx.RevealInternalServerErrors
	mx.RevealInternalServerErrors = true
	t.Cleanup(func() { mx.RevealInternalServerErrors = prev })
	if err := parseFieldValue(t, &form{}, "Partner", "a"); err == nil || !strings.Contains(err.Error(), "no partners in context") {
		t.Errorf("a revealed option-list error: %v", err)
	}
}

// deDate is a date struct that writes and reads only its own German
// spelling, 02.01.2006, not the ISO text a date input sends.
type deDate struct{ Y, M, D int }

func (d deDate) MarshalText() ([]byte, error) {
	if d == (deDate{}) {
		return nil, nil
	}
	return []byte(fmt.Sprintf("%02d.%02d.%04d", d.D, d.M, d.Y)), nil
}

func (d *deDate) UnmarshalText(b []byte) error {
	t, err := time.Parse("02.01.2006", string(b))
	if err != nil {
		return err
	}
	*d = deDate{t.Year(), int(t.Month()), t.Day()}
	return nil
}

// A date type that reads only its own spelling cannot take what a date
// input sends, so it renders as a text input that asks for no ISO format,
// stored or empty, and saves its own spelling. A min or max cannot be
// checked on such text, so the form fails on GET instead of rejecting
// every save.
func TestReflectFormHandler_OwnSpellingDateType(t *testing.T) {
	type record struct {
		Due  deDate `form:"widget=date"`
		Next deDate `form:"widget=date"`
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{Due: deDate{2026, 7, 6}}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	body := get.Body.String()
	for _, want := range []string{
		`<input type="text" name="Due" id="Due" value="06.07.2026"/>`,
		`<input type="text" name="Next" id="Next"/>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("want %s in %s", want, body)
		}
	}
	if strings.Contains(body, "YYYY-MM-DD") || strings.Contains(body, "aria-invalid") {
		t.Errorf("an own-spelling date type must not ask for ISO text: %s", body)
	}
	vals := submittedValues(body)
	vals.Set("Next", "01.08.2026")
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil || saved.Due != (deDate{2026, 7, 6}) || saved.Next != (deDate{2026, 8, 1}) {
		t.Errorf("own spelling: status %d, saved %+v", rec.Code, saved)
	}

	type bounded struct {
		Due deDate `form:"widget=date,min=2026-01-01"`
	}
	bh := mx.ReflectFormHandler(
		func(context.Context) (*bounded, error) { return &bounded{Due: deDate{2026, 7, 6}}, nil },
		func(context.Context, *bounded) error { return nil },
		FieldDecider,
	)
	prev := mx.RevealInternalServerErrors
	mx.RevealInternalServerErrors = true
	t.Cleanup(func() { mx.RevealInternalServerErrors = prev })
	rec := httptest.NewRecorder()
	bh(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), "min") {
		t.Errorf("a bound on an own-spelling date type: status %d, body %s", rec.Code, rec.Body.String())
	}
}

// A file field posted urlencoded, as by a form that missed the multipart
// enctype, carries only the file's name: that is an error, not a save that
// silently drops the upload. A blank file input keeps the stored file.
func TestFieldDecider_FileFieldNotMultipart(t *testing.T) {
	type form struct {
		Doc []byte `form:"widget=file"`
	}
	f := &form{Doc: []byte("stored")}
	if err := parseFieldValue(t, f, "Doc", "scan.pdf"); err == nil || string(f.Doc) != "stored" {
		t.Errorf("a file name posted urlencoded: error %v, Doc %q", err, f.Doc)
	}
	if err := parseFieldValue(t, f, "Doc", ""); err != nil || string(f.Doc) != "stored" {
		t.Errorf("a blank file input posted urlencoded: error %v, Doc %q", err, f.Doc)
	}
}

// A file input left empty keeps the stored file, so a required file field
// that holds one does not make the browser demand a new upload on every
// save; with nothing stored it does.
func TestFieldDecider_StoredRequiredFileNotDemanded(t *testing.T) {
	type form struct {
		Doc []byte `form:"widget=file,required"`
	}
	stored := renderFieldHTML(t, &form{Doc: []byte("pdf")}, "Doc")
	if strings.Contains(stored, "required") {
		t.Errorf("a stored file must not make the input required: %s", stored)
	}
	if empty := renderFieldHTML(t, &form{}, "Doc"); !strings.Contains(empty, `required="required"`) {
		t.Errorf("an empty required file input must be required: %s", empty)
	}
}

func init() {
	mx.RegisterNamedOptions("test-html-floors", func(context.Context) ([]mx.NamedOption, error) {
		return []mx.NamedOption{{Name: "Ground", Value: "0"}, {Name: "First", Value: "1"}}, nil
	})
}

// floorNo reads and writes its digits as text; its zero, the ground
// floor, is an option of every list below.
type floorNo int

func (f floorNo) MarshalText() ([]byte, error) { return []byte(strconv.Itoa(int(f))), nil }

func (f *floorNo) UnmarshalText(b []byte) error {
	n, err := strconv.Atoi(string(b))
	*f = floorNo(n)
	return err
}

// namedFloor lists its options itself, the zero among them.
type namedFloor int

func (namedFloor) NamedOptions() []mx.NamedOption {
	return []mx.NamedOption{{Name: "Ground", Value: "0"}, {Name: "First", Value: "1"}}
}
func (f namedFloor) MarshalText() ([]byte, error)  { return floorNo(f).MarshalText() }
func (f *namedFloor) UnmarshalText(b []byte) error { return (*floorNo)(f).UnmarshalText(b) }

// stringsFloor is an enum by EnumStrings alone, without Enums.
type stringsFloor int

func (stringsFloor) EnumStrings() []string            { return []string{"0", "1"} }
func (f stringsFloor) MarshalText() ([]byte, error)  { return floorNo(f).MarshalText() }
func (f *stringsFloor) UnmarshalText(b []byte) error { return (*floorNo)(f).UnmarshalText(b) }

// A required select whose zero value is one of its options saves that
// option whatever lists it: a form:"options=…" provider, NamedOptions or
// EnumStrings, not only an Enums method.
func TestReflectFormHandler_RequiredZeroOptionSaves(t *testing.T) {
	type record struct {
		Tagged  floorNo      `form:"required,options=test-html-floors"`
		Named   namedFloor   `form:"required"`
		Strings stringsFloor `form:"required"`
	}
	for _, field := range []string{"Tagged", "Named", "Strings"} {
		var saved *record
		h := mx.ReflectFormHandler(
			func(context.Context) (*record, error) { return &record{Tagged: 1, Named: 1, Strings: 1}, nil },
			func(_ context.Context, r *record) error { saved = r; return nil },
			FieldDecider,
		)
		vals := url.Values{}
		vals.Set(mx.PresentSentinelName(mx.FieldPath(field)), "1")
		vals.Set(field, "0")
		if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil {
			t.Errorf("%s: choosing the listed option 0: status %d, saved %+v, body %s", field, rec.Code, saved, rec.Body.String())
		}
	}
}

// A date type that reads only its own spelling is a text input also when
// null or sensitive, where a date picker would send text it rejects, and a
// caller's placeholder is its hint.
func TestReflectFormHandler_OwnSpellingDateNullAndSensitive(t *testing.T) {
	type record struct {
		Due    *deDate `form:"widget=date,placeholder=TT.MM.JJJJ"`
		Secret deDate  `form:"widget=date,sensitive"`
	}
	var saved *record
	h := mx.ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{Secret: deDate{2026, 7, 6}}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		FieldDecider,
	)
	get := httptest.NewRecorder()
	h(get, httptest.NewRequest(http.MethodGet, "/", nil))
	body := get.Body.String()
	for _, want := range []string{
		`<input type="text" name="Due" id="Due" placeholder="TT.MM.JJJJ"/>`,
		`<input type="text" name="Secret" id="Secret"/>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("want %s in %s", want, body)
		}
	}
	vals := submittedValues(body)
	vals.Set("Due", "01.08.2026")
	vals.Set("Secret", "02.08.2026")
	if rec := postValues(t, h, vals); rec.Code != http.StatusSeeOther || saved == nil || saved.Due == nil || *saved.Due != (deDate{2026, 8, 1}) || saved.Secret != (deDate{2026, 8, 2}) {
		t.Errorf("own spelling, null and sensitive: status %d, saved %+v", rec.Code, saved)
	}
}

// An optional select or checkbox set refers to its option note before a
// context list resolves, so with nothing removed the note still exists,
// empty and hidden: the aria-describedby never points at nothing.
func TestFieldDecider_OptionNoteReferenceNeverDangles(t *testing.T) {
	type draftForm struct {
		Partner string   `form:"options=test-html-partners"`
		Tags    []string `form:"options=test-html-partners"`
	}
	ctx := context.WithValue(context.Background(), partnersCtxKey{},
		[]mx.NamedOption{{Name: "Partner One", Value: "p1"}})
	v := reflect.ValueOf(&draftForm{Partner: "p1", Tags: []string{"p1"}}).Elem()
	for _, name := range []string{"Partner", "Tags"} {
		field, _ := v.Type().FieldByName(name)
		value := v.FieldByName(name)
		comp := FieldDecider(mx.FieldPath(name), field, value).Render(mx.FieldPath(name), field, value, nil)
		var b strings.Builder
		if err := comp.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
			t.Fatalf("%s: render: %v", name, err)
		}
		out := b.String()
		if !strings.Contains(out, `aria-describedby="`+name+`-note"`) || !strings.Contains(out, `<p id="`+name+`-note" hidden="hidden"></p>`) {
			t.Errorf("%s: with nothing removed the reference must point at an empty hidden note: %s", name, out)
		}
	}
}

// A file field may store only the uploaded file's name (a string
// destination binds the file name). Holding a name, the required input
// must not make the browser demand a new upload on every save.
func TestFieldDecider_StoredRequiredFileNameNotDemanded(t *testing.T) {
	type form struct {
		Doc string `form:"widget=file,required"`
	}
	if stored := renderFieldHTML(t, &form{Doc: "scan.pdf"}, "Doc"); strings.Contains(stored, "required") {
		t.Errorf("a stored file name must not make the input required: %s", stored)
	}
	if empty := renderFieldHTML(t, &form{}, "Doc"); !strings.Contains(empty, `required="required"`) {
		t.Errorf("an empty required file input must be required: %s", empty)
	}
}

// rtPrio is an int enum whose zero is none of its constants, bound from its
// digits, so a forged "0" binds to the zero.
type rtPrio int

func (rtPrio) Enums() []rtPrio { return []rtPrio{1, 2} }

func (p rtPrio) Validate() error {
	if p != 1 && p != 2 {
		return errors.New("not a priority")
	}
	return nil
}

func (p rtPrio) MarshalText() ([]byte, error) { return []byte(strconv.Itoa(int(p))), nil }

func (p *rtPrio) UnmarshalText(b []byte) error {
	n, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}
	*p = rtPrio(n)
	return nil
}

// A forged non-empty value that binds a required enum to a zero that is no
// constant is no blank choice: it is validated and rejected, not saved,
// also in a row.
func TestReflectFormHandler_RequiredEnumForgedZeroRejected(t *testing.T) {
	type line struct {
		Prio rtPrio `form:"required"`
	}
	type record struct {
		Prio  rtPrio `form:"required"`
		Lines []line `form:"repeatable"`
	}
	for _, raw := range []string{"0", "00", "-0"} {
		for _, row := range []bool{false, true} {
			var saved *record
			h := mx.ReflectFormHandler(
				func(context.Context) (*record, error) { return &record{Prio: 2}, nil },
				func(_ context.Context, r *record) error { saved = r; return nil },
				FieldDecider,
			)
			vals := url.Values{}
			vals.Set(mx.PresentSentinelName("Prio"), "1")
			vals.Set("Prio", raw)
			if row {
				vals.Set("Prio", "2")
				vals.Set(mx.PresentSentinelName("Lines"), "1")
				vals.Set(mx.RowSentinelName("Lines-0"), "new")
				vals.Set(mx.PresentSentinelName("Lines-0-Prio"), "1")
				vals.Set("Lines-0-Prio", raw)
			}
			if rec := postValues(t, h, vals); saved != nil || rec.Code != http.StatusOK {
				t.Errorf("%q (row %t): status %d, saved %+v", raw, row, rec.Code, saved)
			}
		}
	}
}

// A required date wrapper whose embedded time pointer is nil holds no
// value: the required check names it instead of panicking in the promoted
// IsZero of the nil pointer.
func TestReflectFormHandler_RequiredNilWrappedTimeIsBlank(t *testing.T) {
	type record struct {
		At ptrWrappedTime `form:"widget=date,required"`
	}
	var saved *record
	h := mx.ReflectFormHandler(nil, func(_ context.Context, r *record) error { saved = r; return nil }, FieldDecider)
	vals := url.Values{}
	vals.Set(mx.PresentSentinelName("At"), "1")
	vals.Set("At", "")
	if rec := postValues(t, h, vals); rec.Code != http.StatusOK || saved != nil || !strings.Contains(rec.Body.String(), ">required<") {
		t.Errorf("a blank required wrapper: status %d, saved %+v, body %s", rec.Code, saved, rec.Body.String())
	}
}

// A sensitive own-spelling date's error names no format: the input's
// YYYY-MM-DD is text the type rejects. A type that reads the input's text
// but only dates of its own range is no own-spelling type.
func TestFieldDecider_OwnSpellingDateErrors(t *testing.T) {
	type form struct {
		Secret deDate     `form:"widget=date,sensitive"`
		Future futureDate `form:"widget=date"`
	}
	f := &form{Secret: deDate{2026, 7, 6}}
	err := parseFieldValue(t, f, "Secret", "2026-08-02")
	if err == nil || strings.Contains(err.Error(), "YYYY-MM-DD") || strings.Contains(err.Error(), "2026") {
		t.Errorf("a sensitive own-spelling date's error: %v", err)
	}
	if in := mx.DateTimeInput(reflect.ValueOf(futureDate{}), mx.FormTag{Widget: "date"}); in.Type != "date" || in.Format != "YYYY-MM-DD" {
		t.Errorf("a range-holding ISO date type must keep the date input: %+v", in)
	}
}

// futureDate reads ISO dates but takes none before 2020.
type futureDate struct{ t time.Time }

func (d futureDate) MarshalText() ([]byte, error) {
	if d.t.IsZero() {
		return nil, nil
	}
	return []byte(d.t.Format(time.DateOnly)), nil
}

func (d *futureDate) UnmarshalText(b []byte) error {
	t, err := time.Parse(time.DateOnly, string(b))
	if err != nil {
		return err
	}
	if t.Year() < 2020 {
		return errors.New("before 2020")
	}
	d.t = t
	return nil
}

// An enum whose zero is none of its options holds no choice at its zero,
// as on a new record: the select offers "(none)" or "Choose…", not an
// unavailable value with a note that saving removes it.
func TestFieldDecider_UnlistedZeroIsNoSelection(t *testing.T) {
	type form struct {
		Optional colorCode
		Required colorCode `form:"required"`
	}
	f := &form{}
	optional := renderFieldHTML(t, f, "Optional")
	if !strings.Contains(optional, `<option value="" selected="selected">(none)</option>`) || strings.Contains(optional, "not available") || strings.Contains(optional, "removed when you save") {
		t.Errorf("an optional unlisted zero: %s", optional)
	}
	required := renderFieldHTML(t, f, "Required")
	if !strings.Contains(required, `selected="selected">Choose…</option>`) || strings.Contains(required, "not available") {
		t.Errorf("a required unlisted zero: %s", required)
	}
}
