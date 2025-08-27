package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type APIEndpoint struct {
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Description string                 `json:"description"`
	Parameters  []Parameter            `json:"parameters,omitempty"`
	RequestBody *RequestBody           `json:"requestBody,omitempty"`
	Responses   map[string]Response    `json:"responses"`
	Tags        []string               `json:"tags,omitempty"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // "path", "query", "header"
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Example     string `json:"example,omitempty"`
}

type RequestBody struct {
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Content     map[string]ContentType `json:"content"`
}

type ContentType struct {
	Schema  Schema                 `json:"schema"`
	Example map[string]interface{} `json:"example,omitempty"`
}

type Schema struct {
	Type       string             `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string           `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Example     string `json:"example,omitempty"`
}

type Response struct {
	Description string                 `json:"description"`
	Content     map[string]ContentType `json:"content,omitempty"`
}

type OpenAPISpec struct {
	OpenAPI string                    `json:"openapi"`
	Info    Info                      `json:"info"`
	Servers []Server                  `json:"servers"`
	Paths   map[string]map[string]APIEndpoint `json:"paths"`
	Components Components               `json:"components,omitempty"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Contact     Contact `json:"contact,omitempty"`
}

type Contact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type   string `json:"type"`
	Scheme string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

type APIDocGenerator struct {
	spec       *OpenAPISpec
	endpoints  []APIEndpoint
	rootDir    string
}

func NewAPIDocGenerator(rootDir string) *APIDocGenerator {
	return &APIDocGenerator{
		spec: &OpenAPISpec{
			OpenAPI: "3.0.3",
			Info: Info{
				Title:       "Bookmark Sync Service API",
				Description: "A comprehensive bookmark synchronization service with cross-browser support",
				Version:     "1.0.0",
				Contact: Contact{
					Name:  "Bookmark Sync Team",
					Email: "api@bookmark-sync.com",
				},
			},
			Servers: []Server{
				{
					URL:         "http://localhost:8080/api/v1",
					Description: "Development server",
				},
				{
					URL:         "https://api.bookmark-sync.com/v1",
					Description: "Production server",
				},
			},
			Paths: make(map[string]map[string]APIEndpoint),
			Components: Components{
				SecuritySchemes: map[string]SecurityScheme{
					"BearerAuth": {
						Type:         "http",
						Scheme:       "bearer",
						BearerFormat: "JWT",
					},
				},
				Schemas: make(map[string]Schema),
			},
		},
		endpoints: make([]APIEndpoint, 0),
		rootDir:   rootDir,
	}
}

func (g *APIDocGenerator) GenerateFromCode() error {
	// Parse Go source files to extract API endpoints
	err := filepath.Walk(g.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") {
			return nil
		}

		if strings.Contains(path, "/cmd/") || strings.Contains(path, "/pkg/") {
			return nil
		}

		return g.parseGoFile(path)
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %v", err)
	}

	// Add predefined endpoints based on project structure
	g.addPredefinedEndpoints()

	// Build OpenAPI spec from endpoints
	g.buildOpenAPISpec()

	return nil
}

func (g *APIDocGenerator) parseGoFile(filePath string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Look for handler functions and route definitions
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name != nil && strings.Contains(strings.ToLower(x.Name.Name), "handler") {
				// Extract API information from handler functions
				g.extractEndpointFromHandler(x, fset)
			}
		}
		return true
	})

	return nil
}

func (g *APIDocGenerator) extractEndpointFromHandler(fn *ast.FuncDecl, fset *token.FileSet) {
	// Extract endpoint information from function comments and signature
	// This is a simplified implementation
	if fn.Doc != nil {
		for _, comment := range fn.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if strings.Contains(strings.ToLower(text), "endpoint") ||
			   strings.Contains(strings.ToLower(text), "handler") {
				// Try to extract method and path from comments
				g.parseEndpointComment(text, fn.Name.Name)
			}
		}
	}
}

func (g *APIDocGenerator) parseEndpointComment(comment, handlerName string) {
	// Simple parsing logic - in practice this would be more sophisticated
	endpoint := APIEndpoint{
		Method:      "GET", // Default
		Path:        "/unknown",
		Description: comment,
		Responses: map[string]Response{
			"200": {
				Description: "Success",
				Content: map[string]ContentType{
					"application/json": {
						Schema: Schema{Type: "object"},
					},
				},
			},
		},
	}

	g.endpoints = append(g.endpoints, endpoint)
}

func (g *APIDocGenerator) addPredefinedEndpoints() {
	// Add predefined endpoints based on project knowledge
	predefinedEndpoints := []APIEndpoint{
		{
			Method:      "POST",
			Path:        "/auth/login",
			Description: "Authenticate user and return JWT token",
			Tags:        []string{"Authentication"},
			RequestBody: &RequestBody{
				Description: "User credentials",
				Required:    true,
				Content: map[string]ContentType{
					"application/json": {
						Schema: Schema{
							Type: "object",
							Properties: map[string]Property{
								"email": {
									Type:        "string",
									Description: "User email address",
									Example:     "user@example.com",
								},
								"password": {
									Type:        "string",
									Description: "User password",
									Example:     "secretpassword",
								},
							},
							Required: []string{"email", "password"},
						},
					},
				},
			},
			Responses: map[string]Response{
				"200": {
					Description: "Login successful",
					Content: map[string]ContentType{
						"application/json": {
							Schema: Schema{
								Type: "object",
								Properties: map[string]Property{
									"token": {
										Type:        "string",
										Description: "JWT authentication token",
									},
									"user": {
										Type:        "object",
										Description: "User information",
									},
								},
							},
						},
					},
				},
				"401": {
					Description: "Invalid credentials",
				},
			},
		},
		{
			Method:      "POST",
			Path:        "/auth/register",
			Description: "Register new user account",
			Tags:        []string{"Authentication"},
			RequestBody: &RequestBody{
				Description: "User registration data",
				Required:    true,
				Content: map[string]ContentType{
					"application/json": {
						Schema: Schema{
							Type: "object",
							Properties: map[string]Property{
								"email": {
									Type:        "string",
									Description: "User email address",
								},
								"password": {
									Type:        "string",
									Description: "User password",
								},
								"name": {
									Type:        "string",
									Description: "User display name",
								},
							},
							Required: []string{"email", "password", "name"},
						},
					},
				},
			},
			Responses: map[string]Response{
				"201": {
					Description: "User created successfully",
				},
				"400": {
					Description: "Invalid input data",
				},
			},
		},
		{
			Method:      "GET",
			Path:        "/bookmarks",
			Description: "List user bookmarks with pagination and filtering",
			Tags:        []string{"Bookmarks"},
			Parameters: []Parameter{
				{
					Name:        "limit",
					In:          "query",
					Description: "Number of bookmarks to return",
					Required:    false,
					Type:        "integer",
					Example:     "50",
				},
				{
					Name:        "offset",
					In:          "query",
					Description: "Number of bookmarks to skip",
					Required:    false,
					Type:        "integer",
					Example:     "0",
				},
				{
					Name:        "search",
					In:          "query",
					Description: "Search term for filtering bookmarks",
					Required:    false,
					Type:        "string",
				},
			},
			Responses: map[string]Response{
				"200": {
					Description: "List of bookmarks",
					Content: map[string]ContentType{
						"application/json": {
							Schema: Schema{
								Type: "object",
								Properties: map[string]Property{
									"bookmarks": {
										Type:        "array",
										Description: "Array of bookmark objects",
									},
									"total": {
										Type:        "integer",
										Description: "Total number of bookmarks",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Method:      "POST",
			Path:        "/bookmarks",
			Description: "Create a new bookmark",
			Tags:        []string{"Bookmarks"},
			RequestBody: &RequestBody{
				Description: "Bookmark data",
				Required:    true,
				Content: map[string]ContentType{
					"application/json": {
						Schema: Schema{
							Type: "object",
							Properties: map[string]Property{
								"title": {
									Type:        "string",
									Description: "Bookmark title",
								},
								"url": {
									Type:        "string",
									Description: "Bookmark URL",
								},
								"description": {
									Type:        "string",
									Description: "Bookmark description",
								},
								"tags": {
									Type:        "array",
									Description: "Bookmark tags",
								},
							},
							Required: []string{"title", "url"},
						},
					},
				},
			},
			Responses: map[string]Response{
				"201": {
					Description: "Bookmark created successfully",
				},
				"400": {
					Description: "Invalid bookmark data",
				},
			},
		},
		{
			Method:      "GET",
			Path:        "/collections",
			Description: "List user collections",
			Tags:        []string{"Collections"},
			Responses: map[string]Response{
				"200": {
					Description: "List of collections",
				},
			},
		},
		{
			Method:      "POST",
			Path:        "/collections",
			Description: "Create a new collection",
			Tags:        []string{"Collections"},
			RequestBody: &RequestBody{
				Description: "Collection data",
				Required:    true,
				Content: map[string]ContentType{
					"application/json": {
						Schema: Schema{
							Type: "object",
							Properties: map[string]Property{
								"name": {
									Type:        "string",
									Description: "Collection name",
								},
								"description": {
									Type:        "string",
									Description: "Collection description",
								},
							},
							Required: []string{"name"},
						},
					},
				},
			},
			Responses: map[string]Response{
				"201": {
					Description: "Collection created successfully",
				},
			},
		},
	}

	g.endpoints = append(g.endpoints, predefinedEndpoints...)
}

func (g *APIDocGenerator) buildOpenAPISpec() {
	// Group endpoints by path
	for _, endpoint := range g.endpoints {
		if g.spec.Paths[endpoint.Path] == nil {
			g.spec.Paths[endpoint.Path] = make(map[string]APIEndpoint)
		}
		g.spec.Paths[endpoint.Path][strings.ToLower(endpoint.Method)] = endpoint
	}
}

func (g *APIDocGenerator) GenerateOpenAPISpec(outputPath string) error {
	// Create docs directory
	docsDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %v", err)
	}

	// Marshal spec to JSON
	specJSON, err := json.MarshalIndent(g.spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec: %v", err)
	}

	// Write to file
	if err := ioutil.WriteFile(outputPath, specJSON, 0644); err != nil {
		return fmt.Errorf("failed to write OpenAPI spec: %v", err)
	}

	log.Printf("Generated OpenAPI spec: %s", outputPath)
	return nil
}

func (g *APIDocGenerator) GenerateMarkdownDocs(outputPath string) error {
	// Create docs directory
	docsDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %v", err)
	}

	var sb strings.Builder

	// Write header
	sb.WriteString("# Bookmark Sync Service API Documentation\n\n")
	sb.WriteString("## Overview\n")
	sb.WriteString(g.spec.Info.Description + "\n\n")
	sb.WriteString("## Base URLs\n")

	for _, server := range g.spec.Servers {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", server.Description, server.URL))
	}
	sb.WriteString("\n## Authentication\n")
	sb.WriteString("This API uses JWT Bearer token authentication. Include the token in the Authorization header:\n")
	sb.WriteString("```\nAuthorization: Bearer <your-jwt-token>\n```\n\n")

	// Group endpoints by tags
	tagGroups := make(map[string][]APIEndpoint)
	for _, pathMethods := range g.spec.Paths {
		for _, endpoint := range pathMethods {
			for _, tag := range endpoint.Tags {
				tagGroups[tag] = append(tagGroups[tag], endpoint)
			}
		}
	}

	// Write endpoints by tag
	for tag, endpoints := range tagGroups {
		sb.WriteString(fmt.Sprintf("## %s\n\n", tag))

		for _, endpoint := range endpoints {
			sb.WriteString(fmt.Sprintf("### %s %s\n\n", strings.ToUpper(endpoint.Method), endpoint.Path))
			sb.WriteString(fmt.Sprintf("%s\n\n", endpoint.Description))

			// Parameters
			if len(endpoint.Parameters) > 0 {
				sb.WriteString("#### Parameters\n\n")
				sb.WriteString("| Name | Type | In | Required | Description |\n")
				sb.WriteString("|------|------|----|---------| ----------- |\n")

				for _, param := range endpoint.Parameters {
					required := "No"
					if param.Required {
						required = "Yes"
					}
					sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
						param.Name, param.Type, param.In, required, param.Description))
				}
				sb.WriteString("\n")
			}

			// Request body
			if endpoint.RequestBody != nil {
				sb.WriteString("#### Request Body\n\n")
				sb.WriteString(fmt.Sprintf("%s\n\n", endpoint.RequestBody.Description))
				sb.WriteString("```json\n")
				if content, ok := endpoint.RequestBody.Content["application/json"]; ok {
					if content.Example != nil {
						exampleJSON, _ := json.MarshalIndent(content.Example, "", "  ")
						sb.WriteString(string(exampleJSON))
					} else {
						sb.WriteString("{\n  // Request body example\n}")
					}
				}
				sb.WriteString("\n```\n\n")
			}

			// Responses
			sb.WriteString("#### Responses\n\n")
			for status, response := range endpoint.Responses {
				sb.WriteString(fmt.Sprintf("**%s**: %s\n\n", status, response.Description))
			}
			sb.WriteString("\n---\n\n")
		}
	}

	// Write to file
	if err := ioutil.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write markdown docs: %v", err)
	}

	log.Printf("Generated Markdown docs: %s", outputPath)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run doc_generator.go <backend-directory>")
	}

	backendDir := os.Args[1]
	generator := NewAPIDocGenerator(backendDir)

	// Generate from code
	if err := generator.GenerateFromCode(); err != nil {
		log.Printf("Error generating from code: %v", err)
	}

	// Create output directory
	docsDir := "./docs/api"
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		log.Fatalf("Failed to create docs directory: %v", err)
	}

	// Generate OpenAPI spec
	if err := generator.GenerateOpenAPISpec(filepath.Join(docsDir, "openapi.json")); err != nil {
		log.Fatalf("Failed to generate OpenAPI spec: %v", err)
	}

	// Generate Markdown documentation
	if err := generator.GenerateMarkdownDocs(filepath.Join(docsDir, "README.md")); err != nil {
		log.Fatalf("Failed to generate Markdown docs: %v", err)
	}

	fmt.Println("✅ API documentation generated successfully!")
	fmt.Printf("📄 OpenAPI spec: %s\n", filepath.Join(docsDir, "openapi.json"))
	fmt.Printf("📚 Markdown docs: %s\n", filepath.Join(docsDir, "README.md"))
}
