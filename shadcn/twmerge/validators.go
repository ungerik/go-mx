package twmerge

// Validators ported from tailwind-merge v3.6.0 (src/lib/validators.ts).
// They classify a class value fragment so the class-group trie can decide
// which utility group a class belongs to.

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	arbitraryValueRe    = regexp.MustCompile(`(?i)^\[(?:(\w[\w-]*):)?(.+)\]$`)
	arbitraryVariableRe = regexp.MustCompile(`(?i)^\((?:(\w[\w-]*):)?(.+)\)$`)
	fractionRe          = regexp.MustCompile(`^\d+(?:\.\d+)?/\d+(?:\.\d+)?$`)
	tshirtUnitRe        = regexp.MustCompile(`^(\d+(\.\d+)?)?(xs|sm|md|lg|xl)$`)
	lengthUnitRe        = regexp.MustCompile(`\d+(%|px|r?em|[sdl]?v([hwib]|min|max)|pt|pc|in|cm|mm|cap|ch|ex|r?lh|cq(w|h|i|b|min|max))|\b(calc|min|max|clamp)\(.+\)|^0$`)
	colorFunctionRe     = regexp.MustCompile(`^(rgba?|hsla?|hwb|(ok)?(lab|lch)|color-mix)\(.+\)$`)
	shadowRe            = regexp.MustCompile(`^(inset_)?-?((\d+)?\.?(\d+)[a-z]+|0)_-?((\d+)?\.?(\d+)[a-z]+|0)`)
	imageRe             = regexp.MustCompile(`^(url|image|image-set|cross-fade|element|(repeating-)?(linear|radial|conic)-gradient)\(.+\)$`)
)

func isFraction(v string) bool { return fractionRe.MatchString(v) }

// isNumber mirrors JS `!!value && !Number.isNaN(Number(value))`.
func isNumber(v string) bool {
	if v == "" {
		return false
	}
	f, err := strconv.ParseFloat(v, 64)
	return err == nil && !math.IsNaN(f)
}

// isInteger mirrors JS `!!value && Number.isInteger(Number(value))`.
func isInteger(v string) bool {
	if v == "" {
		return false
	}
	f, err := strconv.ParseFloat(v, 64)
	return err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) && f == math.Trunc(f)
}

func isPercent(v string) bool {
	return strings.HasSuffix(v, "%") && isNumber(v[:len(v)-1])
}

func isTshirtSize(v string) bool { return tshirtUnitRe.MatchString(v) }

func isAny(string) bool { return true }

func isNever(string) bool { return false }

// isLengthOnly excludes color functions, which may contain percentages that
// would otherwise be misclassified as lengths (e.g. `hsl(0 0% 0%)`).
func isLengthOnly(v string) bool {
	return lengthUnitRe.MatchString(v) && !colorFunctionRe.MatchString(v)
}

func isShadow(v string) bool { return shadowRe.MatchString(v) }

func isImage(v string) bool { return imageRe.MatchString(v) }

func isAnyNonArbitrary(v string) bool {
	return !isArbitraryValue(v) && !isArbitraryVariable(v)
}

func isNamedContainerQuery(v string) bool {
	if !strings.HasPrefix(v, "@container") {
		return false
	}
	at := func(i int) byte {
		if i < len(v) {
			return v[i]
		}
		return 0
	}
	return (at(10) == '/' && len(v) > 11) ||
		(at(11) == 's' && len(v) > 16 && strings.HasPrefix(v[10:], "-size/")) ||
		(at(11) == 'n' && len(v) > 18 && strings.HasPrefix(v[10:], "-normal/"))
}

func isArbitraryValue(v string) bool    { return arbitraryValueRe.MatchString(v) }
func isArbitraryVariable(v string) bool { return arbitraryVariableRe.MatchString(v) }

func isArbitrarySize(v string) bool       { return getIsArbitraryValue(v, isLabelSize, isNever) }
func isArbitraryLength(v string) bool     { return getIsArbitraryValue(v, isLabelLength, isLengthOnly) }
func isArbitraryNumber(v string) bool     { return getIsArbitraryValue(v, isLabelNumber, isNumber) }
func isArbitraryWeight(v string) bool     { return getIsArbitraryValue(v, isLabelWeight, isAny) }
func isArbitraryFamilyName(v string) bool { return getIsArbitraryValue(v, isLabelFamilyName, isNever) }
func isArbitraryPosition(v string) bool   { return getIsArbitraryValue(v, isLabelPosition, isNever) }
func isArbitraryImage(v string) bool      { return getIsArbitraryValue(v, isLabelImage, isImage) }
func isArbitraryShadow(v string) bool     { return getIsArbitraryValue(v, isLabelShadow, isShadow) }

func isArbitraryVariableLength(v string) bool {
	return getIsArbitraryVariable(v, isLabelLength, false)
}
func isArbitraryVariableFamilyName(v string) bool {
	return getIsArbitraryVariable(v, isLabelFamilyName, false)
}
func isArbitraryVariablePosition(v string) bool {
	return getIsArbitraryVariable(v, isLabelPosition, false)
}
func isArbitraryVariableSize(v string) bool {
	return getIsArbitraryVariable(v, isLabelSize, false)
}
func isArbitraryVariableImage(v string) bool {
	return getIsArbitraryVariable(v, isLabelImage, false)
}
func isArbitraryVariableShadow(v string) bool {
	return getIsArbitraryVariable(v, isLabelShadow, true)
}
func isArbitraryVariableWeight(v string) bool {
	return getIsArbitraryVariable(v, isLabelWeight, true)
}

func getIsArbitraryValue(v string, testLabel, testValue func(string) bool) bool {
	m := arbitraryValueRe.FindStringSubmatch(v)
	if m == nil {
		return false
	}
	if m[1] != "" {
		return testLabel(m[1])
	}
	return testValue(m[2])
}

func getIsArbitraryVariable(v string, testLabel func(string) bool, shouldMatchNoLabel bool) bool {
	m := arbitraryVariableRe.FindStringSubmatch(v)
	if m == nil {
		return false
	}
	if m[1] != "" {
		return testLabel(m[1])
	}
	return shouldMatchNoLabel
}

func isLabelPosition(l string) bool   { return l == "position" || l == "percentage" }
func isLabelImage(l string) bool      { return l == "image" || l == "url" }
func isLabelSize(l string) bool       { return l == "length" || l == "size" || l == "bg-size" }
func isLabelLength(l string) bool     { return l == "length" }
func isLabelNumber(l string) bool     { return l == "number" }
func isLabelFamilyName(l string) bool { return l == "family-name" }
func isLabelWeight(l string) bool     { return l == "number" || l == "weight" }
func isLabelShadow(l string) bool     { return l == "shadow" }
