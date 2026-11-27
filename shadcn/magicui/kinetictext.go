//go:generate go -C ../../tools tool go-enum ../shadcn/magicui/$GOFILE

package magicui

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// KineticTextTag is the element a [KineticText] renders as.
type KineticTextTag string //#enum

const (
	// KineticTextH1 renders an h1 heading, upstream's default.
	KineticTextH1 KineticTextTag = "h1"
	// KineticTextH2 renders an h2 heading.
	KineticTextH2 KineticTextTag = "h2"
	// KineticTextH3 renders an h3 heading.
	KineticTextH3 KineticTextTag = "h3"
	// KineticTextH4 renders an h4 heading.
	KineticTextH4 KineticTextTag = "h4"
	// KineticTextH5 renders an h5 heading.
	KineticTextH5 KineticTextTag = "h5"
	// KineticTextH6 renders an h6 heading.
	KineticTextH6 KineticTextTag = "h6"
	// KineticTextP renders a paragraph.
	KineticTextP KineticTextTag = "p"
	// KineticTextSpan renders a span.
	KineticTextSpan KineticTextTag = "span"
)

// Valid indicates if k is any of the valid values for KineticTextTag
func (k KineticTextTag) Valid() bool {
	switch k {
	case
		KineticTextH1,
		KineticTextH2,
		KineticTextH3,
		KineticTextH4,
		KineticTextH5,
		KineticTextH6,
		KineticTextP,
		KineticTextSpan:
		return true
	}
	return false
}

// Validate returns an error if k is none of the valid values for KineticTextTag
func (k KineticTextTag) Validate() error {
	if !k.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.KineticTextTag", k)
	}
	return nil
}

// ValidOr returns k if it is any of the valid values for KineticTextTag, else defaultVal
func (k KineticTextTag) ValidOr(defaultVal KineticTextTag) KineticTextTag {
	if k.Valid() {
		return k
	}
	return defaultVal
}

// Enums returns all valid values for KineticTextTag
func (KineticTextTag) Enums() []KineticTextTag {
	return []KineticTextTag{
		KineticTextH1,
		KineticTextH2,
		KineticTextH3,
		KineticTextH4,
		KineticTextH5,
		KineticTextH6,
		KineticTextP,
		KineticTextSpan,
	}
}

// EnumStrings returns all valid values for KineticTextTag as strings
func (KineticTextTag) EnumStrings() []string {
	return []string{
		"h1",
		"h2",
		"h3",
		"h4",
		"h5",
		"h6",
		"p",
		"span",
	}
}

// String implements the fmt.Stringer interface for KineticTextTag
func (k KineticTextTag) String() string {
	return string(k)
}

// KineticTextOptions configures a [KineticText].
type KineticTextOptions struct {
	// Text is the text whose letters react to the pointer.
	Text string
	// As is the element to render; "" means [KineticTextH1].
	As KineticTextTag
}

// kineticTextLetterClasses are the classes of one letter of a [KineticText].
const kineticTextLetterClasses = "[will-change:font-weight,-webkit-text-stroke-width,padding] [-webkit-text-stroke-color:transparent] [-webkit-text-stroke-width:var(--text-stroke-width)] [transition:font-weight_0.4s,_-webkit-text-stroke-color_0.4s,_padding_0.4s] hover:[padding-inline:var(--hover-padding)] hover:font-[900] hover:[-webkit-text-stroke-color:currentcolor] hover:[-webkit-text-stroke-width:calc(var(--text-stroke-width)*2)] has-[+span+span:hover]:font-[400] has-[+span:hover]:[padding-inline:var(--hover-padding)] has-[+span:hover]:font-[600] [:hover+&]:[padding-inline:var(--hover-padding)] [:hover+&]:font-[600] [:hover+span+&]:font-[400]"

// KineticText renders opts.Text one letter per span, so the letter under
// the pointer swells to font weight 900 with a text stroke and its
// neighbours to 600 and 400, all with CSS transitions. The effect needs a
// variable font for the weights in between. The letters are aria-hidden
// and an sr-only copy of the text is read instead. Children in
// attribsChildren follow the letters (upstream has none).
//
// The CSS variables --hover-padding and --text-stroke-width can be
// overridden with a caller style, as upstream allows. The text is split
// into Unicode code points, where upstream splits UTF-16 code units. It
// needs no [CSS] of its own.
func KineticText(opts KineticTextOptions, attribsChildren ...any) *mx.Element {
	root := mx.NewElement(string(opts.As.ValidOr(KineticTextH1)), attribsChildren...)
	if root.Err != nil {
		return root
	}
	children := root.Children
	root.Children = make([]mx.Component, 0, len(opts.Text)+1+len(children))
	for _, letter := range opts.Text {
		if letter == ' ' {
			letter = ' '
		}
		root.Children = append(root.Children, html.Span(
			html.Class(kineticTextLetterClasses),
			ariaHidden,
			string(letter),
		))
	}
	root.Children = append(root.Children, html.SpanClass("sr-only", opts.Text))
	root.Children = append(root.Children, children...)
	return finish(root, "kinetic-text", "flex flex-wrap font-[300]",
		"--hover-padding:calc(1em / 12);--text-stroke-width:calc(1em * 125 / 6000)",
	)
}
