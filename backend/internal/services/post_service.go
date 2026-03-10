package services

import (
	"context"
	"errors"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/kshitijdhara/blog/internal/domain"
	"github.com/kshitijdhara/blog/internal/repositories"
)

var (
	ErrNotAuthor   = errors.New("not the author of this post")
	ErrPostNotFound = repositories.ErrPostNotFound
)

type PostService struct {
	postRepo repositories.PostRepository
}

func NewPostService(postRepo repositories.PostRepository) *PostService {
	return &PostService{postRepo: postRepo}
}

func (s *PostService) Create(ctx context.Context, req domain.CreatePostRequest, authorID int64) (*domain.Post, error) {
	status := req.Status
	if status == "" {
		status = "draft"
	}

	var publishedAt *time.Time
	if status == "published" {
		now := time.Now()
		publishedAt = &now
	}

	post := &domain.Post{
		Title:       req.Title,
		Slug:        generateSlug(req.Title),
		Content:     req.Content,
		Excerpt:     generateExcerpt(req.Content),
		Status:      status,
		AuthorID:    authorID,
		PublishedAt: publishedAt,
	}

	return s.postRepo.Create(ctx, post)
}

func (s *PostService) GetBySlug(ctx context.Context, slug string, userID int64) (*domain.PostWithAuthor, error) {
	return s.postRepo.GetBySlug(ctx, slug, userID)
}

func (s *PostService) Update(ctx context.Context, slug string, req domain.UpdatePostRequest, userID int64) (*domain.Post, error) {
	existing, err := s.postRepo.GetBySlug(ctx, slug, 0)
	if err != nil {
		return nil, err
	}
	if existing.AuthorID != userID {
		return nil, ErrNotAuthor
	}

	post := &existing.Post

	if req.Title != nil {
		post.Title = *req.Title
		post.Slug = generateSlug(*req.Title)
	}
	if req.Content != nil {
		post.Content = *req.Content
		post.Excerpt = generateExcerpt(*req.Content)
	}
	if req.Status != nil {
		if *req.Status == "published" && post.Status == "draft" {
			now := time.Now()
			post.PublishedAt = &now
		}
		post.Status = *req.Status
	}

	return s.postRepo.Update(ctx, post)
}

func (s *PostService) Delete(ctx context.Context, slug string, userID int64) error {
	existing, err := s.postRepo.GetBySlug(ctx, slug, 0)
	if err != nil {
		return err
	}
	if existing.AuthorID != userID {
		return ErrNotAuthor
	}
	return s.postRepo.Delete(ctx, existing.ID)
}

func (s *PostService) List(ctx context.Context, params domain.PostListParams) (*domain.PostListResponse, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 50 {
		params.PerPage = 20
	}

	posts, total, err := s.postRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return &domain.PostListResponse{
		Data: posts,
		Meta: domain.PaginationMeta{
			Page:    params.Page,
			PerPage: params.PerPage,
			Total:   total,
		},
	}, nil
}

func (s *PostService) ListDraftsByAuthor(ctx context.Context, authorID int64) ([]domain.PostWithAuthor, error) {
	return s.postRepo.ListDraftsByAuthor(ctx, authorID)
}

func (s *PostService) ToggleLike(ctx context.Context, userID int64, slug string) (bool, int64, error) {
	post, err := s.postRepo.GetBySlug(ctx, slug, 0)
	if err != nil {
		return false, 0, err
	}
	return s.postRepo.ToggleLike(ctx, userID, post.ID)
}

func (s *PostService) ListLikedByUser(ctx context.Context, userID int64) ([]domain.PostWithAuthor, error) {
	return s.postRepo.ListLikedByUser(ctx, userID)
}

func (s *PostService) ToggleBookmark(ctx context.Context, userID int64, slug string) (bool, error) {
	post, err := s.postRepo.GetBySlug(ctx, slug, 0)
	if err != nil {
		return false, err
	}
	return s.postRepo.ToggleBookmark(ctx, userID, post.ID)
}

func (s *PostService) ListBookmarkedByUser(ctx context.Context, userID int64) ([]domain.PostWithAuthor, error) {
	return s.postRepo.ListBookmarkedByUser(ctx, userID)
}

func (s *PostService) ListByAuthor(ctx context.Context, authorID int64, params domain.PostListParams) ([]domain.PostWithAuthor, int64, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 50 {
		params.PerPage = 20
	}
	return s.postRepo.ListByAuthor(ctx, authorID, params)
}

var nonAlphanumRegex = regexp.MustCompile(`[^a-z0-9]+`)

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
			return r
		}
		return -1
	}, slug)
	slug = nonAlphanumRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	suffix := make([]byte, 6)
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	for i := range suffix {
		suffix[i] = chars[rand.Intn(len(chars))]
	}
	return slug + "-" + string(suffix)
}

func generateExcerpt(content string) *string {
	// Strip markdown-ish characters for a plain text excerpt
	stripped := content
	stripped = regexp.MustCompile(`[#*_\[\]()>~` + "`" + `]`).ReplaceAllString(stripped, "")
	stripped = strings.TrimSpace(stripped)

	if len(stripped) > 200 {
		stripped = stripped[:200] + "..."
	}
	if stripped == "" {
		return nil
	}
	return &stripped
}
