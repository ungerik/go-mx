package wordpress

import "time"

// ContentHTML is raw WordPress body markup: classic HTML, Gutenberg block
// comments (<!-- wp:… -->), and shortcodes ([gallery], [caption], …). It is NOT
// plain text and must NOT be rendered by escaping it as a string — it goes
// through the content pipeline (parse, sanitize, rewrite) before rendering. The
// distinct type keeps that contract visible at every call site.
type ContentHTML string

// Status is a WordPress post status from <wp:status>.
type Status string

const (
	// StatusPublish marks a published, publicly visible post.
	StatusPublish Status = "publish"
	// StatusDraft marks an unpublished draft.
	StatusDraft Status = "draft"
	// StatusPending marks a post awaiting review before publication.
	StatusPending Status = "pending"
	// StatusPrivate marks a published post visible only to authorized users.
	StatusPrivate Status = "private"
	// StatusFuture marks a post scheduled to publish at a future date.
	StatusFuture Status = "future"
	// StatusInherit marks a post (typically an attachment) that inherits its
	// parent's status.
	StatusInherit Status = "inherit" // attachments inherit their parent's status
	// StatusTrash marks a post moved to the trash.
	StatusTrash Status = "trash"
)

// Site is a whole imported WordPress site as pure, encoding/json-serializable
// data. Relationships are stored as int64 IDs and slugs (see the field docs),
// never as parent↔child pointer cycles, so a Site always marshals to a JSON
// tree. Lookup maps and reverse indexes are built separately for rendering (see
// [Site.index]); they are derived, not stored, and not part of the JSON shape.
type Site struct {
	Title      string `json:"title"`
	Tagline    string `json:"tagline,omitempty"`
	BaseURL    string `json:"baseURL,omitempty"` // canonical site URL, for link rewriting
	Language   string `json:"language,omitempty"`
	WXRVersion string `json:"wxrVersion,omitempty"`

	Authors     []*Author     `json:"authors,omitempty"`
	Posts       []*Post       `json:"posts,omitempty"`
	Pages       []*Page       `json:"pages,omitempty"`
	Categories  []*Term       `json:"categories,omitempty"` // taxonomy "category"
	Tags        []*Term       `json:"tags,omitempty"`       // taxonomy "post_tag"
	Attachments []*Attachment `json:"attachments,omitempty"`
	MenuItems   []*MenuItem   `json:"menuItems,omitempty"`
}

// Author is a WXR <wp:author>. Posts reference their author by Login
// (<dc:creator> carries the login, not the numeric ID), so resolve via login.
type Author struct {
	ID          int64  `json:"id,omitempty"`
	Login       string `json:"login,omitempty"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
}

// Post is a WXR item with post_type "post".
type Post struct {
	ID      int64       `json:"id"`
	Title   string      `json:"title"`
	Slug    string      `json:"slug,omitempty"`
	Excerpt string      `json:"excerpt,omitempty"`
	Content ContentHTML `json:"content,omitempty"`

	AuthorLogin string `json:"authorLogin,omitempty"` // resolve against Site.Authors by Login
	Status      Status `json:"status,omitempty"`

	Published time.Time `json:"published,omitzero"`
	Modified  time.Time `json:"modified,omitzero"`

	Link string `json:"link,omitempty"`
	GUID string `json:"guid,omitempty"`

	CategorySlugs []string `json:"categorySlugs,omitempty"` // resolve against Site.Categories by Slug
	TagSlugs      []string `json:"tagSlugs,omitempty"`      // resolve against Site.Tags by Slug

	FeaturedImageID int64 `json:"featuredImageID,omitempty"` // _thumbnail_id; 0 = none

	Comments []*Comment `json:"comments,omitempty"`
	Meta     []Meta     `json:"meta,omitempty"`
}

// Page is a WXR item with post_type "page". Pages form a tree via ParentID
// (<wp:post_parent>); 0 means top level.
type Page struct {
	ID       int64       `json:"id"`
	Title    string      `json:"title"`
	Slug     string      `json:"slug,omitempty"`
	Content  ContentHTML `json:"content,omitempty"`
	ParentID int64       `json:"parentID,omitempty"`
	Order    int         `json:"order,omitempty"` // <wp:menu_order>

	AuthorLogin string    `json:"authorLogin,omitempty"`
	Status      Status    `json:"status,omitempty"`
	Published   time.Time `json:"published,omitzero"`
	Modified    time.Time `json:"modified,omitzero"`

	Link string `json:"link,omitempty"`
	GUID string `json:"guid,omitempty"`

	FeaturedImageID int64  `json:"featuredImageID,omitempty"`
	Meta            []Meta `json:"meta,omitempty"`
}

// Term is a taxonomy term: a category (taxonomy "category"), a tag ("post_tag"),
// or a nav menu ("nav_menu"). Categories form a tree via ParentSlug.
type Term struct {
	ID          int64  `json:"id,omitempty"`
	Taxonomy    string `json:"taxonomy,omitempty"`
	Slug        string `json:"slug"` // the stable key terms are referenced by
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ParentSlug  string `json:"parentSlug,omitempty"` // categories only; "" = top level
}

// Comment is a WXR <wp:comment>. Comments thread via ParentID
// (<wp:comment_parent>); 0 means a top-level comment.
type Comment struct {
	ID       int64     `json:"id"`
	Author   string    `json:"author,omitempty"`
	Email    string    `json:"email,omitempty"`
	URL      string    `json:"url,omitempty"`
	Date     time.Time `json:"date,omitzero"`
	Content  string    `json:"content,omitempty"` // comment bodies are plain-ish text, escaped on render
	Approved bool      `json:"approved,omitempty"`
	ParentID int64     `json:"parentID,omitempty"`
	Type     string    `json:"type,omitempty"` // "" for a user comment, "pingback"/"trackback" otherwise
}

// Attachment is a WXR item with post_type "attachment" (media). The binary is
// not in the export; URL is the original absolute location.
type Attachment struct {
	ID    int64  `json:"id"`
	Title string `json:"title,omitempty"`
	Slug  string `json:"slug,omitempty"`
	URL   string `json:"url,omitempty"` // <wp:attachment_url>
	Alt   string `json:"alt,omitempty"` // _wp_attachment_image_alt
	Meta  []Meta `json:"meta,omitempty"`
}

// MenuItem is a WXR item with post_type "nav_menu_item". Items belong to the
// menu named by MenuSlug (the nav_menu taxonomy term) and nest via ParentItemID.
type MenuItem struct {
	ID           int64  `json:"id"`
	Title        string `json:"title,omitempty"`
	Order        int    `json:"order,omitempty"`
	MenuSlug     string `json:"menuSlug,omitempty"`     // which nav_menu this belongs to
	ParentItemID int64  `json:"parentItemID,omitempty"` // _menu_item_menu_item_parent
	Type         string `json:"type,omitempty"`         // post_type | taxonomy | custom
	ObjectID     int64  `json:"objectID,omitempty"`     // _menu_item_object_id
	Object       string `json:"object,omitempty"`       // page | category | custom | …
	URL          string `json:"url,omitempty"`          // _menu_item_url (custom links)
}

// Meta is one <wp:postmeta> key/value pair. Some values are PHP-serialized; the
// content pipeline decodes those it needs in a later step.
type Meta struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}
