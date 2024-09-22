package xml

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
