package dbdoc

import (
	"github.com/domonda/go-types/notnull"
	"github.com/domonda/go-types/uu"
)

// Document is a stored document identified by ID and a name.
type Document struct {
	ID   uu.ID                 `db:"id,pk"`
	Name notnull.TrimmedString `db:"name"`
}
