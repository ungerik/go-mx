package mx

import (
	"context"
	"net/http"
)

// Middleware returns an HTTP middleware that installs d into every
// request's context. Wrap a handler tree once and every
// [ReflectFormHandler] beneath it will use d for field rendering /
// parsing / validation.
//
// Multiple Middleware wraps may be nested to pick different deciders
// for different subtrees (an admin section that uses shadcn, a public
// section that uses html). The innermost wrap wins because the
// outermost middleware runs first and overwrites whatever an inner
// middleware just wrote — except that the inner middleware writes
// the context value AFTER the outer one for the request's pass into
// the handler. (In practice the inner wrap is the per-route choice
// and the outer is the app-wide default; the per-route choice should
// reach the handler.)
func Middleware(d FieldDecider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxKeyDecider{}, d)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ContextWithDecider returns a derived context that carries d. Useful
// for tests that build a context directly instead of going through an
// HTTP server, and for callers that compose deciders manually.
func ContextWithDecider(ctx context.Context, d FieldDecider) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxKeyDecider{}, d)
}
