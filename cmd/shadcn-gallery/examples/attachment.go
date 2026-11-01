package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// AttachmentDemo renders attachments in the done, uploading and error
// states. The first one opens on click through its AttachmentTrigger.
func AttachmentDemo() mx.Component {
	return html.DivClass("flex w-full max-w-sm flex-col gap-3",
		shadcn.Attachment("", "", "",
			shadcn.AttachmentMedia("", iconFileText()),
			shadcn.AttachmentContent(
				shadcn.AttachmentTitle("quarterly-report.pdf"),
				shadcn.AttachmentDescription("PDF · 2.4 MB"),
			),
			shadcn.AttachmentTrigger(html.HRef("#attachment"), html.Attrib("aria-label", "Open quarterly-report.pdf")),
		),
		shadcn.Attachment(shadcn.AttachmentUploading, "", "",
			shadcn.AttachmentMedia("", shadcn.Spinner()),
			shadcn.AttachmentContent(
				shadcn.AttachmentTitle("screenshot.png"),
				shadcn.AttachmentDescription("Uploading… 64%"),
			),
			shadcn.AttachmentActions(
				shadcn.AttachmentAction("", "", html.Attrib("aria-label", "Cancel upload"), iconX()),
			),
		),
		shadcn.Attachment(shadcn.AttachmentError, "", "",
			shadcn.AttachmentMedia("", iconImage()),
			shadcn.AttachmentContent(
				shadcn.AttachmentTitle("diagram.svg"),
				shadcn.AttachmentDescription("Upload failed"),
			),
			shadcn.AttachmentActions(
				shadcn.AttachmentAction("", "", html.Attrib("aria-label", "Retry upload"), iconRetry()),
				shadcn.AttachmentAction("", "", html.Attrib("aria-label", "Remove"), iconX()),
			),
		),
	)
}

// AttachmentVertical renders attachments as tiles, with the actions over the
// media's corner.
func AttachmentVertical() mx.Component {
	return html.DivClass("flex flex-wrap gap-3",
		shadcn.Attachment("", "", shadcn.AttachmentVertical,
			shadcn.AttachmentMedia("", iconFileText()),
			shadcn.AttachmentContent(
				shadcn.AttachmentTitle("notes.md"),
				shadcn.AttachmentDescription("4 KB"),
			),
			shadcn.AttachmentActions(
				shadcn.AttachmentAction("", "", html.Attrib("aria-label", "Remove"), iconX()),
			),
		),
		shadcn.Attachment(shadcn.AttachmentIdle, "", shadcn.AttachmentVertical,
			shadcn.AttachmentMedia("", iconImage()),
			shadcn.AttachmentContent(
				shadcn.AttachmentTitle("photo.jpg"),
				shadcn.AttachmentDescription("Not sent"),
			),
		),
	)
}

// AttachmentGroupDemo renders small attachments in a row that scrolls
// sideways, as above a prompt input.
func AttachmentGroupDemo() mx.Component {
	file := func(name, details string) mx.Component {
		return shadcn.Attachment("", shadcn.AttachmentSizeSM, "",
			shadcn.AttachmentMedia("", iconFileText()),
			shadcn.AttachmentContent(
				shadcn.AttachmentTitle(name),
				shadcn.AttachmentDescription(details),
			),
			shadcn.AttachmentActions(
				shadcn.AttachmentAction("", "", html.Attrib("aria-label", "Remove "+name), iconX()),
			),
		)
	}
	return shadcn.AttachmentGroup(html.Class("w-full max-w-md"),
		file("main.go", "Go · 3 KB"),
		file("go.mod", "Text · 1 KB"),
		file("README.md", "Markdown · 9 KB"),
		file("CHANGELOG.md", "Markdown · 40 KB"),
	)
}
