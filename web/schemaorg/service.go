package schemaorg

// https://schema.org/Service

// Service is a service offered by an organization or a person, such as
// consulting, a training or a repair.
//
// https://schema.org/Service
type Service struct {
	Thing

	// ServiceType is the kind of service, e.g. "Software consulting".
	ServiceType string `json:"serviceType,omitempty"`

	// Provider is who provides the service: an [Organization] or a [Person].
	Provider any `json:"provider,omitempty"`

	// AreaServed is where the service is available: a [Place], a [Thing] of
	// an administrative area type ("Country", "City", …), the name of the
	// area as text, or a slice of those.
	AreaServed any `json:"areaServed,omitempty"`

	// Offers are the terms the service is offered on: an [Offer] or a slice
	// of them.
	Offers any `json:"offers,omitempty"`

	// Category is a category of the service, as text or the URL of a
	// category in some taxonomy.
	Category string `json:"category,omitempty"`
}

// NewService returns a Service with the given name.
func NewService(name string) *Service {
	return &Service{Thing: newThing("Service", name)}
}
