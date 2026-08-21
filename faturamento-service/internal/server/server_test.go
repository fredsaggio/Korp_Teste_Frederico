package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"korp/faturamento-service/internal/invoices"
)

type invoiceServiceStub struct{}

type invoiceClosingServiceStub struct{}

func (invoiceServiceStub) Create(_ context.Context, input invoices.CreateInput) (*invoices.Invoice, error) {
	return &invoices.Invoice{Number: 1, Status: invoices.StatusOpen, Items: input.Items}, nil
}

func (invoiceServiceStub) List(context.Context) ([]invoices.Invoice, error) {
	return []invoices.Invoice{{Number: 1, Status: invoices.StatusOpen}}, nil
}

func (invoiceServiceStub) GetByNumber(_ context.Context, number int64) (*invoices.Invoice, error) {
	return &invoices.Invoice{Number: number, Status: invoices.StatusOpen}, nil
}

func (invoiceClosingServiceStub) Close(_ context.Context, number int64) (*invoices.Invoice, error) {
	return &invoices.Invoice{Number: number, Status: invoices.StatusClosed}, nil
}

func newTestHandler() http.Handler {
	invoiceHandler := invoices.NewInvoiceHandler(invoiceServiceStub{})
	invoiceClosingHandler := invoices.NewInvoiceClosingHandler(invoiceClosingServiceStub{})
	return New(Handlers{
		InvoiceHandler:        invoiceHandler,
		InvoiceClosingHandler: invoiceClosingHandler,
	}).Handler
}

func TestHealthRoute(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	newTestHandler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := strings.TrimSpace(response.Body.String()); body != `{"status":"ok"}` {
		t.Fatalf("body = %q, want health response", body)
	}
}

func TestInvoiceRoutes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "create",
			method:     http.MethodPost,
			path:       "/api/v1/invoices",
			body:       `{"items":[{"product_id":1,"quantity":2}]}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "list",
			method:     http.MethodGet,
			path:       "/api/v1/invoices",
			wantStatus: http.StatusOK,
		},
		{
			name:       "get by number",
			method:     http.MethodGet,
			path:       "/api/v1/invoices/1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "close",
			method:     http.MethodPost,
			path:       "/api/v1/invoices/1/close",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			newTestHandler().ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}
