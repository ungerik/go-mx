package shadcn

import (
	"slices"
	"strings"
)

// Cn combines Tailwind CSS class strings into a single space-separated
// string, dropping earlier classes that a later class supersedes.
//
// Conflict resolution is a focused subset of the npm tailwind-merge
// algorithm, not a full reimplementation:
//
//   - The directional spacing families padding (p, px, py, pt, pr, pb,
//     pl, ps, pe) and margin (m, mx, ...) are resolved per physical
//     side. A later class only drops an earlier one when it covers a
//     superset of the same sides, so "p-4" after "px-2" wins, while
//     "px-2" after "p-4" is kept because it still overrides the x-axis.
//   - Every other utility is grouped by its first dash-separated
//     segment ("bg", "text", "flex", ...) and the last one wins. This
//     is correct for colors and most utilities but does not distinguish
//     sub-groups that share a segment: "text-lg" and "text-red-500"
//     would wrongly be treated as conflicting.
//
// Input order is preserved in the output. Empty strings and non-string
// arguments (such as a bool from a conditional expression) are ignored.
func Cn(classes ...any) string {
	var kept []string                 // surviving classes, in input order
	keys := make(map[string][]string) // class -> conflict keys it occupies

	for _, class := range classes {
		s, ok := class.(string)
		if !ok {
			continue // ignore non-strings, e.g. bool from a conditional
		}
		for field := range strings.FieldsSeq(s) {
			newKeys := conflictKeys(field)
			// Drop earlier classes fully superseded by this one.
			kept = slices.DeleteFunc(kept, func(old string) bool {
				if subsetOf(keys[old], newKeys) {
					delete(keys, old)
					return true
				}
				return false
			})
			keys[field] = newKeys
			kept = append(kept, field)
		}
	}
	return strings.Join(kept, " ")
}

// spacingSides maps a Tailwind directional side letter to the physical
// sides it covers. The unsuffixed form (e.g. "p-4") covers allSides.
var spacingSides = map[byte][]string{
	'x': {"l", "r"},
	'y': {"t", "b"},
	't': {"t"},
	'r': {"r"},
	'b': {"b"},
	'l': {"l"},
	's': {"l"}, // logical inline-start, approximated as left
	'e': {"r"}, // logical inline-end, approximated as right
}

var allSides = []string{"t", "r", "b", "l"}

// conflictKeys returns the set of conflict keys a class occupies. Two
// classes conflict when one key set is a subset of the other.
func conflictKeys(class string) []string {
	// A leading "!" (important) or "-" (negative value) does not change
	// which utility group the class belongs to.
	core := strings.TrimPrefix(class, "!")
	core = strings.TrimPrefix(core, "-")

	if len(core) > 0 {
		switch core[0] {
		case 'p':
			if keys := directionalKeys("pad", core[1:]); keys != nil {
				return keys
			}
		case 'm':
			if keys := directionalKeys("mar", core[1:]); keys != nil {
				return keys
			}
		}
	}

	// Generic fallback: group by the first dash-separated segment.
	seg, _, _ := strings.Cut(core, "-")
	return []string{"seg:" + seg}
}

// directionalKeys returns the conflict keys for a padding/margin class,
// or nil if rest is not a directional spacing suffix. rest is the class
// with the leading p/m and any "!"/"-" prefixes already removed, e.g.
// "-4" for "p-4" or "x-2" for "px-2".
func directionalKeys(family, rest string) []string {
	var sides []string
	switch {
	case len(rest) > 0 && rest[0] == '-':
		sides = allSides // unsuffixed form, e.g. "p-4"
	case len(rest) > 1 && rest[1] == '-':
		s, ok := spacingSides[rest[0]]
		if !ok {
			return nil // not a side letter, e.g. "placeholder-..."
		}
		sides = s
	default:
		return nil
	}
	keys := make([]string, len(sides))
	for i, side := range sides {
		keys[i] = family + ":" + side
	}
	return keys
}

// subsetOf reports whether every key in sub is also present in super.
func subsetOf(sub, super []string) bool {
	for _, key := range sub {
		if !slices.Contains(super, key) {
			return false
		}
	}
	return true
}
