//go:generate go -C ../tools run ./gen-stdoptions -out ../stdoptions/cldrdata.go

// Package stdoptions provides standard option lists for reflected form
// fields — ISO 4217 currencies, ISO 639-1 languages and ISO 3166-1 alpha-2
// countries — as named options providers of [mx.RegisterNamedOptions]:
//
//	stdoptions.Register() // once, at program start
//
//	type Invoice struct {
//		Currency string `form:"options=iso4217"`
//		Language string `form:"options=iso639-1"`
//		Country  string `form:"options=iso3166-1"`
//	}
//
// An option's Value is the code ("EUR", "de", "AT") and its Name the
// localized name ("Euro", "German", "Austria"), sorted by that name in the
// collation of its language. Names exist in English and German; the
// providers take the language from the render context, installed with
// [ContextWithLanguage], and fall back to English.
//
// The lists are generated from Unicode CLDR (release [CLDRRelease]) by
// tools/gen-stdoptions, see the README. The package is separate from the
// root mx package so that only binaries importing it carry the data.
package stdoptions

import (
	"context"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/ungerik/go-mx"
)

// The names [Register] registers the option lists under, for use in
// form:"options=…" tags.
const (
	CurrenciesOptionsName = "iso4217"
	LanguagesOptionsName  = "iso639-1"
	CountriesOptionsName  = "iso3166-1"
)

var registerOnce sync.Once

// Register registers [Currencies], [Languages] and [Countries] with
// [mx.RegisterNamedOptions] under [CurrenciesOptionsName],
// [LanguagesOptionsName] and [CountriesOptionsName]. Call it once at
// program start, before the first form renders; further calls do nothing.
// It panics if the application registered one of the names itself.
//
// To register a list under another name, or a filtered list, pass the
// provider functions to [mx.RegisterNamedOptions] directly instead.
func Register() {
	registerOnce.Do(func() {
		mx.RegisterNamedOptions(CurrenciesOptionsName, Currencies)
		mx.RegisterNamedOptions(LanguagesOptionsName, Languages)
		mx.RegisterNamedOptions(CountriesOptionsName, Countries)
	})
}

// Currencies returns the ISO 4217 currencies that are legal tender in at
// least one country, named in the language of ctx. Withdrawn currencies,
// funds codes and units of account (BOV, XDR, XAU, …) are not included.
// It is a [mx.NamedOptionsFunc] and never fails.
func Currencies(ctx context.Context) ([]mx.NamedOption, error) {
	return localized(ctx, currencies), nil
}

// Languages returns the ISO 639-1 languages, named in the language of ctx.
// Only codes CLDR treats as canonical are included, so withdrawn codes
// (iw, in, …) and codes CLDR maps to another language (sh, tl, tw) are
// not. It is a [mx.NamedOptionsFunc] and never fails.
func Languages(ctx context.Context) ([]mx.NamedOption, error) {
	return localized(ctx, languages), nil
}

// Countries returns the officially assigned ISO 3166-1 alpha-2 country
// codes, named in the language of ctx. Exceptionally reserved codes (EU,
// UN, AC, …), user-assigned codes (XK, ZZ, …) and deprecated codes are not
// included. It is a [mx.NamedOptionsFunc] and never fails.
func Countries(ctx context.Context) ([]mx.NamedOption, error) {
	return localized(ctx, countries), nil
}

// localized returns a copy of the list in the language of ctx, so a caller
// that reorders or edits its options cannot change the shared data.
func localized(ctx context.Context, lists map[string][]mx.NamedOption) []mx.NamedOption {
	return slices.Clone(lists[LanguageFromContext(ctx)])
}

// ctxKeyLanguage is the context key under which [ContextWithLanguage]
// stores the resolved language.
type ctxKeyLanguage struct{}

// ContextWithLanguage returns a derived context whose option lists are
// named in lang, a BCP 47 language tag such as "de", "de-AT" or "en-US".
// Only the primary language subtag counts; a language without names falls
// back to English. Wire it once per request:
//
//	next.ServeHTTP(w, r.WithContext(stdoptions.ContextWithLanguage(r.Context(), lang)))
func ContextWithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, ctxKeyLanguage{}, resolveLanguage(lang))
}

// LanguageFromContext returns the language the option lists of ctx are
// named in: the one installed with [ContextWithLanguage] if names exist
// for it, otherwise "en".
func LanguageFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(ctxKeyLanguage{}).(string); ok {
		return lang
	}
	return fallbackLanguage
}

// SupportedLanguages returns the languages the option lists have names in.
func SupportedLanguages() []string {
	return slices.Sorted(maps.Keys(currencies))
}

// resolveLanguage maps a BCP 47 tag to the language of its primary subtag
// if the lists have names in it, and to [fallbackLanguage] otherwise.
func resolveLanguage(tag string) string {
	primary, _, _ := strings.Cut(tag, "-")
	primary, _, _ = strings.Cut(primary, "_")
	primary = strings.ToLower(strings.TrimSpace(primary))
	if _, ok := currencies[primary]; ok {
		return primary
	}
	return fallbackLanguage
}
