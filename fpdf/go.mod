module github.com/ungerik/go-mx/fpdf

go 1.26

require (
	codeberg.org/go-pdf/fpdf v0.12.0
	github.com/domonda/go-pretty v1.0.0
)

// The parity tests compare this legacy wrapper's output against the native
// github.com/ungerik/go-mx/pdf package; resolved locally via go.work.
require github.com/ungerik/go-mx v0.0.0
