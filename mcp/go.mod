module github.com/ungerik/go-mx/mcp

go 1.27.0

// go-mx is the parent module, co-developed in this repo. The workspace (go.work)
// resolves it locally; this replace makes non-workspace builds resolve it too.
replace github.com/ungerik/go-mx => ../

require (
	github.com/domonda/go-errs v1.0.4
	github.com/modelcontextprotocol/go-sdk v1.8.0
	github.com/stretchr/testify v1.12.1
	github.com/ungerik/go-mx v0.0.0-00010101000000-000000000000
)

require (
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/adrg/frontmatter v0.2.0 // indirect
	github.com/domonda/go-pretty v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/google/jsonschema-go v0.4.3 // indirect
	github.com/pkg/xattr v0.4.12 // indirect
	github.com/segmentio/asm v1.1.5 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/ungerik/go-fs v0.1.0 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
