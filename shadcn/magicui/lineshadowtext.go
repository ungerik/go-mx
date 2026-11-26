//go:generate go -C ../../tools tool go-enum ../shadcn/magicui/$GOFILE

package magicui

import (
	"fmt"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
)

// LineShadowTextTag is the element a [LineShadowText] renders as.
type LineShadowTextTag string //#enum

const (
	// LineShadowTextSpan renders a span, upstream's default.
	LineShadowTextSpan LineShadowTextTag = "span"
	// LineShadowTextArticle renders an article.
	LineShadowTextArticle LineShadowTextTag = "article"
	// LineShadowTextDiv renders a div.
	LineShadowTextDiv LineShadowTextTag = "div"
	// LineShadowTextH1 renders an h1 heading.
	LineShadowTextH1 LineShadowTextTag = "h1"
	// LineShadowTextH2 renders an h2 heading.
	LineShadowTextH2 LineShadowTextTag = "h2"
	// LineShadowTextH3 renders an h3 heading.
	LineShadowTextH3 LineShadowTextTag = "h3"
	// LineShadowTextH4 renders an h4 heading.
	LineShadowTextH4 LineShadowTextTag = "h4"
	// LineShadowTextH5 renders an h5 heading.
	LineShadowTextH5 LineShadowTextTag = "h5"
	// LineShadowTextH6 renders an h6 heading.
	LineShadowTextH6 LineShadowTextTag = "h6"
	// LineShadowTextLi renders a list item.
	LineShadowTextLi LineShadowTextTag = "li"
	// LineShadowTextP renders a paragraph.
	LineShadowTextP LineShadowTextTag = "p"
	// LineShadowTextSection renders a section.
	LineShadowTextSection LineShadowTextTag = "section"
)

// Valid indicates if l is any of the valid values for LineShadowTextTag
func (l LineShadowTextTag) Valid() bool {
	switch l {
	case
		LineShadowTextSpan,
		LineShadowTextArticle,
		LineShadowTextDiv,
		LineShadowTextH1,
		LineShadowTextH2,
		LineShadowTextH3,
		LineShadowTextH4,
		LineShadowTextH5,
		LineShadowTextH6,
		LineShadowTextLi,
		LineShadowTextP,
		LineShadowTextSection:
		return true
	}
	return false
}

// Validate returns an error if l is none of the valid values for LineShadowTextTag
func (l LineShadowTextTag) Validate() error {
	if !l.Valid() {
		return fmt.Errorf("invalid value %#v for type magicui.LineShadowTextTag", l)
	}
	return nil
}

// ValidOr returns l if it is any of the valid values for LineShadowTextTag, else defaultVal
func (l LineShadowTextTag) ValidOr(defaultVal LineShadowTextTag) LineShadowTextTag {
	if l.Valid() {
		return l
	}
	return defaultVal
}

// Enums returns all valid values for LineShadowTextTag
func (LineShadowTextTag) Enums() []LineShadowTextTag {
	return []LineShadowTextTag{
		LineShadowTextSpan,
		LineShadowTextArticle,
		LineShadowTextDiv,
		LineShadowTextH1,
		LineShadowTextH2,
		LineShadowTextH3,
		LineShadowTextH4,
		LineShadowTextH5,
		LineShadowTextH6,
		LineShadowTextLi,
		LineShadowTextP,
		LineShadowTextSection,
	}
}

// EnumStrings returns all valid values for LineShadowTextTag as strings
func (LineShadowTextTag) EnumStrings() []string {
	return []string{
		"span",
		"article",
		"div",
		"h1",
		"h2",
		"h3",
		"h4",
		"h5",
		"h6",
		"li",
		"p",
		"section",
	}
}

// String implements the fmt.Stringer interface for LineShadowTextTag
func (l LineShadowTextTag) String() string {
	return string(l)
}

// LineShadowTextOptions configures a [LineShadowText]. The zero value is
// upstream's default: a black shadow on a span.
type LineShadowTextOptions struct {
	// ShadowColor is the CSS color of the shadow's lines; "" means "black".
	ShadowColor string
	// As is the element to render; "" means [LineShadowTextSpan].
	As LineShadowTextTag
}

// LineShadowText renders its text with a hatched shadow behind it whose
// diagonal lines keep moving. The shadow is an ::after copy of the text,
// taken from the data-text attribute, so the children must be strings, as
// upstream's children: string; another child makes it render an error, as
// does a shadow color that could escape its CSS declaration. The color is
// the CSS variable --shadow-color.
//
// Upstream renders a motion element only to accept motion props; nothing
// of the effect needs it, so this is a plain element.
//
// Needs [CSS] in the app's Tailwind input.
func LineShadowText(opts LineShadowTextOptions, attribsChildren ...any) *mx.Element {
	style, err := styleVars("--shadow-color", or(opts.ShadowColor, "black"))
	if err != nil {
		return mx.NewErrElement(err)
	}
	root := mx.NewElement(string(opts.As.ValidOr(LineShadowTextSpan)), attribsChildren...)
	if root.Err != nil {
		return root
	}
	text, err := childText("LineShadowText", root)
	if err != nil {
		return mx.NewErrElement(err)
	}
	root.Attribs = append(root.Attribs, html.DataAttr("text", text))
	return finish(root, "line-shadow-text",
		"relative z-0 inline-flex after:absolute after:top-[0.04em] after:left-[0.04em] after:content-[attr(data-text)] after:bg-[linear-gradient(45deg,transparent_45%,var(--shadow-color)_45%,var(--shadow-color)_55%,transparent_0)] after:-z-10 after:bg-size-[0.06em_0.06em] after:bg-clip-text after:text-transparent after:animate-line-shadow",
		style,
	)
}
