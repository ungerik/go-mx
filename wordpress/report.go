package wordpress

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// Report is the import audit trail — the diagnostics a developer needs to find
// every place a WXR export did not translate cleanly. v1 surfaces it three ways
// (stdout summary, JSON file, and an HTML page); this is the importer's wedge.
//
// The parser fills SourceFiles, WXRVersion, Counts and SkippedItems. The content
// pipeline fills UnknownShortcodes, UnsupportedBlocks, DroppedHTML and the link/
// media findings as posts render.
type Report struct {
	SourceFiles  []string      `json:"sourceFiles,omitempty"`
	WXRVersion   string        `json:"wxrVersion,omitempty"`
	Counts       Counts        `json:"counts"`
	SkippedItems []SkippedItem `json:"skippedItems,omitempty"`

	UnknownShortcodes []Finding `json:"unknownShortcodes,omitempty"`
	UnsupportedBlocks []Finding `json:"unsupportedBlocks,omitempty"` // non-core Gutenberg blocks (plugin blocks)
	DroppedHTML       []Finding `json:"droppedHTML,omitempty"`       // elements/attributes removed by the sanitizer
	RewrittenLinks    int       `json:"rewrittenLinks,omitempty"`
	BlockedURLs       []Finding `json:"blockedURLs,omitempty"` // dangerous-scheme URLs removed

	// insertion-order indexes for the *Finding slices above (not serialized)
	idx map[string]map[string]int
}

// Counts summarizes what was imported.
type Counts struct {
	Posts       int `json:"posts"`
	Pages       int `json:"pages"`
	Categories  int `json:"categories"`
	Tags        int `json:"tags"`
	Authors     int `json:"authors"`
	Attachments int `json:"attachments"`
	MenuItems   int `json:"menuItems"`
	Comments    int `json:"comments"`
	Skipped     int `json:"skipped"`
}

// SkippedItem records a WXR item the importer did not turn into model content
// (an unsupported post_type, or an item that failed to map).
type SkippedItem struct {
	PostID   int64  `json:"postID,omitempty"`
	PostType string `json:"postType,omitempty"`
	Title    string `json:"title,omitempty"`
	Reason   string `json:"reason"`
}

// Finding is one class of content issue (a shortcode name, a block name, a drop
// reason), with how often it occurred, what the importer did about it, and which
// source posts it came from — so each finding is a concrete to-do, not a mystery.
type Finding struct {
	Name        string  `json:"name"`
	Count       int     `json:"count"`
	Disposition string  `json:"disposition,omitempty"` // what the importer did
	PostIDs     []int64 `json:"postIDs,omitempty"`     // capped sample of affected posts
}

const maxFindingPosts = 25

// record adds one occurrence of name to the named finding slice, de-duplicating
// by name in first-seen order and capping the per-finding post sample.
func (r *Report) record(kind, name, disposition string, postID int64) {
	if r.idx == nil {
		r.idx = map[string]map[string]int{}
	}
	if r.idx[kind] == nil {
		r.idx[kind] = map[string]int{}
	}
	slice := r.sliceFor(kind)
	if i, ok := r.idx[kind][name]; ok {
		f := &(*slice)[i]
		f.Count++
		f.addPost(postID)
		return
	}
	r.idx[kind][name] = len(*slice)
	f := Finding{Name: name, Count: 1, Disposition: disposition}
	f.addPost(postID)
	*slice = append(*slice, f)
}

func (r *Report) sliceFor(kind string) *[]Finding {
	switch kind {
	case "shortcode":
		return &r.UnknownShortcodes
	case "block":
		return &r.UnsupportedBlocks
	case "dropped":
		return &r.DroppedHTML
	case "blockedURL":
		return &r.BlockedURLs
	default:
		return &r.DroppedHTML
	}
}

func (f *Finding) addPost(postID int64) {
	if postID == 0 || len(f.PostIDs) >= maxFindingPosts {
		return
	}
	if slices.Contains(f.PostIDs, postID) {
		return
	}
	f.PostIDs = append(f.PostIDs, postID)
}

// JSON renders the report as indented, machine-readable JSON — diffable across
// re-runs and gateable in CI.
func (r *Report) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// Summary is the human stdout digest printed at the end of every run: counts,
// then the top offenders per finding class.
func (r *Report) Summary() string {
	var b strings.Builder
	c := r.Counts
	fmt.Fprintf(&b, "Imported %d posts, %d pages, %d categories, %d tags, %d authors, %d attachments, %d comments.\n",
		c.Posts, c.Pages, c.Categories, c.Tags, c.Authors, c.Attachments, c.Comments)
	summarizeFindings(&b, "unknown shortcodes", r.UnknownShortcodes)
	summarizeFindings(&b, "plugin blocks (rendered as raw inner HTML)", r.UnsupportedBlocks)
	summarizeFindings(&b, "removed elements/attributes", r.DroppedHTML)
	summarizeFindings(&b, "blocked URLs", r.BlockedURLs)
	if r.RewrittenLinks > 0 {
		fmt.Fprintf(&b, "  %d internal links rewritten\n", r.RewrittenLinks)
	}
	if c.Skipped > 0 {
		fmt.Fprintf(&b, "⚠ %d items skipped (unsupported post types)\n", c.Skipped)
	}
	return b.String()
}

func summarizeFindings(b *strings.Builder, label string, fs []Finding) {
	if len(fs) == 0 {
		return
	}
	total := 0
	for _, f := range fs {
		total += f.Count
	}
	fmt.Fprintf(b, "⚠ %d %s: ", total, label)
	parts := make([]string, 0, len(fs))
	for i, f := range fs {
		if i >= 6 {
			parts = append(parts, "…")
			break
		}
		parts = append(parts, fmt.Sprintf("%s ×%d", f.Name, f.Count))
	}
	b.WriteString(strings.Join(parts, ", "))
	b.WriteByte('\n')
}

// InheritParse copies parse-time fields (source files, skipped items) from the
// report returned by [Parse]/[ParseFile] into the report returned by
// [WriteStatic], so one report carries the full picture for the stdout summary.
func (r *Report) InheritParse(parse *Report) {
	if parse == nil {
		return
	}
	r.SourceFiles = parse.SourceFiles
	if r.WXRVersion == "" {
		r.WXRVersion = parse.WXRVersion
	}
	r.SkippedItems = append(r.SkippedItems, parse.SkippedItems...)
	r.Counts.Skipped += parse.Counts.Skipped
}
