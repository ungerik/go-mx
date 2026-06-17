package shadcn

import (
	"github.com/domonda/go-errs"
)

// PanicOnInvalidID controls how [validateID] reports an invalid id.
//
// Component ids are developer-supplied (constants or trusted values), not user
// input, so an invalid id is a programming bug rather than a runtime condition.
// When PanicOnInvalidID is true (the default) validateID panics on an invalid
// id, surfacing that bug immediately at the call site.
//
// Set it to false to make validateID return the error instead. The component
// constructors then defer it to render time via [mx.NewErrElement]: an invalid
// id yields an Element whose Render returns the error, so a stray bad id can
// never inject unescaped markup and never aborts the program.
var PanicOnInvalidID = true

// validateID checks that id is a non-empty string of letters, digits, '-' and
// '_'. Components in this package interpolate the id into an onclick handler,
// an id, name, for, aria-controls or aria-labelledby attribute, so it must be a
// safe, valid HTML id.
//
// On an invalid id validateID panics if [PanicOnInvalidID] is true (the
// default), otherwise it returns the error for the caller to defer via
// [mx.NewErrElement]. It returns nil for a valid id.
func validateID(id string) error {
	if id == "" {
		err := errs.New("shadcn: id must not be empty")
		if PanicOnInvalidID {
			panic(err)
		}
		return err
	}
	for _, r := range id {
		ok := r == '-' || r == '_' ||
			(r >= '0' && r <= '9') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z')
		if !ok {
			err := errs.Errorf("shadcn: id must contain only letters, digits, '-' and '_', got: %q", id)
			if PanicOnInvalidID {
				panic(err)
			}
			return err
		}
	}
	return nil
}
