package products

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type productServiceStub struct {
	create func(context.Context, ProductInput) (*Product, error)
	list   func(context.Context) ([]Product, error)
}

func (s productServiceStub) Create(ctx context.Context, input ProductInput) (*Product, error) {
	return s.create(ctx, input)
}

func (s productServiceStub) List(ctx context.Context) ([]Product, error) {
	return s.list(ctx)
}

func TestProductHandlerCreate(t *testing.T) {
	wantInput := ProductInput{Code: "PROD-001", Description: "Product one", Balance: 10}
	wantProduct := &Product{ID: 1, Code: wantInput.Code, Description: wantInput.Description, Balance: wantInput.Balance}
	service := productServiceStub{
		create: func(_ context.Context, input ProductInput) (*Product, error) {
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Create input = %#v, want %#v", input, wantInput)
			}
			return wantProduct, nil
		},
	}
	handler := NewProductHandler(service)

	request := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(`{
		"code":"PROD-001",
		"description":"Product one",
		"balance":10
	}`))
	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	var got Product
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, *wantProduct) {
		t.Fatalf("response = %#v, want %#v", got, *wantProduct)
	}
}

func TestProductHandlerCreateErrors(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		serviceErr      error
		wantStatus      int
		wantMessage     string
		wantServiceCall bool
	}{
		{
			name:        "malformed JSON",
			body:        `{`,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Não foi possível processar os dados enviados.",
		},
		{
			name:        "unknown field",
			body:        `{"code":"PROD-001","description":"Product","balance":1,"unknown":true}`,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Não foi possível processar os dados enviados.",
		},
		{
			name:        "missing balance",
			body:        `{"code":"PROD-001","description":"Product"}`,
			wantStatus:  http.StatusUnprocessableEntity,
			wantMessage: "O saldo é obrigatório.",
		},
		{
			name:            "invalid product",
			body:            `{"code":"","description":"Product","balance":1}`,
			serviceErr:      fmt.Errorf("validate product: %w", ErrInvalidInput),
			wantStatus:      http.StatusUnprocessableEntity,
			wantMessage:     "Código e descrição são obrigatórios, e o saldo não pode ser negativo.",
			wantServiceCall: true,
		},
		{
			name:            "duplicate code",
			body:            `{"code":"PROD-001","description":"Product","balance":1}`,
			serviceErr:      ErrCodeAlreadyExists,
			wantStatus:      http.StatusConflict,
			wantMessage:     "Já existe um produto com esse código.",
			wantServiceCall: true,
		},
		{
			name:            "unexpected error",
			body:            `{"code":"PROD-001","description":"Product","balance":1}`,
			serviceErr:      errors.New("database password leaked"),
			wantStatus:      http.StatusInternalServerError,
			wantMessage:     "Erro inesperado no servidor.",
			wantServiceCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := productServiceStub{
				create: func(context.Context, ProductInput) (*Product, error) {
					serviceCalled = true
					return nil, tt.serviceErr
				},
			}
			handler := NewProductHandler(service)
			request := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			handler.Create(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if serviceCalled != tt.wantServiceCall {
				t.Fatalf("service called = %t, want %t", serviceCalled, tt.wantServiceCall)
			}
			body := response.Body.Bytes()
			var got struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatal(err)
			}
			if got.Error != tt.wantMessage {
				t.Fatalf("error = %q, want %q", got.Error, tt.wantMessage)
			}
			if bytes.Contains(body, []byte("database password leaked")) {
				t.Fatal("response exposed internal error")
			}
		})
	}
}

func TestProductHandlerList(t *testing.T) {
	want := []Product{{ID: 1, Code: "PROD-001", Description: "Product", Balance: 10}}
	service := productServiceStub{
		list: func(context.Context) ([]Product, error) {
			return want, nil
		},
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/products", nil)

	NewProductHandler(service).List(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got []Product
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("response = %#v, want %#v", got, want)
	}
}

func TestProductHandlerListReturnsEmptyArray(t *testing.T) {
	service := productServiceStub{
		list: func(context.Context) ([]Product, error) {
			return nil, nil
		},
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/products", nil)

	NewProductHandler(service).List(response, request)

	if body := strings.TrimSpace(response.Body.String()); body != "[]" {
		t.Fatalf("body = %q, want []", body)
	}
}
