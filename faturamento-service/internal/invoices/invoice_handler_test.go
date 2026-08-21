package invoices

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type invoiceServiceStub struct {
	create      func(context.Context, CreateInput) (*Invoice, error)
	list        func(context.Context) ([]Invoice, error)
	getByNumber func(context.Context, int64) (*Invoice, error)
}

func (s invoiceServiceStub) Create(ctx context.Context, input CreateInput) (*Invoice, error) {
	return s.create(ctx, input)
}

func (s invoiceServiceStub) List(ctx context.Context) ([]Invoice, error) {
	return s.list(ctx)
}

func (s invoiceServiceStub) GetByNumber(ctx context.Context, number int64) (*Invoice, error) {
	return s.getByNumber(ctx, number)
}

func TestInvoiceHandlerCreate(t *testing.T) {
	wantInput := CreateInput{Items: []InvoiceItem{{ProductID: 1, Quantity: 2}}}
	wantInvoice := &Invoice{Number: 1, Status: StatusOpen, Items: wantInput.Items}
	service := invoiceServiceStub{
		create: func(_ context.Context, input CreateInput) (*Invoice, error) {
			if !reflect.DeepEqual(input, wantInput) {
				t.Fatalf("Create input = %#v, want %#v", input, wantInput)
			}
			return wantInvoice, nil
		},
	}
	request := httptest.NewRequest(http.MethodPost, "/invoices", strings.NewReader(`{
		"items":[{"product_id":1,"quantity":2}]
	}`))
	response := httptest.NewRecorder()

	NewInvoiceHandler(service).Create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	var got Invoice
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, *wantInvoice) {
		t.Fatalf("response = %#v, want %#v", got, *wantInvoice)
	}
}

func TestInvoiceHandlerCreateErrors(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		serviceErr      error
		wantStatus      int
		wantServiceCall bool
	}{
		{name: "malformed JSON", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "unknown field", body: `{"items":[],"unknown":true}`, wantStatus: http.StatusBadRequest},
		{
			name:            "invalid invoice",
			body:            `{"items":[]}`,
			serviceErr:      ErrInvalidInput,
			wantStatus:      http.StatusUnprocessableEntity,
			wantServiceCall: true,
		},
		{
			name:            "unexpected error",
			body:            `{"items":[{"product_id":1,"quantity":1}]}`,
			serviceErr:      errors.New("database password leaked"),
			wantStatus:      http.StatusInternalServerError,
			wantServiceCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceCalled := false
			service := invoiceServiceStub{
				create: func(context.Context, CreateInput) (*Invoice, error) {
					serviceCalled = true
					return nil, tt.serviceErr
				},
			}
			request := httptest.NewRequest(http.MethodPost, "/invoices", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			NewInvoiceHandler(service).Create(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if serviceCalled != tt.wantServiceCall {
				t.Fatalf("service called = %t, want %t", serviceCalled, tt.wantServiceCall)
			}
			if bytes.Contains(response.Body.Bytes(), []byte("database password leaked")) {
				t.Fatal("response exposed internal error")
			}
		})
	}
}

func TestInvoiceHandlerList(t *testing.T) {
	want := []Invoice{{Number: 1, Status: StatusOpen}}
	service := invoiceServiceStub{
		list: func(context.Context) ([]Invoice, error) { return want, nil },
	}
	request := httptest.NewRequest(http.MethodGet, "/invoices", nil)
	response := httptest.NewRecorder()

	NewInvoiceHandler(service).List(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got []Invoice
	if err := json.NewDecoder(response.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("response = %#v, want %#v", got, want)
	}
}

func TestInvoiceHandlerListReturnsEmptyArray(t *testing.T) {
	service := invoiceServiceStub{
		list: func(context.Context) ([]Invoice, error) { return nil, nil },
	}
	request := httptest.NewRequest(http.MethodGet, "/invoices", nil)
	response := httptest.NewRecorder()

	NewInvoiceHandler(service).List(response, request)

	if body := strings.TrimSpace(response.Body.String()); body != "[]" {
		t.Fatalf("body = %q, want []", body)
	}
}

func TestInvoiceHandlerGetByNumber(t *testing.T) {
	want := &Invoice{Number: 7, Status: StatusOpen}
	service := invoiceServiceStub{
		getByNumber: func(_ context.Context, number int64) (*Invoice, error) {
			if number != want.Number {
				t.Fatalf("number = %d, want %d", number, want.Number)
			}
			return want, nil
		},
	}
	request := httptest.NewRequest(http.MethodGet, "/invoices/7", nil)
	request.SetPathValue("number", "7")
	response := httptest.NewRecorder()

	NewInvoiceHandler(service).GetByNumber(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestInvoiceHandlerGetByNumberErrors(t *testing.T) {
	tests := []struct {
		name       string
		number     string
		serviceErr error
		wantStatus int
	}{
		{name: "invalid number", number: "invalid", wantStatus: http.StatusBadRequest},
		{name: "missing invoice", number: "7", serviceErr: ErrInvoiceNotFound, wantStatus: http.StatusNotFound},
		{name: "unexpected error", number: "7", serviceErr: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := invoiceServiceStub{
				getByNumber: func(context.Context, int64) (*Invoice, error) {
					return nil, tt.serviceErr
				},
			}
			request := httptest.NewRequest(http.MethodGet, "/invoices/"+tt.number, nil)
			request.SetPathValue("number", tt.number)
			response := httptest.NewRecorder()

			NewInvoiceHandler(service).GetByNumber(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}
