package pdf

import (
	"context"
	"iter"
)

// If renders comps only when cond is true, mirroring mx.If. Use the returned
// [IfElse]'s Else / ElseIf methods for the false branch.
func If(cond bool, comps ...Component) IfElse {
	return IfElse{cond: cond, comps: comps}
}

// Iff is like [If] but takes the condition as a function, so the condition can
// be computed lazily at call time.
func Iff(condFunc func() bool, comps ...Component) IfElse {
	return IfElse{cond: condFunc(), comps: comps}
}

// IfElse is a conditional component returned by [If] and [Iff].
type IfElse struct {
	cond  bool
	comps Components
}

// Render draws the wrapped components when the condition is true and does
// nothing otherwise.
func (i IfElse) Render(ctx context.Context, r *Renderer) error {
	if !i.cond {
		return nil
	}
	return i.comps.Render(ctx, r)
}

// Else returns the components to render when the condition was false.
func (i IfElse) Else(comps ...Component) Components {
	if i.cond {
		return i.comps
	}
	return comps
}

// ElseIf starts a new conditional branch when the previous condition was false.
func (i IfElse) ElseIf(cond bool, comps ...Component) IfElse {
	if i.cond {
		return i
	}
	return IfElse{cond: cond, comps: comps}
}

// ForEach builds a component for every value in a slice, mirroring mx.ForEach.
func ForEach[V any, C Component](values []V, componentForValue func(V) C) Components {
	comps := make(Components, 0, len(values))
	for _, val := range values {
		comps = append(comps, componentForValue(val))
	}
	return comps
}

// ForEachIter is [ForEach] over an iterator sequence.
func ForEachIter[V any, C Component](values iter.Seq[V], componentForValue func(V) C) Components {
	var comps Components
	for val := range values {
		comps = append(comps, componentForValue(val))
	}
	return comps
}
