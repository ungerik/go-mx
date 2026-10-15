package schemaorg

// https://schema.org/Person
// https://developers.google.com/search/docs/appearance/structured-data/profile-page

// Person is a person, for example the author of a page or the subject of a
// profile or about page.
//
// https://schema.org/Person
type Person struct {
	Thing

	// GivenName is the first name, which together with FamilyName spells out
	// a Name that alone would be ambiguous about its parts.
	GivenName string `json:"givenName,omitempty"`

	// FamilyName is the last name.
	FamilyName string `json:"familyName,omitempty"`

	// JobTitle is the person's job title, e.g. "Software Engineer".
	JobTitle string `json:"jobTitle,omitempty"`

	// Email is the email address, without a "mailto:" prefix.
	Email string `json:"email,omitempty"`

	// Telephone is the phone number, best in international format
	// ("+43 1 234567").
	Telephone string `json:"telephone,omitempty"`

	// WorksFor is the organization the person works for.
	WorksFor *Organization `json:"worksFor,omitempty"`

	// Address is the person's postal address.
	Address *PostalAddress `json:"address,omitempty"`

	// KnowsLanguage are the languages the person speaks, as BCP 47 language
	// tags ("en", "de-AT").
	KnowsLanguage []string `json:"knowsLanguage,omitempty"`
}

// NewPerson returns a Person with the given full name.
func NewPerson(name string) *Person {
	return &Person{Thing: newThing("Person", name)}
}
