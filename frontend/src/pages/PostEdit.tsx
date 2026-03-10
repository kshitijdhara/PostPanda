import { useParams } from 'react-router-dom';
import PostForm from '../components/PostForm/PostForm';
import { usePost, useUpdatePost } from '../hooks/usePosts';

export default function PostEdit() {
  const { slug } = useParams<{ slug: string }>();
  const { data: post, isLoading } = usePost(slug!);
  const updatePost = useUpdatePost(slug!);

  if (isLoading) return <div style={{ textAlign: 'center', padding: '4rem' }}>Loading...</div>;
  if (!post) return <div style={{ textAlign: 'center', padding: '4rem' }}>Post not found</div>;

  return (
    <PostForm
      initialTitle={post.title}
      initialContent={post.content}
      onSubmit={async (data) => {
        const updated = await updatePost.mutateAsync(data);
        return { slug: updated.slug };
      }}
      isSubmitting={updatePost.isPending}
    />
  );
}
