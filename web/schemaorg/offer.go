//go:generate go -C ../../tools tool go-enum ../web/schemaorg/$GOFILE

package schemaorg

// https://schema.org/Offer
// https://schema.org/ItemAvailability

import (
	"fmt"
	"strconv"
)

// Offer is the terms something is offered on: a ticket for an [Event], the
// price of a [Service].
//
// https://schema.org/Offer
type Offer struct {
	Thing

	// Price is the amount as a decimal string built with [Price], without a
	// currency symbol or thousands separators.
	Price string `json:"price,omitempty"`

	// PriceCurrency is the currency of Price as a three-letter ISO 4217 code,
	// e.g. "EUR".
	PriceCurrency string `json:"priceCurrency,omitempty"`

	// PriceValidUntil is the date after which Price is no longer valid, built
	// with [Date]. Search engines stop showing a price past it.
	PriceValidUntil string `json:"priceValidUntil,omitempty"`

	// Availability tells whether the offer can currently be taken.
	Availability Availability `json:"availability,omitempty"`

	// ValidFrom is when the offer becomes available, e.g. when ticket sales
	// start, built with [Date] or [DateTime].
	ValidFrom string `json:"validFrom,omitempty"`

	// ValidThrough is when the offer ends, built with [Date] or [DateTime].
	ValidThrough string `json:"validThrough,omitempty"`

	// Seller is who sells the offer: an [Organization] or a [Person].
	Seller any `json:"seller,omitempty"`
}

// NewOffer returns an Offer at the given price, built with [Price], in the
// currency with the given ISO 4217 code: NewOffer(Price(18000), "EUR").
func NewOffer(price, currency string) *Offer {
	return &Offer{Thing: newThing("Offer"), Price: price, PriceCurrency: currency}
}

// Availability is the availability of an [Offer]. The values are the full
// schema.org URLs, the form Google's documentation uses. They are the subset
// of schema.org's ItemAvailability that applies to tickets and services; a
// value outside it can still be converted, Availability(
// "https://schema.org/BackOrder"), but is not Valid.
type Availability string //#enum

const (
	// AvailabilityInStock means the offer can be taken.
	AvailabilityInStock Availability = "https://schema.org/InStock"
	// AvailabilityLimited means only a few are left.
	AvailabilityLimited Availability = "https://schema.org/LimitedAvailability"
	// AvailabilitySoldOut means the offer sold out, the usual state of the
	// tickets of a full event.
	AvailabilitySoldOut Availability = "https://schema.org/SoldOut"
	// AvailabilityPreOrder means the offer can be ordered before it becomes
	// available, e.g. tickets before the date sales open.
	AvailabilityPreOrder Availability = "https://schema.org/PreOrder"
	// AvailabilityDiscontinued means the offer is no longer made.
	AvailabilityDiscontinued Availability = "https://schema.org/Discontinued"
	// AvailabilityOutOfStock means the offer cannot currently be taken.
	AvailabilityOutOfStock Availability = "https://schema.org/OutOfStock"
)

// Valid indicates if a is any of the valid values for Availability
func (a Availability) Valid() bool {
	switch a {
	case
		AvailabilityInStock,
		AvailabilityLimited,
		AvailabilitySoldOut,
		AvailabilityPreOrder,
		AvailabilityDiscontinued,
		AvailabilityOutOfStock:
		return true
	}
	return false
}

// Validate returns an error if a is none of the valid values for Availability
func (a Availability) Validate() error {
	if !a.Valid() {
		return fmt.Errorf("invalid value %#v for type schemaorg.Availability", a)
	}
	return nil
}

// Enums returns all valid values for Availability
func (Availability) Enums() []Availability {
	return []Availability{
		AvailabilityInStock,
		AvailabilityLimited,
		AvailabilitySoldOut,
		AvailabilityPreOrder,
		AvailabilityDiscontinued,
		AvailabilityOutOfStock,
	}
}

// EnumStrings returns all valid values for Availability as strings
func (Availability) EnumStrings() []string {
	return []string{
		"https://schema.org/InStock",
		"https://schema.org/LimitedAvailability",
		"https://schema.org/SoldOut",
		"https://schema.org/PreOrder",
		"https://schema.org/Discontinued",
		"https://schema.org/OutOfStock",
	}
}

// String implements the fmt.Stringer interface for Availability
func (a Availability) String() string {
	return string(a)
}

// Price formats an amount given in cents, the minor unit of the currency, as
// a decimal string with exactly two decimals: 18000 is "180.00", -5 is
// "-0.05". It uses integer arithmetic only, so every amount is exact, which
// binary floating point cannot promise: there 0.1 + 0.2 is 0.30000000000000004.
//
// For a currency without a minor unit, such as JPY, pass the amount times 100:
// Price(1000 * 100) is "1000.00", whose zero decimals state the same value.
func Price(cents int64) string {
	// The magnitude goes through uint64 because the negation of
	// math.MinInt64 does not fit an int64. Negating the uint64 conversion
	// instead wraps around to the right magnitude for every negative value,
	// that one included.
	magnitude := uint64(cents)
	var b []byte
	if cents < 0 {
		magnitude = -magnitude
		b = append(b, '-')
	}
	b = strconv.AppendUint(b, magnitude/100, 10)
	fraction := magnitude % 100
	return string(append(b, '.', byte('0'+fraction/10), byte('0'+fraction%10)))
}
