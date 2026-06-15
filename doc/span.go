package doc

// Strike enumerates the line decorations that can be applied to a Span.
type Strike int

const (
	// NoStrike applies no line decoration to the text.
	NoStrike Strike = iota
	// Strikethrough draws a line through the text.
	Strikethrough
	// Underline draws a line under the text.
	Underline
)

// Color is an RGBA color packed into a 32-bit value.
type Color uint32

// Span is an inline run of text with optional styling such as bold,
// italic, strike decoration, color, and font.
type Span struct {
	ID     string
	Text   string
	Bold   bool
	Italic bool
	Strike Strike
	Color  Color
	Font   *Font
}
