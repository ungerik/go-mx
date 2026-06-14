package svg

import (
	"iter"

	"github.com/ungerik/go-mx"
)

// If renders comps when cond is true. Use the Else method on the returned
// mx.IfElse to provide a fallback rendered when cond is false.
func If(cond bool, comps ...mx.Component) mx.IfElse {
	return mx.If(cond, comps...)
}

// Iff is like If but evaluates condFunc lazily at render time.
func Iff(condFunc func() bool, comps ...mx.Component) mx.IfElse {
	return mx.Iff(condFunc, comps...)
}

// ForEach renders componentForValue for each value in values and returns the
// resulting components in order.
func ForEach[V any, C mx.Component](values []V, componentForValue func(V) C) mx.Components {
	return mx.ForEach(values, componentForValue)
}

// ForEachIter renders componentForValue for each value yielded by the iterator
// seq and returns the resulting components in order.
func ForEachIter[V any, C mx.Component](values iter.Seq[V], componentForValue func(V) C) mx.Components {
	return mx.ForEachIter(values, componentForValue)
}
