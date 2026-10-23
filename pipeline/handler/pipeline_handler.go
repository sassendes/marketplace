package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"mercatus/pipeline/middleware"
	"mercatus/pipeline/model"
	"mercatus/pipeline/service"

	"github.com/go-chi/chi/v5"
)

type PipelineHandler struct {
	checkout  *service.CheckoutService
	jwtSecret string
}

func NewPipelineHandler(c *service.CheckoutService, jwtSecret string) *PipelineHandler {
	return &PipelineHandler{checkout: c, jwtSecret: jwtSecret}
}

func (h *PipelineHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(h.jwtSecret))
	r.Post("/", h.StartCheckout)
	r.Get("/{id}", h.GetSaga)
	return r
}

func writeJSON(w http.ResponseWriter, s int, d any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(s)
	_ = json.NewEncoder(w).Encode(d)
}

func writeErr(w http.ResponseWriter, s int, m string) { writeJSON(w, s, map[string]string{"error": m}) }

func (h *PipelineHandler) StartCheckout(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	var req model.StartCheckoutRequest
	if r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}
	if k := r.Header.Get("Idempotency-Key"); k != "" {
		req.IdempotencyKey = k
	}
	res, err := h.checkout.Checkout(r.Context(), uid, req.IdempotencyKey)
	switch {
	case err == nil:
		code := http.StatusCreated
		if res.Status != "completed" {
			code = http.StatusAccepted
		}
		writeJSON(w, code, res)
	case errors.Is(err, service.ErrInvalidKey):
		writeErr(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrEmptyCart):
		writeErr(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrCheckoutRunning):
		writeErr(w, http.StatusConflict, err.Error())
	case res != nil:
		slog.WarnContext(r.Context(), "checkout failed", "saga_id", res.SagaID, "order_id", res.OrderID, "error", err)
		writeJSON(w, http.StatusConflict, res)
	default:
		slog.ErrorContext(r.Context(), "checkout error", "error", err)
		writeErr(w, http.StatusInternalServerError, "checkout could not be completed")
	}
}

func (h *PipelineHandler) GetSaga(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	saga, steps, err := h.checkout.Get(r.Context(), uid, id)
	if err != nil {
		if errors.Is(err, service.ErrSagaNotFound) {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		slog.ErrorContext(r.Context(), "could not read checkout", "error", err)
		writeErr(w, http.StatusInternalServerError, "could not read checkout")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saga": saga, "steps": steps})
}
