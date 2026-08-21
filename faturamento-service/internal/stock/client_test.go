package stock

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func newTestClient(t *testing.T, transport roundTripFunc, timeout time.Duration) *HTTPClient {
	t.Helper()

	client, err := NewHTTPClient("http://stock-service", timeout)
	if err != nil {
		t.Fatal(err)
	}
	client.client.Transport = transport
	return client
}

func TestHTTPClientDebitSuccess(t *testing.T) {
	wantInput := DebitInput{
		Reference: "invoice:10",
		Items:     []Item{{ProductID: 1, Quantity: 2}},
	}

	for _, status := range []int{http.StatusCreated, http.StatusOK} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodPost || r.URL.Path != "/api/v1/stock/debits" {
					t.Errorf("request = %s %s, want POST /api/v1/stock/debits", r.Method, r.URL.Path)
				}
				if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", contentType)
				}

				var gotInput DebitInput
				if err := json.NewDecoder(r.Body).Decode(&gotInput); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(gotInput, wantInput) {
					t.Errorf("input = %#v, want %#v", gotInput, wantInput)
				}
				return &http.Response{
					StatusCode: status,
					Body:       io.NopCloser(strings.NewReader(`{"already_processed":false}`)),
				}, nil
			}, time.Second)
			if err := client.Debit(context.Background(), wantInput); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestHTTPClientDebitReturnsRejection(t *testing.T) {
	client := newTestClient(t, func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusConflict,
			Body:       io.NopCloser(strings.NewReader(`{"error":"Saldo insuficiente para concluir a operação."}`)),
		}, nil
	}, time.Second)
	err := client.Debit(context.Background(), DebitInput{})

	var rejection *RejectionError
	if !errors.As(err, &rejection) {
		t.Fatalf("Debit() error = %v, want RejectionError", err)
	}
	if rejection.StatusCode != http.StatusConflict || rejection.Message != "Saldo insuficiente para concluir a operação." {
		t.Fatalf("rejection = %#v", rejection)
	}
}

func TestHTTPClientDebitReturnsUnavailable(t *testing.T) {
	client := newTestClient(t, func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}, time.Second)
	err := client.Debit(context.Background(), DebitInput{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Debit() error = %v, want ErrUnavailable", err)
	}
}

func TestHTTPClientDebitTimesOut(t *testing.T) {
	client := newTestClient(t, func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	}, time.Millisecond)
	err := client.Debit(context.Background(), DebitInput{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Debit() error = %v, want ErrUnavailable", err)
	}
}

func TestNewHTTPClientRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		timeout time.Duration
	}{
		{name: "invalid URL", baseURL: "://invalid", timeout: time.Second},
		{name: "unsupported scheme", baseURL: "ftp://stock", timeout: time.Second},
		{name: "invalid timeout", baseURL: "http://stock", timeout: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewHTTPClient(tt.baseURL, tt.timeout); err == nil {
				t.Fatal("NewHTTPClient() error = nil, want error")
			}
		})
	}
}
