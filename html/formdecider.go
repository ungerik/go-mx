package html

import (
	"cmp"
	"context"
	"encoding"
	"fmt"
	"net/http"
	"reflect"
	"sort"
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
		t := stringWidgetType(tag)
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
	in := mx.DateTimeInput(value, tag)
	attribs := []mx.Attrib{
		Type(in.Type),
		Name(string(path)),
		ID(string(path)),
	}
	if in.Value != "" {
		attribs = append(attribs, Value(in.Value))
	}
	// A text input shows stored text the picker cannot hold: it has no
	// date range, it says which format to enter instead, and it is invalid
	// from the start, because saving that text unchanged fails.
	textFallback := in.TextFallback()
	if textFallback {
		attribs = append(attribs,
			Placeholder(cmp.Or(tag.Placeholder, in.Format)),
			mx.NewAttrib("aria-describedby", formatHintID(path)),
		)
	} else {
		if tag.Min != "" {
			attribs = append(attribs, Min(tag.Min))
		}
		if tag.Max != "" {
			attribs = append(attribs, Max(tag.Max))
		}
	}
	if tag.Required {
		attribs = append(attribs, Required)
	}
	if tag.Readonly {
		attribs = append(attribs, Readonly)
	}
	if len(errs) > 0 || textFallback {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	if !textFallback {
		return Input(attribs...)
	}
	return mx.Components{
		Input(attribs...),
		mx.NewElement("small",
			mx.Attribute{Name: "class", Value: "form-help"},
			mx.Attribute{Name: "id", Value: formatHintID(path)},
			mx.Text("Expected format: "+in.Format),
		),
	}
}

// formatHintID is the id of the format hint under a date or time field
// that fell back to a text input. The lowercase suffix cannot collide with
// a field path, whose segments are exported Go field names.
func formatHintID(path mx.FieldPath) string {
	return string(path) + "-format"
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
	contextList := mx.OptionsNeedContext(tag, field.Type)
	options := optionChildren(value, tag, field.Type, func(opt mx.NamedOption) mx.Component {
		oAttr := []any{Value(opt.Value)}
		if opt.Value == selected {
			oAttr = append(oAttr, Selected)
		}
		oAttr = append(oAttr, mx.Text(displayLabel(opt)))
		return Option(oAttr...)
	}, func(options []mx.NamedOption) mx.Component {
		return selectPrefaceOption(mx.SelectPreface(selected, tag, contextList, options))
	})
	return Select(append(attribs, options)...)
}

// selectPrefaceOption renders the leading option [mx.SelectPreface] chose.
// Captions are generic on purpose: a filtered-out value is not echoed.
func selectPrefaceOption(preface mx.OptionPreface) mx.Component {
	switch preface {
	case mx.OptionPrefaceUnavailable:
		return Option(Value(""), Disabled, Selected, mx.Text("(current value not available)"))
	case mx.OptionPrefaceChoose:
		return Option(Value(""), Disabled, Selected, mx.Text("Choose…"))
	case mx.OptionPrefaceEmpty:
		return Option(Value(""), Selected, mx.Text("(none)"))
	case mx.OptionPrefaceClear:
		return Option(Value(""), mx.Text("(none)"))
	default:
		return nil
	}
}

func enumSetInput(path mx.FieldPath, value reflect.Value, field reflect.StructField, tag mx.FormTag, errs []error) mx.Component {
	keyType := setKeyType(field.Type)
	if keyType == nil {
		return mx.Text("[enum-set: cannot infer key type for " + string(path) + "]")
	}
	selectedSet := setSelectedValues(value)
	contextList := mx.OptionsNeedContext(tag, keyType)
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
	}, func(options []mx.NamedOption) mx.Component {
		preface := mx.EnumSetPreface(setMemberValues(value), tag, contextList, options)
		if preface != mx.OptionPrefaceValuesUnavailable {
			return nil
		}
		return mx.NewElement("p",
			mx.Attribute{Name: "class", Value: "form-help"},
			mx.Text("(some current values are not available)"),
		)
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
		return setTime(value, raw, tag)
	case mx.FieldKindEnum:
		raw := form.Get(string(path))
		// A required select's empty option must not become a zero non-string
		// value. isEffectivelyEmpty never treats a number as empty, so the
		// handler's required check would accept the zero of an int enum.
		if raw == "" && tag.Required {
			return fmt.Errorf("required")
		}
		if raw != "" {
			if err := validateOptionMembership(r.Context(), value, tag, field.Type, raw); err != nil {
				return err
			}
		}
		return setScalar(value, raw)
	case mx.FieldKindEnumSet:
		vals := form[string(path)]
		if keyType := setKeyType(field.Type); keyType != nil && mx.OptionsNeedContext(tag, keyType) {
			// The list is the trust boundary for the set, same as a select.
			// Submitted values outside it are rejected. Members the list
			// does not offer are not kept: the checkboxes the user saw are
			// the whole set that a save stores, so an uncheck cannot be
			// undone by the list changing before POST.
			options, err := mx.CollectOptions(r.Context(), reflect.Value{}, tag, keyType)
			if err != nil {
				return err
			}
			if err := rejectUnlistedOptions(options, vals); err != nil {
				return err
			}
		}
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

// validateOptionMembership rejects submitted values that are not in
// the field's per-request option list. It only applies to fields whose
// options need the request context (form:"options=…" registry entries,
// [mx.NamedOptionsContextProvider] types): for those the option list
// is the trust boundary the form advertises — a tenant-scoped dropdown
// must not bind another tenant's ID on POST just because the browser
// was told to only offer this tenant's. Static enum kinds are checked by
// the validation chain instead: a single value by its type's Validate
// (generated by go-enum), an enum set element by element.
//
// The rejection is a plain validation error (rendered inline at the
// field); a failure to resolve the option list itself propagates as-is.
func validateOptionMembership(ctx context.Context, value reflect.Value, tag mx.FormTag, t reflect.Type, vals ...string) error {
	if len(vals) == 0 || !mx.OptionsNeedContext(tag, t) {
		return nil
	}
	options, err := mx.CollectOptions(ctx, value, tag, t)
	if err != nil {
		return err
	}
	return rejectUnlistedOptions(options, vals)
}

// rejectUnlistedOptions is the membership check itself, split out so an
// enum set can collect the list once and both reject new values and keep
// stored ones that were not rendered.
func rejectUnlistedOptions(options []mx.NamedOption, vals []string) error {
	if len(vals) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(options))
	for _, opt := range options {
		allowed[opt.Value] = struct{}{}
	}
	for _, v := range vals {
		if _, ok := allowed[v]; !ok {
			return fmt.Errorf("%q is not one of the allowed options", v)
		}
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
	// An empty submission is "no choice", including the optional select's
	// "(none)" option. A non-string TextUnmarshaler must not see "" — that
	// is not a value it can parse, and the zero value is the clear.
	if raw == "" {
		value.SetZero()
		return nil
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

func setTime(value reflect.Value, raw string, tag mx.FormTag) error {
	if derefType(value.Type()).Kind() == reflect.String {
		return setTimeText(value, raw, tag)
	}
	if value.Kind() == reflect.Pointer {
		if raw == "" {
			value.SetZero()
			return nil
		}
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setTime(value.Elem(), raw, tag)
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
	switch tag.Widget {
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

// setTimeText binds a submitted date or time into a string-kind field
// (through any pointers) that holds it as text, the counterpart of the
// text values [mx.DateTimeInput] prefills. [mx.ParseDateTimeInput] checks
// the submission and gives the text to store. An empty submission stores
// the zero value, which a required field rejects. A valid value goes
// through UnmarshalText when the type has it and is set as the string
// otherwise.
//
// A rejected submission is a validation error, and the field keeps the
// submitted text so every re-render shows what was posted: a repeatable
// row is bound from the submission alone, so an add-row click would
// otherwise blank the stored text and the next save would store "". The
// form handler never calls onSubmit while a field has an error, so the
// rejected text is not saved.
func setTimeText(value reflect.Value, raw string, tag mx.FormTag) error {
	text, err := mx.ParseDateTimeInput(tag, raw)
	if err != nil {
		setDerefString(value, raw)
		return err
	}
	if text == "" {
		value.SetZero()
		return nil
	}
	parsed := reflect.New(derefType(value.Type()))
	if u, ok := parsed.Interface().(encoding.TextUnmarshaler); ok {
		if err := u.UnmarshalText([]byte(text)); err != nil {
			setDerefString(value, raw)
			return err
		}
	} else {
		parsed.Elem().SetString(text)
	}
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}
	value.Set(parsed.Elem())
	return nil
}

// setDerefString sets the string-kind value behind any pointers to s,
// allocating nil pointers.
func setDerefString(value reflect.Value, s string) {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}
	value.SetString(s)
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

// derefType returns t with any pointers removed.
func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
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
	// TextUnmarshaler before the string-kind assignment. A named string
	// whose MarshalText is "id:listed" must store "listed", not the prefixed
	// form, or the next render no longer matches the option.
	if u, ok := dst.Addr().Interface().(encoding.TextUnmarshaler); ok {
		if err := u.UnmarshalText([]byte(s)); err != nil {
			return reflect.Value{}, err
		}
		return dst, nil
	}
	if t.Kind() == reflect.String {
		dst.SetString(s)
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
	members := setMemberValues(value)
	out := make(map[string]struct{}, len(members))
	for _, member := range members {
		out[member] = struct{}{}
	}
	return out
}

// setMemberValues returns each set member in [mx.FormatOptionValue], the
// same form Enums() options use. Map keys are sorted so the order is
// stable; slice order is kept.
func setMemberValues(value reflect.Value) []string {
	for value.Kind() == reflect.Pointer {
		if !value.IsValid() || value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil
	}
	switch value.Kind() {
	case reflect.Map:
		out := make([]string, 0, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			out = append(out, mx.FormatOptionValue(iter.Key()))
		}
		sort.Strings(out)
		return out
	case reflect.Slice:
		out := make([]string, 0, value.Len())
		for i := range value.Len() {
			out = append(out, mx.FormatOptionValue(value.Index(i)))
		}
		return out
	default:
		return nil
	}
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
//
// prepend, when non-nil, is called with the resolved option list and its
// result (if non-nil) is rendered before the options. Selects use it for
// the leading empty option ([mx.SelectPreface]); enum sets use it for the
// hidden-members note.
func optionChildren(value reflect.Value, tag mx.FormTag, t reflect.Type, render func(opt mx.NamedOption) mx.Component, prepend func(options []mx.NamedOption) mx.Component) mx.Component {
	assemble := func(options []mx.NamedOption) mx.Components {
		children := make(mx.Components, 0, len(options)+1)
		if prepend != nil {
			if first := prepend(options); first != nil {
				children = append(children, first)
			}
		}
		for _, opt := range options {
			children = append(children, render(opt))
		}
		return children
	}
	if !mx.OptionsNeedContext(tag, t) {
		options, err := mx.CollectOptions(context.Background(), value, tag, t)
		if err != nil {
			return mx.NewErrElement(err)
		}
		return assemble(options)
	}
	return mx.ComponentFunc(func(ctx context.Context, w mx.Writer) error {
		options, err := mx.CollectOptions(ctx, value, tag, t)
		if err != nil {
			return err
		}
		return assemble(options).Render(ctx, w)
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
	// The zero value of a fixed-size ID is what an empty submission stores
	// (setScalar), so it is no value, not 00000000-0000-0000-0000-000000000000.
	if v.Kind() == reflect.Array && v.IsZero() {
		return ""
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

// stringWidgetType returns the input type of a string-kind field. [mx.DetectField]
// has already put a [mx.FormWidgetHint]'s widget into tag, so the tag is
// the only place to look.
func stringWidgetType(tag mx.FormTag) string {
	switch tag.Widget {
	case "email", "url", "tel", "password":
		return tag.Widget
	}
	return "text"
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
