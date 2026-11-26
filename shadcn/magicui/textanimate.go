//go:generate go -C ../../tools tool go-enum ../shadcn/magicui/$GOFILE

package magicui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// TextAnimateBy is how a [TextAnimate] splits its text into the segments
// that animate one after another.
type TextAnimateBy string //#enum

const (
	// TextAnimateByText animates the whole text as one segment.
	TextAnimateByText TextAnimateBy = "text"
	// TextAnimateByWord animates word by word, upstream's default. The
	// whitespace between the words is a segment of its own, as upstream.
	TextAnimateByWord TextAnimateBy = "word"
	// TextAnimateByCharacter animates character (rune) by character.
	TextAnimateByCharacter TextAnimateBy = "character"
	// TextAnimateByLine animates line by line, splitting at "\n".
	TextAnimateByLine TextAnimateBy = "line"
)

// Valid indicates if t is any of the valid values for TextAnimateBy
func (t TextAnimateBy) Valid() bool {
	switch t {
	case
		TextAnimateByText,
		TextAnimateByWord,
		TextAnimateByCharacter,
		TextAnimateByLine:
		return true
	}
	return false
}

// Validate returns an error if t is none of the valid values for TextAnimateBy
func (t TextAnimateBy) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.TextAnimateBy", t)
	}
	return nil
}

// ValidOr returns t if it is any of the valid values for TextAnimateBy, else defaultVal
func (t TextAnimateBy) ValidOr(defaultVal TextAnimateBy) TextAnimateBy {
	if t.Valid() {
		return t
	}
	return defaultVal
}

// Enums returns all valid values for TextAnimateBy
func (TextAnimateBy) Enums() []TextAnimateBy {
	return []TextAnimateBy{
		TextAnimateByText,
		TextAnimateByWord,
		TextAnimateByCharacter,
		TextAnimateByLine,
	}
}

// EnumStrings returns all valid values for TextAnimateBy as strings
func (TextAnimateBy) EnumStrings() []string {
	return []string{
		"text",
		"word",
		"character",
		"line",
	}
}

// String implements the fmt.Stringer interface for TextAnimateBy
func (t TextAnimateBy) String() string {
	return string(t)
}

// TextAnimateAnimation is the entrance preset of a [TextAnimate]'s
// segments, named like upstream's animation prop.
type TextAnimateAnimation string //#enum

const (
	// TextAnimateFadeIn fades the segments in while they rise 20px,
	// upstream's default.
	TextAnimateFadeIn TextAnimateAnimation = "fadeIn"
	// TextAnimateBlurIn fades the segments in from a 10px blur.
	TextAnimateBlurIn TextAnimateAnimation = "blurIn"
	// TextAnimateBlurInUp fades the segments in from a 10px blur while
	// they rise 20px.
	TextAnimateBlurInUp TextAnimateAnimation = "blurInUp"
	// TextAnimateBlurInDown fades the segments in from a 10px blur while
	// they drop 20px.
	TextAnimateBlurInDown TextAnimateAnimation = "blurInDown"
	// TextAnimateSlideUp slides the segments up 20px while they fade in.
	TextAnimateSlideUp TextAnimateAnimation = "slideUp"
	// TextAnimateSlideDown slides the segments down 20px while they fade in.
	TextAnimateSlideDown TextAnimateAnimation = "slideDown"
	// TextAnimateSlideLeft slides the segments 20px to the left while they
	// fade in.
	TextAnimateSlideLeft TextAnimateAnimation = "slideLeft"
	// TextAnimateSlideRight slides the segments 20px to the right while
	// they fade in.
	TextAnimateSlideRight TextAnimateAnimation = "slideRight"
	// TextAnimateScaleUp springs the segments up from half their size while
	// they fade in.
	TextAnimateScaleUp TextAnimateAnimation = "scaleUp"
	// TextAnimateScaleDown springs the segments down from one and a half
	// times their size while they fade in.
	TextAnimateScaleDown TextAnimateAnimation = "scaleDown"
)

// Valid indicates if t is any of the valid values for TextAnimateAnimation
func (t TextAnimateAnimation) Valid() bool {
	switch t {
	case
		TextAnimateFadeIn,
		TextAnimateBlurIn,
		TextAnimateBlurInUp,
		TextAnimateBlurInDown,
		TextAnimateSlideUp,
		TextAnimateSlideDown,
		TextAnimateSlideLeft,
		TextAnimateSlideRight,
		TextAnimateScaleUp,
		TextAnimateScaleDown:
		return true
	}
	return false
}

// Validate returns an error if t is none of the valid values for TextAnimateAnimation
func (t TextAnimateAnimation) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.TextAnimateAnimation", t)
	}
	return nil
}

// ValidOr returns t if it is any of the valid values for TextAnimateAnimation, else defaultVal
func (t TextAnimateAnimation) ValidOr(defaultVal TextAnimateAnimation) TextAnimateAnimation {
	if t.Valid() {
		return t
	}
	return defaultVal
}

// Enums returns all valid values for TextAnimateAnimation
func (TextAnimateAnimation) Enums() []TextAnimateAnimation {
	return []TextAnimateAnimation{
		TextAnimateFadeIn,
		TextAnimateBlurIn,
		TextAnimateBlurInUp,
		TextAnimateBlurInDown,
		TextAnimateSlideUp,
		TextAnimateSlideDown,
		TextAnimateSlideLeft,
		TextAnimateSlideRight,
		TextAnimateScaleUp,
		TextAnimateScaleDown,
	}
}

// EnumStrings returns all valid values for TextAnimateAnimation as strings
func (TextAnimateAnimation) EnumStrings() []string {
	return []string{
		"fadeIn",
		"blurIn",
		"blurInUp",
		"blurInDown",
		"slideUp",
		"slideDown",
		"slideLeft",
		"slideRight",
		"scaleUp",
		"scaleDown",
	}
}

// String implements the fmt.Stringer interface for TextAnimateAnimation
func (t TextAnimateAnimation) String() string {
	return string(t)
}

// TextAnimateTag is the element a [TextAnimate] renders as, one of the
// elements upstream's as prop allows.
type TextAnimateTag string //#enum

const (
	// TextAnimateP renders a paragraph, upstream's default.
	TextAnimateP TextAnimateTag = "p"
	// TextAnimateArticle renders an article.
	TextAnimateArticle TextAnimateTag = "article"
	// TextAnimateDiv renders a div.
	TextAnimateDiv TextAnimateTag = "div"
	// TextAnimateH1 renders an h1 heading.
	TextAnimateH1 TextAnimateTag = "h1"
	// TextAnimateH2 renders an h2 heading.
	TextAnimateH2 TextAnimateTag = "h2"
	// TextAnimateH3 renders an h3 heading.
	TextAnimateH3 TextAnimateTag = "h3"
	// TextAnimateH4 renders an h4 heading.
	TextAnimateH4 TextAnimateTag = "h4"
	// TextAnimateH5 renders an h5 heading.
	TextAnimateH5 TextAnimateTag = "h5"
	// TextAnimateH6 renders an h6 heading.
	TextAnimateH6 TextAnimateTag = "h6"
	// TextAnimateLi renders a list item.
	TextAnimateLi TextAnimateTag = "li"
	// TextAnimateSection renders a section.
	TextAnimateSection TextAnimateTag = "section"
	// TextAnimateSpan renders a span.
	TextAnimateSpan TextAnimateTag = "span"
)

// Valid indicates if t is any of the valid values for TextAnimateTag
func (t TextAnimateTag) Valid() bool {
	switch t {
	case
		TextAnimateP,
		TextAnimateArticle,
		TextAnimateDiv,
		TextAnimateH1,
		TextAnimateH2,
		TextAnimateH3,
		TextAnimateH4,
		TextAnimateH5,
		TextAnimateH6,
		TextAnimateLi,
		TextAnimateSection,
		TextAnimateSpan:
		return true
	}
	return false
}

// Validate returns an error if t is none of the valid values for TextAnimateTag
func (t TextAnimateTag) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.TextAnimateTag", t)
	}
	return nil
}

// ValidOr returns t if it is any of the valid values for TextAnimateTag, else defaultVal
func (t TextAnimateTag) ValidOr(defaultVal TextAnimateTag) TextAnimateTag {
	if t.Valid() {
		return t
	}
	return defaultVal
}

// Enums returns all valid values for TextAnimateTag
func (TextAnimateTag) Enums() []TextAnimateTag {
	return []TextAnimateTag{
		TextAnimateP,
		TextAnimateArticle,
		TextAnimateDiv,
		TextAnimateH1,
		TextAnimateH2,
		TextAnimateH3,
		TextAnimateH4,
		TextAnimateH5,
		TextAnimateH6,
		TextAnimateLi,
		TextAnimateSection,
		TextAnimateSpan,
	}
}

// EnumStrings returns all valid values for TextAnimateTag as strings
func (TextAnimateTag) EnumStrings() []string {
	return []string{
		"p",
		"article",
		"div",
		"h1",
		"h2",
		"h3",
		"h4",
		"h5",
		"h6",
		"li",
		"section",
		"span",
	}
}

// String implements the fmt.Stringer interface for TextAnimateTag
func (t TextAnimateTag) String() string {
	return string(t)
}

// TextAnimateOptions configures a [TextAnimate]. The zero value is
// upstream's default apart from StartOnView: a paragraph fading in word by
// word as soon as the page shows it.
type TextAnimateOptions struct {
	// Animation is the entrance preset; "" means [TextAnimateFadeIn].
	Animation TextAnimateAnimation
	// By is how the text is split into segments; "" means
	// [TextAnimateByWord].
	By TextAnimateBy
	// As is the element to render; "" means [TextAnimateP].
	As TextAnimateTag
	// Delay is the time before the first segment starts.
	Delay time.Duration
	// Duration is the time over which the segments' starts are spread:
	// segment i starts Delay + i*Duration/segments after the page shows
	// it, as upstream's staggerChildren. Every segment itself animates for
	// the preset's fixed duration (0.3s; the scale presets' spring settles
	// in 0.85s). 0 means 300ms.
	Duration time.Duration
	// SegmentClass is added to the class of every segment, as upstream's
	// segmentClassName. A caller's own animate-* utility there replaces
	// the preset; it can read the segment's start time from the CSS
	// variables --text-animate-delay, --text-animate-stagger and --i.
	SegmentClass string
	// StartOnView ties the animation to scrolling instead of time: the
	// segments animate while the element scrolls into view and run
	// backwards when it scrolls out again, with scroll-driven animations
	// (animation-timeline: view()). Browsers without them show the text
	// at once. Delay and Duration do not apply; the segments still start
	// one after another.
	StartOnView bool
	// NoScreenReaderText renders the segments without the sr-only copy of
	// the whole text and without aria-hidden on the segments, as upstream's
	// accessible={false}.
	NoScreenReaderText bool
}

// TextAnimate animates its text in segment by segment with one of
// upstream's entrance presets. The children must be strings, as upstream's
// children: string; another child makes it render an error. The segments
// are aria-hidden spans and an sr-only span holds the whole text, so
// assistive technology reads it once.
//
// Upstream animates with the motion library and plays the entrance when
// the text scrolls into view (startOnView, default true), which needs
// JavaScript to observe the viewport; here the entrance is CSS that plays
// when the page shows the text, and StartOnView instead links it to the
// scroll position. Not ported: exit animations (there is no unmount), the
// once prop (a scroll-driven animation cannot stay finished), custom motion
// variants (see SegmentClass), and the aria-label upstream puts on the
// element besides the sr-only copy.
//
// The element is a flex container (a column by line), so that whitespace
// between the segment elements, such as the line breaks of an indenting
// writer, is not rendered; align its text with justify-* classes rather
// than text-*.
//
// Needs [CSS] in the app's Tailwind input.
func TextAnimate(opts TextAnimateOptions, attribsChildren ...any) *mx.Element {
	root := mx.NewElement(string(opts.As.ValidOr(TextAnimateP)), attribsChildren...)
	if root.Err != nil {
		return root
	}
	text, err := childText("TextAnimate", root)
	if err != nil {
		return mx.NewErrElement(err)
	}
	by := opts.By.ValidOr(TextAnimateByWord)
	segments := textAnimateSegments(text, by)
	stagger := or(opts.Duration, 300*time.Millisecond) / time.Duration(max(len(segments), 1))

	animation := opts.Animation.ValidOr(TextAnimateFadeIn)
	animateClass := textAnimateClasses[animation]
	// A flex container, unlike upstream's inline flow, ignores whitespace
	// between the segment elements, so markup an indenting writer breaks
	// into lines renders like compact markup.
	rootClass := "flex flex-wrap whitespace-pre-wrap"
	if by == TextAnimateByLine {
		rootClass = "flex flex-col whitespace-pre-wrap"
	}
	pairs := []string{
		"--text-animate-delay", seconds(opts.Delay, 0),
		"--text-animate-stagger", seconds(stagger.Round(time.Microsecond), 0),
	}
	if opts.StartOnView {
		animateClass = "animate-text-animate-on-view"
		rootClass += " [view-timeline-name:--text-animate]"
		pairs = append(pairs,
			"--n", strconv.Itoa(len(segments)),
			"--text-animate-keyframes", textAnimateViewKeyframes[animation],
		)
	}
	style, err := styleVars(pairs...)
	if err != nil {
		return mx.NewErrElement(err)
	}

	segmentClass := shadcn.Cn(
		map[string]bool{"block": by == TextAnimateByLine, "inline-block whitespace-pre": by != TextAnimateByLine},
		animateClass,
		opts.SegmentClass,
	)
	root.Children = make([]mx.Component, 0, len(segments)+1)
	if !opts.NoScreenReaderText {
		root.Children = append(root.Children, html.Span(html.Class("sr-only"), text))
	}
	// By character, the letters of a word share an inline-flex span, so
	// the flex container wraps between words, not inside one.
	var word *mx.Element
	for i, segment := range segments {
		span := html.Span(html.Class(segmentClass), html.Style("--i:"+strconv.Itoa(i)), segment)
		if !opts.NoScreenReaderText {
			span.Attribs = append(span.Attribs, ariaHidden)
		}
		if by != TextAnimateByCharacter || strings.TrimSpace(segment) == "" {
			word = nil
			root.Children = append(root.Children, span)
			continue
		}
		if word == nil {
			word = html.Span(html.Class("inline-flex"))
			root.Children = append(root.Children, word)
		}
		word.Children = append(word.Children, span)
	}
	return finish(root, "text-animate", rootClass, style)
}

// textAnimateViewKeyframes are the names of the @keyframes a StartOnView
// [TextAnimate] animates its segments with, an animation-name list.
var textAnimateViewKeyframes = map[TextAnimateAnimation]string{
	TextAnimateFadeIn:     "text-animate-up",
	TextAnimateBlurIn:     "text-animate-blur",
	TextAnimateBlurInUp:   "text-animate-blur-up, text-animate-fade",
	TextAnimateBlurInDown: "text-animate-blur-down, text-animate-fade",
	TextAnimateSlideUp:    "text-animate-up",
	TextAnimateSlideDown:  "text-animate-down",
	TextAnimateSlideLeft:  "text-animate-left",
	TextAnimateSlideRight: "text-animate-right",
	TextAnimateScaleUp:    "text-animate-scale-up, text-animate-fade",
	TextAnimateScaleDown:  "text-animate-scale-down, text-animate-fade",
}

// textAnimateClasses are the animate-* utilities of the presets.
var textAnimateClasses = map[TextAnimateAnimation]string{
	TextAnimateFadeIn:     "animate-text-animate-fade-in",
	TextAnimateBlurIn:     "animate-text-animate-blur-in",
	TextAnimateBlurInUp:   "animate-text-animate-blur-in-up",
	TextAnimateBlurInDown: "animate-text-animate-blur-in-down",
	TextAnimateSlideUp:    "animate-text-animate-slide-up",
	TextAnimateSlideDown:  "animate-text-animate-slide-down",
	TextAnimateSlideLeft:  "animate-text-animate-slide-left",
	TextAnimateSlideRight: "animate-text-animate-slide-right",
	TextAnimateScaleUp:    "animate-text-animate-scale-up",
	TextAnimateScaleDown:  "animate-text-animate-scale-down",
}

// textAnimateWordRE matches the whitespace runs upstream splits words at,
// keeping them as segments of their own.
var textAnimateWordRE = regexp.MustCompile(`\s+`)

// textAnimateSegments splits text like upstream's TextAnimate, leaving
// out the empty segments JavaScript's split produces at the edges.
func textAnimateSegments(text string, by TextAnimateBy) []string {
	var segments []string
	switch by {
	case TextAnimateByWord:
		last := 0
		for _, loc := range textAnimateWordRE.FindAllStringIndex(text, -1) {
			segments = append(segments, text[last:loc[0]], text[loc[0]:loc[1]])
			last = loc[1]
		}
		segments = append(segments, text[last:])
	case TextAnimateByCharacter:
		for _, r := range text {
			segments = append(segments, string(r))
		}
		return segments
	case TextAnimateByLine:
		segments = strings.Split(text, "\n")
	default:
		segments = []string{text}
	}
	nonEmpty := segments[:0]
	for _, s := range segments {
		if s != "" || by == TextAnimateByLine {
			nonEmpty = append(nonEmpty, s)
		}
	}
	return nonEmpty
}
