//go:generate go -C ../../tools tool go-enum ../web/schemaorg/$GOFILE

package schemaorg

// https://schema.org/Event
// https://developers.google.com/search/docs/appearance/structured-data/event

import (
	"fmt"
	"time"
)

// Event is something happening at a time and place: a concert, a talk, a
// workshop, a webinar. Google shows an event in its event search with a name,
// a StartDate and a Location; the rest is optional.
//
// https://schema.org/Event
type Event struct {
	Thing

	// StartDate is when the event starts, built with [DateTime], or with
	// [Date] for an all-day event.
	StartDate string `json:"startDate,omitempty"`

	// EndDate is when the event ends, built like StartDate.
	EndDate string `json:"endDate,omitempty"`

	// PreviousStartDate is the original start of a rescheduled event, whose
	// StartDate holds the new one and EventStatus [EventStatusRescheduled].
	PreviousStartDate string `json:"previousStartDate,omitempty"`

	// EventStatus tells whether the event takes place as planned. Empty is
	// read as [EventStatusScheduled].
	EventStatus EventStatus `json:"eventStatus,omitempty"`

	// EventAttendanceMode tells whether the event is attended in person,
	// online or both. Empty is read as in person.
	EventAttendanceMode EventAttendanceMode `json:"eventAttendanceMode,omitempty"`

	// Location is where the event takes place: a [Place] for an in-person
	// event, a [VirtualLocation] or the URL as text for an online one, or a
	// []any holding one of each for a mixed event.
	Location any `json:"location,omitempty"`

	// Organizer is who runs the event: an [Organization] or a [Person].
	Organizer any `json:"organizer,omitempty"`

	// Performer is who performs or speaks at the event: a [Person] or an
	// [Organization], or a slice of them.
	Performer any `json:"performer,omitempty"`

	// Offers are the tickets: an [Offer] or a slice of them.
	Offers any `json:"offers,omitempty"`

	// InLanguage is the language the event is held in as a BCP 47 language
	// tag ("en", "de-AT").
	InLanguage string `json:"inLanguage,omitempty"`
}

// NewEvent returns an Event with the given name, starting at start formatted
// with [DateTime]. For an all-day event overwrite StartDate with [Date].
func NewEvent(name string, start time.Time) *Event {
	event := &Event{Thing: newThing("Event"), StartDate: DateTime(start)}
	event.Name = name
	return event
}

// EventStatus is the status of an [Event]. The values are the full schema.org
// URLs, the form Google's documentation uses. A changed event keeps its page
// and its structured data with the new status rather than disappearing, so a
// search result can show the change.
type EventStatus string //#enum

const (
	// EventStatusScheduled is an event taking place as planned, the
	// default.
	EventStatusScheduled EventStatus = "https://schema.org/EventScheduled"
	// EventStatusCancelled is an event that will not take place.
	EventStatusCancelled EventStatus = "https://schema.org/EventCancelled"
	// EventStatusMovedOnline is an in-person event that will take place
	// online instead; its EventAttendanceMode changes accordingly.
	EventStatusMovedOnline EventStatus = "https://schema.org/EventMovedOnline"
	// EventStatusPostponed is an event put off to a date not yet known; the
	// StartDate stays the original one until it is rescheduled.
	EventStatusPostponed EventStatus = "https://schema.org/EventPostponed"
	// EventStatusRescheduled is an event moved to a new StartDate, with the
	// original one in PreviousStartDate.
	EventStatusRescheduled EventStatus = "https://schema.org/EventRescheduled"
)

// Valid indicates if e is any of the valid values for EventStatus
func (e EventStatus) Valid() bool {
	switch e {
	case
		EventStatusScheduled,
		EventStatusCancelled,
		EventStatusMovedOnline,
		EventStatusPostponed,
		EventStatusRescheduled:
		return true
	}
	return false
}

// Validate returns an error if e is none of the valid values for EventStatus
func (e EventStatus) Validate() error {
	if !e.Valid() {
		return fmt.Errorf("invalid value %#v for type schemaorg.EventStatus", e)
	}
	return nil
}

// Enums returns all valid values for EventStatus
func (EventStatus) Enums() []EventStatus {
	return []EventStatus{
		EventStatusScheduled,
		EventStatusCancelled,
		EventStatusMovedOnline,
		EventStatusPostponed,
		EventStatusRescheduled,
	}
}

// EnumStrings returns all valid values for EventStatus as strings
func (EventStatus) EnumStrings() []string {
	return []string{
		"https://schema.org/EventScheduled",
		"https://schema.org/EventCancelled",
		"https://schema.org/EventMovedOnline",
		"https://schema.org/EventPostponed",
		"https://schema.org/EventRescheduled",
	}
}

// String implements the fmt.Stringer interface for EventStatus
func (e EventStatus) String() string {
	return string(e)
}

// EventAttendanceMode tells how an [Event] is attended. The values are the
// full schema.org URLs, the form Google's documentation uses.
type EventAttendanceMode string //#enum

const (
	// EventAttendanceModeOffline is an event attended in person at a
	// [Place].
	EventAttendanceModeOffline EventAttendanceMode = "https://schema.org/OfflineEventAttendanceMode"
	// EventAttendanceModeOnline is an event attended online at a
	// [VirtualLocation].
	EventAttendanceModeOnline EventAttendanceMode = "https://schema.org/OnlineEventAttendanceMode"
	// EventAttendanceModeMixed is an event attended both in person and
	// online.
	EventAttendanceModeMixed EventAttendanceMode = "https://schema.org/MixedEventAttendanceMode"
)

// Valid indicates if e is any of the valid values for EventAttendanceMode
func (e EventAttendanceMode) Valid() bool {
	switch e {
	case
		EventAttendanceModeOffline,
		EventAttendanceModeOnline,
		EventAttendanceModeMixed:
		return true
	}
	return false
}

// Validate returns an error if e is none of the valid values for EventAttendanceMode
func (e EventAttendanceMode) Validate() error {
	if !e.Valid() {
		return fmt.Errorf("invalid value %#v for type schemaorg.EventAttendanceMode", e)
	}
	return nil
}

// Enums returns all valid values for EventAttendanceMode
func (EventAttendanceMode) Enums() []EventAttendanceMode {
	return []EventAttendanceMode{
		EventAttendanceModeOffline,
		EventAttendanceModeOnline,
		EventAttendanceModeMixed,
	}
}

// EnumStrings returns all valid values for EventAttendanceMode as strings
func (EventAttendanceMode) EnumStrings() []string {
	return []string{
		"https://schema.org/OfflineEventAttendanceMode",
		"https://schema.org/OnlineEventAttendanceMode",
		"https://schema.org/MixedEventAttendanceMode",
	}
}

// String implements the fmt.Stringer interface for EventAttendanceMode
func (e EventAttendanceMode) String() string {
	return string(e)
}
