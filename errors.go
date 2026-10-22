package mx

import (
	"context"
	"errors"
	"net/http"
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
	//     errors and recovered panics of [ComponentHTTPHandler],
	//     [ComponentFuncHandler] and the reflected form handlers, and the 500
	//     a reflected form answers when its load function fails;
	//   - the in-band event of [SSEResponse.SendError];
	//   - an error a reflected form shows as [FormErrorText]: one onSubmit
	//     returned, an error no input of the form shows, and a context option
	//     list that failed while a POST was parsed;
	//   - the errors of a form:"sensitive" field, which read "not a valid
	//     value". Their text may quote the value the form keeps off the page,
	//     so a hook that logs them must treat it as secret as the value.
	//
	// It gets each such error once, whether or not
	// [RevealInternalServerErrors] is set, but not an error the user sees: a
	// [UserMessageError] or an error shown at its field. Nor does it get a
	// context.Canceled or context.DeadlineExceeded, because the client that
	// would have been told has disconnected. A recovered panic arrives as an
	// error carrying the debug stack of the panic.
	//
	// ctx is the request's context without its cancellation: it carries the
	// request's values, such as a request-scoped logger, and a client that
	// went away does not cut a report short. ReportError is called from the
	// goroutines serving requests, so it must be safe for concurrent use; set
	// it once at start-up. A decider or handler of the application hiding an
	// error reports it through [ReportHiddenError].
	ReportError func(ctx context.Context, err error)
)

// UserMessageError is an error whose message is meant for the user, such as
// a rule a submitted value breaks. Where go-mx hides the text of any other
// error, it shows the UserMessage instead: as a reflected form's message for
// an onSubmit error ([FormErrorText]), and for a form:"sensitive" field.
//
// go-mx recognizes the method set wherever [errors.As] finds it in an
// error's chain, also through errors.Join, so a domain package marks its
// user-fixable errors with a UserMessage method without importing go-mx.
// [FormMessage] is a ready-made one.
type UserMessageError interface {
	error

	// UserMessage returns the text shown to the user. It stands alone: the
	// Error text of the error, or of one wrapping it, may add internal
	// context the user must not see.
	UserMessage() string
}

// UserMessage returns the UserMessage of the first [UserMessageError] in the
// chain of err, as [errors.As] finds it, also through errors.Join, and
// whether there is one. It is the test go-mx applies wherever it decides
// whether the user may read an error.
func UserMessage(err error) (message string, ok bool) {
	if m, ok := errors.AsType[UserMessageError](err); ok {
		return m.UserMessage(), true
	}
	return "", false
}

// ReportHiddenError passes err, an error the client is shown a generic text
// for instead, to [ReportError] with ctx stripped of its cancellation. It
// does nothing when ReportError or err is nil, or when err is a
// context.Canceled or context.DeadlineExceeded. go-mx calls it wherever it
// hides an error; a decider or handler of the application that hides one
// calls it too.
func ReportHiddenError(ctx context.Context, err error) {
	if ReportError == nil || err == nil || isContextError(err) {
		return
	}
	ReportError(context.WithoutCancel(ctx), err)
}

// isContextError reports whether err is a context.Canceled or
// context.DeadlineExceeded, which go-mx takes for a client that has
// disconnected.
func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// hiddenErrorText returns the text the client is shown in place of err, an
// error not meant for it, and reports err ([ReportHiddenError]): generic, or
// the error's own text when [RevealInternalServerErrors] is set. ok is false
// for a context.Canceled or context.DeadlineExceeded, which is neither shown
// nor reported, because the client that would read it has disconnected.
//
// It is the policy of the 500s go-mx answers ([RespondNonContextError],
// respondLoadError), of [SSEResponse.SendError] (which sends the text as an
// in-band event, because by then the 200 is committed) and of
// [FormErrorText], so they cannot drift apart.
func hiddenErrorText(ctx context.Context, err error, generic string) (text string, ok bool) {
	if isContextError(err) {
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
// context.Canceled or context.DeadlineExceeded, in which case it does
// nothing because the client has disconnected. If
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
