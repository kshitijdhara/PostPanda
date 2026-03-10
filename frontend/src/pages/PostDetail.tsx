import { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeft, Calendar, Edit03, Trash01, User01, Heart, ThumbsUp, ThumbsDown, Bookmark, BookmarkCheck } from '@untitled-ui/icons-react';
import { usePost, useDeletePost, useToggleLike, useToggleBookmark } from '../hooks/usePosts';
import { useComments, useCreateComment, useDeleteComment, useVoteComment } from '../hooks/useComments';
import { useAuth } from '../features/auth/AuthContext';
import MarkdownRenderer from '../components/MarkdownRenderer/MarkdownRenderer';
import type { Comment } from '../types/comment';
import styles from './PostDetail.module.scss';

export default function PostDetail() {
  const { slug } = useParams<{ slug: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { data: post, isLoading, error } = usePost(slug!);
  const deletePost = useDeletePost();
  const toggleLike = useToggleLike(slug!);
  const toggleBookmark = useToggleBookmark(slug!);
  const { data: comments } = useComments(slug!);
  const createComment = useCreateComment(slug!);
  const deleteComment = useDeleteComment(slug!);
  const voteComment = useVoteComment(slug!);
  const [commentText, setCommentText] = useState('');
  const [replyTo, setReplyTo] = useState<number | null>(null);

  if (isLoading) return <div className={styles.loading}>Loading...</div>;
  if (error || !post) return (
    <div className={styles.container}>
      <div className={styles.notFound}>
        <h2>Post not found</h2>
        <p>The post you're looking for doesn't exist or has been removed.</p>
        <Link to="/" className={styles.backLink}>Back to home</Link>
      </div>
    </div>
  );

  const isAuthor = user?.id === post.author_id;
  const date = new Date(post.published_at || post.created_at).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  });

  const handleDelete = async () => {
    if (window.confirm('Are you sure you want to delete this post?')) {
      await deletePost.mutateAsync(slug!);
      navigate('/');
    }
  };

  const handleComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commentText.trim()) return;
    await createComment.mutateAsync({
      content: commentText,
      parent_id: replyTo || undefined,
    });
    setCommentText('');
    setReplyTo(null);
  };

  const handleVote = (commentId: number, currentVote: number | undefined, newValue: number) => {
    // Clicking the same vote removes it (toggle off)
    const value = currentVote === newValue ? 0 : newValue;
    voteComment.mutate({ id: commentId, value });
  };

  // Build threaded comments
  const topLevel = comments?.filter(c => !c.parent_id) || [];
  const replies = (parentId: number) => comments?.filter(c => c.parent_id === parentId) || [];

  const renderComment = (comment: Comment, isReply = false) => (
    <div key={comment.id} className={`${styles.comment} ${isReply ? styles.reply : ''}`}>
      <div className={styles.commentMeta}>
        <span className={styles.commentAuthor}>{comment.author_display_name}</span>
        <span className={styles.commentDate}>
          {new Date(comment.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
        </span>
        {user && user.id === comment.author_id && (
          <button
            className={styles.commentDelete}
            style={{ marginLeft: 'auto' }}
            onClick={() => deleteComment.mutate(comment.id)}
          >
            Delete
          </button>
        )}
      </div>

      <p className={styles.commentContent}>{comment.content}</p>

      <div className={styles.commentActions}>
        {/* Upvote / downvote */}
        <div className={styles.voteGroup}>
          <button
            className={`${styles.voteBtn} ${comment.user_vote === 1 ? styles.votedUp : ''}`}
            onClick={() => user && handleVote(comment.id, comment.user_vote, 1)}
            title={user ? 'Upvote' : 'Login to vote'}
            disabled={!user}
          >
            <ThumbsUp width={14} height={14} />
            {comment.upvotes > 0 && <span>{comment.upvotes}</span>}
          </button>
          <button
            className={`${styles.voteBtn} ${comment.user_vote === -1 ? styles.votedDown : ''}`}
            onClick={() => user && handleVote(comment.id, comment.user_vote, -1)}
            title={user ? 'Downvote' : 'Login to vote'}
            disabled={!user}
          >
            <ThumbsDown width={14} height={14} />
            {comment.downvotes > 0 && <span>{comment.downvotes}</span>}
          </button>
        </div>

        {user && !isReply && (
          <button
            className={styles.replyBtn}
            onClick={() => setReplyTo(replyTo === comment.id ? null : comment.id)}
          >
            {replyTo === comment.id ? 'Cancel' : 'Reply'}
          </button>
        )}
      </div>

      {replyTo === comment.id && (
        <form onSubmit={handleComment} className={styles.commentForm} style={{ marginTop: '0.5rem' }}>
          <textarea
            className={styles.commentTextarea}
            placeholder="Write a reply..."
            value={commentText}
            onChange={e => setCommentText(e.target.value)}
            style={{ minHeight: '60px' }}
          />
          <button type="submit" className={styles.commentSubmit} disabled={createComment.isPending}>
            Reply
          </button>
        </form>
      )}
      {replies(comment.id).map(r => renderComment(r, true))}
    </div>
  );

  return (
    <div className={styles.container}>
      <Link to="/" className={styles.backLink}>
        <ArrowLeft width={18} height={18} />
        <span>Back to posts</span>
      </Link>

      <article>
        <header className={styles.header}>
          <h1 className={styles.title}>{post.title}</h1>
          <div className={styles.meta}>
            <User01 width={16} height={16} />
            <Link to={`/profile/${post.author_username}`} className={styles.authorName}>
              {post.author_display_name}
            </Link>
            <span>&middot;</span>
            <Calendar width={16} height={16} />
            <span>{date}</span>
            {post.status === 'draft' && (
              <span className={styles.draftBadge}>Draft</span>
            )}
            {isAuthor && (
              <div className={styles.actions}>
                <Link to={`/posts/${post.slug}/edit`} className={styles.editBtn}>
                  <Edit03 width={16} height={16} />
                  <span>Edit</span>
                </Link>
                <button onClick={handleDelete} className={styles.deleteBtn} disabled={deletePost.isPending}>
                  <Trash01 width={16} height={16} />
                  <span>Delete</span>
                </button>
              </div>
            )}
          </div>
        </header>

        <div className={styles.content}>
          <MarkdownRenderer content={post.content} />
        </div>

        {/* Like + Bookmark bar — published posts only */}
        {post.status === 'published' && (
          <div className={styles.likeBar}>
            <button
              className={`${styles.likeBtn} ${post.liked_by_user ? styles.liked : ''}`}
              onClick={() => user && toggleLike.mutate()}
              disabled={!user || toggleLike.isPending}
              title={user ? (post.liked_by_user ? 'Unlike' : 'Like') : 'Login to like'}
            >
              <Heart width={18} height={18} />
              <span>{post.like_count > 0 ? post.like_count : ''} {post.liked_by_user ? 'Liked' : 'Like'}</span>
            </button>
            {user && (
              <button
                className={`${styles.bookmarkBtn} ${post.bookmarked_by_user ? styles.bookmarked : ''}`}
                onClick={() => toggleBookmark.mutate()}
                disabled={toggleBookmark.isPending}
                title={post.bookmarked_by_user ? 'Remove bookmark' : 'Save bookmark'}
              >
                {post.bookmarked_by_user
                  ? <BookmarkCheck width={18} height={18} />
                  : <Bookmark width={18} height={18} />
                }
                <span>{post.bookmarked_by_user ? 'Saved' : 'Save'}</span>
              </button>
            )}
            {!user && <span className={styles.likeHint}><Link to="/login">Login</Link> to like or save this post</span>}
          </div>
        )}
      </article>

      {/* Comments — published posts only */}
      {post.status === 'published' && (
        <section className={styles.commentSection}>
          <h2 className={styles.commentHeader}>Comments ({comments?.length || 0})</h2>

          {user && !replyTo && (
            <form onSubmit={handleComment} className={styles.commentForm}>
              <textarea
                className={styles.commentTextarea}
                placeholder="Share your thoughts..."
                value={commentText}
                onChange={e => setCommentText(e.target.value)}
              />
              <button type="submit" className={styles.commentSubmit} disabled={createComment.isPending}>
                {createComment.isPending ? 'Posting...' : 'Post Comment'}
              </button>
            </form>
          )}

          {!user && (
            <p style={{ color: '#6b7280', marginBottom: '1.5rem' }}>
              <Link to="/login" style={{ color: '#2563eb' }}>Login</Link> to leave a comment.
            </p>
          )}

          <div className={styles.commentList}>
            {topLevel.map(c => renderComment(c))}
          </div>
        </section>
      )}
    </div>
  );
}
