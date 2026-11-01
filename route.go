package mx

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/domonda/go-errs"
)

// Route is an http.Handler with a path pattern.
type Route interface {
	http.Handler

	// Pattern of the route compatible with http.ServeMux.
	Pattern() string

	// HTTP methods that the route handles.
	// If empty then the route handles all methods.
	Methods() []string

	// Path returns the URL path of the route: the path of its pattern with
	// every wildcard replaced by the value of the same name in values. It
	// returns an error rather than a path with a wildcard left in it.
	Path(values map[string]any) (string, error)

	Register(mux *http.ServeMux)
}

var _ http.Handler = new(ComponentFuncHandler[int])

// ComponentFuncHandler is an http.Handler that builds an argument of type T
// from the request, calls compFunc to obtain a [Component], and writes its
// rendered markup as the response. T is expected to be a struct: each of its
// fields is populated from the request's path value or form value of the same
// name (only string fields are currently supported; others are skipped). The
// optional response headers are added before the body is written.
type ComponentFuncHandler[T any] struct {
	compFunc      func(T) Component
	writerFactory WriterFactory
	headers       []http.Header
}

// NewComponentFuncHandler returns a [ComponentFuncHandler] for the component
// constructor compFunc, using writerFactory to create the render [Writer] (the
// package [DefaultWriterFactory] is used if it is nil) and adding the given
// response headers to each response.
func NewComponentFuncHandler[T any](compFunc func(T) Component, writerFactory WriterFactory, headers ...http.Header) *ComponentFuncHandler[T] {
	return &ComponentFuncHandler[T]{
		compFunc:      compFunc,
		writerFactory: writerFactory,
		headers:       headers,
	}
}

func (h *ComponentFuncHandler[T]) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// Recover panics from reflection: a non-struct T or an unsettable
	// field would otherwise crash the request without a response.
	defer func() {
		if p := recover(); p != nil {
			RespondNonContextError(response, request, errs.AsErrorWithDebugStack(p))
		}
	}()
	var funcArg T
	// TODO other kinds than struct
	for field, fieldVal := range ReflectStructFields(reflect.ValueOf(&funcArg)) {
		requestVal := request.PathValue(field.Name)
		if requestVal == "" {
			requestVal = request.FormValue(field.Name)
			if requestVal == "" {
				continue
			}
		}
		// TODO convert types other than string. Until then non-string
		// fields are skipped rather than panicking in reflect.SetString.
		if fieldVal.Kind() == reflect.String && fieldVal.CanSet() {
			fieldVal.SetString(requestVal)
		}
	}
	var body bytes.Buffer
	factory := h.writerFactory
	if factory == nil {
		factory = DefaultWriterFactory
	}
	writer := factory.NewWriter(&body)
	err := h.compFunc(funcArg).Render(request.Context(), writer)
	if err != nil {
		RespondNonContextError(response, request, err)
		return
	}
	for _, header := range h.headers {
		for key, values := range header {
			for _, value := range values {
				response.Header().Add(key, value)
			}
		}
	}
	response.Write(body.Bytes())
}

var _ Route = new(TypedRoute[struct{}])

// TypedRoute is a [Route] that pairs a [ComponentFuncHandler] of type T with an
// http.ServeMux path pattern, so the same handler can both serve requests and
// report its pattern and methods for registration.
type TypedRoute[T any] struct {
	ComponentFuncHandler[T]
	pattern string
}

// NewTypedRoute returns a [TypedRoute] for the given http.ServeMux pattern and
// component constructor compFunc. writerFactory and headers are passed through
// to the embedded [ComponentFuncHandler].
func NewTypedRoute[T any](pattern string, compFunc func(T) Component, writerFactory WriterFactory, headers ...http.Header) *TypedRoute[T] {
	return &TypedRoute[T]{
		ComponentFuncHandler: ComponentFuncHandler[T]{
			compFunc:      compFunc,
			writerFactory: writerFactory,
			headers:       headers,
		},
		pattern: pattern,
	}
}

// Pattern returns the http.ServeMux pattern of the route.
func (r *TypedRoute[T]) Pattern() string {
	return r.pattern
}

// Methods returns the HTTP method specified at the start of the pattern as a
// single-element slice, or nil if the pattern names no method (so the route
// handles all methods).
func (r *TypedRoute[T]) Methods() []string {
	method := patternMethod(r.pattern)
	if method == "" {
		return nil
	}
	return []string{method}
}

// Path returns the URL path of the route: the path of its pattern, without
// the method and host, with every wildcard replaced by the value of the same
// name in values, formatted with [FormatPathValue].
//
// The wildcards are those of http.ServeMux: a {name} segment takes one escaped
// value, a {name...} segment takes a value whose slashes separate segments
// (each escaped on its own, so a slash in the value stays a separator), and
// {$} only anchors the pattern and is dropped, keeping the trailing slash
// before it.
//
// A wildcard without a value and a value without a wildcard are both errors:
// the first would leave a "{name}" in the URL, and the second is almost always
// a misspelt name that would leave the wildcard it meant without its value. So
// is a value that formats as a "." or ".." segment, which would resolve to
// another path.
func (r *TypedRoute[T]) Path(values map[string]any) (string, error) {
	p := patternPath(r.pattern)
	// A pattern may name a host before its path, as in "example.com/blog/".
	if i := strings.IndexByte(p, '/'); i > 0 {
		p = p[i:]
	}
	segments := strings.Split(p, "/")
	used := make(map[string]struct{}, len(values))
	for i, segment := range segments {
		if len(segment) < 2 || segment[0] != '{' || segment[len(segment)-1] != '}' {
			continue
		}
		name := segment[1 : len(segment)-1]
		if name == "$" {
			segments[i] = ""
			continue
		}
		name, remainder := strings.CutSuffix(name, "...")
		value, ok := values[name]
		if !ok {
			return "", errs.Errorf("no value for wildcard %q of route pattern %q", name, r.pattern)
		}
		used[name] = struct{}{}
		if !remainder {
			formatted, err := formatPathSegment(r.pattern, name, value)
			if err != nil {
				return "", err
			}
			segments[i] = formatted
			continue
		}
		parts := strings.Split(fmt.Sprint(value), "/")
		for j, part := range parts {
			formatted, err := formatPathSegment(r.pattern, name, part)
			if err != nil {
				return "", err
			}
			parts[j] = formatted
		}
		segments[i] = strings.Join(parts, "/")
	}
	if len(used) < len(values) {
		for name := range values {
			if _, ok := used[name]; !ok {
				return "", errs.Errorf("route pattern %q has no wildcard %q", r.pattern, name)
			}
		}
	}
	return strings.Join(segments, "/"), nil
}

// formatPathSegment formats value as one segment of a path of pattern with
// [FormatPathValue]. Escaping leaves "." and ".." as they are, and a client
// or http.ServeMux resolves such a segment to another path, so either is an
// error.
func formatPathSegment(pattern, name string, value any) (string, error) {
	formatted, err := FormatPathValue(name, value)
	if err != nil {
		return "", err
	}
	if formatted == "." || formatted == ".." {
		return "", errs.Errorf("value %q for wildcard %q of route pattern %q is a dot segment", formatted, name, pattern)
	}
	return formatted, nil
}

// Register registers the route on mux under its pattern.
func (r *TypedRoute[T]) Register(mux *http.ServeMux) {
	mux.Handle(r.Pattern(), r)
}

func patternMethod(pattern string) string {
	if i := strings.IndexAny(pattern, " \t"); i != -1 {
		return pattern[:i]
	}
	return ""
}

func patternPath(pattern string) string {
	i := strings.LastIndexAny(pattern, " \t")
	return pattern[i+1:]
}

// PathValueNames returns the set of placeholder names in the path portion of an
// http.ServeMux pattern, that is the names inside "{...}" segments (any method
// prefix is ignored).
func PathValueNames(pattern string) map[string]struct{} {
	names := make(map[string]struct{})
	for part := range strings.SplitSeq(patternPath(pattern), "/") {
		if len(part) > 0 && part[0] == '{' {
			if i := strings.IndexByte(part, '}'); i != -1 {
				names[part[1:i]] = struct{}{}
			}
		}
	}
	return names
}
