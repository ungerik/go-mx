//go:generate go -C ../../tools tool go-enum ../shadcn/magicui/$GOFILE

package magicui

import (
	"context"
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
	"github.com/ungerik/go-mx/svg"
)

// HeroVideoAnimationStyle selects how the video of a [HeroVideoDialog]
// enters and leaves the screen.
type HeroVideoAnimationStyle string //#enum

const (
	// HeroVideoFromBottom slides the video in from below and out downwards.
	HeroVideoFromBottom HeroVideoAnimationStyle = "from-bottom"
	// HeroVideoFromCenter grows the video from half its size and shrinks it
	// back; the default.
	HeroVideoFromCenter HeroVideoAnimationStyle = "from-center"
	// HeroVideoFromTop slides the video in from above and out upwards.
	HeroVideoFromTop HeroVideoAnimationStyle = "from-top"
	// HeroVideoFromLeft slides the video in from the left and out to the left.
	HeroVideoFromLeft HeroVideoAnimationStyle = "from-left"
	// HeroVideoFromRight slides the video in from the right and out to the
	// right.
	HeroVideoFromRight HeroVideoAnimationStyle = "from-right"
	// HeroVideoFade only fades the video in and out.
	HeroVideoFade HeroVideoAnimationStyle = "fade"
	// HeroVideoTopInBottomOut slides the video in from above and out
	// downwards.
	HeroVideoTopInBottomOut HeroVideoAnimationStyle = "top-in-bottom-out"
	// HeroVideoLeftInRightOut slides the video in from the left and out to
	// the right.
	HeroVideoLeftInRightOut HeroVideoAnimationStyle = "left-in-right-out"
)

// Valid indicates if h is any of the valid values for HeroVideoAnimationStyle
func (h HeroVideoAnimationStyle) Valid() bool {
	switch h {
	case
		HeroVideoFromBottom,
		HeroVideoFromCenter,
		HeroVideoFromTop,
		HeroVideoFromLeft,
		HeroVideoFromRight,
		HeroVideoFade,
		HeroVideoTopInBottomOut,
		HeroVideoLeftInRightOut:
		return true
	}
	return false
}

// Validate returns an error if h is none of the valid values for HeroVideoAnimationStyle
func (h HeroVideoAnimationStyle) Validate() error {
	if !h.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.HeroVideoAnimationStyle", h)
	}
	return nil
}

// ValidOr returns h if it is any of the valid values for HeroVideoAnimationStyle, else defaultVal
func (h HeroVideoAnimationStyle) ValidOr(defaultVal HeroVideoAnimationStyle) HeroVideoAnimationStyle {
	if h.Valid() {
		return h
	}
	return defaultVal
}

// Enums returns all valid values for HeroVideoAnimationStyle
func (HeroVideoAnimationStyle) Enums() []HeroVideoAnimationStyle {
	return []HeroVideoAnimationStyle{
		HeroVideoFromBottom,
		HeroVideoFromCenter,
		HeroVideoFromTop,
		HeroVideoFromLeft,
		HeroVideoFromRight,
		HeroVideoFade,
		HeroVideoTopInBottomOut,
		HeroVideoLeftInRightOut,
	}
}

// EnumStrings returns all valid values for HeroVideoAnimationStyle as strings
func (HeroVideoAnimationStyle) EnumStrings() []string {
	return []string{
		"from-bottom",
		"from-center",
		"from-top",
		"from-left",
		"from-right",
		"fade",
		"top-in-bottom-out",
		"left-in-right-out",
	}
}

// String implements the fmt.Stringer interface for HeroVideoAnimationStyle
func (h HeroVideoAnimationStyle) String() string {
	return string(h)
}

// heroVideoAnimationClasses are upstream's motion variants as the
// closed/exit (plain), entrance (starting:open:) and open (open:) styles of
// the shadcn Dialog's CSS transitions. Every style fades through the
// dialog's own opacity classes; the scale-100 of the sliding styles cancels
// the dialog's zoom.
var heroVideoAnimationClasses = map[HeroVideoAnimationStyle]string{
	HeroVideoFromBottom:     "scale-100 starting:open:scale-100 translate-y-full starting:open:translate-y-full open:translate-y-0",
	HeroVideoFromCenter:     "scale-50 starting:open:scale-50",
	HeroVideoFromTop:        "scale-100 starting:open:scale-100 -translate-y-full starting:open:-translate-y-full open:translate-y-0",
	HeroVideoFromLeft:       "scale-100 starting:open:scale-100 -translate-x-full starting:open:-translate-x-full open:translate-x-0",
	HeroVideoFromRight:      "scale-100 starting:open:scale-100 translate-x-full starting:open:translate-x-full open:translate-x-0",
	HeroVideoFade:           "scale-100 starting:open:scale-100",
	HeroVideoTopInBottomOut: "scale-100 starting:open:scale-100 translate-y-full starting:open:-translate-y-full open:translate-y-0",
	HeroVideoLeftInRightOut: "scale-100 starting:open:scale-100 translate-x-full starting:open:-translate-x-full open:translate-x-0",
}

// heroVideoContentClasses turn the shadcn DialogContent box into upstream's
// bare video frame: no padding, border, background or shadow, the width of
// upstream's `mx-4 w-full max-w-4xl md:mx-0` (a margin would undo the
// dialog's m-auto centering), overflow visible for the close button above
// it, upstream's blurred overlay on the ::backdrop, a 300ms ease-out for
// upstream's spring, and the built-in close button (always the dialog's last
// child) hidden in favour of upstream's, which comes first.
const heroVideoContentClasses = "aspect-video w-full max-w-[calc(100%-2rem)] sm:max-w-[calc(100%-2rem)] md:max-w-4xl gap-0 overflow-visible rounded-none border-0 bg-transparent p-0 shadow-none " +
	"duration-300 ease-out backdrop:backdrop-blur-md [&>[data-slot=dialog-close]:last-child]:hidden"

// HeroVideoDialogOptions configures a [HeroVideoDialog].
type HeroVideoDialogOptions struct {
	// AnimationStyle selects the video's entrance and exit; empty or unknown
	// means [HeroVideoFromCenter].
	AnimationStyle HeroVideoAnimationStyle
	// VideoSrc is the URL of the embedded video player, e.g. a YouTube
	// /embed/ URL, loaded in an <iframe>.
	VideoSrc string
	// ThumbnailSrc is the image shown with the play button.
	ThumbnailSrc string
	// ThumbnailAlt is the thumbnail's alt text; empty means
	// "Video thumbnail".
	ThumbnailAlt string
	// ID is the id of the <dialog>. It must be unique in the document and
	// may only contain letters, digits, '_' and '-', any other character
	// makes the component render an error; empty derives one with
	// [mx.UniqueID].
	ID string
}

// HeroVideoDialog renders a video thumbnail with a play button that opens
// the video in a modal dialog over a blurred backdrop. It is built on
// [shadcn.DialogTrigger] and [shadcn.DialogContent], so the native <dialog>
// brings the focus trap, Escape to close and the light dismiss on a backdrop
// click, and the button above the video closes it too; it gets the
// screen-reader text of [shadcn.Labels].DialogClose, which upstream's
// unlabelled button lacks. The caller's attributes go to the root element,
// which has data-slot="hero-video-dialog", and children follow the dialog.
//
// Upstream mounts the <iframe> only while the dialog is open. Here it is
// always in the page, so it is loaded lazily (it stays unloaded while the
// dialog is closed) and the dialog reloads it on close, which stops the
// video, through an onclose handler attribute.
//
// Upstream's motion variants become CSS transitions between the dialog's
// closed and open styles, entrance and exit as upstream (exit transitions
// need a browser with the CSS overlay property, elsewhere the dialog closes
// at once, see [shadcn.Dialog]); its spring becomes a 300ms ease-out.
func HeroVideoDialog(opts HeroVideoDialogOptions, attribsChildren ...any) *mx.Element {
	id, err := svgRefID(opts.ID)
	if err != nil {
		return mx.NewErrElement(err)
	}
	animation := heroVideoAnimationClasses[opts.AnimationStyle.ValidOr(HeroVideoFromCenter)]

	trigger := shadcn.DialogTrigger(id,
		html.Class("group relative cursor-pointer border-0 bg-transparent p-0"),
		mx.NewAttrib("aria-label", "Play video"),
		html.Img(
			html.Src(opts.ThumbnailSrc),
			html.Alt(or(opts.ThumbnailAlt, "Video thumbnail")),
			html.Width(1920),
			html.Height(1080),
			html.Class("w-full rounded-md border shadow-lg transition-all duration-200 ease-out group-hover:brightness-[0.8]"),
		),
		html.DivClass("absolute inset-0 flex scale-[0.9] items-center justify-center rounded-2xl transition-all duration-200 ease-out group-hover:scale-100",
			html.DivClass("bg-primary/10 flex size-28 items-center justify-center rounded-full backdrop-blur-md",
				html.DivClass("from-primary/30 to-primary relative flex size-20 scale-100 items-center justify-center rounded-full bg-linear-to-b shadow-md transition-all duration-200 ease-out group-hover:scale-[1.2]",
					heroVideoIcon("size-8 scale-100 fill-white text-white transition-transform duration-200 ease-out group-hover:scale-105 [filter:drop-shadow(0_4px_3px_rgb(0_0_0/0.07))_drop-shadow(0_2px_2px_rgb(0_0_0/0.06))]",
						svg.Path(svg.D("M5 5a2 2 0 0 1 3.008-1.728l11.997 6.998a2 2 0 0 1 .003 3.458l-12 7A2 2 0 0 1 5 19z")),
					),
				),
			),
		),
	)
	if trigger.Err != nil {
		return trigger
	}

	content := shadcn.DialogContent(id,
		html.Class(heroVideoContentClasses+" "+animation),
		mx.NewAttrib("onclose", "var f=this.querySelector('iframe');if(f)f.src=f.src"),
		shadcn.DialogClose(
			html.Class("absolute -top-16 right-0 rounded-full bg-neutral-900/50 p-2 text-xl text-white ring-1 backdrop-blur-md dark:bg-neutral-100/50 dark:text-black"),
			heroVideoIcon("size-5", svg.Path(svg.D("M18 6 6 18")), svg.Path(svg.D("m6 6 12 12"))),
			html.SpanClass("sr-only", mx.ComponentFunc(func(ctx context.Context, w mx.Writer) error {
				return w.EscapeText(shadcn.LabelsFromContext(ctx).DialogClose)
			})),
		),
		html.DivClass("relative isolate z-1 size-full overflow-hidden rounded-2xl border-2 border-white",
			html.IFrame(
				html.Src(opts.VideoSrc),
				html.Title("Hero Video player"),
				html.Class("mt-0 size-full rounded-2xl"),
				html.BoolAttrib("allowfullscreen"),
				html.Allow("accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"),
				html.LoadingLazy,
			),
		),
	)
	if content.Err != nil {
		return content
	}

	root := html.Div(attribsChildren...)
	root.Children = append([]mx.Component{trigger, content}, root.Children...)
	return finish(root, "hero-video-dialog", "relative", "")
}

// heroVideoIcon renders a lucide icon (24×24, stroked with currentColor)
// from its shapes, with class merged in like lucide-react's className.
func heroVideoIcon(class string, attribsShapes ...any) *mx.Element {
	return svg.SVG(append([]any{
		svg.XMLNS,
		svg.Width(24),
		svg.Height(24),
		svg.ViewBox(0, 0, 24, 24),
		svg.Fill("none"),
		svg.Stroke("currentColor"),
		svg.StrokeWidth(2),
		svg.StrokeLineCapRound,
		svg.StrokeLineJoinRound,
		svg.Class(class),
		ariaHidden,
	}, attribsShapes...)...)
}
