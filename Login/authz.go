package main

import (
	"context"
	"net/http"
)

// Context key for authenticated user
type authUserKey struct{}

// Authentication Middleware
func (a *App) requireAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check existing session
		u, err := a.user(r)

		if err != nil {
			fail(w, http.StatusUnauthorized, "Please log in.")
			return
		}

		// Store verified user in request context
		ctx := context.WithValue(r.Context(), authUserKey{}, u)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Authorization Middleware
func requireRoles(allowedRoles ...string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Get authenticated user
			u, ok := r.Context().Value(authUserKey{}).(User)

			if !ok {
				fail(w, http.StatusUnauthorized, "Authentication required.")
				return
			}

			// Check user role
			for _, role := range allowedRoles {

				if u.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			// No permission
			fail(w, http.StatusForbidden, "You do not have permission.")
		})
	}
}
