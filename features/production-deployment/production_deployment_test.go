// BDD Test Runner for Production Deployment Features
// Task 26: Production deployment infrastructure with container orchestration

package production_deployment_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"bookmark-sync-service/features/production-deployment/step_definitions"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress", // can be changed to "pretty" for more detailed output
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opts.Paths = []string{"features/production-deployment"}

	status := godog.TestSuite{
		Name:                "Production Deployment",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	// Optional: print additional information about the test run
	if status == 2 {
		fmt.Println("Random test execution order.")
	}

	os.Exit(status)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	// Create a test instance for each scenario
	t := &testing.T{}

	// Initialize production infrastructure context
	prodInfraCtx := step_definitions.NewProductionInfrastructureContext(t)
	prodInfraCtx.RegisterSteps(ctx)

	// Initialize Docker Swarm orchestration context
	swarmCtx := step_definitions.NewDockerSwarmOrchestrationContext(t)
	swarmCtx.RegisterSteps(ctx)

	// Add hooks for setup and teardown
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Setup before each scenario
		fmt.Printf("Starting scenario: %s\n", sc.Name)
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// Cleanup after each scenario
		if err != nil {
			fmt.Printf("Scenario failed: %s - %v\n", sc.Name, err)
		} else {
			fmt.Printf("Scenario passed: %s\n", sc.Name)
		}
		return ctx, nil
	})
}

// getAllFeatureScenarios parses all feature files and returns scenario names
func getAllFeatureScenarios(featureDir string) (map[string]struct{}, error) {
	scenarios := make(map[string]struct{})

	err := filepath.Walk(featureDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".feature") {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			gherkinDocument, err := io.ParseGherkinDocument(gio.NewReader(f))
			if err != nil {
				return err
			}

			if gherkinDocument.Feature != nil {
				for _, child := range gherkinDocument.Feature.Children {
					if child.Scenario != nil {
						scenarios[child.Scenario.Name] = struct{}{}
					}
				}
			}
		}
		return nil
	})

	return scenarios, err
}

// TestProductionDeploymentFeatures runs the BDD tests for production deployment
func TestProductionDeploymentFeatures(t *testing.T) {
	featureDir := "features/production-deployment"

	// Get all scenarios from feature files
	allScenarios, err := getAllFeatureScenarios(featureDir)
	if err != nil {
		t.Fatalf("Failed to parse feature files: %v", err)
	}

	executedScenarios := make(map[string]struct{})

	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			// Initialize production infrastructure context
			prodInfraCtx := step_definitions.NewProductionInfrastructureContext(t)
			prodInfraCtx.RegisterSteps(ctx)

			// Initialize Docker Swarm orchestration context
			swarmCtx := step_definitions.NewDockerSwarmOrchestrationContext(t)
			swarmCtx.RegisterSteps(ctx)

			// Track executed scenarios
			ctx.BeforeScenario(func(sc *godog.Scenario) {
				executedScenarios[sc.Name] = struct{}{}
			})
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{featureDir},
			TestingT: t,
		},
	}

	status := suite.Run()

	// After test run, check for unexecuted scenarios
	missingScenarios := []string{}
	for scenario := range allScenarios {
		if _, ok := executedScenarios[scenario]; !ok {
			missingScenarios = append(missingScenarios, scenario)
		}
	}

	if len(missingScenarios) > 0 {
		t.Errorf("The following scenarios were not executed (possibly due to missing step definitions):\n%s",
			strings.Join(missingScenarios, "\n"))
	}

	if status != 0 {
		t.Fail()
	}
}

// TestProductionInfrastructure tests the production infrastructure scenarios
func TestProductionInfrastructure(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			prodInfraCtx := step_definitions.NewProductionInfrastructureContext(t)
			prodInfraCtx.RegisterSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/production-deployment/production-infrastructure.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("production infrastructure tests failed")
	}
}

// TestDockerSwarmOrchestration tests the Docker Swarm orchestration scenarios
func TestDockerSwarmOrchestration(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			swarmCtx := step_definitions.NewDockerSwarmOrchestrationContext(t)
			swarmCtx.RegisterSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/production-deployment/docker-swarm-orchestration.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("Docker Swarm orchestration tests failed")
	}
}
