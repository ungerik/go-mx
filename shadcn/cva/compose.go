// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go port of the compose function of cva v1.0.0-beta
// (class-variance-authority's successor), Copyright 2022 Joe Bell, licensed
// under the Apache License, Version 2.0. The function has been reimplemented
// in Go (a modified derivative work). See THIRD-PARTY-LICENSES.md for the
// full license text.

package cva

import "strings"

// Compose merges several [Variants] resolvers into one, so a component can
// build on another component's variant table instead of restating it.
//
// The composed resolver passes the same props to every resolver, minus the
// reserved "class" key, joins their non-empty results in argument order, and
// appends the props "class" override once at the end. Withholding "class"
// from the parts keeps a caller override from being repeated once per part
// and keeps it last, where a tailwind-merge step lets it win.
//
// Each part reads only the variant names its own [Config] declares, so props
// meant for one part are ignored by the others. Two parts that declare the
// same variant name both resolve it; their classes are concatenated, not
// merged, exactly as in the JavaScript original.
func Compose(parts ...Variants) Variants {
	return func(props map[string]string) string {
		partProps := props
		if _, ok := props["class"]; ok {
			// Copy rather than delete, so the caller's map is left as given.
			partProps = make(map[string]string, len(props))
			for name, value := range props {
				if name != "class" {
					partProps[name] = value
				}
			}
		}

		classes := make([]string, 0, len(parts)+1)
		for _, part := range parts {
			if cls := part(partProps); cls != "" {
				classes = append(classes, cls)
			}
		}
		if cls := props["class"]; cls != "" {
			classes = append(classes, cls)
		}
		return strings.Join(classes, " ")
	}
}
