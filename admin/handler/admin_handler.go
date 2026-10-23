package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"mercatus/admin/middleware"
	"mercatus/admin/model"
	"mercatus/admin/service"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	admin     *service.AdminService
	jwtSecret string
}

func NewAdminHandler(a *service.AdminService, jwtSecret string) *AdminHandler {
	return &AdminHandler{admin: a, jwtSecret: jwtSecret}
}

func (h *AdminHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(h.jwtSecret), middleware.RequireRole("admin"))
	r.Post("/moderate", h.Moderate)
	r.Get("/audit", h.Audit)
	r.Get("/stats", h.Stats)
	return r
}

func writeJSON(w http.ResponseWriter, s int, d any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(s)
	_ = json.NewEncoder(w).Encode(d)
}

func internal(w http.ResponseWriter, r *http.Request, err error, msg string) {
	slog.ErrorContext(r.Context(), msg, "error", err)
	writeErr(w, http.StatusInternalServerError, msg)
}
func writeErr(w http.ResponseWriter, s int, m string) { writeJSON(w, s, map[string]string{"error": m}) }

func (h *AdminHandler) Moderate(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	var req model.ModerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	a, err := h.admin.Moderate(r.Context(), uid, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTarget), errors.Is(err, service.ErrInvalidAction), errors.Is(err, service.ErrReasonRequired):
			writeErr(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrNotFound):
			writeErr(w, http.StatusNotFound, err.Error())
		default:
			internal(w, r, err, "could not moderate")
		}
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (h *AdminHandler) Audit(w http.ResponseWriter, r *http.Request) {
	items, err := h.admin.AuditLog(r.Context())
	if err != nil {
		internal(w, r, err, "could not read audit log")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	s, err := h.admin.Stats(r.Context())
	if err != nil {
		internal(w, r, err, "could not read stats")
		return
	}
	writeJSON(w, http.StatusOK, s)
}
