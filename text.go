package mx

import (
	"context"
	"fmt"
)

type Text string

func (t Text) Render(_ context.Context, w Writer) error {
	return w.EscapeText(string(t))
}

func Textf(format string, args ...any) Text {
	return Text(fmt.Sprintf(format, args...))
}
