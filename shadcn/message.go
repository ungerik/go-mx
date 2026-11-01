//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// Message is a Go port of shadcn/ui's Message, the layout of one chat turn:
// an avatar next to a column of content, with an optional header and footer.
// The content is usually one or more [Bubble]s. Class strings are
// reconstructed from the post-July-2026 registry (message.tsx skeleton + the
// style-vega.css .cn-message* rules) — see "Upstream restructure (July 2026)"
// in TODOS.md.

// ChatAlign is the side a [Message], [Bubble] or [BubbleReactions] sits on.
// Upstream gives each of them its own start/end align prop; one type serves
// all three here.
type ChatAlign string //#enum

const (
	// ChatAlignStart puts the element on the inline-start side, where the
	// other party's messages go.
	ChatAlignStart ChatAlign = "start"
	// ChatAlignEnd puts the element on the inline-end side, where the user's
	// own messages go.
	ChatAlignEnd ChatAlign = "end"
)

// Valid indicates if c is any of the valid values for ChatAlign
func (c ChatAlign) Valid() bool {
	switch c {
	case
		ChatAlignStart,
		ChatAlignEnd:
		return true
	}
	return false
}

// Validate returns an error if c is none of the valid values for ChatAlign
func (c ChatAlign) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.ChatAlign", c)
	}
	return nil
}

// ValidOr returns c if it is any of the valid values for ChatAlign, else defaultVal
func (c ChatAlign) ValidOr(defaultVal ChatAlign) ChatAlign {
	if c.Valid() {
		return c
	}
	return defaultVal
}

// Enums returns all valid values for ChatAlign
func (ChatAlign) Enums() []ChatAlign {
	return []ChatAlign{
		ChatAlignStart,
		ChatAlignEnd,
	}
}

// EnumStrings returns all valid values for ChatAlign as strings
func (ChatAlign) EnumStrings() []string {
	return []string{
		"start",
		"end",
	}
}

// String implements the fmt.Stringer interface for ChatAlign
func (c ChatAlign) String() string {
	return string(c)
}

// MessageGroup stacks consecutive [Message]s.
func MessageGroup(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "message-group",
		"flex min-w-0 flex-col gap-2")
}

// Message renders one chat turn. align may be "" for [ChatAlignStart]; an
// unknown value also resolves to it. [ChatAlignEnd] mirrors the row, so the
// avatar goes to the end and the content's bubbles and footer align there.
// Compose it with [MessageAvatar] and [MessageContent].
func Message(align ChatAlign, attribsChildren ...any) *mx.Element {
	a := string(align.ValidOr(ChatAlignStart))
	e := html.Div(attribsChildren...)
	e.Attribs = append(e.Attribs, html.DataAttr("align", a))
	return finish(e, "message",
		"group/message relative flex w-full min-w-0 data-[align=end]:flex-row-reverse text-sm gap-2")
}

// MessageAvatar holds a [Message]'s avatar, typically an [Avatar], aligned to
// the last bubble; with a [MessageFooter] it moves up to stay beside the
// bubbles.
func MessageAvatar(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "message-avatar",
		"flex w-fit shrink-0 items-center justify-center self-end overflow-hidden rounded-full bg-muted min-w-8 group-has-data-[slot=message-footer]/message:-translate-y-8")
}

// MessageContent holds a [Message]'s [MessageHeader], bubbles and
// [MessageFooter].
func MessageContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "message-content",
		"flex w-full min-w-0 flex-col wrap-break-word gap-2.5 group-data-[align=end]/message:*:data-slot:self-end")
}

// messageHeaderFooterClasses is shared by [MessageHeader] and
// [MessageFooter]. Upstream drops the inset when the message holds any
// element with data-variant="ghost"; this port's [Button] always emits a
// data-variant, so a ghost copy or retry button in the footer would trigger
// it, and the selector is scoped to a ghost [Bubble].
const messageHeaderFooterClasses = "flex max-w-full min-w-0 items-center text-xs font-medium text-muted-foreground px-3 group-has-[[data-slot=bubble][data-variant=ghost]]/message:px-0"

// MessageHeader renders a line above a [Message]'s bubbles, such as the
// sender's name.
func MessageHeader(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "message-header", messageHeaderFooterClasses)
}

// MessageFooter renders a line below a [Message]'s bubbles, such as a time
// or actions; in a [ChatAlignEnd] message it aligns to the end.
func MessageFooter(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "message-footer",
		Cn(messageHeaderFooterClasses, "group-data-[align=end]/message:justify-end"))
}
