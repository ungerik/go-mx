package shadcn

// Class-group resolution, ported from tailwind-merge v3.6.0
// (src/lib/class-group-utils.ts). A trie ("class map") is built from the
// config's classGroups and used to resolve any class name to its group id.

import "strings"

// classGroup is a list of class definitions. Each element is one of:
// string, themeGetter, classObject, or a validator func(string) bool.
type classGroup = []any

// classObject maps a class-name path segment to a nested classGroup.
type classObject map[string]classGroup

// themeGetter references a theme scale by key; it expands to that scale's
// classGroup when the class map is built.
type themeGetter struct{ key string }

type classPart struct {
	nextPart     map[string]*classPart
	validators   []classValidatorObj
	classGroupID string
}

type classValidatorObj struct {
	classGroupID string
	validator    func(string) bool
}

func newClassPart() *classPart {
	return &classPart{nextPart: map[string]*classPart{}}
}

// getPart walks (creating as needed) to the trie node for a `-`-separated path.
func (cp *classPart) getPart(path string) *classPart {
	current := cp
	for part := range strings.SplitSeq(path, "-") {
		next, ok := current.nextPart[part]
		if !ok {
			next = newClassPart()
			current.nextPart[part] = next
		}
		current = next
	}
	return current
}

func buildClassMap(theme map[string]classGroup, classGroups []classGroupEntry) *classPart {
	root := newClassPart()
	for _, entry := range classGroups {
		processClasses(entry.group, root, entry.id, theme)
	}
	return root
}

func processClasses(group classGroup, node *classPart, classGroupID string, theme map[string]classGroup) {
	for _, def := range group {
		switch d := def.(type) {
		case string:
			target := node
			if d != "" {
				target = node.getPart(d)
			}
			target.classGroupID = classGroupID
		case themeGetter:
			processClasses(theme[d.key], node, classGroupID, theme)
		case classObject:
			// Object keys are independent trie branches, so iteration
			// order does not affect the resulting class map.
			for key, sub := range d {
				processClasses(sub, node.getPart(key), classGroupID, theme)
			}
		case func(string) bool:
			node.validators = append(node.validators, classValidatorObj{classGroupID, d})
		}
	}
}

// getClassGroupID resolves a class name (without modifiers) to its class-group
// id, or "" if it is not a recognized Tailwind class.
func getClassGroupID(className string) string {
	if strings.HasPrefix(className, "[") && strings.HasSuffix(className, "]") {
		return getGroupIDForArbitraryProperty(className)
	}
	classParts := strings.Split(className, "-")
	startIndex := 0
	// Classes like `-inset-1` produce an empty first part; skip it.
	if classParts[0] == "" && len(classParts) > 1 {
		startIndex = 1
	}
	return getGroupRecursive(classParts, startIndex, state.classMap)
}

func getGroupRecursive(classParts []string, startIndex int, node *classPart) string {
	if len(classParts)-startIndex == 0 {
		return node.classGroupID
	}
	current := classParts[startIndex]
	if next, ok := node.nextPart[current]; ok {
		if res := getGroupRecursive(classParts, startIndex+1, next); res != "" {
			return res
		}
	}
	if len(node.validators) == 0 {
		return ""
	}
	classRest := strings.Join(classParts[startIndex:], "-")
	for _, vo := range node.validators {
		if vo.validator(classRest) {
			return vo.classGroupID
		}
	}
	return ""
}

func getGroupIDForArbitraryProperty(className string) string {
	content := className[1 : len(className)-1]
	property, _, found := strings.Cut(content, ":")
	if !found || property == "" {
		return ""
	}
	// Two dots: one dot is reserved as the class-group prefix for plugins.
	return "arbitrary.." + property
}

func getConflictingClassGroupIDs(classGroupID string, hasPostfixModifier bool) []string {
	if hasPostfixModifier {
		mod := state.conflictingClassGroupModifiers[classGroupID]
		base := state.conflictingClassGroups[classGroupID]
		if mod != nil {
			if base != nil {
				out := make([]string, 0, len(base)+len(mod))
				out = append(out, base...)
				return append(out, mod...)
			}
			return mod
		}
		return base
	}
	return state.conflictingClassGroups[classGroupID]
}
