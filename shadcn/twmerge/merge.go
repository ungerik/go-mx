// Package twmerge is a Go port of tailwind-merge v3.6.0 (targeting Tailwind
// CSS v4): [Merge] resolves Tailwind utility-class conflicts so that a later
// class overrides an earlier conflicting one.
//
// It is a faithful transcription of the upstream TypeScript — the full default
// config (every utility class group, the class-group trie, modifier /
// important / postfix parsing and the conflict tables) is spread across
// merge.go, classmap.go, validators.go, parse.go and defaultconfig.go.
//
// twmerge is the merge half of the shadcn cn helper; the flatten half is
// package clsx. See [github.com/ungerik/go-mx/shadcn.Cn].
package twmerge

// The tailwind-merge conflict-resolution algorithm, ported from
// tailwind-merge v3.6.0 (src/lib/merge-classlist.ts and config-utils.ts).

import (
	"slices"
	"strings"
)

// twConfig is the data half of a tailwind-merge configuration.
type twConfig struct {
	theme                          map[string]classGroup
	classGroups                    []classGroupEntry // order-significant
	conflictingClassGroups         map[string][]string
	conflictingClassGroupModifiers map[string][]string
	postfixLookupClassGroups       []string
	orderSensitiveModifiers        []string
}

// classGroupEntry is one classGroups map entry. A slice is used instead of a
// map because the processing order of class groups determines validator
// precedence inside the class map.
type classGroupEntry struct {
	id    string
	group classGroup
}

// twState is the prepared, read-only runtime state. It is built once and only
// read afterwards, so concurrent Merge calls are safe.
type twState struct {
	classMap                       *classPart
	conflictingClassGroups         map[string][]string
	conflictingClassGroupModifiers map[string][]string
	postfixLookup                  map[string]bool
	orderSensitive                 map[string]struct{}
}

var state = buildState()

func buildState() *twState {
	cfg := getDefaultConfig()
	s := &twState{
		conflictingClassGroups:         cfg.conflictingClassGroups,
		conflictingClassGroupModifiers: cfg.conflictingClassGroupModifiers,
		postfixLookup:                  make(map[string]bool, len(cfg.postfixLookupClassGroups)),
		orderSensitive:                 make(map[string]struct{}, len(cfg.orderSensitiveModifiers)),
	}
	for _, id := range cfg.postfixLookupClassGroups {
		s.postfixLookup[id] = true
	}
	for _, m := range cfg.orderSensitiveModifiers {
		s.orderSensitive[m] = struct{}{}
	}
	s.classMap = buildClassMap(cfg.theme, cfg.classGroups)
	return s
}

// Merge merges a space-separated class list, removing earlier classes that a
// later class overrides. It is the Go port of tailwind-merge's twMerge.
func Merge(classList string) string {
	classNames := strings.Fields(classList)

	// conflictKeys is a set of `{modifierId}{classGroupId}` strings already
	// seen; any class mapping to one of them is a conflict and dropped.
	conflictKeys := make(map[string]struct{}, len(classNames))
	kept := make([]string, 0, len(classNames))

	// Iterate from last to first so the last occurrence of each class wins.
	for _, original := range slices.Backward(classNames) {
		p := parseClassName(original)

		hasPostfixModifier := p.postfixModifierPos != 0
		var classGroupID string
		if hasPostfixModifier {
			end := min(p.postfixModifierPos, len(p.baseClassName))
			classGroupID = getClassGroupID(p.baseClassName[:end])

			var withPostfix string
			if classGroupID != "" && state.postfixLookup[classGroupID] {
				withPostfix = getClassGroupID(p.baseClassName)
			}
			if withPostfix != "" && withPostfix != classGroupID {
				classGroupID = withPostfix
				hasPostfixModifier = false
			}
		} else {
			classGroupID = getClassGroupID(p.baseClassName)
		}

		if classGroupID == "" {
			if !hasPostfixModifier {
				kept = append(kept, original) // not a Tailwind class
				continue
			}
			classGroupID = getClassGroupID(p.baseClassName)
			if classGroupID == "" {
				kept = append(kept, original) // not a Tailwind class
				continue
			}
			hasPostfixModifier = false
		}

		var variantModifier string
		switch len(p.modifiers) {
		case 0:
			variantModifier = ""
		case 1:
			variantModifier = p.modifiers[0]
		default:
			variantModifier = strings.Join(sortModifiers(p.modifiers), ":")
		}
		modifierID := variantModifier
		if p.hasImportantModifier {
			modifierID += "!"
		}
		classID := modifierID + classGroupID

		if _, conflict := conflictKeys[classID]; conflict {
			continue // omitted due to conflict
		}
		conflictKeys[classID] = struct{}{}
		for _, g := range getConflictingClassGroupIDs(classGroupID, hasPostfixModifier) {
			conflictKeys[modifierID+g] = struct{}{}
		}
		kept = append(kept, original)
	}

	// kept was collected last-to-first; reverse to restore source order.
	for l, r := 0, len(kept)-1; l < r; l, r = l+1, r-1 {
		kept[l], kept[r] = kept[r], kept[l]
	}
	return strings.Join(kept, " ")
}
