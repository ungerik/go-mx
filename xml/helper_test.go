package xml_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ungerik/go-mx"
)

// render renders c with a default (double-quoting) CheckedWriter and fails the
// test on a render error.
func render(t *testing.T, c mx.Component) string {
	t.Helper()
	var b strings.Builder
	require.NoError(t, c.Render(context.Background(), mx.NewCheckedWriter(&b)))
	return b.String()
}

// renderIndent renders c with a two-space indenting CheckedWriter.
func renderIndent(t *testing.T, c mx.Component) string {
	t.Helper()
	var b strings.Builder
	require.NoError(t, c.Render(context.Background(), mx.NewCheckedWriter(&b).WithIndent("", "  ")))
	return b.String()
}
