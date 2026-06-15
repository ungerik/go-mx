package mx

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// shortWriter accepts at most limit bytes total, then reports a short write:
// it returns the number of bytes it actually accepted along with an error. It
// exists to verify CheckedWriter.Write propagates the underlying writer's count
// instead of discarding it.
type shortWriter struct {
	limit   int
	written int
}

func (s *shortWriter) Write(p []byte) (int, error) {
	n := len(p)
	if s.written+n > s.limit {
		n = s.limit - s.written
	}
	s.written += n
	if n < len(p) {
		return n, errors.New("short write")
	}
	return n, nil
}

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

	t.Run(`trailing newline after "?>" is dropped (encoding/xml.Header)`, func(t *testing.T) {
		var b strings.Builder
		// A declaration carrying its own trailing newline (like the standard
		// library's encoding/xml.Header) renders identically to the newline-free
		// form: the "\n" is dropped and the single break before the next element
		// is produced by the writer, not stacked on the value's own newline.
		got := renderTo(t, NewCheckedWriter(&b), Raw("<?xml version=\"1.0\"?>\n"), NewElement("root"))
		require.Equal(t, "<?xml version=\"1.0\"?>\n<root></root>", got)
	})

	t.Run(`trailing newline after "?>" with nothing following adds no break`, func(t *testing.T) {
		var b strings.Builder
		got := renderTo(t, NewCheckedWriter(&b), Raw("<?a?>\n"))
		require.Equal(t, "<?a?>", got)
	})

	t.Run("indented: a processing-instruction child is broken and indented like a sibling", func(t *testing.T) {
		var b strings.Builder
		w := NewCheckedWriter(&b).WithIndent("", "  ")
		// A top-level instruction stays at column 0, but one used as an element
		// child is indented to its siblings' depth instead of being glued to the
		// parent's start tag.
		got := renderTo(t, w, procInst("pi"), NewElement("root", procInst("child-pi"), NewElement("child")))
		require.Equal(t, "<?pi?>\n<root>\n  <?child-pi?>\n  <child></child>\n</root>", got)
	})

	t.Run(`short write in the "?>\n" branch reports bytes actually written`, func(t *testing.T) {
		// The branch strips the trailing newline and writes the 5 leading bytes
		// "<?a?>"; the sink accepts only 3 and errors. Write must report the 3
		// bytes the underlying writer accepted, not 0.
		w := NewCheckedWriter(&shortWriter{limit: 3})
		n, err := w.Write([]byte("<?a?>\n"))
		require.Error(t, err)
		require.Equal(t, 3, n)
	})
}
