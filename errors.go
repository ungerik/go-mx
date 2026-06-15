package mx

import (
	"context"
	"errors"
	"net/http"
)

var (
	// RevealInternalServerErrors controls whether [RespondNonContextError]
	// includes the actual error message in a 500 response. It defaults to false,
	// which sends a generic "Internal Server Error"; set it to true (typically in
	// development) to expose the error text.
	RevealInternalServerErrors = false
)

// RespondNonContextError responds with an internal server error,
// unless the error is a context.Canceled or context.DeadlineExceeded,
// in which case it does nothing because the client has disconnected.
// If RevealInternalServerErrors is true, the error message is included
// in the response, otherwise a generic "Internal Server Error" is used.
func RespondNonContextError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return
	}
	errString := "Internal Server Error"
	if RevealInternalServerErrors {
		errString = err.Error()
	}
	http.Error(w, errString, http.StatusInternalServerError)
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
