package doc

import (
	"context"
	"io"
)

type Renderer interface {
	RenderParagraphOpening(context.Context, io.Writer, *Paragraph) error
	RenderParagraphChildren(context.Context, io.Writer, *Paragraph) error
	RenderParagraphClosing(context.Context, io.Writer, *Paragraph) error

	RenderTextOpening(context.Context, io.Writer, Text) error
	RenderTextChildren(context.Context, io.Writer, Text) error
	RenderTextClosing(context.Context, io.Writer, Text) error
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

// func JSONRenderer() Renderer {
// 	return jsonRenderer{}
// }

// type jsonRenderer struct{}

// func (jsonRenderer) RenderParagraphOpening(context.Context, io.Writer, *Paragraph) error  { return nil }
// func (jsonRenderer) RenderParagraphChildren(context.Context, io.Writer, *Paragraph) error { return nil }
// func (jsonRenderer) RenderParagraphClosing(context.Context, io.Writer, *Paragraph) error  { return nil }

// func (jsonRenderer) RenderTextOpening(ctx context.Context, w io.Writer, text Text) error {
// 	j, err := json.Marshal(string(text))
// 	if err != nil {
// 		return err
// 	}
// 	_, err = w.Write(j)
// 	return err
// }

// func (jsonRenderer) RenderTextChildren(ctx context.Context, w io.Writer, text Text) error { return nil }
// func (jsonRenderer) RenderTextClosing(ctx context.Context, w io.Writer, text Text) error  { return nil }
