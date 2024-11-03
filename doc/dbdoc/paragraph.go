package dbdoc

import (
	"time"

	"github.com/domonda/go-types/uu"
)

type Paragraph struct {
	ID        uu.ID     `db:"id,pk"`
	CreatedAt time.Time `db:"created_at,default"`
}

type DocVersionParagraph struct {
	DocVersionID uu.ID `db:"doc_version_id,pk"`
	ParagraphID  uu.ID `db:"paragraph_id,pk"`
	Position     int32 `db:"position"`
}
