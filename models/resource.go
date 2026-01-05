package models

// Resource represents a single Minecraft resource entry
type Resource struct {
	Name            string
	Slug            string
	URL             string
	Description     string
	Platform        string
	Audience        string
	Price           string
	CategorySlug    string
	SubcategorySlug string
	CategoryName    string
	SubcategoryName string
}

// Subcategory represents a subcategory within a main category
type Subcategory struct {
	Name      string
	Slug      string
	Resources []Resource
}

// Category represents a main category of resources
type Category struct {
	Name          string
	Slug          string
	Subcategories []Subcategory
}

// SearchResult represents a resource with its category context for search results
type SearchResult struct {
	Resource    Resource
	Category    string
	Subcategory string
}

// SiteData holds all parsed data from the markdown file
type SiteData struct {
	Categories     []Category
	TotalResources int
}
