package html

import (
	"cmp"
	"context"
	"database/sql/driver"
	"encoding"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/domonda/go-errs"

	"github.com/ungerik/go-mx"
)

// FieldDecider is the plain-HTML implementation of [mx.FieldDecider].
// It implements the dispatch table from the ReflectFormHandler design:
// each detected [mx.FieldKind] picks an HTML element from this package
// (Input / Select / TextArea / Label) and a matching parser. It emits
// the __clear sentinel for nullable fields and form:"sensitive" text,
// number and date fields that are neither required nor readonly; the form
// handler emits the __present sentinel.
//
// Layered renderers (hx, shadcn) wrap this decider — calling it for
// the kinds they don't customize and falling through for the rest.
var FieldDecider mx.FieldDecider = func(path mx.FieldPath, field reflect.StructField, value reflect.Value) mx.FieldBehavior {
	kind, tag := mx.DetectField(path, field, value)
	return mx.FieldBehavior{
		Render: func(path mx.FieldPath, field reflect.StructField, value reflect.Value, errs []error) mx.Component {
			// A probe asks only whether the field is in the form.
			if mx.IsRenderProbe(errs) && rendersAlways(kind) {
				return inForm
			}
			return renderField(path, field, value, kind, tag, errs)
		},
		Parse: func(path mx.FieldPath, field reflect.StructField, value reflect.Value, r *http.Request) error {
			return parseField(path, field, value, kind, tag, r)
		},
	}
}

// inForm answers a render probe for a field in the form, boxed once.
var inForm mx.Component = mx.Components{}

// rendersAlways reports whether renderField returns a Component for every
// field of kind, so a render probe ([mx.IsRenderProbe]) is answered without
// building the markup.
func rendersAlways(kind mx.FieldKind) bool {
	switch kind {
	case mx.FieldKindHidden, mx.FieldKindString, mx.FieldKindTextarea, mx.FieldKindNumber,
		mx.FieldKindBool, mx.FieldKindDateTime, mx.FieldKindFile, mx.FieldKindEnum,
		mx.FieldKindEnumSet, mx.FieldKindCatchAll:
		return true
	}
	return false
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

	input := renderInput(path, field, value, kind, mx.BrowserTag(tag, kind, value), errs)
	if input == nil {
		return nil
	}

	parts := mx.Components{}
	if labelText := mx.FieldLabel(field, tag); labelText != "" && kind == mx.FieldKindEnumSet {
		// A set has no control with the path as its id: the label names
		// its group.
		parts = append(parts, Label(ID(groupLabelID(path)), labelText+":"))
	} else if labelText != "" && kind != mx.FieldKindBool {
		parts = append(parts, Label(For(string(path)), labelText+":"))
	}
	parts = append(parts, input)
	if kind == mx.FieldKindBool {
		if labelText := mx.FieldLabel(field, tag); labelText != "" {
			parts = append(parts, Label(For(string(path)), " "+labelText))
		}
	}
	// The control is described by its help and each error message, after
	// what it refers to itself, such as a date's format hint.
	var describedBy []string
	if tag.Help != "" {
		describedBy = append(describedBy, string(path)+"-help")
		parts = append(parts, mx.NewElement("small",
			mx.Attribute{Name: "class", Value: "form-help"},
			mx.Attribute{Name: "id", Value: string(path) + "-help"},
			mx.Text(tag.Help),
		))
	}
	// A sensitive input never shows its value, so it says that one is
	// stored and that a blank keeps it.
	if !tag.Readonly && mx.KeepsStoredValue(tag, kind, value) {
		describedBy = append(describedBy, string(path)+"-stored")
		parts = append(parts, mx.NewElement("small",
			mx.Attribute{Name: "class", Value: "form-help"},
			mx.Attribute{Name: "id", Value: string(path) + "-stored"},
			mx.Text("Leave blank to keep the current value."),
		))
	}
	// Stored text the input cannot show unchanged says so.
	if problem, ok := FieldTextProblem(kind, value, tag); ok {
		describedBy = append(describedBy, string(path)+"-text")
		parts = append(parts, mx.NewElement("small",
			mx.Attribute{Name: "class", Value: "form-help"},
			mx.Attribute{Name: "id", Value: string(path) + "-text"},
			mx.Text(textProblemNote(problem)),
		))
	}
	// A readonly field is never written, so it offers no clear control.
	// A blank sensitive input keeps the stored value, so the clear control
	// is how it is emptied, ticked while it is empty so a re-render keeps
	// a clear the user asked for.
	sensitive := tag.Sensitive && kind.KeepsStoredWhenBlank()
	// An unchecked box stores false, so a database/sql bool such as
	// sql.NullBool stores its null with the clear control.
	nullableBool := kind == mx.FieldKindBool && isScanner(value)
	if (isNullable(value) || sensitive || nullableBool) && !tag.Required && !tag.Readonly {
		parts = append(parts, clearSentinelInput(path, mx.FieldLabel(field, tag), sensitive && !mx.KeepsStoredValue(tag, kind, value)))
	}
	firstError := len(describedBy)
	for _, e := range errs {
		for _, text := range mx.FieldErrorTexts(e) {
			id := string(path) + "-error-" + strconv.Itoa(len(describedBy)-firstError)
			describedBy = append(describedBy, id)
			parts = append(parts, mx.NewElement("p",
				mx.Attribute{Name: "class", Value: "form-error"},
				mx.Attribute{Name: "id", Value: id},
				mx.Attribute{Name: "data-error", Value: string(path)},
				mx.Text(text),
			))
		}
	}
	mx.AddDescribedBy(input, describedBy...)
	return parts
}

// FieldTextProblem reports why the text input or textarea of a string-like
// field cannot show its stored database/sql text unchanged
// ([mx.DriverValueTextProblem]); a sensitive field shows none of it. The
// renderers of this package and of shadcn share it.
func FieldTextProblem(kind mx.FieldKind, value reflect.Value, tag mx.FormTag) (mx.TextProblem, bool) {
	switch {
	case tag.Sensitive:
		return 0, false
	case kind == mx.FieldKindString, kind == mx.FieldKindCatchAll:
		return mx.DriverValueTextProblem(value, false)
	case kind == mx.FieldKindTextarea:
		return mx.DriverValueTextProblem(value, true)
	}
	return 0, false
}

// textProblemNote is the note under an input that cannot show its stored
// text unchanged.
func textProblemNote(problem mx.TextProblem) string {
	switch problem {
	case mx.TextProblemNUL:
		return "The stored text holds a NUL character, shown as \uFFFD; saving it unchanged keeps it."
	case mx.TextProblemInvalidUTF8:
		return "The stored text is not valid UTF-8, shown as \uFFFD; saving it unchanged keeps it."
	}
	return "The stored text holds line breaks this input cannot show; saving it unchanged keeps it."
}

// renderInput emits the input element for one field based on its
// [mx.FieldKind].
func renderInput(path mx.FieldPath, field reflect.StructField, value reflect.Value, kind mx.FieldKind, tag mx.FormTag, errs []error) mx.Component {
	switch kind {
	case mx.FieldKindString:
		t := mx.StringInputType(tag)
		return inputElement(path, t, mx.FormValueText(value), field, tag, errs)
	case mx.FieldKindTextarea:
		return textareaElement(path, mx.FormValueText(value), field, tag, errs)
	case mx.FieldKindNumber:
		return numberInput(path, value, tag, errs)
	case mx.FieldKindBool:
		return checkboxInput(path, value, tag, errs)
	case mx.FieldKindDateTime:
		return datetimeInput(path, value, tag, errs)
	case mx.FieldKindFile:
		return fileInput(path, value, tag, errs)
	case mx.FieldKindEnum:
		return selectInput(path, value, tag, field, errs)
	case mx.FieldKindEnumSet:
		return enumSetInput(path, value, field, tag, errs)
	case mx.FieldKindCatchAll:
		return inputElement(path, "text", mx.FormValueText(value), field, tag, errs)
	}
	return nil
}

// renderHidden writes the value as [mx.FormValueText] does; a hidden time, such
// as an UpdatedAt kept for optimistic locking, round-trips as RFC 3339.
func renderHidden(path mx.FieldPath, value reflect.Value) mx.Component {
	return VoidElement("input",
		Type("hidden"),
		Name(string(path)),
		Value(mx.FormValueText(value)),
	)
}

// isTimeType reports whether t is time.Time or a type defined as it.
func isTimeType(t reflect.Type) bool {
	return t.Kind() == reflect.Struct && t.ConvertibleTo(reflect.TypeFor[time.Time]())
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
	attribs = append(attribs, TextTagAttribs(tag)...)
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

// FileTagAttribs returns the attributes a file input gets from its field: its
// form:"accept=…", and multiple for a field of several files
// ([mx.MultipleFiles]). The renderers of this package and of shadcn share
// it.
func FileTagAttribs(value reflect.Value, tag mx.FormTag) []mx.Attrib {
	var a []mx.Attrib
	if tag.Accept != "" {
		a = append(a, Accept(tag.Accept))
	}
	if value.IsValid() && mx.MultipleFiles(value.Type()) {
		a = append(a, Multiple)
	}
	return a
}

// TextTagAttribs returns the attributes a text input or textarea gets from
// the tag's form:"minlength=…,maxlength=…,autocomplete=…,inputmode=…": the
// renderers of this package and of shadcn share it. An inputmode that is no
// [InputMode] defers its error to render time.
func TextTagAttribs(tag mx.FormTag) []mx.Attrib {
	var a []mx.Attrib
	if tag.MinLength > 0 {
		a = append(a, MinLength(tag.MinLength))
	}
	if tag.MaxLength > 0 {
		a = append(a, MaxLength(tag.MaxLength))
	}
	if tag.Autocomplete != "" {
		a = append(a, AutoComplete(tag.Autocomplete))
	}
	if tag.InputMode != "" {
		a = append(a, InputMode(tag.InputMode))
	}
	return a
}

func textareaElement(path mx.FieldPath, val string, _ reflect.StructField, tag mx.FormTag, errs []error) mx.Component {
	attribs := []any{
		Name(string(path)),
		ID(string(path)),
	}
	if tag.Placeholder != "" {
		attribs = append(attribs, Placeholder(tag.Placeholder))
	}
	for _, a := range TextTagAttribs(tag) {
		attribs = append(attribs, a)
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
	}
	// A sensitive number never shows its value, like a sensitive text input,
	// and a null or nil one has none.
	if text := mx.FormValueText(value); !tag.Sensitive && text != "" {
		attribs = append(attribs, Value(text))
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
	if mx.FormValueChecked(value) {
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
	if err := mx.DateTimeTagError(value, tag); err != nil {
		return mx.NewErrElement(err)
	}
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
	// date range and it says which format to enter. Text the parser rejects
	// is invalid from the start, because saving it unchanged fails. A
	// readonly field is never saved, so it asks for nothing, and a type
	// that reads only its own spelling (no Format) asks for no format: only
	// a caller's placeholder hints at it.
	textFallback := in.TextFallback()
	askFormat := textFallback && !tag.Readonly && in.Format != ""
	// A caller's placeholder is the only hint for a type that reads only
	// its own spelling.
	if placeholder := cmp.Or(tag.Placeholder, in.Format); textFallback && !tag.Readonly && placeholder != "" {
		attribs = append(attribs, Placeholder(placeholder))
	}
	if askFormat {
		attribs = append(attribs, mx.NewAttrib("aria-describedby", formatHintID(path)))
	}
	if !textFallback {
		if tag.Min != "" {
			attribs = append(attribs, Min(tag.Min))
		}
		if tag.Max != "" {
			attribs = append(attribs, Max(tag.Max))
		}
		if in.Step != "" {
			attribs = append(attribs, Step(in.Step))
		}
	}
	if tag.Required {
		attribs = append(attribs, Required)
	}
	if tag.Readonly {
		attribs = append(attribs, Readonly)
	}
	if len(errs) > 0 || (in.Invalid && !tag.Readonly) {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	if !askFormat {
		return Input(attribs...)
	}
	return mx.Components{
		Input(attribs...),
		mx.NewElement("small",
			mx.Attribute{Name: "class", Value: "form-help"},
			mx.Attribute{Name: "id", Value: formatHintID(path)},
			mx.Text(formatHint(in, errs)),
		),
	}
}

// formatHint names the format the parser reads, which a placeholder from
// the tag does not replace, and says so when the stored text is not in it.
// After a submit, the field error says that already.
func formatHint(in mx.DateTimeInputSpec, errs []error) string {
	if in.Invalid && len(errs) == 0 {
		return "Not in the expected format: " + in.Format
	}
	return "Expected format: " + in.Format
}

// formatHintID is the id of the format hint under a date or time field
// that fell back to a text input. The lowercase suffix cannot collide with
// a field path, whose segments are exported Go field names.
func formatHintID(path mx.FieldPath) string {
	return string(path) + "-format"
}

func fileInput(path mx.FieldPath, value reflect.Value, tag mx.FormTag, errs []error) mx.Component {
	attribs := []mx.Attrib{
		Type("file"),
		Name(string(path)),
		ID(string(path)),
	}
	attribs = append(attribs, FileTagAttribs(value, tag)...)
	// The browser labels the input "No file chosen" also while the field
	// holds a file, which a note next to it says, and that a blank keeps it.
	name, stored := mx.StoredFile(value)
	var note mx.Component
	if stored {
		text := "A file is stored."
		if name != "" {
			text = "Current file: " + name
		}
		if !tag.Readonly {
			text += " Leave empty to keep it."
		}
		attribs = append(attribs, mx.NewAttrib("aria-describedby", fileNoteID(path)))
		note = mx.NewElement("small",
			mx.Attribute{Name: "class", Value: "form-help"},
			mx.Attribute{Name: "id", Value: fileNoteID(path)},
			mx.Text(text),
		)
	}
	// A file input cannot be readonly, and a readonly field is never
	// saved, so it is disabled and asks for nothing.
	switch {
	case tag.Readonly:
		attribs = append(attribs, Disabled)
	case tag.Required:
		attribs = append(attribs, Required)
	}
	if len(errs) > 0 {
		attribs = append(attribs, mx.NewAttrib("aria-invalid", "true"))
	}
	if note == nil {
		return Input(attribs...)
	}
	return mx.Components{Input(attribs...), note}
}

// fileNoteID is the id of the note next to a file input whose field holds a
// file. The lowercase suffix cannot collide with a field path, whose
// segments are exported Go field names.
func fileNoteID(path mx.FieldPath) string {
	return string(path) + "-file"
}

func selectInput(path mx.FieldPath, value reflect.Value, tag mx.FormTag, field reflect.StructField, errs []error) mx.Component {
	selected := mx.FormValueText(value)
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
	// Saving drops a value the list does not offer, which the note after
	// the select says, also to a screen reader. The start tag is written
	// before a context list resolves, so a select that can drop its value
	// always refers to the note, which is empty and hidden when it drops
	// nothing.
	removable := !tag.Required && !tag.Readonly
	if removable {
		attribs = append(attribs, mx.NewAttrib("aria-describedby", optionNoteID(path)))
	}
	contextList := mx.OptionsNeedContext(tag, field.Type)
	removed := false
	options := mx.OptionComponents(value, tag, field.Type, func(opt mx.NamedOption) mx.Component {
		oAttr := []any{Value(opt.Value)}
		if opt.Value == selected {
			oAttr = append(oAttr, Selected)
		}
		oAttr = append(oAttr, mx.Text(opt.Label()))
		return Option(oAttr...)
	}, func(options []mx.NamedOption) mx.Component {
		// Runs before the options render, so they and the preface see the
		// same selection: an unlisted zero selects the list's own empty
		// option, if it has one, or the preface's.
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
			text := ""
			if removed {
				text = "The current value is removed when you save."
			}
			return optionNote(path, text).Render(ctx, w)
		}),
	}
}

// optionNote is the note a select or checkbox set describes itself by
// (see [optionNoteID]): text, or, for "", an empty hidden element, so the
// reference written before a context list resolved stays valid.
func optionNote(path mx.FieldPath, text string) mx.Component {
	if text == "" {
		return mx.NewElement("p",
			mx.Attribute{Name: "id", Value: optionNoteID(path)},
			Hidden,
		)
	}
	return mx.NewElement("p",
		mx.Attribute{Name: "class", Value: "form-help"},
		mx.Attribute{Name: "id", Value: optionNoteID(path)},
		mx.Text(text),
	)
}

// optionNoteID is the id of the note under a select or checkbox set whose
// stored value the option list does not offer. The lowercase suffix cannot
// collide with a field path, whose segments are exported Go field names.
func optionNoteID(path mx.FieldPath) string {
	return string(path) + "-note"
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
	keyType := mx.SetElemType(field.Type)
	if keyType == nil {
		return mx.Text("[enum-set: cannot infer key type for " + string(path) + "]")
	}
	members := mx.SetMemberValues(value)
	contextList := mx.OptionsNeedContext(tag, keyType)
	items := mx.OptionComponents(reflect.Value{}, tag, keyType, func(opt mx.NamedOption) mx.Component {
		inputAttribs := []mx.Attrib{
			Type("checkbox"),
			Name(string(path)),
			Value(opt.Value),
		}
		if _, ok := slices.BinarySearch(members, opt.Value); ok {
			inputAttribs = append(inputAttribs, Checked)
		}
		if tag.Readonly {
			inputAttribs = append(inputAttribs, Disabled)
		}
		return Label(Input(inputAttribs...), " "+opt.Label())
	}, func(options []mx.NamedOption) mx.Component {
		preface := mx.EnumSetPreface(members, tag, contextList, options)
		if preface != mx.OptionPrefaceValuesUnavailable {
			return optionNote(path, "")
		}
		note := "(some current values are not available)"
		if preface.RemovesValue(tag) {
			note = "(some current values are not available and are removed when you save)"
		}
		return optionNote(path, note)
	})
	// The group is described by the note on members the list does not
	// offer, which is empty and hidden when there are none (see
	// optionNote).
	group := []any{mx.Attribute{Name: "role", Value: "group"}}
	if mx.FieldLabel(field, tag) != "" {
		group = append(group, mx.Attribute{Name: "aria-labelledby", Value: groupLabelID(path)})
	}
	group = append(group, mx.Attribute{Name: "aria-describedby", Value: optionNoteID(path)})
	if len(errs) > 0 {
		group = append(group, mx.Attribute{Name: "aria-invalid", Value: "true"})
	}
	return mx.NewElement("div", append(group, items)...)
}

// parseField writes the submitted form value into value (which is
// addressable). For multipart file uploads, parseField stores the
// uploaded file's bytes when the destination is []byte; richer file
// handling is left to higher layers.
func parseField(path mx.FieldPath, field reflect.StructField, value reflect.Value, kind mx.FieldKind, tag mx.FormTag, r *http.Request) error {
	if !value.CanSet() {
		return errors.New("field not settable: " + string(path))
	}
	form := r.PostForm
	if r.MultipartForm != nil {
		form = r.MultipartForm.Value
	}
	switch kind {
	case mx.FieldKindHidden, mx.FieldKindString, mx.FieldKindTextarea, mx.FieldKindCatchAll:
		raw := form.Get(string(path))
		if kind == mx.FieldKindHidden && isTimeType(derefType(value.Type())) {
			// A hidden time is written as RFC 3339, whatever the field's
			// widget or range, and setTime reads that without them, also
			// for a type defined as time.Time, which has no UnmarshalText.
			return setTime(r.Context(), value, raw, mx.FormTag{Sensitive: tag.Sensitive})
		}
		if kind == mx.FieldKindHidden && isPlainNumber(derefType(value.Type())) {
			// A hidden number, such as a per-row key, binds as a number
			// does, and a hidden bool as a checkbox does; setScalar takes
			// only text.
			return guardBind(r.Context(), value, raw, tag, func() error { return setNumeric(value, raw) })
		}
		if kind == mx.FieldKindHidden && derefType(value.Type()).Kind() == reflect.Bool &&
			!reflect.PointerTo(derefType(value.Type())).Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) {
			// A nil *bool renders as "" and stays nil.
			if value.Kind() == reflect.Pointer && raw == "" {
				value.SetZero()
				return nil
			}
			return guardBind(r.Context(), value, raw, tag, func() error { return setBool(value, raw) })
		}
		if (kind == mx.FieldKindString || kind == mx.FieldKindCatchAll) && keepsStoredLines(value, raw) {
			return nil
		}
		return guardBind(r.Context(), value, raw, tag, func() error {
			if err := setScalar(value, raw); err != nil {
				return err
			}
			// Checked once bound, so the re-render shows what was posted, as
			// a number outside its range does. Only the inputs that render
			// the pattern attribute are held to it.
			if kind == mx.FieldKindString || kind == mx.FieldKindCatchAll {
				if err := checkPattern(tag.Pattern, raw); err != nil {
					return err
				}
			}
			if kind == mx.FieldKindString || kind == mx.FieldKindCatchAll || kind == mx.FieldKindTextarea {
				if err := mx.CheckTextLength(tag, raw); err != nil {
					return err
				}
			}
			return checkScannedRange(value, raw, tag)
		})
	case mx.FieldKindNumber:
		raw := form.Get(string(path))
		// No number input submits more; a longer text is no number, and
		// parsing it would only cost time.
		if len(raw) > maxNumberLen {
			return errors.New("too long for a number")
		}
		return guardBind(r.Context(), value, raw, tag, func() error {
			if err := setNumeric(value, raw); err != nil {
				return err
			}
			// A number that is no pointer stores 0 for a blank, which a
			// required one rejects, as a required select rejects its empty
			// option; the browser does too, so only a tampered POST sends it.
			if raw == "" && tag.Required && value.Kind() != reflect.Pointer {
				return errors.New("required")
			}
			// The range applies on the server too, not only in the browser.
			// The number stays bound, as a date outside its range does.
			if derefValue := reflect.Indirect(value); derefValue.IsValid() && isScanner(derefValue) {
				return checkScannedRange(value, raw, tag)
			}
			return mx.CheckNumberRange(value, raw, tag)
		})
	case mx.FieldKindBool:
		raw := form.Get(string(path))
		return setBool(value, raw)
	case mx.FieldKindDateTime:
		raw := form.Get(string(path))
		return setTime(r.Context(), value, raw, tag)
	case mx.FieldKindEnum:
		raw := form.Get(string(path))
		// A required select's empty option must not become a zero non-string
		// value. isEffectivelyEmpty never treats a number as empty, so the
		// handler's required check would accept the zero of an int enum.
		if raw == "" && tag.Required {
			return fmt.Errorf("required")
		}
		return guardBind(r.Context(), value, raw, tag, func() error {
			if raw != "" {
				if err := validateOptionMembership(r.Context(), value, tag, field.Type, raw); err != nil {
					return err
				}
			}
			return setScalar(value, raw)
		})
	case mx.FieldKindEnumSet:
		vals := form[string(path)]
		return guardBind(r.Context(), value, strings.Join(vals, ","), tag, func() error {
			if keyType := mx.SetElemType(field.Type); keyType != nil && mx.OptionsNeedContext(tag, keyType) {
				// The list is the trust boundary for the set, same as a select.
				// Submitted values outside it are rejected. Members the list
				// does not offer are not kept: the checkboxes the user saw are
				// the whole set that a save stores, so an uncheck cannot be
				// undone by the list changing before POST.
				options, err := mx.CollectOptions(r.Context(), reflect.Value{}, tag, keyType)
				if err != nil {
					return optionListError(r.Context(), err, tag)
				}
				if err := rejectUnlistedOptions(options, vals); err != nil {
					return err
				}
			}
			return setEnumSet(value, field.Type, vals)
		})
	case mx.FieldKindFile:
		if r.MultipartForm == nil {
			// A urlencoded body carries only a chosen file's name: the
			// form missed its multipart enctype, and saving would drop the
			// upload. A blank input keeps the stored file.
			if r.PostForm.Get(string(path)) != "" {
				return errors.New("the file was not uploaded: the form must be sent as multipart/form-data")
			}
			return nil
		}
		files, ok := r.MultipartForm.File[string(path)]
		if !ok || len(files) == 0 {
			return nil
		}
		if mx.MultipleFiles(value.Type()) {
			return setFiles(r.Context(), value, files)
		}
		return setFile(value, files[0], mx.FormMaxMemory(r.Context()))
	}
	return nil
}

// setFile binds one upload to a file field: a string gets its file name, a
// *multipart.FileHeader the upload itself, a []byte its content, read into
// memory only up to limit bytes, as the multipart parse is bounded.
func setFile(value reflect.Value, fh *multipart.FileHeader, limit int64) error {
	switch {
	case value.Kind() == reflect.String:
		value.SetString(fh.Filename)
		return nil
	case value.Type() == reflect.TypeFor[*multipart.FileHeader]():
		value.Set(reflect.ValueOf(fh))
		return nil
	case value.Kind() != reflect.Slice || value.Type().Elem().Kind() != reflect.Uint8:
		return errors.New("file upload requires a []byte, string or *multipart.FileHeader field")
	}
	// The multipart parser keeps a large file on disk; reading it into
	// memory is bounded like the parse.
	if fh.Size > limit {
		return fmt.Errorf("the file is larger than %d bytes", limit)
	}
	f, err := fh.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, fh.Size)
	if _, err := io.ReadFull(f, buf); err != nil {
		return err
	}
	value.SetBytes(buf)
	return nil
}

// setFiles binds the uploads of a multiple file input to a field of several
// files ([mx.MultipleFiles]), one element per file in the order sent. The
// contents read into memory are bounded together by [mx.FormMaxMemory].
func setFiles(ctx context.Context, value reflect.Value, files []*multipart.FileHeader) error {
	limit := mx.FormMaxMemory(ctx)
	slice := reflect.MakeSlice(derefType(value.Type()), len(files), len(files))
	for i, fh := range files {
		if err := setFile(slice.Index(i), fh, limit); err != nil {
			return err
		}
		if slice.Index(i).Kind() == reflect.Slice {
			limit -= fh.Size
		}
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}
	value.Set(slice)
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
// field). A failure to resolve the option list itself becomes
// [optionListError]: the [mx.UserMessage] of an error the list returned,
// else [mx.FormErrorMessage] unless [mx.RevealInternalServerErrors] is set,
// with the error reported to [mx.ReportError].
func validateOptionMembership(ctx context.Context, value reflect.Value, tag mx.FormTag, t reflect.Type, vals ...string) error {
	if len(vals) == 0 || !mx.OptionsNeedContext(tag, t) {
		return nil
	}
	options, err := mx.CollectOptions(ctx, value, tag, t)
	if err != nil {
		return optionListError(ctx, err, tag)
	}
	return rejectUnlistedOptions(options, vals)
}

// optionListError is the field error for a context option list that failed
// while a POST was parsed. Like an onSubmit error, its text may carry
// internal details and a call stack, so the field shows [mx.FormErrorText]:
// the [mx.UserMessage] of a [mx.UserMessageError] the list returned, else
// [mx.FormErrorMessage] unless [mx.RevealInternalServerErrors] is set, with
// the error reported to [mx.ReportError]. A form:"sensitive" field reveals
// nothing, as the list's error may quote the value. The text is a
// [mx.FormMessage], meant for the user, so the sensitive field's guard shows
// it as it is and does not report the error again.
func optionListError(ctx context.Context, err error, tag mx.FormTag) error {
	text := mx.FormErrorText(ctx, err)
	if _, ok := mx.UserMessage(err); !ok && tag.Sensitive {
		text = mx.FormErrorMessage
	}
	return mx.FormMessage(text)
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
			return errors.New(quotedValue(v) + " is not one of the allowed options")
		}
	}
	return nil
}

// guardBind runs bind, which writes raw into value and may call the type's
// own UnmarshalText. A method that panics, as the promoted method of a nil
// embedded pointer does, is an error instead of a failed request, and is
// reported with its stack ([mx.ReportHiddenError]). Its error
// is bounded by [mx.MaxFieldErrorLen] ([mx.BoundFieldError]): an option list,
// UnmarshalText or strconv error quotes raw whole, into the page, or, where
// no input shows it, through [mx.ReportError] into the application's logs. A form:"sensitive"
// field's error reads "not a valid value" ([mx.MaskSensitiveError]), leaving raw out
// at any length. (Date and time fields guard their own panics and mask
// their errors with their format instead.)
func guardBind(ctx context.Context, value reflect.Value, raw string, tag mx.FormTag, bind func() error) (err error) {
	defer func() {
		if p := recover(); p != nil {
			// A bug in the application's type, which the user is told of only
			// as an invalid value.
			mx.ReportHiddenError(ctx, errs.AsErrorWithDebugStack(p))
			err = fmt.Errorf("%s cannot be stored as %s", quotedValue(raw), value.Type())
		}
		if err != nil {
			// A sensitive field's substitute is its mask, whatever its length.
			substitute := mx.FormMessage("the text is not a valid value")
			if tag.Sensitive {
				substitute = "not a valid value"
			}
			err = mx.BoundFieldError(ctx, err, raw, substitute)
		}
		if err != nil && tag.Sensitive {
			err = mx.MaskSensitiveError(ctx, err, "not a valid value")
		}
	}()
	return bind()
}

// quotedValue returns raw quoted for an error, or "the text" for raw longer
// than [mx.MaxFieldErrorLen]: the error would be past the bound and dropped
// anyway ([mx.BoundFieldError]), and quoting first would copy a forged
// value of any size, %q growing a control byte to four.
func quotedValue(raw string) string {
	if maxLen := mx.MaxFieldErrorLen; maxLen > 0 && len(raw) > maxLen {
		return "the text"
	}
	return strconv.Quote(raw)
}

// patterns caches the compiled form:"pattern=…" of each tag, nil for one
// Go cannot compile. The patterns come from struct tags, so they are few.
var patterns sync.Map // pattern string → *regexp.Regexp

// checkPattern returns a validation error when raw, a non-empty submitted
// text, does not match pattern, the form:"pattern=…" the input renders as
// its pattern attribute, so a scripted POST is held to it too. Like the
// attribute, the pattern must match the whole text and does not apply to
// an empty one, which form:"required" decides. A pattern Go's RE2 cannot
// compile, such as one with a JavaScript lookahead, is not checked, as a
// browser ignores a pattern it cannot compile. The error quotes neither
// text nor pattern.
func checkPattern(pattern, raw string) error {
	if pattern == "" || raw == "" {
		return nil
	}
	cached, ok := patterns.Load(pattern)
	if !ok {
		re, err := regexp.Compile("^(?:" + pattern + ")$")
		if err != nil {
			re = nil
		}
		cached, _ = patterns.LoadOrStore(pattern, re)
	}
	if re := cached.(*regexp.Regexp); re != nil && !re.MatchString(raw) {
		return errors.New("does not match the required format")
	}
	return nil
}

// maxNumberLen is the longest text a number field takes: the longest a
// float64 is written in by the renderers, strconv.FormatFloat(f, 'f', -1,
// 64), as for -2.2250738585072014e-308, so every value a field shows can be
// posted back. An integer takes at most 20 characters.
const maxNumberLen = 327

// dateFormatError is [mx.DateTimeFormatError] for raw, except for a type
// that reads only its own spelling, whose format the error cannot name
// (its [mx.DateTimeInput] has no Format): the input's format would send
// the user to text the type rejects.
func dateFormatError(value reflect.Value, tag mx.FormTag, raw string) error {
	if value.IsValid() && mx.DateTimeInput(value, tag).Format == "" {
		return mx.FormMessage("not a valid value")
	}
	return mx.DateTimeFormatError(tag, raw)
}

// maskDateTimeError is [mx.MaskSensitiveError] for err, the error of a
// form:"sensitive" date or time field, with formatErr, the format to enter,
// as its mask: a user message in err shows, else formatErr does, kept as the
// [mx.DateTimeError] it is, so a renderer can still translate it. err is
// reported where it is hidden.
func maskDateTimeError(ctx context.Context, err, formatErr error) error {
	if _, ok := mx.UserMessage(err); ok {
		return mx.MaskSensitiveError(ctx, err, formatErr.Error())
	}
	mx.ReportHiddenError(ctx, err)
	return formatErr
}

// bindPointee binds into a new pointee of the pointer value that starts as
// a copy of the old one, so a submission, bound or rejected, never reaches
// a record that shares the old pointee, such as one in a store a
// shallow-copying load left the field pointing into.
func bindPointee(value reflect.Value, bind func(reflect.Value) error) error {
	pointee := reflect.New(value.Type().Elem())
	if !value.IsNil() {
		pointee.Elem().Set(value.Elem())
	}
	err := bind(pointee.Elem())
	value.Set(pointee)
	return err
}

// setScalar writes raw into a string-shaped, TextUnmarshaler or
// database/sql (see setScanned) destination. A pointer gets a new pointee.
func setScalar(value reflect.Value, raw string) error {
	if value.Kind() == reflect.Pointer {
		if raw == "" {
			value.SetZero()
			return nil
		}
		return bindPointee(value, func(v reflect.Value) error { return setScalar(v, raw) })
	}
	if isScanner(value) {
		return setScanned(value, raw)
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
		if _, ok := value.Addr().Interface().(encoding.TextUnmarshaler); ok {
			// Unmarshal into a copy: an UnmarshalText that fails may still
			// change its value, as netip.Addr's zeroes it, and a rejected
			// field must keep the stored value.
			dst := reflect.New(value.Type())
			dst.Elem().Set(value)
			if err := dst.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(raw)); err != nil {
				return err
			}
			value.Set(dst.Elem())
			return nil
		}
	}
	if u, ok := value.Interface().(encoding.TextUnmarshaler); ok {
		return u.UnmarshalText([]byte(raw))
	}
	return fmt.Errorf("cannot parse value into %s: implement encoding.TextUnmarshaler or database/sql.Scanner, or use a form:\"widget=...\" override", value.Type())
}

// scanner is database/sql.Scanner, without linking database/sql.
type scanner interface{ Scan(src any) error }

// isScanner reports whether value is a database/sql type such as
// sql.NullInt64 ([mx.DriverValueType]), whose Scan method reads back what
// [mx.DriverValueText] writes. It is decided by the type, so a value
// holding text an input cannot show, such as a NUL, is still scanned.
func isScanner(value reflect.Value) bool {
	return value.CanAddr() && mx.DriverValueType(value.Type())
}

// lineBreakStripper removes line breaks as a browser does from the value of
// a one-line input.
var lineBreakStripper = strings.NewReplacer("\r", "", "\n", "")

// unchangedText reports whether raw, a posted value, is shown, the text the
// input was rendered with, coming back unchanged: as it was, or with the
// CRLF line breaks a browser submits for a textarea's LF.
func unchangedText(raw, shown string) bool {
	return raw == shown || strings.Contains(raw, "\r\n") && strings.ReplaceAll(raw, "\r\n", "\n") == shown
}

// keepsStoredLines reports whether raw, posted by a one-line input for a
// string or database/sql text field, is the field's stored text without the
// line breaks the input stripped from its value, so that saving the form
// unchanged keeps them.
func keepsStoredLines(value reflect.Value, raw string) bool {
	v := reflect.Indirect(value)
	var stored string
	switch {
	case !v.IsValid():
		return false
	case isScanner(v):
		stored = mx.FormValueText(v)
	case v.Kind() == reflect.String:
		stored = v.String()
	default:
		return false
	}
	return raw != stored && strings.ContainsAny(stored, "\r\n") && raw == lineBreakStripper.Replace(stored)
}

// setScanned binds raw to a database/sql type such as sql.NullInt64 with
// its Scan method, which reads the text [mx.DriverValueText] writes: as a
// string, as a driver gives text; as []byte, which the Scan of a JSON
// column may require; or as the time.Time of RFC 3339 text. A blank is
// null. Text unchanged from the stored value keeps it, so a valid empty
// string stays valid and a time keeps its zone. A number is held to the
// rules of a number field: finite and at most maxNumberLen long (the
// handler checks form:"min=…,max=…", see checkScannedRange).
func setScanned(value reflect.Value, raw string) error {
	// The shown text comes back unchanged also when the input could not
	// hold the stored text: a NUL or invalid UTF-8 shown as U+FFFD. The line
	// breaks a one-line input strips are kept by parseField
	// (keepsStoredLines), which knows the input.
	if shown := mx.FormValueText(value); unchangedText(raw, shown) {
		return nil
	}
	// Text longer than any number is rejected before a type that scans only
	// numbers or bools sees it: each failing Scan quotes the text, once as a
	// string and once as bytes, so a forged value would be copied many times
	// over into errors that are dropped.
	if len(raw) > maxNumberLen {
		if stored, ok := scansOnlyNumbers(value.Type()); ok {
			if _, isBool := stored.(bool); !isBool {
				return errors.New("too long for a number")
			}
			return scanError(raw, stored)
		}
	}
	// Scanned into a new value: a Scan that fails may still change its
	// value, and a rejected field keeps the stored one.
	scan := func(src any) (reflect.Value, error) {
		dst := reflect.New(value.Type())
		return dst, dst.Interface().(scanner).Scan(src)
	}
	sources := []any{nil} // the null, for a blank
	if raw != "" {
		sources = []any{raw, []byte(raw)}
		if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
			sources = append(sources, t)
		}
	}
	var dst reflect.Value
	var err error
	for _, src := range sources {
		if dst, err = scan(src); err == nil {
			break
		}
	}
	if err != nil {
		// The driver's error names Go types; the stored value tells what
		// to enter.
		return scanError(raw, driverValueOf(value))
	}
	switch n := driverValueOf(dst.Elem()).(type) {
	case int64, float64:
		if len(raw) > maxNumberLen {
			return errors.New("too long for a number")
		}
		if f, ok := n.(float64); ok && (math.IsNaN(f) || math.IsInf(f, 0)) {
			return fmt.Errorf("%q is not a finite number", raw)
		}
	}
	value.Set(dst.Elem())
	return nil
}

// checkScannedRange checks form:"min=…,max=…" on a database/sql number,
// such as a sql.NullInt64 that setScanned bound, as on a number field.
func checkScannedRange(value reflect.Value, raw string, tag mx.FormTag) error {
	if tag.Min == "" && tag.Max == "" {
		return nil
	}
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if !isScanner(value) {
		return nil
	}
	switch n := driverValueOf(value).(type) {
	case int64, float64:
		return mx.CheckNumberRange(reflect.ValueOf(n), raw, tag)
	}
	return nil
}

// scansOnlyNumbers reports whether the database/sql type t scans a number
// but no text, as sql.NullInt64, sql.NullFloat64 and sql.NullBool do, and
// returns the driver value it holds after scanning one, which tells
// [scanError] what to ask for. A type whose Scan panics scans nothing.
func scansOnlyNumbers(t reflect.Type) (stored any, ok bool) {
	defer func() {
		if recover() != nil {
			stored, ok = nil, false
		}
	}()
	if reflect.New(t).Interface().(scanner).Scan("x") == nil {
		return nil, false
	}
	dst := reflect.New(t)
	if dst.Interface().(scanner).Scan(int64(0)) != nil {
		return nil, false
	}
	return driverValueOf(dst.Elem()), true
}

// driverValueOf returns the driver value an addressable database/sql value
// holds, or nil.
func driverValueOf(value reflect.Value) any {
	if !value.CanAddr() {
		return nil
	}
	valuer, ok := value.Addr().Interface().(driver.Valuer)
	if !ok {
		return nil
	}
	dv, err := valuer.Value()
	if err != nil {
		return nil
	}
	return dv
}

// scanError is the validation error for raw, text a database/sql type does
// not scan: it names what to enter, judged by the stored driver value. Text
// longer than any number or time is not quoted, so a submission of any size
// is not echoed into the page.
func scanError(raw string, stored any) error {
	// Checked before quoting, so a forged value is not copied into a text
	// that is then dropped.
	subject := "the text is"
	if len(raw) <= maxNumberLen {
		subject = strconv.Quote(raw) + " is"
	}
	switch stored.(type) {
	case int64:
		return fmt.Errorf("%s not a whole number", subject)
	case float64:
		return fmt.Errorf("%s not a number", subject)
	case bool:
		return fmt.Errorf("%s not true or false", subject)
	case time.Time:
		return fmt.Errorf("%s not a time in RFC 3339 format, such as 2026-01-02T15:04:05Z", subject)
	}
	return fmt.Errorf("%s not a valid value", subject)
}

func setNumeric(value reflect.Value, raw string) error {
	if value.Kind() == reflect.Pointer {
		if raw == "" {
			value.SetZero()
			return nil
		}
		return bindPointee(value, func(v reflect.Value) error { return setNumeric(v, raw) })
	}
	// A database/sql type holding a number, such as sql.NullInt64, scans
	// it, and a blank as its null.
	if isScanner(value) {
		return setScanned(value, raw)
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
		// No input submits NaN or an infinity, and no min or max can bound
		// NaN, which compares false with everything.
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return fmt.Errorf("%q is not a finite number", raw)
		}
		value.SetFloat(f)
	default:
		return fmt.Errorf("cannot store a number in a %s", value.Kind())
	}
	return nil
}

func setBool(value reflect.Value, raw string) error {
	on := raw == "on" || raw == "true" || raw == "1"
	// A database/sql type holding a bool, such as sql.NullBool, scans it;
	// its clear control stores the null an unchecked box cannot.
	if isScanner(value) {
		return setScanned(value, strconv.FormatBool(on))
	}
	if value.Kind() == reflect.Pointer {
		if !value.CanSet() {
			return errors.New("pointer not settable")
		}
		// The pointee may be a database/sql type itself, as a
		// *sql.NullBool's is.
		return bindPointee(value, func(v reflect.Value) error { return setBool(v, raw) })
	}
	value.SetBool(on)
	return nil
}

func setTime(ctx context.Context, value reflect.Value, raw string, tag mx.FormTag) error {
	// Text longer than any date or time input sends is no date. It is
	// rejected before it is bound, so the re-render does not echo a
	// submission of any size, and the field keeps its stored value.
	if len(raw) > mx.MaxDateTimeLen {
		// The format is all it says, a user message, so it is neither
		// masked nor reported (see below).
		return dateFormatError(value, tag, raw)
	}
	if value.Kind() == reflect.Pointer {
		if raw == "" {
			value.SetZero()
			return nil
		}
		return bindPointee(value, func(v reflect.Value) error { return setTime(ctx, v, raw, tag) })
	}
	if value.Kind() == reflect.String {
		return setTimeText(ctx, value, raw, tag)
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
	case "date", "datepicker":
		layouts = []string{"2006-01-02"}
	case "time":
		layouts = []string{"15:04:05", "15:04"}
	}
	var t time.Time
	var parseErr error
	zoned := false
	for _, layout := range layouts {
		if t, parseErr = time.Parse(layout, raw); parseErr == nil {
			zoned = layout == time.RFC3339
			break
		}
	}
	// time.Parse accepts zone offsets up to 99 hours; RFC 3339 caps them
	// below 24, and a time.Time with a larger one cannot be marshaled again.
	if _, offset := t.Zone(); parseErr == nil && (offset <= -24*60*60 || offset >= 24*60*60) {
		parseErr = mx.DateTimeFormatError(tag, raw)
	}
	// The range applies on the server too, not only in the browser. A value
	// outside it is still bound, as rejected text is
	// (see [setTimeText]), so every re-render shows what was posted, also
	// after an add-row click, which discards the errors, and in an added
	// repeatable row, which has no stored value to fall back on. The form
	// handler never calls onSubmit while a field has an error, so the value
	// is not saved.
	var rangeErr error
	if parseErr == nil {
		shownTag := tag
		shownTag.Sensitive = false
		in := mx.DateTimeInput(value, shownTag)
		shown := in.Time
		layout := mx.DateTimeInputLayout(tag)
		// The input shows a stored time without its zone, so reading it
		// back could move it, as in the hour repeated when clocks go back:
		// an unchanged submission keeps the stored value as it is.
		if !zoned && !shown.IsZero() && t.Format(layout) == shown.Format(layout) {
			return mx.CheckDateTimeRange(tag, raw, t)
		}
		t = inShownLocation(t, zoned, shown, in.Type)
		// Checked as the form will show it, after a posted offset is
		// converted and a time in a DST gap moved.
		rangeErr = mx.CheckDateTimeRange(tag, raw, t)
	} else if tag.Sensitive {
		// A sensitive value is not echoed through the parse error. The format
		// is all it says, and the type may still read raw in its own spelling,
		// so it is not reported; a user message is not masked again.
		parseErr = mx.DateTimeFormatError(tag, raw)
	}
	// A database/sql type holding a time, such as sql.NullTime, scans it.
	if isScanner(value) {
		if parseErr != nil {
			return mx.DateTimeFormatError(tag, raw)
		}
		scanned := reflect.New(value.Type())
		if err := scanned.Interface().(scanner).Scan(t); err != nil {
			return mx.DateTimeFormatError(tag, raw)
		}
		value.Set(scanned.Elem())
		return rangeErr
	}
	timeType := reflect.TypeFor[time.Time]()
	if value.Type() == timeType || (value.Kind() == reflect.Struct && value.Type().ConvertibleTo(timeType)) {
		if parseErr != nil {
			return parseErr
		}
		// A type defined as time.Time takes the time converted.
		value.Set(reflect.ValueOf(t).Convert(value.Type()))
		return rangeErr
	}
	if reflect.PointerTo(value.Type()).Implements(reflect.TypeFor[encoding.TextUnmarshaler]()) {
		if err := unmarshalTime(ctx, value, raw, t, parseErr, tag); err != nil {
			return err
		}
		if parseErr != nil && (tag.Min != "" || tag.Max != "") {
			return checkOwnTextRange(value, raw, tag)
		}
		return rangeErr
	}
	if parseErr != nil {
		return parseErr
	}
	return fmt.Errorf("cannot assign time to %s", value.Type())
}

// inShownLocation puts a posted time into the location the input showed
// the stored time in, so the wall clock means what the user saw. A date is
// the first instant of its day, which is not midnight where the clocks
// jump over it, and a time keeps the stored date. A time posted with an
// offset keeps its instant, in that location, so a re-render shows it as
// the stored times are shown. Without a stored time, as on a new record,
// the posted time stays as parsed: a zone-less one is UTC. inputType is
// the input's type attribute, as [mx.DateTimeInputSpec.Type].
func inShownLocation(t time.Time, zoned bool, shown time.Time, inputType string) time.Time {
	if shown.IsZero() {
		return t
	}
	loc := shown.Location()
	switch {
	case zoned:
		return t.In(loc)
	case inputType == "time":
		return time.Date(shown.Year(), shown.Month(), shown.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
	case inputType == "date":
		day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		for range 3 {
			if day.Day() == t.Day() {
				break
			}
			day = day.Add(time.Hour)
		}
		return day
	}
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

// checkOwnTextRange checks form:"min=…,max=…" for a value that
// [unmarshalTime] bound from raw in the type's own spelling, which no input
// layout reads. The range is checked on the value as [mx.DateTimeInput]
// shows it. A value it can only show as text, or not at all, as when its
// MarshalText fails, cannot be checked, so it is rejected with the error
// that names the format to enter.
func checkOwnTextRange(value reflect.Value, raw string, tag mx.FormTag) error {
	shown := tag
	shown.Sensitive = false
	in := mx.DateTimeInput(value, shown)
	if in.TextFallback() || in.Value == "" {
		return mx.DateTimeFormatError(tag, raw)
	}
	_, err := mx.ParseDateTimeInput(tag, in.Value)
	return err
}

// unmarshalTime binds the submitted raw text into a type with UnmarshalText,
// trying in turn:
//
//  1. raw in the type's own form, such as a date struct's 2006-01-02 or the
//     text a text input showed because the picker could not, so an
//     unchanged submit binds back to the stored value. A type whose own
//     form is RFC 3339, as a struct embedding time.Time, is never shown in
//     it, so it only takes what the input layouts read;
//  2. t, the time raw was parsed into (parseErr nil), in the input type's
//     layout with seconds, which a time-of-day type may need where the
//     input left zero seconds out;
//  3. t in RFC 3339, the only text an embedded time.Time reads and the
//     text of a nullable time such as null.Time.
//
// The promoted UnmarshalText of an embedded time.Time sets only the time,
// not a validity flag next to it, so a result of any step that reports
// itself null ([mx.IsNull]) was half set and is refused.
//
// A method that panics, as the promoted method of a nil embedded pointer
// does, is an error. A form:"sensitive" field's error does not echo raw
// ([mx.MaskSensitiveError]).
func unmarshalTime(ctx context.Context, value reflect.Value, raw string, t time.Time, parseErr error, tag mx.FormTag) (err error) {
	typ := value.Type()
	defer func() {
		if p := recover(); p != nil {
			mx.ReportHiddenError(ctx, errs.AsErrorWithDebugStack(p))
			err = fmt.Errorf("%q cannot be stored as %s", raw, typ)
		}
		if err != nil && tag.Sensitive {
			err = maskDateTimeError(ctx, err, dateFormatError(value, tag, raw))
		}
	}()
	if parseErr != nil && mx.TextIsRFC3339(typ) {
		return parseErr
	}
	// Unmarshal into a copy: a failing UnmarshalText may still change its
	// value, as time.Time's zeroes it, and the field must keep the stored
	// value when the submission is rejected. A *time.Time field gets a copy
	// of its time, so a promoted UnmarshalText never writes the time the
	// stored value points to, and a nil embedded one a zero time to take
	// the submission, where its promoted methods would panic.
	halfSet := false
	try := func(text string) error {
		dst := reflect.New(typ)
		dst.Elem().Set(value)
		if typ.Kind() == reflect.Struct {
			for i := range typ.NumField() {
				f := dst.Elem().Field(i)
				if f.Type() != reflect.TypeFor[*time.Time]() || !f.CanSet() {
					continue
				}
				switch {
				case !f.IsNil():
					f.Set(reflect.ValueOf(new(*f.Interface().(*time.Time))))
				case typ.Field(i).Anonymous:
					f.Set(reflect.ValueOf(new(time.Time)))
				}
			}
		}
		if err := dst.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(text)); err != nil {
			return err
		}
		if mx.IsNull(dst.Interface()) {
			halfSet = true
			return errors.New("half set")
		}
		value.Set(dst.Elem())
		return nil
	}
	err = try(raw)
	if err == nil {
		return nil
	}
	if parseErr == nil && (try(t.Format(mx.DateTimeInputLayout(tag))) == nil || try(t.Format(time.RFC3339Nano)) == nil) {
		return nil
	}
	if halfSet {
		return fmt.Errorf("%q cannot be stored as %s", raw, typ)
	}
	return err
}

// setTimeText binds a submitted date or time into a string-kind field
// that holds it as text, the counterpart of the
// text values [mx.DateTimeInput] prefills. [mx.ParseDateTimeInput] checks
// the submission and gives the text to store. An empty submission stores
// the zero value, which a required field rejects. A valid value goes
// through UnmarshalText when the type has it and is set as the string
// otherwise.
//
// A rejected submission is a validation error, and the field keeps the
// submitted text so every re-render shows what was posted, also after an
// add-row click, which discards the errors, and in an added repeatable
// row. The form handler never calls onSubmit while a field has an error,
// so the rejected text is not saved.
func setTimeText(ctx context.Context, value reflect.Value, raw string, tag mx.FormTag) (bindErr error) {
	// An UnmarshalText that panics rejects the value, which stays bound as
	// posted, as other rejected text does.
	defer func() {
		if p := recover(); p != nil {
			mx.ReportHiddenError(ctx, errs.AsErrorWithDebugStack(p))
			value.SetString(raw)
			bindErr = mx.DateTimeFormatError(tag, raw)
		}
	}()
	// A type that reads only its own spelling, such as a type DEDate string
	// taking 02.01.2006, renders as a text input of its own text (its input
	// asks for no Format), so its UnmarshalText reads what was posted.
	if mx.DateTimeInput(value, tag).Format == "" {
		if raw == "" {
			value.SetZero()
			return nil
		}
		// No form:"min=…,max=…" can be checked on its spelling. The form
		// renders the tag as an error, and a scripted POST that gets past it
		// is rejected rather than saved unchecked.
		if err := mx.DateTimeTagError(value, tag); err != nil {
			value.SetString(raw)
			return err
		}
		parsed := reflect.New(value.Type())
		if err := parsed.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(raw)); err != nil {
			value.SetString(raw)
			if tag.Sensitive {
				return mx.MaskSensitiveError(ctx, err, "not a valid value")
			}
			return err
		}
		value.Set(parsed.Elem())
		return nil
	}
	text, err := mx.ParseDateTimeInput(tag, raw)
	if err != nil {
		value.SetString(raw)
		return err
	}
	if text == "" {
		value.SetZero()
		return nil
	}
	parsed := reflect.New(value.Type())
	if _, ok := parsed.Interface().(encoding.TextUnmarshaler); ok {
		err := parsed.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(text))
		if err != nil && strings.Count(text, ":") == 1 {
			// A type whose text always carries seconds, as civil.Time's
			// does, reads the value with the ":00" that the shortest form
			// leaves out for zero seconds (in which a time has one colon,
			// a date none).
			parsed = reflect.New(value.Type())
			if parsed.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(text+":00")) == nil {
				err = nil
			}
		}
		if err != nil {
			value.SetString(raw)
			if tag.Sensitive {
				return maskDateTimeError(ctx, err, mx.DateTimeFormatError(tag, raw))
			}
			return err
		}
	} else {
		parsed.Elem().SetString(text)
	}
	value.Set(parsed.Elem())
	return nil
}

func setEnumSet(value reflect.Value, fieldType reflect.Type, vals []string) error {
	keyType := mx.SetElemType(fieldType)
	if keyType == nil {
		return errors.New("not a recognized set type")
	}
	if value.Kind() == reflect.Pointer {
		return bindPointee(value, func(v reflect.Value) error { return setEnumSet(v, value.Type().Elem(), vals) })
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
		return fmt.Errorf("cannot store a set in a %s", value.Kind())
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
	return reflect.Value{}, fmt.Errorf("cannot convert %s to %s", quotedValue(s), t)
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

// isPlainNumber reports whether t is an integer or float type that reads
// no text of its own.
func isPlainNumber(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return !reflect.PointerTo(t).Implements(reflect.TypeFor[encoding.TextUnmarshaler]())
	}
	return false
}

// groupLabelID is the id of the label that names a field's group of
// controls, such as an enum set's checkboxes.
func groupLabelID(path mx.FieldPath) string {
	return string(path) + "-label"
}

func clearSentinelInput(path mx.FieldPath, label string, checked bool) mx.Component {
	box := []mx.Attrib{
		Type("checkbox"),
		Name(mx.ClearSentinelName(path)),
		Value("1"),
		// The visible word comes first in the name, which says which field
		// a form full of clear boxes, as in repeatable rows, clears.
		mx.NewAttrib("aria-label", strings.TrimSpace("clear "+label)),
	}
	if checked {
		box = append(box, Checked)
	}
	return Label(
		Input(box...),
		mx.Text(" clear"),
	)
}
