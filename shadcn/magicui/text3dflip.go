//go:generate go -C ../../tools tool go-enum ../shadcn/magicui/$GOFILE

package magicui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

// Text3DFlipTag is the element a [Text3DFlip] renders as.
type Text3DFlipTag string //#enum

const (
	// Text3DFlipP renders a paragraph, upstream's default.
	Text3DFlipP Text3DFlipTag = "p"
	// Text3DFlipDiv renders a div.
	Text3DFlipDiv Text3DFlipTag = "div"
	// Text3DFlipSpan renders a span.
	Text3DFlipSpan Text3DFlipTag = "span"
	// Text3DFlipH1 renders an h1 heading.
	Text3DFlipH1 Text3DFlipTag = "h1"
	// Text3DFlipH2 renders an h2 heading.
	Text3DFlipH2 Text3DFlipTag = "h2"
	// Text3DFlipH3 renders an h3 heading.
	Text3DFlipH3 Text3DFlipTag = "h3"
	// Text3DFlipH4 renders an h4 heading.
	Text3DFlipH4 Text3DFlipTag = "h4"
	// Text3DFlipH5 renders an h5 heading.
	Text3DFlipH5 Text3DFlipTag = "h5"
	// Text3DFlipH6 renders an h6 heading.
	Text3DFlipH6 Text3DFlipTag = "h6"
	// Text3DFlipLi renders a list item.
	Text3DFlipLi Text3DFlipTag = "li"
)

// Valid indicates if t is any of the valid values for Text3DFlipTag
func (t Text3DFlipTag) Valid() bool {
	switch t {
	case
		Text3DFlipP,
		Text3DFlipDiv,
		Text3DFlipSpan,
		Text3DFlipH1,
		Text3DFlipH2,
		Text3DFlipH3,
		Text3DFlipH4,
		Text3DFlipH5,
		Text3DFlipH6,
		Text3DFlipLi:
		return true
	}
	return false
}

// Validate returns an error if t is none of the valid values for Text3DFlipTag
func (t Text3DFlipTag) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.Text3DFlipTag", t)
	}
	return nil
}

// ValidOr returns t if it is any of the valid values for Text3DFlipTag, else defaultVal
func (t Text3DFlipTag) ValidOr(defaultVal Text3DFlipTag) Text3DFlipTag {
	if t.Valid() {
		return t
	}
	return defaultVal
}

// Enums returns all valid values for Text3DFlipTag
func (Text3DFlipTag) Enums() []Text3DFlipTag {
	return []Text3DFlipTag{
		Text3DFlipP,
		Text3DFlipDiv,
		Text3DFlipSpan,
		Text3DFlipH1,
		Text3DFlipH2,
		Text3DFlipH3,
		Text3DFlipH4,
		Text3DFlipH5,
		Text3DFlipH6,
		Text3DFlipLi,
	}
}

// EnumStrings returns all valid values for Text3DFlipTag as strings
func (Text3DFlipTag) EnumStrings() []string {
	return []string{
		"p",
		"div",
		"span",
		"h1",
		"h2",
		"h3",
		"h4",
		"h5",
		"h6",
		"li",
	}
}

// String implements the fmt.Stringer interface for Text3DFlipTag
func (t Text3DFlipTag) String() string {
	return string(t)
}

// Text3DFlipRotateDirection is the edge a [Text3DFlip]'s characters roll
// towards.
type Text3DFlipRotateDirection string //#enum

const (
	// Text3DFlipTop rolls the characters towards the top edge.
	Text3DFlipTop Text3DFlipRotateDirection = "top"
	// Text3DFlipRight rolls the characters towards the right edge,
	// upstream's default.
	Text3DFlipRight Text3DFlipRotateDirection = "right"
	// Text3DFlipBottom rolls the characters towards the bottom edge.
	Text3DFlipBottom Text3DFlipRotateDirection = "bottom"
	// Text3DFlipLeft rolls the characters towards the left edge.
	Text3DFlipLeft Text3DFlipRotateDirection = "left"
)

// Valid indicates if t is any of the valid values for Text3DFlipRotateDirection
func (t Text3DFlipRotateDirection) Valid() bool {
	switch t {
	case
		Text3DFlipTop,
		Text3DFlipRight,
		Text3DFlipBottom,
		Text3DFlipLeft:
		return true
	}
	return false
}

// Validate returns an error if t is none of the valid values for Text3DFlipRotateDirection
func (t Text3DFlipRotateDirection) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.Text3DFlipRotateDirection", t)
	}
	return nil
}

// ValidOr returns t if it is any of the valid values for Text3DFlipRotateDirection, else defaultVal
func (t Text3DFlipRotateDirection) ValidOr(defaultVal Text3DFlipRotateDirection) Text3DFlipRotateDirection {
	if t.Valid() {
		return t
	}
	return defaultVal
}

// Enums returns all valid values for Text3DFlipRotateDirection
func (Text3DFlipRotateDirection) Enums() []Text3DFlipRotateDirection {
	return []Text3DFlipRotateDirection{
		Text3DFlipTop,
		Text3DFlipRight,
		Text3DFlipBottom,
		Text3DFlipLeft,
	}
}

// EnumStrings returns all valid values for Text3DFlipRotateDirection as strings
func (Text3DFlipRotateDirection) EnumStrings() []string {
	return []string{
		"top",
		"right",
		"bottom",
		"left",
	}
}

// String implements the fmt.Stringer interface for Text3DFlipRotateDirection
func (t Text3DFlipRotateDirection) String() string {
	return string(t)
}

// Text3DFlipStaggerFrom is the character a [Text3DFlip]'s flip starts
// from; the others follow with a delay growing with their distance to it.
type Text3DFlipStaggerFrom string //#enum

const (
	// Text3DFlipFromFirst starts at the first character, upstream's
	// default.
	Text3DFlipFromFirst Text3DFlipStaggerFrom = "first"
	// Text3DFlipFromLast starts at the last character.
	Text3DFlipFromLast Text3DFlipStaggerFrom = "last"
	// Text3DFlipFromCenter starts at the middle character.
	Text3DFlipFromCenter Text3DFlipStaggerFrom = "center"
	// Text3DFlipFromRandom gives every character the delay of its distance
	// to a character drawn at random (from Text3DFlipOptions.Seed), as
	// upstream does for each character anew.
	Text3DFlipFromRandom Text3DFlipStaggerFrom = "random"
	// Text3DFlipFromIndex starts at the character
	// Text3DFlipOptions.StaggerIndex, upstream's numeric staggerFrom.
	Text3DFlipFromIndex Text3DFlipStaggerFrom = "index"
)

// Valid indicates if t is any of the valid values for Text3DFlipStaggerFrom
func (t Text3DFlipStaggerFrom) Valid() bool {
	switch t {
	case
		Text3DFlipFromFirst,
		Text3DFlipFromLast,
		Text3DFlipFromCenter,
		Text3DFlipFromRandom,
		Text3DFlipFromIndex:
		return true
	}
	return false
}

// Validate returns an error if t is none of the valid values for Text3DFlipStaggerFrom
func (t Text3DFlipStaggerFrom) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.Text3DFlipStaggerFrom", t)
	}
	return nil
}

// ValidOr returns t if it is any of the valid values for Text3DFlipStaggerFrom, else defaultVal
func (t Text3DFlipStaggerFrom) ValidOr(defaultVal Text3DFlipStaggerFrom) Text3DFlipStaggerFrom {
	if t.Valid() {
		return t
	}
	return defaultVal
}

// Enums returns all valid values for Text3DFlipStaggerFrom
func (Text3DFlipStaggerFrom) Enums() []Text3DFlipStaggerFrom {
	return []Text3DFlipStaggerFrom{
		Text3DFlipFromFirst,
		Text3DFlipFromLast,
		Text3DFlipFromCenter,
		Text3DFlipFromRandom,
		Text3DFlipFromIndex,
	}
}

// EnumStrings returns all valid values for Text3DFlipStaggerFrom as strings
func (Text3DFlipStaggerFrom) EnumStrings() []string {
	return []string{
		"first",
		"last",
		"center",
		"random",
		"index",
	}
}

// String implements the fmt.Stringer interface for Text3DFlipStaggerFrom
func (t Text3DFlipStaggerFrom) String() string {
	return string(t)
}

// Text3DFlipOptions configures a [Text3DFlip]. The zero value is
// upstream's default: a paragraph whose characters roll to the right one
// after another, 50ms apart.
type Text3DFlipOptions struct {
	// As is the element to render; "" means [Text3DFlipP].
	As Text3DFlipTag
	// TextClass is added to the class of every character's front face.
	TextClass string
	// FlipTextClass is added to the class of every character's second
	// face, the one the flip turns to the front.
	FlipTextClass string
	// StaggerDuration is the delay between neighboring characters; 0
	// means 50ms.
	StaggerDuration time.Duration
	// StaggerFrom is where the flip starts; "" means [Text3DFlipFromFirst].
	StaggerFrom Text3DFlipStaggerFrom
	// StaggerIndex is the character the flip starts at for
	// [Text3DFlipFromIndex].
	StaggerIndex int
	// RotateDirection is the edge the characters roll towards; "" means
	// [Text3DFlipRight].
	RotateDirection Text3DFlipRotateDirection
	// Duration is how long one character's flip takes, eased like
	// upstream's default spring (damping 30, stiffness 300); 0 means
	// 460ms, that spring's settling time.
	Duration time.Duration
	// Seed draws the characters' delays for [Text3DFlipFromRandom].
	Seed uint64
}

// text3DFlipTransforms are upstream's transforms per rotate direction: of
// a character box at rest (container), when flipped (turn), and of its two
// faces.
var text3DFlipTransforms = map[Text3DFlipRotateDirection]struct{ container, turn, front, second string }{
	Text3DFlipTop: {
		"translateZ(-0.5lh)", "rotateX(90deg)",
		"translateZ(0.5lh)", "rotateX(-90deg) translateZ(0.5lh)",
	},
	Text3DFlipRight: {
		"rotateY(90deg) translateX(50%) rotateY(-90deg)", "rotateY(90deg)",
		"rotateY(-90deg) translateX(50%) rotateY(90deg)", "rotateY(90deg) translateX(50%) rotateY(-90deg) translateX(-50%) rotateY(-90deg) translateX(50%)",
	},
	Text3DFlipBottom: {
		"translateZ(-0.5lh)", "rotateX(-90deg)",
		"translateZ(0.5lh)", "rotateX(90deg) translateZ(0.5lh)",
	},
	Text3DFlipLeft: {
		"rotateY(90deg) translateX(50%) rotateY(-90deg)", "rotateY(-90deg)",
		"rotateY(90deg) translateX(50%) rotateY(-90deg)", "rotateY(90deg) translateX(50%) rotateY(-90deg) translateX(50%) rotateY(-90deg) translateX(50%)",
	},
}

// text3DFlipSpring is upstream's default spring, damping 30 and stiffness
// 300, sampled over its settling time as a linear() easing.
const text3DFlipSpring = "linear(0, 0.15, 0.413, 0.644, 0.807, 0.908, 0.963, 0.99, 1.001, 1.004, 1.004, 1.003, 1)"

// Text3DFlip rolls every character of its text over in 3D when the pointer
// enters it, one character after another, so each turns its second face
// to the front. The children must be strings (upstream extracts the text
// of its children); another child makes it render an error. Words wrap as
// a whole; an sr-only copy holds the text for assistive technology and the
// characters are aria-hidden (upstream exposes both).
//
// Upstream starts a motion animation on mouseenter that rolls the
// characters over and then snaps them back, while the pointer stays. Here
// it is a CSS transition on hover (group-hover, so only on devices that
// can hover): the characters stay flipped while the pointer is over the
// element and snap back without a transition when it leaves. With equal
// TextClass and FlipTextClass both look the same. Not ported: the
// transition prop (see Duration).
func Text3DFlip(opts Text3DFlipOptions, attribsChildren ...any) *mx.Element {
	root := mx.NewElement(string(opts.As.ValidOr(Text3DFlipP)), attribsChildren...)
	if root.Err != nil {
		return root
	}
	text, err := childText("Text3DFlip", root)
	if err != nil {
		return mx.NewErrElement(err)
	}
	direction := opts.RotateDirection.ValidOr(Text3DFlipRight)
	transforms := text3DFlipTransforms[direction]
	style, err := styleVars(
		"--text-3d-flip-rest", transforms.container,
		"--text-3d-flip-turn", transforms.turn,
		"--text-3d-flip-duration", seconds(opts.Duration, 460*time.Millisecond),
		"--text-3d-flip-easing", text3DFlipSpring,
	)
	if err != nil {
		return mx.NewErrElement(err)
	}

	words := strings.Split(text, " ")
	total := 0
	for _, word := range words {
		total += len([]rune(word))
	}
	delays := text3DFlipDelays(opts, total)
	charClass := "text-3d-flip-char inline transform-3d [transform:var(--text-3d-flip-rest)] group-hover/text-3d-flip:[transform:var(--text-3d-flip-turn)] group-hover/text-3d-flip:[transition:transform_var(--text-3d-flip-duration)_var(--text-3d-flip-easing)_var(--text-3d-flip-delay)]"
	frontClass := shadcn.Cn("relative h-[1lh] backface-hidden", opts.TextClass)
	secondClass := shadcn.Cn("absolute top-0 left-0 h-[1lh] backface-hidden", opts.FlipTextClass)

	root.Children = make([]mx.Component, 0, len(words)+1)
	root.Children = append(root.Children, html.Span(html.Class("sr-only"), text))
	index := 0
	for i, word := range words {
		chars := []any{html.Class("inline-flex"), ariaHidden}
		for _, r := range word {
			chars = append(chars, html.Span(
				html.Class(charClass),
				html.Style("--text-3d-flip-delay:"+seconds(delays[index], 0)),
				html.Span(html.Class(frontClass), html.Style("transform:"+transforms.front), string(r)),
				html.Span(html.Class(secondClass), html.Style("transform:"+transforms.second), string(r)),
			))
			index++
		}
		if i < len(words)-1 {
			chars = append(chars, html.Span(html.Class("whitespace-pre"), " "))
		}
		root.Children = append(root.Children, html.Span(chars...))
	}
	return finish(root, "text-3d-flip", "group/text-3d-flip relative flex flex-wrap", style)
}

// text3DFlipDelays returns the transition delay of each of total
// characters, as upstream's getStaggerDelay.
func text3DFlipDelays(opts Text3DFlipOptions, total int) []time.Duration {
	stagger := or(opts.StaggerDuration, 50*time.Millisecond)
	rng := newRand(opts.Seed)
	delays := make([]time.Duration, total)
	for i := range delays {
		var distance int
		switch opts.StaggerFrom.ValidOr(Text3DFlipFromFirst) {
		case Text3DFlipFromFirst:
			distance = i
		case Text3DFlipFromLast:
			distance = total - 1 - i
		case Text3DFlipFromCenter:
			distance = total/2 - i
		case Text3DFlipFromRandom:
			distance = rng.IntN(total) - i
		case Text3DFlipFromIndex:
			distance = opts.StaggerIndex - i
		}
		delays[i] = time.Duration(math.Abs(float64(distance))) * stagger
	}
	return delays
}
