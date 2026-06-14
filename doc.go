// Package mx builds HTML, SVG and XML markup as a tree of [Component] values —
// [Element]s with their [Attrib]s, [Text] and other components — that is written
// out by [Component.Render].
//
// # Deferred errors
//
// Element and attribute constructors are made to nest and chain so that building
// the tree reads like the markup it produces, for example (using the html and
// svg packages that build on this one):
//
//	html.Div(html.Class("card"), svg.Circle(svg.CX(50), svg.R(40)))
//
// To keep that composition simple, these constructors return a single value and
// never a separate error: an error result on every call would make the nested,
// variadic form impossible to write. A constructor that nevertheless detects an
// invalid input therefore does not panic or silently drop it — it returns an
// Element or Attrib implementation that holds the error and reports it when the
// tree is rendered:
//
//   - [NewErrElement] returns an [Element] whose Err field holds the error;
//     [Element.Render] returns that error instead of writing the element.
//   - [ErrAttrib] is an [Attrib] whose AttribValue always returns its error,
//     which aborts rendering of the element that carries it.
//
// A build-time problem is thus never lost: it surfaces the first time the
// affected subtree is rendered ([Element.String], used in tests and debugging,
// renders it as "mx.Element.String: <error>"). Code that needs to detect such a
// problem earlier can inspect [Element.Err] or call [Attrib.AttribValue] before
// rendering. The same shape recurs in subpackages — for example svg.KeySplines
// returns an [ErrAttrib] for an invalid value count, and the svg keyword-enum
// attributes return their generated Validate error from AttribValue.
package mx
