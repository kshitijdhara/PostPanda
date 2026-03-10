import type { PostListResponse, CreatePostRequest, UpdatePostRequest, Post } from '../types/post';
import type { User, RegisterRequest, LoginRequest, UpdateProfileRequest, ChangePasswordRequest } from '../types/user';
import type { Comment, CreateCommentRequest } from '../types/comment';

const BASE_URL = '/api/v1';

class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(
      res.status,
      body?.error?.code || 'UNKNOWN',
      body?.error?.message || `Request failed with status ${res.status}`
    );
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return res.json();
}

// Auth
export const auth = {
  register: (data: RegisterRequest) =>
    request<{ data: User }>('/auth/register', { method: 'POST', body: JSON.stringify(data) }).then(r => r.data),

  login: (data: LoginRequest) =>
    request<{ data: User }>('/auth/login', { method: 'POST', body: JSON.stringify(data) }).then(r => r.data),

  logout: () =>
    request<void>('/auth/logout', { method: 'POST' }),

  me: () =>
    request<{ data: User }>('/auth/me').then(r => r.data),
};

// Posts
export const posts = {
  list: (page = 1, perPage = 20, search?: string) => {
    const params = new URLSearchParams({ page: String(page), per_page: String(perPage) });
    if (search) params.set('search', search);
    return request<PostListResponse>(`/posts?${params}`);
  },

  get: (slug: string) =>
    request<{ data: Post }>(`/posts/${slug}`).then(r => r.data),

  create: (data: CreatePostRequest) =>
    request<{ data: Post }>('/posts', { method: 'POST', body: JSON.stringify(data) }).then(r => r.data),

  update: (slug: string, data: UpdatePostRequest) =>
    request<{ data: Post }>(`/posts/${slug}`, { method: 'PUT', body: JSON.stringify(data) }).then(r => r.data),

  delete: (slug: string) =>
    request<void>(`/posts/${slug}`, { method: 'DELETE' }),

  drafts: () =>
    request<{ data: Post[] }>('/posts/drafts/mine').then(r => r.data),
};

// Comments
export const comments = {
  list: (slug: string) =>
    request<{ data: Comment[] }>(`/posts/${slug}/comments`).then(r => r.data),

  create: (slug: string, data: CreateCommentRequest) =>
    request<{ data: Comment }>(`/posts/${slug}/comments`, { method: 'POST', body: JSON.stringify(data) }).then(r => r.data),

  delete: (id: number) =>
    request<void>(`/comments/${id}`, { method: 'DELETE' }),

  vote: (id: number, value: number) =>
    request<{ data: { upvotes: number; downvotes: number; user_vote?: number } }>(
      `/comments/${id}/vote`, { method: 'POST', body: JSON.stringify({ value }) }
    ).then(r => r.data),
};

// Post likes
export const likes = {
  toggle: (slug: string) =>
    request<{ data: { liked: boolean; like_count: number } }>(`/posts/${slug}/like`, { method: 'POST' }).then(r => r.data),
};

// Post bookmarks
export const bookmarks = {
  toggle: (slug: string) =>
    request<{ data: { bookmarked: boolean } }>(`/posts/${slug}/bookmark`, { method: 'POST' }).then(r => r.data),
};

// Users
export const users = {
  getByUsername: (username: string) =>
    request<{ data: User }>(`/users/${username}`).then(r => r.data),

  getPostsByUsername: (username: string, page = 1, perPage = 20) => {
    const params = new URLSearchParams({ page: String(page), per_page: String(perPage) });
    return request<PostListResponse>(`/users/${username}/posts?${params}`);
  },

  getCommentsByUsername: (username: string) =>
    request<{ data: Comment[] }>(`/users/${username}/comments`).then(r => r.data),

  getLikedPosts: () =>
    request<{ data: Post[] }>('/users/me/liked-posts').then(r => r.data),

  getBookmarkedPosts: () =>
    request<{ data: Post[] }>('/users/me/bookmarked-posts').then(r => r.data),

  getMyComments: () =>
    request<{ data: Comment[] }>('/users/me/comments').then(r => r.data),

  updateProfile: (data: UpdateProfileRequest) =>
    request<{ data: User }>('/users/me', { method: 'PUT', body: JSON.stringify(data) }).then(r => r.data),

  changePassword: (data: ChangePasswordRequest) =>
    request<void>('/users/me/password', { method: 'PUT', body: JSON.stringify(data) }),
};

export { ApiError };
