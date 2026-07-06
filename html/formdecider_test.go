package html

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

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
	if strings.Contains(out, "selected") {
		t.Errorf("p1 must not be selected: %q", out)
	}

	out = renderWith([]mx.NamedOption{{Name: "Partner Two", Value: "p2"}})
	if !strings.Contains(out, "Partner Two") {
		t.Errorf("same component must render second context's options: %q", out)
	}
	if !strings.Contains(out, "selected") {
		t.Errorf("p2 matches the field value and must be selected: %q", out)
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

// fakeID is a stand-in for uu.ID-shaped foreign key types: a [16]byte
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
