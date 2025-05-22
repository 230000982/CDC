package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"text/template"
	"time"

	"v0/config"
	"v0/database"
	"v0/models"
	"v0/services"

	"v0/middleware"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// ConcursoHandler handles concurso-related requests
type ConcursoHandler struct {
	db         *sql.DB
	store      *sessions.CookieStore
	cfg        *config.Config
	logService *services.LogService
}

// NewConcursoHandler creates a new ConcursoHandler
func NewConcursoHandler(db *sql.DB, store *sessions.CookieStore, cfg *config.Config) *ConcursoHandler {
	return &ConcursoHandler{
		db:         db,
		store:      store,
		cfg:        cfg,
		logService: services.NewLogService(db),
	}
}

// List handles the concursos list page
func (h *ConcursoHandler) List(w http.ResponseWriter, r *http.Request) {
	// Get search parameter from URL
	entidade := r.URL.Query().Get("entidade")

	// Get concursos from database
	concursos, err := models.GetConcursos(h.db, entidade)
	if err != nil {
		log.Printf("Error fetching concursos: %v", err)
		http.Error(w, "Erro ao buscar concursos", http.StatusInternalServerError)
		return
	}

	// Format prices for display
	type ConcursoDisplay struct {
		models.Concurso
		PrecoFormatted string
	}

	var concursosDisplay []ConcursoDisplay
	for _, c := range concursos {
		concursosDisplay = append(concursosDisplay, ConcursoDisplay{
			Concurso:       c,
			PrecoFormatted: fmt.Sprintf("%.2f€", c.Preco),
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

	// Check if user can edit (Admin or SAV)
	canEdit := cargoID == middleware.RoleAdmin || cargoID == middleware.RoleSAV

	// Render template
	tmpl := template.Must(template.ParseFiles("templates/layout/base.html", "templates/concursos/list.html"))
	data := struct {
		Title     string
		User      interface{}
		Concursos []ConcursoDisplay
		CanEdit   bool
	}{
		Title:     "Lista de Concursos",
		User:      user,
		Concursos: concursosDisplay,
		CanEdit:   canEdit,
	}
	tmpl.Execute(w, data)
}

// Edit handles the edit concurso page
func (h *ConcursoHandler) Edit(w http.ResponseWriter, r *http.Request) {
	// Get concurso ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID do concurso inválido", http.StatusBadRequest)
		return
	}

	// Get concurso from database
	concurso, err := models.GetConcursoByID(h.db, id)
	if err != nil {
		log.Printf("Error fetching concurso: %v", err)
		http.Error(w, "Erro ao buscar concurso", http.StatusInternalServerError)
		return
	}

	// Get related data
	tipos, err := database.GetTipos(h.db)
	if err != nil {
		log.Printf("Error fetching tipos: %v", err)
		http.Error(w, "Erro ao buscar tipos", http.StatusInternalServerError)
		return
	}

	plataformas, err := database.GetPlataformas(h.db)
	if err != nil {
		log.Printf("Error fetching plataformas: %v", err)
		http.Error(w, "Erro ao buscar plataformas", http.StatusInternalServerError)
		return
	}

	estados, err := database.GetEstados(h.db)
	if err != nil {
		log.Printf("Error fetching estados: %v", err)
		http.Error(w, "Erro ao buscar estados", http.StatusInternalServerError)
		return
	}

	resultado, err := database.GetAdjudicatarios(h.db)
	if err != nil {
		log.Printf("Error fetching adjudicatarios: %v", err)
		http.Error(w, "Erro ao buscar adjudicatários", http.StatusInternalServerError)
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

	// Prepare data for template
	data := struct {
		Title    string
		User     interface{}
		Concurso *models.Concurso
		Tipos    []struct {
			ID        int
			Descricao string
		}
		Plataformas []struct {
			ID        int
			Descricao string
		}
		Estados []struct {
			ID        int
			Descricao string
		}
		Resultado []struct {
			ID        int
			Descricao string
		}
	}{
		Title:       "Editar Concurso",
		User:        user,
		Concurso:    concurso,
		Tipos:       tipos,
		Plataformas: plataformas,
		Estados:     estados,
		Resultado:   resultado,
	}

	// Render template
	tmpl := template.Must(template.ParseFiles("templates/layout/base.html", "templates/concursos/edit.html"))
	tmpl.Execute(w, data)
}

// Update handles the update concurso form submission
func (h *ConcursoHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Get concurso ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID do concurso inválido", http.StatusBadRequest)
		return
	}

	// Get user ID from session
	session, _ := h.store.Get(r, "session-name")
	userID, _ := session.Values["user_id"].(int)

	// Get original concurso for logging
	oldConcurso, err := models.GetConcursoByID(h.db, id)
	if err != nil {
		log.Printf("Error fetching original concurso: %v", err)
		http.Error(w, "Erro ao buscar concurso original", http.StatusInternalServerError)
		return
	}

	// Process form
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			log.Printf("Form parse error: %v", err)
			http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
			return
		}

		// Get form values
		preco, err := strconv.ParseFloat(r.FormValue("preco"), 64)
		if err != nil {
			http.Error(w, "Preço inválido", http.StatusBadRequest)
			return
		}

		tipoID, err := strconv.Atoi(r.FormValue("tipo_id"))
		if err != nil {
			http.Error(w, "Tipo inválido", http.StatusBadRequest)
			return
		}

		plataformaID, err := strconv.Atoi(r.FormValue("plataforma_id"))
		if err != nil {
			http.Error(w, "Plataforma inválida", http.StatusBadRequest)
			return
		}

		estadoID, err := strconv.Atoi(r.FormValue("estado_id"))
		if err != nil {
			http.Error(w, "Estado inválido", http.StatusBadRequest)
			return
		}

		// Create concurso object
		concurso := &models.Concurso{
			ID:            oldConcurso.ID,
			Preco:         preco,
			Referencia:    r.FormValue("referencia"),
			Entidade:      r.FormValue("entidade"),
			ReferenciaBC:  r.FormValue("referencia_bc"),
			DiaErro:       models.ParseNullString(r.FormValue("dia_erro")),
			HoraErro:      models.ParseNullString(r.FormValue("hora_erro")),
			DiaProposta:   models.ParseNullString(r.FormValue("dia_proposta")),
			HoraProposta:  models.ParseNullString(r.FormValue("hora_proposta")),
			DiaAudiencia:  models.ParseNullString(r.FormValue("dia_audiencia")),
			HoraAudiencia: models.ParseNullString(r.FormValue("hora_audiencia")),
			Preliminar:    r.FormValue("preliminar") == "on",
			Final:         r.FormValue("final") == "on",
			Recurso:       r.FormValue("recurso") == "on",
			Impugnacao:    r.FormValue("impugnacao") == "on",
			TipoID:        tipoID,
			PlataformaID:  plataformaID,
			EstadoID:      estadoID,
		}

		// Update concurso in database
		if err := models.UpdateConcurso(h.db, id, concurso); err != nil {
			log.Printf("Error updating concurso: %v", err)
			http.Error(w, "Erro ao atualizar concurso", http.StatusInternalServerError)
			return
		}

		// Log the update action
		if err := h.logService.LogUpdate(userID, "concurso", oldConcurso, concurso); err != nil {
			log.Printf("Error logging update: %v", err)
		}

		// Get tipo description for email
		var tipoDesc string
		rows, err := h.db.Query("SELECT descricao FROM tipo WHERE id_tipo = ?", tipoID)
		if err == nil {
			defer rows.Close()
			if rows.Next() {
				rows.Scan(&tipoDesc)
			}
		}

		// Get estado description for email
		var estadoDesc string
		rows, err = h.db.Query("SELECT descricao FROM estado WHERE id_estado = ?", estadoID)
		if err == nil {
			defer rows.Close()
			if rows.Next() {
				rows.Scan(&estadoDesc)
			}
		}

		// Send update email
		emailService := services.NewEmailService(h.cfg.Email)
		err = emailService.SendUpdateEmail(
			concurso.Referencia,
			concurso.Entidade,
			tipoDesc,
			estadoDesc,
			concurso.DiaErro.NullString.String,
			concurso.HoraErro.NullString.String,
			concurso.DiaProposta.NullString.String,
			concurso.HoraProposta.NullString.String,
			concurso.DiaAudiencia.NullString.String,
			concurso.HoraAudiencia.NullString.String,
			concurso.Preliminar,
			concurso.Final,
			concurso.Recurso,
			concurso.Impugnacao,
			h.db,
		)
		if err != nil {
			log.Printf("Error sending email: %v", err)
		}

		// Redirect to concursos list
		http.Redirect(w, r, "/concursos", http.StatusSeeOther)
	}
}

// Create handles the create concurso page
func (h *ConcursoHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get related data
	tipos, err := database.GetTipos(h.db)
	if err != nil {
		log.Printf("Error fetching tipos: %v", err)
		http.Error(w, "Erro ao buscar tipos", http.StatusInternalServerError)
		return
	}

	plataformas, err := database.GetPlataformas(h.db)
	if err != nil {
		log.Printf("Error fetching plataformas: %v", err)
		http.Error(w, "Erro ao buscar plataformas", http.StatusInternalServerError)
		return
	}

	estados, err := database.GetEstados(h.db)
	if err != nil {
		log.Printf("Error fetching estados: %v", err)
		http.Error(w, "Erro ao buscar estados", http.StatusInternalServerError)
		return
	}

	resultado, err := database.GetAdjudicatarios(h.db)
	if err != nil {
		log.Printf("Error fetching adjudicatarios: %v", err)
		http.Error(w, "Erro ao buscar adjudicatários", http.StatusInternalServerError)
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

	// Prepare data for template
	data := struct {
		Title string
		User  interface{}
		Tipos []struct {
			ID        int
			Descricao string
		}
		Plataformas []struct {
			ID        int
			Descricao string
		}
		Estados []struct {
			ID        int
			Descricao string
		}
		Resultado []struct {
			ID        int
			Descricao string
		}
	}{
		Title:       "Criar Concurso",
		User:        user,
		Tipos:       tipos,
		Plataformas: plataformas,
		Estados:     estados,
		Resultado:   resultado,
	}

	// Render template
	tmpl := template.Must(template.ParseFiles("templates/layout/base.html", "templates/concursos/create.html"))
	tmpl.Execute(w, data)
}

// Save handles the save concurso form submission
func (h *ConcursoHandler) Save(w http.ResponseWriter, r *http.Request) {
	// Get user ID from session
	session, _ := h.store.Get(r, "session-name")
	userID, _ := session.Values["user_id"].(int)

	// Process form
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			log.Printf("Form parse error: %v", err)
			http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
			return
		}

		// Get form values
		precoStr := r.FormValue("preco")
		tipoIDStr := r.FormValue("tipo_id")
		plataformaIDStr := r.FormValue("plataforma_id")
		estadoIDStr := r.FormValue("estado_id")

		// Convert strings to their respective types with default values
		var preco float64
		var err error
		if precoStr != "" {
			preco, err = strconv.ParseFloat(precoStr, 64)
			if err != nil {
				http.Error(w, "Preço inválido", http.StatusBadRequest)
				return
			}
		}

		var tipoID int
		if tipoIDStr != "" {
			tipoID, err = strconv.Atoi(tipoIDStr)
			if err != nil {
				http.Error(w, "Tipo inválido", http.StatusBadRequest)
				return
			}
		}

		var plataformaID int
		if plataformaIDStr != "" {
			plataformaID, err = strconv.Atoi(plataformaIDStr)
			if err != nil {
				http.Error(w, "Plataforma inválida", http.StatusBadRequest)
				return
			}
		}

		var estadoID int
		if estadoIDStr != "" {
			estadoID, err = strconv.Atoi(estadoIDStr)
			if err != nil {
				http.Error(w, "Estado inválido", http.StatusBadRequest)
				return
			}
		}

		// Create concurso object
		concurso := &models.Concurso{
			Preco:         preco,
			Referencia:    r.FormValue("referencia"),
			Entidade:      r.FormValue("entidade"),
			ReferenciaBC:  r.FormValue("referencia_bc"),
			DiaErro:       models.ParseNullString(r.FormValue("dia_erro")),
			HoraErro:      models.ParseNullString(r.FormValue("hora_erro")),
			DiaProposta:   models.ParseNullString(r.FormValue("dia_proposta")),
			HoraProposta:  models.ParseNullString(r.FormValue("hora_proposta")),
			DiaAudiencia:  models.ParseNullString(r.FormValue("dia_audiencia")),
			HoraAudiencia: models.ParseNullString(r.FormValue("hora_audiencia")),
			Preliminar:    r.FormValue("preliminar") == "on",
			Final:         r.FormValue("final") == "on",
			Recurso:       r.FormValue("recurso") == "on",
			Impugnacao:    r.FormValue("impugnacao") == "on",
			TipoID:        tipoID,
			PlataformaID:  plataformaID,
			EstadoID:      estadoID,
		}

		// Create concurso in database
		if err := models.CreateConcurso(h.db, concurso); err != nil {
			log.Printf("Error creating concurso: %v", err)
			http.Error(w, "Erro ao criar concurso", http.StatusInternalServerError)
			return
		}

		// Log the create action
		if err := h.logService.LogCreate(userID, "concurso", concurso); err != nil {
			log.Printf("Error logging create: %v", err)
		}

		// Get tipo description for email
		var tipoDesc string
		rows, err := h.db.Query("SELECT descricao FROM tipo WHERE id_tipo = ?", tipoID)
		if err == nil {
			defer rows.Close()
			if rows.Next() {
				rows.Scan(&tipoDesc)
			}
		}

		// Get estado description for email
		var estadoDesc string
		rows, err = h.db.Query("SELECT descricao FROM estado WHERE id_estado = ?", estadoID)
		if err == nil {
			defer rows.Close()
			if rows.Next() {
				rows.Scan(&estadoDesc)
			}
		}

		// Send update email
		emailService := services.NewEmailService(h.cfg.Email)
		err = emailService.SendUpdateEmail(
			concurso.Referencia,
			concurso.Entidade,
			tipoDesc,
			estadoDesc,
			concurso.DiaErro.NullString.String,
			concurso.HoraErro.NullString.String,
			concurso.DiaProposta.NullString.String,
			concurso.HoraProposta.NullString.String,
			concurso.DiaAudiencia.NullString.String,
			concurso.HoraAudiencia.NullString.String,
			concurso.Preliminar,
			concurso.Final,
			concurso.Recurso,
			concurso.Impugnacao,
			h.db,
		)
		if err != nil {
			log.Printf("Error sending email: %v", err)
		}

		// Redirect to concursos list
		http.Redirect(w, r, "/concursos", http.StatusSeeOther)
	}
}

// Delete handles the delete concurso request
func (h *ConcursoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Get concurso ID from URL
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "ID do concurso inválido", http.StatusBadRequest)
		return
	}

	// Get user ID from session
	session, _ := h.store.Get(r, "session-name")
	userID, _ := session.Values["user_id"].(int)

	// Get original concurso for logging
	oldConcurso, err := models.GetConcursoByID(h.db, id)
	if err != nil {
		log.Printf("Error fetching concurso for deletion: %v", err)
		http.Error(w, "Erro ao buscar concurso", http.StatusInternalServerError)
		return
	}

	// Delete concurso
	if err := models.DeleteConcurso(h.db, id); err != nil {
		log.Printf("Error deleting concurso: %v", err)
		http.Error(w, "Erro ao excluir concurso", http.StatusInternalServerError)
		return
	}

	// Log the delete action
	if err := h.logService.LogDelete(userID, "concurso", oldConcurso); err != nil {
		log.Printf("Error logging delete: %v", err)
	}

	// Redirect to concursos list
	http.Redirect(w, r, "/concursos", http.StatusSeeOther)
}

// ListOrdered handles the ordered concursos list page
func (h *ConcursoHandler) ListOrdered(w http.ResponseWriter, r *http.Request) {
	// Get current date and time
	now := time.Now()

	// Get future concursos from database
	items, err := models.GetFutureConcursos(h.db)
	if err != nil {
		log.Printf("Error fetching future concursos: %v", err)
		http.Error(w, "Erro ao buscar concursos futuros", http.StatusInternalServerError)
		return
	}

	// Sort items by date and time
	sort.Slice(items, func(i, j int) bool {
		// Compare first by date
		if items[i].Data != items[j].Data {
			return items[i].Data < items[j].Data
		}
		// If date is equal, compare by time
		return items[i].Hora < items[j].Hora
	})

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

	// Register daysUntil function for use in template
	funcMap := template.FuncMap{
		"daysUntil": func(date string) int {
			targetDate, err := time.Parse("2006-01-02", date)
			if err != nil {
				return -1
			}
			duration := targetDate.Sub(now)
			return int(duration.Hours() / 24)
		},
	}

	// Load template with custom functions
	tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles("templates/layout/base.html", "templates/concursos/ordered.html")
	if err != nil {
		log.Printf("Error loading template: %v", err)
		http.Error(w, "Erro ao carregar template", http.StatusInternalServerError)
		return
	}

	// Data for template
	data := struct {
		Title       string
		User        interface{}
		Items       []models.ConcursoItem
		CurrentTime string
	}{
		Title:       "Concursos Futuros",
		User:        user,
		Items:       items,
		CurrentTime: now.Format("2006-01-02 15:04:05"),
	}

	// Execute template
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
		return
	}
}
