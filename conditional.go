package mx

import (
	"context"
	"iter"
)

// If returns an [IfElse] that renders comps only if cond is true. Chain
// [IfElse.Else], [IfElse.ElseIf] or [IfElse.ElseIff] to supply alternatives.
func If(cond bool, comps ...Component) IfElse {
	return IfElse{cond: cond, comps: comps}
}

// Iff is like [If] but evaluates the condition by calling condFunc. Because the
// function is called eagerly at construction time, it offers no short-circuit
// benefit; it is a convenience for conditions expressed as a closure.
func Iff(condFunc func() bool, comps ...Component) IfElse {
	return IfElse{cond: condFunc(), comps: comps}
}

// IfElse is a conditional [Component] built by [If] or [Iff]. It renders its
// components when its condition is true and otherwise renders nothing, unless
// an alternative is supplied via Else, ElseIf or ElseIff.
type IfElse struct {
	cond  bool
	comps Components
}

// Render renders the held components if the condition is true, otherwise nothing.
func (i IfElse) Render(ctx context.Context, w Writer) error {
	if !i.cond {
		return nil
	}
	return i.comps.Render(ctx, w)
}

// Else returns the original components if the condition was true, otherwise the
// fallback comps. The result is a [Components] slice, ending the chain.
func (i IfElse) Else(comps ...Component) Components {
	if i.cond {
		return i.comps
	}
	return comps
}

// ElseIf continues the chain with another condition. Note that, like Go's own
// if/else if, only one branch should ultimately render; this returns a fresh
// [IfElse] for cond and does not retain the receiver's condition, so build the
// chain so that earlier true conditions are handled before reaching here.
func (i IfElse) ElseIf(cond bool, comps ...Component) IfElse {
	return IfElse{cond: cond, comps: comps}
}

// ElseIff is like [IfElse.ElseIf] but evaluates the condition by calling condFunc.
func (i IfElse) ElseIff(condFunc func() bool, comps ...Component) IfElse {
	return IfElse{cond: condFunc(), comps: comps}
}

// ForEach maps each element of values to a [Component] via componentForValue
// and returns them as a [Components] slice, the markup equivalent of a range
// loop.
func ForEach[V any, C Component](values []V, componentForValue func(V) C) Components {
	var comps Components
	for _, val := range values {
		comps = append(comps, componentForValue(val))
	}
	return comps
}

// ForEachIter is like [ForEach] but ranges over an [iter.Seq] sequence instead
// of a slice.
func ForEachIter[V any, C Component](values iter.Seq[V], componentForValue func(V) C) Components {
	var comps Components
	for val := range values {
		comps = append(comps, componentForValue(val))
	}
	return comps
}
