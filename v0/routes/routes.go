package routes

import (
	"database/sql"
	"net/http"

	"v0/config"
	"v0/handlers"
	"v0/middleware"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(db *sql.DB, cfg *config.Config) http.Handler {
	// Create session store
	store := sessions.NewCookieStore([]byte(cfg.Session.Secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}

	// Create router
	router := mux.NewRouter()

	// Create handler instances
	authHandler := handlers.NewAuthHandler(db, store)
	concursoHandler := handlers.NewConcursoHandler(db, store, cfg)
	pdfHandler := handlers.NewPDFHandler(db)
	userHandler := handlers.NewUserHandler(db, store)

	// Public routes
	router.HandleFunc("/", handlers.IndexHandler).Methods("GET")
	router.HandleFunc("/login", authHandler.Login).Methods("GET", "POST")
	router.HandleFunc("/register", authHandler.Register).Methods("GET", "POST")
	router.HandleFunc("/logout", authHandler.Logout).Methods("GET")

	// Protected routes - View Only (All authenticated users)
	viewOnly := router.PathPrefix("/").Subrouter()
	viewOnly.Use(middleware.ViewOnlyMiddleware(store))

	viewOnly.HandleFunc("/concursos", concursoHandler.List).Methods("GET")
	viewOnly.HandleFunc("/concursos-ordenados", concursoHandler.ListOrdered).Methods("GET")
	viewOnly.HandleFunc("/download-pdf", pdfHandler.Download).Methods("GET")

	// Protected routes - Admin and SAV Only
	adminSAV := router.PathPrefix("/").Subrouter()
	adminSAV.Use(middleware.AdminSAVMiddleware(store))

	adminSAV.HandleFunc("/edit-concurso/{id}", concursoHandler.Edit).Methods("GET")
	adminSAV.HandleFunc("/update-concurso/{id}", concursoHandler.Update).Methods("POST")
	adminSAV.HandleFunc("/create-concurso", concursoHandler.Create).Methods("GET")
	adminSAV.HandleFunc("/save-concurso", concursoHandler.Save).Methods("POST")
	adminSAV.HandleFunc("/delete-concurso/{id}", concursoHandler.Delete).Methods("GET", "POST")

	// Admin Only routes
	adminOnly := router.PathPrefix("/admin").Subrouter()
	adminOnly.Use(middleware.AdminOnlyMiddleware(store))

	adminOnly.HandleFunc("/users", userHandler.List).Methods("GET")
	adminOnly.HandleFunc("/users/create", userHandler.Create).Methods("GET")
	adminOnly.HandleFunc("/users/save", userHandler.Save).Methods("POST")
	adminOnly.HandleFunc("/users/edit/{id}", userHandler.Edit).Methods("GET")
	adminOnly.HandleFunc("/users/update/{id}", userHandler.Update).Methods("POST")
	adminOnly.HandleFunc("/users/delete/{id}", userHandler.Delete).Methods("GET", "POST")

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	return router
}
