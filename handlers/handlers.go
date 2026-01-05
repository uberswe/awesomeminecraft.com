package handlers

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/uberswe/awesomeminecraft.com/models"
	"github.com/uberswe/awesomeminecraft.com/parser"
)

// Handler holds the application dependencies
type Handler struct {
	Data      *models.SiteData
	Templates *template.Template
}

// NewHandler creates a new handler with the given data and templates
func NewHandler(data *models.SiteData, templates *template.Template) *Handler {
	return &Handler{
		Data:      data,
		Templates: templates,
	}
}

// HomeData contains data for the home page template
type HomeData struct {
	Categories     []models.Category
	TotalResources int
}

// CategoryData contains data for the category page template
type CategoryData struct {
	Category       *models.Category
	Categories     []models.Category
	TotalResources int
}

// SearchData contains data for the search results page template
type SearchData struct {
	Query          string
	Results        []models.SearchResult
	ResultCount    int
	Categories     []models.Category
	TotalResources int
}

// ResourceData contains data for the resource page template
type ResourceData struct {
	Resource       *models.Resource
	Categories     []models.Category
	TotalResources int
}

// Home handles the home page
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := HomeData{
		Categories:     h.Data.Categories,
		TotalResources: h.Data.TotalResources,
	}

	err := h.Templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Category handles category pages
func (h *Handler) Category(w http.ResponseWriter, r *http.Request) {
	// Extract slug from path /category/{slug}
	path := strings.TrimPrefix(r.URL.Path, "/category/")
	slug := strings.TrimSuffix(path, "/")

	if slug == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	category := parser.GetCategoryBySlug(h.Data, slug)
	if category == nil {
		http.NotFound(w, r)
		return
	}

	data := CategoryData{
		Category:       category,
		Categories:     h.Data.Categories,
		TotalResources: h.Data.TotalResources,
	}

	err := h.Templates.ExecuteTemplate(w, "category.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Search handles search requests
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	var results []models.SearchResult
	if query != "" {
		results = parser.Search(h.Data, query)
	}

	data := SearchData{
		Query:          query,
		Results:        results,
		ResultCount:    len(results),
		Categories:     h.Data.Categories,
		TotalResources: h.Data.TotalResources,
	}

	err := h.Templates.ExecuteTemplate(w, "search.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Resource handles individual resource pages
func (h *Handler) Resource(w http.ResponseWriter, r *http.Request) {
	// Extract category and resource slugs from path /resource/{category}/{resource}
	path := strings.TrimPrefix(r.URL.Path, "/resource/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	categorySlug := parts[0]
	resourceSlug := parts[1]

	resource := parser.GetResourceBySlug(h.Data, categorySlug, resourceSlug)
	if resource == nil {
		http.NotFound(w, r)
		return
	}

	data := ResourceData{
		Resource:       resource,
		Categories:     h.Data.Categories,
		TotalResources: h.Data.TotalResources,
	}

	err := h.Templates.ExecuteTemplate(w, "resource.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
