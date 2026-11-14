package mx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// germanFormTexts translates every field, so a test sees each one replaced.
var germanFormTexts = FormTexts{
	Submit:          "Speichern",
	Error:           "Das Formular konnte nicht gespeichert werden.",
	RetypeSensitive: "Bitte das verborgene Feld erneut eingeben.",
	AddRow:          "Zeile hinzufügen",
	RemoveRow:       "Entfernen",
	NoRows:          "Keine",
}

// withFormTexts installs texts in every request's context, as an
// application's language middleware would.
func withFormTexts(h http.Handler, texts FormTexts) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(ContextWithFormTexts(r.Context(), texts)))
	})
}

// orDefaults fills every field: the zero FormTexts resolves to the built-in
// English, and a complete FormTexts is returned unchanged. It fails when a
// new field is added to FormTexts but not to orDefaults.
func TestFormTextsOrDefaultsCoversAllFields(t *testing.T) {
	got := reflect.ValueOf(FormTexts{}.orDefaults())
	for i := range got.NumField() {
		if got.Field(i).String() == "" {
			t.Errorf("%s resolves to an empty text", got.Type().Field(i).Name)
		}
	}
	if germanFormTexts.orDefaults() != germanFormTexts {
		t.Errorf("a complete FormTexts was changed: %+v", germanFormTexts.orDefaults())
	}
}

// A context without FormTexts, also a nil one, yields DefaultFormTexts, and
// a partial FormTexts keeps the defaults for the fields it leaves empty.
func TestFormTextsFromContext(t *testing.T) {
	if got := FormTextsFromContext(context.Background()); got != DefaultFormTexts {
		t.Errorf("no texts installed: got %+v", got)
	}
	if got := FormTextsFromContext(nil); got != DefaultFormTexts { //nolint:staticcheck // a nil context must not panic
		t.Errorf("nil context: got %+v", got)
	}
	got := FormTextsFromContext(ContextWithFormTexts(context.Background(), FormTexts{AddRow: "Zeile hinzufügen"}))
	if got.AddRow != "Zeile hinzufügen" || got.RemoveRow != DefaultFormTexts.RemoveRow {
		t.Errorf("partial texts: got %+v", got)
	}
}

// A partial DefaultFormTexts assignment keeps the built-in English for the
// fields it leaves empty instead of rendering an empty caption.
func TestFormTextsPartialDefault(t *testing.T) {
	saved := DefaultFormTexts
	t.Cleanup(func() { DefaultFormTexts = saved })
	DefaultFormTexts = FormTexts{Submit: "Absenden"}
	got := FormTextsFromContext(context.Background())
	if got.Submit != "Absenden" || got.AddRow != englishFormTexts.AddRow {
		t.Errorf("got %+v", got)
	}
}

// A reflected form takes every text it renders on its own from the
// request's context: the submit caption, the row buttons, the empty readonly
// line, the generic onSubmit error and the retype notice.
func TestReflectFormHandler_FormTextsFromContext(t *testing.T) {
	t.Run("submit and row buttons", func(t *testing.T) {
		h := withFormTexts(ReflectFormHandler(
			func(context.Context) (*invoice, error) {
				return &invoice{Lines: []lineItem{{Desc: "a"}}}, nil
			},
			func(context.Context, *invoice) error { return nil },
			newTestDecider(),
		), germanFormTexts)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		body := rec.Body.String()
		for _, text := range []string{">Speichern<", ">Zeile hinzufügen<", ">Entfernen<"} {
			if !strings.Contains(body, text) {
				t.Errorf("missing %q in %s", text, body)
			}
		}
		for _, text := range []string{">Save<", ">Add row<", ">Remove<"} {
			if strings.Contains(body, text) {
				t.Errorf("English %q left in %s", text, body)
			}
		}
	})

	t.Run("SubmitLabel still wins", func(t *testing.T) {
		h := withFormTexts(ReflectFormHandlerWith(
			ReflectFormConfig{SubmitLabel: "Rechnung anlegen"},
			func(context.Context) (*invoice, error) { return &invoice{}, nil },
			func(context.Context, *invoice) error { return nil },
			newTestDecider(),
		), germanFormTexts)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if body := rec.Body.String(); !strings.Contains(body, ">Rechnung anlegen<") || strings.Contains(body, ">Speichern<") {
			t.Errorf("want the handler's SubmitLabel: %s", body)
		}
	})

	t.Run("readonly field without rows", func(t *testing.T) {
		h := withFormTexts(ReflectFormHandler(
			func(context.Context) (*lockedSheet, error) { return &lockedSheet{}, nil },
			func(context.Context, *lockedSheet) error { return nil },
			newTestDecider(),
		), germanFormTexts)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if body := rec.Body.String(); !strings.Contains(body, `<p data-repeatable-empty="Lines">Keine</p>`) {
			t.Errorf("want the translated no-rows line: %s", body)
		}
	})

	t.Run("generic onSubmit error", func(t *testing.T) {
		h := withFormTexts(ReflectFormHandler(
			func(context.Context) (*invoice, error) { return &invoice{}, nil },
			func(context.Context, *invoice) error { return errors.New("pq: connection refused") },
			newTestDecider(),
		), germanFormTexts)
		form := url.Values{}
		form.Set(PresentSentinelName("Title"), "1")
		form.Set("Title", "x")
		body := postForm(t, h, form).Body.String()
		if !strings.Contains(body, germanFormTexts.Error) || strings.Contains(body, "pq:") {
			t.Errorf("want the translated generic error: %s", body)
		}
	})

	t.Run("retype notice", func(t *testing.T) {
		type record struct {
			Lines []sensitiveLine `form:"repeatable"`
		}
		h := withFormTexts(ReflectFormHandler(
			func(context.Context) (*record, error) { return &record{Lines: []sensitiveLine{{}}}, nil },
			func(context.Context, *record) error { return nil },
			newTestDecider(),
		), germanFormTexts)
		form := url.Values{}
		addRow(form, "Lines", 0, map[string]string{"Token": "typed"})
		form.Set(sentinelCommand, AddRowCommand("Lines"))
		if body := postForm(t, h, form).Body.String(); !strings.Contains(body, germanFormTexts.RetypeSensitive) {
			t.Errorf("want the translated retype notice: %s", body)
		}
	})
}
