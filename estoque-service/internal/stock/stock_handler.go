package stock

import (
	"errors"
	"log/slog"
	"net/http"

	"korp/estoque-service/internal/httputils"
)

type DebitHandler struct {
	service DebitService
}

func NewDebitHandler(service DebitService) *DebitHandler {
	return &DebitHandler{service: service}
}

func (h *DebitHandler) Debit(w http.ResponseWriter, r *http.Request) {
	var input DebitInput
	if err := httputils.DecodeJSON(w, r, &input); err != nil {
		httputils.RespondError(w, http.StatusBadRequest, "Não foi possível processar os dados enviados.")
		return
	}

	result, err := h.service.Debit(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			httputils.RespondError(w, http.StatusUnprocessableEntity, "Referência e itens válidos são obrigatórios.")
		case errors.Is(err, ErrIdempotencyConflict):
			httputils.RespondError(w, http.StatusConflict, "A referência já foi utilizada com itens diferentes.")
		case errors.Is(err, ErrProductNotFound):
			httputils.RespondError(w, http.StatusNotFound, "Produto não encontrado.")
		case errors.Is(err, ErrInsufficientStock):
			httputils.RespondError(w, http.StatusConflict, "Saldo insuficiente para concluir a operação.")
		default:
			slog.Error("failed to debit stock", "error", err)
			httputils.RespondError(w, http.StatusInternalServerError, "Erro inesperado no servidor.")
		}
		return
	}

	status := http.StatusCreated
	if result.AlreadyProcessed {
		status = http.StatusOK
	}
	httputils.Respond(w, status, result)
}
