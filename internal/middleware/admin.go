package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dutakey/shellgate/internal/types"
)

func AdminAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				token = r.Header.Get("X-Admin-Secret")
			}

			if token == "" || !strings.EqualFold(token, secret) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(types.ErrorResponse{
					Error: types.APIError{
						Message: "invalid admin secret",
						Type:    "auth_error",
					},
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
