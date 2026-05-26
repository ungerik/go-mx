package mx_test

import (
	"encoding/json"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// TestImportInvariant enforces the four-layer architecture from the
// ReflectFormHandler design (D1): mx must not import html, hx, or
// shadcn. html must not import hx or shadcn. hx must not import
// shadcn.
//
// Catching a violation here is cheap; catching it after a release
// requires renaming and re-exporting types. The test uses `go list
// -json` so the dependency graph comes from the toolchain instead of
// hand-rolled AST walking.
func TestImportInvariant(t *testing.T) {
	t.Parallel()

	cases := []struct {
		pkg       string
		forbidden []string
	}{
		{
			pkg: "github.com/ungerik/go-mx",
			forbidden: []string{
				"github.com/ungerik/go-mx/html",
				"github.com/ungerik/go-mx/hx",
				"github.com/ungerik/go-mx/shadcn",
			},
		},
		{
			pkg: "github.com/ungerik/go-mx/html",
			forbidden: []string{
				"github.com/ungerik/go-mx/hx",
				"github.com/ungerik/go-mx/shadcn",
			},
		},
		{
			pkg: "github.com/ungerik/go-mx/hx",
			forbidden: []string{
				"github.com/ungerik/go-mx/shadcn",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.pkg, func(t *testing.T) {
			imps := packageImports(t, c.pkg)
			for _, f := range c.forbidden {
				if slices.Contains(imps, f) {
					t.Errorf("%s must not import %s — direct import found", c.pkg, f)
				}
			}
		})
	}
}

// packageImports returns the (non-test) direct imports of pkg as
// reported by `go list -json`.
func packageImports(t *testing.T, pkg string) []string {
	t.Helper()
	out, err := exec.Command("go", "list", "-json", pkg).Output()
	if err != nil {
		var stderr string
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		t.Fatalf("go list -json %s: %v\n%s", pkg, err, stderr)
	}
	var info struct {
		ImportPath string
		Imports    []string
	}
	dec := json.NewDecoder(strings.NewReader(string(out)))
	if err := dec.Decode(&info); err != nil {
		t.Fatalf("decode go list output: %v", err)
	}
	return info.Imports
}
