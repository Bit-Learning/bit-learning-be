package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/lcaohoanq/bit-learning-be-v2/internal/platform/httpx"
)

type TokenParser interface {
	Parse(raw string) (uuid.UUID, error)
}

type userIDContextKey struct{}

func Authenticate(tokens TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "a Bearer token is required")
				return
			}
			id, err := tokens.Parse(parts[1])
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "token is invalid or expired")
				return
			}
			ctx := context.WithValue(r.Context(), userIDContextKey{}, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDContextKey{}).(uuid.UUID)
	return id, ok
}
