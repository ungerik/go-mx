package highlight

// The Go spec's "Predeclared identifiers" section. They are highlighted
// distinctly from user identifiers, the way editors do, even though they are
// ordinary identifiers (not keywords) to the scanner.

// predeclaredTypes are the predeclared type names, including the alias any and
// the constraint comparable.
var predeclaredTypes = map[string]bool{
	"any":        true,
	"bool":       true,
	"byte":       true,
	"comparable": true,
	"complex64":  true,
	"complex128": true,
	"error":      true,
	"float32":    true,
	"float64":    true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"rune":       true,
	"string":     true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uintptr":    true,
}

// predeclaredConsts are the predeclared constants and the predeclared zero
// value nil.
var predeclaredConsts = map[string]bool{
	"true":  true,
	"false": true,
	"iota":  true,
	"nil":   true,
}

// predeclaredFuncs are the predeclared (builtin) functions.
var predeclaredFuncs = map[string]bool{
	"append":  true,
	"cap":     true,
	"clear":   true,
	"close":   true,
	"complex": true,
	"copy":    true,
	"delete":  true,
	"imag":    true,
	"len":     true,
	"make":    true,
	"max":     true,
	"min":     true,
	"new":     true,
	"panic":   true,
	"print":   true,
	"println": true,
	"real":    true,
	"recover": true,
}
