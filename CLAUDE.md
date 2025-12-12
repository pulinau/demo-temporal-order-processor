# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based Temporal order processing demonstration. The workflow orchestrates order validation, inventory checks, picking, shipping, and delivery tracking using Temporal.io's durable execution engine.

## Architecture

### Core Components

**Workflow**: `internal/temporal/workflow.go:45`
- Main workflow: `ProccessOrder` orchestrates the complete order lifecycle
- Handles signals for: order picking, shipping, delivery, and cancellation
- Query handler: `GetOrderStatus` returns current order status
- Signal names defined as constants at `internal/temporal/workflow.go:34-39`

**Activities**: `internal/temporal/activities.go`
- `Validate`: Validates order and checks inventory via external API
- `Process`: Placeholder for order processing logic (not yet implemented at `internal/temporal/activities.go:65`)

**Order Lifecycle States**: `internal/temporal/status.go`
- `PLACED` → `PICKED` → `SHIPPED` → `COMPLETED`
- Can be `CANCELLED` at any point before picking
- `UNABLE_TO_COMPLETE` for validation or processing failures

### External Dependencies

**Inventory Service Client**: `internal/integrations/inventory/client.go`
- HTTP client that checks product availability
- Interprets status codes:
  - 200: Success
  - 503/500: Retryable errors (Temporal will retry based on activity retry policy)
  - 400: Non-retryable errors (workflow fails immediately)

**WireMock**: Mock inventory service for testing
- Runs on `localhost:8080`
- Scenarios managed via `./wiremock/scenarios.sh`
- See `wiremock/README.md` for detailed documentation

### Configuration

Uses Viper for YAML config loading with validation.

**Worker Config**: `config/worker/local/config.yaml`
- Temporal connection settings (host, port, task queue)
- Inventory API base URL

**Client Config**: `config/client/local/config.yaml`
- Similar to worker config but for workflow client

## Common Commands

### Build & Test

```bash
# Format and tidy dependencies
make tidy

# Run all tests with race detection and coverage
make test

# Generate HTML coverage report
make cover

# Generate mocks (uses mockery)
make generate.mocks
```

### Run Locally

```bash
# Start dependencies (Temporal server + WireMock)
make worker.deps.start

# Stop dependencies
make worker.deps.stop

# Restart dependencies
make worker.deps.restart

# Start the Temporal worker
make worker.start

# Run the client (requires JSON order payload)
go run cmd/client/main.go -config="./config/client/local/config.yaml" -order='{"id":"...","line_items":[...]}'
```

### Docker Services

```bash
# Start all services
docker compose up -d

# Stop all services
docker compose down
```

**Temporal Web UI**: http://localhost:8233
**WireMock Admin**: http://localhost:8080/__admin

### Testing Inventory Scenarios

Pre-approved commands for testing different failure scenarios:

```bash
# Test intermittent failures (first attempt fails, second succeeds)
./wiremock/scenarios.sh test-intermittent

# Test successful inventory checks
./wiremock/scenarios.sh test-success

# Test non-retryable failures (validation errors)
./wiremock/scenarios.sh test-non-retryable
```

## Key Implementation Details

### Workflow Signal Pattern

The workflow uses Temporal's selector pattern to wait for signals. Example at `internal/temporal/workflow.go:67-81`:
- Creates selector with multiple signal channels
- Blocks until one signal is received
- Updates order status based on received signal

### Retry Configuration

Two activity option sets defined at `internal/temporal/workflow.go:12-30`:
- `validateActivityOptions`: 10s timeout, 5 max attempts (for quick validation)
- `defaultActivityOptions`: 1m timeout, 4 max attempts (for longer operations)

### Inventory Check Retry Logic

The inventory client returns Go errors that Temporal interprets based on:
- Regular errors from 503/500: Temporal retries automatically
- Non-retryable application errors from 400: Workflow fails without retry (see `internal/temporal/activities.go:54-58`)

### Testing with Mocks

Generated mocks are in `internal/temporal/mocks/`. The `InventoryChecker` interface is mocked for unit testing activities without real HTTP calls.

## Development Notes

- Temporal SDK version: v1.38.0
- Go version: 1.25.5
- Worker task queue name: `order-proccesor-queue` (note: typo in original config)
- The `Process` activity is intentionally unimplemented - returns error at `internal/temporal/activities.go:66`
