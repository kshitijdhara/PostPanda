import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { posts as postsApi, likes as likesApi, bookmarks as bookmarksApi } from '../services/api';
import type { CreatePostRequest, UpdatePostRequest, Post } from '../types/post';

export function usePosts(page = 1, perPage = 20, search?: string) {
  return useQuery({
    queryKey: ['posts', page, perPage, search],
    queryFn: () => postsApi.list(page, perPage, search),
  });
}

export function usePost(slug: string) {
  return useQuery({
    queryKey: ['post', slug],
    queryFn: () => postsApi.get(slug),
    enabled: !!slug,
  });
}

export function useDrafts() {
  return useQuery({
    queryKey: ['drafts'],
    queryFn: () => postsApi.drafts(),
  });
}

export function useCreatePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreatePostRequest) => postsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['posts'] });
      queryClient.invalidateQueries({ queryKey: ['drafts'] });
    },
  });
}

export function useUpdatePost(slug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdatePostRequest) => postsApi.update(slug, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['posts'] });
      queryClient.invalidateQueries({ queryKey: ['post', slug] });
      queryClient.invalidateQueries({ queryKey: ['drafts'] });
    },
  });
}

export function useDeletePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (slug: string) => postsApi.delete(slug),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['posts'] });
      queryClient.invalidateQueries({ queryKey: ['drafts'] });
    },
  });
}

export function useToggleLike(slug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => likesApi.toggle(slug),
    onSuccess: (result) => {
      // Optimistically update the post cache
      queryClient.setQueryData(['post', slug], (old: any) => {
        if (!old) return old;
        return { ...old, like_count: result.like_count, liked_by_user: result.liked };
      });
      queryClient.invalidateQueries({ queryKey: ['posts'] });
    },
  });
}

export function useToggleBookmark(slug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => bookmarksApi.toggle(slug),
    onSuccess: (result) => {
      // Update the individual post cache
      queryClient.setQueryData(['post', slug], (old: any) => {
        if (!old) return old;
        return { ...old, bookmarked_by_user: result.bookmarked };
      });
      // Update all paginated post list caches
      queryClient.setQueriesData({ queryKey: ['posts'] }, (old: any) => {
        if (!old || !old.data) return old;
        return {
          ...old,
          data: old.data.map((p: Post) =>
            p.slug === slug ? { ...p, bookmarked_by_user: result.bookmarked } : p
          ),
        };
      });
      queryClient.invalidateQueries({ queryKey: ['bookmarked-posts'] });
    },
  });
}
