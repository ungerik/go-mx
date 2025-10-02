package mx

import (
	"net/http"
)

var (
	RevealInternalServerErrors = true
)

func RespondError(w http.ResponseWriter, err error) {
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
