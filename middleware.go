package authtokens

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

var ErrMissingToken = errors.New("authorization token not provided")

type contextKey struct{}

// Middleware returns HTTP middleware that extracts a Bearer token from the
// Authorization header, validates it, and injects the resulting Claims
// into the request context. On failure it responds with 401 JSON.
func Middleware(v Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				writeError(w, ErrMissingToken)
				return
			}

			raw := strings.TrimPrefix(header, "Bearer ")
			claims, err := v.Validate(raw)
			if err != nil {
				writeError(w, err)
				return
			}

			ctx := context.WithValue(r.Context(), contextKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext retrieves the Claims injected by the Middleware.
func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(contextKey{}).(Claims)
	return claims, ok
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}
