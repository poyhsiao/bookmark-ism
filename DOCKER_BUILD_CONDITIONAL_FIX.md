# Docker Build Conditional Logic Fix

## Issue Addressed

**Bug Risk**: Setting a placeholder digest may cause downstream steps to behave unexpectedly.

## Root Cause

The original implementation used a placeholder digest (`digest=sha256:placeholder`) when Docker build was skipped, which could cause downstream steps that depend on the build outputs to behave unexpectedly.

## Solution Implemented

### 1. **Proper Output Handling**

**Before (Problematic)**:
```yaml
echo "digest=sha256:placeholder" >> $GITHUB_OUTPUT
```

**After (Fixed)**:
```yaml
# Set outputs to indicate build was skipped
echo "skipped=true" >> $GITHUB_OUTPUT
echo "digest=" >> $GITHUB_OUTPUT
```

### 2. **Conditional Step Execution**

Added proper conditional logic to prevent downstream steps from running when build is skipped:

```yaml
# All deployment jobs now check for skipped condition
if: github.ref == 'refs/heads/main' && vars.ENABLE_AWS_DEPLOYMENT == 'true' && needs.build-and-push.outputs.skipped != 'true'
```

### 3. **Dedicated Skipped Build Handling**

Added a specific job to handle the skipped build case:

```yaml
docker-build-skipped:
  name: Docker Build Skipped
  runs-on: ubuntu-latest
  needs: build-and-push
  if: needs.build-and-push.outputs.skipped == 'true'
```

### 4. **SBOM Generation Conditional Logic**

```yaml
- name: Generate SBOM
  if: steps.build.outputs.skipped != 'true'
  # ... normal SBOM generation

- name: Skip SBOM generation
  if: steps.build.outputs.skipped == 'true'
  # ... create placeholder SBOM
```

### 5. **Enhanced Notification Logic**

Updated notification job to handle three scenarios:
- **Success with build**: Normal successful pipeline
- **Success with skipped build**: Pipeline completed but Docker build was skipped
- **Failure**: Build or other steps failed

## Jobs Affected

### Modified Jobs
- `build-and-push`: Added `skipped` output
- `deploy-staging`: Added skipped condition check
- `deploy-production`: Added skipped condition check
- `skip-deployment`: Added skipped condition check
- `cleanup`: Added skipped condition check
- `notify-deployment`: Enhanced to handle skipped builds

### New Jobs
- `docker-build-skipped`: Dedicated handling for skipped builds

## Benefits

1. **No Unexpected Behavior**: Downstream steps won't receive invalid digest values
2. **Clear Status Reporting**: Explicit handling of skipped vs successful builds
3. **Proper Conditional Logic**: Steps only run when appropriate
4. **Comprehensive Notifications**: Users get clear feedback about pipeline status
5. **Safe Re-enabling**: Automated script to restore normal operation

## Re-enabling Process

The `scripts/re-enable-docker-builds.sh` script automatically:
1. Validates the Docker build fix locally
2. Removes all conditional logic for skipped builds
3. Restores normal Docker build and push operations
4. Updates job dependencies and outputs
5. Validates YAML syntax

## Testing

### Local Validation
```bash
./validate_docker_build.sh
```

### Re-enable Builds
```bash
./scripts/re-enable-docker-builds.sh
```

### Monitor Results
- GitHub Actions will show clear status for each scenario
- No more placeholder values causing unexpected behavior
- Proper conditional execution of all dependent steps

## Implementation Details

### Output Structure
```yaml
outputs:
  image-tag: ${{ steps.meta.outputs.tags }}
  image-digest: ${{ steps.build.outputs.digest }}
  skipped: ${{ steps.build.outputs.skipped }}
```

### Conditional Patterns
```yaml
# For steps that should skip when build is skipped
if: needs.build-and-push.outputs.skipped != 'true'

# For steps that should only run when build is skipped
if: needs.build-and-push.outputs.skipped == 'true'
```

This implementation follows GitHub Actions best practices for conditional workflow execution and ensures no downstream steps receive invalid or placeholder values.