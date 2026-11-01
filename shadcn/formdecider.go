package shadcn

import (
	"context"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"

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
		// A render probe asks only whether the field is in the form, which
		// base answers for every kind shadcn renders, without markup.
		if mx.IsRenderProbe(errs) {
			if base.Render == nil {
				return nil
			}
			return base.Render(path, field, value, errs)
		}
		if comp := renderShadcnField(path, field, value, kind, tag, errs); comp != nil {
			return comp
		}
		// Fall through if shadcn doesn't customize this kind.
		if base.Render == nil {
			return nil
		}
		return base.Render(path, field, value, errs)
	}
	// Parsing, validation and Multipart stay base's; only Render is
	// shadcn's.
	beh := base
	beh.Render = render
	return beh
}

func renderShadcnField(path mx.FieldPath, field reflect.StructField, value reflect.Value, kind mx.FieldKind, tag mx.FormTag, errs []error) mx.Component {
	// The input renders without required when a sensitive value is stored,
	// as in the html renderer; the label and clear control keep the tag.
	inputTag := mx.BrowserTag(tag, kind, value)
	switch kind {
	case mx.FieldKindString, mx.FieldKindCatchAll:
		// A catch-all value, such as a registered money type, is a text
		// input through its MarshalText and UnmarshalText.
		return shadcnField(path, field, value, kind, tag, errs,
			Input(stringAttribs(path, value, inputTag, errs)...),
		)
	case mx.FieldKindNumber:
		return shadcnField(path, field, value, kind, tag, errs,
			Input(numberAttribs(path, value, inputTag, errs)...),
		)
	case mx.FieldKindBool:
		return shadcnBoolField(path, field, value, tag, errs)
	case mx.FieldKindDateTime:
		// Stored text the parser rejects marks the input invalid before any
		// submit; the Field root follows, so the label matches the control.
		if err := mx.DateTimeTagError(value, tag); err != nil {
			return mx.NewErrElement(err)
		}
		in := mx.DateTimeInput(value, tag)
		invalid := len(errs) > 0 || (in.Invalid && !tag.Readonly)
		return shadcnFieldMarked(path, field, value, kind, tag, errs, invalid,
			shadcnDateTime(path, in, inputTag, errs),
		)
	case mx.FieldKindTextarea:
		return shadcnField(path, field, value, kind, tag, errs,
			Textarea(textareaAttribs(path, value, inputTag, errs)...),
		)
	case mx.FieldKindEnum:
		return shadcnField(path, field, value, kind, tag, errs,
			shadcnEnum(path, field, value, tag, errs),
		)
	case mx.FieldKindEnumSet:
		return shadcnField(path, field, value, kind, tag, errs,
			shadcnEnumSet(path, field, value, tag, errs),
		)
	case mx.FieldKindFile:
		return shadcnField(path, field, value, kind, tag, errs, shadcnFile(path, value, inputTag, errs))
	}
	return nil
}

// shadcnField wraps a single input element in a [Field] stack:
//
//	<Field [data-invalid=true when errored]>
//	  <FieldLabel for=path>label</FieldLabel> (when set)
//	  <input ...>
//	  <FieldDescription>help</FieldDescription> (when set)
//	  <Label for=clear>clear</Label> (when nullable or sensitive, not required and not readonly)
//	  <FieldError data-error=path>msg</FieldError> (one per message of each error, role="alert")
//	</Field>
//
// The error look keys off data-invalid on the [Field] root (destructive
// text) together with each [FieldError]; the control also carries
// aria-invalid from its own attribs so its border turns destructive.
func shadcnField(path mx.FieldPath, field reflect.StructField, value reflect.Value, kind mx.FieldKind, tag mx.FormTag, errs []error, input mx.Component) mx.Component {
	return shadcnFieldMarked(path, field, value, kind, tag, errs, len(errs) > 0, input)
}

// shadcnFieldMarked is [shadcnField] with the Field root marked
// data-invalid when invalid is true, which a control can be without an
// error, such as a text input holding stored text the parser rejects.
func shadcnFieldMarked(path mx.FieldPath, field reflect.StructField, value reflect.Value, kind mx.FieldKind, tag mx.FormTag, errs []error, invalid bool, input mx.Component) mx.Component {
	parts := make([]any, 0, 4+len(errs))
	if labelText := mx.FieldLabel(field, tag); labelText != "" && kind == mx.FieldKindEnumSet {
		// A set has no control with the path as its id: the label names
		// its group.
		parts = append(parts, FieldLabel(html.ID(groupLabelID(path)), labelText))
	} else if labelText != "" {
		parts = append(parts, FieldLabelFor(string(path), labelText))
	}
	parts = append(parts, input)
	// The control is described by its help and each error message, after
	// what it refers to itself, such as a date's format hint.
	var describedBy []string
	if tag.Help != "" {
		describedBy = append(describedBy, string(path)+"-help")
		parts = append(parts, FieldDescription(html.ID(string(path)+"-help"), mx.Text(tag.Help)))
	}
	// A sensitive input never shows its value, so it says that one is
	// stored and that a blank keeps it.
	if !tag.Readonly && mx.KeepsStoredValue(tag, kind, value) {
		describedBy = append(describedBy, string(path)+"-stored")
		parts = append(parts, FieldDescription(html.ID(string(path)+"-stored"),
			labelText(func(l Labels) string { return l.FormKeepStoredValue })))
	}
	// A readonly field is never written, so it offers no clear control. A
	// blank sensitive input keeps the stored value, so the clear control is
	// how it is emptied, ticked while it is empty so a re-render keeps a
	// clear the user asked for.
	if !tag.Readonly && isNullable(field, kind, tag) {
		parts = append(parts, clearControl(path, mx.FieldLabel(field, tag), tag.Sensitive && kind.KeepsStoredWhenBlank() && !mx.KeepsStoredValue(tag, kind, value)))
	}
	parts, describedBy = appendFieldErrors(parts, describedBy, path, errs)
	mx.AddDescribedBy(input, describedBy...)
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

// appendFieldErrors appends one [FieldError] (role="alert") per message of
// each error ([mx.FieldErrorTexts]), each tagged data-error=path so a client
// can target a specific field's message (the same convention the html layer
// uses).
func appendFieldErrors(parts []any, describedBy []string, path mx.FieldPath, errs []error) ([]any, []string) {
	firstError := len(describedBy)
	for _, e := range errs {
		// A date or time error is translated from its fields (Labels).
		if dateTimeErr, ok := e.(*mx.DateTimeError); ok {
			id := string(path) + "-error-" + strconv.Itoa(len(describedBy)-firstError)
			describedBy = append(describedBy, id)
			parts = append(parts, FieldError(
				html.ID(id),
				mx.NewAttrib("data-error", string(path)),
				labelText(func(l Labels) string { return l.dateTimeError(dateTimeErr) }),
			))
			continue
		}
		for _, text := range mx.FieldErrorTexts(e) {
			id := string(path) + "-error-" + strconv.Itoa(len(describedBy)-firstError)
			describedBy = append(describedBy, id)
			parts = append(parts, FieldError(
				html.ID(id),
				mx.NewAttrib("data-error", string(path)),
				mx.Text(text),
			))
		}
	}
	return parts, describedBy
}

func shadcnBoolField(path mx.FieldPath, field reflect.StructField, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	attribs := []mx.Attrib{
		html.Name(string(path)),
		html.ID(string(path)),
		html.Value("on"),
	}
	if mx.FormValueChecked(value) {
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
	if labelText := mx.FieldLabel(field, tag); labelText != "" {
		row = append(row, FieldLabelFor(string(path), labelText))
	}
	parts := []any{
		mx.NewElement("div",
			mx.Attribute{Name: "class", Value: "flex items-center gap-2"},
			row,
		),
	}
	var describedBy []string
	if tag.Help != "" {
		describedBy = append(describedBy, string(path)+"-help")
		parts = append(parts, FieldDescription(html.ID(string(path)+"-help"), mx.Text(tag.Help)))
	}
	// An unchecked box stores false, so a database/sql bool such as
	// sql.NullBool stores its null with the clear control.
	if _, isDriverValue := mx.DriverValueText(reflect.Indirect(value)); isDriverValue && !tag.Required && !tag.Readonly {
		parts = append(parts, clearControl(path, mx.FieldLabel(field, tag), false))
	}
	parts, describedBy = appendFieldErrors(parts, describedBy, path, errs)
	mx.AddDescribedBy(input, describedBy...)
	return fieldWrap(len(errs) > 0, parts...)
}

func shadcnEnum(path mx.FieldPath, field reflect.StructField, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	selected := mx.FormValueText(value)
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
	// As in the html renderer, a select that can drop its value refers to
	// the note saying so, which is empty and hidden when it drops nothing.
	removable := !tag.Required && !tag.Readonly
	if removable {
		attribs = append(attribs, mx.NewAttrib("aria-describedby", optionNoteID(path)))
	}
	contextList := mx.OptionsNeedContext(tag, field.Type)
	removed := false
	options := mx.OptionComponents(value, tag, field.Type, func(opt mx.NamedOption) mx.Component {
		oAttr := []any{html.Value(opt.Value)}
		if opt.Value == selected {
			oAttr = append(oAttr, html.Selected)
		}
		oAttr = append(oAttr, mx.Text(opt.Label()))
		return html.Option(oAttr...)
	}, func(options []mx.NamedOption) mx.Component {
		// Runs before the options render, so they and the preface see the
		// same selection (see the html renderer).
		var preface mx.OptionPreface
		preface, selected = mx.SelectPreface(value, selected, tag, contextList, options)
		removed = preface.RemovesValue(tag)
		return selectPrefaceOption(preface)
	})
	return mx.Components{
		Select(append(attribs, options)...),
		// The options decide whether saving drops the value, also those a
		// context list resolves while the select renders, before this.
		mx.ComponentFunc(func(ctx context.Context, w mx.Writer) error {
			if !removable {
				return nil
			}
			if !removed {
				return hiddenOptionNote(path).Render(ctx, w)
			}
			return FieldDescription(html.ID(optionNoteID(path)),
				labelText(func(l Labels) string { return l.FormValueRemovedOnSave })).Render(ctx, w)
		}),
	}
}

// hiddenOptionNote stands in for the note on values the option list does
// not offer when there are none: empty and hidden, it keeps the
// aria-describedby written before a context list resolved valid.
func hiddenOptionNote(path mx.FieldPath) mx.Component {
	return mx.NewElement("p", html.ID(optionNoteID(path)), html.Hidden)
}

// optionNoteID is the id of the note under a select or checkbox set whose
// stored value the option list does not offer. The lowercase suffix cannot
// collide with a field path, whose segments are exported Go field names.
func optionNoteID(path mx.FieldPath) string {
	return string(path) + "-note"
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
	keyType := mx.SetElemType(field.Type)
	if keyType == nil {
		return mx.Text("[enum-set: cannot infer key type for " + string(path) + "]")
	}
	members := mx.SetMemberValues(value)
	contextList := mx.OptionsNeedContext(tag, keyType)
	items := mx.OptionComponents(reflect.Value{}, tag, keyType, func(opt mx.NamedOption) mx.Component {
		// The label names the box by its id, derived from the value, so any
		// value, also one with spaces, makes a valid id.
		id := mx.KeyedIDValue(path, opt.Value)
		attribs := []mx.Attrib{
			html.ID(id),
			html.Name(string(path)),
			html.Value(opt.Value),
			hx.Trigger("change"),
		}
		if _, ok := slices.BinarySearch(members, opt.Value); ok {
			attribs = append(attribs, html.Checked)
		}
		if tag.Readonly {
			attribs = append(attribs, html.Disabled)
		}
		return mx.NewElement("div",
			mx.Attribute{Name: "class", Value: "flex items-center gap-2"},
			Checkbox(attribs...),
			LabelFor(id, mx.Text(opt.Label())),
		)
	}, func(options []mx.NamedOption) mx.Component {
		preface := mx.EnumSetPreface(members, tag, contextList, options)
		if preface != mx.OptionPrefaceValuesUnavailable {
			return hiddenOptionNote(path)
		}
		if !preface.RemovesValue(tag) {
			return FieldDescription(html.Class("col-span-2"), html.ID(optionNoteID(path)),
				labelText(func(l Labels) string { return l.FormValuesNotAvailableReadonly }))
		}
		return FieldDescription(html.Class("col-span-2"), html.ID(optionNoteID(path)),
			labelText(func(l Labels) string { return l.FormValuesNotAvailable }))
	})
	// The group is described by the note on members the list does not
	// offer, which is empty and hidden when there are none.
	wrapAttribs := []any{
		mx.Attribute{Name: "role", Value: "group"},
		mx.Attribute{Name: "class", Value: "grid grid-cols-2 gap-2"},
	}
	if mx.FieldLabel(field, tag) != "" {
		wrapAttribs = append(wrapAttribs, mx.Attribute{Name: "aria-labelledby", Value: groupLabelID(path)})
	}
	wrapAttribs = append(wrapAttribs, mx.Attribute{Name: "aria-describedby", Value: optionNoteID(path)})
	if len(errs) > 0 {
		wrapAttribs = append(wrapAttribs, mx.NewAttrib("aria-invalid", "true"))
	}
	wrapAttribs = append(wrapAttribs, items)
	return mx.NewElement("div", wrapAttribs...)
}

func stringAttribs(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) []mx.Attrib {
	inputType := mx.StringInputType(tag)
	a := []mx.Attrib{
		html.Type(inputType),
		html.Name(string(path)),
		html.ID(string(path)),
		hx.Trigger("change"),
	}
	if val := mx.FormValueText(value); val != "" && !tag.Sensitive {
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
	// A sensitive number never shows its value, like a sensitive text input,
	// and a null or nil one has none.
	if text := mx.FormValueText(value); !tag.Sensitive && text != "" {
		a = append(a, html.Value(text))
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
func shadcnDateTime(path mx.FieldPath, in mx.DateTimeInputSpec, tag mx.FormTag, errs []error) mx.Component {
	a := []mx.Attrib{
		html.Type(in.Type),
		html.Name(string(path)),
		html.ID(string(path)),
		hx.Trigger("change"),
	}
	if in.Value != "" {
		a = append(a, html.Value(in.Value))
	}
	// A readonly field is never saved, so its text input asks for nothing,
	// and a type that reads only its own spelling (no Format) asks for no
	// format: only a caller's placeholder hints at it.
	textFallback := in.TextFallback()
	askFormat := textFallback && !tag.Readonly && in.Format != ""
	// A caller's placeholder is the only hint for a type that reads only
	// its own spelling.
	if textFallback && !tag.Readonly {
		switch {
		case tag.Placeholder != "":
			a = append(a, html.Placeholder(tag.Placeholder))
		case in.Format != "":
			// The same translated format as the hint below.
			a = append(a, labelAttrib{name: "placeholder", get: func(l Labels) string { return l.dateTimeFormat(in.Format) }})
		}
	}
	if askFormat {
		a = append(a, mx.NewAttrib("aria-describedby", formatHintID(path)))
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
			labelText(func(l Labels) string { return caption(l) + " " + l.dateTimeFormat(in.Format) }),
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
	if val := mx.FormValueText(value); val != "" && !tag.Sensitive {
		a = append(a, mx.Text(val))
	}
	return a
}

// shadcnFile renders a file input and, while the field holds a file, a note
// next to it saying so ([mx.StoredFile]), as the browser labels the input
// "No file chosen" either way, and on an editable one that a blank keeps it.
func shadcnFile(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	attribs := fileAttribs(path, tag, errs)
	name, stored := mx.StoredFile(value)
	if !stored {
		return Input(attribs...)
	}
	noteID := string(path) + "-file"
	attribs = append(attribs, mx.NewAttrib("aria-describedby", noteID))
	note := labelText(func(l Labels) string {
		text := l.FormStoredFileUnnamed
		if name != "" {
			text = strings.Replace(l.FormStoredFile, "%s", name, 1)
		}
		if !tag.Readonly {
			text += " " + l.FormKeepStoredFile
		}
		return text
	})
	return mx.Components{Input(attribs...), FieldDescription(html.ID(noteID), note)}
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

// groupLabelID is the id of the label that names a field's group of
// controls, such as an enum set's checkboxes (as in the html renderer).
func groupLabelID(path mx.FieldPath) string {
	return string(path) + "-label"
}

func clearControl(path mx.FieldPath, label string, checked bool) mx.Component {
	box := []mx.Attrib{
		html.ID(mx.ClearSentinelName(path)),
		html.Name(mx.ClearSentinelName(path)),
		html.Value("1"),
		// The name holds the visible word and says which field a form full
		// of clear boxes, as in repeatable rows, clears.
		labelAttrib{name: "aria-label", get: func(l Labels) string {
			return strings.TrimSpace(strings.Replace(l.FormClearField, "%s", label, 1))
		}},
	}
	if checked {
		box = append(box, html.Checked)
	}
	// The Field stretches its children; the wrapper takes that, so only the
	// box and its word are clickable, not the strip below the input.
	return html.Div(
		LabelFor(mx.ClearSentinelName(path),
			html.Class("w-fit"),
			Checkbox(box...),
			mx.NewElement("span",
				mx.Attribute{Name: "class", Value: "text-sm"},
				labelText(func(l Labels) string { return l.FormClear }),
			),
		),
	)
}

func isNullable(field reflect.StructField, kind mx.FieldKind, tag mx.FormTag) bool {
	if tag.Required {
		return false
	}
	t := field.Type
	if t.Implements(reflect.TypeFor[mx.Nullable]()) || reflect.PointerTo(t).Implements(reflect.TypeFor[mx.Nullable]()) {
		return true
	}
	// A form:"sensitive" text, number or date input keeps the stored value
	// when left blank, so it is emptied with the clear control, as in the
	// html renderer.
	return tag.Sensitive && kind.KeepsStoredWhenBlank()
}

// FormMessage renders a message above a reflected form, such as an onSubmit
// error or an error no input shows, as a destructive [Alert], so it stands
// out like the field errors in red: it is often the only sign that a save
// failed. A shadcn form sets it as its mx.ReflectFormConfig.FormMessage.
func FormMessage(text string) mx.Component {
	return Alert(AlertDestructive, AlertDescription(mx.Text(text)))
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
