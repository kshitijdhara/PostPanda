import { Link } from 'react-router-dom';
import { useDrafts } from '../hooks/usePosts';
import PostCard from '../components/PostCard/PostCard';
import styles from './PostList.module.scss';

export default function DraftList() {
  const { data: drafts, isLoading } = useDrafts();

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>My Drafts</h1>
        <Link to="/posts/new" className={styles.ctaBtn}>New Post</Link>
      </div>

      {isLoading ? (
        <div className={styles.loading}>Loading...</div>
      ) : drafts && drafts.length > 0 ? (
        <div className={styles.grid}>
          {drafts.map((post) => (
            <PostCard key={post.id} post={post} showDraftBadge />
          ))}
        </div>
      ) : (
        <div className={styles.empty}>
          <h2>No drafts</h2>
          <p>You don't have any drafts yet.</p>
          <Link to="/posts/new" className={styles.ctaBtn}>Start writing</Link>
        </div>
      )}
    </div>
  );
}
