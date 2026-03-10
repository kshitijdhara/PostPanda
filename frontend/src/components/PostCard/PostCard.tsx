import { Link } from 'react-router-dom';
import { Calendar, User01, Bookmark, BookmarkCheck } from '@untitled-ui/icons-react';
import type { Post } from '../../types/post';
import { useAuth } from '../../features/auth/AuthContext';
import { useToggleBookmark } from '../../hooks/usePosts';
import styles from './PostCard.module.scss';

interface PostCardProps {
  post: Post;
  showDraftBadge?: boolean;
  showBookmark?: boolean;
}

export default function PostCard({ post, showDraftBadge, showBookmark }: PostCardProps) {
  const { user } = useAuth();
  const toggleBookmark = useToggleBookmark(post.slug);

  const date = new Date(post.published_at || post.created_at).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });

  return (
    <div className={styles.card}>
      <Link to={`/posts/${post.slug}`} className={styles.cardLink}>
        <div className={styles.header}>
          <div className={styles.meta}>
            <User01 width={16} height={16} />
            <span className={styles.author}>{post.author_display_name}</span>
          </div>
          <span className={styles.dot}>&middot;</span>
          <div className={styles.meta}>
            <Calendar width={16} height={16} />
            <span className={styles.date}>{date}</span>
          </div>
          {showDraftBadge && post.status === 'draft' && (
            <span className={styles.draftBadge}>Draft</span>
          )}
        </div>
        <h2 className={styles.title}>{post.title}</h2>
        {post.excerpt && <p className={styles.excerpt}>{post.excerpt}</p>}
      </Link>
      {showBookmark && user && post.status === 'published' && (
        <div className={styles.cardFooter}>
          <button
            className={`${styles.bookmarkBtn} ${post.bookmarked_by_user ? styles.bookmarked : ''}`}
            onClick={() => toggleBookmark.mutate()}
            disabled={toggleBookmark.isPending}
            title={post.bookmarked_by_user ? 'Remove bookmark' : 'Save bookmark'}
            aria-label={post.bookmarked_by_user ? 'Remove bookmark' : 'Save bookmark'}
          >
            {post.bookmarked_by_user
              ? <BookmarkCheck width={16} height={16} />
              : <Bookmark width={16} height={16} />
            }
          </button>
        </div>
      )}
    </div>
  );
}
