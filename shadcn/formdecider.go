package shadcn

import (
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
// [Card]) for every kind it recognizes and delegating to
// [hx.FieldDecider] (and through it [html.FieldDecider]) for anything
// it does not. Parsing is always delegated upward — shadcn is purely a
// rendering concern.
//
// On validation failure the rendered element receives
// `aria-invalid="true"`, which the shadcn class strings already turn
// into a visible destructive border.
var FieldDecider mx.FieldDecider = func(path mx.FieldPath, field reflect.StructField, value reflect.Value) mx.FieldBehavior {
	kind, tag := mx.DetectField(path, field, value)
	base := hx.FieldDecider(path, field, value)

	switch kind {
	case mx.FieldKindSkip, mx.FieldKindHidden, mx.FieldKindCatchAll:
		// Skip and hidden fields don't need styling; the catch-all
		// path is the unstyled fallback by definition.
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
	case mx.FieldKindString:
		return shadcnField(path, field, tag, errs,
			Input(stringAttribs(path, value, field.Type, tag, errs)...),
		)
	case mx.FieldKindNumber:
		return shadcnField(path, field, tag, errs,
			Input(numberAttribs(path, value, tag, errs)...),
		)
	case mx.FieldKindBool:
		return shadcnBoolField(path, field, value, tag, errs)
	case mx.FieldKindDateTime:
		return shadcnField(path, field, tag, errs,
			Input(datetimeAttribs(path, value, tag, errs)...),
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

// shadcnField wraps a single input element in a labeled stack:
//   <Label for=path>label</Label>
//   <input ...>
//   <small>help</small> (when set)
//   <Label for=clear>clear</Label> (when nullable)
//   <p data-error>err</p> (one per error)
func shadcnField(path mx.FieldPath, field reflect.StructField, tag mx.FormTag, errs []error, input mx.Component) mx.Component {
	parts := mx.Components{}
	if labelText := fieldLabel(field, tag); labelText != "" {
		parts = append(parts, Label(html.For(string(path)), labelText))
	}
	parts = append(parts, input)
	if tag.Help != "" {
		parts = append(parts, mx.NewElement("small",
			mx.Attribute{Name: "class", Value: "text-muted-foreground text-sm"},
			mx.Text(tag.Help),
		))
	}
	if isNullable(field, tag) {
		parts = append(parts, clearControl(path))
	}
	for _, e := range errs {
		parts = append(parts, mx.NewElement("p",
			mx.Attribute{Name: "class", Value: "text-destructive text-sm"},
			mx.Attribute{Name: "data-error", Value: string(path)},
			mx.Text(e.Error()),
		))
	}
	return mx.NewElement("div",
		mx.Attribute{Name: "class", Value: "grid gap-1.5"},
		parts,
	)
}

func shadcnBoolField(path mx.FieldPath, field reflect.StructField, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	attribs := []any{
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

	row := mx.Components{input}
	if labelText := fieldLabel(field, tag); labelText != "" {
		row = append(row, Label(html.For(string(path)),
			mx.NewElement("span",
				mx.Attribute{Name: "class", Value: "ml-2"},
				mx.Text(labelText),
			),
		))
	}
	parts := mx.Components{
		mx.NewElement("div",
			mx.Attribute{Name: "class", Value: "flex items-center"},
			row,
		),
	}
	if tag.Help != "" {
		parts = append(parts, mx.NewElement("small",
			mx.Attribute{Name: "class", Value: "text-muted-foreground text-sm"},
			mx.Text(tag.Help),
		))
	}
	for _, e := range errs {
		parts = append(parts, mx.NewElement("p",
			mx.Attribute{Name: "class", Value: "text-destructive text-sm"},
			mx.Attribute{Name: "data-error", Value: string(path)},
			mx.Text(e.Error()),
		))
	}
	return mx.NewElement("div",
		mx.Attribute{Name: "class", Value: "grid gap-1.5"},
		parts,
	)
}

func shadcnEnum(path mx.FieldPath, field reflect.StructField, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	options := collectOptionsForField(field.Type, value)
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
	for _, opt := range options {
		oAttr := []any{html.Value(opt.Value)}
		if opt.Value == selected {
			oAttr = append(oAttr, html.Selected)
		}
		oAttr = append(oAttr, mx.Text(optionDisplay(opt)))
		attribs = append(attribs, html.Option(oAttr...))
	}
	return Select(attribs...)
}

func shadcnEnumSet(path mx.FieldPath, field reflect.StructField, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	keyType := setKeyType(field.Type)
	if keyType == nil {
		return mx.Text("[enum-set: cannot infer key type for " + string(path) + "]")
	}
	options := collectOptionsForField(keyType, reflect.New(keyType).Elem())
	selected := setMembers(value)
	items := mx.Components{}
	for _, opt := range options {
		attribs := []any{
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
		items = append(items, mx.NewElement("div",
			mx.Attribute{Name: "class", Value: "flex items-center gap-2"},
			Checkbox(attribs...),
			Label(html.For(string(path)+"-"+opt.Value),
				mx.Text(optionDisplay(opt))),
		))
	}
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

func stringAttribs(path mx.FieldPath, value reflect.Value, t reflect.Type, tag mx.FormTag, errs []error) []any {
	inputType := stringWidget(tag, t)
	a := []any{
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

func numberAttribs(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) []any {
	a := []any{
		html.Type("number"),
		html.Name(string(path)),
		html.ID(string(path)),
		html.Value(displayValue(value)),
		hx.Trigger("change"),
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

func datetimeAttribs(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) []any {
	t := "datetime-local"
	switch tag.Widget {
	case "date":
		t = "date"
	case "time":
		t = "time"
	}
	a := []any{
		html.Type(t),
		html.Name(string(path)),
		html.ID(string(path)),
		hx.Trigger("change"),
	}
	if val := timeDisplay(value, t); val != "" {
		a = append(a, html.Value(val))
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

func fileAttribs(path mx.FieldPath, tag mx.FormTag, errs []error) []any {
	a := []any{
		html.Type("file"),
		html.Name(string(path)),
		html.ID(string(path)),
	}
	if tag.Required {
		a = append(a, html.Required)
	}
	if len(errs) > 0 {
		a = append(a, mx.NewAttrib("aria-invalid", "true"))
	}
	return a
}

func clearControl(path mx.FieldPath) mx.Component {
	return Label(html.For(mx.ClearSentinelName(path)),
		Checkbox(
			html.Name(mx.ClearSentinelName(path)),
			html.Value("1"),
		),
		mx.NewElement("span",
			mx.Attribute{Name: "class", Value: "ml-2 text-sm"},
			mx.Text("clear"),
		),
	)
}

// --- shared local helpers (duplicated from html.formdecider, narrow on purpose) ---

func stringWidget(tag mx.FormTag, t reflect.Type) string {
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

func displayValue(v reflect.Value) string {
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

func timeDisplay(v reflect.Value, inputType string) string {
	if !v.IsValid() {
		return ""
	}
	x := v
	for x.Kind() == reflect.Pointer {
		if x.IsNil() {
			return ""
		}
		x = x.Elem()
	}
	t, ok := x.Interface().(time.Time)
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
			out[fmt.Sprint(iter.Key().Interface())] = struct{}{}
		}
	case reflect.Slice:
		for i := range v.Len() {
			out[fmt.Sprint(v.Index(i).Interface())] = struct{}{}
		}
	}
	return out
}

func collectOptionsForField(t reflect.Type, value reflect.Value) []mx.NamedOption {
	probe := func(iface any) []mx.NamedOption {
		if iface == nil {
			return nil
		}
		if np, ok := iface.(mx.NamedOptionsProvider); ok {
			return np.NamedOptions()
		}
		if op, ok := iface.(mx.OptionsProvider); ok {
			out := make([]mx.NamedOption, len(op.Options()))
			for i, s := range op.Options() {
				out[i] = mx.NamedOption{Name: s, Value: s}
			}
			return out
		}
		v := reflect.ValueOf(iface)
		if v.IsValid() {
			if m := v.MethodByName("EnumStrings"); m.IsValid() {
				ret := m.Call(nil)
				if len(ret) == 1 && ret[0].Kind() == reflect.Slice && ret[0].Type().Elem().Kind() == reflect.String {
					out := make([]mx.NamedOption, ret[0].Len())
					for i := range ret[0].Len() {
						s := ret[0].Index(i).String()
						out[i] = mx.NamedOption{Name: s, Value: s}
					}
					return out
				}
			}
			if m := v.MethodByName("Enums"); m.IsValid() {
				ret := m.Call(nil)
				if len(ret) == 1 && ret[0].Kind() == reflect.Slice {
					out := make([]mx.NamedOption, ret[0].Len())
					for i := range ret[0].Len() {
						s := fmt.Sprint(ret[0].Index(i).Interface())
						out[i] = mx.NamedOption{Name: s, Value: s}
					}
					return out
				}
			}
		}
		return nil
	}
	if value.IsValid() && value.CanAddr() {
		if opts := probe(value.Addr().Interface()); opts != nil {
			return opts
		}
	}
	if value.IsValid() && value.CanInterface() {
		if opts := probe(value.Interface()); opts != nil {
			return opts
		}
	}
	fresh := reflect.New(t).Elem()
	if fresh.CanInterface() {
		if opts := probe(fresh.Interface()); opts != nil {
			return opts
		}
	}
	if t.Kind() != reflect.Pointer {
		ptr := reflect.New(t)
		if opts := probe(ptr.Interface()); opts != nil {
			return opts
		}
	}
	return nil
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
