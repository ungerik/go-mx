package schemaorg_test

import (
	"testing"
	"time"

	"github.com/ungerik/go-mx/web/schemaorg"
)

func TestItemList(t *testing.T) {
	requireJSON(t, `{"@context":"https://schema.org","@type":"ItemList"}`, schemaorg.NewItemList())

	first := schemaorg.NewEvent("Autumn Concert", time.Date(2026, 10, 15, 19, 30, 0, 0, cest))
	first.Context = ""
	first.URL = "https://example.com/events/autumn-concert"
	second := schemaorg.NewEvent("Winter Talk", time.Date(2026, 12, 3, 18, 0, 0, 0, cest))
	second.Context = ""
	second.URL = "https://example.com/events/winter-talk"

	list := schemaorg.NewItemList(
		schemaorg.NewListItem(1, first),
		schemaorg.NewListItem(2, second),
	)
	list.Name = "Upcoming events"
	list.NumberOfItems = 2
	requireGolden(t, "itemlist", list)
}

func TestItemListBreadcrumbs(t *testing.T) {
	// A BreadcrumbList is an ItemList subtype without properties of its own,
	// so the ItemList struct with the type overwritten describes it.
	home := schemaorg.NewListItem(1, "https://example.com/")
	home.Context = ""
	home.Name = "Home"
	events := schemaorg.NewListItem(2, "https://example.com/events/")
	events.Context = ""
	events.Name = "Events"

	list := schemaorg.NewItemList(home, events)
	list.Type = "BreadcrumbList"
	requireGolden(t, "breadcrumblist", list)
}

func TestListItem(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"ListItem","position":1,"item":"https://example.com/"}`,
		schemaorg.NewListItem(1, "https://example.com/"),
	)
	// Position zero and a nil item are left out, not rendered as 0 and null.
	requireJSON(t, `{"@context":"https://schema.org","@type":"ListItem"}`, schemaorg.NewListItem(0, nil))
}
