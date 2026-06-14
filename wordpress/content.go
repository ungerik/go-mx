package wordpress

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// contentOptions controls how body HTML is turned into components.
type contentOptions struct {
	baseHost string                     // host of Site.BaseURL, for internal-link detection
	resolve  func(rawURL string) string // optional: map an internal URL to a local path; nil keeps it
}

// renderContent converts raw WordPress body HTML (classic markup + Gutenberg
// inner HTML + shortcodes) into a safe mx.Component wrapped in a .wp-content
// prose container. It is the importer's security boundary: the parsed nodes are
// re-emitted through the html.* constructors (so go-mx's CheckedWriter escapes
// them) and run through an element/attribute allowlist — body HTML is NEVER
// passed to html.Raw. Issues are recorded in rep against postID.
func renderContent(raw ContentHTML, postID int64, opt contentOptions, rep *Report) mx.Component {
	s := stripGutenbergComments(string(raw), postID, rep)
	s = processShortcodes(s, postID, rep)

	ctx := &xhtml.Node{Type: xhtml.ElementNode, Data: "body", DataAtom: atom.Body}
	nodes, err := xhtml.ParseFragment(strings.NewReader(s), ctx)
	if err != nil {
		// x/net/html.ParseFragment is robust and effectively never errors on
		// real input; if it does, drop the body rather than emit anything raw.
		rep.record("dropped", "unparseable body HTML", "dropped", postID)
		return html.Div(html.Class(wpContentClass))
	}

	w := &walker{opt: opt, rep: rep, postID: postID}
	children := make([]any, 0, len(nodes))
	for _, n := range nodes {
		if c := w.node(n, 0, false); c != nil {
			children = append(children, c)
		}
	}
	return html.Div(append([]any{html.Class(wpContentClass)}, children...)...)
}

// maxNodeDepth bounds the recursive HTML walk. x/net/html already caps parse
// depth, but this is defense-in-depth against pathological nesting regardless of
// the parser, and against the renderer's own recursion.
const maxNodeDepth = 256

const wpContentClass = "wp-content"

// --- Gutenberg block comments ---

var gutenbergRe = regexp.MustCompile(`<!--\s*/?\s*wp:([a-zA-Z0-9_-]+(?:/[a-zA-Z0-9_-]+)?)[^>]*-->`)

// stripGutenbergComments removes the <!-- wp:… --> delimiters, leaving the inner
// HTML (which is already valid). Plugin blocks (a "namespace/name" form) are
// recorded as unsupported — their inner HTML still renders, but they were not
// specially handled.
func stripGutenbergComments(s string, postID int64, rep *Report) string {
	return gutenbergRe.ReplaceAllStringFunc(s, func(m string) string {
		if sub := gutenbergRe.FindStringSubmatch(m); len(sub) > 1 && strings.Contains(sub[1], "/") {
			rep.record("block", sub[1], "rendered as inner HTML (block not specially handled)", postID)
		}
		return ""
	})
}

// --- shortcodes ---

var shortcodeRe = regexp.MustCompile(`\[(/?)([a-zA-Z][a-zA-Z0-9_-]*)([^\]]*)\]`)

// processShortcodes strips shortcode delimiters while keeping their inner
// content, honoring the [[escaped]] form. Unknown shortcodes are recorded.
// Plugin shortcodes cannot execute, so their delimiters are removed (never
// emitted as raw HTML) and the inner text is preserved.
func processShortcodes(s string, postID int64, rep *Report) string {
	const lb, rb = "\x00LB\x00", "\x00RB\x00"
	s = strings.ReplaceAll(s, "[[", lb)
	s = strings.ReplaceAll(s, "]]", rb)
	s = shortcodeRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := shortcodeRe.FindStringSubmatch(m)
		if sub[1] == "/" {
			return "" // closing delimiter
		}
		switch sub[2] {
		case "caption", "embed", "gallery", "audio", "video":
			// core media shortcodes: keep the wrapped HTML/inner content
		default:
			rep.record("shortcode", sub[2], "delimiters stripped, inner content kept", postID)
		}
		return ""
	})
	s = strings.ReplaceAll(s, lb, "[")
	s = strings.ReplaceAll(s, rb, "]")
	return s
}

// --- safe HTML walk ---

type walker struct {
	opt    contentOptions
	rep    *Report
	postID int64
}

// dropElements are removed with their entire subtree (scripts, embeds, forms,
// and foreign-content roots that can carry script).
var dropElements = map[string]bool{
	"script": true, "style": true, "iframe": true, "object": true, "embed": true,
	"form": true, "input": true, "button": true, "select": true, "option": true,
	"textarea": true, "link": true, "meta": true, "base": true, "noscript": true,
	"template": true, "applet": true, "frame": true, "frameset": true,
	"svg": true, "math": true, "head": true, "html": true, "body": true, "title": true,
}

var voidElements = map[string]bool{
	"area": true, "br": true, "col": true, "hr": true, "img": true,
	"source": true, "track": true, "wbr": true,
}

var nameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`)

func (w *walker) node(n *xhtml.Node, depth int, inPre bool) mx.Component {
	if depth > maxNodeDepth {
		return nil
	}
	switch n.Type {
	case xhtml.TextNode:
		return mx.Text(n.Data) // escaped by CheckedWriter
	case xhtml.ElementNode:
		tag := strings.ToLower(n.Data)
		if dropElements[tag] {
			w.rep.record("dropped", "<"+tag+">", "removed (disallowed element)", w.postID)
			return nil
		}
		if !nameRe.MatchString(tag) {
			return nil
		}
		nextPre := inPre || tag == "pre" || tag == "code"
		tag = demoteHeading(tag, inPre)
		attribs := w.attribs(n, tag)

		if voidElements[tag] {
			return html.VoidElement(tag, attribs...)
		}
		args := make([]any, 0, len(attribs)+2)
		for _, a := range attribs {
			args = append(args, a)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if comp := w.node(c, depth+1, nextPre); comp != nil {
				args = append(args, comp)
			}
		}
		return html.Element(tag, args...)
	default:
		return nil // comments, doctype, etc.
	}
}

func (w *walker) attribs(n *xhtml.Node, tag string) []mx.Attrib {
	out := make([]mx.Attrib, 0, len(n.Attr))
	seen := make(map[string]bool, len(n.Attr))
	hasTarget := false
	for _, a := range n.Attr {
		name := strings.ToLower(a.Key)
		if !nameRe.MatchString(name) || seen[name] {
			continue // invalid name, or duplicate (CheckedWriter rejects dups)
		}
		if strings.HasPrefix(name, "on") { // event handlers
			w.rep.record("dropped", name+"=", "removed (event handler)", w.postID)
			continue
		}
		if name == "style" || name == "ping" { // inline styles + tracking pings
			continue
		}
		val := a.Val
		switch name {
		case "href", "cite", "action", "formaction":
			safe, ok := w.safeURL(val, false)
			if !ok {
				w.rep.record("blockedURL", schemeOf(val), "removed (disallowed URL scheme)", w.postID)
				continue
			}
			val = safe
		case "src", "poster":
			safe, ok := w.safeURL(val, true)
			if !ok {
				w.rep.record("blockedURL", schemeOf(val), "removed (disallowed URL scheme)", w.postID)
				continue
			}
			val = safe
		case "srcset":
			val = w.safeSrcset(val)
		case "target":
			hasTarget = true
		}
		seen[name] = true
		out = append(out, html.Attrib(name, val))
	}
	if tag == "a" && hasTarget && !seen["rel"] {
		out = append(out, html.Attrib("rel", "noopener noreferrer")) // reverse-tabnabbing guard
	}
	if tag == "img" {
		if !seen["loading"] {
			out = append(out, html.Attrib("loading", "lazy"))
		}
		// Graceful fallback for missing media (a literal handler, not from input).
		out = append(out, html.Attrib("onerror", "this.style.display='none'"))
	}
	return out
}

// safeURL enforces a per-context scheme allowlist (not a blocklist). It returns
// (value, ok); ok=false means the URL must be dropped. For link contexts
// (imageContext=false) only http(s)/mailto/tel and relative/fragment URLs pass.
// For media contexts (imageContext=true) http(s), relative, and data: for raster
// image types pass — never data:image/svg+xml (it can carry script) or any other
// scheme (javascript:, vbscript:, file:, blob:, …). Internal links are rewritten
// to local routes when a resolver is set.
func (w *walker) safeURL(raw string, imageContext bool) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw, true
	}
	// Relative, root-relative, protocol-relative, query- and fragment-only URLs
	// have no scheme and are safe.
	if strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "#") ||
		strings.HasPrefix(trimmed, "?") || strings.HasPrefix(trimmed, "./") ||
		strings.HasPrefix(trimmed, "../") {
		return w.maybeRewrite(trimmed), true
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return "", false
	}
	if u.Scheme == "" {
		return w.maybeRewrite(trimmed), true
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		return w.maybeRewrite(trimmed), true
	case "mailto", "tel":
		return trimmed, !imageContext
	case "data":
		return raw, imageContext && isRasterImageData(trimmed)
	default: // javascript, vbscript, file, blob, and anything else
		return "", false
	}
}

// maybeRewrite turns an internal absolute URL into its local route when a
// resolver is configured; otherwise it returns the URL unchanged.
func (w *walker) maybeRewrite(raw string) string {
	if w.opt.resolve == nil {
		return raw
	}
	if u, err := url.Parse(raw); err == nil && u.Host != "" && hostMatches(u.Host, w.opt.baseHost) {
		if local := w.opt.resolve(raw); local != "" {
			w.rep.RewrittenLinks++
			return local
		}
	}
	return raw
}

// isRasterImageData reports whether a data: URI is a raster image type — the
// only data: form safe to keep. data:image/svg+xml is excluded: SVG can carry
// scripts.
func isRasterImageData(raw string) bool {
	lower := strings.ToLower(raw)
	for _, p := range []string{"data:image/png", "data:image/jpeg", "data:image/jpg", "data:image/gif", "data:image/webp", "data:image/avif", "data:image/bmp"} {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

func (w *walker) safeSrcset(raw string) string {
	parts := strings.Split(raw, ",")
	for i, p := range parts {
		fields := strings.Fields(strings.TrimSpace(p))
		if len(fields) == 0 {
			continue
		}
		if safe, ok := w.safeURL(fields[0], true); ok {
			fields[0] = safe
			parts[i] = " " + strings.Join(fields, " ")
		} else {
			parts[i] = ""
		}
	}
	return strings.TrimLeft(strings.Join(parts, ","), ", ")
}

func demoteHeading(tag string, inPre bool) string {
	if inPre {
		return tag
	}
	switch tag {
	case "h1":
		return "h2"
	case "h2":
		return "h3"
	case "h3":
		return "h4"
	case "h4":
		return "h5"
	case "h5", "h6":
		return "h6"
	}
	return tag
}

func schemeOf(raw string) string {
	if i := strings.IndexByte(raw, ':'); i > 0 {
		return strings.ToLower(strings.TrimSpace(raw[:i])) + ":"
	}
	return raw
}

func hostMatches(host, base string) bool {
	if base == "" {
		return false
	}
	host = strings.ToLower(strings.TrimPrefix(host, "www."))
	base = strings.ToLower(strings.TrimPrefix(base, "www."))
	return host == base
}
