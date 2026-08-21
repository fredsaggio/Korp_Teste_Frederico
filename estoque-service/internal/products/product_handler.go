package products

import (
	"errors"
	"log/slog"
	"net/http"

	"korp/estoque-service/internal/httputils"
)

type ProductHandler struct {
	service ProductService
}

type createProductRequest struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Balance     *int64 `json:"balance"`
}

func NewProductHandler(service ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request createProductRequest
	if err := httputils.DecodeJSON(w, r, &request); err != nil {
		httputils.RespondError(w, http.StatusBadRequest, "Não foi possível processar os dados enviados.")
		return
	}
	if request.Balance == nil {
		httputils.RespondError(w, http.StatusUnprocessableEntity, "O saldo é obrigatório.")
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
			httputils.RespondError(w, http.StatusUnprocessableEntity, "Código e descrição são obrigatórios, e o saldo não pode ser negativo.")
		case errors.Is(err, ErrCodeAlreadyExists):
			httputils.RespondError(w, http.StatusConflict, "Já existe um produto com esse código.")
		default:
			slog.Error("failed to create product", "error", err)
			httputils.RespondError(w, http.StatusInternalServerError, "Erro inesperado no servidor.")
		}
		return
	}

	httputils.Respond(w, http.StatusCreated, product)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.List(r.Context())
	if err != nil {
		slog.Error("failed to list products", "error", err)
		httputils.RespondError(w, http.StatusInternalServerError, "Erro inesperado no servidor.")
		return
	}
	if products == nil {
		products = make([]Product, 0)
	}

	httputils.Respond(w, http.StatusOK, products)
}
