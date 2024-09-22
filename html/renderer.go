package html

import (
	"context"

	"github.com/ungerik/go-mx/xml"
)

type Renderer = xml.Renderer

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
