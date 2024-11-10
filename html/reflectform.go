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

func ReflectFormComponents(formStruct any) (comps mx.Components) {
	for field, val := range mx.FlatExportedStructFieldsAndValues(reflect.ValueOf(formStruct)) {
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
			for _, attr := range strings.Split(inputTag, "|") {
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
			inputAttribs = append(inputAttribs, Name(field.Name))
		}
		if inputType == "" {
			switch {
			case field.Type.Implements(reflect.TypeFor[OptionsProvider]()):
				selectOptions = val.Interface().(OptionsProvider).Options()
			case field.Type == reflect.TypeFor[time.Time]() || field.Type == reflect.TypeFor[*time.Time]():
				inputType = "datetime-local"
			default:
				switch field.Type.Kind() {
				case reflect.Bool:
					inputType = "checkbox"
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					inputType = "number"
				case reflect.Float32, reflect.Float64:
					inputType = "number"
				}
			}
			if inputType != "" {
				inputAttribs = append(inputAttribs, Type(inputType))
			}
		}

		// Should we check with && !mx.IsNull(val.Interface()) for non zero values that can represent NULL?
		if !val.IsZero() {
			var value string
			switch inputType {
			case "checkbox":
				if field.Type.Kind() == reflect.Pointer && val.Bool() {
					value = "on" // or use the checked attribute?
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
			comps = append(comps,
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
			comps = append(comps, inputElement)
		case "checkbox", "radio":
			// Postfix input with label
			comps = append(comps, Label(inputElement, label))
		default:
			// Prefix input with label
			if !strings.HasSuffix(label, ":") {
				label += ":"
			}
			comps = append(comps, Label(label, inputElement))
		}

	}
	return comps
}
