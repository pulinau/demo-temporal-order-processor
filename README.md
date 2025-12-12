# Temporal Order Processing Demo
 
A demonstration of order processing workflows using Temporal.io, featuring durable 
execution, signal handling, and retry policies.
 
 ## Prerequisites
 
 - [mise](https://mise.jit.su/) - Development tool version manager
 - [Docker](https://www.docker.com/) - For running Temporal server and WireMock
 
 ## Setup
 
 ### 1. Install Development Tools
 
 This project uses `mise` to manage Go and mockery versions:
 
 ```bash
 # Install mise (if not already installed)
 curl https://mise.run | sh
 
 # Install project tools (Go 1.25.5 and mockery 3.6.1)
 mise install
 ```
 
 ### 2. Install Dependencies
 
 ```bash
 go mod download
 ```
 
 ### 3. Start Services
 
 ```bash
 # Start Temporal server and WireMock
 make worker.deps.start
 ```
 
 This starts:
 - **Temporal Server**: `localhost:7233`
 - **Temporal Web UI**: http://localhost:8233
 - **WireMock (Inventory API)**: http://localhost:8080
 
 ## Running the Application
 
 ### Start the Worker
 
 The worker runs workflows and activities:
 
 ```bash
 make worker.start
 ```
 
 This starts a worker listening on the `order-proccesor-queue` task queue.
 
 ### Execute a Workflow
 
 In a separate terminal, run the client to start an order workflow:
 
 ```bash
 go run cmd/client/main.go \
   -config="./config/client/local/config.yaml" \
   -order='{
     "id": "00000000-0000-0000-0000-000000000001",
     "line_items": [
       {
         "product_id": "00000000-0000-0000-0000-000000000001",
         "quantity": 10,
         "price_per_item": "29.99"
       }
     ]
   }'
 ```
 
 The workflow will:
 1. Validate the order and check inventory
 2. Wait for a `pickOrder` signal (or `cancelOrder`)
 3. Process the order
 4. Wait for `shipOrder` signal
 5. Wait for `markOrderAsDelivered` signal
 6. Complete with status
 
 ### Interacting with Workflows
 
 Send signals using the Temporal CLI or Web UI:
 
 ```bash
 # Pick the order (moves from PLACED to PICKED)
 temporal workflow signal \
   --workflow-id order-<uuid> \
   --name pickOrder
 
 # Ship the order (moves to SHIPPED)
 temporal workflow signal \
   --workflow-id order-<uuid> \
   --name shipOrder
 
 # Mark as delivered (moves to COMPLETED)
 temporal workflow signal \
   --workflow-id order-<uuid> \
   --name markOrderAsDelivered
 
 # Or cancel the order (before picking)
 temporal workflow signal \
   --workflow-id order-<uuid> \
   --name cancelOrder
 ```
 
 Query the order status:
 
 ```bash
 temporal workflow query \
   --workflow-id order-<uuid> \
   --name GetOrderStatus
 ```
 
 ## Testing
 
 ```bash
 # Run all tests with coverage
 make test
 
 # Generate HTML coverage report
 make cover
 
 # Test different inventory scenarios
 ./wiremock/scenarios.sh test-success
 ./wiremock/scenarios.sh test-intermittent
 ./wiremock/scenarios.sh test-non-retryable
 ```
 
 ## Development
 
 ```bash
 # Format code and tidy dependencies
 make tidy
 
 # Generate mocks
 make generate.mocks
 
 # Stop services
 make worker.deps.stop
 ```
 
 ## Project Structure
 
 ```
 .
 ├── cmd/
 │   ├── worker/          # Temporal worker entrypoint
 │   └── client/          # Workflow execution client
 ├── internal/
 │   ├── temporal/        # Workflows and activities
 │   └── integrations/    # External service clients (inventory API)
 ├── config/              # YAML configuration files
 ├── wiremock/           # Mock inventory service
 └── Makefile            # Build and run targets
 ```
 
 For detailed architecture information, see [CLAUDE.md](./CLAUDE.md).