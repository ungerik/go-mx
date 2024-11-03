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

type CheckedWriter struct {
	io.Writer
	writingAttribs bool
	singleQuote    bool
	textEscaper    *strings.Replacer
	parentElems    []string
	allowedElems   map[string]struct{}
	prefix         string
	indent         string
}

func (w *CheckedWriter) WithIndet(prefix, indent string) *CheckedWriter {
	w.prefix = prefix
	w.indent = indent
	return w
}

func (w *CheckedWriter) currentElem() string {
	if len(w.parentElems) == 0 {
		return RootElementName
	}
	return w.parentElems[len(w.parentElems)-1]
}

func (w *CheckedWriter) BeginElement(elem string) error {
	if elem == "" {
		return fmt.Errorf("empty element name")
	}
	if w.writingAttribs {
		return fmt.Errorf("can't BeginElement while writing attributes of element %s", w.currentElem())
	}
	if w.allowedElems != nil {
		if _, ok := w.allowedElems[elem]; !ok {
			return fmt.Errorf("element %s not allowed", elem)
		}
	}
	w.parentElems = append(w.parentElems, elem)
	w.writingAttribs = true
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
	if !w.writingAttribs {
		return fmt.Errorf("can't write attribute while writing children of element %s", w.currentElem())
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

func (w *CheckedWriter) CloseAndEndElement() error {
	if len(w.parentElems) == 0 {
		return errors.New("can't CloseAndEndElement without BeginElement")
	}
	if !w.writingAttribs {
		return fmt.Errorf("can't CloseAndEndElement while writing children of element %s", w.currentElem())
	}
	w.parentElems = w.parentElems[:len(w.parentElems)-1]
	w.writingAttribs = false
	_, err := w.Write([]byte{'/', '>'})
	if err == nil && w.indent != "" {
		err = w.Newline()
	}
	return err
}

func (w *CheckedWriter) CloseElement() error {
	if len(w.parentElems) == 0 {
		return errors.New("can't CloseElement without BeginElement")
	}
	if !w.writingAttribs {
		return fmt.Errorf("can't CloseElement while writing children of element %s", w.currentElem())
	}
	w.writingAttribs = false
	_, err := w.Write([]byte{'>'})
	return err
}

func (w *CheckedWriter) EscapeText(text string) error {
	if w.writingAttribs {
		return fmt.Errorf("can't EscapeText while writing attributes of element %s", w.currentElem())
	}
	_, err := w.textEscaper.WriteString(w, text)
	return err
}

func (w *CheckedWriter) EndElement() error {
	if len(w.parentElems) == 0 {
		return errors.New("can't EndElement without BeginElement")
	}
	if w.writingAttribs {
		return fmt.Errorf("can't EndElement while writing attributes of element %s", w.currentElem())
	}
	elem := w.currentElem()
	w.parentElems = w.parentElems[:len(w.parentElems)-1]
	_, err := fmt.Fprintf(w, "</%s>", elem)
	if err == nil && w.indent != "" {
		err = w.Newline()
	}
	return err
}

func (w *CheckedWriter) Comment(text string) error {
	// TODO The text part of comments has the following restrictions:
	// must not start with a ">" character
	// must not start with the string "->"
	// must not contain the string "--"
	// must not end with a "-" character
	_, err := fmt.Fprintf(w, "<!-- %s -->", text)
	if err == nil && w.indent != "" {
		err = w.Newline()
	}
	return err
}

func (w *CheckedWriter) CDATA(text string) error {
	if strings.Contains(text, "]]>") {
		return fmt.Errorf("CDATA text contains ']]>': %s", text)
	}
	_, err := fmt.Fprintf(w, "<![CDATA[%s]]>", text)
	if err == nil && w.indent != "" {
		err = w.Newline()
	}
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
	for range len(w.parentElems) {
		_, err = w.Write([]byte(w.indent))
		if err != nil {
			return err
		}
	}
	return nil
}
