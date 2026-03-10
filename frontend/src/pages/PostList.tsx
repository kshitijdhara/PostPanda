import { useState, useEffect, useRef } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { SearchSm } from '@untitled-ui/icons-react';
import { usePosts } from '../hooks/usePosts';
import PostCard from '../components/PostCard/PostCard';
import { useAuth } from '../features/auth/AuthContext';
import styles from './PostList.module.scss';

export default function PostList() {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get('page')) || 1;

  const [searchInput, setSearchInput] = useState(searchParams.get('search') || '');
  const [debouncedSearch, setDebouncedSearch] = useState(searchInput);
  const isFirstRender = useRef(true);

  const { data, isLoading, isFetching } = usePosts(page, 20, debouncedSearch || undefined);
  const { isAuthenticated } = useAuth();

  // Debounce: update query + URL after 300ms of no typing
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchInput);

      if (isFirstRender.current) {
        isFirstRender.current = false;
        return;
      }

      const params: Record<string, string> = {};
      if (searchInput) params.search = searchInput;
      setSearchParams(params, { replace: true });
    }, 300);
    return () => clearTimeout(timer);
  }, [searchInput, setSearchParams]);

  const totalPages = data ? Math.ceil(data.meta.total / data.meta.per_page) : 0;
  const isSearching = isFetching && debouncedSearch !== '';

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>Latest Posts</h1>
      </div>

      <div className={styles.searchWrapper}>
        <SearchSm className={styles.searchIcon} width={18} height={18} />
        <input
          type="search"
          placeholder="Search posts..."
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          className={styles.searchInput}
        />
        {isSearching && <span className={styles.searchSpinner} />}
      </div>

      {isLoading ? (
        <div className={styles.loading}>Loading posts...</div>
      ) : data && data.data.length > 0 ? (
        <>
          {debouncedSearch && (
            <p className={styles.searchMeta}>
              {data.meta.total} result{data.meta.total !== 1 ? 's' : ''} for &ldquo;{debouncedSearch}&rdquo;
            </p>
          )}
          <div className={styles.grid}>
            {data.data.map((post) => (
              <PostCard key={post.id} post={post} showBookmark />
            ))}
          </div>

          {totalPages > 1 && (
            <div className={styles.pagination}>
              <button
                className={styles.pageBtn}
                disabled={page <= 1}
                onClick={() => setSearchParams({ page: String(page - 1), ...(debouncedSearch && { search: debouncedSearch }) })}
              >
                Previous
              </button>
              <span className={styles.pageInfo}>Page {page} of {totalPages}</span>
              <button
                className={styles.pageBtn}
                disabled={page >= totalPages}
                onClick={() => setSearchParams({ page: String(page + 1), ...(debouncedSearch && { search: debouncedSearch }) })}
              >
                Next
              </button>
            </div>
          )}
        </>
      ) : (
        <div className={styles.empty}>
          {debouncedSearch ? (
            <>
              <h2>No results found</h2>
              <p>No posts match &ldquo;{debouncedSearch}&rdquo;. Try a different search.</p>
            </>
          ) : (
            <>
              <h2>No posts yet</h2>
              <p>Be the first to share something with the community.</p>
              {isAuthenticated && (
                <Link to="/posts/new" className={styles.ctaBtn}>Write your first post</Link>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
}
