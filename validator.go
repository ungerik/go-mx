package mx

// Validator is the last-resort, boolean validator in the validation
// chain run by [ReflectFormHandler]. A type that has no Normalize or
// Validate method but does implement Valid() bool is asked whether the
// parsed value is acceptable. The chain synthesizes a generic "invalid
// value" error when Valid returns false because there is no message to
// surface.
//
// Types that can produce a richer error should implement
// Validate() error, or Normalize() error / Normalize() []error if they
// also canonicalize the value in place. See the validation chain in
// validation_chain.go for the full ordering.
type Validator interface {
	Valid() bool
}

// Normalizer is the richest validator in the chain: it both validates
// and (potentially) mutates the receiver into a canonical form,
// returning every error it finds. Implementations are typically pointer
// receivers because they need to write back the normalized value.
//
// When a type implements multiple validator methods, Normalize() []error
// wins over the others (see validation_chain.go).
type Normalizer interface {
	Normalize() []error
}

// SingleErrNormalizer is the single-error variant of [Normalizer]: same
// canonicalization semantics, but returns at most one error.
type SingleErrNormalizer interface {
	Normalize() error
}

// SingleErrValidator validates without normalizing and returns a single
// error. Implementations are typically value receivers because they do
// not need to write back to the receiver.
type SingleErrValidator interface {
	Validate() error
}
