package html

import (
	"iter"

	"github.com/ungerik/go-mx"
)

func If(cond bool, comps ...mx.Component) mx.IfElse {
	return mx.If(cond, comps...)
}

func Iff(condFunc func() bool, comps ...mx.Component) mx.IfElse {
	return mx.Iff(condFunc, comps...)
}

func ForEach[V any, C mx.Component](values []V, componentForValue func(V) C) mx.Components {
	return mx.ForEach(values, componentForValue)
}

func ForEachIter[V any, C mx.Component](values iter.Seq[V], componentForValue func(V) C) mx.Components {
	return mx.ForEachIter(values, componentForValue)
}
