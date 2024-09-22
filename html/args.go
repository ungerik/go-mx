package html

import "github.com/ungerik/go-mx/xml"

type (
	Attrib     = xml.Attrib
	Attributes = xml.Attributes
	Attribs    = xml.Attribs
	Attribute  = xml.Attribute
)

type ID string

func (a ID) Get() (name, value string) {
	return "id", string(a)
}

type Class string

func (a Class) Get() (name, value string) {
	return "class", string(a)
}

type StyleArg string

func (a StyleArg) Get() (name, value string) {
	return "style", string(a)
}

type Lang string

func (a Lang) Get() (name, value string) {
	return "lang", string(a)
}
