package html_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// jsonLDBody returns the content between the start and end tag of a rendered
// JSON-LD script, failing the test if the markup has another shape.
func jsonLDBody(t *testing.T, rendered string) string {
	t.Helper()
	const start = `<script type="` + html.JSONLDType + `">`
	const end = `</script>`
	i := strings.Index(rendered, start)
	require.GreaterOrEqual(t, i, 0, "no JSON-LD start tag in %s", rendered)
	body, _, found := strings.Cut(rendered[i+len(start):], end)
	require.True(t, found, "no end tag in %s", rendered)
	return body
}

// renderJSONLD renders the JSON-LD script of v, failing the test on an error.
func renderJSONLD(t *testing.T, v any) string {
	t.Helper()
	var b strings.Builder
	err := html.JSONLD(v).Render(context.Background(), mx.NewCheckedWriter(&b))
	require.NoError(t, err)
	return b.String()
}

func TestJSONLD(t *testing.T) {
	type person struct {
		Context string   `json:"@context"`
		Type    string   `json:"@type"`
		Name    string   `json:"name"`
		Langs   []string `json:"knowsLanguage"`
	}
	value := person{
		Context: "https://schema.org",
		Type:    "Person",
		Name:    "Ada Lovelace",
		Langs:   []string{"en", "fr"},
	}

	element := html.JSONLD(value)
	require.NoError(t, element.Err)
	rendered := renderJSONLD(t, value)
	require.Equal(t,
		`<script type="application/ld+json">`+
			`{"@context":"https://schema.org","@type":"Person","name":"Ada Lovelace","knowsLanguage":["en","fr"]}`+
			`</script>`,
		rendered,
	)

	var roundTrip person
	err := json.Unmarshal([]byte(jsonLDBody(t, rendered)), &roundTrip)
	require.NoError(t, err)
	require.Equal(t, value, roundTrip)
}

func TestJSONLDHostileStrings(t *testing.T) {
	// A string value is the one place JSON can carry "<", and the JSON encoder
	// escapes it there, so neither an end tag nor a comment opener reaches the
	// markup however hostile the string is.
	hostile := map[string]string{
		"@type":       "Thing",
		"name":        `</script><script>alert(1)</script>`,
		"description": `<!-- <script> & </SCRIPT >`,
	}
	rendered := renderJSONLD(t, hostile)
	require.Equal(t, 1, strings.Count(rendered, "</script>"), rendered)
	require.True(t, strings.HasSuffix(rendered, "</script>"), rendered)
	require.NotContains(t, strings.ToLower(rendered), "</script><script>")
	require.NotContains(t, rendered, "<!--")
	require.NotContains(t, rendered, "alert(1)</")

	var roundTrip map[string]string
	err := json.Unmarshal([]byte(jsonLDBody(t, rendered)), &roundTrip)
	require.NoError(t, err)
	require.Equal(t, hostile, roundTrip, "escaping must not change the data")

	// The output of a json.Marshaler is escaped the same way: the encoder
	// re-compacts it with HTML escaping instead of copying it verbatim.
	raw := json.RawMessage(`{"name":"</script><script>alert(1)</script>"}`)
	rendered = renderJSONLD(t, raw)
	require.Equal(t, 1, strings.Count(rendered, "</script>"), rendered)
	require.Contains(t, rendered, `\u003c/script\u003e`)

	// U+2028 and U+2029 are valid in JSON strings but line terminators in
	// older JavaScript; they are escaped too.
	rendered = renderJSONLD(t, map[string]string{"text": "a\u2028b\u2029c"})
	require.Contains(t, rendered, `a\u2028b\u2029c`)
}

func TestJSONLDDeferredErrors(t *testing.T) {
	var nilPerson *struct{ Name string }
	tests := []struct {
		name    string
		value   any
		wantErr string
	}{
		{name: "nil", value: nil, wantErr: "nil value"},
		{name: "typed nil pointer", value: nilPerson, wantErr: "marshals to null"},
		{name: "nil map", value: map[string]any(nil), wantErr: "marshals to null"},
		{name: "unmarshalable", value: make(chan int), wantErr: "unsupported type"},
		{name: "unmarshalable field", value: map[string]any{"f": func() {}}, wantErr: "unsupported type"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The error is visible before rendering ...
			element := html.JSONLD(tt.value)
			require.ErrorContains(t, element.Err, tt.wantErr)

			// ... and fails the render without writing anything.
			var b strings.Builder
			err := element.Render(context.Background(), mx.NewCheckedWriter(&b))
			require.ErrorIs(t, err, element.Err)
			require.Empty(t, b.String())
		})
	}
}

func TestJSONLDInDocumentHead(t *testing.T) {
	doc := &html.Document{
		Title: "Structured",
		HeadCustom: html.JSONLD(map[string]any{
			"@context": "https://schema.org",
			"@type":    "WebPage",
			"name":     "Structured",
		}),
		Body: html.P("Hello"),
	}
	var b strings.Builder
	err := doc.Render(context.Background(), mx.NewCheckedWriter(&b).WithIndent("", "  "))
	require.NoError(t, err)
	rendered := b.String()

	head, _, found := strings.Cut(rendered, "</head>")
	require.True(t, found, rendered)
	require.Contains(t, head,
		`<script type="application/ld+json">{"@context":"https://schema.org","@type":"WebPage","name":"Structured"}</script>`,
	)

	var roundTrip map[string]any
	err = json.Unmarshal([]byte(jsonLDBody(t, rendered)), &roundTrip)
	require.NoError(t, err)
	require.Equal(t, "WebPage", roundTrip["@type"])

	// A deferred JSON-LD error fails the whole document.
	doc.HeadCustom = html.JSONLD(make(chan int))
	err = doc.Render(context.Background(), mx.NewCheckedWriter(&b))
	require.Error(t, err)
}
