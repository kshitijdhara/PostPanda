import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import MDEditor from '@uiw/react-md-editor';
import styles from './PostForm.module.scss';

interface PostFormProps {
  initialTitle?: string;
  initialContent?: string;
  onSubmit: (data: { title: string; content: string; status: 'draft' | 'published' }) => Promise<{ slug: string }>;
  isSubmitting?: boolean;
}

export default function PostForm({
  initialTitle = '',
  initialContent = '',
  onSubmit,
  isSubmitting = false,
}: PostFormProps) {
  const navigate = useNavigate();
  const [title, setTitle] = useState(initialTitle);
  const [content, setContent] = useState(initialContent);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent, status: 'draft' | 'published') => {
    e.preventDefault();
    setError('');

    if (!title.trim()) {
      setError('Title is required');
      return;
    }
    if (!content.trim()) {
      setError('Content is required');
      return;
    }

    try {
      const result = await onSubmit({ title: title.trim(), content, status });
      navigate(`/posts/${result.slug}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong');
    }
  };

  return (
    <div className={styles.container}>
      <form className={styles.form}>
        {error && <div className={styles.error}>{error}</div>}

        <input
          type="text"
          placeholder="Post title..."
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className={styles.titleInput}
        />

        <div className={styles.editorWrapper} data-color-mode="light">
          <MDEditor
            value={content}
            onChange={(val) => setContent(val || '')}
            height={400}
            preview="live"
          />
        </div>

        <div className={styles.footer}>
          <button type="button" className={styles.cancelBtn} onClick={() => navigate(-1)}>
            Cancel
          </button>

          <div className={styles.actions}>
            <button
              type="button"
              className={styles.draftBtn}
              onClick={(e) => handleSubmit(e, 'draft')}
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Saving...' : 'Save as Draft'}
            </button>
            <button
              type="button"
              className={styles.publishBtn}
              onClick={(e) => handleSubmit(e, 'published')}
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Publishing...' : 'Publish'}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}
