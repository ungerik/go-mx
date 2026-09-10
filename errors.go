package mx

import (
	"context"
	"errors"
	"net/http"
)

var (
	// RevealInternalServerErrors controls whether [RespondNonContextError] and
	// [SSEResponse.SendError] include the actual error message in what they
	// report. It defaults to false, which sends a generic "Internal Server
	// Error"; set it to true (typically in development) to expose the error text.
	RevealInternalServerErrors = false
)

// nonContextErrorMessage returns the message to report for err and whether it
// should be reported at all. A context.Canceled or context.DeadlineExceeded is
// not reported because the client that would read it has disconnected.
//
// It is the one place the [RevealInternalServerErrors] policy is applied, shared
// by [RespondNonContextError] (which reports it as a 500) and
// [SSEResponse.SendError] (which reports it as an in-band event, because by then
// the 200 is committed).
func nonContextErrorMessage(err error) (message string, report bool) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return "", false
	}
	if RevealInternalServerErrors {
		return err.Error(), true
	}
	return "Internal Server Error", true
}

// RespondNonContextError responds with an internal server error,
// unless the error is a context.Canceled or context.DeadlineExceeded,
// in which case it does nothing because the client has disconnected.
// If RevealInternalServerErrors is true, the error message is included
// in the response, otherwise a generic "Internal Server Error" is used.
func RespondNonContextError(w http.ResponseWriter, err error) {
	message, report := nonContextErrorMessage(err)
	if !report {
		return
	}
	http.Error(w, message, http.StatusInternalServerError)
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
