package stock

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrUnavailable = errors.New("stock service unavailable")

type Item struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type DebitInput struct {
	Reference string `json:"reference"`
	Items     []Item `json:"items"`
}

type Client interface {
	Debit(ctx context.Context, input DebitInput) error
}

type RejectionError struct {
	StatusCode int
	Message    string
}

func (e *RejectionError) Error() string {
	return fmt.Sprintf("stock service rejected debit with status %d: %s", e.StatusCode, e.Message)
}

type HTTPClient struct {
	baseURL *url.URL
	client  *http.Client
}

func NewHTTPClient(baseURL string, timeout time.Duration) (*HTTPClient, error) {
	parsedURL, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, errors.New("invalid stock service URL")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, errors.New("stock service URL must use HTTP or HTTPS")
	}
	if timeout <= 0 {
		return nil, errors.New("stock service timeout must be positive")
	}

	return &HTTPClient{
		baseURL: parsedURL,
		client:  &http.Client{Timeout: timeout},
	}, nil
}

func (c *HTTPClient) Debit(ctx context.Context, input DebitInput) error {
	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("encode stock debit request: %w", err)
	}

	endpoint := c.baseURL.JoinPath("api", "v1", "stock", "debits")
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create stock debit request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
		return nil
	}
	if response.StatusCode >= http.StatusBadRequest && response.StatusCode < http.StatusInternalServerError {
		return decodeRejection(response)
	}

	return fmt.Errorf("%w: stock service returned status %d", ErrUnavailable, response.StatusCode)
}

func decodeRejection(response *http.Response) error {
	var payload struct {
		Error string `json:"error"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(&payload); err != nil || strings.TrimSpace(payload.Error) == "" {
		payload.Error = "Operação rejeitada pelo serviço de estoque."
	}

	return &RejectionError{
		StatusCode: response.StatusCode,
		Message:    payload.Error,
	}
}
