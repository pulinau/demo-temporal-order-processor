#!/bin/bash

WIREMOCK_URL="http://localhost:8080"

show_help() {
    echo "WireMock Scenario Manager"
    echo ""
    echo "Usage: ./scenarios.sh [command]"
    echo ""
    echo "Commands:"
    echo "  success              - Enable success scenario (default)"
    echo "  intermittent         - Enable intermittent failure scenario"
    echo "  non-retryable        - Enable non-retryable failure scenario"
    echo "  reset                - Reset all scenarios to default"
    echo "  status               - Show current scenario states"
    echo "  test-success         - Test success scenario"
    echo "  test-intermittent    - Test intermittent failure scenario"
    echo "  test-non-retryable   - Test non-retryable failure scenario"
    echo "  help                 - Show this help message"
}

enable_success() {
    echo "Enabling success scenario (default)..."
    curl -X POST "${WIREMOCK_URL}/__admin/scenarios/reset"
    echo ""
    echo "Success scenario is now active (default behavior)"
}

enable_intermittent() {
    echo "Enabling intermittent failure scenario..."
    # Add both intermittent failure mappings
    curl -X POST "${WIREMOCK_URL}/__admin/mappings" \
        -H "Content-Type: application/json" \
        -d @wiremock/scenarios/inventory-intermittent-failure.json
    echo ""
    curl -X POST "${WIREMOCK_URL}/__admin/mappings" \
        -H "Content-Type: application/json" \
        -d @wiremock/scenarios/inventory-intermittent-failure-recovery.json
    echo ""
    # Initialize the scenario state
    curl -X PUT "${WIREMOCK_URL}/__admin/scenarios/IntermittentFailure/state" \
        -H "Content-Type: application/json" \
        -d '{"state": "Started"}'
    echo ""
    echo "Intermittent failure scenario is now active"
    echo "Next request will fail with 503, subsequent request will succeed"
}

enable_non_retryable() {
    echo "Enabling non-retryable failure scenario..."
    # Add the non-retryable mapping via admin API
    curl -X POST "${WIREMOCK_URL}/__admin/mappings" \
        -H "Content-Type: application/json" \
        -d @wiremock/scenarios/inventory-non-retryable-failure.json
    echo ""
    echo "Non-retryable failure scenario is now active"
    echo "All requests will return 400 Bad Request"
    echo "Run './scenarios.sh reset' to restore default scenario"
}

reset_scenarios() {
    echo "Resetting all scenarios..."
    curl -X POST "${WIREMOCK_URL}/__admin/scenarios/reset"
    echo ""
    curl -X POST "${WIREMOCK_URL}/__admin/mappings/reset"
    echo ""
    docker compose restart wiremock > /dev/null 2>&1
    echo "All scenarios reset to default (success)"
}

show_status() {
    echo "Current scenario states:"
    echo "------------------------"
    curl -s "${WIREMOCK_URL}/__admin/scenarios" | jq '.'
    echo ""
    echo "Available mappings:"
    echo "-------------------"
    curl -s "${WIREMOCK_URL}/__admin/mappings" | jq '.mappings[] | {id: .id, name: .name, priority: .priority}'
}

test_success() {
    echo "Testing success scenario..."
    curl -X POST "${WIREMOCK_URL}/inventory/check" \
        -H "Content-Type: application/json" \
        -d '{
            "product_id": "00000000-0000-0000-0000-000000000001",
            "quantity": 10
        }' | jq '.'
}

test_intermittent() {
    echo "Testing intermittent failure scenario..."
    echo "Adding intermittent failure mappings..."
    curl -s -X POST "${WIREMOCK_URL}/__admin/mappings" \
        -H "Content-Type: application/json" \
        -d @wiremock/scenarios/inventory-intermittent-failure.json > /dev/null
    curl -s -X POST "${WIREMOCK_URL}/__admin/mappings" \
        -H "Content-Type: application/json" \
        -d @wiremock/scenarios/inventory-intermittent-failure-recovery.json > /dev/null

    echo "Setting scenario state to Started..."
    curl -s -X PUT "${WIREMOCK_URL}/__admin/scenarios/IntermittentFailure/state" \
        -H "Content-Type: application/json" \
        -d '{"state": "Started"}' > /dev/null
    echo ""

    echo "First request (should fail with 503):"
    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${WIREMOCK_URL}/inventory/check" \
        -H "Content-Type: application/json" \
        -d '{"product_id": "00000000-0000-0000-0000-000000000002", "quantity": 5}')
    HTTP_CODE=$(echo "$RESPONSE" | grep HTTP_STATUS | cut -d: -f2)
    BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS/d')
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    echo "HTTP Status: ${HTTP_CODE}"

    echo ""
    echo "Second request (should succeed with 200):"
    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${WIREMOCK_URL}/inventory/check" \
        -H "Content-Type: application/json" \
        -d '{"product_id": "00000000-0000-0000-0000-000000000002", "quantity": 5}')
    HTTP_CODE=$(echo "$RESPONSE" | grep HTTP_STATUS | cut -d: -f2)
    BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS/d')
    echo "$BODY" | jq '.'
    echo "HTTP Status: ${HTTP_CODE}"

    echo ""
    echo "Resetting to default..."
    curl -s -X POST "${WIREMOCK_URL}/__admin/mappings/reset" > /dev/null
    docker compose restart wiremock > /dev/null 2>&1
    echo "Done"
}

test_non_retryable() {
    echo "Testing non-retryable failure scenario..."
    echo "Adding non-retryable mapping..."
    curl -s -X POST "${WIREMOCK_URL}/__admin/mappings" \
        -H "Content-Type: application/json" \
        -d @wiremock/scenarios/inventory-non-retryable-failure.json > /dev/null
    echo ""

    echo "Request (should fail with 400):"
    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${WIREMOCK_URL}/inventory/check" \
        -H "Content-Type: application/json" \
        -d '{"product_id": "00000000-0000-0000-0000-000000000003", "quantity": 1}')
    HTTP_CODE=$(echo "$RESPONSE" | grep HTTP_STATUS | cut -d: -f2)
    BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS/d')
    echo "$BODY" | jq '.'
    echo "HTTP Status: ${HTTP_CODE}"

    echo ""
    echo "Resetting to default..."
    curl -s -X POST "${WIREMOCK_URL}/__admin/mappings/reset" > /dev/null
    docker compose restart wiremock > /dev/null 2>&1
    echo "Done"
}

# Main command handling
case "${1}" in
    success)
        enable_success
        ;;
    intermittent)
        enable_intermittent
        ;;
    non-retryable)
        enable_non_retryable
        ;;
    reset)
        reset_scenarios
        ;;
    status)
        show_status
        ;;
    test-success)
        test_success
        ;;
    test-intermittent)
        test_intermittent
        ;;
    test-non-retryable)
        test_non_retryable
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        show_help
        exit 1
        ;;
esac
