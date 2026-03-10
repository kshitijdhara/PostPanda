package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kshitijdhara/blog/internal/domain"
	"github.com/kshitijdhara/blog/internal/middleware"
	"github.com/kshitijdhara/blog/internal/repositories"
	"github.com/kshitijdhara/blog/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.With(middleware.Auth(h.authService)).Get("/me", h.Me)
	return r
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "username, email, password, and display_name are required")
		return
	}

	user, token, err := h.authService.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, repositories.ErrUserExists) {
			writeError(w, http.StatusConflict, "CONFLICT", "username or email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to register")
		return
	}

	setAuthCookie(w, token)
	writeJSON(w, http.StatusCreated, domain.DataResponse{Data: user.ToResponse()})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "username and password are required")
		return
	}

	user, token, err := h.authService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid username or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to login")
		return
	}

	setAuthCookie(w, token)
	writeJSON(w, http.StatusOK, domain.DataResponse{Data: user.ToResponse()})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get user")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: user.ToResponse()})
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(24 * time.Hour / time.Second),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
