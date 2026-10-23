package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"mercatus/notification/middleware"
	"mercatus/notification/model"
	"mercatus/notification/service"

	"github.com/go-chi/chi/v5"
)

type NotificationHandler struct {
	notifications *service.NotificationService
	jwtSecret     string
}

func NewNotificationHandler(n *service.NotificationService, jwtSecret string) *NotificationHandler {
	return &NotificationHandler{notifications: n, jwtSecret: jwtSecret}
}

func (h *NotificationHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth(h.jwtSecret))
	r.Get("/", h.List)
	r.Get("/unread-count", h.Unread)
	r.Post("/read-all", h.MarkAllRead)
	r.Patch("/{id}/read", h.MarkRead)
	r.With(middleware.RequireRole("admin")).Post("/send", h.Send)
	return r
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func internal(w http.ResponseWriter, r *http.Request, err error, msg string) {
	slog.ErrorContext(r.Context(), msg, "error", err)
	writeError(w, http.StatusInternalServerError, msg)
}

func (h *NotificationHandler) Send(w http.ResponseWriter, r *http.Request) {
	var req model.SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	n, err := h.notifications.Send(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		internal(w, r, err, "could not send notification")
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	items, err := h.notifications.ListForUser(r.Context(), uid)
	if err != nil {
		internal(w, r, err, "could not list notifications")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *NotificationHandler) Unread(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	n, err := h.notifications.UnreadCount(r.Context(), uid)
	if err != nil {
		internal(w, r, err, "could not count notifications")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"unread": n})
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.notifications.MarkRead(r.Context(), id, uid); err != nil {
		writeError(w, http.StatusNotFound, "notification not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "marked read"})
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	if err := h.notifications.MarkAllRead(r.Context(), uid); err != nil {
		internal(w, r, err, "could not update notifications")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "all marked read"})
}
