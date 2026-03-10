package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/kshitijdhara/blog/internal/domain"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) (*domain.Comment, error)
	ListByPostID(ctx context.Context, postID, userID int64) ([]domain.CommentWithAuthor, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*domain.Comment, error)
	VoteComment(ctx context.Context, userID, commentID int64, value int) (int64, int64, *int, error)
	ListByAuthor(ctx context.Context, authorID int64) ([]domain.CommentWithAuthor, error)
}

type commentRepository struct {
	db *sqlx.DB
}

func NewCommentRepository(db *sqlx.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *domain.Comment) (*domain.Comment, error) {
	query := `
		INSERT INTO comments (content, post_id, author_id, parent_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, content, post_id, author_id, parent_id, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		comment.Content, comment.PostID, comment.AuthorID, comment.ParentID,
	).StructScan(comment)
	return comment, err
}

func (r *commentRepository) ListByPostID(ctx context.Context, postID, userID int64) ([]domain.CommentWithAuthor, error) {
	query := `
		SELECT c.*,
		       u.username AS author_username, u.display_name AS author_display_name,
		       COALESCE(SUM(CASE WHEN cv.value = 1  THEN 1 ELSE 0 END), 0) AS upvotes,
		       COALESCE(SUM(CASE WHEN cv.value = -1 THEN 1 ELSE 0 END), 0) AS downvotes,
		       (SELECT value FROM comment_votes WHERE user_id = $2 AND comment_id = c.id) AS user_vote
		FROM comments c
		JOIN users u ON u.id = c.author_id
		LEFT JOIN comment_votes cv ON cv.comment_id = c.id
		WHERE c.post_id = $1
		GROUP BY c.id, u.id
		ORDER BY c.created_at ASC`

	var comments []domain.CommentWithAuthor
	err := r.db.SelectContext(ctx, &comments, query, postID, userID)
	if comments == nil {
		comments = []domain.CommentWithAuthor{}
	}
	return comments, err
}

func (r *commentRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM comments WHERE id = $1", id)
	return err
}

func (r *commentRepository) GetByID(ctx context.Context, id int64) (*domain.Comment, error) {
	var comment domain.Comment
	err := r.db.GetContext(ctx, &comment, "SELECT * FROM comments WHERE id = $1", id)
	return &comment, err
}

func (r *commentRepository) VoteComment(ctx context.Context, userID, commentID int64, value int) (int64, int64, *int, error) {
	if value == 0 {
		// Remove vote
		_, err := r.db.ExecContext(ctx,
			"DELETE FROM comment_votes WHERE user_id = $1 AND comment_id = $2", userID, commentID)
		if err != nil {
			return 0, 0, nil, err
		}
	} else {
		// Upsert vote
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO comment_votes (user_id, comment_id, value) VALUES ($1, $2, $3)
			ON CONFLICT (user_id, comment_id) DO UPDATE SET value = $3`,
			userID, commentID, value)
		if err != nil {
			return 0, 0, nil, err
		}
	}

	var upvotes, downvotes int64
	err := r.db.GetContext(ctx, &upvotes,
		"SELECT COUNT(*) FROM comment_votes WHERE comment_id = $1 AND value = 1", commentID)
	if err != nil {
		return 0, 0, nil, err
	}
	err = r.db.GetContext(ctx, &downvotes,
		"SELECT COUNT(*) FROM comment_votes WHERE comment_id = $1 AND value = -1", commentID)
	if err != nil {
		return 0, 0, nil, err
	}

	var userVote *int
	if value != 0 {
		v := value
		userVote = &v
	}
	return upvotes, downvotes, userVote, nil
}

func (r *commentRepository) ListByAuthor(ctx context.Context, authorID int64) ([]domain.CommentWithAuthor, error) {
	query := `
		SELECT c.*,
		       u.username AS author_username, u.display_name AS author_display_name,
		       p.slug AS post_slug, p.title AS post_title,
		       COALESCE(SUM(CASE WHEN cv.value = 1  THEN 1 ELSE 0 END), 0) AS upvotes,
		       COALESCE(SUM(CASE WHEN cv.value = -1 THEN 1 ELSE 0 END), 0) AS downvotes,
		       NULL::int AS user_vote
		FROM comments c
		JOIN users u ON u.id = c.author_id
		JOIN posts p ON p.id = c.post_id
		LEFT JOIN comment_votes cv ON cv.comment_id = c.id
		WHERE c.author_id = $1
		GROUP BY c.id, u.id, p.id
		ORDER BY c.created_at DESC`

	var comments []domain.CommentWithAuthor
	err := r.db.SelectContext(ctx, &comments, query, authorID)
	if comments == nil {
		comments = []domain.CommentWithAuthor{}
	}
	return comments, err
}
