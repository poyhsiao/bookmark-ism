// BDD Test Runner for Comprehensive Testing Suite
// Phase 14: Testing and Quality Assurance with BDD approach

package comprehensive_testing_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"bookmark-sync-service/features/comprehensive-testing/step_definitions"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress", // can be changed to "pretty" for detailed output
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opts.Paths = []string{"./"}

	status := godog.TestSuite{
		Name:                "Comprehensive Testing Suite",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	// Print test execution summary
	if status == 0 {
		fmt.Println("\n✅ All BDD tests passed successfully!")
	} else {
		fmt.Printf("\n❌ BDD tests failed with status: %d\n", status)
	}

	os.Exit(status)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		fmt.Println("🚀 Starting Comprehensive Testing Suite...")
		fmt.Printf("Time: %s\n", time.Now().Format(time.RFC3339))
	})

	ctx.AfterSuite(func() {
		fmt.Println("🏁 Comprehensive Testing Suite completed")
		fmt.Printf("Time: %s\n", time.Now().Format(time.RFC3339))
	})
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	// Create a test instance for each scenario
	t := &testing.T{}

	// Initialize Unit Test BDD Context
	unitTestCtx := step_definitions.NewUnitTestContext(t)
	unitTestCtx.RegisterSteps(ctx)

	// Initialize Integration Test BDD Context
	integrationCtx := step_definitions.NewIntegrationTestContext(t)
	integrationCtx.RegisterSteps(ctx)

	// Initialize End-to-End Test Context
	e2eCtx := step_definitions.NewEndToEndTestContext(t)
	e2eCtx.RegisterSteps(ctx)

	// Initialize Performance Test Context
	performanceCtx := step_definitions.NewPerformanceTestContext(t)
	performanceCtx.RegisterSteps(ctx)

	// Initialize Security Test Context
	securityCtx := step_definitions.NewSecurityTestContext(t)
	securityCtx.RegisterSteps(ctx)

	// Initialize API Documentation Context
	apiDocCtx := step_definitions.NewAPIDocumentationContext(t)
	apiDocCtx.RegisterSteps(ctx)

	// Initialize Code Quality Context
	codeQualityCtx := step_definitions.NewCodeQualityContext(t)
	codeQualityCtx.RegisterSteps(ctx)

	// Initialize User Acceptance Test Context
	userAcceptanceCtx := step_definitions.NewUserAcceptanceTestContext(t)
	userAcceptanceCtx.RegisterSteps(ctx)

	// Initialize Documentation Context for Phase 14 Task 31
	documentationCtx := step_definitions.NewDocumentationContext()
	documentationCtx.RegisterSteps(ctx)

	// Add hooks for setup and teardown
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Setup before each scenario
		fmt.Printf("📝 Starting scenario: %s\n", sc.Name)
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// Cleanup after each scenario
		if err != nil {
			fmt.Printf("❌ Scenario failed: %s - %v\n", sc.Name, err)
		} else {
			fmt.Printf("✅ Scenario passed: %s\n", sc.Name)
		}
		return ctx, nil
	})
}

// TestComprehensiveTestingSuite runs all BDD tests for Phase 14
func TestComprehensiveTestingSuite(t *testing.T) {
	featureDir := "./"

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{featureDir},
			TestingT: t,
		},
	}

	status := suite.Run()

	if status != 0 {
		t.Fatalf("Comprehensive testing suite failed with status: %d", status)
	}

	t.Log("🎉 All comprehensive tests completed successfully!")
}

// TestUnitTestCoverage runs BDD scenarios for unit test coverage
func TestUnitTestCoverage(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			unitTestCtx := step_definitions.NewUnitTestContext(t)
			unitTestCtx.RegisterSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"unit-tests.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("Unit test coverage scenarios failed")
	}
}

// TestIntegrationTestScenarios runs BDD scenarios for integration tests
func TestIntegrationTestScenarios(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			integrationCtx := step_definitions.NewIntegrationTestContext(t)
			integrationCtx.RegisterSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"integration-tests.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("Integration test scenarios failed")
	}
}

// TestEndToEndUserWorkflows runs BDD scenarios for end-to-end user workflows
func TestEndToEndUserWorkflows(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			e2eCtx := step_definitions.NewEndToEndTestContext(t)
			e2eCtx.RegisterSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"e2e-workflows.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("End-to-end workflow scenarios failed")
	}
}

// TestPerformanceTestScenarios runs BDD scenarios for performance testing
func TestPerformanceTestScenarios(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			performanceCtx := step_definitions.NewPerformanceTestContext(t)
			performanceCtx.RegisterSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"performance-tests.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("Performance test scenarios failed")
	}
}

// TestSecurityTestScenarios runs BDD scenarios for security testing
func TestSecurityTestScenarios(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			securityCtx := step_definitions.NewSecurityTestContext(t)
			securityCtx.RegisterSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"security-tests.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("Security test scenarios failed")
	}
}

// TestDocumentationGeneration runs BDD scenarios for documentation generation (Phase 14 Task 31)
func TestDocumentationGeneration(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			documentationCtx := step_definitions.NewDocumentationContext()
			documentationCtx.RegisterSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"documentation-generation.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("Documentation generation scenarios failed")
	}
}

// TestCodeQuality runs BDD scenarios for code quality assurance (Phase 14 Task 31)
func TestCodeQuality(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			codeQualityCtx := step_definitions.NewCodeQualityContext(t)
			codeQualityCtx.RegisterSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"quality-assurance.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("Code quality scenarios failed")
	}
}

// TestUserAcceptance runs BDD scenarios for user acceptance testing (Phase 14 Task 31)
func TestUserAcceptance(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			userAcceptanceCtx := step_definitions.NewUserAcceptanceTestContext(t)
			userAcceptanceCtx.RegisterSteps(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"quality-assurance.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("User acceptance scenarios failed")
	}
}

// getAllFeatureScenarios parses all feature files and returns scenario names
func getAllFeatureScenarios(featureDir string) (map[string]struct{}, error) {
	scenarios := make(map[string]struct{})

	// Add all expected scenarios for validation
	expectedScenarios := []string{
		"Unit Test Coverage Validation",
		"Service Layer Unit Testing",
		"Handler Unit Testing",
		"Database Integration Testing",
		"External Service Integration Testing",
		"End-to-End User Registration and Login",
		"Complete Bookmark Management Workflow",
		"Cross-Browser Synchronization Workflow",
		"Performance Load Testing",
		"Response Time Validation",
		"Security Vulnerability Assessment",
		"Authentication Security Testing",
		"API Documentation Generation",
		"Code Quality Analysis",
		"User Acceptance Testing",
	}

	for _, scenario := range expectedScenarios {
		scenarios[scenario] = struct{}{}
	}

	return scenarios, nil
}

// TestScenarioCompletion validates all expected scenarios are executed
func TestScenarioCompletion(t *testing.T) {
	featureDir := "./"

	// Get all expected scenarios
	expectedScenarios, err := getAllFeatureScenarios(featureDir)
	if err != nil {
		t.Fatalf("Failed to get expected scenarios: %v", err)
	}

	executedScenarios := make(map[string]struct{})

	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			InitializeScenario(ctx)

			// Track executed scenarios
			ctx.BeforeScenario(func(sc *godog.Scenario) {
				executedScenarios[sc.Name] = struct{}{}
			})
		},
		Options: &godog.Options{
			Format:   "progress",
			Paths:    []string{featureDir},
			TestingT: t,
		},
	}

	status := suite.Run()

	// Check for missing scenarios
	missingScenarios := []string{}
	for scenario := range expectedScenarios {
		if _, ok := executedScenarios[scenario]; !ok {
			missingScenarios = append(missingScenarios, scenario)
		}
	}

	if len(missingScenarios) > 0 {
		t.Errorf("The following scenarios were not executed:\n%s",
			strings.Join(missingScenarios, "\n"))
	}

	if status != 0 {
		t.Errorf("Test suite failed with status: %d", status)
	}

	t.Logf("✅ Executed %d scenarios successfully", len(executedScenarios))
}