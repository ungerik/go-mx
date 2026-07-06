package html

import (
	"context"
	"encoding"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/domonda/go-errs"
	"github.com/ungerik/go-mx"
)

// FieldDecider is the plain-HTML implementation of [mx.FieldDecider].
// It implements the dispatch table from the ReflectFormHandler design:
// each detected [mx.FieldKind] picks an HTML element from this package
// (Input / Select / TextArea / Label) and a matching parser. It emits
// the __clear sentinel for nullable fields; the form handler emits the
// __present sentinel.
//
// Layered renderers (hx, shadcn) wrap this decider — calling it for
// the kinds they don't customize and falling through for the rest.
var FieldDecider mx.FieldDecider = func(path mx.FieldPath, field reflect.StructField, value reflect.Value) mx.FieldBehavior {
	kind, tag := mx.DetectField(path, field, value)
	return mx.FieldBehavior{
		Render: func(path mx.FieldPath, field reflect.StructField, value reflect.Value, errs []error) mx.Component {
			return renderField(path, field, value, kind, tag, errs)
		},
		Parse: func(path mx.FieldPath, field reflect.StructField, value reflect.Value, r *http.Request) error {
			return parseField(path, field, value, kind, tag, r)
		},
	}
}

// renderField produces the HTML for one form field: an optional
// <label>, the input element itself, optional help text, the __clear
// sentinel for nullable fields, and inline error messages. The
// form-level <form> wrapper and __present sentinel are emitted by the
// handler — this function only renders the per-field markup.
func renderField(path mx.FieldPath, field reflect.StructField, value reflect.Value, kind mx.FieldKind, tag mx.FormTag, errs []error) mx.Component {
	switch kind {
	case mx.FieldKindHidden:
		return renderHidden(path, value)
	case mx.FieldKindSkip:
		return nil
	}

	input := renderInput(path, field, value, kind, tag, errs)
	if input == nil {
		return nil
	}

	parts := mx.Components{}
	if labelText := fieldLabel(field, tag); labelText != "" && kind != mx.FieldKindBool {
		parts = append(parts, Label(For(string(path)), labelText+":"))
	}
	parts = append(parts, input)
	if kind == mx.FieldKindBool {
		if labelText := fieldLabel(field, tag); labelText != "" {
			parts = append(parts, Label(For(string(path)), " "+labelText))
		}
	}
	if tag.Help != "" {
		parts = append(parts, mx.NewElement("small",
			mx.Attribute{Name: "class", Value: "form-help"},
			mx.Text(tag.Help),
		))
	}
	if isNullable(value) && !tag.Required {
		parts = append(parts, clearSentinelInput(path))
	}
	if len(errs) > 0 {
		for _, e := range errs {
			parts = append(parts, mx.NewElement("p",
				mx.Attribute{Name: "class", Value: "form-error"},
				mx.Attribute{Name: "data-error", Value: string(path)},
				mx.Text(e.Error()),
			))
		}
	}
	return parts
}

// renderInput emits the input element for one field based on its
// [mx.FieldKind].
func renderInput(path mx.FieldPath, field reflect.StructField, value reflect.Value, kind mx.FieldKind, tag mx.FormTag, errs []error) mx.Component {
	switch kind {
	case mx.FieldKindString:
		t := stringWidgetType(tag, field.Type)
		return inputElement(path, t, displayValue(value, tag), field, tag, errs)
	case mx.FieldKindTextarea:
		return textareaElement(path, displayValue(value, tag), field, tag, errs)
	case mx.FieldKindNumber:
		return numberInput(path, value, tag, errs)
	case mx.FieldKindBool:
		return checkboxInput(path, value, tag, errs)
	case mx.FieldKindDateTime:
		return datetimeInput(path, value, tag, errs)
	case mx.FieldKindFile:
		return fileInput(path, tag, errs)
	case mx.FieldKindEnum:
		return selectInput(path, value, tag, field, errs)
	case mx.FieldKindEnumSet:
		return enumSetInput(path, value, field, tag, errs)
	case mx.FieldKindCatchAll:
		return inputElement(path, "text", displayValue(value, tag), field, tag, errs)
	}
	return nil
}

func renderHidden(path mx.FieldPath, value reflect.Value) mx.Component {
	return VoidElement("input",
		Type("hidden"),
		Name(string(path)),
		Value(displayValue(value, mx.FormTag{})),
	)
}

func inputElement(path mx.FieldPath, inputType, val string, _ reflect.StructField, tag mx.FormTag, errs []error) mx.Component {
	attribs := []mx.Attrib{
		Type(inputType),
		Name(string(path)),
		ID(string(path)),
	}
	if val != "" && !tag.Sensitive {
		attribs = append(attribs, Value(val))
	}
	if tag.Placeholder != "" {
		attribs = append(attribs, Placeholder(tag.Placeholder))
	}
	if tag.Pattern != "" {
		attribs = append(attribs, Pattern(tag.Pattern))
	}
	if tag.Required {
		attribs = append(attribs, Required)
	}
	if tag.Readonly {
		attribs = append(attribs, Readonly)
	}
	if len(errs) > 0 {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	return Input(attribs...)
}

func textareaElement(path mx.FieldPath, val string, _ reflect.StructField, tag mx.FormTag, errs []error) mx.Component {
	attribs := []any{
		Name(string(path)),
		ID(string(path)),
	}
	if tag.Placeholder != "" {
		attribs = append(attribs, Placeholder(tag.Placeholder))
	}
	if tag.Required {
		attribs = append(attribs, Required)
	}
	if tag.Readonly {
		attribs = append(attribs, Readonly)
	}
	if len(errs) > 0 {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	if val != "" && !tag.Sensitive {
		attribs = append(attribs, mx.Text(val))
	}
	return TextArea(attribs...)
}

func numberInput(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	attribs := []mx.Attrib{
		Type("number"),
		Name(string(path)),
		ID(string(path)),
		Value(displayValue(value, tag)),
	}
	if tag.Min != "" {
		attribs = append(attribs, Min(tag.Min))
	}
	if tag.Max != "" {
		attribs = append(attribs, Max(tag.Max))
	}
	if tag.Step != "" {
		attribs = append(attribs, Step(tag.Step))
	}
	if tag.Required {
		attribs = append(attribs, Required)
	}
	if tag.Readonly {
		attribs = append(attribs, Readonly)
	}
	if len(errs) > 0 {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	return Input(attribs...)
}

func checkboxInput(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	attribs := []mx.Attrib{
		Type("checkbox"),
		Name(string(path)),
		ID(string(path)),
		Value("on"),
	}
	if isTrueBool(value) {
		attribs = append(attribs, Checked)
	}
	if tag.Readonly {
		attribs = append(attribs, Disabled)
	}
	if len(errs) > 0 {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	return Input(attribs...)
}

func datetimeInput(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	t := "datetime-local"
	switch tag.Widget {
	case "date":
		t = "date"
	case "time":
		t = "time"
	}
	val := timeDisplayValue(value, t)
	attribs := []mx.Attrib{
		Type(t),
		Name(string(path)),
		ID(string(path)),
	}
	if val != "" {
		attribs = append(attribs, Value(val))
	}
	if tag.Min != "" {
		attribs = append(attribs, Min(tag.Min))
	}
	if tag.Max != "" {
		attribs = append(attribs, Max(tag.Max))
	}
	if tag.Required {
		attribs = append(attribs, Required)
	}
	if tag.Readonly {
		attribs = append(attribs, Readonly)
	}
	if len(errs) > 0 {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	return Input(attribs...)
}

func fileInput(path mx.FieldPath, tag mx.FormTag, errs []error) mx.Component {
	attribs := []mx.Attrib{
		Type("file"),
		Name(string(path)),
		ID(string(path)),
	}
	if tag.Required {
		attribs = append(attribs, Required)
	}
	if len(errs) > 0 {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	return Input(attribs...)
}

func selectInput(path mx.FieldPath, value reflect.Value, tag mx.FormTag, field reflect.StructField, errs []error) mx.Component {
	selected := displayValue(value, tag)
	attribs := []any{
		Name(string(path)),
		ID(string(path)),
	}
	if tag.Required {
		attribs = append(attribs, Required)
	}
	if tag.Readonly {
		attribs = append(attribs, Disabled)
	}
	if len(errs) > 0 {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	options := optionChildren(value, tag, field.Type, func(opt mx.NamedOption) mx.Component {
		oAttr := []any{Value(opt.Value)}
		if opt.Value == selected {
			oAttr = append(oAttr, Selected)
		}
		oAttr = append(oAttr, mx.Text(displayLabel(opt)))
		return Option(oAttr...)
	})
	return Select(append(attribs, options)...)
}

func enumSetInput(path mx.FieldPath, value reflect.Value, field reflect.StructField, tag mx.FormTag, errs []error) mx.Component {
	keyType := setKeyType(field.Type)
	if keyType == nil {
		return mx.Text("[enum-set: cannot infer key type for " + string(path) + "]")
	}
	selectedSet := setSelectedValues(value)
	items := optionChildren(reflect.Value{}, tag, keyType, func(opt mx.NamedOption) mx.Component {
		inputAttribs := []mx.Attrib{
			Type("checkbox"),
			Name(string(path)),
			Value(opt.Value),
		}
		if _, ok := selectedSet[opt.Value]; ok {
			inputAttribs = append(inputAttribs, Checked)
		}
		if tag.Readonly {
			inputAttribs = append(inputAttribs, Disabled)
		}
		return Label(Input(inputAttribs...), " "+displayLabel(opt))
	})
	if len(errs) > 0 {
		return mx.Components{
			mx.NewElement("div",
				mx.Attribute{Name: "role", Value: "group"},
				mx.Attribute{Name: "aria-invalid", Value: "true"},
				items,
			),
		}
	}
	return mx.NewElement("div", mx.Attribute{Name: "role", Value: "group"}, items)
}

// parseField writes the submitted form value into value (which is
// addressable). For multipart file uploads, parseField stores the
// uploaded file's bytes when the destination is []byte; richer file
// handling is left to higher layers.
func parseField(path mx.FieldPath, field reflect.StructField, value reflect.Value, kind mx.FieldKind, tag mx.FormTag, r *http.Request) error {
	if !value.CanSet() {
		return errs.New("field not settable: " + string(path))
	}
	form := r.PostForm
	if r.MultipartForm != nil {
		form = r.MultipartForm.Value
	}
	switch kind {
	case mx.FieldKindHidden, mx.FieldKindString, mx.FieldKindTextarea, mx.FieldKindCatchAll:
		raw := form.Get(string(path))
		return setScalar(value, raw)
	case mx.FieldKindNumber:
		raw := form.Get(string(path))
		return setNumeric(value, raw)
	case mx.FieldKindBool:
		raw := form.Get(string(path))
		return setBool(value, raw)
	case mx.FieldKindDateTime:
		raw := form.Get(string(path))
		return setTime(value, raw, tag.Widget)
	case mx.FieldKindEnum:
		raw := form.Get(string(path))
		return setScalar(value, raw)
	case mx.FieldKindEnumSet:
		vals := form[string(path)]
		return setEnumSet(value, field.Type, vals)
	case mx.FieldKindFile:
		if r.MultipartForm == nil {
			return nil
		}
		files, ok := r.MultipartForm.File[string(path)]
		if !ok || len(files) == 0 {
			return nil
		}
		fh := files[0]
		f, err := fh.Open()
		if err != nil {
			return err
		}
		defer f.Close()
		buf := make([]byte, fh.Size)
		if _, err := f.Read(buf); err != nil && err.Error() != "EOF" {
			return err
		}
		if value.Kind() == reflect.Slice && value.Type().Elem().Kind() == reflect.Uint8 {
			value.SetBytes(buf)
			return nil
		}
		// String destinations get the filename only.
		if value.Kind() == reflect.String {
			value.SetString(fh.Filename)
			return nil
		}
		return errs.New("file upload requires []byte or string field")
	}
	return nil
}

// setScalar writes raw into a string-shaped or TextUnmarshaler
// destination. Pointer types are allocated on demand.
func setScalar(value reflect.Value, raw string) error {
	if value.Kind() == reflect.Pointer {
		if raw == "" {
			value.SetZero()
			return nil
		}
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setScalar(value.Elem(), raw)
	}
	if value.Kind() == reflect.String {
		value.SetString(raw)
		return nil
	}
	if value.CanAddr() {
		if u, ok := value.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return u.UnmarshalText([]byte(raw))
		}
	}
	if u, ok := value.Interface().(encoding.TextUnmarshaler); ok {
		return u.UnmarshalText([]byte(raw))
	}
	// Fallback: try fmt.Sscan-style conversion via reflect.
	return errs.Errorf("cannot parse value into %s — implement encoding.TextUnmarshaler or use a form:\"widget=...\" override", value.Type())
}

func setNumeric(value reflect.Value, raw string) error {
	if value.Kind() == reflect.Pointer {
		if raw == "" {
			value.SetZero()
			return nil
		}
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setNumeric(value.Elem(), raw)
	}
	if raw == "" && value.CanSet() {
		value.SetZero()
		return nil
	}
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, value.Type().Bits())
		if err != nil {
			return err
		}
		value.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, value.Type().Bits())
		if err != nil {
			return err
		}
		value.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw, value.Type().Bits())
		if err != nil {
			return err
		}
		value.SetFloat(f)
	default:
		return errs.Errorf("setNumeric: unsupported kind %s", value.Kind())
	}
	return nil
}

func setBool(value reflect.Value, raw string) error {
	on := raw == "on" || raw == "true" || raw == "1"
	if value.Kind() == reflect.Pointer {
		if !value.CanSet() {
			return errs.New("pointer not settable")
		}
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value.Elem().SetBool(on)
		return nil
	}
	value.SetBool(on)
	return nil
}

func setTime(value reflect.Value, raw, widget string) error {
	if value.Kind() == reflect.Pointer {
		if raw == "" {
			value.SetZero()
			return nil
		}
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setTime(value.Elem(), raw, widget)
	}
	if raw == "" {
		value.SetZero()
		return nil
	}
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"15:04:05",
		"15:04",
		time.RFC3339,
	}
	switch widget {
	case "date":
		layouts = []string{"2006-01-02"}
	case "time":
		layouts = []string{"15:04:05", "15:04"}
	}
	var lastErr error
	for _, layout := range layouts {
		t, err := time.Parse(layout, raw)
		if err == nil {
			if value.Type() == reflect.TypeFor[time.Time]() {
				value.Set(reflect.ValueOf(t))
				return nil
			}
			// time.Time wrapped in a named type — best effort
			if value.Kind() == reflect.Struct && value.CanAddr() {
				if u, ok := value.Addr().Interface().(encoding.TextUnmarshaler); ok {
					return u.UnmarshalText([]byte(raw))
				}
			}
			return errs.New("cannot assign time to " + value.Type().String())
		}
		lastErr = err
	}
	return lastErr
}

func setEnumSet(value reflect.Value, fieldType reflect.Type, vals []string) error {
	keyType := setKeyType(fieldType)
	if keyType == nil {
		return errs.New("not a recognized set type")
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setEnumSet(value.Elem(), value.Type().Elem(), vals)
	}
	switch value.Kind() {
	case reflect.Map:
		newMap := reflect.MakeMapWithSize(value.Type(), len(vals))
		for _, v := range vals {
			kv, err := stringToType(v, keyType)
			if err != nil {
				return err
			}
			newMap.SetMapIndex(kv, reflect.Zero(value.Type().Elem()))
		}
		value.Set(newMap)
	case reflect.Slice:
		newSlice := reflect.MakeSlice(value.Type(), 0, len(vals))
		for _, v := range vals {
			kv, err := stringToType(v, keyType)
			if err != nil {
				return err
			}
			newSlice = reflect.Append(newSlice, kv)
		}
		value.Set(newSlice)
	default:
		return errs.Errorf("setEnumSet: unsupported kind %s", value.Kind())
	}
	return nil
}

// setKeyType returns the element type of a set-shaped field:
// map[T]struct{} → T, []T → T, *T of either → T. Returns nil if t is
// not set-shaped.
func setKeyType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil {
		return nil
	}
	switch t.Kind() {
	case reflect.Map:
		if t.Elem().Kind() == reflect.Struct && t.Elem().NumField() == 0 {
			return t.Key()
		}
	case reflect.Slice:
		if t.Elem().Kind() != reflect.Uint8 {
			return t.Elem()
		}
	}
	return nil
}

func stringToType(s string, t reflect.Type) (reflect.Value, error) {
	dst := reflect.New(t).Elem()
	if t.Kind() == reflect.String {
		dst.SetString(s)
		return dst, nil
	}
	if u, ok := dst.Addr().Interface().(encoding.TextUnmarshaler); ok {
		if err := u.UnmarshalText([]byte(s)); err != nil {
			return reflect.Value{}, err
		}
		return dst, nil
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		dst.SetInt(n)
		return dst, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, t.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		dst.SetUint(n)
		return dst, nil
	}
	return reflect.Value{}, errs.Errorf("cannot convert %q to %s", s, t)
}

func setSelectedValues(value reflect.Value) map[string]struct{} {
	out := map[string]struct{}{}
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return out
		}
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Map:
		iter := value.MapRange()
		for iter.Next() {
			out[fmt.Sprint(iter.Key().Interface())] = struct{}{}
		}
	case reflect.Slice:
		for i := range value.Len() {
			out[fmt.Sprint(value.Index(i).Interface())] = struct{}{}
		}
	}
	return out
}

// optionChildren resolves the field's option list via
// [mx.CollectOptions] and renders each entry through render. Static
// option lists are resolved immediately so build-time tree enrichment
// (the hx layer's attribute injection) still sees the elements; option
// lists that need the request context (form:"options=…" registry
// entries, [mx.NamedOptionsContextProvider] types) are collected when
// the component renders, where the context passed to Render is the
// real request context. A collection error surfaces at render time
// (deferred-error pattern).
func optionChildren(value reflect.Value, tag mx.FormTag, t reflect.Type, render func(opt mx.NamedOption) mx.Component) mx.Component {
	if !mx.OptionsNeedContext(tag, t) {
		options, err := mx.CollectOptions(context.Background(), value, tag, t)
		if err != nil {
			return mx.NewErrElement(err)
		}
		children := make(mx.Components, len(options))
		for i, opt := range options {
			children[i] = render(opt)
		}
		return children
	}
	return mx.ComponentFunc(func(ctx context.Context, w mx.Writer) error {
		options, err := mx.CollectOptions(ctx, value, tag, t)
		if err != nil {
			return err
		}
		for _, opt := range options {
			if err := render(opt).Render(ctx, w); err != nil {
				return err
			}
		}
		return nil
	})
}

func displayLabel(o mx.NamedOption) string {
	if o.Name != "" {
		return o.Name
	}
	return o.Value
}

// displayValue returns the printable string for value, honoring
// nullable, time, and TextMarshaler conventions.
func displayValue(value reflect.Value, tag mx.FormTag) string {
	_ = tag
	if !value.IsValid() {
		return ""
	}
	if mx.IsNull(value.Interface()) {
		return ""
	}
	v := value
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if t, ok := v.Interface().(time.Time); ok {
		return t.Format("2006-01-02T15:04:05")
	}
	if v.CanInterface() {
		if tm, ok := v.Interface().(encoding.TextMarshaler); ok {
			b, err := tm.MarshalText()
			if err == nil {
				return string(b)
			}
		}
	}
	if v.CanAddr() {
		if tm, ok := v.Addr().Interface().(encoding.TextMarshaler); ok {
			b, err := tm.MarshalText()
			if err == nil {
				return string(b)
			}
		}
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, v.Type().Bits())
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.String {
			parts := make([]string, v.Len())
			for i := range v.Len() {
				parts[i] = v.Index(i).String()
			}
			return strings.Join(parts, "\n")
		}
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return string(v.Bytes())
		}
	}
	return fmt.Sprint(v.Interface())
}

func timeDisplayValue(value reflect.Value, inputType string) string {
	if !value.IsValid() {
		return ""
	}
	v := value
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	t, ok := v.Interface().(time.Time)
	if !ok || t.IsZero() {
		return ""
	}
	switch inputType {
	case "date":
		return t.Format("2006-01-02")
	case "time":
		return t.Format("15:04:05")
	}
	return t.Format("2006-01-02T15:04:05")
}

func stringWidgetType(tag mx.FormTag, t reflect.Type) string {
	switch tag.Widget {
	case "email", "url", "tel", "password":
		return tag.Widget
	}
	if hint, ok := callFormWidgetHint(t); ok {
		switch hint {
		case "email", "url", "tel", "password":
			return hint
		}
	}
	return "text"
}

func callFormWidgetHint(t reflect.Type) (string, bool) {
	if t == nil {
		return "", false
	}
	hintType := reflect.TypeFor[mx.FormWidgetHint]()
	if t.Implements(hintType) {
		v := reflect.New(t).Elem()
		if h, ok := v.Interface().(mx.FormWidgetHint); ok {
			return h.FormWidget(), true
		}
	}
	if t.Kind() != reflect.Pointer && reflect.PointerTo(t).Implements(hintType) {
		v := reflect.New(t)
		if h, ok := v.Interface().(mx.FormWidgetHint); ok {
			return h.FormWidget(), true
		}
	}
	return "", false
}

func fieldLabel(field reflect.StructField, tag mx.FormTag) string {
	if tag.Label != "" {
		return tag.Label
	}
	return field.Name
}

func isNullable(value reflect.Value) bool {
	if !value.IsValid() {
		return false
	}
	if value.CanInterface() {
		if _, ok := value.Interface().(mx.Nullable); ok {
			return true
		}
	}
	if value.CanAddr() {
		if _, ok := value.Addr().Interface().(mx.Nullable); ok {
			return true
		}
	}
	return false
}

func isTrueBool(value reflect.Value) bool {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return false
		}
		value = value.Elem()
	}
	return value.Kind() == reflect.Bool && value.Bool()
}

func clearSentinelInput(path mx.FieldPath) mx.Component {
	return Label(
		Input(
			Type("checkbox"),
			Name(mx.ClearSentinelName(path)),
			Value("1"),
		),
		mx.Text(" clear"),
	)
}

