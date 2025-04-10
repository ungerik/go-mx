package mx

/*
import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestReflectRoutes(t *testing.T) {
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
				for pattern, handler := range ReflectRoutes(tt.routesStruct) {
					gotPatterns = append(gotPatterns, pattern)
					require.NotNilf(t, handler, "nil handler for pattern %s from ReflectRoutes", pattern)
					if route, ok := handler.(Route); ok {
						gotPaths = append(gotPaths, route.Path(nil))
					}
				}
			}()
			require.Equal(t, tt.wantPanic, gotPanic, "ReflectRoutes panics")
			require.Equal(t, tt.wantPatterns, gotPatterns, "ReflectRoutes patterns")
			require.Equal(t, tt.wantPaths, gotPaths, "ReflectRoutes -> Route.Path(nil)")
		})
	}
}
*/
