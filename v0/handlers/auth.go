package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"text/template"

	"v0/models"
	"v0/services"

	"github.com/gorilla/sessions"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db         *sql.DB
	store      *sessions.CookieStore
	logService *services.LogService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(db *sql.DB, store *sessions.CookieStore) *AuthHandler {
	return &AuthHandler{
		db:         db,
		store:      store,
		logService: services.NewLogService(db),
	}
}

// IndexHandler handles the index page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	data := struct {
		Title string
	}{
		Title: "Controlo de Concursos",
	}
	tmpl.Execute(w, data)
}

// Login handles the login page and login form submission
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Authenticate user
		userID, cargoID, err := models.Authenticate(h.db, email, password)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Email não encontrado", http.StatusUnauthorized)
			} else {
				log.Printf("Login error: %v", err)
				http.Error(w, "Erro ao fazer login", http.StatusInternalServerError)
			}
			return
		}

		// Log the login action
		h.logService.LogAction(userID, "user", "login", nil, map[string]interface{}{
			"user_id": userID,
			"email":   email,
		})

		// Create session
		session, _ := h.store.Get(r, "session-name")
		session.Values["authenticated"] = true
		session.Values["user_id"] = userID
		session.Values["cargo"] = cargoID
		if err := session.Save(r, w); err != nil {
			log.Printf("Session save error: %v", err)
			http.Error(w, "Erro ao salvar sessão", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/concursos", http.StatusSeeOther)
	} else {
		// Use a simple login page without the base template
		tmpl := template.Must(template.ParseFiles("templates/auth/login.html"))
		data := struct {
			Title string
		}{
			Title: "Login",
		}
		tmpl.Execute(w, data)
	}
}

// Register handles the registration form submission
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")

		// Basic validations
		if password != confirmPassword {
			http.Error(w, "As senhas não coincidem", http.StatusBadRequest)
			return
		}

		// Check if email already exists
		exists, err := models.EmailExists(h.db, email)
		if err != nil {
			log.Printf("Email check error: %v", err)
			http.Error(w, "Erro ao verificar email", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "Email já em uso", http.StatusBadRequest)
			return
		}

		// Create user (cargo_id 4 = Guest by default)
		if err := models.CreateUser(h.db, "user", email, password, 4); err != nil {
			log.Printf("User creation error: %v", err)
			http.Error(w, "Erro ao criar usuário", http.StatusInternalServerError)
			return
		}

		// Get the newly created user's ID for logging
		var userID int
		err = h.db.QueryRow("SELECT id_user FROM user WHERE email = ?", email).Scan(&userID)
		if err == nil {
			// Log the registration action
			h.logService.LogCreate(userID, "user", map[string]interface{}{
				"user_id":  userID,
				"email":    email,
				"cargo_id": 4,
			})
		}

		// Redirect to login after successful registration
		http.Redirect(w, r, "/login?registered=true", http.StatusSeeOther)
	} else {
		// If not POST, show the template normally
		tmpl := template.Must(template.ParseFiles("templates/layout/base.html", "templates/auth/login.html"))
		data := struct {
			Title string
		}{
			Title: "Registro",
		}
		tmpl.Execute(w, data)
	}
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "session-name")

	// Get user ID before clearing session
	userID, _ := session.Values["user_id"].(int)

	// Log the logout action if we have a valid user ID
	if userID > 0 {
		h.logService.LogAction(userID, "user", "logout", nil, map[string]interface{}{
			"user_id": userID,
		})
	}

	session.Values["authenticated"] = false
	session.Values["user_id"] = nil
	session.Values["cargo"] = nil
	session.Options.MaxAge = -1 // Delete the cookie
	if err := session.Save(r, w); err != nil {
		log.Printf("Session save error: %v", err)
		http.Error(w, "Erro ao fazer logout", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
