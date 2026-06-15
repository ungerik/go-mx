package mx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

func NewCheckedWriter(dest io.Writer) *CheckedWriter {
	if dest == nil {
		dest = io.Discard
	}
	return &CheckedWriter{Writer: dest, writtenAttribs: make(map[string]struct{}), textEscaper: TextEscaper}
}

type elemState struct {
	element    string
	hasNewline bool
}

type CheckedWriter struct {
	// Configuration:
	io.Writer
	singleQuote             bool
	equalNameValueSkipValue bool
	textEscaper             *strings.Replacer
	allowedElems            map[string]struct{}
	prefix                  string
	indent                  string
	// Render state:
	inStartTag     bool
	writtenAttribs map[string]struct{}
	elemStack      []elemState
	afterProcInst  bool // previous Write ended a processing instruction ("?>")
}

func (w *CheckedWriter) Clone(dest io.Writer) *CheckedWriter {
	return &CheckedWriter{
		// Configuration:
		Writer:                  dest,
		singleQuote:             w.singleQuote,
		equalNameValueSkipValue: w.equalNameValueSkipValue,
		textEscaper:             w.textEscaper,
		allowedElems:            w.allowedElems,
		prefix:                  w.prefix,
		indent:                  w.indent,
		// Render state:
		inStartTag:     false,
		writtenAttribs: make(map[string]struct{}),
		elemStack:      nil,
		afterProcInst:  false,
	}
}

// Write writes p to the underlying writer, overriding the embedded io.Writer so
// the renderer can recognize XML processing instructions. A processing
// instruction or XML declaration (xml.Declaration, xml.ProcInst, …) is written
// raw and ends with "?>". When the previous Write ended one, a newline is
// inserted before the next written content so the declaration or instruction
// sits on its own line in both compact and indented output, instead of being
// glued to the following element. Indentation already breaks the line before
// the next element (via Newline, which clears the flag first), so no duplicate
// blank line is produced.
//
// When such an instruction is itself written inside an element's content (a
// processing instruction used as a child rather than in the document prolog),
// an indenting writer breaks and indents before it too, matching how
// BeginElement, Comment and CDATA indent themselves — Write is the only path
// that reaches the renderer without going through one of those.
//
// Escaped text and attribute values never end a Write with "?>" (">" is escaped
// to "&gt;" and attribute values are quote-terminated), so this is a no-op for
// ordinary markup. The exception is raw caller-supplied content (mx.Raw,
// mx.RawBytes) ending in "?>": it is treated like a processing instruction and
// gets the same line break.
//
// A declaration or instruction may already carry a trailing newline (notably
// the standard library's encoding/xml.Header, which is the XML declaration plus
// "\n"). That newline is dropped and the "?>" recognized, so such a value
// renders identically to the newline-free xml.Declaration: the separating line
// break is produced uniformly by the logic above rather than stacking on the
// value's own newline and the one an indenting writer inserts before the next
// element.
func (w *CheckedWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	procInst := bytes.HasSuffix(p, []byte("?>")) || bytes.HasSuffix(p, []byte("?>\n"))
	// A processing instruction written inside an element's content must be
	// broken and indented before it, just as BeginElement/Comment/CDATA indent
	// themselves. Newline clears afterProcInst, so the bare break below is
	// skipped and no duplicate newline results. Top-level instructions (an empty
	// element stack — a document's declaration and prolog) keep their compact,
	// un-indented spacing from the afterProcInst path instead.
	if procInst && w.indent != "" && !w.inStartTag && len(w.elemStack) > 0 {
		if err := w.Newline(); err != nil {
			return 0, err
		}
	}
	if w.afterProcInst {
		w.afterProcInst = false
		if _, err := w.Writer.Write([]byte{'\n'}); err != nil {
			return 0, err
		}
	}
	if bytes.HasSuffix(p, []byte("?>\n")) {
		// Strip the trailing newline and remember the instruction so the break
		// is emitted by the afterProcInst path on the next Write, exactly as for
		// a newline-free declaration. On success every byte of p is accounted
		// for (the dropped newline included), so report len(p); on a short write
		// report the count the underlying writer actually accepted.
		n, err := w.Writer.Write(p[:len(p)-1])
		if err != nil {
			return n, err
		}
		w.afterProcInst = true
		return len(p), nil
	}
	n, err := w.Writer.Write(p)
	if err == nil {
		w.afterProcInst = bytes.HasSuffix(p, []byte("?>"))
	}
	return n, err
}

func (w *CheckedWriter) WithIndent(prefix, indent string) *CheckedWriter {
	w.prefix = prefix
	w.indent = indent
	return w
}

func (w *CheckedWriter) WithSingleQuoteAttribs() *CheckedWriter {
	w.singleQuote = true
	return w
}

func (w *CheckedWriter) WithDoubleQuoteAttribs() *CheckedWriter {
	w.singleQuote = false
	return w
}

func (w *CheckedWriter) currentElemName() string {
	if len(w.elemStack) == 0 {
		return "ROOT_ELEMENT"
	}
	return w.elemStack[len(w.elemStack)-1].element
}

func (w *CheckedWriter) BeginElement(elem string) error {
	// TODO regex for valid element name
	if elem == "" {
		return fmt.Errorf("empty element name")
	}
	if w.inStartTag {
		return fmt.Errorf("can't BeginElement while writing attributes of element %s", w.currentElemName())
	}
	if w.allowedElems != nil {
		if _, ok := w.allowedElems[elem]; !ok {
			return fmt.Errorf("element %s not allowed", elem)
		}
	}
	if w.indent != "" {
		err := w.Newline()
		if err != nil {
			return err
		}
	}
	w.elemStack = append(w.elemStack, elemState{element: elem})
	w.inStartTag = true
	// Reset duplicate-attribute tracking for this element's start tag.
	// It must be cleared per start tag, not on EndElement: a nested child
	// begins its start tag while its parent (and ancestors) are still open.
	clear(w.writtenAttribs)
	_, err := w.Write(append([]byte{'<'}, elem...))
	return err
}

func (w *CheckedWriter) Attribute(name, value string) (err error) {
	if !w.inStartTag {
		return fmt.Errorf("can't write attribute while writing children of element %s", w.currentElemName())
	}
	// TODO regex for valid attribute name
	if name == "" {
		return fmt.Errorf("empty attribute name")
	}
	if _, duplicate := w.writtenAttribs[name]; duplicate {
		return fmt.Errorf("duplicate attribute %s in element %s", name, w.currentElemName())
	}
	w.writtenAttribs[name] = struct{}{}

	switch {
	case w.equalNameValueSkipValue && name == value:
		_, err = fmt.Fprintf(w, ` %s`, name)
	case w.singleQuote:
		_, err = fmt.Fprintf(w, ` %s='%s'`, name, singleQuoteAttribEscaper.Replace(value))
	default:
		_, err = fmt.Fprintf(w, ` %s="%s"`, name, doubleQuoteAttribEscaper.Replace(value))
	}
	return err
}

func (w *CheckedWriter) CloseElementStartTag() error {
	if len(w.elemStack) == 0 {
		return errors.New("can't CloseElementStartTag without BeginElement")
	}
	if !w.inStartTag {
		return fmt.Errorf("can't CloseElementStartTag while writing children of element %s", w.currentElemName())
	}
	w.inStartTag = false
	_, err := w.Write([]byte{'>'})
	return err
}

func (w *CheckedWriter) EndElement() (err error) {
	if len(w.elemStack) == 0 {
		return errors.New("can't EndElement without BeginElement")
	}
	this := w.elemStack[len(w.elemStack)-1]
	w.elemStack = w.elemStack[:len(w.elemStack)-1]
	if this.hasNewline {
		err = w.Newline()
		if err != nil {
			return err
		}
	}
	if w.inStartTag {
		// Void element
		w.inStartTag = false
		_, err = fmt.Fprint(w, "/>")
	} else {
		_, err = fmt.Fprintf(w, "</%s>", this.element)
	}
	return err
}

func (w *CheckedWriter) EscapeText(text string) error {
	if w.inStartTag {
		return fmt.Errorf("can't EscapeText while writing start tag of element %s", w.currentElemName())
	}
	_, err := w.textEscaper.WriteString(w, text)
	return err
}

func (w *CheckedWriter) Comment(text string) error {
	// TODO The text part of comments has the following restrictions:
	// must not start with a ">" character
	// must not start with the string "->"
	// must not contain the string "--"
	// must not end with a "-" character
	if w.indent != "" {
		err := w.Newline()
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "<!-- %s -->", text)
	return err
}

func (w *CheckedWriter) CDATA(text string) error {
	if strings.Contains(text, "]]>") {
		return fmt.Errorf("CDATA text contains ']]>': %s", text)
	}
	if w.indent != "" {
		err := w.Newline()
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "<![CDATA[%s]]>", text)
	return err
}

func (w *CheckedWriter) Newline() error {
	// Indentation provides this line break itself, so clear the
	// processing-instruction flag to avoid Write inserting a second newline.
	w.afterProcInst = false
	_, err := w.Write([]byte{'\n'})
	if err != nil {
		return err
	}
	if w.prefix != "" {
		_, err = w.Write([]byte(w.prefix))
		if err != nil {
			return err
		}
	}
	for range len(w.elemStack) {
		_, err = w.Write([]byte(w.indent))
		if err != nil {
			return err
		}
	}
	for i := range w.elemStack {
		w.elemStack[i].hasNewline = true
	}
	return nil
}
