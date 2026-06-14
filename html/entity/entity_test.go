package entity_test

import (
	"context"
	"os"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html/entity"
)

func Example() {
	mx.Components{
		entity.Copyright,
		entity.Registered,
		entity.Trademark,
		entity.Euro,
		entity.RightArrow,
		entity.NonBreakingSpace,
	}.Render(context.Background(), mx.NewCheckedWriter(os.Stdout))

	// Output: &copy;&reg;&trade;&euro;&rarr;&nbsp;
}
