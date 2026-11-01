package stdoptions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

type invoice struct {
	Currency string `form:"options=iso4217"`
	Language string `form:"options=iso639-1"`
	Country  string `form:"options=iso3166-1"`
}

func TestRegister(t *testing.T) {
	// A second call must not panic on the duplicate registration that
	// mx.RegisterNamedOptions would otherwise reject.
	Register()
	Register()

	for name, want := range map[string]string{
		CurrenciesOptionsName: "EUR",
		LanguagesOptionsName:  "de",
		CountriesOptionsName:  "AT",
	} {
		opts, err := mx.CollectOptions(t.Context(), reflect.Value{}, mx.FormTag{Options: name}, reflect.TypeFor[string]())
		require.NoError(t, err, name)
		assert.True(t, hasValue(opts, want), "%s has no %s", name, want)
	}
}

func TestRegister_FormTag(t *testing.T) {
	Register()
	ctx := ContextWithLanguage(t.Context(), "de")

	want := map[string]mx.NamedOption{
		"Currency": {Name: "Euro", Value: "EUR"},
		"Language": {Name: "Deutsch", Value: "de"},
		"Country":  {Name: "Österreich", Value: "AT"},
	}
	typ := reflect.TypeFor[invoice]()
	for i := range typ.NumField() {
		field := typ.Field(i)
		opts, err := mx.CollectOptions(ctx, reflect.Value{}, mx.ParseFormTag(field), field.Type)
		require.NoError(t, err, field.Name)
		assert.Contains(t, opts, want[field.Name], field.Name)
	}
}

// The reflected form handler renders a tagged field as a select of the
// localized options and accepts only a listed code on submit.
func TestRegister_ReflectFormHandler(t *testing.T) {
	Register()
	var saved invoice
	handler := mx.ReflectFormHandler(
		func(context.Context) (*invoice, error) {
			return &invoice{Currency: "EUR", Language: "de", Country: "AT"}, nil
		},
		func(_ context.Context, v *invoice) error { saved = *v; return nil },
		html.FieldDecider,
	)
	withLanguage := func(r *http.Request) *http.Request {
		return r.WithContext(ContextWithLanguage(r.Context(), "de-AT"))
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, withLanguage(httptest.NewRequest(http.MethodGet, "/invoice", nil)))
	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `value="AT"`)
	assert.Contains(t, body, "Österreich")
	assert.Contains(t, body, "US-Dollar")
	assert.NotContains(t, body, "Austria")

	post := func(country string) *httptest.ResponseRecorder {
		form := url.Values{
			"Currency": {"USD"}, "Language": {"en"}, "Country": {country},
			// The sentinels mark the fields as rendered, so they are bound.
			"__present__Currency": {"1"}, "__present__Language": {"1"}, "__present__Country": {"1"},
		}
		req := httptest.NewRequest(http.MethodPost, "/invoice", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, withLanguage(req))
		return rec
	}
	rec = post("CH")
	require.Equal(t, http.StatusSeeOther, rec.Code, rec.Body.String())
	assert.Equal(t, invoice{Currency: "USD", Language: "en", Country: "CH"}, saved)

	saved = invoice{}
	rec = post("XK") // user-assigned, not in the list
	assert.Equal(t, http.StatusOK, rec.Code, "re-rendered form, not a redirect")
	assert.Contains(t, rec.Body.String(), `id="Country" aria-invalid="true"`)
	assert.Equal(t, invoice{}, saved)
}

func TestLocalizedNames(t *testing.T) {
	en := t.Context()
	de := ContextWithLanguage(t.Context(), "de")

	for _, tc := range []struct {
		list   func(context.Context) ([]mx.NamedOption, error)
		code   string
		en, de string
	}{
		{Currencies, "EUR", "Euro", "Euro"},
		{Currencies, "USD", "US Dollar", "US-Dollar"},
		{Currencies, "CHF", "Swiss Franc", "Schweizer Franken"},
		{Languages, "de", "German", "Deutsch"},
		{Languages, "en", "English", "Englisch"},
		{Countries, "AT", "Austria", "Österreich"},
		{Countries, "DE", "Germany", "Deutschland"},
		{Countries, "GB", "United Kingdom", "Vereinigtes Königreich"},
	} {
		assert.Equal(t, tc.en, nameOf(t, en, tc.list, tc.code), tc.code)
		assert.Equal(t, tc.de, nameOf(t, de, tc.list, tc.code), tc.code)
	}
}

func TestLists_SameCodesInEveryLanguage(t *testing.T) {
	for _, lists := range []map[string][]mx.NamedOption{currencies, languages, countries} {
		want := sortedValues(lists[fallbackLanguage])
		require.NotEmpty(t, want)
		for lang, opts := range lists {
			assert.Equal(t, want, sortedValues(opts), lang)
		}
	}
}

// The lists are sorted by the localized name with accented letters filed
// under their base letter, not by code and not in byte order, which would
// put every name starting with Ä, Å or Ö after Z.
func TestLists_SortedByLocalizedName(t *testing.T) {
	en := t.Context()
	de := ContextWithLanguage(t.Context(), "de")

	for _, tc := range []struct {
		ctx   context.Context
		list  func(context.Context) ([]mx.NamedOption, error)
		names []string
	}{
		{en, Countries, []string{"Afghanistan", "Åland Islands", "Albania"}},
		{en, Countries, []string{"Côte d’Ivoire", "Croatia", "Cuba", "Curaçao"}},
		{en, Countries, []string{"Türkiye", "Turkmenistan"}},
		{de, Countries, []string{"Afghanistan", "Ägypten", "Ålandinseln", "Albanien"}},
		{de, Countries, []string{"Oman", "Österreich", "Pakistan"}},
		{de, Countries, []string{"Turkmenistan", "Tuvalu"}},
		{en, Currencies, []string{"Afghan Afghani", "Albanian Lek", "Algerian Dinar"}},
		{de, Languages, []string{"Dänisch", "Deutsch"}},
	} {
		opts, err := tc.list(tc.ctx)
		require.NoError(t, err)
		last := -1
		for _, name := range tc.names {
			i := slices.IndexFunc(opts, func(o mx.NamedOption) bool { return o.Name == name })
			require.GreaterOrEqual(t, i, 0, "no option named %q", name)
			assert.Greater(t, i, last, "%q is out of order in %v", name, tc.names)
			last = i
		}
	}
}

func TestLists_Exclusions(t *testing.T) {
	for _, tc := range []struct {
		list  func(context.Context) ([]mx.NamedOption, error)
		codes []string
	}{
		// Withdrawn (DEM, BGN, ANG), funds codes and units of account.
		{Currencies, []string{"DEM", "BGN", "ANG", "BOV", "CHE", "XDR", "XAU", "XXX"}},
		// Withdrawn (iw, in), and aliased to another code by CLDR (sh, tl, tw).
		{Languages, []string{"iw", "in", "sh", "tl", "tw"}},
		// Exceptionally reserved, user-assigned and deprecated.
		{Countries, []string{"EU", "UN", "AC", "XK", "ZZ", "QO", "AN", "YU"}},
	} {
		opts, err := tc.list(t.Context())
		require.NoError(t, err)
		for _, code := range tc.codes {
			assert.False(t, hasValue(opts, code), "%s must not be listed", code)
		}
	}
}

func TestLists_ReturnCopies(t *testing.T) {
	opts, err := Countries(t.Context())
	require.NoError(t, err)
	first := opts[0]
	opts[0] = mx.NamedOption{Name: "changed", Value: "changed"}

	again, err := Countries(t.Context())
	require.NoError(t, err)
	assert.Equal(t, first, again[0])
}

func TestContextWithLanguage(t *testing.T) {
	assert.Equal(t, "en", LanguageFromContext(t.Context()))
	for tag, want := range map[string]string{
		"de":    "de",
		"de-AT": "de",
		"DE_ch": "de",
		"en-US": "en",
		"fr":    "en",
		"":      "en",
	} {
		assert.Equal(t, want, LanguageFromContext(ContextWithLanguage(t.Context(), tag)), tag)
	}
	assert.Equal(t, []string{"de", "en"}, SupportedLanguages())
}

func hasValue(opts []mx.NamedOption, value string) bool {
	return slices.ContainsFunc(opts, func(o mx.NamedOption) bool { return o.Value == value })
}

func nameOf(t *testing.T, ctx context.Context, list func(context.Context) ([]mx.NamedOption, error), code string) string {
	t.Helper()
	opts, err := list(ctx)
	require.NoError(t, err)
	i := slices.IndexFunc(opts, func(o mx.NamedOption) bool { return o.Value == code })
	require.GreaterOrEqual(t, i, 0, "no option %s", code)
	return opts[i].Name
}

func sortedValues(opts []mx.NamedOption) []string {
	values := make([]string, len(opts))
	for i, o := range opts {
		values[i] = o.Value
	}
	slices.Sort(values)
	return values
}
