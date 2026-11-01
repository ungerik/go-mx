package mx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

var (
	// RevealInternalServerErrors controls whether [RespondNonContextError] and
	// [SSEResponse.SendError] include the actual error message in what they
	// send, and whether a reflected form shows the text of an onSubmit
	// error instead of [FormErrorMessage] ([FormErrorText]). It defaults to
	// false, which sends a generic "Internal Server Error"; set it to true
	// (typically in development) to expose the error text. [ReportError]
	// gets the error either way.
	RevealInternalServerErrors = false

	// ReportError, when not nil, gets every error go-mx hides from the
	// client behind a generic text, so the application can log it:
	//
	//   - the 500 of [RespondNonContextError], which also answers the render
	//     errors of [ComponentHTTPHandler], [ComponentFuncHandler] and the
	//     reflected form handlers and the panics the first two recover, and
	//     the 500 a reflected form answers when its load function fails;
	//   - the in-band event of [SSEResponse.SendError];
	//   - the 500 of pdf.Document.ServeHTTP;
	//   - an error a reflected form shows as [FormErrorText]: one onSubmit
	//     returned, an error no input of the form shows, and a context option
	//     list that failed while a POST was parsed;
	//   - the errors of a form:"sensitive" field, which read "not a valid
	//     value", for a date or time the format to enter. Their text may quote
	//     the value the form keeps off the page, so a hook that logs them must
	//     treat it as secret as the value.
	//
	// It gets each such error once, whether or not
	// [RevealInternalServerErrors] is set, but not an error the user sees: the
	// [UserMessageError] a form shows, or an error shown at its field. (A 500
	// or an SSE error event shows no user message, so they report a
	// UserMessageError like any other error.) Nor does it get a
	// context.Canceled or context.DeadlineExceeded of a request whose context
	// is done, because the client that would have been told has disconnected
	// (one a call's own timeout returned while the client waits is reported);
	// the error of a request body a
	// reflected form cannot read (its 400, 413 or 415), which is the client's;
	// or the errors of the parse a repeatable field's add-row or remove-row
	// click runs, which the form discards and the next save reports. A panic
	// [ComponentHTTPHandler] or [ComponentFuncHandler] recovers arrives as an
	// error carrying the debug stack of the panic. The page tools of the mcp
	// module report to their own PageToolsOptions.OnError instead.
	//
	// ctx is the request's context without its cancellation: it carries the
	// request's values, such as a request-scoped logger, and a client that
	// went away does not cut a report short. ReportError is called from the
	// goroutines serving requests, before go-mx answers the request, so it
	// must be safe for concurrent use, return quickly and not panic; set it
	// once at start-up. A decider or handler of the application hiding an
	// error reports it through [ReportHiddenError].
	ReportError func(ctx context.Context, err error)
)

// UserMessageError is an error whose message is meant for the user, such as
// a rule a submitted value breaks. Where go-mx hides the text of any other
// error, it shows the UserMessage instead: as a reflected form's message for
// an onSubmit error ([FormErrorText]), and for a form:"sensitive" field. A
// field shows it in place of the error's own text ([FieldErrorTexts]), so the
// Error text of a wrapper never reaches the page.
//
// go-mx recognizes the method set wherever [errors.As] finds it in an
// error's chain, also through errors.Join, so a domain package marks its
// user-fixable errors with a UserMessage method without importing go-mx.
// [FormMessage] is a ready-made one.
//
// A form:"sensitive" field shows the UserMessage too, so it must not quote
// the value it rejects, which after a blank submission is the stored secret.
// An empty UserMessage counts as none: the error is hidden like any other.
type UserMessageError interface {
	error

	// UserMessage returns the text shown to the user. It stands alone: the
	// Error text of the error, or of one wrapping it, may add internal
	// context the user must not see.
	UserMessage() string
}

// UserMessage returns the first UserMessage in the tree of err that is not
// empty, of a [UserMessageError] in its chain, also through errors.Join, or
// one an As method of an error in it supplies, as [errors.As] finds them,
// and whether there is one. An empty message does not hide one deeper in
// the tree. It is the test go-mx applies wherever it decides whether the
// user may read an error.
func UserMessage(err error) (message string, ok bool) {
	if message, ok := ownUserMessage(err); ok {
		return message, true
	}
	switch wrapped := err.(type) {
	case interface{ Unwrap() error }:
		return UserMessage(wrapped.Unwrap())
	case interface{ Unwrap() []error }:
		for _, part := range wrapped.Unwrap() {
			if message, ok := UserMessage(part); ok {
				return message, true
			}
		}
	}
	return "", false
}

// ownUserMessage returns the UserMessage of err itself, not of an error it
// wraps: that of err as a [UserMessageError], or of the one its As method
// supplies, and whether it is not empty.
func ownUserMessage(err error) (message string, ok bool) {
	if m, isMessage := err.(UserMessageError); isMessage {
		if message := m.UserMessage(); message != "" {
			return message, true
		}
	}
	if x, hasAs := err.(interface{ As(any) bool }); hasAs {
		var m UserMessageError
		if x.As(&m) && m != nil {
			if message := m.UserMessage(); message != "" {
				return message, true
			}
		}
	}
	return "", false
}

// userMessages returns every user message in the tree of err, in the order
// [FieldErrorTexts] shows them, and none of the text of the errors around
// or next to them: each as the error itself where its text is its message,
// as with a [FormMessage] or a [DateTimeError], so a renderer can still read
// its type, else as a FormMessage of the message.
func userMessages(err error) []error {
	if message, ok := ownUserMessage(err); ok {
		if err.Error() == message {
			return []error{err}
		}
		return []error{FormMessage(message)}
	}
	switch wrapped := err.(type) {
	case interface{ Unwrap() error }:
		return userMessages(wrapped.Unwrap())
	case interface{ Unwrap() []error }:
		var messages []error
		for _, part := range wrapped.Unwrap() {
			messages = append(messages, userMessages(part)...)
		}
		return messages
	}
	return nil
}

// FieldErrorTexts returns the messages a field shows for err, one of its
// errors, one text each: the UserMessage of a [UserMessageError], which
// speaks for everything it wraps, else the error's own text, as a field error
// that is no user message is meant to be read where it is shown. Where a user
// message is under it, a wrapper's own text is left out, and a joined error
// gives the messages of each of its parts, also when it is wrapped, so two
// joined user messages both show; an error with none under it shows its own
// text whole, or "not a valid value" for an error without text, so a field
// marked invalid always says something. A decider renders a field's errors
// with it.
func FieldErrorTexts(err error) []string {
	if err == nil {
		return nil
	}
	if message, ok := ownUserMessage(err); ok {
		return []string{message}
	}
	if _, ok := UserMessage(err); ok {
		switch wrapped := err.(type) {
		case interface{ Unwrap() []error }:
			var texts []string
			for _, part := range wrapped.Unwrap() {
				// A part without text, such as an empty message, says
				// nothing next to one that does.
				if part != nil && part.Error() != "" {
					texts = append(texts, FieldErrorTexts(part)...)
				}
			}
			return texts
		case interface{ Unwrap() error }:
			return FieldErrorTexts(wrapped.Unwrap())
		}
	}
	if text := err.Error(); text != "" {
		return []string{text}
	}
	return []string{"not a valid value"}
}

// BoundFieldError is err, an error a field shows for the posted text raw,
// bounded by [MaxFieldErrorLen] part by part, as [FieldErrorTexts] shows
// them: a part whose text is longer than the bound, a user message
// included, reads substitute, and it is reported when raw is within the
// bound, as its text was then not built from a forged value; for a longer
// raw a stand-in is reported, naming the error's type and the lengths of
// its text and of raw, so a legitimate long value's long error is not lost.
// For raw longer than 64 times the bound the text of a part without a user
// message is not built at all, as it may quote raw at several times its
// size, so even a short one reads substitute. A short part of a long raw
// keeps its text and nothing else, so what it holds does not reach
// [ReportError] either. The
// substitute is a [FormMessage], meant for the user, so it is not reported
// again. go-mx bounds every parse and validation error of a reflected form
// field with it; a decider that builds an error from posted text can call it
// too.
func BoundFieldError(ctx context.Context, err error, raw string, substitute FormMessage) error {
	bounded, _ := boundFieldError(ctx, err, raw, substitute)
	return bounded
}

// boundFieldError is [BoundFieldError], reporting whether it changed err,
// which an error type that is not comparable cannot tell by ==.
func boundFieldError(ctx context.Context, err error, raw string, substitute FormMessage) (error, bool) {
	maxLen := MaxFieldErrorLen
	if err == nil || maxLen <= 0 {
		return err, false
	}
	// The same walk as FieldErrorTexts: an error's own user message, else
	// the parts of an error holding one, else its text.
	if message, ok := ownUserMessage(err); ok {
		if len(message) <= maxLen {
			return err, false
		}
		reportHiddenFieldError(ctx, err, len(message), raw)
		return substitute, true
	}
	if _, ok := UserMessage(err); ok {
		switch wrapped := err.(type) {
		case interface{ Unwrap() []error }:
			parts := wrapped.Unwrap()
			bounded := make([]error, len(parts))
			changed := false
			for i, part := range parts {
				var partChanged bool
				bounded[i], partChanged = boundFieldError(ctx, part, raw, substitute)
				changed = changed || partChanged
			}
			if !changed {
				return err, false
			}
			return errors.Join(bounded...), true
		case interface{ Unwrap() error }:
			// Only the wrapped error is shown, not the text around it.
			if bounded, changed := boundFieldError(ctx, wrapped.Unwrap(), raw, substitute); changed {
				return bounded, true
			}
			return err, false
		}
	}
	if len(raw)/64 > maxLen {
		reportHiddenFieldError(ctx, err, -1, raw)
		return substitute, true
	}
	text := err.Error()
	switch {
	case len(text) > maxLen:
		reportHiddenFieldError(ctx, err, len(text), raw)
		return substitute, true
	case len(raw) > maxLen:
		return errors.New(text), true
	}
	return err, false
}

// reportHiddenFieldError reports err, a field's error [BoundFieldError]
// hides for being longer than the bound, textLen bytes long or -1 when its
// text was not built: err itself for raw within the bound, otherwise a
// stand-in naming its type and the lengths, which quotes nothing of a value
// that may be forged at any size.
func reportHiddenFieldError(ctx context.Context, err error, textLen int, raw string) {
	if len(raw) <= MaxFieldErrorLen {
		ReportHiddenError(ctx, err)
		return
	}
	text := "not built"
	if textLen >= 0 {
		text = strconv.Itoa(textLen) + " bytes"
	}
	ReportHiddenError(ctx, fmt.Errorf("a field error of type %T was hidden for its length (text %s) for %d bytes of posted text", err, text, len(raw)))
}

// ReportHiddenError passes err, an error the client is shown a generic text
// for instead, to [ReportError] with ctx stripped of its cancellation. It
// does nothing when ReportError or err is nil, when err is a
// context.Canceled or context.DeadlineExceeded while ctx is done, the
// client having disconnected, or during the parse of a
// repeatable field's row command, whose errors the form discards. go-mx
// calls it wherever it hides an error; a decider or handler of the
// application that hides one calls it too.
func ReportHiddenError(ctx context.Context, err error) {
	report := ReportError
	if report == nil || err == nil || clientGone(ctx, err) || ctx.Value(unreportedKey{}) != nil {
		return
	}
	report(context.WithoutCancel(ctx), err)
}

// MaskSensitiveError is the error a form:"sensitive" field shows in place of
// err, whose text may quote the submitted or stored value: the user
// messages in it, which quote nothing, joined when there are several,
// without what they were wrapped or joined with, which is reported; each is
// the error itself where its text is its message, as a [DateTimeError]'s
// is, else a [FormMessage]. Without one it is the FormMessage mask, with
// err reported ([ReportHiddenError]). Either is meant for the user, so a
// later stage neither masks nor reports it again.
func MaskSensitiveError(ctx context.Context, err error, mask string) error {
	if messages := userMessages(err); len(messages) > 0 {
		reportHiddenParts(ctx, err)
		if len(messages) == 1 {
			return messages[0]
		}
		return errors.Join(messages...)
	}
	ReportHiddenError(ctx, err)
	return FormMessage(mask)
}

// unreportedKey marks a context whose hidden errors [ReportHiddenError] does
// not report (see [withoutReports]).
type unreportedKey struct{}

// withoutReports returns ctx marked so that [ReportHiddenError] reports
// nothing, for a parse whose errors are discarded.
func withoutReports(ctx context.Context) context.Context {
	return context.WithValue(ctx, unreportedKey{}, true)
}

// reportHiddenParts reports the parts of err, an error the user is shown
// through a user message or [FieldErrors] in it, that are neither: each part
// of a joined error that holds no UserMessage and no FieldErrors is an
// internal error the page does not show, such as a failed audit write
// joined with a validation result ([ReportHiddenError]). The text wrapping a
// user message is not reported, as that would report every validation
// failure, nor is an error whose own UserMessage speaks for what it wraps.
func reportHiddenParts(ctx context.Context, err error) {
	if _, ok := ownUserMessage(err); ok {
		return
	}
	if _, ok := err.(FieldErrors); ok {
		return
	}
	switch wrapped := err.(type) {
	case interface{ Unwrap() []error }:
		for _, part := range wrapped.Unwrap() {
			_, hasMessage := UserMessage(part)
			if _, hasFieldErrors := errors.AsType[FieldErrors](part); hasMessage || hasFieldErrors {
				reportHiddenParts(ctx, part)
			} else {
				ReportHiddenError(ctx, part)
			}
		}
	case interface{ Unwrap() error }:
		reportHiddenParts(ctx, wrapped.Unwrap())
	}
}

// clientGone reports whether err is the cancellation of ctx, the request's
// context: a context.Canceled or context.DeadlineExceeded while ctx is done,
// which go-mx takes for a client that has disconnected. A deadline a call
// set for itself, such as a database query under context.WithTimeout, or a
// cancel of its own, as fan-out results collected after one give, fails
// with the same errors while the client still waits: that is an error of
// the server.
func clientGone(ctx context.Context, err error) bool {
	return ctx.Err() != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}

// hiddenErrorText returns the text the client is shown in place of err, an
// error not meant for it, and reports err ([ReportHiddenError]): generic, or
// the error's own text when [RevealInternalServerErrors] is set. ok is false
// for a context.Canceled or context.DeadlineExceeded while ctx is done,
// which is neither shown nor reported, because the client that would read it
// has disconnected ([clientGone]).
//
// It is the policy of the 500s go-mx answers ([RespondNonContextError],
// respondLoadError) and of [SSEResponse.SendError] (which sends the text as
// an in-band event, because by then the 200 is committed), so they cannot
// drift apart. [FormErrorText] reveals and reports the same way, but always
// returns a text, because the form is re-rendered either way.
func hiddenErrorText(ctx context.Context, err error, generic string) (text string, ok bool) {
	if clientGone(ctx, err) {
		return "", false
	}
	ReportHiddenError(ctx, err)
	if RevealInternalServerErrors {
		return err.Error(), true
	}
	return generic, true
}

// RespondNonContextError responds with an internal server error and reports
// err to [ReportError] with the context of r, unless err is a
// context.Canceled or context.DeadlineExceeded and the context of r is done,
// in which case it does nothing because the client has disconnected. A
// timeout of a call the handler made, while the client still waits, is a
// 500 like any other error. If
// [RevealInternalServerErrors] is true, the error message is included in the
// response, otherwise a generic "Internal Server Error" is used.
func RespondNonContextError(w http.ResponseWriter, r *http.Request, err error) {
	if message, ok := hiddenErrorText(r.Context(), err, "Internal Server Error"); ok {
		http.Error(w, message, http.StatusInternalServerError)
	}
}

// func respondMarshalJSON(w http.ResponseWriter, v any) {
// 	j, err := json.MarshalIndent(v, "", "  ")
// 	if err != nil {
// 		RespondError(w, err)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(j)
// }
