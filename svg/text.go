package svg

import "github.com/ungerik/go-mx"

// Note: the <text> element is provided by the Text function in elements.go,
// so the mx.Text type is not aliased here to avoid a name collision. Plain Go
// strings passed as children are escaped and rendered as text content.

// Raw is an alias for mx.Raw: a string rendered without HTML/XML escaping.
type Raw = mx.Raw

// RawBytes is an alias for mx.RawBytes: bytes rendered without HTML/XML escaping.
type RawBytes = mx.RawBytes
