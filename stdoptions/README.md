# stdoptions

Standard option lists for reflected form fields, registered with
`mx.RegisterNamedOptions`:

| Tag                         | List                                      | Value  | Count |
|-----------------------------|-------------------------------------------|--------|-------|
| `form:"options=iso4217"`    | ISO 4217 currencies in circulation        | `EUR`  | 153   |
| `form:"options=iso639-1"`   | ISO 639-1 languages                       | `de`   | 181   |
| `form:"options=iso3166-1"`  | ISO 3166-1 alpha-2 countries              | `AT`   | 249   |

```go
func main() {
	stdoptions.Register()
	// ...
}

type Invoice struct {
	Currency string `form:"options=iso4217"`
	Country  string `form:"options=iso3166-1"`
}

// per request, e.g. in a middleware:
ctx = stdoptions.ContextWithLanguage(ctx, "de-AT")
```

The option value is the code, the label the localized name without the code
("Euro", "Österreich"). A reflected form submits only a listed code, so a
withdrawn or unknown code is rejected as any other value outside the list.

## Registration

`Register()` is an explicit call, not an `init` side effect. The
`mx.RegisterNamedOptions` registry panics on a duplicate name, so an import
that registered on its own would crash any program that already uses one of
the names, and could not be turned off or renamed. An explicit call keeps the
dependency visible where the program starts. It is guarded by a `sync.Once`, so
calling it from several packages or tests is safe. To use other names or a
filtered list, register `Currencies`, `Languages` or `Countries` (each an
`mx.NamedOptionsFunc`) directly.

## Languages of the names

The names exist in English and German. The providers read the language from
the render context, set with `ContextWithLanguage` (a BCP 47 tag; only the
primary subtag counts), and fall back to English for any other language.
Each list is sorted by its names in the CLDR collation of that language, so
accented letters file under their base letter: "Ägypten" sorts between
"Afghanistan" and "Albanien", not after "Zypern". The sorting happens in the
generator, the package holds presorted data and has no dependency beyond `mx`.

## Data source

The lists are generated from [Unicode CLDR](https://cldr.unicode.org) as
published in [cldr-json](https://github.com/unicode-org/cldr-json), currently
release **49.0.0** (`stdoptions.CLDRRelease`). The release is pinned in the
`cldrRelease` constant of the generator, `tools/gen-stdoptions`, which lives in
the nested tools module because it collates with `golang.org/x/text`. To
update, bump that constant and run:

```bash
go generate ./stdoptions
```

This rewrites `cldrdata.go`. The generator fails if a code has no name in a
language or two codes share a name.

What is left out, and how CLDR marks it:

- **Currencies:** a currency is listed when an entry of the supplemental
  `currencyData` has no `_to` date and is not marked `_tender="false"`. CLDR
  sets `_to` when a currency is withdrawn (DEM, BGN), so a withdrawal announced
  with a future date drops the currency from the next generated list.
  `_tender="false"` marks funds codes and units of account (BOV, CHE, XDR,
  XAU, XXX).
- **Languages:** the codes are the two-letter replacements of CLDR's
  `overlong` language aliases (`deu` → `de`), CLDR's mapping of ISO 639-2/3
  codes to their ISO 639-1 form. Codes CLDR aliases away are left out: the
  withdrawn `in`, `iw`, `ji`, `jw` and `mo`, and `sh`, `tl` and `tw`, which
  CLDR maps to `sr-Latn`, `fil` and `ak`. Every listed value is a canonical
  language code.
- **Countries:** the two-letter regions of the supplemental `codeMappings` with
  an ISO 3166-1 numeric code below 900. That leaves out the exceptionally
  reserved codes, which have no numeric code (AC, CP, CQ, DG, EA, EZ, IC, TA,
  UN), the codes CLDR gives a numeric code in the user-assigned 900–999 range
  (AA, EU, QO, XA, XB, XK, ZZ), and the deprecated regions of the
  `territoryAlias` table (AN, CS, YU, …). Kosovo (XK) is therefore not listed;
  register your own list if you need it.
