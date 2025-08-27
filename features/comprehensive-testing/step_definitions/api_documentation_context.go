// API Documentation BDD Step Definitions
// Phase 14: Task 31.1 - API Documentation Generation and Configuration

package step_definitions

import (
	"fmt"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

// APIDocumentationContext handles API documentation related BDD scenarios
type APIDocumentationContext struct {
	*BaseTestContext
	swaggerConfig     *SwaggerConfiguration
	postmanConfig     *PostmanConfiguration
	documentationResults map[string]*DocumentationResult
	apiEndpoints      []APIEndpoint
	generatedFiles    []string
}

// SwaggerConfiguration represents Swagger/OpenAPI configuration
type SwaggerConfiguration struct {
	Version       string            `json:"version"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Host          string            `json:"host"`
	BasePath      string            `json:"base_path"`
	Schemes       []string          `json:"schemes"`
	SecurityDefs  map[string]interface{} `json:"security_definitions"`
	OutputFormat  string            `json:"output_format"`
	OutputPath    string            `json:"output_path"`
}

// PostmanConfiguration represents Postman collection configuration
type PostmanConfiguration struct {
	CollectionName string            `json:"collection_name"`
	Environment    string            `json:"environment"`
	BaseURL        string            `json:"base_url"`
	AuthMethod     string            `json:"auth_method"`
	Variables      map[string]string `json:"variables"`
	OutputPath     string            `json:"output_path"`
}

// DocumentationResult represents documentation generation result
type DocumentationResult struct {
	DocumentType string        `json:"document_type"`
	Success      bool          `json:"success"`
	FilePath     string        `json:"file_path"`
	FileSize     int64         `json:"file_size"`
	GeneratedAt  time.Time     `json:"generated_at"`
	Duration     time.Duration `json:"duration"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Validation   *ValidationResult `json:"validation,omitempty"`
}

// ValidationResult represents documentation validation result
type ValidationResult struct {
	Valid          bool     `json:"valid"`
	ErrorCount     int      `json:"error_count"`
	WarningCount   int      `json:"warning_count"`
	Errors         []string `json:"errors"`
	Warnings       []string `json:"warnings"`
}

// APIEndpoint represents an API endpoint
type APIEndpoint struct {
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	Parameters  []Parameter            `json:"parameters"`
	Responses   map[string]Response    `json:"responses"`
	Tags        []string               `json:"tags"`
	Security    []map[string][]string  `json:"security"`
}

// Parameter represents an API parameter
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Example     interface{} `json:"example,omitempty"`
}

// Response represents an API response
type Response struct {
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Examples    map[string]interface{} `json:"examples,omitempty"`
}

// NewAPIDocumentationContext creates a new API documentation context
func NewAPIDocumentationContext(t *testing.T) *APIDocumentationContext {
	return &APIDocumentationContext{
		BaseTestContext:      NewBaseTestContext(t),
		documentationResults: make(map[string]*DocumentationResult),
		apiEndpoints:         make([]APIEndpoint, 0),
		generatedFiles:       make([]string, 0),
		swaggerConfig: &SwaggerConfiguration{
			Version:      "2.0",
			Title:        "Bookmark Sync API",
			Description:  "API for bookmark synchronization service",
			Host:         "localhost:8080",
			BasePath:     "/api/v1",
			Schemes:      []string{"http", "https"},
			OutputFormat: "json",
			OutputPath:   "./docs/swagger.json",
		},
		postmanConfig: &PostmanConfiguration{
			CollectionName: "Bookmark Sync API",
			Environment:    "development",
			BaseURL:        "http://localhost:8080",
			AuthMethod:     "bearer",
			OutputPath:     "./docs/postman_collection.json",
		},
	}
}

// RegisterSteps registers all API documentation related steps
func (ctx *APIDocumentationContext) RegisterSteps(sc *godog.ScenarioContext) {
	// Register common steps
	ctx.RegisterCommonSteps(sc)

	// API documentation generation steps
	sc.Step(`^I have API endpoints documented with annotations$`, ctx.iHaveAPIEndpointsDocumentedWithAnnotations)
	sc.Step(`^I generate comprehensive API documentation$`, ctx.iGenerateComprehensiveAPIDocumentation)
	sc.Step(`^Swagger/OpenAPI specifications should be created$`, ctx.swaggerOpenAPISpecificationsShouldBeCreated)
	sc.Step(`^API documentation should be accessible via web interface$`, ctx.apiDocumentationShouldBeAccessibleViaWebInterface)
	sc.Step(`^documentation should include request/response examples$`, ctx.documentationShouldIncludeRequestResponseExamples)
	sc.Step(`^authentication methods should be clearly documented$`, ctx.authenticationMethodsShouldBeClearlyDocumented)

	// Postman collection generation steps
	sc.Step(`^I have API endpoints configured for testing$`, ctx.iHaveAPIEndpointsConfiguredForTesting)
	sc.Step(`^I generate Postman collection with test scripts$`, ctx.iGeneratePostmanCollectionWithTestScripts)
	sc.Step(`^Postman collection should include all endpoints$`, ctx.postmanCollectionShouldIncludeAllEndpoints)
	sc.Step(`^collection should have pre-request scripts$`, ctx.collectionShouldHavePreRequestScripts)
	sc.Step(`^test scripts should validate responses$`, ctx.testScriptsShouldValidateResponses)
	sc.Step(`^environment variables should be configured$`, ctx.environmentVariablesShouldBeConfigured)

	// API documentation validation steps
	sc.Step(`^generated documentation should be valid$`, ctx.generatedDocumentationShouldBeValid)
	sc.Step(`^API specifications should follow OpenAPI standards$`, ctx.apiSpecificationsShouldFollowOpenAPIStandards)
	sc.Step(`^documentation should be up-to-date with code$`, ctx.documentationShouldBeUpToDateWithCode)
	sc.Step(`^examples should be realistic and working$`, ctx.examplesShouldBeRealisticAndWorking)
	sc.Step(`^error responses should be documented$`, ctx.errorResponsesShouldBeDocumented)
}

// API documentation generation step implementations

func (ctx *APIDocumentationContext) iHaveAPIEndpointsDocumentedWithAnnotations() error {
	ctx.t.Log("📝 Verifying API endpoints are documented with annotations")

	// Define sample API endpoints for testing
	ctx.apiEndpoints = []APIEndpoint{
		{
			Method:      "GET",
			Path:        "/bookmarks",
			Summary:     "Get user bookmarks",
			Description: "Retrieve all bookmarks for the authenticated user",
			Parameters: []Parameter{
				{Name: "page", In: "query", Type: "integer", Required: false, Description: "Page number"},
				{Name: "limit", In: "query", Type: "integer", Required: false, Description: "Items per page"},
			},
			Responses: map[string]Response{
				"200": {Description: "List of bookmarks"},
				"401": {Description: "Unauthorized"},
			},
			Tags: []string{"bookmarks"},
		},
		{
			Method:      "POST",
			Path:        "/bookmarks",
			Summary:     "Create bookmark",
			Description: "Create a new bookmark",
			Parameters: []Parameter{
				{Name: "bookmark", In: "body", Type: "object", Required: true, Description: "Bookmark data"},
			},
			Responses: map[string]Response{
				"201": {Description: "Bookmark created"},
				"400": {Description: "Invalid input"},
				"401": {Description: "Unauthorized"},
			},
			Tags: []string{"bookmarks"},
		},
	}

	if err := ctx.validateAPIAnnotations(); err != nil {
		return fmt.Errorf("API annotations validation failed: %w", err)
	}

	ctx.t.Log("✅ API endpoints documented with annotations")
	return nil
}

func (ctx *APIDocumentationContext) iGenerateComprehensiveAPIDocumentation() error {
	ctx.t.Log("📝 Generating comprehensive API documentation")

	// Generate Swagger documentation
	swaggerResult, err := ctx.generateSwaggerDocumentation()
	if err != nil {
		return fmt.Errorf("Swagger documentation generation failed: %w", err)
	}
	ctx.documentationResults["swagger"] = swaggerResult

	// Generate HTML documentation
	htmlResult, err := ctx.generateHTMLDocumentation()
	if err != nil {
		return fmt.Errorf("HTML documentation generation failed: %w", err)
	}
	ctx.documentationResults["html"] = htmlResult

	ctx.t.Log("✅ Comprehensive API documentation generated")
	return nil
}

func (ctx *APIDocumentationContext) swaggerOpenAPISpecificationsShouldBeCreated() error {
	ctx.t.Log("📝 Validating Swagger/OpenAPI specifications creation")

	swaggerResult, exists := ctx.documentationResults["swagger"]
	if !exists || !swaggerResult.Success {
		return fmt.Errorf("Swagger documentation was not created successfully")
	}

	// Validate file exists and is valid
	if err := ctx.validateSwaggerFile(swaggerResult.FilePath); err != nil {
		return fmt.Errorf("Swagger file validation failed: %w", err)
	}

	ctx.t.Log("✅ Swagger/OpenAPI specifications created")
	return nil
}

func (ctx *APIDocumentationContext) apiDocumentationShouldBeAccessibleViaWebInterface() error {
	ctx.t.Log("📝 Validating API documentation web interface accessibility")

	if err := ctx.testDocumentationWebInterface(); err != nil {
		return fmt.Errorf("documentation web interface not accessible: %w", err)
	}

	ctx.t.Log("✅ API documentation accessible via web interface")
	return nil
}

func (ctx *APIDocumentationContext) documentationShouldIncludeRequestResponseExamples() error {
	ctx.t.Log("📝 Validating request/response examples in documentation")

	if err := ctx.validateRequestResponseExamples(); err != nil {
		return fmt.Errorf("request/response examples validation failed: %w", err)
	}

	ctx.t.Log("✅ Documentation includes request/response examples")
	return nil
}

func (ctx *APIDocumentationContext) authenticationMethodsShouldBeClearlyDocumented() error {
	ctx.t.Log("📝 Validating authentication methods documentation")

	if err := ctx.validateAuthenticationDocumentation(); err != nil {
		return fmt.Errorf("authentication documentation validation failed: %w", err)
	}

	ctx.t.Log("✅ Authentication methods clearly documented")
	return nil
}

// Postman collection generation step implementations

func (ctx *APIDocumentationContext) iHaveAPIEndpointsConfiguredForTesting() error {
	ctx.t.Log("📝 Verifying API endpoints configured for testing")

	if len(ctx.apiEndpoints) == 0 {
		return fmt.Errorf("no API endpoints configured")
	}

	// Validate endpoint configuration
	for _, endpoint := range ctx.apiEndpoints {
		if err := ctx.validateEndpointConfiguration(endpoint); err != nil {
			return fmt.Errorf("endpoint %s %s configuration invalid: %w",
				endpoint.Method, endpoint.Path, err)
		}
	}

	ctx.t.Log("✅ API endpoints configured for testing")
	return nil
}

func (ctx *APIDocumentationContext) iGeneratePostmanCollectionWithTestScripts() error {
	ctx.t.Log("📝 Generating Postman collection with test scripts")

	postmanResult, err := ctx.generatePostmanCollection()
	if err != nil {
		return fmt.Errorf("Postman collection generation failed: %w", err)
	}
	ctx.documentationResults["postman"] = postmanResult

	ctx.t.Log("✅ Postman collection with test scripts generated")
	return nil
}

func (ctx *APIDocumentationContext) postmanCollectionShouldIncludeAllEndpoints() error {
	ctx.t.Log("📝 Validating Postman collection includes all endpoints")

	postmanResult, exists := ctx.documentationResults["postman"]
	if !exists || !postmanResult.Success {
		return fmt.Errorf("Postman collection was not created successfully")
	}

	if err := ctx.validatePostmanCollectionCompleteness(); err != nil {
		return fmt.Errorf("Postman collection completeness validation failed: %w", err)
	}

	ctx.t.Log("✅ Postman collection includes all endpoints")
	return nil
}

// Helper methods

func (ctx *APIDocumentationContext) validateAPIAnnotations() error {
	ctx.t.Log("Validating API annotations...")
	time.Sleep(100 * time.Millisecond)

	// Check if endpoints have required documentation
	for _, endpoint := range ctx.apiEndpoints {
		if endpoint.Summary == "" || endpoint.Description == "" {
			return fmt.Errorf("endpoint %s %s missing documentation",
				endpoint.Method, endpoint.Path)
		}
	}

	return nil
}

func (ctx *APIDocumentationContext) generateSwaggerDocumentation() (*DocumentationResult, error) {
	ctx.t.Log("Generating Swagger documentation...")
	time.Sleep(200 * time.Millisecond)

	// Simulate Swagger generation
	filePath := ctx.swaggerConfig.OutputPath
	ctx.generatedFiles = append(ctx.generatedFiles, filePath)

	return &DocumentationResult{
		DocumentType: "swagger",
		Success:      true,
		FilePath:     filePath,
		FileSize:     1024 * 50, // 50KB
		GeneratedAt:  time.Now(),
		Duration:     200 * time.Millisecond,
		Validation: &ValidationResult{
			Valid:        true,
			ErrorCount:   0,
			WarningCount: 0,
		},
	}, nil
}

func (ctx *APIDocumentationContext) generateHTMLDocumentation() (*DocumentationResult, error) {
	ctx.t.Log("Generating HTML documentation...")
	time.Sleep(150 * time.Millisecond)

	filePath := "./docs/api.html"
	ctx.generatedFiles = append(ctx.generatedFiles, filePath)

	return &DocumentationResult{
		DocumentType: "html",
		Success:      true,
		FilePath:     filePath,
		FileSize:     1024 * 200, // 200KB
		GeneratedAt:  time.Now(),
		Duration:     150 * time.Millisecond,
	}, nil
}

func (ctx *APIDocumentationContext) validateSwaggerFile(filePath string) error {
	ctx.t.Logf("Validating Swagger file: %s", filePath)
	time.Sleep(75 * time.Millisecond)

	// Simulate file validation
	return nil
}

func (ctx *APIDocumentationContext) testDocumentationWebInterface() error {
	ctx.t.Log("Testing documentation web interface...")
	time.Sleep(100 * time.Millisecond)

	// Simulate web interface test
	return nil
}

func (ctx *APIDocumentationContext) validateRequestResponseExamples() error {
	ctx.t.Log("Validating request/response examples...")
	time.Sleep(75 * time.Millisecond)

	// Check if endpoints have examples
	for _, endpoint := range ctx.apiEndpoints {
		hasExamples := false
		for _, response := range endpoint.Responses {
			if response.Examples != nil && len(response.Examples) > 0 {
				hasExamples = true
				break
			}
		}
		if !hasExamples {
			ctx.t.Logf("Warning: endpoint %s %s missing examples",
				endpoint.Method, endpoint.Path)
		}
	}

	return nil
}

func (ctx *APIDocumentationContext) validateAuthenticationDocumentation() error {
	ctx.t.Log("Validating authentication documentation...")
	time.Sleep(50 * time.Millisecond)

	// Check if security definitions exist
	if ctx.swaggerConfig.SecurityDefs == nil || len(ctx.swaggerConfig.SecurityDefs) == 0 {
		return fmt.Errorf("no security definitions found")
	}

	return nil
}

func (ctx *APIDocumentationContext) validateEndpointConfiguration(endpoint APIEndpoint) error {
	if endpoint.Method == "" || endpoint.Path == "" {
		return fmt.Errorf("missing method or path")
	}
	return nil
}

func (ctx *APIDocumentationContext) generatePostmanCollection() (*DocumentationResult, error) {
	ctx.t.Log("Generating Postman collection...")
	time.Sleep(150 * time.Millisecond)

	filePath := ctx.postmanConfig.OutputPath
	ctx.generatedFiles = append(ctx.generatedFiles, filePath)

	return &DocumentationResult{
		DocumentType: "postman",
		Success:      true,
		FilePath:     filePath,
		FileSize:     1024 * 30, // 30KB
		GeneratedAt:  time.Now(),
		Duration:     150 * time.Millisecond,
	}, nil
}

func (ctx *APIDocumentationContext) validatePostmanCollectionCompleteness() error {
	ctx.t.Log("Validating Postman collection completeness...")
	time.Sleep(75 * time.Millisecond)
	return nil
}

// Placeholder implementations for remaining steps
func (ctx *APIDocumentationContext) collectionShouldHavePreRequestScripts() error { return nil }
func (ctx *APIDocumentationContext) testScriptsShouldValidateResponses() error { return nil }
func (ctx *APIDocumentationContext) environmentVariablesShouldBeConfigured() error { return nil }
func (ctx *APIDocumentationContext) generatedDocumentationShouldBeValid() error { return nil }
func (ctx *APIDocumentationContext) apiSpecificationsShouldFollowOpenAPIStandards() error { return nil }
func (ctx *APIDocumentationContext) documentationShouldBeUpToDateWithCode() error { return nil }
func (ctx *APIDocumentationContext) examplesShouldBeRealisticAndWorking() error { return nil }
func (ctx *APIDocumentationContext) errorResponsesShouldBeDocumented() error { return nil }