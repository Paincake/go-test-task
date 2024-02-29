package sort

import (
	"context"
	"net/http"
	"strings"
)

const (
	SortOptionsContextKey = "sort_options"
)

func Middleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		sortBy := r.URL.Query().Get("sort")
		sortOrder := r.URL.Query().Get("order")

		options := Options{
			Field: sortBy,
			Order: strings.ToUpper(sortOrder),
		}

		ctx := context.WithValue(r.Context(), SortOptionsContextKey, options)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

type Options struct {
	Field string
	Order string
}
