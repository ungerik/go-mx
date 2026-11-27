module github.com/ungerik/go-mx/tools

go 1.27.0

tool github.com/ungerik/go-enum

// go-mx is the parent module, co-developed in this repo; gen-docs-site renders
// the documentation site with it. The workspace (go.work) resolves it locally;
// this replace makes non-workspace builds resolve it too.
replace github.com/ungerik/go-mx => ../

require (
	github.com/adrg/frontmatter v0.2.0
	github.com/ungerik/go-mx v0.0.0-00010101000000-000000000000
	github.com/yuin/goldmark v1.8.7
	golang.org/x/text v0.44.0
)

require (
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/domonda/go-errs v1.3.1 // indirect
	github.com/domonda/go-pretty v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/pkg/xattr v0.4.12 // indirect
	github.com/ungerik/go-enum v1.3.0 // indirect
	github.com/ungerik/go-fs v1.0.0-rc.8 // indirect
	golang.org/x/mod v0.43.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/tools v0.53.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
