package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dutakey/shellgate/internal/store"
	"github.com/dutakey/shellgate/internal/types"
)

type contextKey string

const ContextKeyAPIKey contextKey = "api_key"

func Auth(ks *store.KeyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractBearerToken(r)
			if key == "" {
				writeError(w, http.StatusUnauthorized, "missing Authorization header", "auth_error")
				return
			}

			if !ks.Validate(key) {
				writeError(w, http.StatusUnauthorized, "invalid API key", "invalid_api_key")
				return
			}

			go ks.Touch(key)

			ctx := context.WithValue(r.Context(), ContextKeyAPIKey, key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func writeError(w http.ResponseWriter, status int, msg, errType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(types.ErrorResponse{
		Error: types.APIError{
			Message: msg,
			Type:    errType,
		},
	})
}
