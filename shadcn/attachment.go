//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// Attachment is a Go port of shadcn/ui's Attachment, a card for a file
// attached to a message or a prompt: media, name, details and actions, with
// a state for uploads in flight or failed. Class strings are reconstructed
// from the post-July-2026 registry (attachment.tsx skeleton + the
// style-vega.css .cn-attachment* rules) — see "Upstream restructure (July
// 2026)" in TODOS.md.

// AttachmentState is the upload state of an [Attachment].
type AttachmentState string //#enum

const (
	// AttachmentIdle is a file picked but not sent yet: a dashed border.
	AttachmentIdle AttachmentState = "idle"
	// AttachmentUploading pulses the title and fades an image.
	AttachmentUploading AttachmentState = "uploading"
	// AttachmentProcessing pulses the title and fades an image.
	AttachmentProcessing AttachmentState = "processing"
	// AttachmentError colors the border, media and description destructive.
	AttachmentError AttachmentState = "error"
	// AttachmentDone is a sent file, the default.
	AttachmentDone AttachmentState = "done"
)

// Valid indicates if a is any of the valid values for AttachmentState
func (a AttachmentState) Valid() bool {
	switch a {
	case
		AttachmentIdle,
		AttachmentUploading,
		AttachmentProcessing,
		AttachmentError,
		AttachmentDone:
		return true
	}
	return false
}

// Validate returns an error if a is none of the valid values for AttachmentState
func (a AttachmentState) Validate() error {
	if !a.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.AttachmentState", a)
	}
	return nil
}

// ValidOr returns a if it is any of the valid values for AttachmentState, else defaultVal
func (a AttachmentState) ValidOr(defaultVal AttachmentState) AttachmentState {
	if a.Valid() {
		return a
	}
	return defaultVal
}

// Enums returns all valid values for AttachmentState
func (AttachmentState) Enums() []AttachmentState {
	return []AttachmentState{
		AttachmentIdle,
		AttachmentUploading,
		AttachmentProcessing,
		AttachmentError,
		AttachmentDone,
	}
}

// EnumStrings returns all valid values for AttachmentState as strings
func (AttachmentState) EnumStrings() []string {
	return []string{
		"idle",
		"uploading",
		"processing",
		"error",
		"done",
	}
}

// String implements the fmt.Stringer interface for AttachmentState
func (a AttachmentState) String() string {
	return string(a)
}

// AttachmentSize selects an [Attachment]'s padding, media and text size.
type AttachmentSize string //#enum

const (
	// AttachmentSizeDefault is the default size.
	AttachmentSizeDefault AttachmentSize = "default"
	// AttachmentSizeSM is the small size.
	AttachmentSizeSM AttachmentSize = "sm"
	// AttachmentSizeXS is the extra-small size.
	AttachmentSizeXS AttachmentSize = "xs"
)

// Valid indicates if a is any of the valid values for AttachmentSize
func (a AttachmentSize) Valid() bool {
	switch a {
	case
		AttachmentSizeDefault,
		AttachmentSizeSM,
		AttachmentSizeXS:
		return true
	}
	return false
}

// Validate returns an error if a is none of the valid values for AttachmentSize
func (a AttachmentSize) Validate() error {
	if !a.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.AttachmentSize", a)
	}
	return nil
}

// ValidOr returns a if it is any of the valid values for AttachmentSize, else defaultVal
func (a AttachmentSize) ValidOr(defaultVal AttachmentSize) AttachmentSize {
	if a.Valid() {
		return a
	}
	return defaultVal
}

// Enums returns all valid values for AttachmentSize
func (AttachmentSize) Enums() []AttachmentSize {
	return []AttachmentSize{
		AttachmentSizeDefault,
		AttachmentSizeSM,
		AttachmentSizeXS,
	}
}

// EnumStrings returns all valid values for AttachmentSize as strings
func (AttachmentSize) EnumStrings() []string {
	return []string{
		"default",
		"sm",
		"xs",
	}
}

// String implements the fmt.Stringer interface for AttachmentSize
func (a AttachmentSize) String() string {
	return string(a)
}

// AttachmentOrientation selects an [Attachment]'s layout.
type AttachmentOrientation string //#enum

const (
	// AttachmentHorizontal puts the media beside the text, the default.
	AttachmentHorizontal AttachmentOrientation = "horizontal"
	// AttachmentVertical puts the media above the text in a narrow tile and
	// moves the actions over the media's corner.
	AttachmentVertical AttachmentOrientation = "vertical"
)

// Valid indicates if a is any of the valid values for AttachmentOrientation
func (a AttachmentOrientation) Valid() bool {
	switch a {
	case
		AttachmentHorizontal,
		AttachmentVertical:
		return true
	}
	return false
}

// Validate returns an error if a is none of the valid values for AttachmentOrientation
func (a AttachmentOrientation) Validate() error {
	if !a.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.AttachmentOrientation", a)
	}
	return nil
}

// ValidOr returns a if it is any of the valid values for AttachmentOrientation, else defaultVal
func (a AttachmentOrientation) ValidOr(defaultVal AttachmentOrientation) AttachmentOrientation {
	if a.Valid() {
		return a
	}
	return defaultVal
}

// Enums returns all valid values for AttachmentOrientation
func (AttachmentOrientation) Enums() []AttachmentOrientation {
	return []AttachmentOrientation{
		AttachmentHorizontal,
		AttachmentVertical,
	}
}

// EnumStrings returns all valid values for AttachmentOrientation as strings
func (AttachmentOrientation) EnumStrings() []string {
	return []string{
		"horizontal",
		"vertical",
	}
}

// String implements the fmt.Stringer interface for AttachmentOrientation
func (a AttachmentOrientation) String() string {
	return string(a)
}

// attachmentVariants resolves an attachment's base + size + orientation
// classes, declared the same way shadcn/ui's attachment.tsx declares them
// with cva.
var attachmentVariants = cva.New(cva.Config{
	Base: "group/attachment relative flex max-w-full min-w-0 shrink-0 flex-wrap border bg-card text-card-foreground transition-colors has-[>a,>button]:hover:bg-muted/50 data-[state=error]:border-destructive/30 data-[state=idle]:border-dashed rounded-xl w-fit focus-within:ring-1 focus-within:ring-ring/50",
	Variants: map[string]map[string]string{
		"size": {
			"default": "gap-2 has-data-[slot=attachment-content]:px-2.5 has-data-[slot=attachment-content]:py-2 has-data-[slot=attachment-media]:p-2 text-sm",
			"sm":      "gap-2.5 has-data-[slot=attachment-content]:px-2 has-data-[slot=attachment-content]:py-1.5 has-data-[slot=attachment-media]:p-1.5 text-xs",
			"xs":      "gap-1.5 has-data-[slot=attachment-content]:px-1.5 has-data-[slot=attachment-content]:py-1 has-data-[slot=attachment-media]:p-1 text-xs rounded-lg",
		},
		"orientation": {
			"horizontal": "items-center min-w-40",
			"vertical":   "flex-col w-24 has-data-[slot=attachment-content]:w-30",
		},
	},
	DefaultVariants: map[string]string{"size": "default", "orientation": "horizontal"},
})

// Attachment renders one attachment card. state, size and orientation may be
// "" for [AttachmentDone], [AttachmentSizeDefault] and
// [AttachmentHorizontal]; unknown values also resolve to those. Compose it
// with [AttachmentMedia], [AttachmentContent] (holding [AttachmentTitle] and
// [AttachmentDescription]), [AttachmentActions] and an [AttachmentTrigger]
// that makes the whole card clickable.
//
// Upstream's shimmer on the title of an uploading or processing file comes
// from shadcn's own tailwind.css, not from Tailwind; this package emits
// Tailwind classes only, so the title pulses with animate-pulse instead (not
// under prefers-reduced-motion).
func Attachment(state AttachmentState, size AttachmentSize, orientation AttachmentOrientation, attribsChildren ...any) *mx.Element {
	st := string(state.ValidOr(AttachmentDone))
	sz := string(size.ValidOr(AttachmentSizeDefault))
	o := string(orientation.ValidOr(AttachmentHorizontal))
	e := html.Div(attribsChildren...)
	e.Attribs = append(e.Attribs,
		html.DataAttr("state", st),
		html.DataAttr("size", sz),
		html.DataAttr("orientation", o),
	)
	return finish(e, "attachment",
		Cn(attachmentVariants(map[string]string{"size": sz, "orientation": o})))
}

// AttachmentMediaVariant selects how an [AttachmentMedia] frames its
// content.
type AttachmentMediaVariant string //#enum

const (
	// AttachmentMediaIcon frames a file-type icon or a [Spinner] in a muted
	// square, the default.
	AttachmentMediaIcon AttachmentMediaVariant = "icon"
	// AttachmentMediaImage crops an <img> thumbnail to the square, faded
	// while the file uploads.
	AttachmentMediaImage AttachmentMediaVariant = "image"
)

// Valid indicates if a is any of the valid values for AttachmentMediaVariant
func (a AttachmentMediaVariant) Valid() bool {
	switch a {
	case
		AttachmentMediaIcon,
		AttachmentMediaImage:
		return true
	}
	return false
}

// Validate returns an error if a is none of the valid values for AttachmentMediaVariant
func (a AttachmentMediaVariant) Validate() error {
	if !a.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.AttachmentMediaVariant", a)
	}
	return nil
}

// ValidOr returns a if it is any of the valid values for AttachmentMediaVariant, else defaultVal
func (a AttachmentMediaVariant) ValidOr(defaultVal AttachmentMediaVariant) AttachmentMediaVariant {
	if a.Valid() {
		return a
	}
	return defaultVal
}

// Enums returns all valid values for AttachmentMediaVariant
func (AttachmentMediaVariant) Enums() []AttachmentMediaVariant {
	return []AttachmentMediaVariant{
		AttachmentMediaIcon,
		AttachmentMediaImage,
	}
}

// EnumStrings returns all valid values for AttachmentMediaVariant as strings
func (AttachmentMediaVariant) EnumStrings() []string {
	return []string{
		"icon",
		"image",
	}
}

// String implements the fmt.Stringer interface for AttachmentMediaVariant
func (a AttachmentMediaVariant) String() string {
	return string(a)
}

// attachmentMediaVariants resolves the media's base + variant classes,
// declared the same way shadcn/ui's attachment.tsx declares them with cva.
var attachmentMediaVariants = cva.New(cva.Config{
	Base: "relative flex aspect-square shrink-0 items-center justify-center overflow-hidden group-data-[state=error]/attachment:bg-destructive/10 group-data-[state=error]/attachment:text-destructive [&_svg]:pointer-events-none bg-muted text-foreground w-10 rounded-lg group-data-[size=sm]/attachment:w-8 group-data-[size=xs]/attachment:w-7 group-data-[size=xs]/attachment:rounded-md [&_svg:not([class*='size-'])]:size-4 group-data-[size=xs]/attachment:[&_svg:not([class*='size-'])]:size-3.5 group-data-[orientation=vertical]/attachment:w-full group-data-[orientation=vertical]/attachment:[&_svg:not([class*='size-'])]:size-6 group-data-[orientation=vertical]/attachment:*:data-[slot=spinner]:size-6!",
	Variants: map[string]map[string]string{
		"variant": {
			"icon":  "",
			"image": "*:[img]:aspect-square *:[img]:w-full *:[img]:object-cover opacity-60 group-data-[state=idle]/attachment:opacity-100 group-data-[state=done]/attachment:opacity-100",
		},
	},
	DefaultVariants: map[string]string{"variant": "icon"},
})

// AttachmentMedia renders an [Attachment]'s icon, spinner or thumbnail.
// variant may be "" for [AttachmentMediaIcon]; an unknown value also
// resolves to it.
func AttachmentMedia(variant AttachmentMediaVariant, attribsChildren ...any) *mx.Element {
	v := string(variant.ValidOr(AttachmentMediaIcon))
	e := html.Div(attribsChildren...)
	e.Attribs = append(e.Attribs, html.DataAttr("variant", v))
	return finish(e, "attachment-media",
		Cn(attachmentMediaVariants(map[string]string{"variant": v})))
}

// AttachmentContent holds an [Attachment]'s [AttachmentTitle] and
// [AttachmentDescription].
func AttachmentContent(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "attachment-content",
		"max-w-full min-w-0 flex-1 leading-tight group-data-[orientation=vertical]/attachment:px-1")
}

// AttachmentTitle renders an [Attachment]'s file name, truncated to one line.
func AttachmentTitle(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "attachment-title",
		"block max-w-full min-w-0 truncate motion-safe:group-data-[state=processing]/attachment:animate-pulse motion-safe:group-data-[state=uploading]/attachment:animate-pulse font-medium")
}

// AttachmentDescription renders an [Attachment]'s details, such as the file
// type and size or an error, truncated to one line.
func AttachmentDescription(attribsChildren ...any) *mx.Element {
	return finish(html.Span(attribsChildren...), "attachment-description",
		"block min-w-0 truncate text-muted-foreground group-data-[state=error]/attachment:text-destructive/80 max-w-full mt-0.5 text-xs")
}

// AttachmentActions holds an [Attachment]'s [AttachmentAction]s. They stay
// clickable above an [AttachmentTrigger].
func AttachmentActions(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "attachment-actions",
		"flex shrink-0 items-center group-data-[orientation=vertical]/attachment:absolute group-data-[orientation=vertical]/attachment:top-3 group-data-[orientation=vertical]/attachment:right-3 relative z-20 group-data-[orientation=vertical]/attachment:gap-1")
}

// AttachmentAction renders an icon [Button] in [AttachmentActions], such as
// remove or retry. variant and size may be "" for [ButtonGhost] and
// [SizeIconXS], upstream's defaults here. An icon-only action needs an
// aria-label.
func AttachmentAction(variant ButtonVariant, size ButtonSize, attribsChildren ...any) *mx.Element {
	return finish(Button(variant.ValidOr(ButtonGhost), size.ValidOr(SizeIconXS), attribsChildren...),
		"attachment-action", "")
}

// AttachmentTrigger renders a transparent overlay that makes the whole
// [Attachment] clickable, for opening a preview or downloading the file.
// Give it an accessible name with html.Attrib("aria-label", …), since it has
// no text.
//
// Upstream's render prop (React useRender) is replaced by the attributes: the
// trigger is a <button type="button">, and an <a> when given an href (a
// download link).
func AttachmentTrigger(attribsChildren ...any) *mx.Element {
	e := html.Button(attribsChildren...)
	e.Name = renderTag(e, "button")
	if e.Name == "button" && e.AttribIndex("type") < 0 {
		e.Attribs = append(e.Attribs, html.Type("button"))
	}
	return finish(e, "attachment-trigger", "absolute inset-0 z-10 outline-none")
}

// AttachmentGroup lays [Attachment]s out in a row that scrolls sideways and
// snaps to each card. Upstream's scroll-fade-x edge mask comes from shadcn's
// own tailwind.css and is not reproduced.
func AttachmentGroup(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "attachment-group",
		"flex min-w-0 snap-x snap-mandatory [scrollbar-width:none] [&::-webkit-scrollbar]:hidden overflow-x-auto overscroll-x-contain *:data-[slot=attachment]:flex-none *:data-[slot=attachment]:snap-start gap-3 scroll-px-1 py-1")
}
