package html

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	"github.com/ungerik/go-mx"
)

func Serve(addr string, component mx.Component) error {
	return http.ListenAndServe(addr, http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		buf := bytes.NewBuffer(nil)
		writer := mx.NewCheckedWriter(buf).WithIndent("", "  ")
		err := component.Render(request.Context(), writer)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				http.Error(response, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		response.Header().Set("Content-Type", mx.ContentTypeHTML)
		response.Write(buf.Bytes())
	}))
}
