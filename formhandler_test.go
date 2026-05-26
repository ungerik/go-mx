package mx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
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
	if !strings.Contains(body, `method="POST"`) {
		t.Errorf("missing method=POST: %q", body)
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
