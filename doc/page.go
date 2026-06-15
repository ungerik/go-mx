package doc

import "github.com/ungerik/go-mx"

// Page is a top-level document page identified by ID and holding its content
// as a single child Component.
type Page struct {
	ID       string
	Children mx.Component
}
