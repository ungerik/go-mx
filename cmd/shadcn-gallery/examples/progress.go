package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func ProgressDemo() mx.Component {
	return html.DivClass("w-full max-w-md",
		shadcn.Progress(66),
	)
}
