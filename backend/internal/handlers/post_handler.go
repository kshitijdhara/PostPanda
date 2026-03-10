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

type PostHandler struct {
	postService *services.PostService
	authService *services.AuthService
}

func NewPostHandler(postService *services.PostService, authService *services.AuthService) *PostHandler {
	return &PostHandler{postService: postService, authService: authService}
}

func (h *PostHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Public routes
	r.With(middleware.OptionalAuth(h.authService)).Get("/", h.List)
	r.With(middleware.OptionalAuth(h.authService)).Get("/{slug}", h.GetBySlug)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(h.authService))
		r.Get("/drafts/mine", h.ListDrafts)
		r.Post("/", h.Create)
		r.Put("/{slug}", h.Update)
		r.Delete("/{slug}", h.Delete)
	})

	return r
}

func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	params := domain.PostListParams{
		Page:    queryInt(r, "page", 1),
		PerPage: queryInt(r, "per_page", 20),
		Search:  r.URL.Query().Get("search"),
		UserID:  userID,
	}

	result, err := h.postService.List(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list posts")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *PostHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	userID, _ := middleware.GetUserID(r.Context())

	post, err := h.postService.GetBySlug(r.Context(), slug, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get post")
		return
	}

	// Only show drafts to the author
	if post.Status == "draft" {
		if userID == 0 || userID != post.AuthorID {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "post not found")
			return
		}
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: post})
}

func (h *PostHandler) ToggleLike(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	liked, count, err := h.postService.ToggleLike(r.Context(), userID, slug)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to toggle like")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: domain.ToggleLikeResponse{Liked: liked, LikeCount: count}})
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var req domain.CreatePostRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Title == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "title and content are required")
		return
	}

	post, err := h.postService.Create(r.Context(), req, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create post")
		return
	}

	writeJSON(w, http.StatusCreated, domain.DataResponse{Data: post})
}

func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	var req domain.UpdatePostRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	post, err := h.postService.Update(r.Context(), slug, req, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "post not found")
			return
		}
		if errors.Is(err, services.ErrNotAuthor) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "you can only edit your own posts")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update post")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: post})
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	err := h.postService.Delete(r.Context(), slug, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "post not found")
			return
		}
		if errors.Is(err, services.ErrNotAuthor) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "you can only delete your own posts")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete post")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PostHandler) ToggleBookmark(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	bookmarked, err := h.postService.ToggleBookmark(r.Context(), userID, slug)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to toggle bookmark")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: domain.ToggleBookmarkResponse{Bookmarked: bookmarked}})
}

func (h *PostHandler) ListDrafts(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	posts, err := h.postService.ListDraftsByAuthor(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list drafts")
		return
	}

	writeJSON(w, http.StatusOK, domain.DataResponse{Data: posts})
}
