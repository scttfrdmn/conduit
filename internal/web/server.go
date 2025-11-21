package web

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/scttfrdmn/conduit/internal/catalog"
)

//go:embed templates/* static/*
var content embed.FS

// Server represents the web server
type Server struct {
	catalog   *catalog.DB
	templates *template.Template
	addr      string
}

// NewServer creates a new web server
func NewServer(catalogDB *catalog.DB, addr string) (*Server, error) {
	// Create template functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
	}

	// Parse templates with functions
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(content, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Server{
		catalog:   catalogDB,
		templates: tmpl,
		addr:      addr,
	}, nil
}

// Start starts the web server
func (s *Server) Start() error {
	// Setup routes
	mux := http.NewServeMux()

	// Pages
	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/browse", s.handleBrowse)
	mux.HandleFunc("/search", s.handleSearch)
	mux.HandleFunc("/model/", s.handleModelDetail)

	// API endpoints
	mux.HandleFunc("/api/models", s.handleAPIModels)
	mux.HandleFunc("/api/model/", s.handleAPIModel)
	mux.HandleFunc("/api/search", s.handleAPISearch)

	// Static files
	mux.Handle("/static/", http.FileServer(http.FS(content)))

	log.Printf("Starting web server on %s", s.addr)
	log.Printf("Open http://%s in your browser", s.addr)

	return http.ListenAndServe(s.addr, mux)
}
