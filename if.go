package mx

import "context"

func If(cond bool, comps ...Component) IfElse {
	return IfElse{cond: cond, comps: comps}
}

func Iff(condFunc func() bool, comps ...Component) IfElse {
	return IfElse{cond: condFunc(), comps: comps}
}

type IfElse struct {
	cond  bool
	comps Components
}

func (i IfElse) Render(ctx context.Context, w Writer) error {
	if !i.cond {
		return nil
	}
	return i.comps.Render(ctx, w)
}

func (i IfElse) Else(comps ...Component) Components {
	if i.cond {
		return i.comps
	}
	return comps
}

func (i IfElse) ElseIf(cond bool, comps ...Component) IfElse {
	return IfElse{cond: cond, comps: comps}
}

func (i IfElse) ElseIff(condFunc func() bool, comps ...Component) IfElse {
	return IfElse{cond: condFunc(), comps: comps}
}
