package dbdoc

import (
	"time"

	"github.com/domonda/go-types/nullable"
	"github.com/domonda/go-types/uu"
)

type DocVersion struct {
	ID          uu.ID                  `db:"id,pk"`
	DocumentID  uu.ID                  `db:"document_id"`
	Description nullable.TrimmedString `db:"description"`
	CreatedBy   nullable.TrimmedString `db:"created_by"`
	CreatedAt   time.Time              `db:"created_at,default"`
}
