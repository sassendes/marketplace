package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"marketplace/auth/middleware"
	"marketplace/auth/model"
	"marketplace/auth/service"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	auth      *service.AuthService
	jwtMgr    *service.JWTManager
	jwtSecret string
}

func NewAuthHandler(auth *service.AuthService, jwtMgr *service.JWTManager, jwtSecret string) *AuthHandler {
	return &AuthHandler{auth: auth, jwtMgr: jwtMgr, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(h.jwtSecret))
		r.Get("/me", h.Me)
		r.Post("/verify-email", h.VerifyEmail)
	})
	return r
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func internal(w http.ResponseWriter, r *http.Request, err error, msg string) {
	slog.ErrorContext(r.Context(), msg, "error", err)
	writeError(w, http.StatusInternalServerError, msg)
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if !decode(w, r, &req) {
		return
	}
	user, err := h.auth.Register(r.Context(), &req)
	if err != nil {
		var ve *service.ValidationError
		switch {
		case errors.As(err, &ve):
			writeError(w, http.StatusBadRequest, ve.Msg)
		case errors.Is(err, service.ErrEmailTaken):
			writeError(w, http.StatusConflict, err.Error())
		default:
			internal(w, r, err, "could not register")
		}
		return
	}
	resp, err := h.auth.Login(r.Context(), &model.LoginRequest{Email: user.Email, Password: req.Password}, h.jwtMgr)
	if err != nil {
		internal(w, r, err, "registered but could not sign in")
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if !decode(w, r, &req) {
		return
	}
	resp, err := h.auth.Login(r.Context(), &req, h.jwtMgr)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		internal(w, r, err, "could not login")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if !decode(w, r, &req) {
		return
	}
	resp, err := h.auth.Refresh(r.Context(), req.RefreshToken, h.jwtMgr)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid refresh token")
			return
		}
		internal(w, r, err, "could not refresh session")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if !decode(w, r, &req) {
		return
	}
	if err := h.auth.Logout(r.Context(), req.RefreshToken); err != nil {
		internal(w, r, err, "could not logout")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.auth.Me(r.Context(), uid)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		internal(w, r, err, "could not load user")
		return
	}
	writeJSON(w, http.StatusOK, user.ToPublic())
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.UserIDFromContext(r.Context())
	var req model.VerifyEmailRequest
	if !decode(w, r, &req) {
		return
	}
	if err := h.auth.VerifyEmail(r.Context(), uid, req.Code); err != nil {
		if errors.Is(err, service.ErrInvalidCode) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		internal(w, r, err, "could not verify email")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "email verified"})
}
