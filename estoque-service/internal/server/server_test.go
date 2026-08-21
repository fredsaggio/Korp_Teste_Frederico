package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"korp/estoque-service/internal/products"
	"korp/estoque-service/internal/stock"
)

type productServiceStub struct{}

func (productServiceStub) Create(_ context.Context, input products.ProductInput) (*products.Product, error) {
	return &products.Product{
		ID:          1,
		Code:        input.Code,
		Description: input.Description,
		Balance:     input.Balance,
	}, nil
}

func (productServiceStub) List(context.Context) ([]products.Product, error) {
	return []products.Product{{ID: 1, Code: "PROD-001", Description: "Product", Balance: 10}}, nil
}

type debitServiceStub struct{}

func (debitServiceStub) Debit(_ context.Context, input stock.DebitInput) (*stock.DebitResult, error) {
	return &stock.DebitResult{Reference: input.Reference}, nil
}

func newTestHandler() http.Handler {
	productHandler := products.NewProductHandler(productServiceStub{})
	debitHandler := stock.NewDebitHandler(debitServiceStub{})
	return New(Handlers{
		ProductHandler: productHandler,
		DebitHandler:   debitHandler,
	}).Handler
}

func TestStockDebitRoute(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/stock/debits",
		strings.NewReader(`{"reference":"INVOICE-001","items":[{"product_id":1,"quantity":2}]}`),
	)
	response := httptest.NewRecorder()

	newTestHandler().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
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

func TestProductRoutes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{
			name:       "create",
			method:     http.MethodPost,
			body:       `{"code":"PROD-001","description":"Product","balance":10}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "list",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/api/v1/products", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			newTestHandler().ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}
