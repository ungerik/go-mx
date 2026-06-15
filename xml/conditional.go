package xml

import (
	"iter"

	"github.com/ungerik/go-mx"
)

// If renders comps when cond is true and nothing otherwise. Call the Else method
// on the result to provide an alternative for the false case.
func If(cond bool, comps ...mx.Component) mx.IfElse {
	return mx.If(cond, comps...)
}

// Iff is like [If] but evaluates condFunc lazily at render time instead of
// taking a precomputed boolean.
func Iff(condFunc func() bool, comps ...mx.Component) mx.IfElse {
	return mx.Iff(condFunc, comps...)
}

// ForEach renders componentForValue once for each element of values and returns
// the resulting components in order.
func ForEach[V any, C mx.Component](values []V, componentForValue func(V) C) mx.Components {
	return mx.ForEach(values, componentForValue)
}

// ForEachIter is like [ForEach] but iterates over an iter.Seq instead of a slice.
func ForEachIter[V any, C mx.Component](values iter.Seq[V], componentForValue func(V) C) mx.Components {
	return mx.ForEachIter(values, componentForValue)
}
