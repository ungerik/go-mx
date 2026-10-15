module github.com/ungerik/go-mx/fpdf

go 1.26.0

// go-mx is the parent module, co-developed in this repo. The workspace (go.work)
// resolves it locally; this replace makes non-workspace builds resolve it too.
replace github.com/ungerik/go-mx => ../

// The parity tests compare this legacy wrapper's output against the native
// github.com/ungerik/go-mx/pdf package; resolved locally via go.work.
require github.com/ungerik/go-mx v0.0.0-00010101000000-000000000000

require (
	codeberg.org/go-pdf/fpdf v0.12.0
	github.com/domonda/go-pretty v1.0.0
)

require github.com/domonda/go-errs v1.0.4 // indirect
