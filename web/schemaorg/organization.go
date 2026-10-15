package schemaorg

// https://schema.org/Organization
// https://developers.google.com/search/docs/appearance/structured-data/organization

// Organization is a company, association, institution or any other
// organization. Published on the home or about page of a site it is what a
// search engine shows as the entity behind the site, with its logo.
//
// https://schema.org/Organization
type Organization struct {
	Thing

	// LegalName is the registered name, when it differs from Name — the
	// "Example GmbH" behind the brand "Example".
	LegalName string `json:"legalName,omitempty"`

	// Logo is the URL of the organization's logo. Google requires an image
	// of at least 112x112 pixels that looks right on a white background.
	Logo string `json:"logo,omitempty"`

	// Email is the contact email address, without a "mailto:" prefix.
	Email string `json:"email,omitempty"`

	// Telephone is the contact phone number, best in international format
	// ("+43 1 234567").
	Telephone string `json:"telephone,omitempty"`

	// Address is the postal address of the organization.
	Address *PostalAddress `json:"address,omitempty"`

	// VATID is the value-added tax identification number, e.g.
	// "ATU12345678".
	VATID string `json:"vatID,omitempty"`
}

// NewOrganization returns an Organization with the given name.
func NewOrganization(name string) *Organization {
	return &Organization{Thing: newThing("Organization", name)}
}
