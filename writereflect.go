package mx

import (
	"errors"
	"fmt"
	"reflect"
)

func WriteStructFieldsAsAttributes(w Writer, elem string, s any) error {
	for field, val := range FlatExportedStructFields(reflect.ValueOf(s)) {
		if attribs, _ := val.Interface().([]Attrib); attribs != nil {
			for i := range attribs {
				err := w.Attribute(attribs[i].Name, attribs[i].Value)
				if err != nil {
					return err
				}
			}
			continue
		}
		name := field.Tag.Get("attr")
		if name == "" || name == "-" {
			continue
		}
		value := fmt.Sprint(val.Interface())
		if value == "" {
			continue
		}
		err := w.Attribute(name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteStructAsVoidElement(w Writer, elem string, s any) error {
	return errors.Join(
		w.BeginElement(elem),
		WriteStructFieldsAsAttributes(w, elem, s),
		w.CloseAndEndElement(),
	)
}

func WriteStructAsElementStart(w Writer, elem string, s any) error {
	return errors.Join(
		w.BeginElement(elem),
		WriteStructFieldsAsAttributes(w, elem, s),
		w.CloseElement(),
	)
}
