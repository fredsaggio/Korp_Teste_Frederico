package invoices

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"korp/faturamento-service/internal/httputils"
	"korp/faturamento-service/internal/stock"
)

type InvoiceClosingHandler struct {
	service InvoiceClosingService
}

func NewInvoiceClosingHandler(service InvoiceClosingService) *InvoiceClosingHandler {
	return &InvoiceClosingHandler{service: service}
}

func (h *InvoiceClosingHandler) Close(w http.ResponseWriter, r *http.Request) {
	number, err := strconv.ParseInt(r.PathValue("number"), 10, 64)
	if err != nil || number <= 0 {
		httputils.RespondError(w, http.StatusBadRequest, "A numeração da nota é inválida.")
		return
	}

	invoice, err := h.service.Close(r.Context(), number)
	if err != nil {
		var rejection *stock.RejectionError
		switch {
		case errors.Is(err, ErrInvoiceNotFound):
			httputils.RespondError(w, http.StatusNotFound, "Nota fiscal não encontrada.")
		case errors.Is(err, ErrInvoiceNotOpen):
			httputils.RespondError(w, http.StatusConflict, "Apenas notas abertas podem ser impressas.")
		case errors.As(err, &rejection):
			httputils.RespondError(w, rejection.StatusCode, rejection.Message)
		case errors.Is(err, stock.ErrUnavailable):
			httputils.RespondError(w, http.StatusServiceUnavailable, "Serviço de estoque indisponível. Tente novamente.")
		default:
			slog.Error("failed to close invoice", "number", number, "error", err)
			httputils.RespondError(w, http.StatusInternalServerError, "Erro inesperado no servidor.")
		}
		return
	}

	httputils.Respond(w, http.StatusOK, invoice)
}
