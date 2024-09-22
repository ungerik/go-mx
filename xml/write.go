package xml

import (
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/ungerik/go-mx"
)

func WriteStructFieldsAsAttributes(w io.Writer, renderer Renderer, elem string, s any) error {
	for field, val := range mx.FlatExportedStructFields(reflect.ValueOf(s)) {
		if a, _ := val.Interface().(Attribs); a != nil {
			for name, value := range a.Iter() {
				err := renderer.Attribute(w, name, value)
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
		err := renderer.Attribute(w, name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteStructAsVoidElement(w io.Writer, renderer Renderer, elem string, s any) error {
	return errors.Join(
		renderer.OpenElement(w, elem),
		WriteStructFieldsAsAttributes(w, renderer, elem, s),
		renderer.CloseVoidElement(w),
	)
}

func WriteStructAsElementStart(w io.Writer, renderer Renderer, elem string, s any) error {
	return errors.Join(
		renderer.OpenElement(w, elem),
		WriteStructFieldsAsAttributes(w, renderer, elem, s),
		renderer.CloseElement(w),
	)
}
