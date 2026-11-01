package examples

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// MessageScrollerDemo renders a transcript that opens at the latest message.
// Scroll up and the button back to the end appears; new messages appended
// while you are at the bottom are followed.
func MessageScrollerDemo() mx.Component {
	items := []any{}
	for i := 1; i <= 8; i++ {
		items = append(items,
			shadcn.MessageScrollerItem(shadcn.Message(shadcn.ChatAlignEnd,
				shadcn.MessageContent(shadcn.Bubble("", "", shadcn.BubbleContent(fmt.Sprintf("Question %d: what changed in this step?", i)))),
			)),
			shadcn.MessageScrollerItem(shadcn.Message("",
				shadcn.MessageAvatar(shadcn.Avatar(shadcn.AvatarFallback("AI"))),
				shadcn.MessageContent(shadcn.Bubble(shadcn.BubbleMuted, "", shadcn.BubbleContent(
					fmt.Sprintf("Step %d renamed one function and updated its three callers; the tests still pass.", i),
				))),
			)),
		)
	}
	return html.DivClass("h-96 w-full max-w-md rounded-xl border",
		shadcn.MessageScroller(
			shadcn.MessageScrollerViewport(
				shadcn.MessageScrollerContent(append([]any{html.Class("p-4")}, items...)...),
			),
			shadcn.MessageScrollerButton("", ""),
		),
	)
}
