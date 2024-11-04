package mx

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const RootElementName = "ROOT_ELEMENT"

func NewCheckedWriter(w io.Writer) *CheckedWriter {
	if w == nil {
		panic("nil io.Writer")
	}
	return &CheckedWriter{Writer: w, textEscaper: TextEscaper}
}

type elemState struct {
	element    string
	hasNewline bool
}

type CheckedWriter struct {
	io.Writer
	inStartTag   bool
	singleQuote  bool
	textEscaper  *strings.Replacer
	elemStack    []elemState
	allowedElems map[string]struct{}
	prefix       string
	indent       string
}

func (w *CheckedWriter) WithIndent(prefix, indent string) *CheckedWriter {
	w.prefix = prefix
	w.indent = indent
	return w
}

func (w *CheckedWriter) currentElemName() string {
	if len(w.elemStack) == 0 {
		return RootElementName
	}
	return w.elemStack[len(w.elemStack)-1].element
}

func (w *CheckedWriter) currentElemHasNewline() bool {
	if len(w.elemStack) == 0 {
		return false
	}
	return w.elemStack[len(w.elemStack)-1].hasNewline
}

func (w *CheckedWriter) BeginElement(elem string) error {
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
	_, err := w.Write(append([]byte{'<'}, elem...))
	return err
}

var (
	doubleQuote              = []byte{'"'}
	doubleQuoteAttribEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`"`, "&quot;",
	)
	singleQuote              = []byte{'\''}
	singleQuoteAttribEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`'`, "&apos;",
	)
)

func (w *CheckedWriter) Attribute(name, value string) error {
	if !w.inStartTag {
		return fmt.Errorf("can't write attribute while writing children of element %s", w.currentElemName())
	}
	if name == "" {
		return fmt.Errorf("empty attribute name")
	}
	var (
		quote   = []byte{'"'}
		escaper = doubleQuoteAttribEscaper
	)
	if w.singleQuote {
		quote = singleQuote
		escaper = singleQuoteAttribEscaper
	}
	_, err := fmt.Fprintf(w, ` %s=%s`, name, quote)
	if err != nil {
		return err
	}
	_, err = escaper.WriteString(w, value)
	if err != nil {
		return err
	}
	_, err = w.Write(quote)
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
		_, err = w.Write([]byte{'/', '>'})
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
