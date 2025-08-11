package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

type dockerBuildFeature struct {
	buildContext     string
	dockerfilePath   string
	buildOutput      string
	buildError       error
	imageBuilt       bool
	projectStructure map[string]bool
}

func (d *dockerBuildFeature) resetState(*godog.Scenario) {
	d.buildContext = ""
	d.dockerfilePath = ""
	d.buildOutput = ""
	d.buildError = nil
	d.imageBuilt = false
	d.projectStructure = make(map[string]bool)
}

func (d *dockerBuildFeature) iHaveAGoApplicationWithProperModuleStructure() error {
	// Check if go.mod exists
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found in project root")
	}

	// Check if backend directory exists
	if _, err := os.Stat("backend"); os.IsNotExist(err) {
		return fmt.Errorf("backend directory not found")
	}

	// Check if backend/cmd/api exists
	if _, err := os.Stat("backend/cmd/api"); os.IsNotExist(err) {
		return fmt.Errorf("backend/cmd/api directory not found")
	}

	d.projectStructure["go.mod"] = true
	d.projectStructure["backend"] = true
	d.projectStructure["backend/cmd/api"] = true

	return nil
}

func (d *dockerBuildFeature) iHaveAMultiStageDockerfileForProductionBuilds() error {
	if _, err := os.Stat("Dockerfile.prod"); os.IsNotExist(err) {
		return fmt.Errorf("Dockerfile.prod not found")
	}

	d.dockerfilePath = "Dockerfile.prod"
	d.projectStructure["Dockerfile.prod"] = true

	return nil
}

func (d *dockerBuildFeature) iHaveGitHubActionsConfiguredForCICD() error {
	if _, err := os.Stat(".github/workflows"); os.IsNotExist(err) {
		return fmt.Errorf(".github/workflows directory not found")
	}

	d.projectStructure[".github/workflows"] = true
	return nil
}

func (d *dockerBuildFeature) theGoModuleIsProperlyConfigured() error {
	// Read go.mod and verify it has the correct module name
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %v", err)
	}

	if !strings.Contains(string(content), "module bookmark-sync-service") {
		return fmt.Errorf("go.mod does not contain expected module name")
	}

	return nil
}

func (d *dockerBuildFeature) theBackendDirectoryContainsTheAPIServerCode() error {
	mainGoPath := "backend/cmd/api/main.go"
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		return fmt.Errorf("main.go not found at %s", mainGoPath)
	}

	// Verify main.go contains expected content
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %v", err)
	}

	if !strings.Contains(string(content), "package main") {
		return fmt.Errorf("main.go does not contain package main")
	}

	return nil
}

func (d *dockerBuildFeature) theDockerfileProdUsesCorrectBuildContext() error {
	content, err := os.ReadFile("Dockerfile.prod")
	if err != nil {
		return fmt.Errorf("failed to read Dockerfile.prod: %v", err)
	}

	dockerfileContent := string(content)

	// Check for correct COPY commands
	if !strings.Contains(dockerfileContent, "COPY go.mod go.sum ./") {
		return fmt.Errorf("Dockerfile.prod missing correct go.mod copy command")
	}

	if !strings.Contains(dockerfileContent, "COPY backend ./backend") {
		return fmt.Errorf("Dockerfile.prod missing correct backend copy command")
	}

	// Check for correct build command
	if !strings.Contains(dockerfileContent, "./backend/cmd/api") {
		return fmt.Errorf("Dockerfile.prod missing correct build path")
	}

	return nil
}

func (d *dockerBuildFeature) iBuildTheDockerImageUsingGitHubActions() error {
	// Simulate Docker build (in real scenario, this would be done by GitHub Actions)
	d.buildContext = "."

	// Test Docker build locally to verify it works
	cmd := exec.Command("docker", "build", "-f", "Dockerfile.prod", "-t", "test-build", ".")
	output, err := cmd.CombinedOutput()

	d.buildOutput = string(output)
	d.buildError = err

	if err == nil {
		d.imageBuilt = true
	}

	return nil
}

func (d *dockerBuildFeature) theBuildShouldCompleteSuccessfully() error {
	if d.buildError != nil {
		return fmt.Errorf("build failed: %v\nOutput: %s", d.buildError, d.buildOutput)
	}

	if !d.imageBuilt {
		return fmt.Errorf("image was not built successfully")
	}

	return nil
}

func (d *dockerBuildFeature) theImageShouldBeOptimizedForProduction() error {
	// Check if the image uses multi-stage build
	content, err := os.ReadFile("Dockerfile.prod")
	if err != nil {
		return err
	}

	dockerfileContent := string(content)

	// Should have multiple FROM statements
	fromCount := strings.Count(dockerfileContent, "FROM ")
	if fromCount < 2 {
		return fmt.Errorf("Dockerfile should use multi-stage build (found %d FROM statements)", fromCount)
	}

	// Should use distroless or minimal base image
	if !strings.Contains(dockerfileContent, "distroless") && !strings.Contains(dockerfileContent, "alpine") {
		return fmt.Errorf("Dockerfile should use minimal base image")
	}

	return nil
}

func (d *dockerBuildFeature) theImageShouldContainOnlyTheCompiledBinary() error {
	// This would typically be verified by inspecting the built image
	// For now, we check that the Dockerfile copies only the binary in the final stage
	content, err := os.ReadFile("Dockerfile.prod")
	if err != nil {
		return err
	}

	dockerfileContent := string(content)

	// Should copy binary from builder stage
	if !strings.Contains(dockerfileContent, "COPY --from=builder") {
		return fmt.Errorf("Dockerfile should copy binary from builder stage")
	}

	return nil
}

func (d *dockerBuildFeature) theProjectHasAMonorepoStructure() error {
	// Verify monorepo structure
	requiredDirs := []string{"backend", "web", "extensions"}
	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("required directory %s not found", dir)
		}
	}
	return nil
}

func (d *dockerBuildFeature) theGoCodeIsInTheBackendDirectory() error {
	return d.theBackendDirectoryContainsTheAPIServerCode()
}

func (d *dockerBuildFeature) theDockerfileIsInTheRootDirectory() error {
	if _, err := os.Stat("Dockerfile.prod"); os.IsNotExist(err) {
		return fmt.Errorf("Dockerfile.prod not found in root directory")
	}
	return nil
}

func (d *dockerBuildFeature) theDockerBuildProcessStarts() error {
	d.buildContext = "."
	return nil
}

func (d *dockerBuildFeature) itShouldCopyTheCorrectFilesFromTheBuildContext() error {
	// This is verified by checking the Dockerfile content
	return d.theDockerfileProdUsesCorrectBuildContext()
}

func (d *dockerBuildFeature) itShouldBuildTheGoBinaryFromTheCorrectPath() error {
	content, err := os.ReadFile("Dockerfile.prod")
	if err != nil {
		return err
	}

	if !strings.Contains(string(content), "./backend/cmd/api") {
		return fmt.Errorf("Dockerfile does not build from correct path")
	}

	return nil
}

func (d *dockerBuildFeature) itShouldNotFailWithNoSuchFileOrDirectoryErrors() error {
	if d.buildError != nil && strings.Contains(d.buildOutput, "no such file or directory") {
		return fmt.Errorf("build failed with file not found error: %s", d.buildOutput)
	}
	return nil
}

func (d *dockerBuildFeature) iHaveAMultiStageDockerfile() error {
	return d.iHaveAMultiStageDockerfileForProductionBuilds()
}

func (d *dockerBuildFeature) theBuildProcessRuns() error {
	return d.iBuildTheDockerImageUsingGitHubActions()
}

func (d *dockerBuildFeature) itShouldUseGoBuildCacheForFasterBuilds() error {
	content, err := os.ReadFile("Dockerfile.prod")
	if err != nil {
		return err
	}

	if !strings.Contains(string(content), "--mount=type=cache") {
		return fmt.Errorf("Dockerfile should use build cache mounts")
	}

	return nil
}

func (d *dockerBuildFeature) itShouldCreateAMinimalFinalImage() error {
	return d.theImageShouldBeOptimizedForProduction()
}

func (d *dockerBuildFeature) itShouldUseSecurityBestPractices() error {
	content, err := os.ReadFile("Dockerfile.prod")
	if err != nil {
		return err
	}

	dockerfileContent := string(content)

	// Should use non-root user
	if !strings.Contains(dockerfileContent, "USER nonroot") && !strings.Contains(dockerfileContent, "USER appuser") {
		return fmt.Errorf("Dockerfile should use non-root user")
	}

	return nil
}

func (d *dockerBuildFeature) itShouldRunAsNonRootUser() error {
	return d.itShouldUseSecurityBestPractices()
}

func (d *dockerBuildFeature) iHaveACDPipelineConfigured() error {
	return d.iHaveGitHubActionsConfiguredForCICD()
}

func (d *dockerBuildFeature) aPushToMainBranchOccurs() error {
	// This would be triggered by GitHub Actions
	return nil
}

func (d *dockerBuildFeature) itShouldBuildTheDockerImage() error {
	return d.iBuildTheDockerImageUsingGitHubActions()
}

func (d *dockerBuildFeature) itShouldPushToTheContainerRegistry() error {
	// This would be handled by GitHub Actions
	// We can verify the workflow configuration
	content, err := os.ReadFile(".github/workflows/cd.yml")
	if err != nil {
		return fmt.Errorf("CD workflow not found")
	}

	if !strings.Contains(string(content), "docker/build-push-action") {
		return fmt.Errorf("CD workflow should use build-push-action")
	}

	return nil
}

func (d *dockerBuildFeature) itShouldHandleBuildFailuresGracefully() error {
	// Check if the workflow has proper error handling
	content, err := os.ReadFile(".github/workflows/cd.yml")
	if err != nil {
		return err
	}

	if !strings.Contains(string(content), "if: failure()") {
		return fmt.Errorf("CD workflow should handle failures")
	}

	return nil
}

func TestDockerBuildFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeDockerBuildScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/docker_build.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeDockerBuildScenario(ctx *godog.ScenarioContext) {
	d := &dockerBuildFeature{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		d.resetState(sc)
		return ctx, nil
	})

	// Background steps
	ctx.Step(`^I have a Go application with proper module structure$`, d.iHaveAGoApplicationWithProperModuleStructure)
	ctx.Step(`^I have a multi-stage Dockerfile for production builds$`, d.iHaveAMultiStageDockerfileForProductionBuilds)
	ctx.Step(`^I have GitHub Actions configured for CI/CD$`, d.iHaveGitHubActionsConfiguredForCICD)

	// Scenario 1: Build production Docker image successfully
	ctx.Step(`^the Go module is properly configured$`, d.theGoModuleIsProperlyConfigured)
	ctx.Step(`^the backend directory contains the API server code$`, d.theBackendDirectoryContainsTheAPIServerCode)
	ctx.Step(`^the Dockerfile\.prod uses correct build context$`, d.theDockerfileProdUsesCorrectBuildContext)
	ctx.Step(`^I build the Docker image using GitHub Actions$`, d.iBuildTheDockerImageUsingGitHubActions)
	ctx.Step(`^the build should complete successfully$`, d.theBuildShouldCompleteSuccessfully)
	ctx.Step(`^the image should be optimized for production$`, d.theImageShouldBeOptimizedForProduction)
	ctx.Step(`^the image should contain only the compiled binary$`, d.theImageShouldContainOnlyTheCompiledBinary)

	// Scenario 2: Handle build context correctly
	ctx.Step(`^the project has a monorepo structure$`, d.theProjectHasAMonorepoStructure)
	ctx.Step(`^the Go code is in the backend directory$`, d.theGoCodeIsInTheBackendDirectory)
	ctx.Step(`^the Dockerfile is in the root directory$`, d.theDockerfileIsInTheRootDirectory)
	ctx.Step(`^the Docker build process starts$`, d.theDockerBuildProcessStarts)
	ctx.Step(`^it should copy the correct files from the build context$`, d.itShouldCopyTheCorrectFilesFromTheBuildContext)
	ctx.Step(`^it should build the Go binary from the correct path$`, d.itShouldBuildTheGoBinaryFromTheCorrectPath)
	ctx.Step(`^it should not fail with "no such file or directory" errors$`, d.itShouldNotFailWithNoSuchFileOrDirectoryErrors)

	// Scenario 3: Multi-stage build optimization
	ctx.Step(`^I have a multi-stage Dockerfile$`, d.iHaveAMultiStageDockerfile)
	ctx.Step(`^the build process runs$`, d.theBuildProcessRuns)
	ctx.Step(`^it should use Go build cache for faster builds$`, d.itShouldUseGoBuildCacheForFasterBuilds)
	ctx.Step(`^it should create a minimal final image$`, d.itShouldCreateAMinimalFinalImage)
	ctx.Step(`^it should use security best practices$`, d.itShouldUseSecurityBestPractices)
	ctx.Step(`^it should run as non-root user$`, d.itShouldRunAsNonRootUser)

	// Scenario 4: GitHub Actions integration
	ctx.Step(`^I have a CD pipeline configured$`, d.iHaveACDPipelineConfigured)
	ctx.Step(`^a push to main branch occurs$`, d.aPushToMainBranchOccurs)
	ctx.Step(`^it should build the Docker image$`, d.itShouldBuildTheDockerImage)
	ctx.Step(`^it should push to the container registry$`, d.itShouldPushToTheContainerRegistry)
	ctx.Step(`^it should handle build failures gracefully$`, d.itShouldHandleBuildFailuresGracefully)
}
