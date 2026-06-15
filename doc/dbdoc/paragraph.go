package dbdoc

import (
	"time"

	"github.com/domonda/go-types/uu"
)

// Paragraph is a stored document paragraph identified by ID.
type Paragraph struct {
	ID        uu.ID     `db:"id,pk"`
	CreatedAt time.Time `db:"created_at,default"`
}

// DocVersionParagraph is the join record linking a Paragraph to a DocVersion
// at a given position within that version.
type DocVersionParagraph struct {
	DocVersionID uu.ID `db:"doc_version_id,pk"`
	ParagraphID  uu.ID `db:"paragraph_id,pk"`
	Position     int32 `db:"position"`
}
