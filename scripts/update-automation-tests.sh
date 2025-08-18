#!/bin/bash

# Script to update automation test files to use the new base class methods

echo "🔄 Updating automation test files to use shared test base..."

# Update service_test.go
echo "📝 Updating service_test.go..."
sed -i '' 's/suite\.service\./suite.GetTestService()./g' backend/internal/automation/service_test.go
sed -i '' 's/suite\.userID/suite.GetTestUserID()/g' backend/internal/automation/service_test.go

# Update handlers_test.go (remaining occurrences)
echo "📝 Updating handlers_test.go..."
sed -i '' 's/suite\.service\./suite.GetTestService()./g' backend/internal/automation/handlers_test.go
sed -i '' 's/suite\.userID/suite.GetTestUserID()/g' backend/internal/automation/handlers_test.go

echo "✅ Test files updated successfully!"
echo "🧪 Running tests to verify changes..."

# Run tests to verify the changes work
cd backend && go test ./internal/automation -v

if [ $? -eq 0 ]; then
    echo "✅ All tests pass after refactoring!"
else
    echo "❌ Tests failed after refactoring. Please check the changes."
    exit 1
fi