package mx

import (
	"context"
	"errors"
	stdhtml "html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

type lineItem struct {
	Desc string `form:"required"`
	Qty  int
}

type invoice struct {
	Title string
	Lines []lineItem `form:"repeatable,label=Lines"`
}

type invoicePtr struct {
	Lines []*lineItem `form:"repeatable"`
}

func TestDetectField_Repeatable(t *testing.T) {
	t.Run("slice of struct", func(t *testing.T) {
		field, _ := reflect.TypeFor[invoice]().FieldByName("Lines")
		kind, tag := DetectField("Lines", field, reflect.ValueOf(invoice{}).FieldByName("Lines"))
		if kind != FieldKindRepeatable {
			t.Errorf("kind=%s, want repeatable", kind)
		}
		if !tag.Repeatable {
			t.Errorf("tag.Repeatable=false")
		}
	})

	t.Run("slice of pointer to struct", func(t *testing.T) {
		field, _ := reflect.TypeFor[invoicePtr]().FieldByName("Lines")
		kind, _ := DetectField("Lines", field, reflect.ValueOf(invoicePtr{}).FieldByName("Lines"))
		if kind != FieldKindRepeatable {
			t.Errorf("kind=%s, want repeatable", kind)
		}
	})

	t.Run("repeatable tag on non-struct slice is ignored", func(t *testing.T) {
		type withStrings struct {
			Tags []string `form:"repeatable"`
		}
		field, _ := reflect.TypeFor[withStrings]().FieldByName("Tags")
		kind, _ := DetectField("Tags", field, reflect.ValueOf(withStrings{}).FieldByName("Tags"))
		if kind == FieldKindRepeatable {
			t.Errorf("[]string tagged repeatable must not be FieldKindRepeatable")
		}
		if kind != FieldKindTextarea {
			t.Errorf("kind=%s, want textarea ([]string fallback)", kind)
		}
	})

	// Regression: an interface element must NOT reach derefType (which
	// panics calling Elem() on an interface type).
	t.Run("slice of interface does not panic and is not repeatable", func(t *testing.T) {
		type rowIface interface{ Marker() }
		type withIface struct {
			Items []rowIface `form:"repeatable"`
		}
		field, _ := reflect.TypeFor[withIface]().FieldByName("Items")
		kind, _ := DetectField("Items", field, reflect.ValueOf(withIface{}).FieldByName("Items")) // must not panic
		if kind == FieldKindRepeatable {
			t.Errorf("[]interface tagged repeatable must not be FieldKindRepeatable")
		}
	})

	// Regression: multi-level pointer elements bind incorrectly, so they
	// must not be detected as repeatable (single pointer level only).
	t.Run("slice of double pointer is not repeatable", func(t *testing.T) {
		type withPP struct {
			Items []**lineItem `form:"repeatable"`
		}
		field, _ := reflect.TypeFor[withPP]().FieldByName("Items")
		kind, _ := DetectField("Items", field, reflect.ValueOf(withPP{}).FieldByName("Items"))
		if kind == FieldKindRepeatable {
			t.Errorf("[]**T tagged repeatable must not be FieldKindRepeatable")
		}
	})
}

func TestRepeatable_DeleteRowCommandSparseIndices(t *testing.T) {
	// Client removed the middle row's markup, so submitted indices are
	// 0 and 2. The Remove button for the row rendered at submitted index
	// 2 must delete the correct (compacted) row, not no-op or hit the
	// wrong one.
	var submitted bool
	h := ReflectFormHandler(
		func(context.Context) (*invoice, error) { return &invoice{}, nil },
		func(context.Context, *invoice) error { submitted = true; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Desc": "keep", "Qty": "1"})
	addRow(form, "Lines", 2, map[string]string{"Desc": "drop", "Qty": "2"})
	form.Set("__cmd__", DeleteRowCommand("Lines-2"))

	rec := postForm(t, h, form)
	if rec.Code != http.StatusOK || submitted {
		t.Fatalf("status=%d submitted=%v (want 200 re-render, no save)", rec.Code, submitted)
	}
	body := rec.Body.String()
	if strings.Contains(body, `value="drop"`) {
		t.Errorf("wrong row survived — 'drop' still present: %s", body)
	}
	if !strings.Contains(body, `value="keep"`) {
		t.Errorf("survivor 'keep' missing: %s", body)
	}
	if !strings.Contains(body, `data-row="Lines-0"`) || strings.Contains(body, `data-row="Lines-1"`) {
		t.Errorf("survivor not compacted to a single Lines-0 row: %s", body)
	}
}

func TestRepeatable_UnknownCommandFallsThroughToSave(t *testing.T) {
	// An injected/unknown __cmd__ that does not resolve to a repeatable
	// field must NOT suppress the save.
	t.Run("non-repeatable form", func(t *testing.T) {
		var captured *sampleStruct
		h := ReflectFormHandler(
			func(context.Context) (*sampleStruct, error) { return &sampleStruct{}, nil },
			func(_ context.Context, s *sampleStruct) error { captured = s; return nil },
			newTestDecider(),
		)
		form := url.Values{}
		form.Set(PresentSentinelName("Name"), "1")
		form.Set("Name", "saved")
		form.Set(PresentSentinelName("Age"), "1")
		form.Set("Age", "5")
		form.Set("__cmd__", "addrow:Lines") // no such field
		rec := postForm(t, h, form)
		if rec.Code != http.StatusSeeOther || captured == nil || captured.Name != "saved" {
			t.Fatalf("status=%d captured=%v — injected __cmd__ suppressed the save", rec.Code, captured)
		}
	})

	t.Run("repeatable form, garbage command", func(t *testing.T) {
		var submitted bool
		h := ReflectFormHandler(
			func(context.Context) (*invoice, error) { return &invoice{}, nil },
			func(context.Context, *invoice) error { submitted = true; return nil },
			newTestDecider(),
		)
		form := url.Values{}
		form.Set(PresentSentinelName("Title"), "1")
		form.Set("Title", "x")
		form.Set(PresentSentinelName("Lines"), "1") // field rendered, zero rows
		form.Set("__cmd__", "bogus-not-a-command")
		rec := postForm(t, h, form)
		if rec.Code != http.StatusSeeOther || !submitted {
			t.Fatalf("status=%d submitted=%v — garbage __cmd__ suppressed the save", rec.Code, submitted)
		}
	})
}

func TestRepeatable_MissingFieldPresentLeavesSliceUntouched(t *testing.T) {
	// A POST that does not carry the field-level __present marker (field
	// not rendered / tampered away) must leave the loaded slice intact,
	// not wipe it — the mass-assignment allowlist for repeatable fields.
	var got *invoice
	h := ReflectFormHandler(
		func(context.Context) (*invoice, error) {
			return &invoice{Lines: []lineItem{{Desc: "a"}, {Desc: "b"}}}, nil
		},
		func(_ context.Context, in *invoice) error { got = in; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	form.Set(PresentSentinelName("Title"), "1")
	form.Set("Title", "kept")
	// deliberately NO __present__Lines and no rows
	rec := postForm(t, h, form)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	want := []lineItem{{Desc: "a"}, {Desc: "b"}}
	if !reflect.DeepEqual(got.Lines, want) {
		t.Errorf("Lines=%+v, want untouched %+v (no __present should not wipe)", got.Lines, want)
	}
}

func TestRepeatable_NegativeRowIndexIgnored(t *testing.T) {
	var got *invoice
	h := ReflectFormHandler(
		func(context.Context) (*invoice, error) { return &invoice{}, nil },
		func(_ context.Context, in *invoice) error { got = in; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Desc": "real", "Qty": "1"})
	form.Set(RowSentinelName("Lines--1"), "1") // tampered negative index

	rec := postForm(t, h, form)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d", rec.Code)
	}
	if len(got.Lines) != 1 || got.Lines[0].Desc != "real" {
		t.Errorf("Lines=%+v, want exactly one {real 1} (negative index must be ignored)", got.Lines)
	}
}

func TestRepeatable_GetRendersRows(t *testing.T) {
	load := func(ctx context.Context) (*invoice, error) {
		return &invoice{
			Title: "Q1",
			Lines: []lineItem{{Desc: "Widget", Qty: 3}, {Desc: "Gadget", Qty: 7}},
		}, nil
	}
	h := ReflectFormHandler(load, func(context.Context, *invoice) error { return nil }, newTestDecider())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`data-repeatable="Lines"`,
		`data-row="Lines-0"`, `name="Lines-0-Desc"`, `value="Widget"`, `name="Lines-0-Qty"`, `value="3"`,
		`data-row="Lines-1"`, `name="Lines-1-Desc"`, `value="Gadget"`,
		`name="__row__Lines-0"`, `name="__row__Lines-1"`,
		`name="__present__Lines-0-Desc"`,
		`value="addrow:Lines"`, `value="delrow:Lines-0"`, `formnovalidate`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q\n%s", want, body)
		}
	}
}

func TestRepeatable_PostBindsRows(t *testing.T) {
	var got *invoice
	h := ReflectFormHandler(
		func(context.Context) (*invoice, error) { return &invoice{}, nil },
		func(_ context.Context, in *invoice) error { got = in; return nil },
		newTestDecider(),
	)

	form := url.Values{}
	form.Set(PresentSentinelName("Title"), "1")
	form.Set("Title", "Invoice A")
	addRow(form, "Lines", 0, map[string]string{"Desc": "First", "Qty": "2"})
	addRow(form, "Lines", 1, map[string]string{"Desc": "Second", "Qty": "5"})

	rec := postForm(t, h, form)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got == nil {
		t.Fatal("onSubmit not called")
	}
	if got.Title != "Invoice A" {
		t.Errorf("Title=%q", got.Title)
	}
	want := []lineItem{{Desc: "First", Qty: 2}, {Desc: "Second", Qty: 5}}
	if !reflect.DeepEqual(got.Lines, want) {
		t.Errorf("Lines=%+v, want %+v", got.Lines, want)
	}
}

func TestRepeatable_PostFewerRowsAfterClientRemoval(t *testing.T) {
	// Client dropped row 1's markup: only row 0's markers/values are
	// submitted, so the bound slice has exactly one element.
	var got *invoice
	h := ReflectFormHandler(
		func(context.Context) (*invoice, error) {
			return &invoice{Lines: []lineItem{{Desc: "old0"}, {Desc: "old1"}}}, nil
		},
		func(_ context.Context, in *invoice) error { got = in; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Desc": "kept", "Qty": "1"})

	rec := postForm(t, h, form)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(got.Lines) != 1 || got.Lines[0].Desc != "kept" {
		t.Errorf("Lines=%+v, want one {kept 1}", got.Lines)
	}
}

func TestRepeatable_AddRowCommand(t *testing.T) {
	var submitted bool
	h := ReflectFormHandler(
		func(context.Context) (*invoice, error) { return &invoice{}, nil },
		func(context.Context, *invoice) error { submitted = true; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Desc": "keepme", "Qty": "4"})
	form.Set("__cmd__", AddRowCommand("Lines"))

	rec := postForm(t, h, form)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d (want 200 re-render)", rec.Code)
	}
	if submitted {
		t.Error("onSubmit must not run for an add-row command")
	}
	body := rec.Body.String()
	if !strings.Contains(body, `value="keepme"`) {
		t.Errorf("entered value lost across add: %s", body)
	}
	if !strings.Contains(body, `data-row="Lines-1"`) || !strings.Contains(body, `name="__row__Lines-1"`) {
		t.Errorf("new empty row not rendered: %s", body)
	}
}

func TestRepeatable_DeleteRowCommand(t *testing.T) {
	var submitted bool
	h := ReflectFormHandler(
		func(context.Context) (*invoice, error) { return &invoice{}, nil },
		func(context.Context, *invoice) error { submitted = true; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Desc": "first", "Qty": "1"})
	addRow(form, "Lines", 1, map[string]string{"Desc": "second", "Qty": "2"})
	form.Set("__cmd__", DeleteRowCommand("Lines-0"))

	rec := postForm(t, h, form)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d (want 200 re-render)", rec.Code)
	}
	if submitted {
		t.Error("onSubmit must not run for a delete-row command")
	}
	body := rec.Body.String()
	if strings.Contains(body, `value="first"`) {
		t.Errorf("deleted row 0 still rendered: %s", body)
	}
	if !strings.Contains(body, `value="second"`) {
		t.Errorf("surviving row lost: %s", body)
	}
	// Survivor compacts to index 0.
	if !strings.Contains(body, `data-row="Lines-0"`) || strings.Contains(body, `data-row="Lines-1"`) {
		t.Errorf("survivor not compacted to Lines-0: %s", body)
	}
}

func TestRepeatable_PerRowRequiredValidation(t *testing.T) {
	h := ReflectFormHandler(
		func(context.Context) (*invoice, error) { return &invoice{}, nil },
		func(context.Context, *invoice) error {
			t.Fatal("onSubmit must not run when a row fails validation")
			return nil
		},
		newTestDecider(),
	)
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Desc": "", "Qty": "1"}) // Desc required, empty

	rec := postForm(t, h, form)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d (want 200 re-render)", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `data-error="Lines-0-Desc"`) || !strings.Contains(body, "required") {
		t.Errorf("expected inline required error at Lines-0-Desc: %s", body)
	}
}

func TestRepeatable_SparseIndicesDensifyErrorRouting(t *testing.T) {
	// Client submitted rows at indices 0 and 2 (row 1 removed). The bad
	// int in the index-2 row must route its error to the DENSE path
	// Lines-1-Qty so it lines up with the re-render.
	h := ReflectFormHandler(
		func(context.Context) (*invoice, error) { return &invoice{}, nil },
		func(context.Context, *invoice) error {
			t.Fatal("onSubmit must not run on parse error")
			return nil
		},
		newTestDecider(),
	)
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Desc": "ok", "Qty": "1"})
	addRow(form, "Lines", 2, map[string]string{"Desc": "bad", "Qty": "not-an-int"})

	rec := postForm(t, h, form)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `data-error="Lines-1-Qty"`) {
		t.Errorf("parse error not routed to dense Lines-1-Qty: %s", body)
	}
	if !strings.Contains(body, `data-row="Lines-1"`) || strings.Contains(body, `data-row="Lines-2"`) {
		t.Errorf("rows not densified on re-render: %s", body)
	}
}

func TestRepeatable_PointerElements(t *testing.T) {
	var got *invoicePtr
	h := ReflectFormHandler(
		func(context.Context) (*invoicePtr, error) { return &invoicePtr{}, nil },
		func(_ context.Context, in *invoicePtr) error { got = in; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Desc": "ptr", "Qty": "9"})

	rec := postForm(t, h, form)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(got.Lines) != 1 || got.Lines[0] == nil {
		t.Fatalf("Lines=%+v, want one non-nil element", got.Lines)
	}
	if got.Lines[0].Desc != "ptr" || got.Lines[0].Qty != 9 {
		t.Errorf("row=%+v", *got.Lines[0])
	}
}

// addRow writes a full row (the field-level __present marker, the row's
// __row__ marker naming the loaded row of that index as its origin, and
// each cell's __present sentinel and value) into form.
func addRow(form url.Values, field FieldPath, index int, cells map[string]string) {
	form.Set(PresentSentinelName(field), "1")
	rowPath := field.Append(strconv.Itoa(index))
	form.Set(RowSentinelName(rowPath), "o"+strconv.Itoa(index))
	for sub, val := range cells {
		cellPath := rowPath.Append(sub)
		form.Set(PresentSentinelName(cellPath), "1")
		form.Set(string(cellPath), val)
	}
}

var (
	renderedInputRe = regexp.MustCompile(`<input\b([^>]*)>`)
	renderedAttrRe  = regexp.MustCompile(`\b(name|value|type|checked)="([^"]*)"`)
)

// renderedValues returns what submitting the rendered form unchanged
// posts: the name and value of every <input> in body, except checkboxes
// and radios that are not checked.
func renderedValues(body string) url.Values {
	values := url.Values{}
	for _, input := range renderedInputRe.FindAllStringSubmatch(body, -1) {
		attrs := map[string]string{}
		for _, a := range renderedAttrRe.FindAllStringSubmatch(input[1], -1) {
			attrs[a[1]] = stdhtml.UnescapeString(a[2])
		}
		if typ := attrs["type"]; typ == "checkbox" || typ == "radio" {
			if _, checked := attrs["checked"]; !checked {
				continue
			}
		}
		if name, ok := attrs["name"]; ok {
			values.Add(name, attrs["value"])
		}
	}
	return values
}

type approvalLine struct {
	Desc       string
	ApprovedBy string `form:"readonly"`
}

type approvalSheet struct {
	Lines []approvalLine `form:"repeatable"`
}

// A readonly cell of a repeatable row is never parsed, yet keeps the value
// of the loaded row it came from: through the re-renders of a remove-row
// and an add-row click, on its own row after an earlier row was removed,
// and never on the added row, whatever that row posts for it.
func TestRepeatable_ReadonlyCellsKeepLoadedValues(t *testing.T) {
	stored := []approvalLine{{"a", "alice"}, {"b", "bob"}, {"c", "carol"}}
	var saved *approvalSheet
	h := ReflectFormHandler(
		func(context.Context) (*approvalSheet, error) {
			return &approvalSheet{Lines: append([]approvalLine(nil), stored...)}, nil
		},
		func(_ context.Context, s *approvalSheet) error { saved = s; return nil },
		newTestDecider(),
	)
	get := httptest.NewRecorder()
	h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/", nil))
	shown := renderedValues(get.Body.String())
	if got := shown.Get(RowSentinelName("Lines-2")); got != "o2" {
		t.Fatalf("GET marks row 2 with origin %q, want o2", got)
	}

	shown.Set("__cmd__", DeleteRowCommand("Lines-0"))
	removedBody := postForm(t, h, shown).Body.String()
	removed := renderedValues(removedBody)
	if removed.Get("Lines-0-ApprovedBy") != "bob" || removed.Get(RowSentinelName("Lines-0")) != "o1" {
		t.Fatalf("after removing row 0, row 0 must be bob's with origin o1: %s", removedBody)
	}

	removed.Set("__cmd__", AddRowCommand("Lines"))
	added := renderedValues(postForm(t, h, removed).Body.String())
	if got := added.Get(RowSentinelName("Lines-2")); got != "new" {
		t.Fatalf("the added row's origin is %q, want new", got)
	}
	added.Del("__cmd__")
	added.Set("Lines-2-Desc", "d")
	added.Set("Lines-2-ApprovedBy", "mallory")
	if rec := postForm(t, h, added); rec.Code != http.StatusSeeOther {
		t.Fatalf("Save: status %d, body %s", rec.Code, rec.Body.String())
	}
	want := []approvalLine{{"b", "bob"}, {"c", "carol"}, {"d", ""}}
	if saved == nil || !reflect.DeepEqual(saved.Lines, want) {
		t.Errorf("saved %+v, want %+v", saved, want)
	}
}

// The __row__ marker is client-controlled: a loaded row's readonly values
// go to at most one submitted row, and a marker naming no loaded row
// claims none, nor does the bare "1" of a page rendered before markers
// carried origins.
func TestRepeatable_TamperedOriginsClaimOnce(t *testing.T) {
	var saved *approvalSheet
	h := ReflectFormHandler(
		func(context.Context) (*approvalSheet, error) {
			return &approvalSheet{Lines: []approvalLine{{"a", "alice"}, {"b", "bob"}}}, nil
		},
		func(_ context.Context, s *approvalSheet) error { saved = s; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	for i, origin := range []string{"o0", "o0", "o7", "1", "x"} {
		addRow(form, "Lines", i, map[string]string{"Desc": strconv.Itoa(i)})
		form.Set(RowSentinelName(FieldPath("Lines").Append(strconv.Itoa(i))), origin)
	}
	if rec := postForm(t, h, form); rec.Code != http.StatusSeeOther {
		t.Fatalf("status %d", rec.Code)
	}
	want := []approvalLine{{"0", "alice"}, {"1", ""}, {"2", ""}, {"3", ""}, {"4", ""}}
	if saved == nil || !reflect.DeepEqual(saved.Lines, want) {
		t.Errorf("saved %+v, want %+v", saved, want)
	}
}

type reviewLine struct {
	Desc       string `form:"required"`
	ApprovedBy string `form:"readonly"`
}

type reviewSheet struct {
	Lines []reviewLine `form:"repeatable"`
}

// A save rejected after a remove-row click re-renders each row with its
// origin, so the corrected save keeps the readonly cells on their rows,
// whether a field, onSubmit's FieldErrors or onSubmit's error rejected it.
// Rows that onSubmit appended before failing claim no loaded row; the rows
// before them keep theirs.
func TestRepeatable_ReadonlyOriginsSurviveRejectedSave(t *testing.T) {
	stored := []reviewLine{{"a", "alice"}, {"b", "bob"}, {"c", "carol"}}
	for _, tt := range []struct {
		name        string
		blankDesc   bool
		firstSubmit func(*reviewSheet) error
		wantMarkers []string
		want        []reviewLine
	}{
		{
			name:        "required cell blank",
			blankDesc:   true,
			wantMarkers: []string{"o1", "o2"},
			want:        []reviewLine{{"b", "bob"}, {"c", "carol"}},
		},
		{
			name:        "onSubmit field errors",
			firstSubmit: func(*reviewSheet) error { return FieldErrors{"Lines-0-Desc": errors.New("taken")} },
			wantMarkers: []string{"o1", "o2"},
			want:        []reviewLine{{"b", "bob"}, {"c", "carol"}},
		},
		{
			name:        "onSubmit error",
			firstSubmit: func(*reviewSheet) error { return errors.New("try again") },
			wantMarkers: []string{"o1", "o2"},
			want:        []reviewLine{{"b", "bob"}, {"c", "carol"}},
		},
		{
			// Fewer rows may have shifted, so none claims a loaded row.
			name: "onSubmit removed a row",
			firstSubmit: func(s *reviewSheet) error {
				s.Lines = s.Lines[1:]
				return errors.New("try again")
			},
			wantMarkers: []string{"new"},
			want:        []reviewLine{{"b", ""}},
		},
		{
			name: "onSubmit added a row",
			firstSubmit: func(s *reviewSheet) error {
				s.Lines = append(s.Lines, reviewLine{Desc: "d"})
				return errors.New("try again")
			},
			wantMarkers: []string{"o1", "o2", "new"},
			want:        []reviewLine{{"b", "bob"}, {"c", "carol"}, {"d", ""}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			submits := 0
			var saved *reviewSheet
			h := ReflectFormHandler(
				func(context.Context) (*reviewSheet, error) {
					return &reviewSheet{Lines: append([]reviewLine(nil), stored...)}, nil
				},
				func(_ context.Context, s *reviewSheet) error {
					submits++
					if submits == 1 && tt.firstSubmit != nil {
						return tt.firstSubmit(s)
					}
					saved = s
					return nil
				},
				newTestDecider(),
			)
			get := httptest.NewRecorder()
			h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/", nil))
			shown := renderedValues(get.Body.String())
			shown.Set("__cmd__", DeleteRowCommand("Lines-0"))
			removed := renderedValues(postForm(t, h, shown).Body.String())
			removed.Del("__cmd__")
			if tt.blankDesc {
				removed.Set("Lines-0-Desc", "")
			}

			rejected := postForm(t, h, removed)
			if rejected.Code != http.StatusOK || saved != nil {
				t.Fatalf("first save: status %d, saved %+v; want a 200 re-render", rejected.Code, saved)
			}
			again := renderedValues(rejected.Body.String())
			for i, want := range tt.wantMarkers {
				if got := again.Get(RowSentinelName(FieldPath("Lines").Append(strconv.Itoa(i)))); got != want {
					t.Errorf("row %d origin %q, want %q", i, got, want)
				}
			}
			again.Set("Lines-0-Desc", "b")
			if rec := postForm(t, h, again); rec.Code != http.StatusSeeOther {
				t.Fatalf("corrected save: status %d, body %s", rec.Code, rec.Body.String())
			}
			if saved == nil || !reflect.DeepEqual(saved.Lines, tt.want) {
				t.Errorf("saved %+v, want %+v", saved, tt.want)
			}
		})
	}
}

// A rebuilt row starts as a copy of its origin row: a cell the client did
// not send keeps its loaded value instead of becoming zero.
func TestRepeatable_RowCellsStartFromOrigin(t *testing.T) {
	var saved *approvalSheet
	h := ReflectFormHandler(
		func(context.Context) (*approvalSheet, error) {
			return &approvalSheet{Lines: []approvalLine{{"a", "alice"}}}, nil
		},
		func(_ context.Context, s *approvalSheet) error { saved = s; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{}) // no cells sent
	if rec := postForm(t, h, form); rec.Code != http.StatusSeeOther {
		t.Fatalf("status %d", rec.Code)
	}
	if want := []approvalLine{{"a", "alice"}}; saved == nil || !reflect.DeepEqual(saved.Lines, want) {
		t.Errorf("saved %+v, want %+v", saved, want)
	}
}

type lockedSheet struct {
	Lines []approvalLine `form:"repeatable,readonly"`
}

type auditPart struct {
	Lines []approvalLine `form:"repeatable"`
}

type auditedSheet struct {
	Audit auditPart `form:"section=Audit,readonly"`
}

// form:"readonly" on the repeatable field renders its rows readonly with no
// add or remove buttons, and a tampered POST never rebuilds them.
func TestRepeatable_ReadonlyFieldKeepsRows(t *testing.T) {
	var saved *lockedSheet
	h := ReflectFormHandler(
		func(context.Context) (*lockedSheet, error) {
			return &lockedSheet{Lines: []approvalLine{{"a", "alice"}}}, nil
		},
		func(_ context.Context, s *lockedSheet) error { saved = s; return nil },
		newTestDecider(),
	)
	get := httptest.NewRecorder()
	h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/", nil))
	if body := get.Body.String(); strings.Contains(body, "__cmd__") {
		t.Errorf("a readonly repeatable field offers add or remove buttons: %s", body)
	}
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Desc": "tampered"})
	addRow(form, "Lines", 1, map[string]string{"Desc": "injected"})
	if rec := postForm(t, h, form); rec.Code != http.StatusSeeOther {
		t.Fatalf("status %d", rec.Code)
	}
	if want := []approvalLine{{"a", "alice"}}; saved == nil || !reflect.DeepEqual(saved.Lines, want) {
		t.Errorf("saved %+v, want %+v", saved, want)
	}
	// A forged row command on a readonly field is a plain submit.
	form.Set("__cmd__", AddRowCommand("Lines"))
	if rec := postForm(t, h, form); rec.Code != http.StatusSeeOther || len(saved.Lines) != 1 {
		t.Errorf("forged add-row: status %d, saved %+v", rec.Code, saved)
	}
}

// A repeatable inside a readonly section inherits readonly: no row
// buttons, its rows are not rebuilt from the submission, and a forged row
// command is a plain submit.
func TestRepeatable_ReadonlySectionKeepsRows(t *testing.T) {
	var saved *auditedSheet
	h := ReflectFormHandler(
		func(context.Context) (*auditedSheet, error) {
			return &auditedSheet{Audit: auditPart{Lines: []approvalLine{{"a", "alice"}}}}, nil
		},
		func(_ context.Context, s *auditedSheet) error { saved = s; return nil },
		newTestDecider(),
	)
	get := httptest.NewRecorder()
	h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/", nil))
	if body := get.Body.String(); strings.Contains(body, "__cmd__") {
		t.Errorf("a repeatable in a readonly section offers add or remove buttons: %s", body)
	}
	lines := FieldPath("Audit").Append("Lines")
	form := url.Values{}
	addRow(form, lines, 0, map[string]string{"Desc": "tampered"})
	addRow(form, lines, 1, map[string]string{"Desc": "injected"})
	form.Set("__cmd__", AddRowCommand(lines))
	if rec := postForm(t, h, form); rec.Code != http.StatusSeeOther {
		t.Fatalf("status %d", rec.Code)
	}
	if want := []approvalLine{{"a", "alice"}}; saved == nil || !reflect.DeepEqual(saved.Audit.Lines, want) {
		t.Errorf("saved %+v, want %+v", saved, want)
	}
}

type pointerSheet struct {
	Lines []*approvalLine `form:"repeatable"`
}

// Rows of pointer elements start as a copy of their origin row too, a nil
// one as the zero row, and the loaded pointees are never written.
func TestRepeatable_PointerRowsStartFromOrigin(t *testing.T) {
	first := &approvalLine{"a", "alice"}
	var saved *pointerSheet
	h := ReflectFormHandler(
		func(context.Context) (*pointerSheet, error) {
			return &pointerSheet{Lines: []*approvalLine{first, nil}}, nil
		},
		func(_ context.Context, s *pointerSheet) error { saved = s; return nil },
		newTestDecider(),
	)
	form := url.Values{}
	addRow(form, "Lines", 0, map[string]string{"Desc": "a2", "ApprovedBy": "mallory"})
	addRow(form, "Lines", 1, map[string]string{"Desc": "b"})
	if rec := postForm(t, h, form); rec.Code != http.StatusSeeOther {
		t.Fatalf("status %d", rec.Code)
	}
	if saved == nil || len(saved.Lines) != 2 || *saved.Lines[0] != (approvalLine{"a2", "alice"}) || *saved.Lines[1] != (approvalLine{"b", ""}) {
		t.Errorf("saved %+v", saved)
	}
	if *first != (approvalLine{"a", "alice"}) {
		t.Errorf("the loaded row was written: %+v", *first)
	}
}

func postForm(t *testing.T, h http.Handler, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
