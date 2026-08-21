package invoices

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"korp/faturamento-service/internal/stock"
)

type invoiceClosingServiceStub struct {
	close func(context.Context, int64) (*Invoice, error)
}

func (s invoiceClosingServiceStub) Close(ctx context.Context, number int64) (*Invoice, error) {
	return s.close(ctx, number)
}

func TestInvoiceClosingHandlerClose(t *testing.T) {
	want := &Invoice{Number: 7, Status: StatusClosed}
	service := invoiceClosingServiceStub{
		close: func(_ context.Context, number int64) (*Invoice, error) {
			if number != want.Number {
				t.Fatalf("number = %d, want %d", number, want.Number)
			}
			return want, nil
		},
	}
	request := httptest.NewRequest(http.MethodPost, "/invoices/7/close", nil)
	request.SetPathValue("number", "7")
	response := httptest.NewRecorder()

	NewInvoiceClosingHandler(service).Close(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got Invoice
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Number != want.Number || got.Status != StatusClosed {
		t.Fatalf("response = %#v, want closed invoice", got)
	}
}

func TestInvoiceClosingHandlerErrors(t *testing.T) {
	tests := []struct {
		name        string
		number      string
		serviceErr  error
		wantStatus  int
		wantMessage string
	}{
		{name: "invalid number", number: "invalid", wantStatus: http.StatusBadRequest, wantMessage: "A numeração da nota é inválida."},
		{name: "missing invoice", number: "7", serviceErr: ErrInvoiceNotFound, wantStatus: http.StatusNotFound, wantMessage: "Nota fiscal não encontrada."},
		{name: "closed invoice", number: "7", serviceErr: ErrInvoiceNotOpen, wantStatus: http.StatusConflict, wantMessage: "Apenas notas abertas podem ser impressas."},
		{
			name:        "stock rejection",
			number:      "7",
			serviceErr:  &stock.RejectionError{StatusCode: http.StatusConflict, Message: "Saldo insuficiente para concluir a operação."},
			wantStatus:  http.StatusConflict,
			wantMessage: "Saldo insuficiente para concluir a operação.",
		},
		{name: "stock unavailable", number: "7", serviceErr: stock.ErrUnavailable, wantStatus: http.StatusServiceUnavailable, wantMessage: "Serviço de estoque indisponível. Tente novamente."},
		{name: "unexpected error", number: "7", serviceErr: errors.New("database password leaked"), wantStatus: http.StatusInternalServerError, wantMessage: "Erro inesperado no servidor."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := invoiceClosingServiceStub{
				close: func(context.Context, int64) (*Invoice, error) {
					return nil, tt.serviceErr
				},
			}
			request := httptest.NewRequest(http.MethodPost, "/invoices/"+tt.number+"/close", nil)
			request.SetPathValue("number", tt.number)
			response := httptest.NewRecorder()

			NewInvoiceClosingHandler(service).Close(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			body := response.Body.Bytes()
			var payload struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Error != tt.wantMessage {
				t.Fatalf("error = %q, want %q", payload.Error, tt.wantMessage)
			}
			if bytes.Contains(body, []byte("database password leaked")) {
				t.Fatal("response exposed internal error")
			}
		})
	}
}
