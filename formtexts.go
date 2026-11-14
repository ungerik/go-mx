package mx

import (
	"cmp"
	"context"
)

// FormTexts holds every user-visible string that [ReflectFormHandler]
// renders on its own, without a decider or the caller passing it: the
// submit button's caption, the generic form-level error, the notice after a
// row command, and the row buttons and empty line of a repeatable field.
//
// The strings are resolved from the request's context, so one process can
// serve several languages concurrently: install the request's FormTexts with
// [ContextWithFormTexts] and every reflected form handling that request uses
// them, with no change to any handler. [DefaultFormTexts] is used when a
// context carries none.
//
// Fields may be left empty: each one falls back to the corresponding
// DefaultFormTexts field and then to the built-in English, so a partial
// translation renders a default for the strings it does not cover instead
// of an empty caption.
//
// A renderer's own strings, such as the shadcn renderer's shadcn.Labels,
// are installed separately; a translated shadcn form installs both.
type FormTexts struct {
	// Submit is the caption of the form's submit button.
	// [ReflectFormConfig].SubmitLabel overrides it for one handler.
	Submit string

	// Error is the form-level message shown when onSubmit returns an error
	// that is neither [FieldErrors] nor a [UserMessageError]: such an error
	// may carry internal details and a call stack, so its text is shown only
	// when [RevealInternalServerErrors] is set (see [FormErrorText]).
	Error string

	// RetypeSensitive is the form-level message an add-row or remove-row
	// click re-renders the form with when text was typed into a
	// form:"sensitive" input: the input never shows its value, and the save
	// that follows reads the input, so the text must be typed again.
	RetypeSensitive string

	// AddRow and RemoveRow are the captions of a repeatable field's
	// add-row button and of the remove button on each of its rows.
	AddRow    string
	RemoveRow string

	// NoRows is the line a form:"readonly" repeatable field with no rows
	// shows in its fieldset, which has no add-row button, so that "no
	// entries" does not look like a broken render.
	NoRows string
}

// DefaultFormTexts are the [FormTexts] used when a context carries none from
// [ContextWithFormTexts], and to fill the fields a context's FormTexts leave
// empty. Reassign it at init time to change the texts program-wide; a
// partial assignment keeps the built-in English for the fields it leaves
// empty.
var DefaultFormTexts = englishFormTexts

// englishFormTexts is the last fallback of [FormTexts.orDefaults], so that
// no field resolves to "" even after a partial [DefaultFormTexts]
// assignment.
var englishFormTexts = FormTexts{
	Submit:          "Save",
	Error:           "The form could not be saved.",
	RetypeSensitive: "A value typed into a hidden input, such as a password, is not shown again: type it again before saving.",
	AddRow:          "Add row",
	RemoveRow:       "Remove",
	NoRows:          "None",
}

// ctxKeyFormTexts is the context key of [ContextWithFormTexts].
type ctxKeyFormTexts struct{}

// ContextWithFormTexts returns a derived context that carries texts, so the
// reflected forms of a request using it show those strings instead of
// [DefaultFormTexts]. Wire it once per request to serve several languages
// from one process:
//
//	next.ServeHTTP(w, r.WithContext(mx.ContextWithFormTexts(r.Context(), formTextsFor(lang))))
func ContextWithFormTexts(ctx context.Context, texts FormTexts) context.Context {
	return context.WithValue(ctx, ctxKeyFormTexts{}, texts)
}

// FormTextsFromContext returns the [FormTexts] installed by
// [ContextWithFormTexts] with every empty field filled from
// [DefaultFormTexts], or DefaultFormTexts itself if the context carries none
// or is nil. Callers therefore always get a complete set.
func FormTextsFromContext(ctx context.Context) FormTexts {
	var texts FormTexts
	if ctx != nil {
		texts, _ = ctx.Value(ctxKeyFormTexts{}).(FormTexts)
	}
	return texts.orDefaults()
}

// orDefaults returns t with every empty field replaced by the corresponding
// [DefaultFormTexts] field, and then by the englishFormTexts field.
// TestFormTextsOrDefaultsCoversAllFields fails if a new field is not added
// here.
func (t FormTexts) orDefaults() FormTexts {
	return FormTexts{
		Submit:          cmp.Or(t.Submit, DefaultFormTexts.Submit, englishFormTexts.Submit),
		Error:           cmp.Or(t.Error, DefaultFormTexts.Error, englishFormTexts.Error),
		RetypeSensitive: cmp.Or(t.RetypeSensitive, DefaultFormTexts.RetypeSensitive, englishFormTexts.RetypeSensitive),
		AddRow:          cmp.Or(t.AddRow, DefaultFormTexts.AddRow, englishFormTexts.AddRow),
		RemoveRow:       cmp.Or(t.RemoveRow, DefaultFormTexts.RemoveRow, englishFormTexts.RemoveRow),
		NoRows:          cmp.Or(t.NoRows, DefaultFormTexts.NoRows, englishFormTexts.NoRows),
	}
}
