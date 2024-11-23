package mx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNameToPath(t *testing.T) {
	testCases := map[string]string{
		// Non Go names
		"":    "",
		"_":   "_",
		"/":   "/",
		"/./": "/./",
		" ":   " ",
		"  ":  "  ",

		// Normal cases
		"already/a/path":   "already/a/path",
		"/already/a/path/": "/already/a/path/",
		"HelloWorld":       "hello/world",
		"Hello-World":      "hello-world",
		"Hello_World":      "hello_world",
		"Hello.World":      "hello.world",
		"Hello/World":      "hello/world",
		"DocumentID":       "document/id",
		"HTMLHandler":      "htmlhandler",
		"Straßenadresse/":  "straßenadresse/",
		"もしもしWorld":        "もしもし/world",
	}
	for name, expected := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := NameToPath(name, "/")
			require.Equal(t, expected, actual)
		})
	}
}
