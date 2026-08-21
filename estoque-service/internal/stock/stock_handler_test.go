package stock

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

	"korp/estoque-service/internal/httputils"
)

type debitServiceStub struct {
	debit func(context.Context, DebitInput) (*DebitResult, error)
}

func (s debitServiceStub) Debit(ctx context.Context, input DebitInput) (*DebitResult, error) {
	return s.debit(ctx, input)
}

func TestDebitHandlerSuccess(t *testing.T) {
	tests := []struct {
		name       string
		result     *DebitResult
		wantStatus int
	}{
		{
			name:       "first processing",
			result:     &DebitResult{Reference: "invoice:123"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "idempotent retry",
			result:     &DebitResult{Reference: "invoice:123", AlreadyProcessed: true},
			wantStatus: http.StatusOK,
		},
	}

	wantInput := DebitInput{
		Reference: "invoice:123",
		Items:     []DebitItem{{ProductID: 1, Quantity: 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := debitServiceStub{
				debit: func(_ context.Context, input DebitInput) (*DebitResult, error) {
					if !reflect.DeepEqual(input, wantInput) {
						t.Fatalf("Debit input = %#v, want %#v", input, wantInput)
					}
					return tt.result, nil
				},
			}
			handler := NewDebitHandler(service)
			request := httptest.NewRequest(http.MethodPost, "/stock/debits", strings.NewReader(`{
				"reference":"invoice:123",
				"items":[{"product_id":1,"quantity":2}]
			}`))
			response := httptest.NewRecorder()

			handler.Debit(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			var got DebitResult
			if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, *tt.result) {
				t.Fatalf("response = %#v, want %#v", got, *tt.result)
			}
		})
	}
}

func TestDebitHandlerErrors(t *testing.T) {
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
			name:            "invalid input",
			body:            `{"reference":"","items":[]}`,
			serviceErr:      fmt.Errorf("validate debit: %w", ErrInvalidInput),
			wantStatus:      http.StatusUnprocessableEntity,
			wantMessage:     "Referência e itens válidos são obrigatórios.",
			wantServiceCall: true,
		},
		{
			name:            "idempotency conflict",
			body:            `{"reference":"invoice:123","items":[{"product_id":1,"quantity":2}]}`,
			serviceErr:      ErrIdempotencyConflict,
			wantStatus:      http.StatusConflict,
			wantMessage:     "A referência já foi utilizada com itens diferentes.",
			wantServiceCall: true,
		},
		{
			name:            "product not found",
			body:            `{"reference":"invoice:123","items":[{"product_id":1,"quantity":2}]}`,
			serviceErr:      ErrProductNotFound,
			wantStatus:      http.StatusNotFound,
			wantMessage:     "Produto não encontrado.",
			wantServiceCall: true,
		},
		{
			name:            "insufficient stock",
			body:            `{"reference":"invoice:123","items":[{"product_id":1,"quantity":2}]}`,
			serviceErr:      ErrInsufficientStock,
			wantStatus:      http.StatusConflict,
			wantMessage:     "Saldo insuficiente para concluir a operação.",
			wantServiceCall: true,
		},
		{
			name:            "unexpected error",
			body:            `{"reference":"invoice:123","items":[{"product_id":1,"quantity":2}]}`,
			serviceErr:      errors.New("database password leaked"),
			wantStatus:      http.StatusInternalServerError,
			wantMessage:     "Erro inesperado no servidor.",
			wantServiceCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := debitServiceStub{
				debit: func(context.Context, DebitInput) (*DebitResult, error) {
					serviceCalled = true
					return nil, tt.serviceErr
				},
			}
			request := httptest.NewRequest(http.MethodPost, "/stock/debits", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			NewDebitHandler(service).Debit(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if serviceCalled != tt.wantServiceCall {
				t.Fatalf("service called = %t, want %t", serviceCalled, tt.wantServiceCall)
			}
			body := response.Body.Bytes()
			var got httputils.ErrorResponse
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
