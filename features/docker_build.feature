Feature: Docker Build for Go Application
  As a developer
  I want to build Docker images for my Go application
  So that I can deploy it consistently across environments

  Background:
    Given I have a Go application with proper module structure
    And I have a multi-stage Dockerfile for production builds
    And I have GitHub Actions configured for CI/CD

  Scenario: Build production Docker image successfully
    Given the Go module is properly configured
    And the backend directory contains the API server code
    And the Dockerfile.prod uses correct build context
    When I build the Docker image using GitHub Actions
    Then the build should complete successfully
    And the image should be optimized for production
    And the image should contain only the compiled binary

  Scenario: Handle build context correctly
    Given the project has a monorepo structure
    And the Go code is in the backend directory
    And the Dockerfile is in the root directory
    When the Docker build process starts
    Then it should copy the correct files from the build context
    And it should build the Go binary from the correct path
    And it should not fail with "no such file or directory" errors

  Scenario: Multi-stage build optimization
    Given I have a multi-stage Dockerfile
    When the build process runs
    Then it should use Go build cache for faster builds
    And it should create a minimal final image
    And it should use security best practices
    And it should run as non-root user

  Scenario: GitHub Actions integration
    Given I have a CD pipeline configured
    When a push to main branch occurs
    Then it should build the Docker image
    And it should push to the container registry
    And it should handle build failures gracefully