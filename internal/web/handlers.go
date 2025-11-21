package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/scttfrdmn/conduit/internal/catalog"
)

// PageData represents data passed to templates
type PageData struct {
	Title       string
	Models      []catalog.Model
	Model       *catalog.Model
	Versions    []catalog.ModelVersion
	Query       string
	Filters     map[string]string
	TotalModels int
	Page        int
	PerPage     int
}

// handleHome renders the home page
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get popular models (sorted by usage stats)
	models, err := s.catalog.Search(catalog.SearchOptions{
		SortBy: "popular",
		Limit:  12,
	})
	if err != nil {
		log.Printf("Error fetching models: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert SearchResult to Model
	modelList := make([]catalog.Model, len(models))
	for i, result := range models {
		model, err := s.catalog.GetModel(result.Name)
		if err != nil {
			continue
		}
		modelList[i] = *model
	}

	data := PageData{
		Title:  "Conduit - ML Model Catalog",
		Models: modelList,
	}

	if err := s.templates.ExecuteTemplate(w, "home.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleBrowse renders the browse page
func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page := 1
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 20
	offset := (page - 1) * perPage

	// Get filters
	domain := query.Get("domain")
	framework := query.Get("framework")
	tags := query.Get("tags")

	searchOpts := catalog.SearchOptions{
		Limit:  perPage,
		Offset: offset,
	}

	if domain != "" {
		searchOpts.Domain = domain
	}
	if framework != "" {
		searchOpts.Framework = framework
	}
	if tags != "" {
		searchOpts.Tags = strings.Split(tags, ",")
	}

	// Get models
	results, err := s.catalog.Search(searchOpts)
	if err != nil {
		log.Printf("Error searching models: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert SearchResult to Model
	modelList := make([]catalog.Model, 0, len(results))
	for _, result := range results {
		model, err := s.catalog.GetModel(result.Name)
		if err != nil {
			continue
		}
		modelList = append(modelList, *model)
	}

	// Get total count
	allModels, _ := s.catalog.ListModels(0, 0)
	totalModels := len(allModels)

	filters := map[string]string{
		"domain":    domain,
		"framework": framework,
		"tags":      tags,
	}

	data := PageData{
		Title:       "Browse Models",
		Models:      modelList,
		Filters:     filters,
		TotalModels: totalModels,
		Page:        page,
		PerPage:     perPage,
	}

	if err := s.templates.ExecuteTemplate(w, "browse.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleSearch renders the search page
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	var modelList []catalog.Model
	if query != "" {
		// Enable fuzzy search
		results, err := s.catalog.Search(catalog.SearchOptions{
			Query:      query,
			FuzzyMatch: true,
			Limit:      50,
		})
		if err != nil {
			log.Printf("Error searching models: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Convert SearchResult to Model
		modelList = make([]catalog.Model, 0, len(results))
		for _, result := range results {
			model, err := s.catalog.GetModel(result.Name)
			if err != nil {
				continue
			}
			modelList = append(modelList, *model)
		}
	}

	data := PageData{
		Title:  "Search Models",
		Models: modelList,
		Query:  query,
	}

	if err := s.templates.ExecuteTemplate(w, "search.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleModelDetail renders the model detail page
func (s *Server) handleModelDetail(w http.ResponseWriter, r *http.Request) {
	// Extract model name from path /model/{name}
	path := strings.TrimPrefix(r.URL.Path, "/model/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	// Get model
	model, err := s.catalog.GetModel(path)
	if err != nil {
		log.Printf("Error fetching model: %v", err)
		http.NotFound(w, r)
		return
	}

	// Get all versions
	versions, err := s.catalog.ListModelVersions(path)
	if err != nil {
		log.Printf("Error fetching versions: %v", err)
	}

	data := PageData{
		Title:    model.Name + " - Model Details",
		Model:    model,
		Versions: versions,
	}

	if err := s.templates.ExecuteTemplate(w, "model.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// API Handlers

// handleAPIModels returns models as JSON
func (s *Server) handleAPIModels(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	limit := 20
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := query.Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	models, err := s.catalog.ListModels(limit, offset)
	if err != nil {
		log.Printf("Error fetching models: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(models); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// handleAPIModel returns a single model as JSON
func (s *Server) handleAPIModel(w http.ResponseWriter, r *http.Request) {
	// Extract model name from path /api/model/{name}
	path := strings.TrimPrefix(r.URL.Path, "/api/model/")
	if path == "" {
		http.Error(w, "Model name required", http.StatusBadRequest)
		return
	}

	model, err := s.catalog.GetModel(path)
	if err != nil {
		http.Error(w, "Model not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(model); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// handleAPISearch returns search results as JSON
func (s *Server) handleAPISearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter required", http.StatusBadRequest)
		return
	}

	results, err := s.catalog.Search(catalog.SearchOptions{
		Query:      query,
		FuzzyMatch: true,
		Limit:      50,
	})
	if err != nil {
		log.Printf("Error searching models: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
