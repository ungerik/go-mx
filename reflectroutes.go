package mx

// import (
// 	"iter"
// 	"net/http"
// 	"reflect"
// 	"strings"
//

// 	"github.com/domonda/go-errs"
// )

// func ReflectRoutes(routesStruct any, parentPatterns ...string) iter.Seq2[string, http.Handler] {
// 	return func(yield func(string, http.Handler) bool) {
// 		for field, v := range ReflectStructFields(reflect.ValueOf(routesStruct)) {
// 			pattern, ok := field.Tag.Lookup("route")
// 			if !ok {
// 				pattern = NameToPath(field.Name, "/")
// 			} else if pattern == "" {
// 				panic(errs.Errorf("field %s of %T has empty route tag", field.Name, routesStruct))
// 			}
// 			if pattern == "-" {
// 				continue
// 			}
// 			var methods []string
// 			if m := field.Tag.Get("method"); m != "" {
// 				methods = append(methods, strings.Split(m, ",")...)
// 			}

// 			var handler http.Handler
// 			switch {
// 			case field.Type.AssignableTo(reflect.TypeFor[NestedRoute]()):
// 				if canBeNil(field.Type.Kind()) && v.IsNil() {
// 					panic(errs.Errorf("field %s of %T is nil", field.Name, routesStruct))
// 				}
// 				route := v.Interface().(NestedRoute)
// 				route.SetParentPatterns(parentPatterns)
// 				pattern = route.Pattern()
// 				methods = route.Methods()
// 				handler = route

// 			case field.Type.AssignableTo(reflect.TypeFor[Route]()):
// 				if canBeNil(field.Type.Kind()) && v.IsNil() {
// 					panic(errs.Errorf("field %s of %T is nil", field.Name, routesStruct))
// 				}
// 				route := v.Interface().(Route)
// 				pattern = route.Pattern()
// 				methods = route.Methods()
// 				handler = route

// 			case field.Type.AssignableTo(reflect.TypeFor[http.HandlerFunc]()):
// 				if v.IsNil() {
// 					panic(errs.Errorf("field %s of %T is nil", field.Name, routesStruct))
// 				}
// 				handler = v.Interface().(http.HandlerFunc)

// 			case field.Type.AssignableTo(reflect.TypeFor[http.Handler]()):
// 				if canBeNil(field.Type.Kind()) && v.IsNil() {
// 					panic(errs.Errorf("field %s of %T is nil", field.Name, routesStruct))
// 				}
// 				handler = v.Interface().(http.Handler)

// 			case field.Type.Kind() == reflect.Struct || field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Struct && !v.IsNil():
// 				for pattern, handler := range ReflectRoutes(v.Interface(), append(parentPatterns, pattern)...) {
// 					if !yield(pattern, handler) {
// 						return
// 					}
// 				}
// 				continue

// 			default:
// 				continue
// 			}

// 			if strings.Contains(pattern, "..") {
// 				panic(errs.Errorf("field %s of %T pattern contains '..': %s", field.Name, routesStruct, pattern))
// 			}

// 			pattern = JoinPath(append(parentPatterns, pattern))
// 			if len(methods) == 0 {
// 				if !yield(pattern, handler) {
// 					return
// 				}
// 			} else {
// 				if patternMethod(pattern) != "" {
// 					panic(errs.Errorf("field %s of %T has methods from struct tag or mx.Route.Methods but also in pattern: %s", field.Name, routesStruct, pattern))
// 				}
// 				for _, method := range methods {
// 					if !yield(strings.ToUpper(method)+" "+pattern, handler) {
// 						return
// 					}
// 				}
// 			}
// 		}
// 	}
// }

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
