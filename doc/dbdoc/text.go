package dbdoc

import (
	"time"

	"github.com/domonda/go-types/uu"
)

// Text is a stored run of document text with its styling, such as bold,
// italic, underline, strikethrough, color, and an optional font reference.
type Text struct {
	ID            uu.ID         `db:"id,pk"`
	Bold          bool          `db:"bold"`
	Italic        bool          `db:"italic"`
	Underline     bool          `db:"underline"`
	Strikethrough bool          `db:"strikethrough"`
	Color         *Color        `db:"color"`
	FontID        uu.NullableID `db:"font_id"`
	Text          string        `db:"text"`
	CreatedAt     time.Time     `db:"created_at,default"`
}

// ParagraphText is the join record linking a Text to a Paragraph
// at a given position within that paragraph.
type ParagraphText struct {
	ParagraphID uu.ID `db:"paragraph_id,pk"`
	TextID      uu.ID `db:"text_id,pk"`
	Position    int32 `db:"position"`
}

// DocVersionText is the join record linking a Text to a DocVersion
// at a given position within that version.
type DocVersionText struct {
	DocVersionID uu.ID `db:"doc_version_id,pk"`
	TextID       uu.ID `db:"text_id,pk"`
	Position     int32 `db:"position"`
}
