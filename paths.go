package mx

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"unicode"

	"github.com/domonda/go-errs"
)

// JoinPath joins URL path segments with path.Join, but preserves a trailing
// slash if the last segment has one (which path.Join would otherwise strip).
// It returns an empty string for no segments.
func JoinPath(segments []string) string {
	if len(segments) == 0 {
		return ""
	}
	p := path.Join(segments...)
	// path.Join removes trailing slashes
	if strings.HasSuffix(segments[len(segments)-1], "/") {
		p += "/"
	}
	return p
}

// JoinAbsPath is like [JoinPath] but ensures the result begins with a leading
// slash, making it an absolute path.
func JoinAbsPath(segments []string) string {
	p := JoinPath(segments)
	if !strings.HasPrefix(p, "/") {
		return "/" + p
	}
	return p
}

// FormatPathValue formats value (stringified with fmt.Sprint and then
// URL-path-escaped) for substitution into the path placeholder named name. It
// returns an error if name is empty or contains the characters '/', '{' or '}',
// which are invalid in a placeholder name.
func FormatPathValue(name string, value any) (valStr string, err error) {
	if name == "" {
		return "", errs.New("path value name is empty")
	}
	if strings.ContainsAny(name, "/{}") {
		return "", errs.Errorf("path value name %q contains invalid characters", name)
	}
	valStr = url.PathEscape(fmt.Sprint(value))
	// if strings.ContainsAny(valStr, ".:,;/?@&=+$") {
	// 	err = errors.Join(err, errs.Errorf("path value %q contains invalid characters: %q", name, valStr))
	// }
	return valStr, err
}

// NameToPath converts a name to a URL path
// by lower casing everything and inserting sep
// before every new upper case character in the name
// except for the first character and any punctuation or dashes.
func NameToPath(name, sep string) string {
	b := strings.Builder{}
	b.Grow(len(name))
	lastWasUpper := true
	for _, r := range name {
		lr := unicode.ToLower(r)
		isUpper := lr != r
		if isUpper && !lastWasUpper {
			b.WriteString(sep)
		}
		b.WriteRune(lr)
		lastWasUpper = isUpper || unicode.IsPunct(r) || unicode.IsSymbol(r)
	}
	return b.String()
}
