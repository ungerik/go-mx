package xml

import (
	"context"
	"fmt"
	"strings"

	"github.com/domonda/go-errs"

	"github.com/ungerik/go-mx"
)

// Declaration is the standard XML declaration that begins most documents:
// <?xml version="1.0" encoding="UTF-8"?>. It is a bare [Raw] value with no
// trailing newline; the [mx.CheckedWriter] recognizes the closing "?>" and
// breaks the line after it, so the document content starts on the next line in
// both compact and indented output. Use [Decl] to build one with a different
// version or encoding.
const Declaration Raw = `<?xml version="1.0" encoding="UTF-8"?>`

// Decl builds an XML declaration <?xml version="..." encoding="..."?> for the
// given version and encoding. The encoding clause is omitted when encoding is
// empty. For the common UTF-8 case use the [Declaration] constant.
func Decl(version, encoding string) Raw {
	if encoding == "" {
		return Raw(fmt.Sprintf("<?xml version=%q?>", version))
	}
	return Raw(fmt.Sprintf("<?xml version=%q encoding=%q?>", version, encoding))
}

// ProcInst is an XML processing instruction rendered as <?Target Data?>; the
// space and Data are omitted when Data is empty. It carries instructions to the
// application processing the document, for example a stylesheet link:
//
//	xml.ProcInst{Target: "xml-stylesheet", Data: `type="text/xsl" href="t.xsl"`}
//	// <?xml-stylesheet type="text/xsl" href="t.xsl"?>
//
// The Target names the application; it must be a non-empty XML name, must not
// equal "xml" (case-insensitively, which is reserved for the [Declaration]) and
// must not contain whitespace. Neither Target nor Data may contain the
// terminator "?>". A ProcInst violating these rules fails to render with an
// error rather than emitting malformed markup.
type ProcInst struct {
	Target string
	Data   string
}

var _ mx.Component = ProcInst{}

// Render writes the processing instruction, implementing [mx.Component].
func (pi ProcInst) Render(_ context.Context, w mx.Writer) error {
	switch {
	case pi.Target == "":
		return errs.New("xml.ProcInst with empty Target")
	case strings.EqualFold(pi.Target, "xml"):
		return errs.Errorf("xml.ProcInst Target %q is reserved; use xml.Declaration or xml.Decl", pi.Target)
	case strings.ContainsAny(pi.Target, " \t\r\n"):
		return errs.Errorf("xml.ProcInst Target must not contain whitespace: %q", pi.Target)
	case strings.Contains(pi.Target, "?>") || strings.Contains(pi.Data, "?>"):
		return errs.Errorf("xml.ProcInst must not contain %q", "?>")
	}
	if pi.Data == "" {
		_, err := fmt.Fprintf(w, "<?%s?>", pi.Target)
		return err
	}
	_, err := fmt.Fprintf(w, "<?%s %s?>", pi.Target, pi.Data)
	return err
}

// Doctype renders a document type declaration <!DOCTYPE text>. The text is
// written verbatim, so it carries the root element name and any external or
// internal subset:
//
//	xml.Doctype("note")                     // <!DOCTYPE note>
//	xml.Doctype(`note SYSTEM "note.dtd"`)    // <!DOCTYPE note SYSTEM "note.dtd">
func Doctype(text string) Raw {
	return Raw("<!DOCTYPE " + text + ">")
}
