// Package entity provides HTML named character references as ready-to-use
// [mx.Raw] components. Each constant renders its named entity (for example
// entity.Copyright renders "&copy;"), which the browser displays as the
// corresponding character.
//
// They live in their own package so the many short names (Heart, Minus,
// Infinity, ...) do not pollute the html package namespace or collide with
// its element and attribute constructors.
//
// The set is curated for everyday web, typographic, and mathematical use; it
// deliberately omits accented Latin letters (write the UTF-8 character
// directly) and the long tail of rarely used MathML symbols. The constants in
// entity.go are generated and validated against the vendored WHATWG
// entities.json by gen.go.
package entity

//go:generate go run gen.go
