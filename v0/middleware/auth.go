package middleware

import (
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

// Role constants
const (
	RoleAdmin = 1
	RoleSAV   = 2
	RoleDCP   = 3
	RoleGuest = 4
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

// AdminSAVMiddleware creates a middleware that only allows Admin and SAV roles
func AdminSAVMiddleware(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return RoleMiddleware(store, []int{RoleAdmin, RoleSAV})
}

// ViewOnlyMiddleware creates a middleware that allows all roles (everyone can view)
func ViewOnlyMiddleware(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return RoleMiddleware(store, []int{RoleAdmin, RoleSAV, RoleDCP, RoleGuest})
}

// AdminOnlyMiddleware creates a middleware that only allows Admin role
func AdminOnlyMiddleware(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return RoleMiddleware(store, []int{RoleAdmin})
}

// HasRole checks if the user has one of the specified roles
func HasRole(r *http.Request, store *sessions.CookieStore, roles []int) bool {
	session, err := store.Get(r, "session-name")
	if err != nil {
		return false
	}

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		return false
	}

	cargo, ok := session.Values["cargo"].(int)
	if !ok {
		return false
	}

	for _, role := range roles {
		if cargo == role {
			return true
		}
	}

	return false
}

// IsAdminOrSAV checks if the user is an Admin or SAV
func IsAdminOrSAV(r *http.Request, store *sessions.CookieStore) bool {
	return HasRole(r, store, []int{RoleAdmin, RoleSAV})
}

// IsAdmin checks if the user is an Admin
func IsAdmin(r *http.Request, store *sessions.CookieStore) bool {
	return HasRole(r, store, []int{RoleAdmin})
}
