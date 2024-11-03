package dbdoc

import (
	"time"

	"github.com/domonda/go-types/uu"
)

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

type ParagraphText struct {
	ParagraphID uu.ID `db:"paragraph_id,pk"`
	TextID      uu.ID `db:"text_id,pk"`
	Position    int32 `db:"position"`
}

type DocVersionText struct {
	DocVersionID uu.ID `db:"doc_version_id,pk"`
	TextID       uu.ID `db:"text_id,pk"`
	Position     int32 `db:"position"`
}
