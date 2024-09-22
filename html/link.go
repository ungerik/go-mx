package html

import (
	"context"
	"io"
	"net/http"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/xml"
)

var _ mx.Component = Link{}

type Link struct {
	Href           string `attr:"href"`
	Rel            string `attr:"rel"`
	Type           string `attr:"type"`
	As             string `attr:"as"`
	Sizes          string `attr:"sizes"`
	ImageSizes     string `attr:"imagesizes"`
	ImageSRCSet    string `attr:"imagesrcset"`
	Crossorigin    string `attr:"crossorigin"`
	FetchPriority  string `attr:"fetchpriority"`
	ReferrerPolicy string `attr:"referrerpolicy"`
	Integrity      string `attr:"integrity"`
	Media          string `attr:"media"`
	Title          string `attr:"title"`
}

func (link Link) Render(ctx context.Context, w io.Writer) error {
	renderer := RendererFromContext(ctx)
	return xml.WriteStructAsVoidElement(w, renderer, "link", &link)
}

func (link Link) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.ServeComponent(w, r, contentTypeHTML, &link)
}
