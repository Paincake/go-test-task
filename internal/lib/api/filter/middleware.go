package filter

import (
	"context"
	"net/http"
)

const (
	FilterContextKey = "filter_options"
)

func Middleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		filterOptions := r.URL.Query().Get("filter")
		ctx := context.WithValue(r.Context(), FilterContextKey, filterOptions)
		r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
