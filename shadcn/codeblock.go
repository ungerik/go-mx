//go:generate go -C ../tools tool go-enum ../shadcn/$GOFILE

package shadcn

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn/cva"
)

// CodeBlock is a Go port of shadcn/ui's CodeBlock, a framed code listing with
// an optional header, actions such as a copy button, line numbers, diff
// lines and a footer. Class strings are reconstructed from the
// post-July-2026 registry (code-block.tsx skeleton + the style-vega.css
// .cn-code-block* rules) — see "Upstream restructure (July 2026)" in
// TODOS.md.
//
// CodeBlock does not highlight. Pass highlighted markup as the children of
// [CodeBlockContent], for Go code the tokens of the highlight package:
//
//	shadcn.CodeBlockContent(highlight.Default.Components(highlight.TokenizeGo(src)))
//
// plus that package's Theme stylesheet for the colors. The dependency points
// that way on purpose: highlight depends only on mx and html, so that it
// drops into any markup, and shadcn does not import it.

// CodeBlockVariant selects a [CodeBlock]'s colors.
type CodeBlockVariant string //#enum

const (
	// CodeBlockDefault is the default frame.
	CodeBlockDefault CodeBlockVariant = "default"
	// CodeBlockDestructive colors the frame and text destructive, for an
	// error output.
	CodeBlockDestructive CodeBlockVariant = "destructive"
)

// Valid indicates if c is any of the valid values for CodeBlockVariant
func (c CodeBlockVariant) Valid() bool {
	switch c {
	case
		CodeBlockDefault,
		CodeBlockDestructive:
		return true
	}
	return false
}

// Validate returns an error if c is none of the valid values for CodeBlockVariant
func (c CodeBlockVariant) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.CodeBlockVariant", c)
	}
	return nil
}

// ValidOr returns c if it is any of the valid values for CodeBlockVariant, else defaultVal
func (c CodeBlockVariant) ValidOr(defaultVal CodeBlockVariant) CodeBlockVariant {
	if c.Valid() {
		return c
	}
	return defaultVal
}

// Enums returns all valid values for CodeBlockVariant
func (CodeBlockVariant) Enums() []CodeBlockVariant {
	return []CodeBlockVariant{
		CodeBlockDefault,
		CodeBlockDestructive,
	}
}

// EnumStrings returns all valid values for CodeBlockVariant as strings
func (CodeBlockVariant) EnumStrings() []string {
	return []string{
		"default",
		"destructive",
	}
}

// String implements the fmt.Stringer interface for CodeBlockVariant
func (c CodeBlockVariant) String() string {
	return string(c)
}

// CodeBlockSize selects a [CodeBlock]'s padding and text size.
type CodeBlockSize string //#enum

const (
	// CodeBlockSizeDefault is the default size.
	CodeBlockSizeDefault CodeBlockSize = "default"
	// CodeBlockSizeSM is the small size, for output nested in a transcript.
	CodeBlockSizeSM CodeBlockSize = "sm"
)

// Valid indicates if c is any of the valid values for CodeBlockSize
func (c CodeBlockSize) Valid() bool {
	switch c {
	case
		CodeBlockSizeDefault,
		CodeBlockSizeSM:
		return true
	}
	return false
}

// Validate returns an error if c is none of the valid values for CodeBlockSize
func (c CodeBlockSize) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.CodeBlockSize", c)
	}
	return nil
}

// ValidOr returns c if it is any of the valid values for CodeBlockSize, else defaultVal
func (c CodeBlockSize) ValidOr(defaultVal CodeBlockSize) CodeBlockSize {
	if c.Valid() {
		return c
	}
	return defaultVal
}

// Enums returns all valid values for CodeBlockSize
func (CodeBlockSize) Enums() []CodeBlockSize {
	return []CodeBlockSize{
		CodeBlockSizeDefault,
		CodeBlockSizeSM,
	}
}

// EnumStrings returns all valid values for CodeBlockSize as strings
func (CodeBlockSize) EnumStrings() []string {
	return []string{
		"default",
		"sm",
	}
}

// String implements the fmt.Stringer interface for CodeBlockSize
func (c CodeBlockSize) String() string {
	return string(c)
}

// codeBlockVariants resolves a code block's base + variant classes, declared
// the same way shadcn/ui's code-block.tsx declares them with cva.
var codeBlockVariants = cva.New(cva.Config{
	Base: "group/code-block relative flex w-full min-w-0 flex-col overflow-hidden rounded-lg text-sm data-[size=sm]:text-xs has-[>[data-slot=code-block-content]:focus-visible]:ring-3 has-[>[data-slot=code-block-content]:focus-visible]:ring-ring/50",
	Variants: map[string]map[string]string{
		"variant": {
			"default":     "border border-input text-foreground dark:bg-input/30 has-[>[data-slot=code-block-content]:focus-visible]:border-ring",
			"destructive": "border border-destructive/30 text-destructive dark:border-destructive/40",
		},
	},
	DefaultVariants: map[string]string{"variant": "default"},
})

// CodeBlock renders the frame of a code listing. variant and size may be ""
// for the defaults; unknown values also resolve to them. Compose it with
// [CodeBlockHeader], [CodeBlockContent] and [CodeBlockFooter]; a
// [CodeBlockActions] placed directly in the CodeBlock, not in a header,
// floats over the top corner and shows on hover and focus.
func CodeBlock(variant CodeBlockVariant, size CodeBlockSize, attribsChildren ...any) *mx.Element {
	v := string(variant.ValidOr(CodeBlockDefault))
	s := string(size.ValidOr(CodeBlockSizeDefault))
	e := html.Div(attribsChildren...)
	e.Attribs = append(e.Attribs,
		html.DataAttr("variant", v),
		html.DataAttr("size", s),
	)
	return finish(e, "code-block", Cn(codeBlockVariants(map[string]string{"variant": v})))
}

// CodeBlockHeader renders a row above the code, holding a [CodeBlockTitle]
// and [CodeBlockActions].
func CodeBlockHeader(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "code-block-header",
		"flex min-w-0 shrink-0 items-center *:has-[>[data-slot=code-block-title]]:min-w-0 gap-2 px-4 pt-4 text-muted-foreground group-data-[variant=destructive]/code-block:text-destructive group-data-[size=sm]/code-block:px-3 group-data-[size=sm]/code-block:pt-3")
}

// CodeBlockTitle renders a [CodeBlock]'s title, such as a file name, with an
// optional leading icon.
func CodeBlockTitle(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "code-block-title",
		"min-w-0 truncate [&_svg]:pointer-events-none [&_svg]:inline-block [&_svg]:align-middle font-sans text-sm font-medium text-foreground group-data-[variant=destructive]/code-block:text-destructive *:[svg]:me-1.5 [&_svg:not([class*='size-'])]:size-4 group-data-[size=sm]/code-block:text-xs group-data-[size=sm]/code-block:[&_svg:not([class*='size-'])]:size-3.5")
}

// CodeBlockActions holds a [CodeBlock]'s buttons, such as a
// [CodeBlockCopyButton]. In a [CodeBlockHeader] it sits at the row's end;
// directly in the CodeBlock it floats over the top end corner, hidden until
// the block is hovered or holds focus (always shown on touch screens).
func CodeBlockActions(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "code-block-actions",
		"ms-auto flex shrink-0 items-center [[data-slot=code-block]>&]:absolute [[data-slot=code-block]>&]:z-10 [[data-slot=code-block]>&]:opacity-0 [[data-slot=code-block]>&]:transition-opacity [[data-slot=code-block]>&]:group-focus-within/code-block:opacity-100 [[data-slot=code-block]>&]:group-hover/code-block:opacity-100 [[data-slot=code-block]>&]:pointer-coarse:opacity-100 -my-1 -me-1.5 gap-1 [[data-slot=code-block]>&]:m-0 [[data-slot=code-block]>&]:top-3 [[data-slot=code-block]>&]:end-3 group-data-[size=sm]/code-block:[[data-slot=code-block]>&]:top-2 group-data-[size=sm]/code-block:[[data-slot=code-block]>&]:end-2 [[data-slot=code-block]>&]:*:[button]:bg-background [[data-slot=code-block]>&]:*:[button]:ring-1 [[data-slot=code-block]>&]:*:[button]:ring-foreground/10")
}

// codeBlockCopyScript is the once-emitted client behavior behind
// [CodeBlockCopyButton], the counterpart of upstream's React state: it writes
// the button's data-value (decoded, see copyValueEscaper), or else the text
// of the code block around it, to
// the clipboard, then sets data-copied (which swaps the copy icon for a
// check) and the "Copied" label for two seconds. Upstream announces the copy
// through a polite live region next to each button; here one region per page
// is created when the script is defined, because a region inserted at the
// moment its text changes is often not announced.
//
// The text of a block numbered by CodeBlockLine rows is joined from the rows'
// contents, so that the line numbers (CSS counters) stay out and an empty
// row, which holds a newline, gives an empty line rather than two.
const codeBlockCopyScript = /*js*/ `if(!window.mxCodeBlockCopy){(function(){var r=document.createElement('span');r.className='sr-only';r.setAttribute('aria-live','polite');(document.body||document.documentElement).appendChild(r);window.mxCodeBlockCopy=function(b){var v=b.dataset.value;if(v!=null){v=decodeURIComponent(v);}else{var c=b.closest('[data-slot=code-block]');var k=c&&c.querySelector('[data-slot=code-block-code]');var ls=k?k.querySelectorAll('[data-slot=code-block-line-content]'):[];v=ls.length?Array.prototype.map.call(ls,function(l){return l.textContent.replace(/\n$/,'');}).join('\n'):(k?k.textContent:'');}navigator.clipboard.writeText(v).then(function(){b.setAttribute('data-copied','');b.setAttribute('aria-label',b.dataset.copiedLabel);r.textContent=b.dataset.copiedLabel;clearTimeout(b.mxCopyTimer);b.mxCopyTimer=setTimeout(function(){b.removeAttribute('data-copied');b.setAttribute('aria-label',b.dataset.copyLabel);r.textContent='';},2000);}).catch(function(){});};})();}`

// copyValueEscaper percent-encodes the characters of a [CodeBlockCopyButton]
// value that go-mx's attribute escaping would replace — a line break becomes
// a space, a tab two — plus the percent sign itself, so that the script's
// decodeURIComponent restores the value exactly.
var copyValueEscaper = strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", "\t", "%09")

// CodeBlockCopyButton renders an icon [Button] (ghost, icon-xs) that copies
// value to the clipboard, or, when value is "", the text of the [CodeBlock]
// it is in. For two seconds after a copy it shows a check icon and is
// labelled "Copied" instead of "Copy" (both localizable via [Labels]).
//
// It emits one small inline script, guarded so it is defined once per page,
// and calls it from its onclick; a caller-supplied onclick replaces the
// copying.
func CodeBlockCopyButton(value string, attribs ...mx.Attrib) *mx.Element {
	args := make([]any, 0, len(attribs)+6)
	for _, a := range attribs {
		args = append(args, a)
	}
	args = append(args,
		html.ScriptJS(codeBlockCopyScript),
		iconCopy(),
		iconCheck(),
	)
	e := Button(ButtonGhost, SizeIconXS, args...)
	if value != "" {
		e.Attribs = append(e.Attribs, html.DataAttr("value", copyValueEscaper.Replace(value)))
	}
	if e.AttribIndex("aria-label") < 0 {
		e.Attribs = append(e.Attribs, labelAriaLabel(func(l Labels) string { return l.CodeBlockCopy }))
	}
	if e.AttribIndex("onclick") < 0 {
		e.Attribs = append(e.Attribs, html.OnClick("mxCodeBlockCopy(this)"))
	}
	e.Attribs = append(e.Attribs,
		labelAttrib{name: "data-copy-label", get: func(l Labels) string { return l.CodeBlockCopy }},
		labelAttrib{name: "data-copied-label", get: func(l Labels) string { return l.CodeBlockCopied }},
	)
	return finish(e, "code-block-copy-button",
		"[&>.lucide-check]:hidden data-copied:[&>.lucide-copy]:hidden data-copied:[&>.lucide-check]:block")
}

// CodeBlockWrap makes a [CodeBlockContent] wrap long lines instead of
// scrolling sideways.
const CodeBlockWrap = mx.ConstAttrib("data-wrap=")

// CodeBlockLineNumbers numbers the lines of a [CodeBlockContent]. A single
// plain-text child is split into [CodeBlockLine]s for that, as upstream does
// with a string; other children, such as highlighted tokens, must already be
// CodeBlockLines.
const CodeBlockLineNumbers = mx.ConstAttrib("data-numbered=")

// CodeBlockStartLine makes the line numbers of a [CodeBlockContent] start at
// n instead of 1, for an excerpt.
func CodeBlockStartLine(n int) mx.Attrib {
	return html.DataAttr("start-line", strconv.Itoa(n))
}

// CodeBlockContent renders the code as <pre><code>. The attributes go to the
// <pre>, which takes [CodeBlockWrap], [CodeBlockLineNumbers] and
// [CodeBlockStartLine]; the children go to the <code>: text, highlighted
// markup or [CodeBlockLine]s. The <code> lays its children out one per row,
// for the lines; children that are not all CodeBlockLines, such as a run of
// highlighted tokens, are wrapped in one <span> so they flow as text. The <pre> is focusable (tabindex="0") so the
// keyboard can scroll it, and the CodeBlock draws its focus ring.
//
// The <code> is rendered without the line breaks and indentation an
// indenting [mx.CheckedWriter] puts before each element, which inside a
// <pre> would be part of the code.
func CodeBlockContent(attribsChildren ...any) *mx.Element {
	e := html.Pre(attribsChildren...)
	children := flattenComponents(e.Children)
	if e.AttribIndex("data-numbered") >= 0 && len(children) == 1 {
		if text, ok := children[0].(mx.Text); ok {
			lines := strings.Split(strings.TrimSuffix(string(text), "\n"), "\n")
			children = make(mx.Components, len(lines))
			for i, line := range lines {
				children[i] = CodeBlockLine("", line)
			}
		}
	}
	if len(children) > 1 && !allCodeBlockLines(children) {
		// The <code> is a grid with one row per line, which would put every
		// highlighted token on a row of its own; in one span they flow as
		// the text they are.
		children = mx.Components{html.Span(children)}
	}
	start := 1
	if i := e.AttribIndex("data-start-line"); i >= 0 {
		if v, err := e.Attribs[i].AttribValue(context.Background()); err == nil {
			if n, err := strconv.Atoi(v); err == nil {
				start = n
			}
		}
	}
	mergeStyle(e, "counter-reset: line "+strconv.Itoa(start-1))
	if e.AttribIndex("dir") < 0 {
		e.Attribs = append(e.Attribs, html.DirLTR)
	}
	if e.AttribIndex("tabindex") < 0 {
		e.Attribs = append(e.Attribs, html.TabIndex(0))
	}
	code := html.Code(html.DataAttr("slot", "code-block-code"),
		html.Class("grid min-h-lh min-w-max content-start in-data-numbered:grid-cols-[auto_1fr] in-data-wrap:min-w-0"))
	code.Children = children
	e.Children = mx.Components{unindented{code}}
	return finish(e, "code-block-content",
		"min-h-0 min-w-0 [scrollbar-width:thin] overflow-auto text-start font-mono outline-none data-wrap:wrap-break-word data-wrap:whitespace-pre-wrap px-4 py-4 leading-relaxed group-has-data-[slot=code-block-header]/code-block:pt-2 group-has-data-[slot=code-block-footer]/code-block:pb-2 group-data-[size=sm]/code-block:px-3 group-data-[size=sm]/code-block:py-3 group-data-[size=sm]/code-block:group-has-data-[slot=code-block-header]/code-block:pt-1.5")
}

// flattenComponents returns cs with every nested [mx.Components] spliced in,
// so that highlighted tokens passed as one mx.Components count as the tokens
// they are.
func flattenComponents(cs mx.Components) mx.Components {
	flat := mx.Components{}
	for _, c := range cs {
		if nested, ok := c.(mx.Components); ok {
			flat = append(flat, flattenComponents(nested)...)
		} else {
			flat = append(flat, c)
		}
	}
	return flat
}

// allCodeBlockLines reports whether every one of children is a
// [CodeBlockLine].
func allCodeBlockLines(children mx.Components) bool {
	for _, c := range children {
		e, ok := c.(*mx.Element)
		if !ok {
			return false
		}
		i := e.AttribIndex("data-slot")
		if i < 0 {
			return false
		}
		if slot, _ := e.Attribs[i].AttribValue(context.Background()); slot != "code-block-line" {
			return false
		}
	}
	return true
}

// unindented renders c without the line breaks and indentation an indenting
// [mx.CheckedWriter] puts before every element, by rendering it through a
// non-indenting clone of the writer. Any other writer renders c directly.
type unindented struct{ c mx.Component }

// Render implements [mx.Component].
func (u unindented) Render(ctx context.Context, w mx.Writer) error {
	cw, ok := w.(*mx.CheckedWriter)
	if !ok {
		return u.c.Render(ctx, w)
	}
	var buf bytes.Buffer
	if err := u.c.Render(ctx, cw.Clone(&buf).WithIndent("", "")); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

// CodeBlockLineVariant selects how a [CodeBlockLine] is marked.
type CodeBlockLineVariant string //#enum

const (
	// CodeBlockLineDefault is an unmarked line.
	CodeBlockLineDefault CodeBlockLineVariant = "default"
	// CodeBlockLineAdded marks an added line of a diff with a "+".
	CodeBlockLineAdded CodeBlockLineVariant = "added"
	// CodeBlockLineRemoved marks a removed line of a diff with a "−".
	CodeBlockLineRemoved CodeBlockLineVariant = "removed"
	// CodeBlockLineHighlighted gives the line a muted background.
	CodeBlockLineHighlighted CodeBlockLineVariant = "highlighted"
)

// Valid indicates if c is any of the valid values for CodeBlockLineVariant
func (c CodeBlockLineVariant) Valid() bool {
	switch c {
	case
		CodeBlockLineDefault,
		CodeBlockLineAdded,
		CodeBlockLineRemoved,
		CodeBlockLineHighlighted:
		return true
	}
	return false
}

// Validate returns an error if c is none of the valid values for CodeBlockLineVariant
func (c CodeBlockLineVariant) Validate() error {
	if !c.Valid() {
		return fmt.Errorf("invalid value %#v for type shadcn.CodeBlockLineVariant", c)
	}
	return nil
}

// ValidOr returns c if it is any of the valid values for CodeBlockLineVariant, else defaultVal
func (c CodeBlockLineVariant) ValidOr(defaultVal CodeBlockLineVariant) CodeBlockLineVariant {
	if c.Valid() {
		return c
	}
	return defaultVal
}

// Enums returns all valid values for CodeBlockLineVariant
func (CodeBlockLineVariant) Enums() []CodeBlockLineVariant {
	return []CodeBlockLineVariant{
		CodeBlockLineDefault,
		CodeBlockLineAdded,
		CodeBlockLineRemoved,
		CodeBlockLineHighlighted,
	}
}

// EnumStrings returns all valid values for CodeBlockLineVariant as strings
func (CodeBlockLineVariant) EnumStrings() []string {
	return []string{
		"default",
		"added",
		"removed",
		"highlighted",
	}
}

// String implements the fmt.Stringer interface for CodeBlockLineVariant
func (c CodeBlockLineVariant) String() string {
	return string(c)
}

// codeBlockLineVariants resolves a line's base + variant classes, declared
// the same way shadcn/ui's code-block.tsx declares them with cva.
var codeBlockLineVariants = cva.New(cva.Config{
	Base: "relative block min-h-lh after:absolute after:top-0 after:select-none in-data-numbered:col-span-2 in-data-numbered:grid in-data-numbered:grid-cols-subgrid in-data-numbered:[counter-increment:line] in-data-numbered:before:content-[counter(line)] in-data-numbered:before:select-none in-data-numbered:supports-[content:counter(line)_/_'']:before:content-[counter(line)_/_''] -mx-4 px-4 in-data-numbered:gap-x-3 before:min-w-6 before:text-end before:text-muted-foreground after:start-1 after:text-muted-foreground group-data-[size=sm]/code-block:-mx-3 group-data-[size=sm]/code-block:px-3",
	Variants: map[string]map[string]string{
		"variant": {
			"default":     "",
			"added":       "bg-muted after:content-['+']",
			"removed":     "bg-destructive/10 after:text-destructive after:content-['−']",
			"highlighted": "bg-muted",
		},
	},
	DefaultVariants: map[string]string{"variant": "default"},
})

// CodeBlockLine renders one line of a [CodeBlockContent], for line numbers
// and diff or highlight marks. variant may be "" for the default; an unknown
// value also resolves to it. An empty line holds a newline, so that copying
// a selection keeps it.
func CodeBlockLine(variant CodeBlockLineVariant, attribsChildren ...any) *mx.Element {
	v := string(variant.ValidOr(CodeBlockLineDefault))
	e := html.Span(attribsChildren...)
	content := html.Span(html.DataAttr("slot", "code-block-line-content"), html.Class("min-w-0"))
	if len(e.Children) == 0 || len(e.Children) == 1 && e.Children[0] == mx.Text("") {
		content.Children = mx.Components{mx.Text("\n")}
	} else {
		content.Children = e.Children
	}
	e.Children = mx.Components{content}
	e.Attribs = append(e.Attribs, html.DataAttr("variant", v))
	return finish(e, "code-block-line", Cn(codeBlockLineVariants(map[string]string{"variant": v})))
}

// CodeBlockFooter renders a row below the code, such as a caption or a
// status.
func CodeBlockFooter(attribsChildren ...any) *mx.Element {
	return finish(html.Div(attribsChildren...), "code-block-footer",
		"flex min-w-0 shrink-0 items-center gap-2 px-4 pb-4 text-muted-foreground group-data-[size=sm]/code-block:px-3 group-data-[size=sm]/code-block:pb-3")
}
