package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fruganyumisa/igrid-middleware/internal/models"
	"github.com/fruganyumisa/igrid-middleware/internal/pkg/logger"
)

// HTTPClientConfig contains configuration for HTTP client
type HTTPClientConfig struct {
	BaseURL string
	Timeout time.Duration
	Headers map[string]string
}

// NewHTTPClient creates a new HTTP client for sending data to external systems
func NewHTTPClient(config HTTPClientConfig, logger logger.Logger) *HTTPClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}

	// Set default headers
	if _, exists := config.Headers["Content-Type"]; !exists {
		config.Headers["Content-Type"] = "application/json"
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		baseURL: config.BaseURL,
		timeout: config.Timeout,
		logger:  logger,
		headers: config.Headers,
	}
}

// SendToDMS sends payload to DMS application endpointbytes"

// HTTPClient handles HTTP requests to external systems
type HTTPClient struct {
	client  *http.Client
	baseURL string
	timeout time.Duration
	logger  logger.Logger
	headers map[string]string
}

// SendToDMS sends payload to DMS application endpoint
func (h *HTTPClient) SendToDMS(ctx context.Context, payload *models.DMSPayload) error {
	return h.sendJSON(ctx, "/api/v1/devices/data", payload)
} // sendJSON sends a JSON payload to the specified endpoint
func (h *HTTPClient) sendJSON(ctx context.Context, endpoint string, payload interface{}) error {
	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("Failed to marshal payload", map[string]interface{}{
			"error":    err.Error(),
			"endpoint": endpoint,
		})
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create full URL
	url := h.baseURL + endpoint

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		h.logger.Error("Failed to create HTTP request", map[string]interface{}{
			"error": err.Error(),
			"url":   url,
		})
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	// Log the outgoing request
	h.logger.Info("Sending HTTP request to DMS", map[string]interface{}{
		"url":     url,
		"method":  "POST",
		"payload": string(jsonData),
		"headers": h.headers,
	})

	// Send request
	resp, err := h.client.Do(req)
	if err != nil {
		h.logger.Error("HTTP request failed", map[string]interface{}{
			"error": err.Error(),
			"url":   url,
		})
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read response body", map[string]interface{}{
			"error":       err.Error(),
			"status_code": resp.StatusCode,
		})
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log the response
	h.logger.Info("Received HTTP response from DMS", map[string]interface{}{
		"status_code":    resp.StatusCode,
		"response_body":  string(respBody),
		"content_length": len(respBody),
	})

	// Check if request was successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		h.logger.Error("HTTP request returned error status", map[string]interface{}{
			"status_code":   resp.StatusCode,
			"response_body": string(respBody),
			"url":           url,
		})
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	h.logger.Info("Successfully sent data to DMS", map[string]interface{}{
		"status_code": resp.StatusCode,
		"url":         url,
	})

	return nil
}

// SendCustomPayload sends a custom JSON payload to any endpoint
func (h *HTTPClient) SendCustomPayload(ctx context.Context, endpoint string, payload interface{}) error {
	return h.sendJSON(ctx, endpoint, payload)
}

// SetHeader sets a custom header for all requests
func (h *HTTPClient) SetHeader(key, value string) {
	h.headers[key] = value
}

// GetBaseURL returns the base URL of the HTTP client
func (h *HTTPClient) GetBaseURL() string {
	return h.baseURL
}

// GetTimeout returns the timeout duration of the HTTP client
func (h *HTTPClient) GetTimeout() time.Duration {
	return h.timeout
}
