package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kshitijdhara/blog/internal/domain"
	"github.com/kshitijdhara/blog/internal/middleware"
	"github.com/kshitijdhara/blog/internal/repositories"
	"github.com/kshitijdhara/blog/internal/services"
)

type UserHandler struct {
	authService    *services.AuthService
	postService    *services.PostService
	commentService *services.CommentService
}

func NewUserHandler(authService *services.AuthService, postService *services.PostService, commentService *services.CommentService) *UserHandler {
	return &UserHandler{authService: authService, postService: postService, commentService: commentService}
}

func (h *UserHandler) Routes(authService *services.AuthService) chi.Router {
	r := chi.NewRouter()
	r.Get("/{username}", h.GetProfile)
	r.Get("/{username}/posts", h.GetUserPosts)
	r.Get("/{username}/comments", h.GetUserComments)
	r.With(middleware.Auth(authService)).Put("/me", h.UpdateProfile)
	r.With(middleware.Auth(authService)).Put("/me/password", h.ChangePassword)
	r.With(middleware.Auth(authService)).Get("/me/liked-posts", h.GetLikedPosts)
	r.With(middleware.Auth(authService)).Get("/me/bookmarked-posts", h.GetBookmarkedPosts)
	r.With(middleware.Auth(authService)).Get("/me/comments", h.GetMyComments)
	return r
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := h.authService.GetUserByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get user")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: user.ToResponse()})
}

func (h *UserHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := h.authService.GetUserByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get user")
		return
	}

	params := domain.PostListParams{
		Page:    queryInt(r, "page", 1),
		PerPage: queryInt(r, "per_page", 20),
	}

	posts, total, err := h.postService.ListByAuthor(r.Context(), user.ID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get posts")
		return
	}

	writeJSON(w, http.StatusOK, domain.PostListResponse{
		Data: posts,
		Meta: domain.PaginationMeta{
			Page:    params.Page,
			PerPage: params.PerPage,
			Total:   total,
		},
	})
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req domain.UpdateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.DisplayName == "" && req.Username == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "at least display_name or username is required")
		return
	}

	user, err := h.authService.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, repositories.ErrUserExists) {
			writeError(w, http.StatusConflict, "CONFLICT", "username already taken")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update profile")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: user.ToResponse()})
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req domain.ChangePasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "current_password and new_password are required")
		return
	}

	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "new password must be at least 8 characters")
		return
	}

	if err := h.authService.ChangePassword(r.Context(), userID, req); err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "current password is incorrect")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to change password")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetLikedPosts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	posts, err := h.postService.ListLikedByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get liked posts")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: posts})
}

func (h *UserHandler) GetMyComments(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	comments, err := h.commentService.ListByAuthor(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get comments")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: comments})
}

func (h *UserHandler) GetBookmarkedPosts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	posts, err := h.postService.ListBookmarkedByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get bookmarked posts")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: posts})
}

func (h *UserHandler) GetUserComments(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := h.authService.GetUserByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get user")
		return
	}

	comments, err := h.commentService.ListByAuthor(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get comments")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: comments})
}
