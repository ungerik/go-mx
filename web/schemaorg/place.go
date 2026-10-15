package schemaorg

// https://schema.org/Place
// https://schema.org/VirtualLocation
// https://schema.org/PostalAddress

// Place is a physical location, for example the venue of an [Event].
//
// https://schema.org/Place
type Place struct {
	Thing

	// Address is the postal address of the place. Google requires it for
	// the location of an in-person event.
	Address *PostalAddress `json:"address,omitempty"`
}

// NewPlace returns a Place with the given name, e.g. the name of a venue.
func NewPlace(name string) *Place {
	place := &Place{Thing: newThing("Place")}
	place.Name = name
	return place
}

// VirtualLocation is the online location of an [Event], the page or stream
// where it is attended. It has no properties beyond those of [Thing], of which
// URL is the one that matters.
//
// https://schema.org/VirtualLocation
type VirtualLocation struct {
	Thing
}

// NewVirtualLocation returns a VirtualLocation with the URL where the event is
// attended.
func NewVirtualLocation(url string) *VirtualLocation {
	location := &VirtualLocation{Thing: newThing("VirtualLocation")}
	location.URL = url
	return location
}

// PostalAddress is a mailing address. Its parts are separate properties, not
// one formatted string, so a crawler does not have to guess the order a
// country writes them in.
//
// https://schema.org/PostalAddress
type PostalAddress struct {
	Thing

	// StreetAddress is the street and house number, e.g. "Mariahilfer
	// Straße 1".
	StreetAddress string `json:"streetAddress,omitempty"`

	// AddressLocality is the city or town, e.g. "Vienna".
	AddressLocality string `json:"addressLocality,omitempty"`

	// AddressRegion is the state, province or region, e.g. "Vienna" or
	// "CA".
	AddressRegion string `json:"addressRegion,omitempty"`

	// PostalCode is the postal code, e.g. "1060".
	PostalCode string `json:"postalCode,omitempty"`

	// AddressCountry is the country as a two-letter ISO 3166-1 alpha-2 code,
	// e.g. "AT", the form Google recommends over a country name.
	AddressCountry string `json:"addressCountry,omitempty"`
}

// NewPostalAddress returns an empty PostalAddress to fill in. It takes no
// arguments because no single part of an address is required, and several
// positional strings would be too easy to mix up.
func NewPostalAddress() *PostalAddress {
	return &PostalAddress{Thing: newThing("PostalAddress")}
}
