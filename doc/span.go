//go:generate go -C ../tools tool go-enum ../doc/$GOFILE

package doc

import "fmt"

// Strike enumerates the line decorations that can be applied to a Span.
type Strike int //#enum

const (
	// NoStrike applies no line decoration to the text.
	NoStrike Strike = iota
	// Strikethrough draws a line through the text.
	Strikethrough
	// Underline draws a line under the text.
	Underline
)

// Valid indicates if s is any of the valid values for Strike
func (s Strike) Valid() bool {
	switch s {
	case
		NoStrike,
		Strikethrough,
		Underline:
		return true
	}
	return false
}

// Validate returns an error if s is none of the valid values for Strike
func (s Strike) Validate() error {
	if !s.Valid() {
		return fmt.Errorf("invalid value %#v for type doc.Strike", s)
	}
	return nil
}

// ValidOr returns s if it is any of the valid values for Strike, else defaultVal
func (s Strike) ValidOr(defaultVal Strike) Strike {
	if s.Valid() {
		return s
	}
	return defaultVal
}

// Enums returns all valid values for Strike
func (Strike) Enums() []Strike {
	return []Strike{
		NoStrike,
		Strikethrough,
		Underline,
	}
}

// EnumStrings returns all valid values for Strike as strings
func (Strike) EnumStrings() []string {
	return []string{
		"0",
		"1",
		"2",
	}
}

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
