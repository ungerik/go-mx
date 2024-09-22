package xml

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type Renderer interface {
	OpenElement(w io.Writer, elem string) error
	Attribute(w io.Writer, name, value string) error
	CloseElement(w io.Writer) error
	CloseVoidElement(w io.Writer) error
	ElementEnd(w io.Writer, elem string) error
}

var rendererKtxKey int

func RendererFromContext(ctx context.Context) Renderer {
	if r, _ := ctx.Value(&rendererKtxKey).(Renderer); r != nil {
		return r
	}
	return DefaultRenderer
}

func ContextWithRenderer(ctx context.Context, renderer Renderer) context.Context {
	return context.WithValue(ctx, &rendererKtxKey, renderer)
}

type BaseRenderer struct{}

func (BaseRenderer) OpenElement(w io.Writer, elem string) error {
	_, err := fmt.Fprintf(w, "<%s", elem)
	return err
}

var attribEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`<`, "&lt;",
	`"`, "&quot;",
)

func (BaseRenderer) Attribute(w io.Writer, name, value string) error {
	_, err := fmt.Fprintf(w, ` %s="`, name)
	if err != nil {
		return err
	}
	_, err = attribEscaper.WriteString(w, value)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte{'"'})
	return err
}

func (BaseRenderer) CloseElement(w io.Writer) error {
	_, err := w.Write([]byte{'>'})
	return err
}

func (BaseRenderer) CloseVoidElement(w io.Writer) error {
	_, err := w.Write([]byte{'/', '>'})
	return err
}

func (BaseRenderer) ElementEnd(w io.Writer, elem string) error {
	_, err := fmt.Fprintf(w, "</%s>", elem)
	return err
}
