package shadcn

import (
	"regexp"
	"strings"
	// "github.com/gotailwindcss/tailwind"
)

// TailwindClass represents a parsed Tailwind utility class
type TailwindClass struct {
	Prefix    string
	Important bool
	Value     string
}

// parseTailwindClass parses a single Tailwind class into its components
func parseTailwindClass(class string) TailwindClass {
	important := strings.HasPrefix(class, "!")
	if important {
		class = strings.TrimPrefix(class, "!")
	}

	re := regexp.MustCompile(`^([a-z-]+)(.+)$`)
	matches := re.FindStringSubmatch(class)

	if len(matches) < 3 {
		return TailwindClass{Value: class, Important: important}
	}

	return TailwindClass{
		Prefix:    matches[1],
		Value:     matches[2],
		Important: important,
	}
}

// shouldOverride determines if newClass should override oldClass
func shouldOverride(oldClass, newClass TailwindClass) bool {
	// If prefixes don't match, no override
	if oldClass.Prefix != newClass.Prefix {
		return false
	}

	// Important classes override non-important classes
	if !oldClass.Important && newClass.Important {
		return true
	}

	// For same importance level, last one wins
	return true
}

// Cn combines multiple class strings, resolving Tailwind conflicts
func Cn(classes ...interface{}) string {
	var validClasses []string

	// First, collect all valid classes
	for _, class := range classes {
		switch v := class.(type) {
		case string:
			if v != "" {
				validClasses = append(validClasses, strings.Fields(v)...)
			}
		case bool:
			// Skip bool values
			continue
		}
	}

	// Map to store final classes, using prefix as key to handle conflicts
	classMap := make(map[string]TailwindClass)

	// Process each class
	for _, class := range validClasses {
		parsed := parseTailwindClass(class)

		// If class has no prefix (like "hidden"), use full class as key
		key := parsed.Prefix
		if key == "" {
			key = parsed.Value
		}

		// Check for existing class with same prefix
		if existing, exists := classMap[key]; exists {
			if shouldOverride(existing, parsed) {
				classMap[key] = parsed
			}
		} else {
			classMap[key] = parsed
		}
	}

	// Build final class string
	var result []string
	for _, class := range classMap {
		if class.Important {
			result = append(result, "!"+class.Prefix+class.Value)
		} else {
			result = append(result, class.Prefix+class.Value)
		}
	}

	return strings.Join(result, " ")
}
