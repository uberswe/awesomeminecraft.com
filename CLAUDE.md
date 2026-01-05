# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Run the server locally
go run .

# Build binary
go build -o awesomeminecraft .

# Build for production (Docker-style)
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o awesomeminecraft .

# Docker build
docker build -t awesomeminecraft .

# Install Playwright for visual testing
npm install
npx playwright install
```

The server runs on `http://localhost:8080` by default.

## Architecture Overview

This is a Go web application serving a curated Minecraft resource directory. Content is stored as markdown files and parsed at startup.

### Core Flow

```
resources/*.md → parser → models.ResourceData → handlers → templates → HTML
```

### Key Packages

- **main.go** - Entry point, route setup, template initialization with clone-based inheritance
- **handlers/** - HTTP handlers for all routes (home, category, resource, search, SEO, OG images)
- **parser/** - Markdown file parsing with frontmatter support
- **models/** - Data structures (Category, Subcategory, Resource, ResourceLink)
- **og/** - Open Graph image generation using fogleman/gg library

### Template System

Templates use Go's html/template with a clone-based inheritance pattern:
1. Base template (`templates/base.html`) and partials (`templates/partials/`) are parsed first
2. Each page template clones the base and adds page-specific content
3. Standalone templates (like `redirect.html`) don't use base

### Content Structure

Resources are organized in `resources/` directory:
```
resources/
├── {category-slug}/
│   ├── _category.md          # Category metadata (description)
│   └── {resource-slug}.md    # Individual resource files
```

Resource markdown files use frontmatter:
```yaml
---
name: Resource Name
url: https://example.com
description: Short description
platform: Java|Bedrock|Both
audience: Casual|Intermediate|Technical|All
price: Free|Freemium|Paid
---
```

### Routes

| Route | Handler | Purpose |
|-------|---------|---------|
| `/` | Home | Category grid |
| `/category/{slug}` | Category | Resource tables by subcategory |
| `/resource/{cat}/{res}` | Resource | Resource detail page |
| `/search` | Search | Full-text search |
| `/go/{cat}/{res}[/{index}]` | Redirect | External link redirect with countdown |
| `/og/*.png` | OGHandler | Dynamic Open Graph images |
| `/sitemap.xml`, `/robots.txt` | SEO handlers | Search engine optimization |

### Styling

CSS uses custom properties defined in `static/css/style.css`:
- `--bg-primary: #0f0f0f` (main background)
- `--bg-secondary: #1a1a1a` (sidebar, header, footer)
- `--bg-card: #1e1e1e` (cards, table headers)
- `--accent: #4ade80` (green accent color)
- `--border: #333333` (borders)

### Deployment

- Kubernetes manifests in `k8s/prod/` and `k8s/dev/`
- GitHub Actions CI/CD builds Docker images and pushes to GHCR
- Production URL: `https://www.awesomeminecraft.com`

## Adding New Resources

1. Create markdown file in appropriate `resources/{category}/` directory
2. Include required frontmatter (name, url, description, platform, audience, price)
3. Parser automatically picks up new files on restart
