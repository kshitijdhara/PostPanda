import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { users } from '../services/api';

import type { UpdateProfileRequest, ChangePasswordRequest } from '../types/user';
import { useDrafts } from './usePosts';

export function useUserProfile(username: string) {
  return useQuery({
    queryKey: ['user', username],
    queryFn: () => users.getByUsername(username),
    enabled: !!username,
  });
}

export function useUserPosts(username: string, page = 1) {
  return useQuery({
    queryKey: ['user-posts', username, page],
    queryFn: () => users.getPostsByUsername(username, page),
    enabled: !!username,
  });
}

export function useUpdateProfile() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateProfileRequest) => users.updateProfile(data),
    onSuccess: (updatedUser) => {
      queryClient.invalidateQueries({ queryKey: ['user', updatedUser.username] });
      queryClient.invalidateQueries({ queryKey: ['me'] });
    },
  });
}

export function useChangePassword() {
  return useMutation({
    mutationFn: (data: ChangePasswordRequest) => users.changePassword(data),
  });
}

export function useUserComments(username: string) {
  return useQuery({
    queryKey: ['user-comments', username],
    queryFn: () => users.getCommentsByUsername(username),
    enabled: !!username,
  });
}

export function useMyLikedPosts(enabled: boolean) {
  return useQuery({
    queryKey: ['liked-posts'],
    queryFn: () => users.getLikedPosts(),
    enabled,
  });
}

export function useMyComments(enabled: boolean) {
  return useQuery({
    queryKey: ['my-comments'],
    queryFn: () => users.getMyComments(),
    enabled,
  });
}

export function useMyBookmarkedPosts(enabled: boolean) {
  return useQuery({
    queryKey: ['bookmarked-posts'],
    queryFn: () => users.getBookmarkedPosts(),
    enabled,
  });
}

export { useDrafts };
