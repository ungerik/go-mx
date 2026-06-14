package html

import "github.com/ungerik/go-mx"

type (
	// Text is a string rendered as HTML text content with special
	// characters escaped. Alias for [mx.Text].
	Text = mx.Text
	// Raw is a string rendered verbatim without HTML escaping; the caller
	// is responsible for its safety. Alias for [mx.Raw].
	Raw = mx.Raw
	// RawBytes is a byte slice rendered verbatim without HTML escaping;
	// the caller is responsible for its safety. Alias for [mx.RawBytes].
	RawBytes = mx.RawBytes
)

// HTML named character references such as Copyright (©) and NonBreakingSpace
// have moved to the html/entity subpackage to keep this package's namespace
// free of collisions with element and attribute constructors.
