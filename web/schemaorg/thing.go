package schemaorg

// https://schema.org/Thing
// https://www.w3.org/TR/json-ld11/#the-context

// Context is the JSON-LD @context of schema.org, the vocabulary the @type and
// property names of every type in this package belong to. The constructors set
// it so the value they return can be the root of a JSON-LD script.
const Context = "https://schema.org"

// Thing is the most generic schema.org type and the base every other type in
// this package embeds. encoding/json flattens an embedded struct into the
// object of the embedding one, at the position of the embedded field, so these
// properties lead the output, starting with @context and @type.
//
// https://schema.org/Thing
type Thing struct {
	// Context is the JSON-LD @context, [Context] for a value from a
	// constructor. Only the root of a script needs it; see the package
	// documentation for nested values.
	Context string `json:"@context,omitempty"`

	// Type is the schema.org type name, e.g. "Person". It is never omitted:
	// a node without a type tells a crawler nothing about what it describes.
	Type string `json:"@type"`

	// ID is the JSON-LD @id, a URL identifying the node across scripts and
	// pages, usually the URL of the page about it with a fragment such as
	// "https://example.com/#organization". Another value can then refer to
	// the node with a Thing holding just Type and ID instead of repeating it.
	ID string `json:"@id,omitempty"`

	// Name is the name of the item.
	Name string `json:"name,omitempty"`

	// Description is a plain-text description of the item.
	Description string `json:"description,omitempty"`

	// URL is the URL of the item, for a page its canonical URL.
	URL string `json:"url,omitempty"`

	// SameAs are the URLs of other pages that unambiguously identify the item,
	// such as its Wikipedia article or its social media profiles. They are how
	// a crawler connects the item to what it already knows about it.
	SameAs []string `json:"sameAs,omitempty"`

	// Image is the URL of an image of the item.
	Image string `json:"image,omitempty"`
}

// NewThing returns a Thing of the schema.org type typ with the given name. It
// covers types this package has no struct for, as far as the common properties
// describe them — a Country for [Service.AreaServed], a Brand, a Language.
func NewThing(typ, name string) *Thing {
	thing := newThing(typ)
	thing.Name = name
	return &thing
}

// newThing returns the Thing every constructor starts from: the type plus the
// @context that makes the result a valid root.
func newThing(typ string) Thing {
	return Thing{Context: Context, Type: typ}
}
