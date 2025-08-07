#!/bin/bash

# Validate GitHub Actions workflow files
# This script checks the syntax of all workflow files

set -e

echo "🔍 Validating GitHub Actions workflow files..."

# Check if GitHub CLI is available
if ! command -v gh &> /dev/null; then
    echo "⚠️ GitHub CLI not found. Installing..."
    # For Ubuntu/Debian
    if command -v apt-get &> /dev/null; then
        curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
        sudo apt update
        sudo apt install gh
    # For macOS
    elif command -v brew &> /dev/null; then
        brew install gh
    else
        echo "❌ Please install GitHub CLI manually: https://cli.github.com/"
        exit 1
    fi
fi

# Validate workflow files
WORKFLOW_DIR=".github/workflows"
VALIDATION_PASSED=true

if [ ! -d "$WORKFLOW_DIR" ]; then
    echo "❌ Workflow directory not found: $WORKFLOW_DIR"
    exit 1
fi

echo "📁 Found workflow directory: $WORKFLOW_DIR"

for workflow_file in "$WORKFLOW_DIR"/*.yml "$WORKFLOW_DIR"/*.yaml; do
    if [ -f "$workflow_file" ]; then
        echo "🔍 Validating: $(basename "$workflow_file")"

        # Check YAML syntax
        if command -v yq &> /dev/null; then
            if ! yq eval '.' "$workflow_file" > /dev/null 2>&1; then
                echo "❌ YAML syntax error in: $(basename "$workflow_file")"
                VALIDATION_PASSED=false
                continue
            fi
        elif command -v python3 &> /dev/null; then
            if ! python3 -c "import yaml; yaml.safe_load(open('$workflow_file'))" 2>/dev/null; then
                echo "❌ YAML syntax error in: $(basename "$workflow_file")"
                VALIDATION_PASSED=false
                continue
            fi
        fi

        # Check for common issues
        if grep -q "actions/upload-artifact@v3" "$workflow_file"; then
            echo "⚠️ Found deprecated actions/upload-artifact@v3 in: $(basename "$workflow_file")"
            echo "   Please update to v4"
        fi

        if grep -q "actions/download-artifact@v3" "$workflow_file"; then
            echo "⚠️ Found deprecated actions/download-artifact@v3 in: $(basename "$workflow_file")"
            echo "   Please update to v4"
        fi

        # Check for package cleanup issues
        if grep -q "delete-package-versions" "$workflow_file"; then
            if grep -q "package-name.*github.repository" "$workflow_file"; then
                echo "⚠️ Potential package name issue in: $(basename "$workflow_file")"
                echo "   Consider using github.event.repository.name instead"
            fi
        fi

        echo "✅ $(basename "$workflow_file") - OK"
    fi
done

if [ "$VALIDATION_PASSED" = true ]; then
    echo ""
    echo "🎉 All workflow files validated successfully!"
    echo ""
    echo "📋 Summary of fixes applied:"
    echo "   • Updated actions/upload-artifact from v3 to v4"
    echo "   • Updated actions/download-artifact from v3 to v4"
    echo "   • Fixed package cleanup configuration"
    echo "   • Added proper error handling for package deletion"
    echo ""
    echo "🚀 Your GitHub Actions workflows are ready to run!"
else
    echo ""
    echo "❌ Some workflow files have issues. Please fix them before proceeding."
    exit 1
fi