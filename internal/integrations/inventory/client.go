package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	BaseURL string `yaml:"baseUrl" validate:"required,http_url"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type CheckInventoryRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int32     `json:"quantity"`
}

type CheckInventoryResponse struct {
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}

// CheckInventory checks if the requested quantity of a product is available in inventory
func (c *Client) CheckInventory(ctx context.Context, productID uuid.UUID, quantity int32) (bool, error) {
	req := CheckInventoryRequest{
		ProductID: productID,
		Quantity:  quantity,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/inventory/check", bytes.NewBuffer(body))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle different status codes
	switch resp.StatusCode {
	case http.StatusOK:
		var checkResp CheckInventoryResponse
		if err := json.Unmarshal(respBody, &checkResp); err != nil {
			return false, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return checkResp.Available, nil

	case http.StatusBadRequest:
		// Non-retryable error (e.g., invalid product)
		var checkResp CheckInventoryResponse
		if err := json.Unmarshal(respBody, &checkResp); err != nil {
			return false, fmt.Errorf("non-retryable error: invalid request")
		}
		return false, fmt.Errorf("non-retryable error: %s", checkResp.Message)

	case http.StatusServiceUnavailable, http.StatusInternalServerError:
		// Retryable error (service temporarily unavailable)
		return false, fmt.Errorf("retryable error: service unavailable (status: %d)", resp.StatusCode)

	default:
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
