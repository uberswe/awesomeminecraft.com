package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/uberswe/awesomeminecraft.com/models"
)

// slugify converts a string to a URL-friendly slug
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	// Replace spaces and special chars with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// Slugify is exported for use by other packages
func Slugify(s string) string {
	return slugify(s)
}

// parseTableRow parses a markdown table row into a Resource
func parseTableRow(line string, categoryName, categorySlug, subcategoryName, subcategorySlug string) *models.Resource {
	// Skip header and separator rows
	if strings.Contains(line, "Resource") && strings.Contains(line, "URL") {
		return nil
	}
	if strings.Contains(line, "---") {
		return nil
	}

	// Split by | and trim
	parts := strings.Split(line, "|")
	if len(parts) < 7 {
		return nil
	}

	// Extract name (remove ** bold markers)
	name := strings.TrimSpace(parts[1])
	name = strings.Trim(name, "*")

	url := strings.TrimSpace(parts[2])
	description := strings.TrimSpace(parts[3])
	platform := strings.TrimSpace(parts[4])
	audience := strings.TrimSpace(parts[5])
	price := strings.TrimSpace(parts[6])

	if name == "" || url == "" {
		return nil
	}

	// Create URLs slice with the primary URL
	urls := []models.ResourceLink{{URL: url}}

	return &models.Resource{
		Name:            name,
		Slug:            slugify(name),
		URL:             url,
		URLs:            urls,
		Description:     description,
		Platform:        platform,
		Audience:        audience,
		Price:           price,
		CategorySlug:    categorySlug,
		SubcategorySlug: subcategorySlug,
		CategoryName:    categoryName,
		SubcategoryName: subcategoryName,
	}
}

// ParseMarkdown parses the resources.md file and returns structured data
func ParseMarkdown(filename string) (*models.SiteData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var categories []models.Category
	var currentCategory *models.Category
	var currentSubcategory *models.Subcategory
	totalResources := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for main category (## Header)
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			// Save previous category if exists
			if currentCategory != nil {
				if currentSubcategory != nil {
					currentCategory.Subcategories = append(currentCategory.Subcategories, *currentSubcategory)
					currentSubcategory = nil
				}
				categories = append(categories, *currentCategory)
			}

			name := strings.TrimPrefix(line, "## ")
			currentCategory = &models.Category{
				Name: name,
				Slug: slugify(name),
			}
			continue
		}

		// Check for subcategory (### Header)
		if strings.HasPrefix(line, "### ") {
			if currentCategory != nil && currentSubcategory != nil {
				currentCategory.Subcategories = append(currentCategory.Subcategories, *currentSubcategory)
			}

			name := strings.TrimPrefix(line, "### ")
			currentSubcategory = &models.Subcategory{
				Name: name,
				Slug: slugify(name),
			}
			continue
		}

		// Check for inline subcategory (**Bold Text:**)
		if strings.HasPrefix(line, "**") && strings.HasSuffix(strings.TrimSpace(line), ":**") {
			if currentCategory != nil && currentSubcategory != nil {
				currentCategory.Subcategories = append(currentCategory.Subcategories, *currentSubcategory)
			}

			name := strings.TrimPrefix(line, "**")
			name = strings.TrimSuffix(name, ":**")
			name = strings.TrimSuffix(name, ":")
			name = strings.TrimSpace(name)
			currentSubcategory = &models.Subcategory{
				Name: name,
				Slug: slugify(name),
			}
			continue
		}

		// Check for table row
		if strings.HasPrefix(line, "|") {
			catName := ""
			catSlug := ""
			subName := ""
			subSlug := ""
			if currentCategory != nil {
				catName = currentCategory.Name
				catSlug = currentCategory.Slug
			}
			if currentSubcategory != nil {
				subName = currentSubcategory.Name
				subSlug = currentSubcategory.Slug
			}
			resource := parseTableRow(line, catName, catSlug, subName, subSlug)
			if resource != nil {
				totalResources++
				if currentSubcategory != nil {
					currentSubcategory.Resources = append(currentSubcategory.Resources, *resource)
				} else if currentCategory != nil {
					// Create a default subcategory if none exists
					if len(currentCategory.Subcategories) == 0 {
						currentSubcategory = &models.Subcategory{
							Name: "Resources",
							Slug: "resources",
						}
					}
					if currentSubcategory != nil {
						currentSubcategory.Resources = append(currentSubcategory.Resources, *resource)
					}
				}
			}
		}
	}

	// Save last category and subcategory
	if currentCategory != nil {
		if currentSubcategory != nil {
			currentCategory.Subcategories = append(currentCategory.Subcategories, *currentSubcategory)
		}
		categories = append(categories, *currentCategory)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &models.SiteData{
		Categories:     categories,
		TotalResources: totalResources,
	}, nil
}

// parseResourceFile parses an individual resource markdown file
func parseResourceFile(filePath string) (*models.Resource, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var name, url, description, platform, audience, price string
	var urls []models.ResourceLink
	var categoryName, subcategoryName string

	scanner := bufio.NewScanner(file)
	inDetails := false

	for scanner.Scan() {
		line := scanner.Text()

		// Parse name from # heading
		if strings.HasPrefix(line, "# ") && name == "" {
			name = strings.TrimPrefix(line, "# ")
			continue
		}

		// Parse website URL(s) - supports multiple **Website:** lines
		if strings.HasPrefix(line, "**Website:**") {
			// Extract URL and label from markdown link [label](url)
			re := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 3 {
				label := matches[1]
				linkURL := matches[2]
				// If label looks like a URL, use empty label (for backward compat)
				if strings.HasPrefix(label, "http://") || strings.HasPrefix(label, "https://") {
					label = ""
				}
				urls = append(urls, models.ResourceLink{
					URL:   linkURL,
					Label: label,
				})
				// Set primary URL from first link for backward compatibility
				if url == "" {
					url = linkURL
				}
			}
			continue
		}

		// Parse category
		if strings.HasPrefix(line, "**Category:**") {
			catLine := strings.TrimPrefix(line, "**Category:**")
			catLine = strings.TrimSpace(catLine)
			parts := strings.Split(catLine, " > ")
			if len(parts) >= 1 {
				categoryName = strings.TrimSpace(parts[0])
			}
			if len(parts) >= 2 {
				subcategoryName = strings.TrimSpace(parts[1])
			}
			continue
		}

		// Parse overview/description
		if strings.HasPrefix(line, "## Overview") {
			continue
		}

		// Capture description (first non-empty line after Overview)
		if description == "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "**") && !strings.HasPrefix(line, "---") && !strings.HasPrefix(line, "|") && strings.TrimSpace(line) != "" {
			description = strings.TrimSpace(line)
			continue
		}

		// Parse details table
		if strings.HasPrefix(line, "## Details") {
			inDetails = true
			continue
		}

		if inDetails && strings.HasPrefix(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				key := strings.TrimSpace(parts[1])
				key = strings.Trim(key, "*")
				value := strings.TrimSpace(parts[2])

				switch key {
				case "Platform":
					platform = value
				case "Audience":
					audience = value
				case "Price":
					price = value
				}
			}
		}

		// Stop parsing at the end
		if inDetails && strings.HasPrefix(line, "---") {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if name == "" || url == "" {
		return nil, nil
	}

	categorySlug := slugify(categoryName)
	subcategorySlug := slugify(subcategoryName)

	return &models.Resource{
		Name:            name,
		Slug:            slugify(name),
		URL:             url,
		URLs:            urls,
		Description:     description,
		Platform:        platform,
		Audience:        audience,
		Price:           price,
		CategorySlug:    categorySlug,
		SubcategorySlug: subcategorySlug,
		CategoryName:    categoryName,
		SubcategoryName: subcategoryName,
	}, nil
}

// parseCategoryDescription parses a _category.md file and returns the description
func parseCategoryDescription(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	frontmatterDone := false

	for scanner.Scan() {
		line := scanner.Text()

		// Check for frontmatter delimiters
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				frontmatterDone = true
				continue
			}
		}

		// Parse frontmatter
		if inFrontmatter && !frontmatterDone {
			if strings.HasPrefix(line, "description:") {
				desc := strings.TrimPrefix(line, "description:")
				desc = strings.TrimSpace(desc)
				// Remove surrounding quotes if present
				desc = strings.Trim(desc, "\"'")
				return desc
			}
		}

		// If we're past frontmatter, first non-empty line is the description
		if frontmatterDone && strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}

	return ""
}

// ParseResourcesDir parses all resource markdown files from the resources directory
func ParseResourcesDir(dirPath string) (*models.SiteData, error) {
	// Map to collect resources by category and subcategory
	categoryMap := make(map[string]*models.Category)
	categoryDescriptions := make(map[string]string)                   // categorySlug -> description
	subcategoryMap := make(map[string]map[string]*models.Subcategory) // categorySlug -> subcategorySlug -> Subcategory
	totalResources := 0

	// Walk through the resources directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Check for _category.md files
		if strings.HasSuffix(path, "_category.md") {
			// Extract category slug from directory name
			dir := filepath.Dir(path)
			catSlug := filepath.Base(dir)
			desc := parseCategoryDescription(path)
			if desc != "" {
				categoryDescriptions[catSlug] = desc
			}
			return nil
		}

		// Parse the resource file
		resource, err := parseResourceFile(path)
		if err != nil {
			return nil // Skip files that can't be parsed
		}
		if resource == nil {
			return nil
		}

		totalResources++

		// Get or create category
		catSlug := resource.CategorySlug
		if catSlug == "" {
			catSlug = "uncategorized"
			resource.CategoryName = "Uncategorized"
			resource.CategorySlug = catSlug
		}

		if _, exists := categoryMap[catSlug]; !exists {
			categoryMap[catSlug] = &models.Category{
				Name: resource.CategoryName,
				Slug: catSlug,
			}
			subcategoryMap[catSlug] = make(map[string]*models.Subcategory)
		}

		// Get or create subcategory
		subSlug := resource.SubcategorySlug
		if subSlug == "" {
			subSlug = "general"
			resource.SubcategoryName = "General"
			resource.SubcategorySlug = subSlug
		}

		if _, exists := subcategoryMap[catSlug][subSlug]; !exists {
			subcategoryMap[catSlug][subSlug] = &models.Subcategory{
				Name: resource.SubcategoryName,
				Slug: subSlug,
			}
		}

		// Add resource to subcategory
		subcategoryMap[catSlug][subSlug].Resources = append(subcategoryMap[catSlug][subSlug].Resources, *resource)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert maps to slices and sort
	var categories []models.Category
	for catSlug, category := range categoryMap {
		var subcategories []models.Subcategory
		for _, subcategory := range subcategoryMap[catSlug] {
			// Sort resources by name within subcategory
			sort.Slice(subcategory.Resources, func(i, j int) bool {
				return subcategory.Resources[i].Name < subcategory.Resources[j].Name
			})
			subcategories = append(subcategories, *subcategory)
		}
		// Sort subcategories by name
		sort.Slice(subcategories, func(i, j int) bool {
			return subcategories[i].Name < subcategories[j].Name
		})
		category.Subcategories = subcategories

		// Apply category description if available
		if desc, exists := categoryDescriptions[catSlug]; exists {
			category.Description = desc
		}

		categories = append(categories, *category)
	}

	// Sort categories by name
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	return &models.SiteData{
		Categories:     categories,
		TotalResources: totalResources,
	}, nil
}

// Search searches through all resources and returns matching results
func Search(data *models.SiteData, query string) []models.SearchResult {
	var results []models.SearchResult
	query = strings.ToLower(query)

	for _, category := range data.Categories {
		for _, subcategory := range category.Subcategories {
			for _, resource := range subcategory.Resources {
				// Search in name, description, platform, audience
				if strings.Contains(strings.ToLower(resource.Name), query) ||
					strings.Contains(strings.ToLower(resource.Description), query) ||
					strings.Contains(strings.ToLower(resource.Platform), query) ||
					strings.Contains(strings.ToLower(category.Name), query) ||
					strings.Contains(strings.ToLower(subcategory.Name), query) {
					results = append(results, models.SearchResult{
						Resource:    resource,
						Category:    category.Name,
						Subcategory: subcategory.Name,
					})
				}
			}
		}
	}

	return results
}

// GetCategoryBySlug finds a category by its slug
func GetCategoryBySlug(data *models.SiteData, slug string) *models.Category {
	for _, category := range data.Categories {
		if category.Slug == slug {
			return &category
		}
	}
	return nil
}

// GetResourceBySlug finds a resource by category slug and resource slug
func GetResourceBySlug(data *models.SiteData, categorySlug, resourceSlug string) *models.Resource {
	for _, category := range data.Categories {
		if category.Slug == categorySlug {
			for _, subcategory := range category.Subcategories {
				for _, resource := range subcategory.Resources {
					if resource.Slug == resourceSlug {
						return &resource
					}
				}
			}
		}
	}
	return nil
}
