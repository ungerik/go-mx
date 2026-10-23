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
	"math/big"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

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

	input := renderInput(path, field, value, kind, browserTag(tag, kind, value), errs)
	if input == nil {
		return nil
	}

	parts := mx.Components{}
	if labelText := fieldLabel(field, tag); labelText != "" && kind == mx.FieldKindEnumSet {
		// A set has no control with the path as its id: the label names
		// its group.
		parts = append(parts, Label(ID(groupLabelID(path)), labelText+":"))
	} else if labelText != "" && kind != mx.FieldKindBool {
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
	// A readonly field is never written, so it offers no clear control.
	// A blank sensitive input keeps the stored value, so the clear control
	// is how it is emptied, ticked while it is empty so a re-render keeps
	// a clear the user asked for.
	sensitive := tag.Sensitive && kind.KeepsStoredWhenBlank()
	if (isNullable(value) || sensitive) && !tag.Required && !tag.Readonly {
		parts = append(parts, clearSentinelInput(path, fieldLabel(field, tag), sensitive && !mx.KeepsStoredValue(tag, kind, value)))
	}
	if len(errs) > 0 {
		for _, e := range errs {
			for _, text := range mx.FieldErrorTexts(e) {
				parts = append(parts, mx.NewElement("p",
					mx.Attribute{Name: "class", Value: "form-error"},
					mx.Attribute{Name: "data-error", Value: string(path)},
					mx.Text(text),
				))
			}
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

// renderHidden writes the value as displayValue does; a hidden time, such
// as an UpdatedAt kept for optimistic locking, round-trips as RFC 3339.
func renderHidden(path mx.FieldPath, value reflect.Value) mx.Component {
	return VoidElement("input",
		Type("hidden"),
		Name(string(path)),
		Value(displayValue(value, mx.FormTag{})),
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
	}
	// A sensitive number never shows its value, like a sensitive text input.
	if !tag.Sensitive {
		attribs = append(attribs, Value(displayValue(value, tag)))
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

func fileInput(path mx.FieldPath, tag mx.FormTag, errs []error) mx.Component {
	attribs := []mx.Attrib{
		Type("file"),
		Name(string(path)),
		ID(string(path)),
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
	options := optionChildren(value, tag, field.Type, func(opt mx.NamedOption) mx.Component {
		oAttr := []any{Value(opt.Value)}
		if opt.Value == selected {
			oAttr = append(oAttr, Selected)
		}
		oAttr = append(oAttr, mx.Text(displayLabel(opt)))
		return Option(oAttr...)
	}, func(options []mx.NamedOption) mx.Component {
		// Runs before the options render, so they and the preface see the
		// same selection: an unlisted zero selects the list's own empty
		// option, if it has one, or the preface's.
		selected = mx.SelectedOption(value, selected, options)
		preface := mx.SelectPreface(selected, tag, contextList, options)
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
	if fieldLabel(field, tag) != "" {
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
		return guardBind(r.Context(), value, raw, tag, func() error {
			if err := setScalar(value, raw); err != nil {
				return err
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
			return checkNumberRange(value, raw, tag)
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
			if keyType := setKeyType(field.Type); keyType != nil && mx.OptionsNeedContext(tag, keyType) {
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
		fh := files[0]
		// String destinations get the filename only.
		if value.Kind() == reflect.String {
			value.SetString(fh.Filename)
			return nil
		}
		if value.Kind() != reflect.Slice || value.Type().Elem().Kind() != reflect.Uint8 {
			return errors.New("file upload requires []byte or string field")
		}
		// The multipart parser keeps a large file on disk; reading it into
		// memory is bounded like the parse.
		if limit := mx.FormMaxMemory(r.Context()); fh.Size > limit {
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
			return fmt.Errorf("%q is not one of the allowed options", v)
		}
	}
	return nil
}

// guardBind runs bind, which writes raw into value and may call the type's
// own UnmarshalText. A method that panics, as the promoted method of a nil
// embedded pointer does, is an error instead of a failed request. Its error
// is bounded by [mx.MaxFieldErrorLen] ([boundFieldError]): an option list,
// UnmarshalText or strconv error quotes raw whole, into the page, or, where
// no input shows it, through [mx.ReportError] into the application's logs. A form:"sensitive"
// field's error reads "not a valid value" ([maskSensitive]), leaving raw out
// at any length. (Date and time fields guard their own panics and mask
// their errors with their format instead.)
func guardBind(ctx context.Context, value reflect.Value, raw string, tag mx.FormTag, bind func() error) (err error) {
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("%q cannot be stored as %s", raw, value.Type())
		}
		if err != nil {
			err = boundFieldError(ctx, err, raw)
		}
		if err != nil && tag.Sensitive {
			err = maskSensitive(ctx, err, "not a valid value")
		}
	}()
	return bind()
}

// boundFieldError is err, the error of binding the posted text raw, bounded
// by [mx.MaxFieldErrorLen] unless it carries a user message. An error whose
// text is longer than the bound reads "the text is not a valid value"; it is
// reported when raw is within the bound, as its text was then not built from
// a forged value. The text is not built at all for raw far longer than the
// bound, where it could only be a quote of raw several times its size. A
// short error of a long raw keeps its text and nothing else, so what it
// holds does not reach [mx.ReportError] either. The substitute is a
// [mx.FormMessage], meant for the user, so it is not reported again.
func boundFieldError(ctx context.Context, err error, raw string) error {
	maxLen := mx.MaxFieldErrorLen
	if _, ok := mx.UserMessage(err); ok || maxLen <= 0 {
		return err
	}
	const notValid = mx.FormMessage("the text is not a valid value")
	if len(raw) > 64*maxLen {
		return notValid
	}
	text := err.Error()
	switch {
	case len(text) > maxLen:
		if len(raw) <= maxLen {
			mx.ReportHiddenError(ctx, err)
		}
		return notValid
	case len(raw) > maxLen:
		return errors.New(text)
	}
	return err
}

// maskSensitive is the error a form:"sensitive" field shows in place of err,
// whose text may quote the submitted or stored value: the UserMessage of an
// [mx.UserMessageError] in it, which quotes nothing, without what it was
// wrapped or joined with, else mask, with err reported to [mx.ReportError].
// Either is a [mx.FormMessage], meant for the user, so a later stage neither
// masks nor reports it again.
func maskSensitive(ctx context.Context, err error, mask string) error {
	if msg, ok := mx.UserMessage(err); ok {
		return mx.FormMessage(msg)
	}
	mx.ReportHiddenError(ctx, err)
	return mx.FormMessage(mask)
}

// maxNumberLen is the longest text a number field takes.
const maxNumberLen = 64

// dateFormatError is [mx.DateTimeFormatError] for raw, except for a type
// that reads only its own spelling, whose format the error cannot name
// (its [mx.DateTimeInput] has no Format): the input's format would send
// the user to text the type rejects.
func dateFormatError(value reflect.Value, tag mx.FormTag, raw string) error {
	if value.IsValid() && mx.DateTimeInput(value, tag).Format == "" {
		return errors.New("not a valid value")
	}
	return mx.DateTimeFormatError(tag, raw)
}

// maxDateTimeLen is the longest text a date, datetime or time field takes,
// longer than any input sends or a date type spells a date.
const maxDateTimeLen = 64

// checkNumberRange returns a validation error when value, the number just
// bound from raw, lies outside the tag's form:"min=…,max=…". The bound
// value is compared, never raw itself, so whatever text the parser read is
// checked as it is stored, including the 0 a blank stores in a number that
// is no pointer (a nil pointer is no number to check). An integer is
// compared exactly, so an int64 bound above 2^53 holds. A bound that is no
// valid HTML floating-point number is ignored, as a browser ignores it.
func checkNumberRange(value reflect.Value, raw string, tag mx.FormTag) error {
	if tag.Min == "" && tag.Max == "" {
		return nil
	}
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if raw == "" {
		raw = displayValue(value, tag) // the 0 the re-render shows
	}
	n := new(big.Rat)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n.SetInt64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n.SetUint64(value.Uint())
	case reflect.Float32, reflect.Float64:
		return checkFloatRange(value.Float(), value.Type().Bits(), raw, tag)
	default:
		return nil
	}
	if lowest, ok := numberBound(tag.Min); ok && n.Cmp(lowest) < 0 {
		return rangeError(raw, tag, true)
	}
	if highest, ok := numberBound(tag.Max); ok && n.Cmp(highest) > 0 {
		return rangeError(raw, tag, false)
	}
	return nil
}

// rangeError is the error of a number below (or above) its form:"min=…" (or
// "max=…"). A form:"sensitive" field's names only the bound, as an
// [mx.FormMessage], which the masking of its errors lets through.
func rangeError(raw string, tag mx.FormTag, below bool) error {
	switch {
	case tag.Sensitive && below:
		return mx.FormMessage("must be at least " + tag.Min)
	case tag.Sensitive:
		return mx.FormMessage("must be at most " + tag.Max)
	case below:
		return fmt.Errorf("%q is less than %s", raw, tag.Min)
	}
	return fmt.Errorf("%q is greater than %s", raw, tag.Max)
}

// checkFloatRange compares a float at its own precision, as a browser
// compares doubles: compared exactly, the binary value of 99.9 is above
// the decimal 99.9, so a value on its bound would be rejected. f is finite,
// as setNumeric rejects the rest.
func checkFloatRange(f float64, bits int, raw string, tag mx.FormTag) error {
	if lowest, ok := floatBound(tag.Min, bits); ok && f < lowest {
		return rangeError(raw, tag, true)
	}
	if highest, ok := floatBound(tag.Max, bits); ok && f > highest {
		return rangeError(raw, tag, false)
	}
	return nil
}

// floatBound reads a min or max as a browser does, rounded as the field's
// own values are, to a float of the given bits.
func floatBound(bound string, bits int) (float64, bool) {
	if !htmlNumber.MatchString(bound) {
		return 0, false
	}
	// Like a browser, a bound past the range of a double is no bound.
	if _, err := strconv.ParseFloat(bound, 64); err != nil {
		return 0, false
	}
	f, _ := strconv.ParseFloat(bound, bits) // past a float32's range: an infinity
	return f, true
}

// htmlNumber matches a valid HTML floating-point number, the only form a
// browser reads a min or max in.
var htmlNumber = regexp.MustCompile(`^-?(\d+|\d*\.\d+)([eE][-+]?\d+)?$`)

// numberBound reads a min or max as a browser does.
func numberBound(bound string) (*big.Rat, bool) {
	if !htmlNumber.MatchString(bound) {
		return nil, false
	}
	return new(big.Rat).SetString(bound)
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
// sql.NullInt64 that [mx.DriverValueText] writes, and so its Scan method
// reads back.
func isScanner(value reflect.Value) bool {
	if !value.CanAddr() {
		return false
	}
	_, ok := mx.DriverValueText(value)
	return ok
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
	if raw == displayValue(value, mx.FormTag{}) {
		return nil
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
		return checkNumberRange(reflect.ValueOf(n), raw, tag)
	}
	return nil
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
	subject := fmt.Sprintf("%q is", raw)
	if len(raw) > maxNumberLen {
		subject = "the text is"
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
	if value.Kind() == reflect.Pointer {
		if !value.CanSet() {
			return errors.New("pointer not settable")
		}
		return bindPointee(value, func(v reflect.Value) error {
			v.SetBool(on)
			return nil
		})
	}
	value.SetBool(on)
	return nil
}

func setTime(ctx context.Context, value reflect.Value, raw string, tag mx.FormTag) error {
	// Text longer than any date or time input sends is no date. It is
	// rejected before it is bound, so the re-render does not echo a
	// submission of any size, and the field keeps its stored value.
	if len(raw) > maxDateTimeLen {
		if tag.Sensitive {
			// The format is all it says, so it is not reported (see below).
			return mx.FormMessage(dateFormatError(value, tag, raw).Error())
		}
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
	case "date":
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
		// so it is not reported; a FormMessage is not masked again.
		parseErr = mx.FormMessage(mx.DateTimeFormatError(tag, raw).Error())
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
// ([maskSensitive]).
func unmarshalTime(ctx context.Context, value reflect.Value, raw string, t time.Time, parseErr error, tag mx.FormTag) (err error) {
	typ := value.Type()
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("%q cannot be stored as %s", raw, typ)
		}
		if err != nil && tag.Sensitive {
			err = maskSensitive(ctx, err, dateFormatError(value, tag, raw).Error())
		}
	}()
	if parseErr != nil && textIsRFC3339(typ) {
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

// textIsRFC3339 reports whether the zero value of t writes its text as
// RFC 3339, as a time.Time does and so does a type embedding one without
// text methods of its own. It mirrors the check of [mx.DetectField] rule 14.
func textIsRFC3339(t reflect.Type) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	m, isMarshaler := reflect.New(t).Interface().(encoding.TextMarshaler)
	if !isMarshaler {
		return false
	}
	text, err := m.MarshalText()
	if err != nil {
		return false
	}
	_, err = time.Parse(time.RFC3339Nano, string(text))
	return err == nil
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
		if recover() != nil {
			value.SetString(raw)
			bindErr = mx.DateTimeFormatError(tag, raw)
		}
	}()
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
				return maskSensitive(ctx, err, mx.DateTimeFormatError(tag, raw).Error())
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
	keyType := setKeyType(fieldType)
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
	return reflect.Value{}, fmt.Errorf("cannot convert %q to %s", s, t)
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
func displayValue(value reflect.Value, tag mx.FormTag) (text string) {
	_ = tag
	// A method that panics, as the promoted MarshalText of a nil embedded
	// pointer does, shows no value instead of failing the render.
	defer func() {
		if recover() != nil {
			text = ""
		}
	}()
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
	// A time, also of a type defined as one, is written as RFC 3339, the
	// text its parser reads back, zone and fraction included.
	if isTimeType(v.Type()) {
		t := v.Convert(reflect.TypeFor[time.Time]()).Interface().(time.Time)
		if t.IsZero() {
			return ""
		}
		return t.Format(time.RFC3339Nano)
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
	// A database/sql type such as sql.NullInt64 writes the value it holds,
	// which its Scan reads back (see setScanned).
	if text, ok := mx.DriverValueText(v); ok {
		return text
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

// browserTag is the tag the input renders with: a form:"sensitive" input
// that holds a stored value shows none, and a blank keeps the stored
// value, so the browser must not demand one; neither must a file input
// whose field holds a file. The server still rejects a blank required
// field when nothing is stored.
func browserTag(tag mx.FormTag, kind mx.FieldKind, value reflect.Value) mx.FormTag {
	if tag.Required && (mx.KeepsStoredValue(tag, kind, value) || kind == mx.FieldKindFile && hasStoredFile(value)) {
		tag.Required = false
	}
	return tag
}

// hasStoredFile reports whether a file field's value holds a file: bytes or
// a file name. A file input left empty keeps it, so a required one demands
// no new upload; the handler still rejects a blank with nothing stored.
func hasStoredFile(value reflect.Value) bool {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return false
		}
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.String, reflect.Slice:
		return value.Len() > 0
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
