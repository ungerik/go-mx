package mx

import (
	"bytes"
	"net/http"
	"reflect"
	"strings"
)

// Route is an http.Handler with a path pattern.
type Route interface {
	http.Handler

	// Pattern of the route compatible with http.ServeMux.
	Pattern() string

	// HTTP methods that the route handles.
	// If empty then the route handles all methods.
	Methods() []string

	// Path or the route with placeholders replaced by named values.
	Path(values ...map[string]any) string

	Register(mux *http.ServeMux)
}

var _ http.Handler = new(ComponentFuncHandler[int])

type ComponentFuncHandler[T any] struct {
	compFunc      func(T) Component
	writerFactory WriterFactory
	headers       []http.Header
}

func NewComponentFuncHandler[T any](compFunc func(T) Component, writerFactory WriterFactory, headers ...http.Header) *ComponentFuncHandler[T] {
	return &ComponentFuncHandler[T]{
		compFunc:      compFunc,
		writerFactory: writerFactory,
		headers:       headers,
	}
}

func (h *ComponentFuncHandler[T]) ServeHTTP(response http.ResponseWriter, request *http.Request) {
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
		// TODO convert other types than string
		fieldVal.SetString(requestVal)
	}
	var body bytes.Buffer
	factory := h.writerFactory
	if factory == nil {
		factory = DefaultWriterFactory
	}
	writer := factory.NewWriter(&body)
	err := h.compFunc(funcArg).Render(request.Context(), writer)
	if err != nil {
		RespondNonContextError(response, err)
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

type TypedRoute[T any] struct {
	ComponentFuncHandler[T]
	pattern string
}

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

func (r *TypedRoute[T]) Pattern() string {
	return r.pattern
}

func (r *TypedRoute[T]) Methods() []string {
	method := patternMethod(r.pattern)
	if method == "" {
		return nil
	}
	return []string{method}
}

// Path or the route with placeholders replaced by named values.
func (r *TypedRoute[T]) Path(values ...map[string]any) string {
	// TODO
	// p := JoinAbsPath(append(r.parentPatterns, r.pattern))
	// for _, values := range values {
	// 	for name, value := range values {
	// 		valueStr, err := FormatPathValue(name, value)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		p = strings.Replace(p, "{"+name+"}", valueStr, 1)
	// 	}
	// }
	// return p

	return r.pattern
}

func (r *TypedRoute[T]) Register(mux *http.ServeMux) {
	mux.Handle(r.Pattern(), r)
}

// type NestedRoute interface {
// 	Route

// 	ParentPatterns() []string
// 	SetParentPatterns([]string)
// }

// func NewRoute(pattern string, handler http.Handler, methods ...string) *nestedRoute {
// 	if m := patternMethod(pattern); m != "" {
// 		if len(methods) > 0 {
// 			panic("NewRoute: can't have methods in pattern and as argument")
// 		}
// 		methods = strings.Split(m, ",")
// 	}
// 	if strings.Contains(pattern, "..") {
// 		panic("NewRoute: pattern contains '..'")
// 	}
// 	for i, m := range methods {
// 		methods[i] = strings.ToUpper(m)
// 	}
// 	return &nestedRoute{
// 		Handler: handler,
// 		pattern: pattern,
// 		methods: methods,
// 	}
// }

// type nestedRoute struct {
// 	http.Handler
// 	parentPatterns []string
// 	pattern        string
// 	methods        []string
// }

// func (r *nestedRoute) ParentPatterns() []string {
// 	return r.parentPatterns
// }

// func (r *nestedRoute) SetParentPatterns(parentPatterns []string) {
// 	r.parentPatterns = parentPatterns
// }

// func (r *nestedRoute) Pattern() string {
// 	return r.pattern
// }

// func (r *nestedRoute) Methods() []string {
// 	return r.methods
// }

// func (r *nestedRoute) Path(values ...map[string]any) string {
// 	p := JoinAbsPath(append(r.parentPatterns, r.pattern))
// 	for _, values := range values {
// 		for name, value := range values {
// 			valueStr, err := FormatPathValue(name, value)
// 			if err != nil {
// 				panic(err)
// 			}
// 			p = strings.Replace(p, "{"+name+"}", valueStr, 1)
// 		}
// 	}
// 	return p
// }

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

func PathValueNames(pattern string) map[string]struct{} {
	names := make(map[string]struct{})
	for _, part := range strings.Split(patternPath(pattern), "/") {
		if len(part) > 0 && part[0] == '{' {
			if i := strings.IndexByte(part, '}'); i != -1 {
				names[part[1:i]] = struct{}{}
			}
		}
	}
	return names
}
