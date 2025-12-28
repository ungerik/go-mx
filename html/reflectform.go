package html

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
)

type OptionsProvider interface {
	Options() []string
}

type NamedOptionsProvider interface {
	NamedOptions() [][2]string
}

type ReflectFormOption interface {
	ReflectFormOption() // Marker method
}

type ReflectFormOptionInputName func(reflect.StructField, reflect.Value) (inputName string, ok bool)

func (ReflectFormOptionInputName) ReflectFormOption() {}

func (f ReflectFormOptionInputName) InputName(field reflect.StructField, val reflect.Value) (inputName string, ok bool) {
	return f(field, val)
}

type ReflectFormOptionInputType func(reflect.StructField, reflect.Value) (inputType string, ok bool)

func (ReflectFormOptionInputType) ReflectFormOption() {}

func (f ReflectFormOptionInputType) InputType(field reflect.StructField, val reflect.Value) (inputType string, ok bool) {
	return f(field, val)
}

type ReflectFormOptionInputValue func(reflect.StructField, reflect.Value) (inputValue string, ok bool)

func (ReflectFormOptionInputValue) ReflectFormOption() {}

func (f ReflectFormOptionInputValue) InputValue(field reflect.StructField, val reflect.Value) (inputValue string, ok bool) {
	return f(field, val)
}

func ReflectFormComponents(formStruct any, options ...ReflectFormOption) (components mx.Components) {
	for field, val := range mx.ReflectStructFields(reflect.ValueOf(formStruct)) {
		inputTag := field.Tag.Get("input")
		if inputTag == "-" {
			continue
		}
		var (
			hasInputName = false
			// isRequired   = false
			inputType     = ""
			inputAttribs  []mx.Attrib
			selectOptions []string
		)
		if inputTag != "" {
			for attr := range strings.SplitSeq(inputTag, "|") {
				attrName, attrVal, _ := strings.Cut(attr, "=")
				switch attrName {
				case "name":
					hasInputName = true
				case "type":
					inputType = attrVal
					// case "required":
					// 	isRequired = true
				}
				if attrVal == "" {
					attrVal = attrName // Boolean attributes like required
				}
				inputAttribs = append(inputAttribs, Attrib(attrName, attrVal))
			}
		}
		if !hasInputName {
			for _, option := range options {
				if option, ok := option.(ReflectFormOptionInputName); ok {
					if name, ok := option.InputName(field, val); ok {
						inputAttribs = append(inputAttribs, Name(name))
						hasInputName = true
						break
					}
				}
			}
		}
		if !hasInputName {
			inputAttribs = append(inputAttribs, Name(field.Name))
		}
		if inputType == "" {
			for _, option := range options {
				if option, ok := option.(ReflectFormOptionInputType); ok {
					if inputType, ok = option.InputType(field, val); ok {
						inputAttribs = append(inputAttribs, Type(inputType))
						break
					}
				}
			}
		}
		if inputType == "" {
			if field.Type.Implements(reflect.TypeFor[OptionsProvider]()) {
				selectOptions = val.Interface().(OptionsProvider).Options()
			} else {
				inputType = defaultReflectFormInputType(field)
				if inputType != "" {
					inputAttribs = append(inputAttribs, Type(inputType))
				}
			}
		}

		hasInputValue := false
		for _, option := range options {
			if option, ok := option.(ReflectFormOptionInputValue); ok {
				if value, ok := option.InputValue(field, val); ok {
					inputAttribs = append(inputAttribs, Value(value))
					hasInputValue = true
					break
				}
			}
		}
		if !hasInputValue && !val.IsZero() && !mx.IsNull(val.Interface()) {
			var value string
			switch inputType {
			case "checkbox":
				if field.Type.Kind() == reflect.Bool && val.Bool() {
					value = "on"
				} else if field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Bool && val.Elem().Bool() {
					value = "on"
				}
			case "datetime", "datetime-local":
				// TODO worry about format details
				switch field.Type {
				case reflect.TypeFor[time.Time]():
					value = val.Interface().(time.Time).Format("2006-01-02T15:04:05")
				case reflect.TypeFor[*time.Time]():
					value = val.Interface().(*time.Time).Format("2006-01-02T15:04:05")
				default:
					value = fmt.Sprint(val.Interface())
				}
			default:
				value = fmt.Sprint(val.Interface())
			}
			inputAttribs = append(inputAttribs, Value(value))
		}

		label := field.Name
		if l := field.Tag.Get("label"); l != "" {
			label = l
		}

		if selectOptions != nil {
			if !strings.HasSuffix(label, ":") {
				label += ":"
			}
			components = append(components,
				Label(
					label,
					Select(inputAttribs,
						ForEach(selectOptions, func(option string) *mx.Element {
							return Option(Value(option), option)
						}),
					),
				),
			)
			continue
		}

		inputElement := Input(inputAttribs...)

		switch inputType {
		case "hidden", "submit", "image", "reset", "button":
			// No label
			components = append(components, inputElement)
		case "checkbox", "radio":
			// Postfix input with label
			components = append(components, Label(inputElement, label))
		default:
			// Prefix input with label
			if !strings.HasSuffix(label, ":") {
				label += ":"
			}
			components = append(components, Label(label, inputElement))
		}
	}
	return components
}

func defaultReflectFormInputType(field reflect.StructField) string {
	if field.Type == reflect.TypeFor[time.Time]() || field.Type == reflect.TypeFor[*time.Time]() {
		return "datetime-local"
	}
	switch field.Type.Kind() {
	case reflect.Bool:
		return "checkbox"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "number"
	}
	return ""
}
