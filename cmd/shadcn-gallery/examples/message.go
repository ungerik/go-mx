package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// MessageDemo renders a short exchange: the other party on the start side,
// the user's own messages on the end side.
func MessageDemo() mx.Component {
	return shadcn.MessageGroup(html.Class("w-full max-w-md"),
		shadcn.Message("",
			shadcn.MessageAvatar(shadcn.Avatar(shadcn.AvatarFallback("RS"))),
			shadcn.MessageContent(
				shadcn.Bubble(shadcn.BubbleMuted, "", shadcn.BubbleContent("Are we still on for the review at three?")),
			),
		),
		shadcn.Message(shadcn.ChatAlignEnd,
			shadcn.MessageContent(
				shadcn.BubbleGroup(
					shadcn.Bubble("", "", shadcn.BubbleContent("Yes!")),
					shadcn.Bubble("", "", shadcn.BubbleContent("I'll bring the updated numbers.")),
				),
			),
		),
	)
}

// MessageHeaderFooter renders an assistant reply with a name above and
// actions below. The reply is a ghost bubble, so header and footer drop
// their inset to line up with its text; the ghost action buttons do not
// trigger that.
func MessageHeaderFooter() mx.Component {
	return shadcn.MessageGroup(html.Class("w-full max-w-md"),
		shadcn.Message(shadcn.ChatAlignEnd,
			shadcn.MessageContent(
				shadcn.MessageHeader("You"),
				shadcn.Bubble("", "", shadcn.BubbleContent("Why is the deploy so slow?")),
				shadcn.MessageFooter("Read 2:14 PM"),
			),
		),
		shadcn.Message("",
			shadcn.MessageContent(
				shadcn.MessageHeader("Assistant"),
				shadcn.Bubble(shadcn.BubbleGhost, "", shadcn.BubbleContent(
					"Image optimization reruns on every deploy because its cache key includes the commit hash. Key it on the lockfile and the image directory instead.",
				)),
				shadcn.MessageFooter(html.Class("gap-1"),
					shadcn.Button(shadcn.ButtonGhost, shadcn.SizeIconXS, html.Attrib("aria-label", "Copy"), iconCopy()),
					shadcn.Button(shadcn.ButtonGhost, shadcn.SizeIconXS, html.Attrib("aria-label", "Retry"), iconRetry()),
				),
			),
		),
	)
}
