package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"v0/models"
	"v0/services"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// UserHandler handles user-related requests
type UserHandler struct {
	db         *sql.DB
	store      *sessions.CookieStore
	logService *services.LogService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(db *sql.DB, store *sessions.CookieStore) *UserHandler {
	return &UserHandler{
		db:         db,
		store:      store,
		logService: services.NewLogService(db),
	}
}

// List handles the users list page
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	// Get users from database
	users, err := models.GetAllUsers(h.db)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		http.Error(w, "Erro ao buscar usuários", http.StatusInternalServerError)
		return
	}

	// Get cargos for display
	cargos, err := models.GetAllCargos(h.db)
	if err != nil {
		log.Printf("Error fetching cargos: %v", err)
		http.Error(w, "Erro ao buscar cargos", http.StatusInternalServerError)
		return
	}

	// Create a map of cargo IDs to descriptions
	cargoMap := make(map[int]string)
	for _, cargo := range cargos {
		cargoMap[cargo.ID] = cargo.Descricao
	}

	// Add cargo description to each user
	type UserDisplay struct {
		models.User
		CargoDesc string
	}

	var usersDisplay []UserDisplay
	for _, u := range users {
		cargoDesc, ok := cargoMap[u.CargoID]
		if !ok {
			cargoDesc = "Desconhecido"
		}

		usersDisplay = append(usersDisplay, UserDisplay{
			User:      u,
			CargoDesc: cargoDesc,
		})
	}

	// Get user info from session
	session, _ := h.store.Get(r, "session-name")
	userID, _ := session.Values["user_id"].(int)
	cargoID, _ := session.Values["cargo"].(int)

	user := struct {
		ID      int
		Nome    string
		CargoID int
	}{
		ID:      userID,
		Nome:    "Usuário", // Default name
		CargoID: cargoID,
	}

	// Get actual user name if possible
	if userID > 0 {
		var nome string
		err := h.db.QueryRow("SELECT nome FROM user WHERE id_user = ?", userID).Scan(&nome)
		if err == nil {
			user.Nome = nome
		}
	}

	// Render template
	tmpl := template.Must(template.ParseFiles("templates/layout/base.html", "templates/admin/users/list.html"))
	data := struct {
		Title string
		User  interface{}
		Users []UserDisplay
	}{
		Title: "Gerenciar Usuários",
		User:  user,
		Users: usersDisplay,
	}
	tmpl.Execute(w, data)
}

// Create handles the create user page
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get cargos for the form
	cargos, err := models.GetAllCargos(h.db)
	if err != nil {
		log.Printf("Error fetching cargos: %v", err)
		http.Error(w, "Erro ao buscar cargos", http.StatusInternalServerError)
		return
	}

	// Get user info from session
	session, _ := h.store.Get(r, "session-name")
	userID, _ := session.Values["user_id"].(int)
	cargoID, _ := session.Values["cargo"].(int)

	user := struct {
		ID      int
		Nome    string
		CargoID int
	}{
		ID:      userID,
		Nome:    "Usuário", // Default name
		CargoID: cargoID,
	}

	// Get actual user name if possible
	if userID > 0 {
		var nome string
		err := h.db.QueryRow("SELECT nome FROM user WHERE id_user = ?", userID).Scan(&nome)
		if err == nil {
			user.Nome = nome
		}
	}

	// Render template
	tmpl := template.Must(template.ParseFiles("templates/layout/base.html", "templates/admin/users/create.html"))
	data := struct {
		Title  string
		User   interface{}
		Cargos []models.Cargo
	}{
		Title:  "Criar Usuário",
		User:   user,
		Cargos: cargos,
	}
	tmpl.Execute(w, data)
}

// Save handles the save user form submission
func (h *UserHandler) Save(w http.ResponseWriter, r *http.Request) {
	// Get user ID from session for logging
	session, _ := h.store.Get(r, "session-name")
	adminID, _ := session.Values["user_id"].(int)

	// Process form
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			log.Printf("Form parse error: %v", err)
			http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
			return
		}

		nome := r.FormValue("nome")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")
		cargoIDStr := r.FormValue("cargo_id")

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

		// Convert cargo_id to int
		cargoID, err := strconv.Atoi(cargoIDStr)
		if err != nil {
			http.Error(w, "Cargo inválido", http.StatusBadRequest)
			return
		}

		// Create user
		if err := models.CreateUser(h.db, nome, email, password, cargoID); err != nil {
			log.Printf("User creation error: %v", err)
			http.Error(w, "Erro ao criar usuário", http.StatusInternalServerError)
			return
		}

		// Get the newly created user's ID for logging
		var newUserID int
		err = h.db.QueryRow("SELECT id_user FROM user WHERE email = ?", email).Scan(&newUserID)
		if err == nil {
			// Log the user creation
			h.logService.LogCreate(adminID, "user", map[string]interface{}{
				"user_id":  newUserID,
				"nome":     nome,
				"email":    email,
				"cargo_id": cargoID,
			})
		}

		// Redirect to users list
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

// Edit handles the edit user page
func (h *UserHandler) Edit(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID do usuário inválido", http.StatusBadRequest)
		return
	}

	// Convert ID to int
	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "ID do usuário inválido", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := models.GetUserByID(h.db, userID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Error(w, "Erro ao buscar usuário", http.StatusInternalServerError)
		return
	}

	// Get cargos for the form
	cargos, err := models.GetAllCargos(h.db)
	if err != nil {
		log.Printf("Error fetching cargos: %v", err)
		http.Error(w, "Erro ao buscar cargos", http.StatusInternalServerError)
		return
	}

	// Get admin info from session
	session, _ := h.store.Get(r, "session-name")
	adminID, _ := session.Values["user_id"].(int)
	adminCargoID, _ := session.Values["cargo"].(int)

	admin := struct {
		ID      int
		Nome    string
		CargoID int
	}{
		ID:      adminID,
		Nome:    "Usuário", // Default name
		CargoID: adminCargoID,
	}

	// Get actual admin name if possible
	if adminID > 0 {
		var nome string
		err := h.db.QueryRow("SELECT nome FROM user WHERE id_user = ?", adminID).Scan(&nome)
		if err == nil {
			admin.Nome = nome
		}
	}

	// Render template
	tmpl := template.Must(template.ParseFiles("templates/layout/base.html", "templates/admin/users/edit.html"))
	data := struct {
		Title  string
		User   interface{}  // This is the admin
		Target *models.User // This is the user being edited
		Cargos []models.Cargo
	}{
		Title:  "Editar Usuário",
		User:   admin,
		Target: user,
		Cargos: cargos,
	}
	tmpl.Execute(w, data)
}

// Update handles the update user form submission
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID do usuário inválido", http.StatusBadRequest)
		return
	}

	// Convert ID to int
	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "ID do usuário inválido", http.StatusBadRequest)
		return
	}

	// Get admin ID from session for logging
	session, _ := h.store.Get(r, "session-name")
	adminID, _ := session.Values["user_id"].(int)

	// Get original user for logging
	oldUser, err := models.GetUserByID(h.db, userID)
	if err != nil {
		log.Printf("Error fetching original user: %v", err)
		http.Error(w, "Erro ao buscar usuário original", http.StatusInternalServerError)
		return
	}

	// Process form
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			log.Printf("Form parse error: %v", err)
			http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
			return
		}

		nome := r.FormValue("nome")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")
		cargoIDStr := r.FormValue("cargo_id")

		// Check if email changed and if it already exists
		if email != oldUser.Email {
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
		}

		// Convert cargo_id to int
		cargoID, err := strconv.Atoi(cargoIDStr)
		if err != nil {
			http.Error(w, "Cargo inválido", http.StatusBadRequest)
			return
		}

		// Create updated user object
		updatedUser := &models.User{
			ID:      userID,
			Nome:    nome,
			Email:   email,
			CargoID: cargoID,
		}

		// Update user in database
		if err := models.UpdateUser(h.db, updatedUser); err != nil {
			log.Printf("User update error: %v", err)
			http.Error(w, "Erro ao atualizar usuário", http.StatusInternalServerError)
			return
		}

		// Log the update action
		h.logService.LogUpdate(adminID, "user", oldUser, updatedUser)

		// Update password if provided
		if password != "" && confirmPassword != "" {
			if password != confirmPassword {
				http.Error(w, "As senhas não coincidem", http.StatusBadRequest)
				return
			}

			if err := models.UpdatePassword(h.db, userID, password); err != nil {
				log.Printf("Password update error: %v", err)
				http.Error(w, "Erro ao atualizar senha", http.StatusInternalServerError)
				return
			}

			// Log the password update
			h.logService.LogAction(adminID, "user", "password_update", nil, map[string]interface{}{
				"user_id": userID,
			})
		}

		// Redirect to users list
		http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
	}
}

// Delete handles the delete user request
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID do usuário inválido", http.StatusBadRequest)
		return
	}

	// Convert ID to int
	userID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "ID do usuário inválido", http.StatusBadRequest)
		return
	}

	// Get admin ID from session for logging
	session, _ := h.store.Get(r, "session-name")
	adminID, _ := session.Values["user_id"].(int)

	// Prevent self-deletion
	if userID == adminID {
		http.Error(w, "Não é possível excluir o próprio usuário", http.StatusBadRequest)
		return
	}

	// Get original user for logging
	oldUser, err := models.GetUserByID(h.db, userID)
	if err != nil {
		log.Printf("Error fetching user for deletion: %v", err)
		http.Error(w, "Erro ao buscar usuário", http.StatusInternalServerError)
		return
	}

	// Delete user
	if err := models.DeleteUser(h.db, userID); err != nil {
		log.Printf("User deletion error: %v", err)
		http.Error(w, "Erro ao excluir usuário", http.StatusInternalServerError)
		return
	}

	// Log the delete action
	h.logService.LogDelete(adminID, "user", oldUser)

	// Redirect to users list
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
