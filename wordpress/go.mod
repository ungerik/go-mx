module github.com/ungerik/go-mx/wordpress

go 1.26

// go-mx is the parent module, co-developed in this repo. The workspace (go.work)
// resolves it locally; this replace makes non-workspace builds resolve it too.
replace github.com/ungerik/go-mx => ../

require (
	github.com/domonda/go-errs v1.0.3
	github.com/ungerik/go-mx v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.56.0
)

require github.com/domonda/go-pretty v1.0.0 // indirect
