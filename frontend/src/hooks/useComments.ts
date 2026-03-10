import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { comments as commentsApi } from '../services/api';
import type { CreateCommentRequest } from '../types/comment';

export function useComments(slug: string) {
  return useQuery({
    queryKey: ['comments', slug],
    queryFn: () => commentsApi.list(slug),
    enabled: !!slug,
  });
}

export function useCreateComment(slug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateCommentRequest) => commentsApi.create(slug, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments', slug] });
    },
  });
}

export function useDeleteComment(slug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => commentsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments', slug] });
    },
  });
}

export function useVoteComment(slug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, value }: { id: number; value: number }) => commentsApi.vote(id, value),
    onSuccess: (result, { id }) => {
      queryClient.setQueryData(['comments', slug], (old: any[]) => {
        if (!old) return old;
        return old.map(c =>
          c.id === id
            ? { ...c, upvotes: result.upvotes, downvotes: result.downvotes, user_vote: result.user_vote }
            : c
        );
      });
    },
  });
}
