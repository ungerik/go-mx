package shadcn

import (
	"context"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/ungerik/go-mx"
)

// renderWith renders a component with ctx and a double-quote CheckedWriter,
// failing the test on any render error. It is the [render] variant used to
// prove that the built-in labels are resolved from the render context.
func renderWith(t *testing.T, ctx context.Context, c mx.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(ctx, mx.NewCheckedWriter(&b)); err != nil {
		t.Fatalf("render error: %v\npartial output:\n%s", err, b.String())
	}
	return b.String()
}

// germanLabels is a fully translated [Labels] used to prove that every
// built-in string is reachable. It must stay complete: an unset field would
// silently fall back to the English DefaultLabels instead of failing loudly,
// which is what TestLabelsFieldsAllCovered checks.
var germanLabels = Labels{
	BreadcrumbNav:          "Brotkrümelnavigation",
	BreadcrumbEllipsis:     "Mehr Einträge",
	CarouselPrevious:       "Vorheriges Bild",
	CarouselNext:           "Nächstes Bild",
	DialogClose:            "Schließen",
	NavigationMenuNav:      "Hauptnavigation",
	PaginationNav:          "Seitennummerierung",
	PaginationPrevious:     "Zurück",
	PaginationNext:         "Weiter",
	PaginationPreviousPage: "Zur vorherigen Seite",
	PaginationNextPage:     "Zur nächsten Seite",
	PaginationEllipsis:     "Mehr Seiten",
	SidebarTrigger:         "Seitenleiste umschalten",
	Spinner:                "Lädt",
	FormClear:              "leeren",
	FormValueNotAvailable:  "(aktueller Wert nicht verfügbar)",
}

// labelCase pairs a component that renders built-in strings with the exact
// output fragments it must show without configured Labels (english) and with
// [germanLabels] installed (german). The fragments are full attributes or
// text runs rather than bare words so that an assertion cannot be satisfied by
// an unrelated part of the markup such as the data-slot name.
type labelCase struct {
	name    string
	comp    func() mx.Component
	fields  []string // Labels fields this case exercises
	english []string
	german  []string
	// wrapCtx decorates the render context for cases that need more than
	// Labels in it — the reflected select reads its per-request option list
	// from the context too.
	wrapCtx func(context.Context) context.Context
}

// labelCases covers every component of this package that renders a
// user-visible string the caller did not pass. TestLabelsFieldsAllCovered
// fails if a [Labels] field is not listed by any case here.
var labelCases = []labelCase{
	{
		name: "Breadcrumb", comp: func() mx.Component { return Breadcrumb() },
		fields:  []string{"BreadcrumbNav"},
		english: []string{`aria-label="breadcrumb"`},
		german:  []string{`aria-label="Brotkrümelnavigation"`},
	}, {
		name: "BreadcrumbEllipsis", comp: func() mx.Component { return BreadcrumbEllipsis() },
		fields:  []string{"BreadcrumbEllipsis"},
		english: []string{`class="sr-only">More<`},
		german:  []string{`class="sr-only">Mehr Einträge<`},
	}, {
		name: "CarouselPrevious", comp: func() mx.Component { return CarouselPrevious() },
		fields:  []string{"CarouselPrevious"},
		english: []string{`aria-label="Previous slide"`},
		german:  []string{`aria-label="Vorheriges Bild"`},
	}, {
		name: "CarouselNext", comp: func() mx.Component { return CarouselNext() },
		fields:  []string{"CarouselNext"},
		english: []string{`aria-label="Next slide"`},
		german:  []string{`aria-label="Nächstes Bild"`},
	}, {
		name: "DialogContent", comp: func() mx.Component { return DialogContent("dlg") },
		fields:  []string{"DialogClose"},
		english: []string{`class="sr-only">Close<`},
		german:  []string{`class="sr-only">Schließen<`},
	}, {
		name: "SheetContent", comp: func() mx.Component { return SheetContent("sht", SheetRight) },
		fields:  []string{"DialogClose"},
		english: []string{`class="sr-only">Close<`},
		german:  []string{`class="sr-only">Schließen<`},
	}, {
		name: "NavigationMenu", comp: func() mx.Component { return NavigationMenu() },
		fields:  []string{"NavigationMenuNav"},
		english: []string{`aria-label="Main"`},
		german:  []string{`aria-label="Hauptnavigation"`},
	}, {
		name: "Pagination", comp: func() mx.Component { return Pagination() },
		fields:  []string{"PaginationNav"},
		english: []string{`aria-label="pagination"`},
		german:  []string{`aria-label="Seitennummerierung"`},
	}, {
		name: "PaginationPrevious", comp: func() mx.Component { return PaginationPrevious() },
		fields:  []string{"PaginationPreviousPage", "PaginationPrevious"},
		english: []string{`aria-label="Go to previous page"`, `>Previous<`},
		german:  []string{`aria-label="Zur vorherigen Seite"`, `>Zurück<`},
	}, {
		name: "PaginationNext", comp: func() mx.Component { return PaginationNext() },
		fields:  []string{"PaginationNextPage", "PaginationNext"},
		english: []string{`aria-label="Go to next page"`, `>Next<`},
		german:  []string{`aria-label="Zur nächsten Seite"`, `>Weiter<`},
	}, {
		name: "PaginationEllipsis", comp: func() mx.Component { return PaginationEllipsis() },
		fields:  []string{"PaginationEllipsis"},
		english: []string{`class="sr-only">More pages<`},
		german:  []string{`class="sr-only">Mehr Seiten<`},
	}, {
		name: "SidebarTrigger", comp: func() mx.Component { return SidebarTrigger() },
		fields:  []string{"SidebarTrigger"},
		english: []string{`aria-label="Toggle Sidebar"`},
		german:  []string{`aria-label="Seitenleiste umschalten"`},
	}, {
		name: "Spinner", comp: func() mx.Component { return Spinner() },
		fields:  []string{"Spinner"},
		english: []string{`aria-label="Loading"`},
		german:  []string{`aria-label="Lädt"`},
	}, {
		// The reflected-form strings are built-in labels too: the caller
		// never passes them, so before Labels they were English-only in an
		// otherwise translated form.
		name: "FormClearControl", comp: func() mx.Component { return clearControl(mx.FieldPath("Optional")) },
		fields:  []string{"FormClear"},
		english: []string{`text-sm">clear<`},
		german:  []string{`text-sm">leeren<`},
	}, {
		name: "FormSelectOutOfListPlaceholder", comp: outOfListSelect,
		fields:  []string{"FormValueNotAvailable"},
		english: []string{`selected="selected">(current value not available)<`},
		german:  []string{`selected="selected">(aktueller Wert nicht verfügbar)<`},
		wrapCtx: func(ctx context.Context) context.Context {
			return context.WithValue(ctx, shadcnPartnersKey{},
				[]mx.NamedOption{{Name: "Partner One", Value: "p1"}})
		},
	},
}

// outOfListSelect builds the reflected select whose current value is missing
// from the per-request option list, which is the only way to reach the
// FormValueNotAvailable placeholder.
func outOfListSelect() mx.Component {
	type draftForm struct {
		Partner string `form:"options=test-shadcn-partners"`
	}
	s := &draftForm{Partner: "gone"}
	v := reflect.ValueOf(s).Elem()
	f, _ := v.Type().FieldByName("Partner")
	beh := FieldDecider(mx.FieldPath("Partner"), f, v.Field(0))
	return beh.Render(mx.FieldPath("Partner"), f, v.Field(0), nil)
}

// TestLabelsLocalize is the point of the whole mechanism: a component built
// without any localization argument renders the context's language, so a
// handler tree serves German by installing German Labels once — no call site
// changes.
func TestLabelsLocalize(t *testing.T) {
	ctx := ContextWithLabels(context.Background(), germanLabels)
	for _, tt := range labelCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ctx
			if tt.wrapCtx != nil {
				ctx = tt.wrapCtx(ctx)
			}
			out := renderWith(t, ctx, tt.comp())
			for _, want := range tt.german {
				if !strings.Contains(out, want) {
					t.Errorf("missing German %q in %s", want, out)
				}
			}
			// A surviving English fragment would mean the label is still
			// hardcoded somewhere in the component.
			for _, notWant := range tt.english {
				if strings.Contains(out, notWant) {
					t.Errorf("English %q survived localization in %s", notWant, out)
				}
			}
		})
	}
}

// TestLabelsDefault pins the English defaults: rendering without installed
// Labels must keep producing the upstream shadcn/ui strings, so adding the
// mechanism is not a behavior change for existing callers.
func TestLabelsDefault(t *testing.T) {
	for _, tt := range labelCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.wrapCtx != nil {
				ctx = tt.wrapCtx(ctx)
			}
			out := renderWith(t, ctx, tt.comp())
			for _, want := range tt.english {
				if !strings.Contains(out, want) {
					t.Errorf("missing English default %q in %s", want, out)
				}
			}
		})
	}
}

// TestLabelsResolvedAtRenderTime guards the property that makes one process
// able to serve several languages concurrently: the language is not baked into
// the element at construction time, so one built tree renders differently per
// request context.
func TestLabelsResolvedAtRenderTime(t *testing.T) {
	dialog := DialogContent("dlg")

	german := renderWith(t, ContextWithLabels(context.Background(), germanLabels), dialog)
	if !strings.Contains(german, "Schließen") {
		t.Errorf("missing German close label in %s", german)
	}

	english := render(t, dialog)
	if !strings.Contains(english, "Close") {
		t.Errorf("the same element must render English again for a context without Labels: %s", english)
	}
}

// TestLabelsCallerAttribStillWins keeps the per-call-site escape hatch working:
// an aria-label passed by the caller must beat the context label, because a
// single control sometimes needs a name the ambient translation cannot know.
func TestLabelsCallerAttribStillWins(t *testing.T) {
	ctx := ContextWithLabels(context.Background(), germanLabels)
	out := renderWith(t, ctx, Spinner(mx.NewAttrib("aria-label", "Speichert")))
	if !strings.Contains(out, `aria-label="Speichert"`) || strings.Contains(out, "Lädt") {
		t.Errorf("caller aria-label should win over the context label: %s", out)
	}
}

// TestLabelsCallerAttribWinsViaDedup covers the second override mechanism:
// PaginationPrevious/Next put their label attrib FIRST and prepend it to the
// caller's, so the caller only wins through finish's last-value-wins dedup —
// a different path from the AttribIndex guard the other components use.
func TestLabelsCallerAttribWinsViaDedup(t *testing.T) {
	ctx := ContextWithLabels(context.Background(), germanLabels)
	out := renderWith(t, ctx, PaginationPrevious(mx.NewAttrib("aria-label", "Eine Seite zurück")))
	if !strings.Contains(out, `aria-label="Eine Seite zurück"`) {
		t.Errorf("caller aria-label should win over the context label: %s", out)
	}
	if strings.Contains(out, `aria-label="Zur vorherigen Seite"`) {
		t.Errorf("the context label must not survive as a second aria-label: %s", out)
	}
}

// TestLabelsAreEscaped closes the trust boundary: Labels come from translation
// data, which is content, not markup. A translator writing an ampersand or an
// angle bracket must not be able to inject markup or break out of an attribute
// value. Text labels go through Writer.EscapeText and attribute labels through
// the writer's attribute escaper (checkedwriter.go), so both are covered — this
// test pins that neither path is ever switched to a raw write.
func TestLabelsAreEscaped(t *testing.T) {
	ctx := ContextWithLabels(context.Background(), Labels{
		DialogClose: `Close <b>&</b> "done"`,
		Spinner:     `Loading <img> & "wait"`,
	})

	text := renderWith(t, ctx, DialogContent("dlg"))
	if strings.Contains(text, "<b>") || !strings.Contains(text, "&lt;b&gt;") {
		t.Errorf("text label must be escaped, not written as markup: %s", text)
	}

	attr := renderWith(t, ctx, Spinner())
	// The attribute escaper escapes & < and the quote character, but leaves >
	// alone — a bare > cannot terminate a quoted attribute value, so it needs
	// no escape. What matters is that < can never start a tag.
	if strings.Contains(attr, "<img") || !strings.Contains(attr, "&lt;img") {
		t.Errorf("attribute label must be escaped, not written as markup: %s", attr)
	}
	// An unescaped double quote would close the attribute value early and turn
	// the rest of the label into attributes.
	if !strings.Contains(attr, "&quot;wait&quot;") {
		t.Errorf("double quotes in an attribute label must be escaped: %s", attr)
	}
}

// TestPartialDefaultLabelsNeverRendersEmpty is the regression test for the
// trap in a self-referential fallback: DefaultLabels is both the thing callers
// assign and the thing empty fields fall back to, so a partial assignment would
// make every unset field fall back to its own empty value. A control whose only
// accessible name is an aria-label would then render aria-label="" — the exact
// failure the fallback was added to prevent. englishLabels is the frozen last
// resort that makes a partial assignment safe.
func TestPartialDefaultLabelsNeverRendersEmpty(t *testing.T) {
	saved := DefaultLabels
	t.Cleanup(func() { DefaultLabels = saved })
	DefaultLabels = Labels{DialogClose: "Schließen"} // only one field set

	if out := render(t, Spinner()); !strings.Contains(out, `aria-label="Loading"`) {
		t.Errorf("an unset DefaultLabels field must fall through to English, not render empty: %s", out)
	}
	if out := render(t, DialogContent("dlg")); !strings.Contains(out, `class="sr-only">Schließen<`) {
		t.Errorf("the one field that was set must still win: %s", out)
	}
	// A context Labels still overrides the partial DefaultLabels.
	ctx := ContextWithLabels(context.Background(), Labels{Spinner: "Lädt"})
	if out := renderWith(t, ctx, Spinner()); !strings.Contains(out, `aria-label="Lädt"`) {
		t.Errorf("context Labels must win over DefaultLabels: %s", out)
	}
}

// TestLabelsFieldsAllCovered fails when a Labels field is added without a
// [labelCases] entry, so the "every built-in label is localizable" guarantee
// cannot silently rot.
func TestLabelsFieldsAllCovered(t *testing.T) {
	var covered []string
	for _, tt := range labelCases {
		covered = append(covered, tt.fields...)
	}
	typ := reflect.TypeOf(Labels{})
	for i := range typ.NumField() {
		name := typ.Field(i).Name
		if !slices.Contains(covered, name) {
			t.Errorf("Labels.%s is not exercised by any labelCases entry", name)
		}
		if reflect.ValueOf(DefaultLabels).Field(i).String() == "" {
			t.Errorf("DefaultLabels.%s is unset", name)
		}
		if reflect.ValueOf(germanLabels).Field(i).String() == "" {
			t.Errorf("germanLabels.%s is unset", name)
		}
	}
}

// TestLabelsFromContextFallback documents that a context without Labels
// falls back to DefaultLabels.
func TestLabelsFromContextFallback(t *testing.T) {
	if got := LabelsFromContext(context.Background()); got != DefaultLabels {
		t.Errorf("empty context should yield DefaultLabels, got %+v", got)
	}
}

// TestLabelsOrDefaultsCoversAllFields fails when a [Labels] field is added
// without a cmp.Or line in orDefaults: the new field would stay empty here
// while DefaultLabels has a value for it.
func TestLabelsOrDefaultsCoversAllFields(t *testing.T) {
	if got := (Labels{}).orDefaults(); got != DefaultLabels {
		t.Errorf("a zero Labels must fill every field from DefaultLabels, got %+v", got)
	}
}

// TestLabelsPartialFallsBackPerField is why the fallback exists: a translation
// that covers only part of the struct must render the English default for the
// rest rather than an empty label, which would drop an accessible name
// entirely.
func TestLabelsPartialFallsBackPerField(t *testing.T) {
	ctx := ContextWithLabels(context.Background(), Labels{DialogClose: "Schließen"})

	out := renderWith(t, ctx, DialogContent("dlg"))
	if !strings.Contains(out, `class="sr-only">Schließen<`) {
		t.Errorf("the translated field must win: %s", out)
	}

	out = renderWith(t, ctx, Spinner())
	if !strings.Contains(out, `aria-label="Loading"`) {
		t.Errorf("an untranslated field must fall back to DefaultLabels: %s", out)
	}
}
