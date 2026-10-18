package dbdoc

import (
	"time"
	"uuid"
)

// Color is a stored text color value.
type Color int32 // TODO

// Font is a stored font definition holding the typographic properties
// applied to document text, mirroring the corresponding CSS font properties.
type Font struct {
	ID          uuid.UUID `db:"id,pk"`
	FontFamily  string    `db:"font_family"`
	FontSize    string    `db:"font_size"`
	FontStyle   string    `db:"font_style"`
	FontVariant string    `db:"font_variant"`
	FontWeight  string    `db:"font_weight"`
	LineHeight  string    `db:"line_hight"`
	CreatedAt   time.Time `db:"created_at,default"`
}

// DocumentFont is the join record linking a Font to a Document.
type DocumentFont struct {
	DocumentID uuid.UUID `db:"document_id,pk"`
	FontID     uuid.UUID `db:"font_id,pk"`
}
