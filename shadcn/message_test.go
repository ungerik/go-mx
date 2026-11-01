package shadcn

import (
	"strings"
	"testing"
)

func TestMessage(t *testing.T) {
	out := render(t, MessageGroup(
		Message("",
			MessageAvatar(Avatar(AvatarFallback("AI"))),
			MessageContent(
				MessageHeader("Assistant"),
				Bubble(BubbleMuted, "", BubbleContent("Hi")),
				MessageFooter("2:14 PM"),
			),
		),
	))
	for _, want := range []string{
		`data-slot="message-group"`, `data-slot="message"`, `data-align="start"`,
		`data-slot="message-avatar"`, `data-slot="message-content"`,
		`data-slot="message-header"`, `data-slot="message-footer"`, "group/message",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	if strings.Contains(out, "cn-") {
		t.Errorf("unresolved cn-* token leaked into output: %s", out)
	}
}

func TestMessageAlignEnd(t *testing.T) {
	out := render(t, Message(ChatAlignEnd, MessageContent(MessageFooter("Read"))))
	for _, want := range []string{`data-align="end"`, "data-[align=end]:flex-row-reverse",
		"group-data-[align=end]/message:justify-end"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestMessageHeaderFooterGhostScope(t *testing.T) {
	// Upstream drops the header/footer inset for any data-variant=ghost in the
	// message. This port's Button always emits data-variant, so a ghost
	// action button in the footer would trigger it: the selector is scoped
	// to a ghost bubble.
	out := render(t, MessageFooter("x"))
	if strings.Contains(out, "group-has-data-[variant=ghost]") {
		t.Errorf("ghost selector must be scoped to [data-slot=bubble]: %s", out)
	}
	if !strings.Contains(out, "group-has-[[data-slot=bubble][data-variant=ghost]]/message:px-0") {
		t.Errorf("missing bubble-scoped ghost selector: %s", out)
	}
}
