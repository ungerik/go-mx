package mx

import (
	"net/http"
	"strings"
)

type Route interface {
	http.Handler

	Pattern() string
	Methods() []string
	Path(values ...map[string]any) string
}

type NestedRoute interface {
	Route

	ParentPatterns() []string
	SetParentPatterns([]string)
}

func NewRoute(pattern string, handler http.Handler, methods ...string) *nestedRoute {
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
	return &nestedRoute{
		Handler: handler,
		pattern: pattern,
		methods: methods,
	}
}

type nestedRoute struct {
	http.Handler
	parentPatterns []string
	pattern        string
	methods        []string
}

func (r *nestedRoute) ParentPatterns() []string {
	return r.parentPatterns
}

func (r *nestedRoute) SetParentPatterns(parentPatterns []string) {
	r.parentPatterns = parentPatterns
}

func (r *nestedRoute) Pattern() string {
	return r.pattern
}

func (r *nestedRoute) Methods() []string {
	return r.methods
}

func (r *nestedRoute) Path(values ...map[string]any) string {
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
