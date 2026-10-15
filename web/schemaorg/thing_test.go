package schemaorg_test

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx/web/schemaorg"
)

var update = flag.Bool("update", false, "rewrite the golden JSON files in testdata/")

// cest is the zone of the fixture times, a non-UTC offset so the tests see
// that dates keep the offset they were given.
var cest = time.FixedZone("CEST", 2*60*60)

// requireGolden marshals v indented and compares it byte for byte with
// testdata/<name>.json. With -update it rewrites the file instead, which is
// how a deliberate change of the output is accepted.
func requireGolden(t *testing.T, name string, v any) {
	t.Helper()
	got, err := json.MarshalIndent(v, "", "  ")
	require.NoError(t, err)
	got = append(got, '\n')
	path := filepath.Join("testdata", name+".json")
	if *update {
		err = os.WriteFile(path, got, 0o644)
		require.NoError(t, err)
		return
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err, "run go test -update to create it")
	require.Equal(t, string(want), string(got))
}

// requireJSON marshals v compactly and compares it with want byte for byte, so
// the property order is checked along with the content.
func requireJSON(t *testing.T, want string, v any) {
	t.Helper()
	got, err := json.Marshal(v)
	require.NoError(t, err)
	require.Equal(t, want, string(got))
}

func TestThing(t *testing.T) {
	// A constructed value is a valid root: @context and @type, plus what the
	// constructor was given, and nothing else.
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"Country","name":"Austria"}`,
		schemaorg.NewThing("Country", "Austria"),
	)

	// A reference to a node described elsewhere: no context, just the type
	// and the @id the node was published under.
	requireJSON(t,
		`{"@type":"Organization","@id":"https://example.com/#organization"}`,
		schemaorg.Thing{Type: "Organization", ID: "https://example.com/#organization"},
	)

	// @type is never omitted, not even when empty.
	requireJSON(t, `{"@type":""}`, schemaorg.Thing{})

	thing := schemaorg.NewThing("Brand", "Example")
	thing.ID = "https://example.com/#brand"
	thing.Description = "The brand of Example Inc."
	thing.URL = "https://example.com/"
	thing.SameAs = []string{"https://en.wikipedia.org/wiki/Example", "https://github.com/example"}
	thing.Image = "https://example.com/brand.png"
	requireGolden(t, "thing", thing)
}
