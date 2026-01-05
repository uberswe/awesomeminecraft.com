package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/uberswe/awesomeminecraft.com/handlers"
	"github.com/uberswe/awesomeminecraft.com/parser"
)

func main() {
	// Parse resources from individual markdown files
	data, err := parser.ParseResourcesDir("resources")
	if err != nil {
		log.Fatalf("Failed to parse resources directory: %v", err)
	}
	log.Printf("Loaded %d categories with %d total resources", len(data.Categories), data.TotalResources)

	// Parse templates with custom functions
	funcMap := template.FuncMap{
		"subtract": func(a, b int) int {
			return a - b
		},
	}

	templatesPath := filepath.Join("templates", "*.html")
	templates, err := template.New("").Funcs(funcMap).ParseGlob(templatesPath)
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Create handler
	h := handlers.NewHandler(data, templates)

	// Set up routes
	http.HandleFunc("/", h.Home)
	http.HandleFunc("/category/", h.Category)
	http.HandleFunc("/resource/", h.Resource)
	http.HandleFunc("/search", h.Search)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start server
	port := ":8080"
	fmt.Printf("Server starting at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
