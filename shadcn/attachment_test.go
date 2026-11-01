package shadcn

import (
	"strings"
	"testing"

	"github.com/ungerik/go-mx/html"
)

func TestAttachment(t *testing.T) {
	out := render(t, AttachmentGroup(
		Attachment("", "", "",
			AttachmentMedia("", "i"),
			AttachmentContent(
				AttachmentTitle("report.pdf"),
				AttachmentDescription("PDF · 2.4 MB"),
			),
			AttachmentActions(AttachmentAction("", "", html.Attrib("aria-label", "Remove"), "x")),
			AttachmentTrigger(html.Attrib("aria-label", "Open report.pdf")),
		),
	))
	for _, want := range []string{
		`data-slot="attachment-group"`, `data-slot="attachment"`, `data-state="done"`,
		`data-size="default"`, `data-orientation="horizontal"`,
		`data-slot="attachment-media"`, `data-variant="icon"`,
		`data-slot="attachment-content"`, `data-slot="attachment-title"`,
		`data-slot="attachment-description"`, `data-slot="attachment-actions"`,
		`data-slot="attachment-action"`, `<button data-slot="attachment-trigger"`, `type="button"`,
		"min-w-40", "[scrollbar-width:none]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// AttachmentAction is a ghost icon-xs Button re-slotted, like upstream.
	if !strings.Contains(out, `data-variant="ghost"`) || !strings.Contains(out, `data-size="icon-xs"`) {
		t.Errorf("action defaults wrong: %s", out)
	}
	// shimmer and scroll-fade-x are utilities of shadcn's own tailwind.css.
	for _, gone := range []string{"cn-", "shimmer", "scroll-fade", "scrollbar-none"} {
		if strings.Contains(out, gone) {
			t.Errorf("unexpected %q in %s", gone, out)
		}
	}
}

func TestAttachmentVariants(t *testing.T) {
	out := render(t, Attachment(AttachmentUploading, AttachmentSizeXS, AttachmentVertical,
		AttachmentMedia(AttachmentMediaImage, html.Img(html.Src("/a.png"), html.Alt(""))),
	))
	for _, want := range []string{
		`data-state="uploading"`, `data-size="xs"`, `data-orientation="vertical"`,
		"w-24", "flex-col", "rounded-lg", `data-variant="image"`, "opacity-60",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
	// The vertical tile's fixed width replaces the base w-fit.
	if strings.Contains(out, "w-fit") {
		t.Errorf("w-24 should replace w-fit: %s", out)
	}
}

func TestAttachmentTriggerLink(t *testing.T) {
	out := render(t, AttachmentTrigger(html.HRef("/files/report.pdf"), html.Download("")))
	if !strings.HasPrefix(out, "<a ") || strings.Contains(out, `type=`) {
		t.Errorf("href should render an <a> without a type: %s", out)
	}
}
