package wordpress

import (
	"bufio"
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/domonda/go-errs"
)

// WXR namespace URIs. encoding/xml matches a namespaced element by the
// "<space> <local>" tag form; matching only the local name would collide
// content:encoded (nsContent) with excerpt:encoded (nsExcerpt) — both local
// name "encoded". The raw struct tags below spell the URIs out (tags must be
// string literals); nsWP is also used directly in the channel token switch.
//
//	nsContent = "http://purl.org/rss/1.0/modules/content/"
//	nsExcerpt = "http://wordpress.org/export/1.2/excerpt/"
//	nsDC      = "http://purl.org/dc/elements/1.1/"
const nsWP = "http://wordpress.org/export/1.2/"

// parseWXR streams one WXR document from r into site, recording diagnostics in
// rep. It uses a token loop with DecodeElement per <item> rather than
// xml.Unmarshal of the whole tree, so peak memory is the growing model, not
// model + full DOM — required for the multi-GB exports WordPress produces.
func parseWXR(r io.Reader, site *Site, rep *Report) error {
	dec := xml.NewDecoder(r)
	dec.CharsetReader = charsetReader

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errs.Errorf("reading WXR: %w (the export may be truncated or malformed; re-export via WordPress → Tools → Export)", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		switch {
		case start.Name.Local == "image" || start.Name.Local == "rss" || start.Name.Local == "channel":
			// <rss>/<channel> are containers we descend into; <image> is the
			// site-icon block whose nested <title> would shadow the channel
			// title, so skip its whole subtree.
			if start.Name.Local == "image" {
				if err := dec.Skip(); err != nil {
					return errs.Errorf("reading WXR: %w", err)
				}
			}

		case start.Name.Space == nsWP && start.Name.Local == "wxr_version":
			site.WXRVersion = decodeText(dec, &start)
			rep.WXRVersion = site.WXRVersion
		case start.Name.Space == nsWP && start.Name.Local == "base_site_url":
			if site.BaseURL == "" {
				site.BaseURL = decodeText(dec, &start)
			}
		case start.Name.Space == nsWP && start.Name.Local == "base_blog_url":
			if site.BaseURL == "" {
				site.BaseURL = decodeText(dec, &start)
			}

		case start.Name.Space == "" && start.Name.Local == "title":
			if site.Title == "" {
				site.Title = decodeText(dec, &start)
			}
		case start.Name.Space == "" && start.Name.Local == "description":
			if site.Tagline == "" {
				site.Tagline = decodeText(dec, &start)
			}
		case start.Name.Space == "" && start.Name.Local == "language":
			if site.Language == "" {
				site.Language = decodeText(dec, &start)
			}

		case start.Name.Space == nsWP && start.Name.Local == "author":
			var a rawAuthor
			if err := dec.DecodeElement(&a, &start); err != nil {
				return errs.Errorf("decoding wp:author: %w", err)
			}
			site.Authors = append(site.Authors, a.toAuthor())

		case start.Name.Space == nsWP && start.Name.Local == "category":
			var c rawCategory
			if err := dec.DecodeElement(&c, &start); err != nil {
				return errs.Errorf("decoding wp:category: %w", err)
			}
			site.Categories = append(site.Categories, c.toTerm())
		case start.Name.Space == nsWP && start.Name.Local == "tag":
			var t rawTag
			if err := dec.DecodeElement(&t, &start); err != nil {
				return errs.Errorf("decoding wp:tag: %w", err)
			}
			site.Tags = append(site.Tags, t.toTerm())
		case start.Name.Space == nsWP && start.Name.Local == "term":
			var t rawTerm
			if err := dec.DecodeElement(&t, &start); err != nil {
				return errs.Errorf("decoding wp:term: %w", err)
			}
			t.apply(site)

		case start.Name.Local == "item":
			var it rawItem
			if err := dec.DecodeElement(&it, &start); err != nil {
				// A syntax error inside an <item> leaves the decoder mis-positioned,
				// so we cannot reliably resume — report it as fatal and actionable
				// rather than pretending to skip just this item.
				return errs.Errorf("malformed <item> in WXR: %w (re-export the file from WordPress → Tools → Export)", err)
			}
			it.apply(site, rep)
		}
	}
}

// charsetReader lets the decoder read the common non-UTF-8 WXR encodings old
// WordPress installs emit, without a golang.org/x/text dependency: UTF-8/ASCII
// pass through, ISO-8859-1 maps each byte to the same code point, and
// Windows-1252 additionally maps the 0x80–0x9F range (smart quotes, dashes,
// euro, …) via a lookup table. An unrecognized declared charset is a clear,
// actionable error.
func charsetReader(label string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "", "utf-8", "utf8", "us-ascii", "ascii":
		return input, nil
	case "iso-8859-1", "iso8859-1", "latin1", "latin-1":
		return &byteReader{src: bufio.NewReader(input), mapByte: latin1Byte}, nil
	case "windows-1252", "cp1252":
		return &byteReader{src: bufio.NewReader(input), mapByte: cp1252Byte}, nil
	default:
		return nil, errs.Errorf("unsupported XML charset %q declared in the WXR export; re-export it as UTF-8", label)
	}
}

// byteReader converts a single-byte encoding to UTF-8 on the fly using mapByte.
type byteReader struct {
	src     io.Reader
	mapByte func(byte) rune
	pend    []byte
}

func (r *byteReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil // io.Reader contract: consume no input on a zero-length buffer
	}
	if len(r.pend) > 0 {
		n := copy(p, r.pend)
		r.pend = r.pend[n:]
		return n, nil
	}
	var b [1]byte
	if _, err := io.ReadFull(r.src, b[:]); err != nil {
		return 0, err
	}
	var enc [utf8.UTFMax]byte
	m := utf8.EncodeRune(enc[:], r.mapByte(b[0]))
	n := copy(p, enc[:m])
	if n < m {
		r.pend = append(r.pend, enc[n:m]...)
	}
	return n, nil
}

func latin1Byte(b byte) rune { return rune(b) }

func cp1252Byte(b byte) rune {
	if b >= 0x80 && b <= 0x9f {
		return cp1252Hi[b-0x80]
	}
	return rune(b)
}

// cp1252Hi maps the Windows-1252 bytes 0x80–0x9F, the range where it differs
// from ISO-8859-1. Unassigned slots (0x81, 0x8D, 0x8F, 0x90, 0x9D) map to the
// byte value.
var cp1252Hi = [32]rune{
	0x20AC, 0x81, 0x201A, 0x0192, 0x201E, 0x2026, 0x2020, 0x2021,
	0x02C6, 0x2030, 0x0160, 0x2039, 0x0152, 0x8D, 0x017D, 0x8F,
	0x90, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
	0x02DC, 0x2122, 0x0161, 0x203A, 0x0153, 0x9D, 0x017E, 0x0178,
}

// decodeText reads the character data of the current element into a trimmed
// string. CDATA and plain text both arrive as char data, so this handles
// <![CDATA[…]]>-wrapped titles and slugs transparently.
func decodeText(dec *xml.Decoder, start *xml.StartElement) string {
	var s string
	if err := dec.DecodeElement(&s, start); err != nil {
		return ""
	}
	return strings.TrimSpace(s)
}

// --- raw WXR shapes (namespace-qualified) mapped to the model ---

type rawAuthor struct {
	ID      int64  `xml:"http://wordpress.org/export/1.2/ author_id"`
	Login   string `xml:"http://wordpress.org/export/1.2/ author_login"`
	Email   string `xml:"http://wordpress.org/export/1.2/ author_email"`
	Display string `xml:"http://wordpress.org/export/1.2/ author_display_name"`
	First   string `xml:"http://wordpress.org/export/1.2/ author_first_name"`
	Last    string `xml:"http://wordpress.org/export/1.2/ author_last_name"`
}

func (a rawAuthor) toAuthor() *Author {
	return &Author{
		ID:          a.ID,
		Login:       strings.TrimSpace(a.Login),
		Email:       strings.TrimSpace(a.Email),
		DisplayName: strings.TrimSpace(a.Display),
		FirstName:   strings.TrimSpace(a.First),
		LastName:    strings.TrimSpace(a.Last),
	}
}

type rawCategory struct {
	TermID   int64  `xml:"http://wordpress.org/export/1.2/ term_id"`
	Nicename string `xml:"http://wordpress.org/export/1.2/ category_nicename"`
	Parent   string `xml:"http://wordpress.org/export/1.2/ category_parent"`
	Name     string `xml:"http://wordpress.org/export/1.2/ cat_name"`
	Desc     string `xml:"http://wordpress.org/export/1.2/ category_description"`
}

func (c rawCategory) toTerm() *Term {
	return &Term{
		ID:          c.TermID,
		Taxonomy:    "category",
		Slug:        strings.TrimSpace(c.Nicename),
		Name:        strings.TrimSpace(c.Name),
		Description: strings.TrimSpace(c.Desc),
		ParentSlug:  strings.TrimSpace(c.Parent),
	}
}

type rawTag struct {
	TermID int64  `xml:"http://wordpress.org/export/1.2/ term_id"`
	Slug   string `xml:"http://wordpress.org/export/1.2/ tag_slug"`
	Name   string `xml:"http://wordpress.org/export/1.2/ tag_name"`
	Desc   string `xml:"http://wordpress.org/export/1.2/ tag_description"`
}

func (t rawTag) toTerm() *Term {
	return &Term{
		ID:          t.TermID,
		Taxonomy:    "post_tag",
		Slug:        strings.TrimSpace(t.Slug),
		Name:        strings.TrimSpace(t.Name),
		Description: strings.TrimSpace(t.Desc),
	}
}

// rawTerm is the generic <wp:term> form (any taxonomy, including nav_menu).
type rawTerm struct {
	TermID   int64  `xml:"http://wordpress.org/export/1.2/ term_id"`
	Taxonomy string `xml:"http://wordpress.org/export/1.2/ term_taxonomy"`
	Slug     string `xml:"http://wordpress.org/export/1.2/ term_slug"`
	Name     string `xml:"http://wordpress.org/export/1.2/ term_name"`
	Parent   string `xml:"http://wordpress.org/export/1.2/ term_parent"`
	Desc     string `xml:"http://wordpress.org/export/1.2/ term_description"`
}

func (t rawTerm) apply(site *Site) {
	term := &Term{
		ID:          t.TermID,
		Taxonomy:    strings.TrimSpace(t.Taxonomy),
		Slug:        strings.TrimSpace(t.Slug),
		Name:        strings.TrimSpace(t.Name),
		Description: strings.TrimSpace(t.Desc),
		ParentSlug:  strings.TrimSpace(t.Parent),
	}
	switch term.Taxonomy {
	case "category":
		site.Categories = append(site.Categories, term)
	case "post_tag":
		site.Tags = append(site.Tags, term)
		// nav_menu and other taxonomies are not materialized as Terms in v1.
	}
}

type rawCatRef struct {
	Domain   string `xml:"domain,attr"`
	Nicename string `xml:"nicename,attr"`
	Name     string `xml:",chardata"`
}

type rawPostmeta struct {
	Key   string `xml:"http://wordpress.org/export/1.2/ meta_key"`
	Value string `xml:"http://wordpress.org/export/1.2/ meta_value"`
}

type rawComment struct {
	ID       int64  `xml:"http://wordpress.org/export/1.2/ comment_id"`
	Author   string `xml:"http://wordpress.org/export/1.2/ comment_author"`
	Email    string `xml:"http://wordpress.org/export/1.2/ comment_author_email"`
	URL      string `xml:"http://wordpress.org/export/1.2/ comment_author_url"`
	Date     string `xml:"http://wordpress.org/export/1.2/ comment_date_gmt"`
	Content  string `xml:"http://wordpress.org/export/1.2/ comment_content"`
	Approved string `xml:"http://wordpress.org/export/1.2/ comment_approved"`
	Parent   int64  `xml:"http://wordpress.org/export/1.2/ comment_parent"`
	Type     string `xml:"http://wordpress.org/export/1.2/ comment_type"`
}

func (c rawComment) toComment() *Comment {
	return &Comment{
		ID:       c.ID,
		Author:   strings.TrimSpace(c.Author),
		Email:    strings.TrimSpace(c.Email),
		URL:      strings.TrimSpace(c.URL),
		Date:     parseWPTime(c.Date),
		Content:  strings.TrimSpace(c.Content),
		Approved: strings.TrimSpace(c.Approved) == "1",
		ParentID: c.Parent,
		Type:     strings.TrimSpace(c.Type),
	}
}

type rawItem struct {
	Title         string        `xml:"title"`
	Link          string        `xml:"link"`
	PubDate       string        `xml:"pubDate"`
	Creator       string        `xml:"http://purl.org/dc/elements/1.1/ creator"`
	GUID          string        `xml:"guid"`
	Content       string        `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	Excerpt       string        `xml:"http://wordpress.org/export/1.2/excerpt/ encoded"`
	PostID        int64         `xml:"http://wordpress.org/export/1.2/ post_id"`
	PostDateGMT   string        `xml:"http://wordpress.org/export/1.2/ post_date_gmt"`
	PostModified  string        `xml:"http://wordpress.org/export/1.2/ post_modified_gmt"`
	PostName      string        `xml:"http://wordpress.org/export/1.2/ post_name"`
	Status        string        `xml:"http://wordpress.org/export/1.2/ status"`
	PostType      string        `xml:"http://wordpress.org/export/1.2/ post_type"`
	PostParent    int64         `xml:"http://wordpress.org/export/1.2/ post_parent"`
	MenuOrder     int           `xml:"http://wordpress.org/export/1.2/ menu_order"`
	AttachmentURL string        `xml:"http://wordpress.org/export/1.2/ attachment_url"`
	Categories    []rawCatRef   `xml:"category"`
	Postmeta      []rawPostmeta `xml:"http://wordpress.org/export/1.2/ postmeta"`
	Comments      []rawComment  `xml:"http://wordpress.org/export/1.2/ comment"`
}

// apply maps a decoded item onto the model by post_type.
func (it *rawItem) apply(site *Site, rep *Report) {
	switch it.PostType {
	case "post":
		site.Posts = append(site.Posts, it.toPost())
	case "page":
		site.Pages = append(site.Pages, it.toPage())
	case "attachment":
		site.Attachments = append(site.Attachments, it.toAttachment())
	case "nav_menu_item":
		site.MenuItems = append(site.MenuItems, it.toMenuItem())
	case "", "revision", "nav_menu", "custom_css", "customize_changeset", "wp_global_styles", "oembed_cache", "user_request", "wp_block":
		// internal/non-content types: drop without noise
	default:
		// custom post types are not rendered in v1 — flag, never silently drop
		rep.SkippedItems = append(rep.SkippedItems, SkippedItem{
			PostID:   it.PostID,
			PostType: it.PostType,
			Title:    strings.TrimSpace(it.Title),
			Reason:   "unsupported post_type",
		})
		rep.Counts.Skipped++
	}
}

func (it *rawItem) meta() []Meta {
	if len(it.Postmeta) == 0 {
		return nil
	}
	m := make([]Meta, 0, len(it.Postmeta))
	for _, pm := range it.Postmeta {
		m = append(m, Meta{Key: pm.Key, Value: pm.Value})
	}
	return m
}

func (it *rawItem) metaValue(key string) string {
	for _, pm := range it.Postmeta {
		if pm.Key == key {
			return strings.TrimSpace(pm.Value)
		}
	}
	return ""
}

func (it *rawItem) comments() []*Comment {
	if len(it.Comments) == 0 {
		return nil
	}
	cs := make([]*Comment, 0, len(it.Comments))
	for i := range it.Comments {
		cs = append(cs, it.Comments[i].toComment())
	}
	return cs
}

func (it *rawItem) taxonomy() (categories, tags []string) {
	for _, ref := range it.Categories {
		slug := strings.TrimSpace(ref.Nicename)
		if slug == "" {
			continue
		}
		switch ref.Domain {
		case "category":
			categories = append(categories, slug)
		case "post_tag":
			tags = append(tags, slug)
		}
	}
	return categories, tags
}

func (it *rawItem) toPost() *Post {
	cats, tags := it.taxonomy()
	return &Post{
		ID:              it.PostID,
		Title:           strings.TrimSpace(it.Title),
		Slug:            strings.TrimSpace(it.PostName),
		Excerpt:         strings.TrimSpace(it.Excerpt),
		Content:         ContentHTML(it.Content),
		AuthorLogin:     strings.TrimSpace(it.Creator),
		Status:          Status(strings.TrimSpace(it.Status)),
		Published:       parseWPTime(it.PostDateGMT),
		Modified:        parseWPTime(it.PostModified),
		Link:            strings.TrimSpace(it.Link),
		GUID:            strings.TrimSpace(it.GUID),
		CategorySlugs:   cats,
		TagSlugs:        tags,
		FeaturedImageID: parseID(it.metaValue("_thumbnail_id")),
		Comments:        it.comments(),
		Meta:            it.meta(),
	}
}

func (it *rawItem) toPage() *Page {
	return &Page{
		ID:              it.PostID,
		Title:           strings.TrimSpace(it.Title),
		Slug:            strings.TrimSpace(it.PostName),
		Content:         ContentHTML(it.Content),
		ParentID:        it.PostParent,
		Order:           it.MenuOrder,
		AuthorLogin:     strings.TrimSpace(it.Creator),
		Status:          Status(strings.TrimSpace(it.Status)),
		Published:       parseWPTime(it.PostDateGMT),
		Modified:        parseWPTime(it.PostModified),
		Link:            strings.TrimSpace(it.Link),
		GUID:            strings.TrimSpace(it.GUID),
		FeaturedImageID: parseID(it.metaValue("_thumbnail_id")),
		Meta:            it.meta(),
	}
}

func (it *rawItem) toAttachment() *Attachment {
	return &Attachment{
		ID:    it.PostID,
		Title: strings.TrimSpace(it.Title),
		Slug:  strings.TrimSpace(it.PostName),
		URL:   strings.TrimSpace(it.AttachmentURL),
		Alt:   it.metaValue("_wp_attachment_image_alt"),
		Meta:  it.meta(),
	}
}

func (it *rawItem) toMenuItem() *MenuItem {
	mi := &MenuItem{
		ID:           it.PostID,
		Title:        strings.TrimSpace(it.Title),
		Order:        it.MenuOrder,
		ParentItemID: parseID(it.metaValue("_menu_item_menu_item_parent")),
		Type:         it.metaValue("_menu_item_type"),
		ObjectID:     parseID(it.metaValue("_menu_item_object_id")),
		Object:       it.metaValue("_menu_item_object"),
		URL:          it.metaValue("_menu_item_url"),
	}
	for _, ref := range it.Categories {
		if ref.Domain == "nav_menu" {
			mi.MenuSlug = strings.TrimSpace(ref.Nicename)
			break
		}
	}
	return mi
}

// parseWPTime parses the "2006-01-02 15:04:05" GMT timestamp WXR uses for
// wp:post_date_gmt etc. A zero/"0000-00-00 00:00:00"/unparseable value yields
// the zero Time.
func parseWPTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "0000-00-00") {
		return time.Time{}
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse(time.RFC1123Z, s); err == nil { // pubDate fallback
		return t.UTC()
	}
	return time.Time{}
}

func parseID(s string) int64 {
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0
	}
	return n
}
