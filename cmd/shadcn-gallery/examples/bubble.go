package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// BubbleVariants renders a bubble in each of the seven variants.
func BubbleVariants() mx.Component {
	return html.DivClass("flex w-full max-w-md flex-col gap-2",
		shadcn.Bubble("", "", shadcn.BubbleContent("Default")),
		shadcn.Bubble(shadcn.BubbleSecondary, "", shadcn.BubbleContent("Secondary")),
		shadcn.Bubble(shadcn.BubbleMuted, "", shadcn.BubbleContent("Muted")),
		shadcn.Bubble(shadcn.BubbleTinted, "", shadcn.BubbleContent("Tinted")),
		shadcn.Bubble(shadcn.BubbleOutline, "", shadcn.BubbleContent("Outline")),
		shadcn.Bubble(shadcn.BubbleGhost, "", shadcn.BubbleContent("Ghost, for an assistant's reply at full width")),
		shadcn.Bubble(shadcn.BubbleDestructive, "", shadcn.BubbleContent("Message failed to send")),
	)
}

// BubbleAlign renders bubbles on both sides of a conversation.
func BubbleAlign() mx.Component {
	return shadcn.BubbleGroup(html.Class("w-full max-w-md"),
		shadcn.Bubble(shadcn.BubbleMuted, "", shadcn.BubbleContent("Did the build pass?")),
		shadcn.Bubble("", shadcn.ChatAlignEnd, shadcn.BubbleContent("Green on the second try.")),
		shadcn.Bubble("", shadcn.ChatAlignEnd, shadcn.BubbleContent("The first run hit a flaky test.")),
	)
}

// BubbleReactionsDemo renders reactions overlapping a bubble's edges.
func BubbleReactionsDemo() mx.Component {
	return html.DivClass("flex w-full max-w-md flex-col gap-8 py-4",
		shadcn.Bubble(shadcn.BubbleMuted, "",
			shadcn.BubbleContent("Shipped the new onboarding flow 🎉"),
			shadcn.BubbleReactions("", "", "🎉 3"),
		),
		shadcn.Bubble("", shadcn.ChatAlignEnd,
			shadcn.BubbleContent("Nice work, everyone."),
			shadcn.BubbleReactions(shadcn.BubbleReactionsTop, shadcn.ChatAlignStart, "❤️"),
		),
	)
}

// BubbleSuggestions renders suggested replies as bubbles the user can click:
// a type attribute makes BubbleContent a <button>.
func BubbleSuggestions() mx.Component {
	suggestion := func(text string) mx.Component {
		return shadcn.Bubble(shadcn.BubbleOutline, shadcn.ChatAlignEnd,
			shadcn.BubbleContent(html.Type("button"), text),
		)
	}
	return shadcn.BubbleGroup(html.Class("w-full max-w-md"),
		suggestion("Show me the failing test"),
		suggestion("Retry the build"),
		suggestion("Open the logs"),
	)
}
