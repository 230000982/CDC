package middleware

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

// AuthMiddleware creates a middleware that checks if the user is authenticated
func AuthMiddleware(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check authentication
			session, err := store.Get(r, "session-name")
			if err != nil {
				http.Error(w, "Session error", http.StatusInternalServerError)
				log.Printf("Session error: %v", err)
				return
			}

			if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RoleMiddleware creates a middleware that checks if the user has the required role
func RoleMiddleware(store *sessions.CookieStore, allowedRoles []int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check authentication first
			session, err := store.Get(r, "session-name")
			if err != nil {
				http.Error(w, "Session error", http.StatusInternalServerError)
				log.Printf("Session error: %v", err)
				return
			}

			if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Check role
			cargo, ok := session.Values["cargo"].(int)
			if !ok {
				http.Error(w, "Role error", http.StatusInternalServerError)
				return
			}

			// Check if the user's role is in the allowed roles
			allowed := false
			for _, role := range allowedRoles {
				if cargo == role {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Não autorizado para esta operação", http.StatusForbidden)
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddlewareFunc creates a middleware function that checks if the user is authenticated
func AuthMiddlewareFunc(store *sessions.CookieStore, db *sql.DB) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check authentication
			session, err := store.Get(r, "session-name")
			if err != nil {
				http.Error(w, "Session error", http.StatusInternalServerError)
				log.Printf("Session error: %v", err)
				return
			}

			if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Call the next handler
			next(w, r)
		}
	}
}
