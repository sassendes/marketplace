package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"mercatus/payment/middleware"
	"mercatus/payment/service"

	"github.com/go-chi/chi/v5"
)

type PaymentHandler struct {
	payments  *service.PaymentService
	jwtSecret string
}

func NewPaymentHandler(p *service.PaymentService, jwtSecret string) *PaymentHandler {
	return &PaymentHandler{payments: p, jwtSecret: jwtSecret}
}

func (h *PaymentHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(h.jwtSecret), middleware.RequireRole("admin"))
	r.Get("/ledger/balance/{account}", h.Balance)
	r.Get("/ledger/integrity", h.Integrity)
	return r
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *PaymentHandler) Balance(w http.ResponseWriter, r *http.Request) {
	account := chi.URLParam(r, "account")
	bal, err := h.payments.LedgerBalance(r.Context(), account)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAccount) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		slog.ErrorContext(r.Context(), "could not read balance", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read balance"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"account": account, "balance_cents": bal})
}

func (h *PaymentHandler) Integrity(w http.ResponseWriter, r *http.Request) {
	total, err := h.payments.LedgerTotal(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "could not read ledger", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not read ledger"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ledger_sum_cents": total, "balanced": total == 0})
}
