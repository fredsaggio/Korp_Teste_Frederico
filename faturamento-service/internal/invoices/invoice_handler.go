package invoices

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"korp/faturamento-service/internal/httputils"
)

type InvoiceHandler struct {
	service InvoiceService
}

func NewInvoiceHandler(service InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{service: service}
}

func (h *InvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := httputils.DecodeJSON(w, r, &input); err != nil {
		httputils.RespondError(w, http.StatusBadRequest, "Não foi possível processar os dados enviados.")
		return
	}

	invoice, err := h.service.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			httputils.RespondError(w, http.StatusUnprocessableEntity, "A nota deve possuir produtos e quantidades válidos, sem repetições.")
			return
		}
		slog.Error("failed to create invoice", "error", err)
		httputils.RespondError(w, http.StatusInternalServerError, "Erro inesperado no servidor.")
		return
	}

	httputils.Respond(w, http.StatusCreated, invoice)
}

func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.service.List(r.Context())
	if err != nil {
		slog.Error("failed to list invoices", "error", err)
		httputils.RespondError(w, http.StatusInternalServerError, "Erro inesperado no servidor.")
		return
	}
	if invoices == nil {
		invoices = make([]Invoice, 0)
	}

	httputils.Respond(w, http.StatusOK, invoices)
}

func (h *InvoiceHandler) GetByNumber(w http.ResponseWriter, r *http.Request) {
	number, err := strconv.ParseInt(r.PathValue("number"), 10, 64)
	if err != nil || number <= 0 {
		httputils.RespondError(w, http.StatusBadRequest, "A numeração da nota é inválida.")
		return
	}

	invoice, err := h.service.GetByNumber(r.Context(), number)
	if err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			httputils.RespondError(w, http.StatusNotFound, "Nota fiscal não encontrada.")
			return
		}
		slog.Error("failed to get invoice", "number", number, "error", err)
		httputils.RespondError(w, http.StatusInternalServerError, "Erro inesperado no servidor.")
		return
	}

	httputils.Respond(w, http.StatusOK, invoice)
}
