package dbdoc

import (
	"time"
	"uuid"
)

// DocVersion is a stored version of a Document, recording who created it,
// when, and an optional description.
type DocVersion struct {
	ID          uuid.UUID `db:"id,pk"`
	DocumentID  uuid.UUID `db:"document_id"`
	Description *string   `db:"description"`
	CreatedBy   *string   `db:"created_by"`
	CreatedAt   time.Time `db:"created_at,default"`
}
