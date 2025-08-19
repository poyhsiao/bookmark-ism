// BDD Step Definitions for Production Infrastructure
// Task 26: Production deployment infrastructure with container orchestration

package step_definitions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// ProductionInfrastructureContext holds the context for production infrastructure tests
type ProductionInfrastructureContext struct {
	t                    *testing.T
	projectRoot          string
	composeFile          string
	swarmComposeFile     string
	deploymentScriptPath string
	pipelineScriptPath   string
	stackName            string
	deploymentStatus     string
	servicesStatus       map[string]string
	healthCheckResults   map[string]bool
	scalingResults       map[string]int
	errorMessages        []string
}

// NewProductionInfrastructureContext creates a new context for production infrastructure tests
func NewProductionInfrastructureContext(t *testing.T) *ProductionInfrastructureContext {
	projectRoot, _ := os.Getwd()
	return &ProductionInfrastructureContext{
		t:                    t,
		projectRoot:          projectRoot,
		composeFile:          "docker-compose.prod.yml",
		swarmComposeFile:     "docker-compose.swarm.yml",
		deploymentScriptPath: "scripts/deploy-swarm.sh",
		pipelineScriptPath:   "scripts/ci-cd-pipeline.sh",
		stackName:            "bookmark-sync-test",
		servicesStatus:       make(map[string]string),
		healthCheckResults:   make(map[string]bool),
		scalingResults:       make(map[string]int),
		errorMessages:        make([]string, 0),
	}
}

// Background step definitions

func (ctx *ProductionInfrastructureContext) theBookmarkSyncServiceIsFullyDevelopedAndTested() error {
	// Verify that all previous tasks are completed
	requiredFiles := []string{
		"backend/internal/auth/service.go",
		"backend/internal/bookmark/service.go",
		"backend/internal/collection/service.go",
		"backend/internal/sync/service.go",
		"backend/internal/automation/service.go",
		"extensions/chrome/manifest.json",
		"extensions/firefox/manifest.json",
		"extensions/safari/manifest.json",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(ctx.projectRoot, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required file not found: %s", file)
		}
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) allPreviousTasksAreCompletedSuccessfully() error {
	// Run tests to verify all previous tasks work correctly
	cmd := exec.Command("go", "test", "./backend/...", "-v")
	cmd.Dir = ctx.projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tests failed: %s", string(output))
	}

	// Check if all tests passed
	if !strings.Contains(string(output), "PASS") {
		return fmt.Errorf("not all tests passed: %s", string(output))
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) iHaveAccessToProductionEnvironmentResources() error {
	// Check if Docker is available
	if err := exec.Command("docker", "version").Run(); err != nil {
		return fmt.Errorf("Docker is not available: %v", err)
	}

	// Check if Docker Compose is available
	if err := exec.Command("docker", "compose", "version").Run(); err != nil {
		return fmt.Errorf("Docker Compose is not available: %v", err)
	}

	return nil
}

// Scenario 1: Create production Docker Compose configuration

func (ctx *ProductionInfrastructureContext) iNeedToDeployTheServiceInProduction() error {
	// This is a given condition - no action needed
	return nil
}

func (ctx *ProductionInfrastructureContext) iCreateAProductionDockerComposeConfiguration() error {
	// Verify that production compose files exist
	composeFiles := []string{
		ctx.composeFile,
		ctx.swarmComposeFile,
	}

	for _, file := range composeFiles {
		filePath := filepath.Join(ctx.projectRoot, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("production compose file not found: %s", file)
		}
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldIncludeOptimizedSettingsForAllServices() error {
	// Read and validate the swarm compose file
	filePath := filepath.Join(ctx.projectRoot, ctx.swarmComposeFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %v", err)
	}

	composeContent := string(content)

	// Check for production optimizations
	requiredOptimizations := []string{
		"resources:",
		"limits:",
		"reservations:",
		"healthcheck:",
		"restart_policy:",
		"update_config:",
	}

	for _, optimization := range requiredOptimizations {
		if !strings.Contains(composeContent, optimization) {
			return fmt.Errorf("missing production optimization: %s", optimization)
		}
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldSupportHorizontalScalingForGoBackendServices() error {
	// Check if API service has scaling configuration
	filePath := filepath.Join(ctx.projectRoot, ctx.swarmComposeFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %v", err)
	}

	composeContent := string(content)

	// Check for scaling configurations
	scalingConfigs := []string{
		"replicas:",
		"placement:",
		"deploy:",
	}

	for _, config := range scalingConfigs {
		if !strings.Contains(composeContent, config) {
			return fmt.Errorf("missing scaling configuration: %s", config)
		}
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldIncludeResourceLimitsAndHealthChecks() error {
	// Verify resource limits and health checks are present
	filePath := filepath.Join(ctx.projectRoot, ctx.swarmComposeFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %v", err)
	}

	composeContent := string(content)

	// Check for resource limits
	if !strings.Contains(composeContent, "memory:") || !strings.Contains(composeContent, "cpus:") {
		return fmt.Errorf("missing resource limits configuration")
	}

	// Check for health checks
	if !strings.Contains(composeContent, "healthcheck:") || !strings.Contains(composeContent, "test:") {
		return fmt.Errorf("missing health check configuration")
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldUseProductionGradeSecuritySettings() error {
	// Check for security configurations
	filePath := filepath.Join(ctx.projectRoot, ctx.swarmComposeFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %v", err)
	}

	composeContent := string(content)

	// Check for security settings
	securitySettings := []string{
		"secrets:",
		"encrypted:",
		"ssl",
		"tls",
	}

	foundSettings := 0
	for _, setting := range securitySettings {
		if strings.Contains(strings.ToLower(composeContent), strings.ToLower(setting)) {
			foundSettings++
		}
	}

	if foundSettings < 2 {
		return fmt.Errorf("insufficient security settings found")
	}

	return nil
}

// Scenario 2: Implement container orchestration with Docker Swarm

func (ctx *ProductionInfrastructureContext) iHaveAProductionDockerComposeConfiguration() error {
	return ctx.iCreateAProductionDockerComposeConfiguration()
}

func (ctx *ProductionInfrastructureContext) iSetUpDockerSwarmOrchestration() error {
	// Check if deployment script exists
	scriptPath := filepath.Join(ctx.projectRoot, ctx.deploymentScriptPath)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("deployment script not found: %s", scriptPath)
	}

	// Verify script is executable
	info, err := os.Stat(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to check script permissions: %v", err)
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("deployment script is not executable")
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldSupportMultiNodeDeployment() error {
	// Check if compose file has swarm-specific configurations
	filePath := filepath.Join(ctx.projectRoot, ctx.swarmComposeFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %v", err)
	}

	composeContent := string(content)

	// Check for swarm-specific configurations
	swarmConfigs := []string{
		"deploy:",
		"placement:",
		"constraints:",
		"preferences:",
	}

	for _, config := range swarmConfigs {
		if !strings.Contains(composeContent, config) {
			return fmt.Errorf("missing swarm configuration: %s", config)
		}
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldProvideAutomaticFailoverCapabilities() error {
	// Check for failover configurations
	filePath := filepath.Join(ctx.projectRoot, ctx.swarmComposeFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %v", err)
	}

	composeContent := string(content)

	// Check for failover settings
	failoverSettings := []string{
		"restart_policy:",
		"condition: on-failure",
		"max_attempts:",
		"failure_action: rollback",
	}

	foundSettings := 0
	for _, setting := range failoverSettings {
		if strings.Contains(composeContent, setting) {
			foundSettings++
		}
	}

	if foundSettings < 2 {
		return fmt.Errorf("insufficient failover configurations found")
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldEnableServiceScalingAcrossNodes() error {
	// Check for scaling configurations across nodes
	filePath := filepath.Join(ctx.projectRoot, ctx.swarmComposeFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %v", err)
	}

	composeContent := string(content)

	// Check for node distribution settings
	distributionSettings := []string{
		"spread:",
		"node.labels",
		"replicas:",
	}

	for _, setting := range distributionSettings {
		if !strings.Contains(composeContent, setting) {
			return fmt.Errorf("missing node distribution setting: %s", setting)
		}
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldMaintainServiceAvailabilityDuringUpdates() error {
	// Check for rolling update configurations
	filePath := filepath.Join(ctx.projectRoot, ctx.swarmComposeFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %v", err)
	}

	composeContent := string(content)

	// Check for rolling update settings
	updateSettings := []string{
		"update_config:",
		"parallelism:",
		"delay:",
		"order: start-first",
	}

	foundSettings := 0
	for _, setting := range updateSettings {
		if strings.Contains(composeContent, setting) {
			foundSettings++
		}
	}

	if foundSettings < 3 {
		return fmt.Errorf("insufficient rolling update configurations found")
	}

	return nil
}

// Scenario 3: Create automated deployment pipeline

func (ctx *ProductionInfrastructureContext) iHaveDockerSwarmOrchestrationConfigured() error {
	return ctx.iSetUpDockerSwarmOrchestration()
}

func (ctx *ProductionInfrastructureContext) iCreateAnAutomatedDeploymentPipeline() error {
	// Check if CI/CD pipeline script exists
	scriptPath := filepath.Join(ctx.projectRoot, ctx.pipelineScriptPath)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("CI/CD pipeline script not found: %s", scriptPath)
	}

	// Verify script is executable
	info, err := os.Stat(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to check script permissions: %v", err)
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("CI/CD pipeline script is not executable")
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldSupportCICDIntegration() error {
	// Check if pipeline script has CI/CD stages
	scriptPath := filepath.Join(ctx.projectRoot, ctx.pipelineScriptPath)
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read pipeline script: %v", err)
	}

	scriptContent := string(content)

	// Check for CI/CD stages
	cicdStages := []string{
		"stage_checkout",
		"stage_test",
		"stage_build",
		"stage_deploy",
	}

	for _, stage := range cicdStages {
		if !strings.Contains(scriptContent, stage) {
			return fmt.Errorf("missing CI/CD stage: %s", stage)
		}
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldIncludeAutomatedTestingBeforeDeployment() error {
	// Check if pipeline includes testing stages
	scriptPath := filepath.Join(ctx.projectRoot, ctx.pipelineScriptPath)
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read pipeline script: %v", err)
	}

	scriptContent := string(content)

	// Check for testing stages
	testingStages := []string{
		"stage_test",
		"go test",
		"smoke_tests",
		"security_scan",
	}

	foundStages := 0
	for _, stage := range testingStages {
		if strings.Contains(scriptContent, stage) {
			foundStages++
		}
	}

	if foundStages < 2 {
		return fmt.Errorf("insufficient testing stages found")
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldSupportRollingUpdatesWithZeroDowntime() error {
	// Check for zero-downtime deployment configurations
	scriptPath := filepath.Join(ctx.projectRoot, ctx.pipelineScriptPath)
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read pipeline script: %v", err)
	}

	scriptContent := string(content)

	// Check for rolling update configurations
	rollingUpdateConfigs := []string{
		"docker service update",
		"--update-parallelism",
		"--update-delay",
		"--update-failure-action rollback",
	}

	foundConfigs := 0
	for _, config := range rollingUpdateConfigs {
		if strings.Contains(scriptContent, config) {
			foundConfigs++
		}
	}

	if foundConfigs < 3 {
		return fmt.Errorf("insufficient rolling update configurations found")
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) itShouldProvideRollbackCapabilities() error {
	// Check for rollback functionality
	scriptPath := filepath.Join(ctx.projectRoot, ctx.pipelineScriptPath)
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read pipeline script: %v", err)
	}

	scriptContent := string(content)

	// Check for rollback configurations
	rollbackConfigs := []string{
		"rollback_deployment",
		"--rollback",
		"PREVIOUS_VERSION",
	}

	foundConfigs := 0
	for _, config := range rollbackConfigs {
		if strings.Contains(scriptContent, config) {
			foundConfigs++
		}
	}

	if foundConfigs < 2 {
		return fmt.Errorf("insufficient rollback configurations found")
	}

	return nil
}

// Additional helper methods for testing

func (ctx *ProductionInfrastructureContext) validateDockerComposeFile(filePath string) error {
	// Validate compose file syntax
	cmd := exec.Command("docker", "compose", "-f", filePath, "config")
	cmd.Dir = ctx.projectRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compose file validation failed: %s", string(output))
	}

	return nil
}

func (ctx *ProductionInfrastructureContext) checkServiceHealth(serviceName string) error {
	// Simulate health check
	ctx.healthCheckResults[serviceName] = true
	return nil
}

func (ctx *ProductionInfrastructureContext) scaleService(serviceName string, replicas int) error {
	// Simulate service scaling
	ctx.scalingResults[serviceName] = replicas
	return nil
}

// Register step definitions
func (ctx *ProductionInfrastructureContext) RegisterSteps(s *godog.ScenarioContext) {
	// Background steps
	s.Step(`^the bookmark sync service is fully developed and tested$`, ctx.theBookmarkSyncServiceIsFullyDevelopedAndTested)
	s.Step(`^all previous tasks \(1-25\) are completed successfully$`, ctx.allPreviousTasksAreCompletedSuccessfully)
	s.Step(`^I have access to production environment resources$`, ctx.iHaveAccessToProductionEnvironmentResources)

	// Scenario 1 steps
	s.Step(`^I need to deploy the service in production$`, ctx.iNeedToDeployTheServiceInProduction)
	s.Step(`^I create a production Docker Compose configuration$`, ctx.iCreateAProductionDockerComposeConfiguration)
	s.Step(`^it should include optimized settings for all services$`, ctx.itShouldIncludeOptimizedSettingsForAllServices)
	s.Step(`^it should support horizontal scaling for Go backend services$`, ctx.itShouldSupportHorizontalScalingForGoBackendServices)
	s.Step(`^it should include resource limits and health checks$`, ctx.itShouldIncludeResourceLimitsAndHealthChecks)
	s.Step(`^it should use production-grade security settings$`, ctx.itShouldUseProductionGradeSecuritySettings)

	// Scenario 2 steps
	s.Step(`^I have a production Docker Compose configuration$`, ctx.iHaveAProductionDockerComposeConfiguration)
	s.Step(`^I set up Docker Swarm orchestration$`, ctx.iSetUpDockerSwarmOrchestration)
	s.Step(`^it should support multi-node deployment$`, ctx.itShouldSupportMultiNodeDeployment)
	s.Step(`^it should provide automatic failover capabilities$`, ctx.itShouldProvideAutomaticFailoverCapabilities)
	s.Step(`^it should enable service scaling across nodes$`, ctx.itShouldEnableServiceScalingAcrossNodes)
	s.Step(`^it should maintain service availability during updates$`, ctx.itShouldMaintainServiceAvailabilityDuringUpdates)

	// Scenario 3 steps
	s.Step(`^I have Docker Swarm orchestration configured$`, ctx.iHaveDockerSwarmOrchestrationConfigured)
	s.Step(`^I create an automated deployment pipeline$`, ctx.iCreateAnAutomatedDeploymentPipeline)
	s.Step(`^it should support CI/CD integration$`, ctx.itShouldSupportCICDIntegration)
	s.Step(`^it should include automated testing before deployment$`, ctx.itShouldIncludeAutomatedTestingBeforeDeployment)
	s.Step(`^it should support rolling updates with zero downtime$`, ctx.itShouldSupportRollingUpdatesWithZeroDowntime)
	s.Step(`^it should provide rollback capabilities$`, ctx.itShouldProvideRollbackCapabilities)
}
