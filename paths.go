package mx

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"unicode"

	"github.com/domonda/go-errs"
)

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

func JoinAbsPath(segments []string) string {
	p := JoinPath(segments)
	if !strings.HasPrefix(p, "/") {
		return "/" + p
	}
	return p
}

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
