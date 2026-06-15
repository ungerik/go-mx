package mx

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// procInst renders a raw "<?target?>" processing instruction, the kind of
// content whose closing "?>" the CheckedWriter recognizes.
func procInst(s string) Component {
	return Raw("<?" + s + "?>")
}

func renderTo(t *testing.T, w *CheckedWriter, comps ...Component) string {
	t.Helper()
	require.NoError(t, Components(comps).Render(context.Background(), w))
	return w.Writer.(*strings.Builder).String()
}

func TestCheckedWriterProcInstNewline(t *testing.T) {
	t.Run("compact: declaration breaks before following element", func(t *testing.T) {
		var b strings.Builder
		got := renderTo(t, NewCheckedWriter(&b), procInst(`xml version="1.0"`), NewElement("root"))
		require.Equal(t, "<?xml version=\"1.0\"?>\n<root></root>", got)
	})

	t.Run("indented: no blank line after the declaration", func(t *testing.T) {
		var b strings.Builder
		w := NewCheckedWriter(&b).WithIndent("", "  ")
		got := renderTo(t, w, procInst(`xml version="1.0"`), NewElement("root", NewElement("child")))
		require.Equal(t, "<?xml version=\"1.0\"?>\n<root>\n  <child></child>\n</root>", got)
	})

	t.Run("consecutive processing instructions each get their own line", func(t *testing.T) {
		var b strings.Builder
		got := renderTo(t, NewCheckedWriter(&b), procInst("a"), procInst("b"), NewElement("root"))
		require.Equal(t, "<?a?>\n<?b?>\n<root></root>", got)
	})

	t.Run("a lone processing instruction gets no trailing newline", func(t *testing.T) {
		var b strings.Builder
		got := renderTo(t, NewCheckedWriter(&b), procInst("a"))
		require.Equal(t, "<?a?>", got)
	})

	t.Run("ordinary markup is unaffected", func(t *testing.T) {
		var b strings.Builder
		// A "?>" appearing inside text is escaped to "?&gt;", so it never ends a
		// write with the literal "?>" and triggers no extra newline.
		got := renderTo(t, NewCheckedWriter(&b), NewElement("p", Text("1 > 0 ?>"), NewElement("br")))
		require.Equal(t, "<p>1 &gt; 0 ?&gt;<br></br></p>", got)
	})
}
