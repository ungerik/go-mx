package mx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// Minimal text-input decider used across handler tests. It renders a
// hidden __present marker plus a single <input name="Path"> per
// string-shaped field, and parses one form value per field. Numeric
// and bool kinds are handled inline so the tests don't need a full
// html package.
func newTestDecider() FieldDecider {
	return func(path FieldPath, field reflect.StructField, value reflect.Value) FieldBehavior {
		return FieldBehavior{
			Render: func(path FieldPath, field reflect.StructField, value reflect.Value, errs []error) Component {
				attribs := []any{
					Attribute{Name: "name", Value: string(path)},
					Attribute{Name: "value", Value: stringify(value)},
				}
				input := NewVoidElement("input")
				for _, a := range attribs {
					input.Attribs = append(input.Attribs, a.(Attribute))
				}
				if len(errs) > 0 {
					return Components{input, NewElement("p",
						Attribute{Name: "data-error", Value: string(path)},
						Text(errors.Join(errs...).Error()),
					)}
				}
				return input
			},
			Parse: func(path FieldPath, field reflect.StructField, value reflect.Value, r *http.Request) error {
				raw := r.PostForm.Get(string(path))
				return setFromString(value, raw)
			},
		}
	}
}

func stringify(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	}
	return fmt.Sprint(v.Interface())
}

func setFromString(v reflect.Value, raw string) error {
	if !v.CanSet() {
		return errors.New("not settable")
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(n)
	case reflect.Bool:
		v.SetBool(raw == "on" || raw == "true" || raw == "1")
	default:
		return fmt.Errorf("unsupported kind %s", v.Kind())
	}
	return nil
}

type sampleStruct struct {
	Name string
	Age  int
}

func TestReflectFormHandler_GetRenders(t *testing.T) {
	load := func(ctx context.Context) (*sampleStruct, error) {
		return &sampleStruct{Name: "Alice", Age: 42}, nil
	}
	onSubmit := func(ctx context.Context, s *sampleStruct) error { return nil }

	h := ReflectFormHandler(load, onSubmit, newTestDecider())

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `name="Name"`) || !strings.Contains(body, `value="Alice"`) {
		t.Errorf("missing Name input or value: %q", body)
	}
	if !strings.Contains(body, `name="Age"`) || !strings.Contains(body, `value="42"`) {
		t.Errorf("missing Age input or value: %q", body)
	}
	if !strings.Contains(body, `name="__present__Name"`) {
		t.Errorf("missing __present sentinel: %q", body)
	}
	if !strings.Contains(body, `method="post"`) {
		t.Errorf("missing method=post: %q", body)
	}
	// The form must self-submit to the URL that served it so a native
	// submit posts back to this handler, not the current document.
	if !strings.Contains(body, `action="/admin"`) {
		t.Errorf("missing action targeting the request URI: %q", body)
	}
}

// A reflected form is commonly served as an HTMX fragment (hx-get) and
// swapped into a page living at a different, GET-only URL. The rendered
// <form> must carry action = the URL that served it (path+query), so the
// native submit posts back to its own handler instead of the embedding
// page (which would 405).
func TestReflectFormHandler_FormActionSelfSubmits(t *testing.T) {
	load := func(ctx context.Context) (*sampleStruct, error) {
		return &sampleStruct{Name: "Alice", Age: 42}, nil
	}
	onSubmit := func(ctx context.Context, s *sampleStruct) error { return nil }
	h := ReflectFormHandler(load, onSubmit, newTestDecider())

	// Fragment endpoint /partners/new/form embedded into a page at
	// /partners/new — the action must target the serving path, not the
	// embedding page, and must preserve the query string.
	req := httptest.NewRequest(http.MethodGet, "/partners/new/form?tab=main", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `action="/partners/new/form?tab=main"`) {
		t.Errorf("form action must target the serving URL path+query; got %q", body)
	}
	// It must NOT default to the parent/embedding path.
	if strings.Contains(body, `action="/partners/new"`) {
		t.Errorf("form action must not target the embedding page: %q", body)
	}
}

// The whole point of the default action is to preserve the serving URL,
// including a multi-parameter query string. A query with more than one
// parameter contains an ampersand, which the attribute writer escapes to
// &amp; — pin that so a future change can't emit an unescaped (or
// dropped) query string and break self-submission.
func TestReflectFormHandler_FormActionEscapesQuery(t *testing.T) {
	load := func(ctx context.Context) (*sampleStruct, error) {
		return &sampleStruct{}, nil
	}
	onSubmit := func(ctx context.Context, s *sampleStruct) error { return nil }
	h := ReflectFormHandler(load, onSubmit, newTestDecider())

	req := httptest.NewRequest(http.MethodGet, "/partners/new/form?tab=main&sort=asc", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `action="/partners/new/form?tab=main&amp;sort=asc"`) {
		t.Errorf("form action must preserve and escape the full query string; got %q", body)
	}
}

// The default action is derived from untrusted request input. A request
// reachable at a path starting with "//" must NOT render a protocol-
// relative action attribute, which a browser resolves off-origin and would
// use to post the form's fields to an attacker's host on a native submit.
// selfSubmitAction collapses the leading slash run to force a same-origin
// path.
func TestReflectFormHandler_FormActionNotProtocolRelative(t *testing.T) {
	load := func(ctx context.Context) (*sampleStruct, error) {
		return &sampleStruct{}, nil
	}
	onSubmit := func(ctx context.Context, s *sampleStruct) error { return nil }
	h := ReflectFormHandler(load, onSubmit, newTestDecider())

	for _, target := range []string{"//evil.example/collect", "///evil.example/collect"} {
		req := httptest.NewRequest(http.MethodGet, target, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("target %q: status=%d, want 200", target, rec.Code)
		}
		body := rec.Body.String()
		// Must render the same-origin, single-leading-slash path.
		if !strings.Contains(body, `action="/evil.example/collect"`) {
			t.Errorf("target %q: action must be a same-origin absolute path; got %q", target, body)
		}
		// Must NOT render a protocol-relative "//host" action.
		if strings.Contains(body, `action="//`) {
			t.Errorf("target %q: action must not be protocol-relative; got %q", target, body)
		}
	}
}

// ReflectFormConfig.Action overrides the default self-submit target,
// mirroring the Redirect knob.
func TestReflectFormHandler_FormActionOverride(t *testing.T) {
	load := func(ctx context.Context) (*sampleStruct, error) {
		return &sampleStruct{}, nil
	}
	onSubmit := func(ctx context.Context, s *sampleStruct) error { return nil }
	cfg := ReflectFormConfig{
		Action: func(r *http.Request) string { return "/custom/submit" },
	}
	h := ReflectFormHandlerWith(cfg, load, onSubmit, newTestDecider())

	req := httptest.NewRequest(http.MethodGet, "/partners/new/form", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, `action="/custom/submit"`) {
		t.Errorf("Action override not applied; got %q", body)
	}
}

func TestReflectFormHandler_PostUpdates(t *testing.T) {
	var captured *sampleStruct
	load := func(ctx context.Context) (*sampleStruct, error) {
		return &sampleStruct{Name: "before", Age: 1}, nil
	}
	onSubmit := func(ctx context.Context, s *sampleStruct) error {
		captured = s
		return nil
	}

	h := ReflectFormHandler(load, onSubmit, newTestDecider())

	form := url.Values{}
	form.Set(PresentSentinelName("Name"), "1")
	form.Set("Name", "after")
	form.Set(PresentSentinelName("Age"), "1")
	form.Set("Age", "99")

	req := httptest.NewRequest(http.MethodPost, "/admin", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d, want 303", rec.Code)
	}
	if captured == nil {
		t.Fatal("onSubmit not called")
	}
	if captured.Name != "after" || captured.Age != 99 {
		t.Errorf("captured=%+v, want {after 99}", captured)
	}
}

func TestReflectFormHandler_AllowlistByConstruction(t *testing.T) {
	// __present omitted for Name → Name must NOT be overwritten,
	// even though the POST contains a value for it.
	var captured *sampleStruct
	load := func(ctx context.Context) (*sampleStruct, error) {
		return &sampleStruct{Name: "immutable", Age: 1}, nil
	}
	onSubmit := func(ctx context.Context, s *sampleStruct) error {
		captured = s
		return nil
	}

	h := ReflectFormHandler(load, onSubmit, newTestDecider())

	form := url.Values{}
	form.Set("Name", "tampered") // no __present sentinel — allowlist miss
	form.Set(PresentSentinelName("Age"), "1")
	form.Set("Age", "7")

	req := httptest.NewRequest(http.MethodPost, "/admin", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if captured == nil {
		t.Fatal("onSubmit not called")
	}
	if captured.Name != "immutable" {
		t.Errorf("Name=%q, want immutable (mass-assignment defense)", captured.Name)
	}
	if captured.Age != 7 {
		t.Errorf("Age=%d, want 7", captured.Age)
	}
}

type fieldErrorsStruct struct {
	Name string
}

func TestReflectFormHandler_FieldErrorsRoute(t *testing.T) {
	load := func(ctx context.Context) (*fieldErrorsStruct, error) {
		return &fieldErrorsStruct{}, nil
	}
	onSubmit := func(ctx context.Context, s *fieldErrorsStruct) error {
		return FieldErrors{
			"Name": errors.New("name is taken"),
		}
	}

	h := ReflectFormHandler(load, onSubmit, newTestDecider())

	form := url.Values{}
	form.Set(PresentSentinelName("Name"), "1")
	form.Set("Name", "alice")

	req := httptest.NewRequest(http.MethodPost, "/admin", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200 (re-render)", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `data-error="Name"`) || !strings.Contains(body, "name is taken") {
		t.Errorf("inline error missing from %q", body)
	}
}

func TestReflectFormHandler_ParseError(t *testing.T) {
	load := func(ctx context.Context) (*sampleStruct, error) {
		return &sampleStruct{}, nil
	}
	onSubmit := func(ctx context.Context, s *sampleStruct) error {
		t.Fatalf("onSubmit must not run when Parse fails")
		return nil
	}

	h := ReflectFormHandler(load, onSubmit, newTestDecider())

	form := url.Values{}
	form.Set(PresentSentinelName("Age"), "1")
	form.Set("Age", "not-a-number")

	req := httptest.NewRequest(http.MethodPost, "/admin", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200 (re-render with error)", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `data-error="Age"`) {
		t.Errorf("expected per-field error, got: %s", rec.Body.String())
	}
}

func TestReflectFormHandler_LoadError(t *testing.T) {
	load := func(ctx context.Context) (*sampleStruct, error) {
		return nil, errors.New("db down")
	}
	onSubmit := func(ctx context.Context, s *sampleStruct) error { return nil }
	h := ReflectFormHandler(load, onSubmit, newTestDecider())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status=%d, want 500", rec.Code)
	}
}

func TestReflectFormHandler_Unconfigured(t *testing.T) {
	// No decider supplied, no Middleware wrap → unconfiguredDecider's
	// Render emits the "no decider configured" message and Parse
	// surfaces a parse error.
	load := func(ctx context.Context) (*sampleStruct, error) {
		return &sampleStruct{}, nil
	}
	onSubmit := func(ctx context.Context, s *sampleStruct) error { return nil }
	h := ReflectFormHandler(load, onSubmit)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rec.Body.String()
	if !strings.Contains(body, "Middleware") {
		t.Errorf("unconfigured response should mention Middleware; got %q", body)
	}
}

func TestReflectFormHandler_MethodNotAllowed(t *testing.T) {
	h := ReflectFormHandler(
		func(ctx context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil },
		func(ctx context.Context, s *sampleStruct) error { return nil },
		newTestDecider(),
	)
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status=%d, want 405", rec.Code)
	}
}

// Race-detector regression for D2/D11. Each goroutine submits its own
// POST against the same handler instance. The loader returns a fresh
// *T per request, so concurrent submissions must not bleed.
func TestReflectFormHandler_ConcurrentPostsAreIsolated(t *testing.T) {
	const N = 64
	var received sync.Map // requestID -> sampleStruct seen by onSubmit

	load := func(ctx context.Context) (*sampleStruct, error) {
		return &sampleStruct{}, nil
	}
	onSubmit := func(ctx context.Context, s *sampleStruct) error {
		received.Store(s.Name, *s)
		return nil
	}
	h := ReflectFormHandler(load, onSubmit, newTestDecider())

	var wg sync.WaitGroup
	wg.Add(N)
	var failures atomic.Int32

	for i := range N {
		go func(i int) {
			defer wg.Done()
			form := url.Values{}
			form.Set(PresentSentinelName("Name"), "1")
			form.Set("Name", fmt.Sprintf("worker-%03d", i))
			form.Set(PresentSentinelName("Age"), "1")
			form.Set("Age", strconv.Itoa(i))

			req := httptest.NewRequest(http.MethodPost, "/admin",
				strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusSeeOther {
				failures.Add(1)
			}
			io.ReadAll(rec.Body)
		}(i)
	}
	wg.Wait()

	if failures.Load() != 0 {
		t.Fatalf("%d/%d concurrent posts failed", failures.Load(), N)
	}
	for i := range N {
		want := sampleStruct{Name: fmt.Sprintf("worker-%03d", i), Age: i}
		got, ok := received.Load(want.Name)
		if !ok {
			t.Errorf("missing onSubmit record for %s", want.Name)
			continue
		}
		if got != want {
			t.Errorf("for %s: got %+v, want %+v", want.Name, got, want)
		}
	}
}

// Validate readonly fields and __clear sentinel handling end-to-end.
type nullableStr struct {
	v string
	n bool
}

func (n nullableStr) IsNull() bool { return n.n }
func (n *nullableStr) SetNull()    { n.v = ""; n.n = true }

type clearableStruct struct {
	Locked    string      `form:"readonly"`
	Optional  nullableStr // nullable, supports SetNull
	Mandatory string      `form:"required"`
}

type lockedLine struct {
	Approved nullableStr `form:"readonly"`
}

type lockedRecord struct {
	Approved nullableStr  `form:"readonly"`
	Lines    []lockedLine `form:"repeatable"`
}

// A tampered __clear sentinel cannot null a readonly field, at the top
// level or in a repeatable row: readonly never writes a submitted value.
func TestReflectFormHandler_ReadonlyIgnoresClear(t *testing.T) {
	var captured *lockedRecord
	h := ReflectFormHandler(
		func(context.Context) (*lockedRecord, error) {
			return &lockedRecord{Approved: nullableStr{v: "yes"}, Lines: []lockedLine{{nullableStr{v: "yes"}}}}, nil
		},
		func(_ context.Context, r *lockedRecord) error { captured = r; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	form.Set(PresentSentinelName("Approved"), "1")
	form.Set(ClearSentinelName("Approved"), "1")
	addRow(form, "Lines", 0, map[string]string{"Approved": ""})
	form.Set(ClearSentinelName("Lines-0-Approved"), "1")
	if rec := postForm(t, h, form); rec.Code != http.StatusSeeOther {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
	}
	if captured == nil || captured.Approved.IsNull() || len(captured.Lines) != 1 || captured.Lines[0].Approved.IsNull() || captured.Lines[0].Approved.v != "yes" {
		t.Errorf("readonly values cleared: %+v", captured)
	}
}

func TestReflectFormHandler_ReadonlyAndClear(t *testing.T) {
	var captured clearableStruct
	load := func(ctx context.Context) (*clearableStruct, error) {
		return &clearableStruct{
			Locked:    "system",
			Optional:  nullableStr{v: "before"},
			Mandatory: "ok",
		}, nil
	}
	onSubmit := func(ctx context.Context, s *clearableStruct) error {
		captured = *s
		return nil
	}

	// Custom decider — minimal Parse for string + nullable.
	decider := func(path FieldPath, field reflect.StructField, value reflect.Value) FieldBehavior {
		return FieldBehavior{
			Render: func(path FieldPath, field reflect.StructField, value reflect.Value, errs []error) Component {
				return Text(string(path))
			},
			Parse: func(path FieldPath, field reflect.StructField, value reflect.Value, r *http.Request) error {
				raw := r.PostForm.Get(string(path))
				switch v := value.Addr().Interface().(type) {
				case *string:
					*v = raw
				case *nullableStr:
					v.v = raw
					v.n = false
				}
				return nil
			},
		}
	}

	h := ReflectFormHandler(load, onSubmit, decider)

	form := url.Values{}
	form.Set(PresentSentinelName("Locked"), "1")
	form.Set("Locked", "tampered")
	form.Set(PresentSentinelName("Optional"), "1")
	form.Set(ClearSentinelName("Optional"), "1")
	form.Set(PresentSentinelName("Mandatory"), "1")
	form.Set("Mandatory", "still here")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d, body=%q", rec.Code, rec.Body.String())
	}
	if captured.Locked != "system" {
		t.Errorf("readonly field overwritten: %q", captured.Locked)
	}
	if !captured.Optional.IsNull() {
		t.Errorf("nullable field not cleared: %+v", captured.Optional)
	}
	if captured.Mandatory != "still here" {
		t.Errorf("Mandatory=%q", captured.Mandatory)
	}
}

// nil load means "submit-only" — handler seeds GET/POST with new(T).
func TestReflectFormHandler_NilLoadUsesZeroValue(t *testing.T) {
	var captured *sampleStruct
	onSubmit := func(ctx context.Context, s *sampleStruct) error {
		captured = s
		return nil
	}
	h := ReflectFormHandler(nil, onSubmit, newTestDecider())

	// GET seeds with new(T): empty struct renders with zero values.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status=%d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `name="Name"`) {
		t.Errorf("expected Name input on GET: %s", body)
	}
	// Empty Name on the rendered form means value="" (or no value attr).
	if strings.Contains(body, `value="Alice"`) {
		t.Errorf("GET should not seed with prior values: %s", body)
	}

	// POST against the nil-load handler still parses into a fresh
	// new(T) and reaches onSubmit.
	form := url.Values{}
	form.Set(PresentSentinelName("Name"), "1")
	form.Set("Name", "submitted")
	form.Set(PresentSentinelName("Age"), "1")
	form.Set("Age", "21")
	req := httptest.NewRequest(http.MethodPost, "/",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST status=%d body=%s", rec.Code, rec.Body.String())
	}
	if captured == nil {
		t.Fatal("onSubmit not called")
	}
	if captured.Name != "submitted" || captured.Age != 21 {
		t.Errorf("captured=%+v", captured)
	}
}

func TestReflectFormHandler_NilOnSubmitPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for nil onSubmit")
		}
	}()
	load := func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil }
	_ = ReflectFormHandler(load, nil)
}

// postRedirecting drives a urlencoded POST of form through h and returns the
// recorder. Shared by the ReflectFormHandlerRedirecting tests.
func postRedirecting(t *testing.T, h http.HandlerFunc, target string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// validSampleForm submits valid values for both sampleStruct fields, so a POST
// reaches onSubmit without field errors.
func validSampleForm() url.Values {
	return url.Values{
		PresentSentinelName("Name"): {"x"}, "Name": {"x"},
		PresentSentinelName("Age"): {"1"}, "Age": {"1"},
	}
}

// TestReflectFormHandlerRedirecting_ReturnedURL is the whole point of the
// variant: onSubmit learns a per-submission destination (here, a Stripe-style
// checkout URL) and the handler 303s to exactly that URL.
func TestReflectFormHandlerRedirecting_ReturnedURL(t *testing.T) {
	const want = "https://checkout.stripe.test/session/abc"
	load := func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil }
	onSubmit := func(ctx context.Context, s *sampleStruct) (string, error) { return want, nil }
	h := ReflectFormHandlerRedirecting(ReflectFormConfig{}, load, onSubmit, newTestDecider())

	rec := postRedirecting(t, h, "/admin", validSampleForm())
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d, want 303; body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Location"); got != want {
		t.Errorf("Location=%q, want %q", got, want)
	}
}

// TestReflectFormHandlerRedirecting_EmptyFallsBackToConfigRedirect: an empty
// returned URL defers to cfg.Redirect, preserving the existing knob.
func TestReflectFormHandlerRedirecting_EmptyFallsBackToConfigRedirect(t *testing.T) {
	load := func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil }
	onSubmit := func(ctx context.Context, s *sampleStruct) (string, error) { return "", nil }
	cfg := ReflectFormConfig{Redirect: func(*http.Request) string { return "/done" }}
	h := ReflectFormHandlerRedirecting(cfg, load, onSubmit, newTestDecider())

	rec := postRedirecting(t, h, "/admin", validSampleForm())
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d, want 303", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/done" {
		t.Errorf("Location=%q, want /done (cfg.Redirect fallback)", got)
	}
}

// TestReflectFormHandlerRedirecting_EmptyFallsBackToRequestPath: empty URL and
// no cfg.Redirect falls all the way back to r.URL.Path.
func TestReflectFormHandlerRedirecting_EmptyFallsBackToRequestPath(t *testing.T) {
	load := func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil }
	onSubmit := func(ctx context.Context, s *sampleStruct) (string, error) { return "", nil }
	h := ReflectFormHandlerRedirecting(ReflectFormConfig{}, load, onSubmit, newTestDecider())

	rec := postRedirecting(t, h, "/admin/edit", validSampleForm())
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d, want 303", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/admin/edit" {
		t.Errorf("Location=%q, want /admin/edit (r.URL.Path fallback)", got)
	}
}

// TestReflectFormHandlerRedirecting_ErrorReRenders confirms the redirecting
// variant did not shift error behavior: a plain error re-renders the form (no
// redirect), and a FieldErrors still routes to its field.
func TestReflectFormHandlerRedirecting_ErrorReRenders(t *testing.T) {
	load := func(context.Context) (*fieldErrorsStruct, error) { return &fieldErrorsStruct{}, nil }
	form := url.Values{PresentSentinelName("Name"): {"1"}, "Name": {"alice"}}

	t.Run("form-level error", func(t *testing.T) {
		h := ReflectFormHandlerRedirecting(ReflectFormConfig{}, load,
			func(ctx context.Context, s *fieldErrorsStruct) (string, error) {
				return "", errors.New("pq: connection refused to 10.0.0.5")
			}, newTestDecider())
		rec := postRedirecting(t, h, "/", form)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 (re-render)", rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "" {
			t.Errorf("errored submit must not redirect, got Location=%q", loc)
		}
		// An internal error's details never reach the page.
		body := rec.Body.String()
		if !strings.Contains(body, FormErrorMessage) || strings.Contains(body, "10.0.0.5") {
			t.Errorf("a plain error must show only the generic message: %s", body)
		}
	})

	t.Run("FormMessage is shown", func(t *testing.T) {
		h := ReflectFormHandlerRedirecting(ReflectFormConfig{}, load,
			func(ctx context.Context, s *fieldErrorsStruct) (string, error) {
				return "", fmt.Errorf("saving: %w", FormMessage("the plan is full"))
			}, newTestDecider())
		if body := postRedirecting(t, h, "/", form).Body.String(); !strings.Contains(body, ">the plan is full<") {
			t.Errorf("a FormMessage must render: %s", body)
		}
	})

	t.Run("RevealInternalServerErrors shows the error", func(t *testing.T) {
		prev := RevealInternalServerErrors
		RevealInternalServerErrors = true
		t.Cleanup(func() { RevealInternalServerErrors = prev })
		h := ReflectFormHandlerRedirecting(ReflectFormConfig{}, load,
			func(ctx context.Context, s *fieldErrorsStruct) (string, error) {
				return "", errors.New("boom")
			}, newTestDecider())
		if body := postRedirecting(t, h, "/", form).Body.String(); !strings.Contains(body, "boom") {
			t.Errorf("a revealed error must render: %s", body)
		}
	})

	t.Run("FieldErrors routes per field", func(t *testing.T) {
		h := ReflectFormHandlerRedirecting(ReflectFormConfig{}, load,
			func(ctx context.Context, s *fieldErrorsStruct) (string, error) {
				return "", FieldErrors{"Name": errors.New("name is taken")}
			}, newTestDecider())
		rec := postRedirecting(t, h, "/", form)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d, want 200 (re-render)", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, `data-error="Name"`) || !strings.Contains(body, "name is taken") {
			t.Errorf("FieldErrors should route to the Name field: %s", body)
		}
	})
}

type strictEnum string

func (strictEnum) EnumStrings() []string { return []string{"a", "b"} }

func (s strictEnum) Validate() error {
	if s != "a" && s != "b" {
		return errors.New("invalid")
	}
	return nil
}

// A required field whose __present sentinel is missing is still checked:
// leaving the sentinel out of a submit-only form cannot save a blank.
func TestReflectFormHandler_RequiredWithoutSentinel(t *testing.T) {
	type signup struct {
		Email  string   `form:"required"`
		Prio   prioEnum `form:"required"`
		Locked string   `form:"readonly,required"`
	}
	var saved *signup
	handler := ReflectFormHandler(nil, func(_ context.Context, s *signup) error { saved = s; return nil }, newTestDecider())
	rec := postForm(t, handler, url.Values{})
	if rec.Code != http.StatusOK || saved != nil {
		t.Fatalf("status %d, saved %+v; want the required errors and nothing saved", rec.Code, saved)
	}
	body := rec.Body.String()
	for _, path := range []string{"Email", "Prio"} {
		if !strings.Contains(body, `data-error="`+path+`"`) {
			t.Errorf("no required error at %s: %s", path, body)
		}
	}
	// A readonly field the user cannot fill never blocks the save.
	if strings.Contains(body, `data-error="Locked"`) {
		t.Errorf("a readonly field got a required error: %s", body)
	}
}

// A required field the decider does not render has no sentinel and shows
// no error, so it is left alone instead of blocking every save.
func TestReflectFormHandler_UnrenderedRequiredFieldSaves(t *testing.T) {
	type hidden struct {
		Name     string
		Approver string `form:"required"`
		Lines    []struct {
			Desc     string
			Approver string `form:"required"`
		} `form:"repeatable"`
	}
	base := newTestDecider()
	decider := func(p FieldPath, f reflect.StructField, v reflect.Value) FieldBehavior {
		if f.Name == "Approver" {
			return FieldBehavior{}
		}
		return base(p, f, v)
	}
	saved := false
	handler := ReflectFormHandler(nil, func(context.Context, *hidden) error { saved = true; return nil }, decider)
	form := url.Values{}
	form.Set(PresentSentinelName("Name"), "1")
	form.Set("Name", "n")
	addRow(form, "Lines", 0, map[string]string{"Desc": "d"})
	form.Set(RowSentinelName("Lines-0"), "new")
	if rec := postForm(t, handler, form); rec.Code != http.StatusSeeOther || !saved {
		t.Errorf("status %d, saved %t; want the save to go through", rec.Code, saved)
	}
}

// prioEnum is an int enum whose zero is no valid value.
type prioEnum int

func (prioEnum) Enums() []prioEnum { return []prioEnum{1, 2} }

func (p prioEnum) Validate() error {
	if p != 1 && p != 2 {
		return errors.New("invalid prio")
	}
	return nil
}

// An added row's required cell cannot be skipped by leaving out its
// sentinel. (The "(none)" of an optional int enum is covered with the html
// decider, which binds an empty select.)
func TestRepeatable_AddedRowRequiredWithoutSentinel(t *testing.T) {
	type line struct {
		Desc string `form:"required"`
		Prio prioEnum
	}
	type record struct {
		Prio  prioEnum
		Lines []line `form:"repeatable"`
	}
	var saved *record
	handler := ReflectFormHandler(nil, func(_ context.Context, r *record) error { saved = r; return nil }, newTestDecider())
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Prio": "1"})
	form.Set(RowSentinelName("Lines-0"), "new")
	rec := postForm(t, handler, form)
	if rec.Code != http.StatusOK || saved != nil || !strings.Contains(rec.Body.String(), `data-error="Lines-0-Desc"`) {
		t.Errorf("an added row without its required cell: status %d, saved %+v", rec.Code, saved)
	}
}

type enumLine struct {
	Color strictEnum
	Set   map[strictEnum]struct{}
}

type enumSheet struct {
	Lines []enumLine `form:"repeatable"`
}

// A blank optional enum in a repeatable row saves, as at the top level.
// (Set members are validated per element too; the test decider cannot bind
// a set, so html.TestReflectFormHandler_RowEnumSetValidated covers that.)
func TestRepeatable_RowEnumValidation(t *testing.T) {
	var saved *enumSheet
	handler := ReflectFormHandler(nil, func(_ context.Context, s *enumSheet) error { saved = s; return nil }, newTestDecider())
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Color": ""})
	if rec := postForm(t, handler, form); rec.Code != http.StatusSeeOther || saved == nil {
		t.Fatalf("a blank optional enum in a row: status %d, body %s", rec.Code, rec.Body.String())
	}
}

// A body that cannot be parsed is answered with an error, never treated as
// a submission of nothing that saves the unchanged record.
// A field the decider renders as nothing, or not at all, is not in the
// form, so a forged __present sentinel and value cannot write it.
func TestReflectFormHandler_UnrenderedFieldNotWritten(t *testing.T) {
	type account struct {
		Name string
		Role string
		Plan string
	}
	var saved *account
	base := newTestDecider()
	h := ReflectFormHandler(
		func(context.Context) (*account, error) { return &account{Role: "member", Plan: "free"}, nil },
		func(_ context.Context, a *account) error { saved = a; return nil },
		func(path FieldPath, field reflect.StructField, value reflect.Value) FieldBehavior {
			beh := base(path, field, value)
			switch path {
			case "Role": // hidden from this user by rendering nothing
				beh.Render = func(FieldPath, reflect.StructField, reflect.Value, []error) Component { return nil }
			case "Plan":
				beh.Render = nil
			}
			return beh
		},
	)
	rec := postForm(t, h, url.Values{
		PresentSentinelName("Name"): {"1"}, "Name": {"n"},
		PresentSentinelName("Role"): {"1"}, "Role": {"admin"},
		PresentSentinelName("Plan"): {"1"}, "Plan": {"pro"},
	})
	if rec.Code != http.StatusSeeOther || saved == nil || saved.Name != "n" || saved.Role != "member" || saved.Plan != "free" {
		t.Errorf("status %d, saved %+v; want Name written and Role, Plan kept", rec.Code, saved)
	}
}

// A multipart body past MaxUploadSize answers 413, by default past
// DefaultMaxUploadSize; a negative MaxUploadSize sets no bound.
func TestReflectFormHandler_MaxUploadSize(t *testing.T) {
	prev := DefaultMaxUploadSize
	DefaultMaxUploadSize = 2048
	t.Cleanup(func() { DefaultMaxUploadSize = prev })
	upload := func() (*bytes.Buffer, string) {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		_ = mw.WriteField(PresentSentinelName("Name"), "1")
		_ = mw.WriteField("Name", "n")
		part, _ := mw.CreateFormFile("Doc", "doc.bin")
		_, _ = part.Write(bytes.Repeat([]byte("x"), 4096))
		_ = mw.Close()
		return &body, mw.FormDataContentType()
	}
	for _, c := range []struct {
		name  string
		limit int64
		want  int
	}{
		{"default", 0, http.StatusRequestEntityTooLarge},
		{"own limit", 1024, http.StatusRequestEntityTooLarge},
		{"room enough", 1 << 20, http.StatusSeeOther},
		{"no bound", -1, http.StatusSeeOther},
	} {
		saved := false
		h := ReflectFormHandlerWith(ReflectFormConfig{MaxUploadSize: c.limit},
			func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil },
			func(context.Context, *sampleStruct) error { saved = true; return nil }, newTestDecider())
		body, contentType := upload()
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.want || saved != (c.want == http.StatusSeeOther) {
			t.Errorf("%s: status %d, saved %t; want %d", c.name, rec.Code, saved, c.want)
		}
	}
}

// maxFormBody is the size net/http's ParseForm reads of a urlencoded body
// that no http.MaxBytesReader limits.
const maxFormBody = 10 << 20

func TestReflectFormHandler_UnparseableBodyRejected(t *testing.T) {
	saved := false
	h := ReflectFormHandler(
		func(context.Context) (*sampleStruct, error) { return &sampleStruct{Name: "stored"}, nil },
		func(context.Context, *sampleStruct) error { saved = true; return nil },
		newTestDecider(),
	)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("--x\r\nContent-Disposition: form-data; name=\"Name\"\r\n\r\ncut"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=x")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || saved {
		t.Errorf("a truncated multipart body: status %d, saved %t", rec.Code, saved)
	}

	// A URL query that cannot be parsed fails nothing: the body is the
	// submission.
	req = httptest.NewRequest(http.MethodPost, "/edit?id=1;tab=2", strings.NewReader("__present__Name=1&Name=n"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther || !saved {
		t.Errorf("a valid body under a semicolon query: status %d, saved %t", rec.Code, saved)
	}
	saved = false
	// The query's readable values join r.Form, which a decider may read,
	// once, also when a middleware parsed the form already.
	var fromQuery []string
	qh := ReflectFormHandler(
		func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil },
		func(context.Context, *sampleStruct) error { return nil },
		func(path FieldPath, field reflect.StructField, value reflect.Value) FieldBehavior {
			beh := newTestDecider()(path, field, value)
			beh.Parse = func(_ FieldPath, _ reflect.StructField, _ reflect.Value, r *http.Request) error {
				fromQuery = r.Form["x"]
				return nil
			}
			return beh
		},
	)
	preParsed := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		qh.ServeHTTP(w, r)
	})
	for name, handler := range map[string]http.Handler{"fresh": qh, "parsed by a middleware": preParsed} {
		for _, target := range []string{"/edit?src=list;tab=2&x=1", "/edit?x=1"} {
			fromQuery = nil
			req = httptest.NewRequest(http.MethodPost, target, strings.NewReader("__present__Name=1&Name=n"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec = httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusSeeOther || !slices.Equal(fromQuery, []string{"1"}) {
				t.Errorf("%s, %s: status %d, r.Form[x] = %q; want 303 and one value", name, target, rec.Code, fromQuery)
			}
		}
	}
	// A malformed body is rejected, also when the query has the same error.
	for _, target := range []string{"/edit?id=1;tab=2", "/edit"} {
		req = httptest.NewRequest(http.MethodPost, target, strings.NewReader("__present__Name=1&Name=a;b"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest || saved {
			t.Errorf("a malformed body to %s: status %d, saved %t", target, rec.Code, saved)
		}
	}
	// A urlencoded body past net/http's 10 MB answers 413.
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("Name="+strings.Repeat("x", maxFormBody+1)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge || saved {
		t.Errorf("a urlencoded body past 10 MB: status %d, saved %t", rec.Code, saved)
	}
	// Exactly 10 MB is accepted, and so is more under a larger limit a
	// caller set with http.MaxBytesReader.
	prefix := "__present__Name=1&Name="
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(prefix+strings.Repeat("x", maxFormBody-len(prefix))))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther || !saved {
		t.Errorf("a urlencoded body of exactly 10 MB: status %d, saved %t", rec.Code, saved)
	}
	saved = false
	raised := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
		h.ServeHTTP(w, r)
	})
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(prefix+strings.Repeat("x", maxFormBody+1)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	raised.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther || !saved {
		t.Errorf("11 MB under a caller's 50 MB limit: status %d, saved %t", rec.Code, saved)
	}
	saved = false
	// A multipart upload is not held to the urlencoded limit.
	var upload bytes.Buffer
	uw := multipart.NewWriter(&upload)
	_ = uw.WriteField(PresentSentinelName("Name"), "1")
	_ = uw.WriteField("Name", "n")
	part, _ := uw.CreateFormFile("Doc", "big.bin")
	_, _ = part.Write(bytes.Repeat([]byte("x"), maxFormBody+1))
	_ = uw.Close()
	req = httptest.NewRequest(http.MethodPost, "/", &upload)
	req.Header.Set("Content-Type", uw.FormDataContentType())
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther || !saved {
		t.Errorf("a multipart body past 10 MB: status %d, saved %t", rec.Code, saved)
	}
	saved = false

	// A multipart body past the part limit answers 413.
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for i := range 1001 {
		if err := mw.WriteField("F"+strconv.Itoa(i), "x"); err != nil {
			t.Fatal(err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge || saved {
		t.Errorf("a multipart body past the part limit: status %d, saved %t", rec.Code, saved)
	}

	// A body past a size limit the server set answers 413.
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("Name="+strings.Repeat("x", 100)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	req.Body = http.MaxBytesReader(rec, req.Body, 10)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge || saved {
		t.Errorf("an oversized body: status %d, saved %t", rec.Code, saved)
	}
}

type sensitiveLine struct {
	Token string `form:"sensitive"`
}

type sensitiveRecord struct {
	Password string          `form:"sensitive"`
	PIN      int             `form:"sensitive"`
	Opt      bool            `form:"sensitive"`
	Lines    []sensitiveLine `form:"repeatable"`
}

// A sensitive field never shows its value, so a blank submission keeps
// the stored value, at the top level and in a row, while a new value
// replaces it.
func TestReflectFormHandler_SensitiveBlankKeepsStored(t *testing.T) {
	var saved *sensitiveRecord
	h := ReflectFormHandler(
		func(context.Context) (*sensitiveRecord, error) {
			return &sensitiveRecord{Password: "secret", PIN: 1234, Opt: true, Lines: []sensitiveLine{{"tok"}}}, nil
		},
		func(_ context.Context, r *sensitiveRecord) error { saved = r; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	for _, f := range []string{"Password", "PIN"} {
		form.Set(PresentSentinelName(FieldPath(f)), "1")
		form.Set(f, "")
	}
	// An unchecked checkbox sends no value; that is "false", not a blank.
	form.Set(PresentSentinelName("Opt"), "1")
	addRow(form, "Lines", 0, map[string]string{"Token": ""})
	if rec := postForm(t, h, form); rec.Code != http.StatusSeeOther {
		t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
	}
	if saved.Password != "secret" || saved.PIN != 1234 || saved.Lines[0].Token != "tok" {
		t.Errorf("blank sensitive fields must keep the stored values: %+v", saved)
	}
	if saved.Opt {
		t.Error("a sensitive checkbox could not be unchecked")
	}
	// Blank as the parser reads it: only the first value counts.
	form["Password"] = []string{"", "x"}
	if rec := postForm(t, h, form); rec.Code != http.StatusSeeOther || saved.Password != "secret" {
		t.Errorf("a blank first value must keep the stored value: status %d, saved %+v", rec.Code, saved)
	}
	form.Set("Password", "new")
	form.Set("PIN", "42")
	if rec := postForm(t, h, form); rec.Code != http.StatusSeeOther || saved.Password != "new" || saved.PIN != 42 {
		t.Errorf("a new value must replace the stored one: status %d, saved %+v", rec.Code, saved)
	}
}

// A stored sensitive value follows its row through a remove-row click,
// so a row left blank on the next save keeps its own stored value.
func TestReflectFormHandler_SensitiveCellFollowsRow(t *testing.T) {
	var saved *sensitiveRecord
	h := ReflectFormHandler(
		func(context.Context) (*sensitiveRecord, error) {
			return &sensitiveRecord{Lines: []sensitiveLine{{"a"}, {"b"}}}, nil
		},
		func(_ context.Context, r *sensitiveRecord) error { saved = r; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Token": ""})
	addRow(form, "Lines", 1, map[string]string{"Token": ""})
	form.Set(sentinelCommand, DeleteRowCommand("Lines-0"))
	removed := postForm(t, h, form)
	if removed.Code != http.StatusOK {
		t.Fatalf("remove-row: status %d", removed.Code)
	}
	if rec := postForm(t, h, renderedValues(removed.Body.String())); rec.Code != http.StatusSeeOther || len(saved.Lines) != 1 || saved.Lines[0].Token != "b" {
		t.Errorf("the stored token must follow its row: status %d, saved %+v", rec.Code, saved)
	}
}

// OwnerPart is embedded readonly by inheritedRecord.
type OwnerPart struct {
	Owner string
}

type auditInfo struct {
	By string
}

type inheritedRecord struct {
	OwnerPart `form:"readonly"`
	Audit     auditInfo `form:"section=Audit,readonly"`
}

// A field readonly through its embedded struct or its section is never
// written by a POST: not by a tampered value, a blank or a forged clear.
func TestReflectFormHandler_InheritedReadonlyIgnoresPost(t *testing.T) {
	var saved *inheritedRecord
	h := ReflectFormHandler(
		func(context.Context) (*inheritedRecord, error) {
			return &inheritedRecord{OwnerPart{"o"}, auditInfo{"alice"}}, nil
		},
		func(_ context.Context, r *inheritedRecord) error { saved = r; return nil },
		newTestDecider(),
	)
	// Each case reaches its own path: a value, a blank, and a clear alone.
	for _, c := range []struct {
		raw   string
		clear bool
	}{{"mallory", false}, {"", false}, {"", true}} {
		form := url.Values{}
		for _, p := range []FieldPath{"Owner", FieldPath("Audit").Append("By")} {
			form.Set(PresentSentinelName(p), "1")
			form.Set(string(p), c.raw)
			if c.clear {
				form.Set(ClearSentinelName(p), "1")
			}
		}
		if rec := postForm(t, h, form); rec.Code != http.StatusSeeOther || saved.Owner != "o" || saved.Audit.By != "alice" {
			t.Errorf("posting %q (clear %t) wrote an inherited readonly field: status %d, saved %+v", c.raw, c.clear, rec.Code, saved)
		}
	}
}

// hiddenKeyLine and hiddenKeyRecord key their rows with a hidden cell.
type hiddenKeyLine struct {
	Key  string `form:"hidden"`
	Desc string
}

type hiddenKeyRecord struct {
	Name  string
	Lines []hiddenKeyLine `form:"repeatable"`
}

// hiddenKeyDecider is the test decider that renders hidden fields as
// hidden inputs without errors, and the field "Skipped" as nothing.
func hiddenKeyDecider() FieldDecider {
	base := newTestDecider()
	return func(path FieldPath, field reflect.StructField, value reflect.Value) FieldBehavior {
		beh := base(path, field, value)
		kind, _ := DetectField(path, field, value)
		switch {
		case kind == FieldKindHidden:
			beh.Render = func(path FieldPath, _ reflect.StructField, value reflect.Value, _ []error) Component {
				return NewVoidElement("input",
					Attribute{Name: "type", Value: "hidden"},
					Attribute{Name: "name", Value: string(path)},
					Attribute{Name: "value", Value: stringify(value)},
				)
			}
		case field.Name == "Skipped":
			beh.Render = func(FieldPath, reflect.StructField, reflect.Value, []error) Component { return nil }
		}
		return beh
	}
}

// An error no input shows, of a hidden row cell or of a path that names no
// field, becomes a form-level message written as an onSubmit error is: a
// FormMessage's text, otherwise the generic message, once.
func TestReflectFormHandler_UnshownErrorsAtFormLevel(t *testing.T) {
	for _, c := range []struct {
		name string
		errs FieldErrors
		want string
	}{
		{"hidden cell message", FieldErrors{FieldPath("Lines").Append("0").Append("Key"): FormMessage("the row changed meanwhile")}, ">the row changed meanwhile<"},
		{"hidden cell internal", FieldErrors{FieldPath("Lines").Append("0").Append("Key"): errors.New("pq: row 7 locked")}, ">" + FormErrorMessage + "<"},
		{"misspelled path", FieldErrors{"Nmae": FormMessage("orphan message")}, ">orphan message<"},
	} {
		h := ReflectFormHandler(
			func(context.Context) (*hiddenKeyRecord, error) {
				return &hiddenKeyRecord{Name: "n", Lines: []hiddenKeyLine{{"k", "a"}}}, nil
			},
			func(context.Context, *hiddenKeyRecord) error { return c.errs },
			hiddenKeyDecider(),
		)
		form := url.Values{}
		form.Set(PresentSentinelName("Name"), "1")
		form.Set("Name", "n")
		addRow(form, "Lines", 0, map[string]string{"Key": "k", "Desc": "a"})
		body := postForm(t, h, form).Body.String()
		if !strings.Contains(body, `<p role="alert"`+c.want) || strings.Contains(body, "pq:") {
			t.Errorf("%s: body %s, want %s", c.name, body, c.want)
		}
	}
}

// A field the decider renders as nothing is not in the form, so it gets no
// __present sentinel and a save leaves it alone.
func TestReflectFormHandler_NothingRenderedIsLeftAlone(t *testing.T) {
	type record struct {
		Name    string
		Skipped string
	}
	var saved *record
	h := ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{Name: "n", Skipped: "keep"}, nil },
		func(_ context.Context, r *record) error { saved = r; return nil },
		hiddenKeyDecider(),
	)
	body := getBody(t, h)
	if strings.Contains(body, PresentSentinelName("Skipped")) {
		t.Errorf("a field rendered as nothing has a sentinel: %s", body)
	}
	// Posted as rendered, a required field rendered as nothing is neither
	// written nor demanded.
	type required struct {
		Name    string
		Skipped string `form:"required"`
	}
	var kept *required
	req := ReflectFormHandler(
		func(context.Context) (*required, error) { return &required{Name: "n"}, nil },
		func(_ context.Context, r *required) error { kept = r; return nil },
		hiddenKeyDecider(),
	)
	if rec := postForm(t, h, renderedValues(body)); rec.Code != http.StatusSeeOther || saved.Skipped != "keep" {
		t.Errorf("status %d, saved %+v", rec.Code, saved)
	}
	if rec := postForm(t, req, renderedValues(getBody(t, req))); rec.Code != http.StatusSeeOther || kept == nil {
		t.Errorf("a required field rendered as nothing blocks the save: status %d, body %s", rec.Code, rec.Body.String())
	}
}

// Each error shows once: inline where an input shows it, above the form
// otherwise, and one generic message for any number of internal ones.
func TestReflectFormHandler_ErrorsShowOnce(t *testing.T) {
	h := ReflectFormHandler(
		func(context.Context) (*hiddenKeyRecord, error) {
			return &hiddenKeyRecord{Name: "n", Lines: []hiddenKeyLine{{"k1", "a"}, {"k2", "b"}}}, nil
		},
		func(context.Context, *hiddenKeyRecord) error {
			return FieldErrors{
				"Name": errors.New("name taken"),
				FieldPath("Lines").Append("0").Append("Desc"): errors.New("desc too long"),
				"Lines": errors.New("too many lines"),
				FieldPath("Lines").Append("0").Append("Key"): errors.New("pq: row locked"),
				FieldPath("Lines").Append("1").Append("Key"): errors.New("pq: row gone"),
			}
		},
		hiddenKeyDecider(),
	)
	body := postForm(t, h, renderedValues(getBody(t, h))).Body.String()
	for _, msg := range []string{"name taken", "desc too long", "too many lines", FormErrorMessage} {
		if n := strings.Count(body, ">"+msg+"<"); n != 1 {
			t.Errorf("%q shows %d times: %s", msg, n, body)
		}
	}
}

// A field the decider does not parse is not written, not even cleared by a
// forged sentinel; a forged clear on a value without a null is an error;
// a nil FieldErrors joined with another error is that error, not success.
func TestReflectFormHandler_ClearAndFieldErrorsEdges(t *testing.T) {
	type record struct {
		Name    string
		Secret  string      `form:"sensitive"`
		Skipped nullableStr `form:"sensitive"`
	}
	saved := false
	h := ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{"n", "s", nullableStr{v: "keep"}}, nil },
		func(context.Context, *record) error { saved = true; return nil },
		func(path FieldPath, field reflect.StructField, value reflect.Value) FieldBehavior {
			if field.Name == "Skipped" {
				return FieldBehavior{}
			}
			return newTestDecider()(path, field, value)
		},
	)
	var got *record
	form := url.Values{}
	form.Set(PresentSentinelName("Skipped"), "1")
	form.Set(ClearSentinelName("Skipped"), "1")
	keepH := ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{"n", "s", nullableStr{v: "keep"}}, nil },
		func(_ context.Context, r *record) error { got = r; return nil },
		func(path FieldPath, field reflect.StructField, value reflect.Value) FieldBehavior {
			if field.Name == "Skipped" {
				return FieldBehavior{}
			}
			return newTestDecider()(path, field, value)
		},
	)
	if rec := postForm(t, keepH, form); rec.Code != http.StatusSeeOther || got.Skipped.IsNull() || got.Skipped.v != "keep" {
		t.Errorf("a forged clear on an unparsed field: status %d, saved %+v", rec.Code, got)
	}
	form = url.Values{}
	form.Set(PresentSentinelName("Name"), "1")
	form.Set("Name", "n")
	form.Set(ClearSentinelName("Name"), "1")
	if rec := postForm(t, h, form); rec.Code != http.StatusOK || saved || !strings.Contains(rec.Body.String(), "field is not nullable") {
		t.Errorf("a forged clear on a plain string: status %d, saved %t", rec.Code, saved)
	}
	joined := ReflectFormHandler(
		func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil },
		func(context.Context, *sampleStruct) error {
			return errors.Join(FormMessage("the plan is full"), FieldErrors{"Name": nil})
		},
		newTestDecider(),
	)
	form = url.Values{}
	form.Set(PresentSentinelName("Name"), "1")
	if rec := postForm(t, joined, form); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), ">the plan is full<") {
		t.Errorf("a nil FieldErrors joined with an error: status %d", rec.Code)
	}
	if msg := (FieldErrors{"Name": nil}).Error(); msg != "" {
		t.Errorf("FieldErrors of nil entries reads %q", msg)
	}
	if FormMaxMemory(context.Background()) != DefaultMaxMemory {
		t.Error("FormMaxMemory outside a handler must be DefaultMaxMemory")
	}
}

// The server removes a multipart body's temporary files, also of an upload
// past MaxMemory, which the parser keeps on disk.
func TestReflectFormHandler_MultipartTempFilesRemoved(t *testing.T) {
	h := ReflectFormHandlerWith(ReflectFormConfig{MaxMemory: 64},
		func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil },
		func(context.Context, *sampleStruct) error { return nil }, newTestDecider())
	type ctxKey struct{}
	// A middleware that passes a copy of the request on, as most do, hides
	// the parsed form from the server's own cleanup.
	copied := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, 1)))
	})
	for name, handler := range map[string]http.Handler{"mounted directly": h, "behind a middleware": copied} {
		dir := t.TempDir()
		t.Setenv("TMPDIR", dir)
		srv := httptest.NewServer(handler)
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		_ = mw.WriteField(PresentSentinelName("Name"), "1")
		part, _ := mw.CreateFormFile("Doc", "doc.bin")
		_, _ = part.Write(bytes.Repeat([]byte("x"), 4096))
		_ = mw.Close()
		resp, err := http.Post(srv.URL, mw.FormDataContentType(), &body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		srv.Close() // waits for the handler, so the server's cleanup has run
		if entries, _ := os.ReadDir(dir); len(entries) != 0 {
			t.Errorf("%s: multipart temp files left behind: %v", name, entries)
		}
	}
}

// A form posts urlencoded, whose limit is 10,000 values, unless a file
// input, also in a row, needs multipart.
func TestReflectFormHandler_Enctype(t *testing.T) {
	type fileLine struct {
		Doc []byte `form:"widget=file"`
	}
	type withRowFile struct {
		Lines []fileLine `form:"repeatable"`
	}
	type withFile struct {
		Doc []byte `form:"widget=file"`
	}
	for _, c := range []struct {
		name string
		body string
		want string
	}{
		{"no file", getBody(t, ReflectFormHandler(func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil },
			func(context.Context, *sampleStruct) error { return nil }, newTestDecider())), "application/x-www-form-urlencoded"},
		{"file", getBody(t, ReflectFormHandler(func(context.Context) (*withFile, error) { return &withFile{}, nil },
			func(context.Context, *withFile) error { return nil }, newTestDecider())), "multipart/form-data"},
		{"file in an empty repeatable", getBody(t, ReflectFormHandler(func(context.Context) (*withRowFile, error) { return &withRowFile{}, nil },
			func(context.Context, *withRowFile) error { return nil }, newTestDecider())), "multipart/form-data"},
	} {
		if !strings.Contains(c.body, `enctype="`+c.want+`"`) {
			t.Errorf("%s: %s, want enctype %s", c.name, c.body, c.want)
		}
	}
}

type requiredNullLine struct {
	Note nullableStr `form:"required"`
}

type requiredNullRecord struct {
	Note  nullableStr        `form:"required"`
	Lines []requiredNullLine `form:"repeatable"`
}

// The forms offer no clear control on a required field, so a POST that
// clears one is tampered: it is rejected as "required", not saved as null,
// at the top level and in a row.
func TestReflectFormHandler_ForgedClearKeepsRequired(t *testing.T) {
	saved := false
	h := ReflectFormHandler(
		func(context.Context) (*requiredNullRecord, error) {
			return &requiredNullRecord{Note: nullableStr{v: "n"}, Lines: []requiredNullLine{{nullableStr{v: "r"}}}}, nil
		},
		func(context.Context, *requiredNullRecord) error { saved = true; return nil },
		newTestDecider(),
	)
	top := url.Values{}
	top.Set(PresentSentinelName("Note"), "1")
	top.Set("Note", "n")
	top.Set(ClearSentinelName("Note"), "1")
	row := url.Values{}
	addRow(row, "Lines", 0, map[string]string{"Note": "r"})
	row.Set(ClearSentinelName("Lines-0-Note"), "1")
	for name, form := range map[string]url.Values{"top level": top, "row": row} {
		if rec := postForm(t, h, form); rec.Code != http.StatusOK || saved || !strings.Contains(rec.Body.String(), "required") {
			t.Errorf("%s: a forged clear on a required field: status %d, saved %t", name, rec.Code, saved)
		}
	}
}

// apiKey validates itself with a message that quotes its value, as
// go-enum's generated Validate does.
type apiKey string

func (k apiKey) Validate() error {
	if !strings.HasPrefix(string(k), "sk_") {
		return fmt.Errorf("invalid key %q", string(k))
	}
	return nil
}

type keyLine struct {
	Key apiKey `form:"sensitive"`
}

type keyRecord struct {
	Key   apiKey    `form:"sensitive"`
	Lines []keyLine `form:"repeatable"`
}

// A blank sensitive field keeps the stored value, which its validation
// then checks; the error must not quote that value.
func TestReflectFormHandler_SensitiveValidationHidesValue(t *testing.T) {
	h := ReflectFormHandler(
		func(context.Context) (*keyRecord, error) {
			return &keyRecord{Key: "legacy-SECRET", Lines: []keyLine{{"row-SECRET"}}}, nil
		},
		func(context.Context, *keyRecord) error { return nil },
		newTestDecider(),
	)
	form := url.Values{}
	form.Set(PresentSentinelName("Key"), "1")
	form.Set("Key", "")
	addRow(form, "Lines", 0, map[string]string{"Key": ""})
	// The test decider shows values; only the errors are checked here.
	rec := postForm(t, h, form)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || strings.Contains(body, "invalid key") || strings.Count(body, ">not a valid value<") != 2 {
		t.Errorf("status %d, body %s; want two masked errors", rec.Code, body)
	}
}

// pwPolicy is a sensitive value whose rule is meant for the user.
type pwPolicy string

func (p pwPolicy) Validate() error {
	if len(p) < 12 {
		// Joined with a message that quotes the value: only the
		// FormMessage may show.
		return errors.Join(FormMessage("must be at least 12 characters"), fmt.Errorf("%q is too short", string(p)))
	}
	return nil
}

// A sensitive field's FormMessage, such as a password rule, shows as it is;
// a nil entry in onSubmit's FieldErrors is no error.
func TestReflectFormHandler_SensitiveMessagesAndNilFieldErrors(t *testing.T) {
	type record struct {
		PW pwPolicy `form:"sensitive"`
	}
	h := ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{}, nil },
		func(context.Context, *record) error { return nil },
		newTestDecider(),
	)
	form := url.Values{}
	form.Set(PresentSentinelName("PW"), "1")
	form.Set("PW", "short")
	if body := postForm(t, h, form).Body.String(); !strings.Contains(body, ">must be at least 12 characters<") || strings.Contains(body, "too short") || strings.Contains(body, "not a valid value") {
		t.Errorf("a sensitive FormMessage must show as it is: %s", body)
	}
	saved := false
	nilErrs := ReflectFormHandler(
		func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil },
		func(context.Context, *sampleStruct) error { saved = true; return FieldErrors{"Name": nil} },
		newTestDecider(),
	)
	form = url.Values{}
	form.Set(PresentSentinelName("Name"), "1")
	form.Set("Name", "n")
	if rec := postForm(t, nilErrs, form); rec.Code != http.StatusSeeOther || !saved {
		t.Errorf("FieldErrors with only nil entries: status %d", rec.Code)
	}
}

// An error of a repeatable field itself, such as one asking for at least
// one row, shows inside its box, after the legend.
func TestReflectFormHandler_RepeatableFieldErrorShown(t *testing.T) {
	h := ReflectFormHandler(
		func(context.Context) (*keyRecord, error) { return &keyRecord{Key: "sk_1"}, nil },
		func(context.Context, *keyRecord) error {
			return FieldErrors{"Lines": errors.New("add at least one line")}
		},
		newTestDecider(),
	)
	form := url.Values{}
	form.Set(PresentSentinelName("Lines"), "1")
	if rec := postForm(t, h, form); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `<legend>Lines</legend><input type="hidden" name="__present__Lines" value="1"/><p class="form-error" role="alert" data-error="Lines">add at least one line</p>`) {
		t.Errorf("status %d, body %s; want the repeatable field's error after its legend", rec.Code, rec.Body.String())
	}
}

// Pressing Enter in an input clicks the form's first submit button, so a
// form whose rows render "Remove" buttons before Save starts with a default
// button that saves; a form without row buttons has none.
func TestReflectFormHandler_DefaultButtonSaves(t *testing.T) {
	firstButton := regexp.MustCompile(`<button[^>]*>`)
	for _, c := range []struct {
		name   string
		target any
		want   bool
	}{
		{"rows", &keyRecord{Key: "sk_1", Lines: []keyLine{{"sk_2"}}}, true},
		{"readonly rows", &lockedSheet{Lines: []approvalLine{{"a", "alice"}}}, false},
		{"no rows", &sampleStruct{Name: "n"}, false},
	} {
		var body string
		switch target := c.target.(type) {
		case *keyRecord:
			body = getBody(t, ReflectFormHandler(func(context.Context) (*keyRecord, error) { return target, nil },
				func(context.Context, *keyRecord) error { return nil }, newTestDecider()))
		case *lockedSheet:
			body = getBody(t, ReflectFormHandler(func(context.Context) (*lockedSheet, error) { return target, nil },
				func(context.Context, *lockedSheet) error { return nil }, newTestDecider()))
		case *sampleStruct:
			body = getBody(t, ReflectFormHandler(func(context.Context) (*sampleStruct, error) { return target, nil },
				func(context.Context, *sampleStruct) error { return nil }, newTestDecider()))
		}
		first := firstButton.FindString(body)
		if got := strings.Contains(first, `tabindex="-1"`) && !strings.Contains(first, "name="); got != c.want {
			t.Errorf("%s: first button %s, want default button %t", c.name, first, c.want)
		}
	}
}

func getBody(t *testing.T, h http.Handler) string {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec.Body.String()
}

// A hidden field shows no error of its own, so its error becomes a
// form-level message instead of failing the save silently.
func TestReflectFormHandler_HiddenFieldErrorShown(t *testing.T) {
	type record struct {
		Name string
		ID   int `form:"hidden"`
	}
	h := ReflectFormHandler(
		func(context.Context) (*record, error) { return &record{ID: 7}, nil },
		func(context.Context, *record) error { return nil },
		newTestDecider(),
	)
	form := url.Values{}
	form.Set(PresentSentinelName("Name"), "1")
	form.Set("Name", "n")
	form.Set(PresentSentinelName("ID"), "1")
	form.Set("ID", "x")
	rec := postForm(t, h, form)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `<p role="alert">`) {
		t.Errorf("status %d, body %s; want the hidden field's error as a form-level message", rec.Code, rec.Body.String())
	}
}

func TestReflectFormHandler_OptionalEnumAllowsBlank(t *testing.T) {
	type draft struct {
		Color strictEnum
	}
	var saved strictEnum
	handler := ReflectFormHandler(nil, func(_ context.Context, d *draft) error {
		saved = d.Color
		return nil
	}, newTestDecider())
	form := url.Values{}
	form.Set(PresentSentinelName("Color"), "1")
	form.Set("Color", "")
	req := httptest.NewRequest(http.MethodPost, "/draft", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("blank optional enum must save, status %d body %s", rec.Code, rec.Body.String())
	}
	if saved != "" {
		t.Fatalf("saved %q, want blank", saved)
	}
}

func TestReflectFormHandler_OptionalEnumBlankRunsDeciderValidate(t *testing.T) {
	type draft struct {
		Color strictEnum
	}
	base := newTestDecider()
	decider := func(path FieldPath, field reflect.StructField, value reflect.Value) FieldBehavior {
		b := base(path, field, value)
		b.Validate = func(FieldPath, reflect.StructField, reflect.Value) error {
			return errors.New("custom check")
		}
		return b
	}
	handler := ReflectFormHandler(nil, func(context.Context, *draft) error {
		t.Fatal("decider Validate must block the save")
		return nil
	}, decider)
	form := url.Values{}
	form.Set(PresentSentinelName("Color"), "1")
	form.Set("Color", "")
	req := httptest.NewRequest(http.MethodPost, "/draft", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200, body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "custom check") {
		t.Fatalf("decider Validate error missing: %s", rec.Body.String())
	}
}

func TestReflectFormHandlerRedirecting_NilOnSubmitPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for nil onSubmit")
		}
	}()
	load := func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil }
	_ = ReflectFormHandlerRedirecting[sampleStruct](ReflectFormConfig{}, load, nil)
}
