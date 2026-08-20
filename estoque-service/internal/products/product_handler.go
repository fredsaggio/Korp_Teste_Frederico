package products

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

const maxProductRequestBodyBytes = 1 << 20

type ProductHandler struct {
	service ProductService
}

type createProductRequest struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Balance     *int64 `json:"balance"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewProductHandler(service ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxProductRequestBodyBytes)

	var request createProductRequest
	if err := decodeJSON(r.Body, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "Não foi possível processar os dados enviados."})
		return
	}
	if request.Balance == nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: "O saldo é obrigatório."})
		return
	}

	product, err := h.service.Create(r.Context(), ProductInput{
		Code:        request.Code,
		Description: request.Description,
		Balance:     *request.Balance,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: "Código e descrição são obrigatórios, e o saldo não pode ser negativo."})
		case errors.Is(err, ErrCodeAlreadyExists):
			writeJSON(w, http.StatusConflict, errorResponse{Error: "Já existe um produto com esse código."})
		default:
			slog.Error("failed to create product", "error", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Erro inesperado no servidor."})
		}
		return
	}

	writeJSON(w, http.StatusCreated, product)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.List(r.Context())
	if err != nil {
		slog.Error("failed to list products", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "Erro inesperado no servidor."})
		return
	}
	if products == nil {
		products = make([]Product, 0)
	}

	writeJSON(w, http.StatusOK, products)
}

func decodeJSON(body io.Reader, destination any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to encode HTTP response", "error", err)
	}
}
