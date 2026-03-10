package domain

import "time"

type Comment struct {
	ID        int64     `db:"id" json:"id"`
	Content   string    `db:"content" json:"content"`
	PostID    int64     `db:"post_id" json:"post_id"`
	AuthorID  int64     `db:"author_id" json:"author_id"`
	ParentID  *int64    `db:"parent_id" json:"parent_id,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CommentWithAuthor struct {
	Comment
	AuthorUsername    string  `db:"author_username" json:"author_username"`
	AuthorDisplayName string  `db:"author_display_name" json:"author_display_name"`
	PostSlug          string  `db:"post_slug" json:"post_slug,omitempty"`
	PostTitle         string  `db:"post_title" json:"post_title,omitempty"`
	Upvotes           int64   `db:"upvotes" json:"upvotes"`
	Downvotes         int64   `db:"downvotes" json:"downvotes"`
	UserVote          *int    `db:"user_vote" json:"user_vote,omitempty"`
}

type CreateCommentRequest struct {
	Content  string `json:"content"`
	ParentID *int64 `json:"parent_id,omitempty"`
}

type VoteCommentRequest struct {
	Value int `json:"value"` // 1, -1, or 0 to remove
}

type VoteCommentResponse struct {
	Upvotes   int64 `json:"upvotes"`
	Downvotes int64 `json:"downvotes"`
	UserVote  *int  `json:"user_vote,omitempty"`
}
