package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/jaykay/vereinstool/backend/db/generated"
	"github.com/jaykay/vereinstool/backend/service"
)

type contextKey string

const userContextKey contextKey = "user"

// UserFromContext retrieves the authenticated user from request context.
func UserFromContext(ctx context.Context) *generated.User {
	user, _ := ctx.Value(userContextKey).(*generated.User)
	return user
}

// AuthRequired middleware validates the session cookie and loads the user.
func AuthRequired(auth *service.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				jsonError(w, "Nicht angemeldet", http.StatusUnauthorized)
				return
			}

			user, err := auth.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				jsonError(w, "Sitzung ungültig", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminRequired middleware checks that the user has admin role.
func AdminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil || user.Role != "admin" {
			jsonError(w, "Keine Berechtigung", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CORS middleware allows requests from the frontend origin.
func CORS(allowOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (origin == allowOrigin || strings.HasPrefix(origin, "http://localhost")) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
