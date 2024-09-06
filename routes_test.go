package mx

import (
	"net/http"
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

type testServeMuxHandler struct {
	patterns []string
}

func (m *testServeMuxHandler) Handle(pattern string, handler http.Handler) {
	if handler == nil {
		panic("handler is nil")
	}
	m.patterns = append(m.patterns, pattern)
}

// func TestRegisterRoutesStruct(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		mux          testServeMuxHandler
// 		routesStruct any
// 		wantPatterns []string
// 		wantErr      bool
// 	}{
// 		{
// 			name:         "empty",
// 			routesStruct: struct{}{},
// 			wantPatterns: nil,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := RegisterRoutesStruct(&tt.mux, tt.routesStruct)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 			} else {
// 				require.NoError(t, err)
// 			}
// 			require.Equal(t, tt.wantPatterns, tt.mux.patterns)
// 		})
// 	}
// }

func TestStructRoutes(t *testing.T) {
	dummyHandlerFunc := func(w http.ResponseWriter, r *http.Request) {}
	dummyHandler := http.HandlerFunc(dummyHandlerFunc)

	type args struct {
		parentPatterns []string
	}
	tests := []struct {
		name         string
		routesStruct any
		wantPatterns []string
		wantPaths    []string
		wantPanic    bool
	}{
		{
			name:         "empty struct",
			routesStruct: struct{}{},
			wantPatterns: nil,
		},
		{
			name:         "empty struct pointer",
			routesStruct: new(struct{}),
			wantPatterns: nil,
		},
		{
			name: "simple struct",
			routesStruct: struct {
				HelloWorld http.Handler `route:"hello/world"`
			}{
				HelloWorld: dummyHandler,
			},
			wantPatterns: []string{"hello/world"},
		},
		{
			name: "Route",
			routesStruct: struct {
				TheRoute      Route
				PostPathValue Route
			}{
				TheRoute:      NewRoute("the/route", dummyHandler),
				PostPathValue: NewRoute("/value/{x}/", dummyHandler, "POST"),
			},
			wantPatterns: []string{"the/route", "POST /value/{x}/"},
			wantPaths:    []string{"/the/route", "/value/{x}/"},
		},

		// Error cases
		{
			name:         "non struct type",
			routesStruct: 666,
			wantPanic:    true,
		},
		{
			name:         "nil struct pointer",
			routesStruct: (*struct{})(nil),
			wantPanic:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotPatterns []string
				gotPaths    []string
				gotPanic    bool
			)
			func() {
				defer func() {
					if r := recover(); r != nil {
						gotPanic = true
					}
				}()
				for pattern, handler := range StructRoutes(tt.routesStruct) {
					gotPatterns = append(gotPatterns, pattern)
					require.NotNilf(t, handler, "nil handler for pattern %s from StructRoutes", pattern)
					if route, ok := handler.(Route); ok {
						gotPaths = append(gotPaths, route.Path(nil))
					}
				}
			}()
			require.Equal(t, tt.wantPanic, gotPanic, "StructRoutes panics")
			require.Equal(t, tt.wantPatterns, gotPatterns, "StructRoutes patterns")
			require.Equal(t, tt.wantPaths, gotPaths, "StructRoutes -> Route.Path(nil)")
		})
	}
}
