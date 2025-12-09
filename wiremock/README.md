# WireMock Inventory Service Mock

This directory contains WireMock stub mappings for mocking the inventory service API used by the Temporal order processing workflow's Validate activity.

## Architecture

The setup uses a hybrid approach:
- **Default mappings** in `mappings/` are loaded automatically when WireMock starts
- **Scenario-specific mappings** in `scenarios/` can be added dynamically via the admin API
- A helper script (`scenarios.sh`) manages switching between scenarios

## Directory Structure

```
wiremock/
├── mappings/
│   └── inventory-success.json                        # Default success scenario (always loaded)
├── scenarios/
│   ├── inventory-intermittent-failure.json           # Intermittent failure - first attempt
│   ├── inventory-intermittent-failure-recovery.json  # Intermittent failure - recovery
│   └── inventory-non-retryable-failure.json          # Non-retryable error
└── scenarios.sh                                      # Helper script to manage scenarios
```

## Scenarios

### 1. Success Scenario (Default)
- **File**: `mappings/inventory-success.json`
- **Status**: 200 OK
- **Response**: `{"available": true, "message": "Product is in stock"}`
- **Use Case**: Tests successful inventory checks when products are in stock
- **Priority**: 1 (default fallback)
- **Loaded**: Automatically on startup

### 2. Intermittent Failure Scenario (Retryable)
- **Files**:
  - `scenarios/inventory-intermittent-failure.json`
  - `scenarios/inventory-intermittent-failure-recovery.json`
- **Status**: First attempt returns 503, second attempt returns 200
- **Response**:
  - First: `{"available": false, "message": "Service temporarily unavailable"}`
  - Second: `{"available": true, "message": "Product is in stock (recovered after retry)"}`
- **Use Case**: Tests Temporal activity retry mechanism with temporary failures
- **Scenario**: Uses WireMock's scenario state feature to simulate recovery after retry
- **Priority**: 0 (highest - overrides default when scenario state matches)
- **Loaded**: Dynamically via admin API or helper script

### 3. Non-Retryable Failure Scenario
- **File**: `scenarios/inventory-non-retryable-failure.json`
- **Status**: 400 Bad Request
- **Response**: `{"available": false, "message": "Invalid product ID or product not found"}`
- **Use Case**: Tests permanent failures that should not be retried (invalid data)
- **Priority**: 0 (highest - overrides default)
- **Loaded**: Dynamically via admin API or helper script

## Quick Start

### Start WireMock
```bash
docker compose up -d
```

### Stop WireMock
```bash
docker compose down
```

## Testing Scenarios

The easiest way to test scenarios is using the helper script:

### Using the Helper Script

```bash
# Test success scenario (default)
./wiremock/scenarios.sh test-success

# Test intermittent failure scenario
# First request returns 503, second returns 200
./wiremock/scenarios.sh test-intermittent

# Test non-retryable failure scenario
# Returns 400 Bad Request
./wiremock/scenarios.sh test-non-retryable
```

### Helper Script Commands

```bash
./wiremock/scenarios.sh help                # Show all available commands
./wiremock/scenarios.sh status              # View current scenario states and mappings
./wiremock/scenarios.sh success             # Enable success scenario
./wiremock/scenarios.sh intermittent        # Enable intermittent failure scenario
./wiremock/scenarios.sh non-retryable       # Enable non-retryable failure scenario
./wiremock/scenarios.sh reset               # Reset all scenarios to default
```

### Manual Testing with curl

#### Test Success Scenario (Default)
```bash
curl -X POST http://localhost:8080/inventory/check \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "00000000-0000-0000-0000-000000000001",
    "quantity": 10
  }'
```

Expected response:
```json
{
  "available": true,
  "message": "Product is in stock"
}
```

#### Test Intermittent Failure Scenario

First, add the intermittent failure mappings:
```bash
curl -X POST http://localhost:8080/__admin/mappings \
  -H "Content-Type: application/json" \
  -d @wiremock/scenarios/inventory-intermittent-failure.json

curl -X POST http://localhost:8080/__admin/mappings \
  -H "Content-Type: application/json" \
  -d @wiremock/scenarios/inventory-intermittent-failure-recovery.json
```

Set the scenario state:
```bash
curl -X PUT http://localhost:8080/__admin/scenarios/IntermittentFailure/state \
  -H "Content-Type: application/json" \
  -d '{"state": "Started"}'
```

Make two requests:
```bash
# First request - will return 503
curl -X POST http://localhost:8080/inventory/check \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "00000000-0000-0000-0000-000000000002",
    "quantity": 5
  }'

# Second request - will return 200
curl -X POST http://localhost:8080/inventory/check \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "00000000-0000-0000-0000-000000000002",
    "quantity": 5
  }'
```

#### Test Non-Retryable Failure Scenario

Add the non-retryable mapping:
```bash
curl -X POST http://localhost:8080/__admin/mappings \
  -H "Content-Type: application/json" \
  -d @wiremock/scenarios/inventory-non-retryable-failure.json
```

Make a request:
```bash
curl -X POST http://localhost:8080/inventory/check \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "00000000-0000-0000-0000-000000000003",
    "quantity": 1
  }'
```

Expected response:
```json
{
  "available": false,
  "message": "Invalid product ID or product not found"
}
```

Reset to default:
```bash
curl -X POST http://localhost:8080/__admin/mappings/reset
docker compose restart wiremock
```

## WireMock Admin API

### View all mappings
```bash
curl http://localhost:8080/__admin/mappings
```

### View scenario state
```bash
curl http://localhost:8080/__admin/scenarios
```

### Reset all scenarios
```bash
curl -X POST http://localhost:8080/__admin/scenarios/reset
```

### View request journal
```bash
curl http://localhost:8080/__admin/requests
```

## How It Works

### Priority System

WireMock uses priority to determine which stub to use when multiple stubs match:
- **Lower priority number = Higher precedence**
- **Priority 0** = Highest precedence (used for scenario-specific stubs)
- **Priority 1** = Default fallback (used for success scenario)

When you dynamically load scenario-specific mappings with priority 0, they override the default success mapping (priority 1).

### Scenario State Management

The intermittent failure scenario uses WireMock's scenario state feature:
1. Initial state: `Started`
2. First request matches the stub with `requiredScenarioState: "Started"`, returns 503, and transitions to `SecondAttempt`
3. Second request matches the stub with `requiredScenarioState: "SecondAttempt"`, returns 200, and transitions back to `Started`

This creates a deterministic pattern: fail → succeed → fail → succeed...

### Integration with Temporal

The inventory client in `internal/integrations/inventory/client.go` interprets HTTP status codes:
- **200**: Success - workflow continues
- **503/500**: Retryable error - Temporal will retry the activity based on retry policy
- **400**: Non-retryable error - Temporal will fail the activity without retrying

## Configuration

### Retry Policy Configuration

To configure how Temporal retries the Validate activity when it encounters retryable errors (503/500), update your activity options in the workflow:

```go
activityOptions := workflow.ActivityOptions{
    StartToCloseTimeout: 10 * time.Second,
    RetryPolicy: &temporal.RetryPolicy{
        InitialInterval:    time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    time.Minute,
        MaximumAttempts:    5,
    },
}
```

### WireMock Configuration

WireMock settings can be adjusted in `docker-compose.yml`:
- Port mapping: Change `8080:8080` to use a different host port
- Verbose logging: Already enabled with `--verbose` flag
- Response templating: Already enabled with `--global-response-templating` flag
