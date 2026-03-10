import PostForm from '../components/PostForm/PostForm';
import { useCreatePost } from '../hooks/usePosts';

export default function PostCreate() {
  const createPost = useCreatePost();

  return (
    <PostForm
      onSubmit={async (data) => {
        const post = await createPost.mutateAsync(data);
        return { slug: post.slug };
      }}
      isSubmitting={createPost.isPending}
    />
  );
}
