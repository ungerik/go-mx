package mx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestMiddleware_InstallsDecider(t *testing.T) {
	tag := "from-test"
	decider := func(p FieldPath, f reflect.StructField, v reflect.Value) FieldBehavior {
		return FieldBehavior{
			Render: func(p FieldPath, f reflect.StructField, v reflect.Value, errs []error) Component {
				return Text(tag)
			},
		}
	}

	var observed FieldDecider
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed = DeciderFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	srv := Middleware(decider)(inner)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if observed == nil {
		t.Fatal("DeciderFromContext returned nil — middleware did not install")
	}
	beh := observed("X", reflect.StructField{}, reflect.Value{})
	got := renderToString(t, beh.Render("X", reflect.StructField{}, reflect.Value{}, nil))
	if got != tag {
		t.Errorf("rendered %q, want %q", got, tag)
	}
}

func TestMiddleware_NestedOverrides(t *testing.T) {
	outer := stringDecider("outer")
	inner := stringDecider("inner")

	var seen string
	terminal := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := DeciderFromContext(r.Context())
		seen = renderToString(t, d("X", reflect.StructField{}, reflect.Value{}).Render("X", reflect.StructField{}, reflect.Value{}, nil))
		w.WriteHeader(http.StatusOK)
	})

	stack := Middleware(outer)(Middleware(inner)(terminal))
	rec := httptest.NewRecorder()
	stack.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if seen != "inner" {
		t.Errorf("nested middleware: terminal saw %q, want inner", seen)
	}
}

func TestContextWithDecider(t *testing.T) {
	d := stringDecider("ctx-direct")
	ctx := ContextWithDecider(context.Background(), d)
	got := DeciderFromContext(ctx)
	if got == nil {
		t.Fatal("expected decider, got nil")
	}
	beh := got("X", reflect.StructField{}, reflect.Value{})
	if renderToString(t, beh.Render("X", reflect.StructField{}, reflect.Value{}, nil)) != "ctx-direct" {
		t.Errorf("wrong decider retrieved")
	}
}

func stringDecider(tag string) FieldDecider {
	return func(p FieldPath, f reflect.StructField, v reflect.Value) FieldBehavior {
		return FieldBehavior{
			Render: func(p FieldPath, f reflect.StructField, v reflect.Value, errs []error) Component {
				return Text(tag)
			},
		}
	}
}

func renderToString(t *testing.T, c Component) string {
	t.Helper()
	if c == nil {
		return ""
	}
	var buf testBuf
	w := NewCheckedWriter(&buf)
	if err := c.Render(context.Background(), w); err != nil {
		t.Fatalf("render: %v", err)
	}
	return buf.String()
}

type testBuf struct{ b []byte }

func (t *testBuf) Write(p []byte) (int, error) {
	t.b = append(t.b, p...)
	return len(p), nil
}

func (t *testBuf) String() string { return string(t.b) }
