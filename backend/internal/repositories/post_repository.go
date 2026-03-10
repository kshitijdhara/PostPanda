package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/kshitijdhara/blog/internal/domain"
)

var ErrPostNotFound = errors.New("post not found")

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) (*domain.Post, error)
	GetBySlug(ctx context.Context, slug string, userID int64) (*domain.PostWithAuthor, error)
	Update(ctx context.Context, post *domain.Post) (*domain.Post, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params domain.PostListParams) ([]domain.PostWithAuthor, int64, error)
	ListByAuthor(ctx context.Context, authorID int64, params domain.PostListParams) ([]domain.PostWithAuthor, int64, error)
	ListDraftsByAuthor(ctx context.Context, authorID int64) ([]domain.PostWithAuthor, error)
	ToggleLike(ctx context.Context, userID, postID int64) (liked bool, count int64, err error)
	ListLikedByUser(ctx context.Context, userID int64) ([]domain.PostWithAuthor, error)
	ToggleBookmark(ctx context.Context, userID, postID int64) (bookmarked bool, err error)
	ListBookmarkedByUser(ctx context.Context, userID int64) ([]domain.PostWithAuthor, error)
}

type postRepository struct {
	db *sqlx.DB
}

func NewPostRepository(db *sqlx.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	query := `
		INSERT INTO posts (title, slug, content, excerpt, status, author_id, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, slug, content, excerpt, status, author_id, created_at, updated_at, published_at`

	err := r.db.QueryRowxContext(ctx, query,
		post.Title, post.Slug, post.Content, post.Excerpt, post.Status, post.AuthorID, post.PublishedAt,
	).StructScan(post)
	return post, err
}

func (r *postRepository) GetBySlug(ctx context.Context, slug string, userID int64) (*domain.PostWithAuthor, error) {
	var post domain.PostWithAuthor
	query := `
		SELECT p.*,
		       u.username AS author_username, u.display_name AS author_display_name, u.avatar_url AS author_avatar_url,
		       (SELECT COUNT(*) FROM post_likes WHERE post_id = p.id) AS like_count,
		       EXISTS(SELECT 1 FROM post_likes WHERE post_id = p.id AND user_id = $2) AS liked_by_user,
		       EXISTS(SELECT 1 FROM post_bookmarks WHERE post_id = p.id AND user_id = $2) AS bookmarked_by_user
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.slug = $1`

	err := r.db.GetContext(ctx, &post, query, slug, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPostNotFound
	}
	return &post, err
}

func (r *postRepository) Update(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	query := `
		UPDATE posts SET title = $1, slug = $2, content = $3, excerpt = $4, status = $5, updated_at = NOW(), published_at = $6
		WHERE id = $7
		RETURNING id, title, slug, content, excerpt, status, author_id, created_at, updated_at, published_at`

	err := r.db.QueryRowxContext(ctx, query,
		post.Title, post.Slug, post.Content, post.Excerpt, post.Status, post.PublishedAt, post.ID,
	).StructScan(post)
	return post, err
}

func (r *postRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (r *postRepository) List(ctx context.Context, params domain.PostListParams) ([]domain.PostWithAuthor, int64, error) {
	baseQuery := `
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.status = 'published'`
	args := []interface{}{}
	argIdx := 1

	if params.Search != "" {
		baseQuery += fmt.Sprintf(" AND to_tsvector('english', p.title || ' ' || p.content) @@ to_tsquery('english', $%d)", argIdx)
		args = append(args, buildPrefixQuery(params.Search))
		argIdx++
	}

	var total int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	args = append(args, params.UserID)
	selectQuery := fmt.Sprintf(`
		SELECT p.*,
		       u.username AS author_username, u.display_name AS author_display_name, u.avatar_url AS author_avatar_url,
		       (SELECT COUNT(*) FROM post_likes WHERE post_id = p.id) AS like_count,
		       EXISTS(SELECT 1 FROM post_likes WHERE post_id = p.id AND user_id = $%d) AS liked_by_user,
		       EXISTS(SELECT 1 FROM post_bookmarks WHERE post_id = p.id AND user_id = $%d) AS bookmarked_by_user
		%s
		ORDER BY p.published_at DESC NULLS LAST, p.created_at DESC
		LIMIT $%d OFFSET $%d`, argIdx, argIdx, baseQuery, argIdx+1, argIdx+2)
	argIdx++

	args = append(args, params.PerPage, (params.Page-1)*params.PerPage)

	var posts []domain.PostWithAuthor
	err = r.db.SelectContext(ctx, &posts, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	if posts == nil {
		posts = []domain.PostWithAuthor{}
	}
	return posts, total, nil
}

func (r *postRepository) ListByAuthor(ctx context.Context, authorID int64, params domain.PostListParams) ([]domain.PostWithAuthor, int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM posts WHERE author_id = $1 AND status = 'published'", authorID)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT p.*, u.username AS author_username, u.display_name AS author_display_name, u.avatar_url AS author_avatar_url
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.author_id = $1 AND p.status = 'published'
		ORDER BY p.published_at DESC NULLS LAST, p.created_at DESC
		LIMIT $2 OFFSET $3`

	var posts []domain.PostWithAuthor
	err = r.db.SelectContext(ctx, &posts, query, authorID, params.PerPage, (params.Page-1)*params.PerPage)
	if posts == nil {
		posts = []domain.PostWithAuthor{}
	}
	return posts, total, err
}

func (r *postRepository) ListDraftsByAuthor(ctx context.Context, authorID int64) ([]domain.PostWithAuthor, error) {
	query := `
		SELECT p.*, u.username AS author_username, u.display_name AS author_display_name, u.avatar_url AS author_avatar_url
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.author_id = $1 AND p.status = 'draft'
		ORDER BY p.updated_at DESC`

	var posts []domain.PostWithAuthor
	err := r.db.SelectContext(ctx, &posts, query, authorID)
	if posts == nil {
		posts = []domain.PostWithAuthor{}
	}
	return posts, err
}

func (r *postRepository) ToggleLike(ctx context.Context, userID, postID int64) (bool, int64, error) {
	// Try insert; if it already exists, delete instead
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		"SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = $1 AND post_id = $2)", userID, postID)
	if err != nil {
		return false, 0, err
	}

	if exists {
		_, err = r.db.ExecContext(ctx, "DELETE FROM post_likes WHERE user_id = $1 AND post_id = $2", userID, postID)
	} else {
		_, err = r.db.ExecContext(ctx, "INSERT INTO post_likes (user_id, post_id) VALUES ($1, $2)", userID, postID)
	}
	if err != nil {
		return false, 0, err
	}

	var count int64
	err = r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM post_likes WHERE post_id = $1", postID)
	return !exists, count, err
}

func (r *postRepository) ListLikedByUser(ctx context.Context, userID int64) ([]domain.PostWithAuthor, error) {
	query := `
		SELECT p.*, u.username AS author_username, u.display_name AS author_display_name, u.avatar_url AS author_avatar_url,
			(SELECT COUNT(*) FROM post_likes pl2 WHERE pl2.post_id = p.id) AS like_count,
			TRUE AS liked_by_user
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN post_likes pl ON pl.post_id = p.id AND pl.user_id = $1
		WHERE p.status = 'published'
		ORDER BY pl.created_at DESC`

	var posts []domain.PostWithAuthor
	err := r.db.SelectContext(ctx, &posts, query, userID)
	if posts == nil {
		posts = []domain.PostWithAuthor{}
	}
	return posts, err
}

func (r *postRepository) ToggleBookmark(ctx context.Context, userID, postID int64) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		"SELECT EXISTS(SELECT 1 FROM post_bookmarks WHERE user_id = $1 AND post_id = $2)", userID, postID)
	if err != nil {
		return false, err
	}

	if exists {
		_, err = r.db.ExecContext(ctx, "DELETE FROM post_bookmarks WHERE user_id = $1 AND post_id = $2", userID, postID)
	} else {
		_, err = r.db.ExecContext(ctx, "INSERT INTO post_bookmarks (user_id, post_id) VALUES ($1, $2)", userID, postID)
	}
	return !exists, err
}

func (r *postRepository) ListBookmarkedByUser(ctx context.Context, userID int64) ([]domain.PostWithAuthor, error) {
	query := `
		SELECT p.*, u.username AS author_username, u.display_name AS author_display_name, u.avatar_url AS author_avatar_url,
			(SELECT COUNT(*) FROM post_likes pl2 WHERE pl2.post_id = p.id) AS like_count,
			EXISTS(SELECT 1 FROM post_likes WHERE post_id = p.id AND user_id = $1) AS liked_by_user,
			TRUE AS bookmarked_by_user
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN post_bookmarks pb ON pb.post_id = p.id AND pb.user_id = $1
		WHERE p.status = 'published'
		ORDER BY pb.created_at DESC`

	var posts []domain.PostWithAuthor
	err := r.db.SelectContext(ctx, &posts, query, userID)
	if posts == nil {
		posts = []domain.PostWithAuthor{}
	}
	return posts, err
}

// buildPrefixQuery converts a search string into a prefix-aware tsquery.
// "draft 1" → "draft:* & 1:*" so partial words like "dr" match "draft".
func buildPrefixQuery(search string) string {
	words := strings.Fields(search)
	if len(words) == 0 {
		return search
	}
	for i, w := range words {
		words[i] = w + ":*"
	}
	return strings.Join(words, " & ")
}
