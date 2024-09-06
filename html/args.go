package html

import (
	"iter"
)

type Attrib interface {
	Get() (name, value string)
}

type Attributes interface {
	AttribIter() iter.Seq2[string, string]
}

type Attribs []Attrib

func (args Attribs) Iter() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, arg := range args {
			key, value := arg.Get()
			if !yield(key, value) {
				return
			}
		}
	}
}

type Attribute struct {
	Name  string
	Value string
}

func (a Attribute) Get() (name, value string) {
	return a.Name, a.Value
}

type ID string

func (a ID) Get() (name, value string) {
	return "id", string(a)
}

type Class string

func (a Class) Get() (name, value string) {
	return "class", string(a)
}

type Style string

func (a Style) Get() (name, value string) {
	return "style", string(a)
}

type Lang string

func (a Lang) Get() (name, value string) {
	return "lang", string(a)
}
