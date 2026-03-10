package domain

import "time"

type Post struct {
	ID          int64      `db:"id" json:"id"`
	Title       string     `db:"title" json:"title"`
	Slug        string     `db:"slug" json:"slug"`
	Content     string     `db:"content" json:"content"`
	Excerpt     *string    `db:"excerpt" json:"excerpt,omitempty"`
	Status      string     `db:"status" json:"status"`
	AuthorID    int64      `db:"author_id" json:"author_id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	PublishedAt *time.Time `db:"published_at" json:"published_at,omitempty"`
}

type PostWithAuthor struct {
	Post
	AuthorUsername    string  `db:"author_username" json:"author_username"`
	AuthorDisplayName string  `db:"author_display_name" json:"author_display_name"`
	AuthorAvatarURL   *string `db:"author_avatar_url" json:"author_avatar_url,omitempty"`
	LikeCount         int64   `db:"like_count" json:"like_count"`
	LikedByUser       bool    `db:"liked_by_user" json:"liked_by_user"`
	BookmarkedByUser  bool    `db:"bookmarked_by_user" json:"bookmarked_by_user"`
}

type ToggleLikeResponse struct {
	Liked     bool  `json:"liked"`
	LikeCount int64 `json:"like_count"`
}

type ToggleBookmarkResponse struct {
	Bookmarked bool `json:"bookmarked"`
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

type UpdatePostRequest struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
	Status  *string `json:"status,omitempty"`
}

type PostListParams struct {
	Page    int
	PerPage int
	Search  string
	Status  string
	UserID  int64 // optional: for liked_by_user, 0 = anonymous
}

type PostListResponse struct {
	Data []PostWithAuthor `json:"data"`
	Meta PaginationMeta   `json:"meta"`
}

type PaginationMeta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
}
