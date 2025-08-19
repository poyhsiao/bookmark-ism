// BDD Test Runner for Production Deployment Features
// Task 26: Production deployment infrastructure with container orchestration

package production_deployment_test

import (
	"context"
	"flag"
	"fmt"
	"os"
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

// TestProductionDeploymentFeatures runs the BDD tests for production deployment
func TestProductionDeploymentFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			// Initialize production infrastructure context
			prodInfraCtx := step_definitions.NewProductionInfrastructureContext(t)
			prodInfraCtx.RegisterSteps(ctx)

			// Initialize Docker Swarm orchestration context
			swarmCtx := step_definitions.NewDockerSwarmOrchestrationContext(t)
			swarmCtx.RegisterSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/production-deployment"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
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
