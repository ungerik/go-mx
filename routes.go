package mx

import (
	"iter"
	"net/http"
	"reflect"
	"strings"

	"github.com/domonda/go-errs"
)

type Route interface {
	http.Handler

	ParentPatterns() []string
	SetParentPatterns([]string)
	Pattern() string
	Methods() []string
	Path(values ...map[string]any) string
}

func StructRoutes(routesStruct any, parentPatterns ...string) iter.Seq2[string, http.Handler] {
	return func(yield func(string, http.Handler) bool) {
		for field, v := range ReflectStructFields(reflect.ValueOf(routesStruct)) {
			pattern, ok := field.Tag.Lookup("route")
			if !ok {
				pattern = NameToPath(field.Name, "/")
			} else if pattern == "" {
				panic(errs.Errorf("field %s of %T has empty route tag", field.Name, routesStruct))
			}
			if pattern == "-" {
				continue
			}
			var methods []string
			if m := field.Tag.Get("method"); m != "" {
				methods = append(methods, strings.Split(m, ",")...)
			}

			var handler http.Handler
			switch {
			case field.Type.AssignableTo(reflect.TypeFor[Route]()):
				if canBeNil(field.Type.Kind()) && v.IsNil() {
					panic(errs.Errorf("field %s of %T is nil", field.Name, routesStruct))
				}
				route := v.Interface().(Route)
				route.SetParentPatterns(parentPatterns)
				pattern = route.Pattern()
				methods = route.Methods()
				handler = route

			case field.Type.AssignableTo(reflect.TypeFor[http.HandlerFunc]()):
				if v.IsNil() {
					panic(errs.Errorf("field %s of %T is nil", field.Name, routesStruct))
				}
				handler = v.Interface().(http.HandlerFunc)

			case field.Type.AssignableTo(reflect.TypeFor[http.Handler]()):
				if canBeNil(field.Type.Kind()) && v.IsNil() {
					panic(errs.Errorf("field %s of %T is nil", field.Name, routesStruct))
				}
				handler = v.Interface().(http.Handler)

			case field.Type.Kind() == reflect.Struct || field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct && !v.IsNil():
				for pattern, handler := range StructRoutes(v.Interface(), append(parentPatterns, pattern)...) {
					if !yield(pattern, handler) {
						return
					}
				}
				continue

			default:
				continue
			}

			if strings.Contains(pattern, "..") {
				panic(errs.Errorf("field %s of %T pattern contains '..': %s", field.Name, routesStruct, pattern))
			}

			pattern = JoinPath(append(parentPatterns, pattern))
			if len(methods) == 0 {
				if !yield(pattern, handler) {
					return
				}
			} else {
				if patternMethod(pattern) != "" {
					panic(errs.Errorf("field %s of %T has methods from struct tag or mx.Route.Methods but also in pattern: %s", field.Name, routesStruct, pattern))
				}
				for _, method := range methods {
					if !yield(strings.ToUpper(method)+" "+pattern, handler) {
						return
					}
				}
			}
		}
	}
}

func NewRoute(pattern string, handler http.Handler, methods ...string) Route {
	if m := patternMethod(pattern); m != "" {
		if len(methods) > 0 {
			panic("NewRoute: can't have methods in pattern and as argument")
		}
		methods = strings.Split(m, ",")
	}
	if strings.Contains(pattern, "..") {
		panic("NewRoute: pattern contains '..'")
	}
	for i, m := range methods {
		methods[i] = strings.ToUpper(m)
	}
	return &routeImpl{
		Handler: handler,
		pattern: pattern,
		methods: methods,
	}
}

type routeImpl struct {
	http.Handler
	parentPatterns []string
	pattern        string
	methods        []string
}

func (r *routeImpl) ParentPatterns() []string {
	return r.parentPatterns
}

func (r *routeImpl) SetParentPatterns(parentPatterns []string) {
	r.parentPatterns = parentPatterns
}

func (r *routeImpl) Pattern() string {
	return r.pattern
}

func (r *routeImpl) Methods() []string {
	return r.methods
}

func (r *routeImpl) Path(values ...map[string]any) string {
	p := JoinAbsPath(append(r.parentPatterns, r.pattern))
	for _, values := range values {
		for name, value := range values {
			valueStr, err := FormatPathValue(name, value)
			if err != nil {
				panic(err)
			}
			p = strings.Replace(p, "{"+name+"}", valueStr, 1)
		}
	}
	return p
}

func patternMethod(pattern string) string {
	if i := strings.IndexAny(pattern, " \t"); i != -1 {
		return pattern[:i]
	}
	return ""
}

func PathValueNames(pattern string) map[string]struct{} {
	names := make(map[string]struct{})
	for _, part := range strings.Split(pattern, "/") {
		if len(part) > 0 && part[0] == '{' {
			if i := strings.IndexByte(part, '}'); i != -1 {
				names[part[1:i]] = struct{}{}
			}
		}
	}
	return names
}

// type ServeMuxHandler interface {
// 	Handle(pattern string, handler http.Handler)
// }

// func RegisterRoutesStruct(mux ServeMuxHandler, routesStruct any, parentPatterns ...string) {
// 	for pattern, handler := range StructRoutes(routesStruct, parentPatterns...) {
// 		mux.Handle(pattern, handler)
// 	}
// }

// func registerServeMux(mux ServeMuxHandler, pattern string, handler http.Handler) (err error) {
// 	defer errs.RecoverPanicAsError(&err)

// 	mux.Handle(pattern, handler)
// 	return nil
// }
