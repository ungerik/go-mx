package dbdoc

import (
	"time"
	"uuid"
)

// Paragraph is a stored document paragraph identified by ID.
type Paragraph struct {
	ID        uuid.UUID `db:"id,pk"`
	CreatedAt time.Time `db:"created_at,default"`
}

// DocVersionParagraph is the join record linking a Paragraph to a DocVersion
// at a given position within that version.
type DocVersionParagraph struct {
	DocVersionID uuid.UUID `db:"doc_version_id,pk"`
	ParagraphID  uuid.UUID `db:"paragraph_id,pk"`
	Position     int32     `db:"position"`
}
