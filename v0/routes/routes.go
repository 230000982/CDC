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

	// Public routes
	router.HandleFunc("/", handlers.IndexHandler).Methods("GET")
	router.HandleFunc("/login", authHandler.Login).Methods("GET", "POST")
	router.HandleFunc("/register", authHandler.Register).Methods("GET", "POST")
	router.HandleFunc("/logout", authHandler.Logout).Methods("GET")

	// Protected routes
	// Create a subrouter with auth middleware
	protected := router.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(store))

	// Concurso routes
	protected.HandleFunc("/concursos", concursoHandler.List).Methods("GET")
	protected.HandleFunc("/edit-concurso/{id}", concursoHandler.Edit).Methods("GET")
	protected.HandleFunc("/update-concurso/{id}", concursoHandler.Update).Methods("POST")
	protected.HandleFunc("/create-concurso", concursoHandler.Create).Methods("GET")
	protected.HandleFunc("/save-concurso", concursoHandler.Save).Methods("POST")
	protected.HandleFunc("/delete-concurso/{id}", concursoHandler.Delete).Methods("GET", "POST")
	protected.HandleFunc("/concursos-ordenados", concursoHandler.ListOrdered).Methods("GET")
	protected.HandleFunc("/download-pdf", pdfHandler.Download).Methods("GET")

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	return router
}
