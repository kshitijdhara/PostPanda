package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kshitijdhara/blog/internal/domain"
	"github.com/kshitijdhara/blog/internal/middleware"
	"github.com/kshitijdhara/blog/internal/repositories"
	"github.com/kshitijdhara/blog/internal/services"
)

type CommentHandler struct {
	commentService *services.CommentService
	authService    *services.AuthService
}

func NewCommentHandler(commentService *services.CommentService, authService *services.AuthService) *CommentHandler {
	return &CommentHandler{commentService: commentService, authService: authService}
}

func (h *CommentHandler) PostRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListByPost)
	r.With(middleware.Auth(h.authService)).Post("/", h.Create)
	return r
}

func (h *CommentHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.With(middleware.Auth(h.authService)).Delete("/{id}", h.Delete)
	return r
}

func (h *CommentHandler) ListByPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	userID, _ := middleware.GetUserID(r.Context())
	slog.Info("list comments", "slug", slug, "path", r.URL.Path)

	comments, err := h.commentService.ListBySlug(r.Context(), slug, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list comments")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: comments})
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	if slug == "" {
		slog.Error("slug not found in URL params", "path", r.URL.Path)
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "slug is required")
		return
	}

	var req domain.CreateCommentRequest
	if err := decodeJSON(r, &req); err != nil {
		slog.Error("comment decode error", "error", err, "slug", slug, "contentLength", r.ContentLength)
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "content is required")
		return
	}

	comment, err := h.commentService.Create(r.Context(), slug, req, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create comment")
		return
	}

	writeJSON(w, http.StatusCreated, domain.DataResponse{Data: comment})
}

func (h *CommentHandler) Vote(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid comment id")
		return
	}

	var req domain.VoteCommentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	if req.Value != 1 && req.Value != -1 && req.Value != 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "value must be 1, -1, or 0")
		return
	}

	upvotes, downvotes, userVote, err := h.commentService.VoteComment(r.Context(), userID, id, req.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to vote")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: domain.VoteCommentResponse{
		Upvotes:   upvotes,
		Downvotes: downvotes,
		UserVote:  userVote,
	}})
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid comment id")
		return
	}

	err = h.commentService.Delete(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrNotCommentAuthor) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "you can only delete your own comments")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete comment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
