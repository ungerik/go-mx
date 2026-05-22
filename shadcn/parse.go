package shadcn

// Class-name parsing and modifier sorting, ported from tailwind-merge v3.6.0
// (src/lib/parse-class-name.ts and src/lib/sort-modifiers.ts).

import (
	"sort"
	"strings"
)

type parsedClassName struct {
	modifiers            []string
	hasImportantModifier bool
	baseClassName        string
	// postfixModifierPos is the position of a `/` postfix modifier relative to
	// the start of baseClassNameWithImportantModifier, or 0 when there is none.
	// It matches tailwind-merge's `maybePostfixModifierPosition` exactly,
	// including the off-by-one when a legacy leading `!` is present.
	postfixModifierPos int
}

// parseClassName splits a class name into variant modifiers, the base class
// name, an important flag and an optional postfix modifier position.
//
// Modifier separators (`:`) and postfix separators (`/`) are only recognized
// at the top nesting level, so `[&:hover]` and `(--my/var)` are left intact.
func parseClassName(className string) parsedClassName {
	var modifiers []string
	bracketDepth := 0
	parenDepth := 0
	modifierStart := 0
	postfixModifierPosition := -1

	for i := 0; i < len(className); i++ {
		c := className[i]
		if bracketDepth == 0 && parenDepth == 0 {
			if c == ':' {
				modifiers = append(modifiers, className[modifierStart:i])
				modifierStart = i + 1
				continue
			}
			if c == '/' {
				postfixModifierPosition = i
				continue
			}
		}
		switch c {
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		}
	}

	baseWithImportant := className
	if len(modifiers) > 0 {
		baseWithImportant = className[modifierStart:]
	}

	baseClassName := baseWithImportant
	hasImportant := false
	switch {
	case strings.HasSuffix(baseWithImportant, "!"):
		baseClassName = baseWithImportant[:len(baseWithImportant)-1]
		hasImportant = true
	case strings.HasPrefix(baseWithImportant, "!"):
		// Tailwind CSS v3 legacy leading-`!` syntax, still supported.
		baseClassName = baseWithImportant[1:]
		hasImportant = true
	}

	postfixPos := 0
	if postfixModifierPosition > modifierStart {
		postfixPos = postfixModifierPosition - modifierStart
	}

	return parsedClassName{
		modifiers:            modifiers,
		hasImportantModifier: hasImportant,
		baseClassName:        baseClassName,
		postfixModifierPos:   postfixPos,
	}
}

// sortModifiers sorts predefined modifiers alphabetically while preserving the
// position of arbitrary variants (`[...]`) and order-sensitive modifiers, which
// act as boundaries between independently sorted segments.
func sortModifiers(modifiers []string) []string {
	result := make([]string, 0, len(modifiers))
	var segment []string
	flush := func() {
		if len(segment) > 0 {
			sort.Strings(segment)
			result = append(result, segment...)
			segment = nil
		}
	}
	for _, m := range modifiers {
		isArbitrary := len(m) > 0 && m[0] == '['
		_, isOrderSensitive := state.orderSensitive[m]
		if isArbitrary || isOrderSensitive {
			flush()
			result = append(result, m)
		} else {
			segment = append(segment, m)
		}
	}
	flush()
	return result
}
