package mx

import (
	"context"
	"strconv"
)

// UniqueID returns an "id" [Attrib] whose value is unique within the running
// process. It draws from an atomic counter, so it is safe for concurrent use,
// and formats the value as "_" followed by the count in base 36, yielding a
// valid HTML id that does not start with a digit.
func UniqueID() Attrib {
	return uniqueID(idCounter.Add(1))
}

type uniqueID uint64

// AttribName returns "id".
func (id uniqueID) AttribName() string {
	return "id"
}

// AttribValue returns the unique value formatted as "_" followed by the count
// in base 36, and a nil error.
func (id uniqueID) AttribValue(context.Context) (string, error) {
	return "_" + strconv.FormatUint(uint64(id), 36), nil
}
