package shadcn

import (
	"cmp"
	"context"
	"encoding"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/hx"
)

// FieldDecider is the Tailwind/shadcn layer of the layered form
// rendering chain. It implements [mx.FieldDecider] by emitting
// shadcn-styled components ([Input], [Switch], [Select], [Textarea],
// [Card]) laid out with the [Field] system ([Field], [FieldLabel],
// [FieldDescription], [FieldError]) for every kind it recognizes, and
// delegating to [hx.FieldDecider] (and through it [html.FieldDecider])
// for anything it does not. Parsing is always delegated upward — shadcn
// is purely a rendering concern.
//
// On validation failure the control receives `aria-invalid="true"` (its
// class string turns the border destructive), the enclosing [Field] gets
// `data-invalid="true"` (destructive field text) and each message renders
// as a [FieldError] with `role="alert"`.
var FieldDecider mx.FieldDecider = func(path mx.FieldPath, field reflect.StructField, value reflect.Value) mx.FieldBehavior {
	kind, tag := mx.DetectField(path, field, value)
	base := hx.FieldDecider(path, field, value)

	switch kind {
	case mx.FieldKindSkip, mx.FieldKindHidden:
		// Skip and hidden fields don't need styling.
		return base
	}

	render := func(path mx.FieldPath, field reflect.StructField, value reflect.Value, errs []error) mx.Component {
		if comp := renderShadcnField(path, field, value, kind, tag, errs); comp != nil {
			return comp
		}
		// Fall through if shadcn doesn't customize this kind.
		if base.Render == nil {
			return nil
		}
		return base.Render(path, field, value, errs)
	}
	return mx.FieldBehavior{
		Render:   render,
		Parse:    base.Parse,
		Validate: base.Validate,
	}
}

func renderShadcnField(path mx.FieldPath, field reflect.StructField, value reflect.Value, kind mx.FieldKind, tag mx.FormTag, errs []error) mx.Component {
	switch kind {
	case mx.FieldKindString, mx.FieldKindCatchAll:
		// A catch-all value, such as a registered money type, is a text
		// input through its MarshalText and UnmarshalText.
		return shadcnField(path, field, tag, errs,
			Input(stringAttribs(path, value, tag, errs)...),
		)
	case mx.FieldKindNumber:
		return shadcnField(path, field, tag, errs,
			Input(numberAttribs(path, value, tag, errs)...),
		)
	case mx.FieldKindBool:
		return shadcnBoolField(path, field, value, tag, errs)
	case mx.FieldKindDateTime:
		// Stored text the parser rejects marks the input invalid before any
		// submit; the Field root follows, so the label matches the control.
		invalid := len(errs) > 0 || (mx.DateTimeInput(value, tag).Invalid && !tag.Readonly)
		return shadcnFieldMarked(path, field, tag, errs, invalid,
			shadcnDateTime(path, value, tag, errs),
		)
	case mx.FieldKindTextarea:
		return shadcnField(path, field, tag, errs,
			Textarea(textareaAttribs(path, value, tag, errs)...),
		)
	case mx.FieldKindEnum:
		return shadcnField(path, field, tag, errs,
			shadcnEnum(path, field, value, tag, errs),
		)
	case mx.FieldKindEnumSet:
		return shadcnField(path, field, tag, errs,
			shadcnEnumSet(path, field, value, tag, errs),
		)
	case mx.FieldKindFile:
		return shadcnField(path, field, tag, errs,
			Input(fileAttribs(path, tag, errs)...),
		)
	}
	return nil
}

// shadcnField wraps a single input element in a [Field] stack:
//
//	<Field [data-invalid=true when errored]>
//	  <FieldLabel for=path>label</FieldLabel> (when set)
//	  <input ...>
//	  <FieldDescription>help</FieldDescription> (when set)
//	  <Label for=clear>clear</Label> (when nullable, not required and not readonly)
//	  <FieldError data-error=path>err</FieldError> (one per error, role="alert")
//	</Field>
//
// The error look keys off data-invalid on the [Field] root (destructive
// text) together with each [FieldError]; the control also carries
// aria-invalid from its own attribs so its border turns destructive.
func shadcnField(path mx.FieldPath, field reflect.StructField, tag mx.FormTag, errs []error, input mx.Component) mx.Component {
	return shadcnFieldMarked(path, field, tag, errs, len(errs) > 0, input)
}

// shadcnFieldMarked is [shadcnField] with the Field root marked
// data-invalid when invalid is true, which a control can be without an
// error, such as a text input holding stored text the parser rejects.
func shadcnFieldMarked(path mx.FieldPath, field reflect.StructField, tag mx.FormTag, errs []error, invalid bool, input mx.Component) mx.Component {
	parts := make([]any, 0, 4+len(errs))
	if labelText := fieldLabel(field, tag); labelText != "" {
		parts = append(parts, FieldLabelFor(string(path), labelText))
	}
	parts = append(parts, input)
	if tag.Help != "" {
		parts = append(parts, FieldDescription(mx.Text(tag.Help)))
	}
	// A readonly field is never written, so it offers no clear control.
	if isNullable(field, tag) && !tag.Readonly {
		parts = append(parts, clearControl(path))
	}
	parts = appendFieldErrors(parts, path, errs)
	return fieldWrap(invalid, parts...)
}

// fieldWrap builds the [Field] root for a decided form field, marking it
// data-invalid when invalid, as when there are validation errors, so the
// Field turns its text destructive and its label/error key off the same
// flag.
func fieldWrap(invalid bool, children ...any) mx.Component {
	args := make([]any, 0, len(children)+1)
	if invalid {
		args = append(args, html.DataAttr("invalid", "true"))
	}
	args = append(args, children...)
	return Field("", args...)
}

// appendFieldErrors appends one [FieldError] (role="alert") per error, each
// tagged data-error=path so a client can target a specific field's message
// (the same convention the html layer uses).
func appendFieldErrors(parts []any, path mx.FieldPath, errs []error) []any {
	for _, e := range errs {
		parts = append(parts, FieldError(
			mx.NewAttrib("data-error", string(path)),
			mx.Text(e.Error()),
		))
	}
	return parts
}

func shadcnBoolField(path mx.FieldPath, field reflect.StructField, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	attribs := []mx.Attrib{
		html.Name(string(path)),
		html.ID(string(path)),
		html.Value("on"),
	}
	if isTrueBool(value) {
		attribs = append(attribs, html.Checked)
	}
	if tag.Readonly {
		attribs = append(attribs, html.Disabled)
	}
	if len(errs) > 0 {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	attribs = append(attribs, hx.Trigger("change"))

	var input mx.Component
	if tag.Widget == "checkbox" {
		input = Checkbox(attribs...)
	} else {
		input = Switch(attribs...)
	}

	// Control and label sit side by side; description and errors stack below.
	row := mx.Components{input}
	if labelText := fieldLabel(field, tag); labelText != "" {
		row = append(row, FieldLabelFor(string(path), labelText))
	}
	parts := []any{
		mx.NewElement("div",
			mx.Attribute{Name: "class", Value: "flex items-center gap-2"},
			row,
		),
	}
	if tag.Help != "" {
		parts = append(parts, FieldDescription(mx.Text(tag.Help)))
	}
	parts = appendFieldErrors(parts, path, errs)
	return fieldWrap(len(errs) > 0, parts...)
}

func shadcnEnum(path mx.FieldPath, field reflect.StructField, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	selected := stringifyValue(value)
	attribs := []any{
		html.Name(string(path)),
		html.ID(string(path)),
		hx.Trigger("change"),
	}
	if tag.Required {
		attribs = append(attribs, html.Required)
	}
	if tag.Readonly {
		attribs = append(attribs, html.Disabled)
	}
	if len(errs) > 0 {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	contextList := mx.OptionsNeedContext(tag, field.Type)
	options := optionChildren(value, tag, field.Type, func(opt mx.NamedOption) mx.Component {
		oAttr := []any{html.Value(opt.Value)}
		if opt.Value == selected {
			oAttr = append(oAttr, html.Selected)
		}
		oAttr = append(oAttr, mx.Text(optionDisplay(opt)))
		return html.Option(oAttr...)
	}, func(options []mx.NamedOption) mx.Component {
		return selectPrefaceOption(mx.SelectPreface(selected, tag, contextList, options))
	})
	return Select(append(attribs, options)...)
}

// selectPrefaceOption renders the leading option [mx.SelectPreface] chose.
// Captions come from [Labels] and stay generic: a filtered-out value is
// not echoed.
func selectPrefaceOption(preface mx.OptionPreface) mx.Component {
	switch preface {
	case mx.OptionPrefaceUnavailable:
		return html.Option(html.Value(""), html.Disabled, html.Selected,
			labelText(func(l Labels) string { return l.FormValueNotAvailable }))
	case mx.OptionPrefaceChoose:
		return html.Option(html.Value(""), html.Disabled, html.Selected,
			labelText(func(l Labels) string { return l.FormChooseOption }))
	case mx.OptionPrefaceEmpty:
		return html.Option(html.Value(""), html.Selected,
			labelText(func(l Labels) string { return l.FormNoSelection }))
	case mx.OptionPrefaceClear:
		return html.Option(html.Value(""),
			labelText(func(l Labels) string { return l.FormNoSelection }))
	default:
		return nil
	}
}

func shadcnEnumSet(path mx.FieldPath, field reflect.StructField, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	keyType := setKeyType(field.Type)
	if keyType == nil {
		return mx.Text("[enum-set: cannot infer key type for " + string(path) + "]")
	}
	selected := setMembers(value)
	contextList := mx.OptionsNeedContext(tag, keyType)
	items := optionChildren(reflect.Value{}, tag, keyType, func(opt mx.NamedOption) mx.Component {
		attribs := []mx.Attrib{
			html.Name(string(path)),
			html.Value(opt.Value),
			hx.Trigger("change"),
		}
		if _, ok := selected[opt.Value]; ok {
			attribs = append(attribs, html.Checked)
		}
		if tag.Readonly {
			attribs = append(attribs, html.Disabled)
		}
		return mx.NewElement("div",
			mx.Attribute{Name: "class", Value: "flex items-center gap-2"},
			Checkbox(attribs...),
			LabelFor(string(path)+"-"+opt.Value,
				mx.Text(optionDisplay(opt))),
		)
	}, func(options []mx.NamedOption) mx.Component {
		members := make([]string, 0, len(selected))
		for member := range selected {
			members = append(members, member)
		}
		preface := mx.EnumSetPreface(members, tag, contextList, options)
		if preface != mx.OptionPrefaceValuesUnavailable {
			return nil
		}
		return FieldDescription(html.Class("col-span-2"),
			labelText(func(l Labels) string { return l.FormValuesNotAvailable }))
	})
	wrapAttribs := []any{
		mx.Attribute{Name: "role", Value: "group"},
		mx.Attribute{Name: "class", Value: "grid grid-cols-2 gap-2"},
	}
	if len(errs) > 0 {
		wrapAttribs = append(wrapAttribs, mx.NewAttrib("aria-invalid", "true"))
	}
	wrapAttribs = append(wrapAttribs, items)
	return mx.NewElement("div", wrapAttribs...)
}

// Helpers below are intentionally narrow re-implementations of the
// html package's display/parse helpers — the dispatch table is small
// enough that calling private html helpers across packages would be
// uglier than this short duplication.

func stringAttribs(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) []mx.Attrib {
	inputType := stringWidget(tag)
	a := []mx.Attrib{
		html.Type(inputType),
		html.Name(string(path)),
		html.ID(string(path)),
		hx.Trigger("change"),
	}
	if val := displayValue(value); val != "" && !tag.Sensitive {
		a = append(a, html.Value(val))
	}
	if tag.Placeholder != "" {
		a = append(a, html.Placeholder(tag.Placeholder))
	}
	if tag.Pattern != "" {
		a = append(a, html.Pattern(tag.Pattern))
	}
	if tag.Required {
		a = append(a, html.Required)
	}
	if tag.Readonly {
		a = append(a, html.Readonly)
	}
	if len(errs) > 0 {
		a = append(a, mx.NewAttrib("aria-invalid", "true"))
	}
	return a
}

func numberAttribs(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) []mx.Attrib {
	a := []mx.Attrib{
		html.Type("number"),
		html.Name(string(path)),
		html.ID(string(path)),
		hx.Trigger("change"),
	}
	// A sensitive number never shows its value, like a sensitive text input.
	if !tag.Sensitive {
		a = append(a, html.Value(displayValue(value)))
	}
	if tag.Min != "" {
		a = append(a, html.Min(tag.Min))
	}
	if tag.Max != "" {
		a = append(a, html.Max(tag.Max))
	}
	if tag.Step != "" {
		a = append(a, html.Step(tag.Step))
	}
	if tag.Required {
		a = append(a, html.Required)
	}
	if tag.Readonly {
		a = append(a, html.Readonly)
	}
	if len(errs) > 0 {
		a = append(a, mx.NewAttrib("aria-invalid", "true"))
	}
	return a
}

// shadcnDateTime renders the input [mx.DateTimeInput] decides. When the
// stored text is no value the picker can hold, the input is a text input
// followed by a [FieldDescription] naming the format to enter.
func shadcnDateTime(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	in := mx.DateTimeInput(value, tag)
	a := []mx.Attrib{
		html.Type(in.Type),
		html.Name(string(path)),
		html.ID(string(path)),
		hx.Trigger("change"),
	}
	if in.Value != "" {
		a = append(a, html.Value(in.Value))
	}
	// A readonly field is never saved, so its text input asks for nothing.
	textFallback := in.TextFallback()
	askFormat := textFallback && !tag.Readonly
	if askFormat {
		a = append(a,
			html.Placeholder(cmp.Or(tag.Placeholder, in.Format)),
			mx.NewAttrib("aria-describedby", formatHintID(path)),
		)
	}
	if !textFallback {
		if tag.Min != "" {
			a = append(a, html.Min(tag.Min))
		}
		if tag.Max != "" {
			a = append(a, html.Max(tag.Max))
		}
		if in.Step != "" {
			a = append(a, html.Step(in.Step))
		}
	}
	if tag.Required {
		a = append(a, html.Required)
	}
	if tag.Readonly {
		a = append(a, html.Readonly)
	}
	// Text the parser rejects is invalid from the start: saving it
	// unchanged fails.
	if len(errs) > 0 || (in.Invalid && !tag.Readonly) {
		a = append(a, mx.NewAttrib("aria-invalid", "true"))
	}
	// The hint names the format the parser reads, which a placeholder from
	// the tag does not replace, and says so when the stored text is not in
	// it.
	// After a submit, the field error says that already.
	caption := func(l Labels) string { return l.FormExpectedFormat }
	if in.Invalid && len(errs) == 0 {
		caption = func(l Labels) string { return l.FormInvalidFormat }
	}
	if !askFormat {
		return Input(a...)
	}
	return mx.Components{
		Input(a...),
		FieldDescription(html.ID(formatHintID(path)),
			labelText(caption),
			mx.Text(" "+in.Format),
		),
	}
}

// formatHintID is the id of the format hint under a date or time field
// that fell back to a text input. The lowercase suffix cannot collide with
// a field path, whose segments are exported Go field names.
func formatHintID(path mx.FieldPath) string {
	return string(path) + "-format"
}

func textareaAttribs(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) []any {
	a := []any{
		html.Name(string(path)),
		html.ID(string(path)),
		hx.Trigger("change"),
	}
	if tag.Placeholder != "" {
		a = append(a, html.Placeholder(tag.Placeholder))
	}
	if tag.Required {
		a = append(a, html.Required)
	}
	if tag.Readonly {
		a = append(a, html.Readonly)
	}
	if len(errs) > 0 {
		a = append(a, mx.NewAttrib("aria-invalid", "true"))
	}
	if val := displayValue(value); val != "" && !tag.Sensitive {
		a = append(a, mx.Text(val))
	}
	return a
}

func fileAttribs(path mx.FieldPath, tag mx.FormTag, errs []error) []mx.Attrib {
	a := []mx.Attrib{
		html.Type("file"),
		html.Name(string(path)),
		html.ID(string(path)),
	}
	// A file input cannot be readonly, and a readonly field is never
	// saved, so it is disabled and asks for nothing.
	switch {
	case tag.Readonly:
		a = append(a, html.Disabled)
	case tag.Required:
		a = append(a, html.Required)
	}
	if len(errs) > 0 {
		a = append(a, mx.NewAttrib("aria-invalid", "true"))
	}
	return a
}

func clearControl(path mx.FieldPath) mx.Component {
	return LabelFor(mx.ClearSentinelName(path),
		Checkbox(
			html.Name(mx.ClearSentinelName(path)),
			html.Value("1"),
		),
		mx.NewElement("span",
			mx.Attribute{Name: "class", Value: "ml-2 text-sm"},
			labelText(func(l Labels) string { return l.FormClear }),
		),
	)
}

// --- shared local helpers (duplicated from html.formdecider, narrow on purpose) ---

// stringWidget returns the input type of a string-kind field. [mx.DetectField]
// has already put a [mx.FormWidgetHint]'s widget into tag, so the tag is
// the only place to look.
func stringWidget(tag mx.FormTag) string {
	switch tag.Widget {
	case "email", "url", "tel", "password":
		return tag.Widget
	}
	return "text"
}

func displayValue(v reflect.Value) (text string) {
	// A method that panics, as the promoted MarshalText of a nil embedded
	// pointer does, shows no value instead of failing the render.
	defer func() {
		if recover() != nil {
			text = ""
		}
	}()
	if !v.IsValid() {
		return ""
	}
	if v.CanInterface() && mx.IsNull(v.Interface()) {
		return ""
	}
	x := v
	for x.Kind() == reflect.Pointer {
		if x.IsNil() {
			return ""
		}
		x = x.Elem()
	}
	// The zero value of a fixed-size ID is what an empty submission stores,
	// so it is no value, not 00000000-0000-0000-0000-000000000000.
	if x.Kind() == reflect.Array && x.IsZero() {
		return ""
	}
	if t, ok := x.Interface().(time.Time); ok {
		return t.Format("2006-01-02T15:04:05")
	}
	if x.CanInterface() {
		if tm, ok := x.Interface().(encoding.TextMarshaler); ok {
			if b, err := tm.MarshalText(); err == nil {
				return string(b)
			}
		}
	}
	if x.CanAddr() {
		if tm, ok := x.Addr().Interface().(encoding.TextMarshaler); ok {
			if b, err := tm.MarshalText(); err == nil {
				return string(b)
			}
		}
	}
	switch x.Kind() {
	case reflect.String:
		return x.String()
	case reflect.Bool:
		return strconv.FormatBool(x.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(x.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(x.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(x.Float(), 'f', -1, x.Type().Bits())
	case reflect.Slice:
		if x.Type().Elem().Kind() == reflect.String {
			parts := make([]string, x.Len())
			for i := range x.Len() {
				parts[i] = x.Index(i).String()
			}
			return strings.Join(parts, "\n")
		}
		if x.Type().Elem().Kind() == reflect.Uint8 {
			return string(x.Bytes())
		}
	}
	return fmt.Sprint(x.Interface())
}

func stringifyValue(v reflect.Value) string {
	return displayValue(v)
}

func isNullable(field reflect.StructField, tag mx.FormTag) bool {
	if tag.Required {
		return false
	}
	t := field.Type
	return t.Implements(reflect.TypeFor[mx.Nullable]()) ||
		reflect.PointerTo(t).Implements(reflect.TypeFor[mx.Nullable]())
}

func isTrueBool(v reflect.Value) bool {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	return v.Kind() == reflect.Bool && v.Bool()
}

func fieldLabel(field reflect.StructField, tag mx.FormTag) string {
	if tag.Label != "" {
		return tag.Label
	}
	return field.Name
}

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

func setMembers(v reflect.Value) map[string]struct{} {
	out := map[string]struct{}{}
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return out
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			out[mx.FormatOptionValue(iter.Key())] = struct{}{}
		}
	case reflect.Slice:
		for i := range v.Len() {
			out[mx.FormatOptionValue(v.Index(i))] = struct{}{}
		}
	}
	return out
}

// optionChildren resolves the field's option list via
// [mx.CollectOptions] and renders each entry through render. Static
// option lists are resolved immediately; option lists that need the
// request context (form:"options=…" registry entries,
// [mx.NamedOptionsContextProvider] types) are collected when the
// component renders, where the context passed to Render is the real
// request context. A collection error surfaces at render time
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

func optionDisplay(o mx.NamedOption) string {
	if o.Name != "" {
		return o.Name
	}
	return o.Value
}

// SectionCard wraps the per-section group of fields in a shadcn Card.
// Form handlers (or higher-level helpers) that wish to group fields
// using section tags can call SectionCard(name, fields...) — it
// matches the layout the design's screenshots expect (card with
// CardHeader > CardTitle name, then CardContent children).
func SectionCard(name string, children ...mx.Component) mx.Component {
	body := make([]any, 0, len(children))
	for _, c := range children {
		body = append(body, c)
	}
	return Card(
		CardHeader(CardTitle(name)),
		CardContent(body...),
	)
}

// _ keeps net/http imported for future expansion (we currently only
// need it transitively but the linter would complain on a removal).
var _ = http.MethodGet
