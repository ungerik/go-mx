package doc

type Strike int

const (
	NoStrike Strike = iota
	Strikethrough
	Underline
)

type Color uint32

type Span struct {
	ID     string
	Text   string
	Bold   bool
	Italic bool
	Strike Strike
	Color  Color
	Font   *Font
}
