module github.com/ungerik/go-mx/mcp

go 1.26

// go-mx is the parent module, co-developed in this repo. The workspace (go.work)
// resolves it locally; this replace makes non-workspace builds resolve it too.
replace github.com/ungerik/go-mx => ../

require (
	github.com/domonda/go-errs v1.0.4
	github.com/modelcontextprotocol/go-sdk v1.7.0
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
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/ungerik/go-fs v0.0.0-20260608082337-4ce02194af8a // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/oauth2 v0.35.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
