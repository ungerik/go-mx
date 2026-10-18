package dbdoc

import "uuid"

// Document is a stored document identified by ID and a name.
type Document struct {
	ID   uuid.UUID `db:"id,pk"`
	Name string    `db:"name"`
}
