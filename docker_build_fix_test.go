package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// isDockerAvailable checks if Docker is available and running
func isDockerAvailable() bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}

	// Test if Docker daemon is running
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	return cmd.Run() == nil
}

// TestDockerBuildFix tests that the Docker build issue is resolved
func TestDockerBuildFix(t *testing.T) {
	tests := []struct {
		name             string
		dockerfile       string
		expectedPaths    []string
		shouldNotContain []string
	}{
		{
			name:       "Production Dockerfile build context",
			dockerfile: "Dockerfile.prod",
			expectedPaths: []string{
				"WORKDIR /build",
				"COPY go.mod go.sum ./",
				"COPY backend ./backend",
				"go build",
				"./backend/cmd/api",
				"COPY --from=builder /build/main /",
			},
			shouldNotContain: []string{
				"WORKDIR /app",
				"COPY --from=builder /app/main",
			},
		},
		{
			name:       "Development Dockerfile build context",
			dockerfile: "Dockerfile",
			expectedPaths: []string{
				"WORKDIR /build",
				"COPY go.mod go.sum ./",
				"COPY backend ./backend",
				"go build",
				"./backend/cmd/api",
				"COPY --from=builder /build/main .",
			},
			shouldNotContain: []string{
				"WORKDIR /app",
				"COPY --from=builder /app/main",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read Dockerfile content
			content, err := os.ReadFile(tt.dockerfile)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tt.dockerfile, err)
			}

			dockerfileContent := string(content)

			// Check for expected paths
			for _, expectedPath := range tt.expectedPaths {
				if !strings.Contains(dockerfileContent, expectedPath) {
					t.Errorf("Dockerfile %s should contain: %s", tt.dockerfile, expectedPath)
				}
			}

			// Check for paths that should not be present
			for _, shouldNotContain := range tt.shouldNotContain {
				if strings.Contains(dockerfileContent, shouldNotContain) {
					t.Errorf("Dockerfile %s should not contain: %s", tt.dockerfile, shouldNotContain)
				}
			}
		})
	}
}

// TestDockerBuildSyntax tests that the Dockerfile syntax is valid
func TestDockerBuildSyntax(t *testing.T) {
	dockerfiles := []string{"Dockerfile", "Dockerfile.prod"}

	for _, dockerfile := range dockerfiles {
		t.Run(dockerfile, func(t *testing.T) {
			// Check if Docker is available and running
			if !isDockerAvailable() {
				t.Skip("Docker not available or not running, skipping syntax test")
			}

			// Check for common syntax errors by reading the file
			content, err := os.ReadFile(dockerfile)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", dockerfile, err)
			}

			dockerfileContent := string(content)

			// Basic syntax checks
			if !strings.HasPrefix(dockerfileContent, "# syntax=docker/dockerfile:1") {
				t.Errorf("%s should start with syntax directive", dockerfile)
			}

			if !strings.Contains(dockerfileContent, "FROM ") {
				t.Errorf("%s should contain FROM instruction", dockerfile)
			}

			// Check for multi-stage build
			fromCount := strings.Count(dockerfileContent, "FROM ")
			if fromCount < 2 {
				t.Errorf("%s should use multi-stage build (found %d FROM statements)", dockerfile, fromCount)
			}

			// Test actual Docker build syntax by attempting a build with a minimal context
			// This will fail if there are syntax errors in the Dockerfile
			cmd := exec.Command("docker", "build", "-f", dockerfile, "--no-cache", "--target", "builder", ".")
			output, err := cmd.CombinedOutput()

			// Check for syntax errors in the output
			if err != nil && strings.Contains(string(output), "dockerfile parse error") {
				t.Errorf("Dockerfile %s has syntax errors: %s", dockerfile, string(output))
			}
		})
	}
}

// TestProjectStructure verifies the project structure is correct for Docker builds
func TestProjectStructure(t *testing.T) {
	requiredFiles := []string{
		"go.mod",
		"go.sum",
		"backend/cmd/api/main.go",
		"Dockerfile",
		"Dockerfile.prod",
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Required file %s does not exist", file)
		}
	}

	// Check go.mod content
	content, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	if !strings.Contains(string(content), "module bookmark-sync-service") {
		t.Error("go.mod should contain correct module name")
	}

	// Check main.go content
	mainContent, err := os.ReadFile("backend/cmd/api/main.go")
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	if !strings.Contains(string(mainContent), "package main") {
		t.Error("main.go should contain package main")
	}

	if !strings.Contains(string(mainContent), "func main()") {
		t.Error("main.go should contain main function")
	}
}

// TestGitHubActionsWorkflow verifies the GitHub Actions workflow is correctly configured
func TestGitHubActionsWorkflow(t *testing.T) {
	workflowFile := ".github/workflows/cd.yml"

	content, err := os.ReadFile(workflowFile)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", workflowFile, err)
	}

	workflowContent := string(content)

	// Check for correct Docker build configuration
	expectedConfigs := []string{
		"docker/build-push-action@v5",
		"context: .",
		"file: ./Dockerfile.prod",
		"push: true",
	}

	for _, config := range expectedConfigs {
		if !strings.Contains(workflowContent, config) {
			t.Errorf("GitHub Actions workflow should contain: %s", config)
		}
	}

	// Check that it's using the correct Dockerfile
	if !strings.Contains(workflowContent, "Dockerfile.prod") {
		t.Error("GitHub Actions should use Dockerfile.prod for production builds")
	}
}
