package schemaorg

// https://schema.org/ItemList
// https://developers.google.com/search/docs/appearance/structured-data/carousel
// https://developers.google.com/search/docs/appearance/structured-data/breadcrumb

// ItemList is an ordered list of items, such as the events on a calendar page
// or the posts of an archive. With Type set to "BreadcrumbList" the same
// struct describes a breadcrumb trail, a ListItem per level holding its name
// and URL.
//
// https://schema.org/ItemList
type ItemList struct {
	Thing

	// ItemListElement are the entries of the list.
	ItemListElement []*ListItem `json:"itemListElement,omitempty"`

	// NumberOfItems is the total number of items, which may be more than the
	// entries of a paginated list. Zero omits it.
	NumberOfItems int `json:"numberOfItems,omitempty"`
}

// NewItemList returns an ItemList with the given entries.
func NewItemList(items ...*ListItem) *ItemList {
	return &ItemList{Thing: newThing("ItemList", ""), ItemListElement: items}
}

// ListItem is one entry of an [ItemList]. Its position is explicit rather than
// implied by the order of the list because JSON-LD does not preserve the order
// of a property's values.
//
// https://schema.org/ListItem
type ListItem struct {
	Thing

	// Position is the 1-based position of the entry in the list. Zero omits
	// it, so a list must start at 1.
	Position int `json:"position,omitempty"`

	// Item is the listed item: any schema.org type ([Event], [WebPage], …),
	// or its URL as text, as in a breadcrumb trail.
	Item any `json:"item,omitempty"`
}

// NewListItem returns a ListItem for item at the 1-based position.
func NewListItem(position int, item any) *ListItem {
	return &ListItem{Thing: newThing("ListItem", ""), Position: position, Item: item}
}
