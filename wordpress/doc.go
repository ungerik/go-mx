// Package wordpress imports the content of a WordPress site from a WXR export
// (WordPress eXtended RSS, produced by the WordPress admin under Tools → Export)
// and re-renders it generically with go-mx + shadcn. It reproduces the
// structure of the original site (posts, pages, archives, navigation, comments)
// in one clean theme, not the source theme's pixel-perfect CSS.
//
// # Layers
//
// The package is four layers with a pure data model in the middle:
//
//   - Source: [Parse], [ParseFile] and [ParseFiles] turn a WXR file (or a split
//     multi-file export) into a [*Site]. Only this layer knows the input format.
//   - Model: [Site] and its parts ([Post], [Page], [Term], [Author], [Comment],
//     [Attachment], [MenuItem]) are plain, encoding/json-serializable Go structs.
//     WordPress IDs are int64 (bigint auto-increment), not UUIDs. Relationships
//     are stored as IDs or slugs, never as parent/child pointer cycles, so the
//     model marshals to a JSON tree. This is the durable, reusable asset.
//   - Render (later step): each logical component is a composable mx.Component
//     view over the model, so a piece can be embedded in a caller's own go-mx page.
//   - Output (later step): a static-site writer.
//
// Parsing also returns a [*Report] — the import diagnostics (skipped items, and,
// once the content pipeline lands, unsupported blocks/shortcodes, missing media
// and rewritten links). The report is how a developer migrating a site finds
// every place that needs manual attention.
//
// # Scope
//
// v1 ingests WXR exclusively and renders a static site. A dynamic HTTP handler,
// a MySQL/REST source, and Markdown output are explicitly out of scope for v1.
// Page-builder layouts (Elementor, Divi, WPBakery) are detected and flagged in
// the report, never faked.
package wordpress
