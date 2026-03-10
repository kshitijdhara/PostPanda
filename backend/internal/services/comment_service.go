package services

import (
	"context"
	"errors"

	"github.com/kshitijdhara/blog/internal/domain"
	"github.com/kshitijdhara/blog/internal/repositories"
)

var ErrNotCommentAuthor = errors.New("not the author of this comment")

type CommentService struct {
	commentRepo repositories.CommentRepository
	postRepo    repositories.PostRepository
}

func NewCommentService(commentRepo repositories.CommentRepository, postRepo repositories.PostRepository) *CommentService {
	return &CommentService{commentRepo: commentRepo, postRepo: postRepo}
}

func (s *CommentService) Create(ctx context.Context, slug string, req domain.CreateCommentRequest, authorID int64) (*domain.Comment, error) {
	post, err := s.postRepo.GetBySlug(ctx, slug, 0)
	if err != nil {
		return nil, err
	}

	comment := &domain.Comment{
		Content:  req.Content,
		PostID:   post.ID,
		AuthorID: authorID,
		ParentID: req.ParentID,
	}

	return s.commentRepo.Create(ctx, comment)
}

func (s *CommentService) ListBySlug(ctx context.Context, slug string, userID int64) ([]domain.CommentWithAuthor, error) {
	post, err := s.postRepo.GetBySlug(ctx, slug, 0)
	if err != nil {
		return nil, err
	}
	return s.commentRepo.ListByPostID(ctx, post.ID, userID)
}

func (s *CommentService) Delete(ctx context.Context, commentID int64, userID int64) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}
	if comment.AuthorID != userID {
		return ErrNotCommentAuthor
	}
	return s.commentRepo.Delete(ctx, commentID)
}

func (s *CommentService) VoteComment(ctx context.Context, userID, commentID int64, value int) (int64, int64, *int, error) {
	return s.commentRepo.VoteComment(ctx, userID, commentID, value)
}

func (s *CommentService) ListByAuthor(ctx context.Context, authorID int64) ([]domain.CommentWithAuthor, error) {
	return s.commentRepo.ListByAuthor(ctx, authorID)
}
