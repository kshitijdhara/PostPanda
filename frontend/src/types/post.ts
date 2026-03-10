export interface Post {
  id: number;
  title: string;
  slug: string;
  content: string;
  excerpt?: string;
  status: 'draft' | 'published';
  author_id: number;
  author_username: string;
  author_display_name: string;
  author_avatar_url?: string;
  created_at: string;
  updated_at: string;
  published_at?: string;
  like_count: number;
  liked_by_user: boolean;
  bookmarked_by_user: boolean;
}

export interface CreatePostRequest {
  title: string;
  content: string;
  status: 'draft' | 'published';
}

export interface UpdatePostRequest {
  title?: string;
  content?: string;
  status?: 'draft' | 'published';
}

export interface PostListResponse {
  data: Post[];
  meta: PaginationMeta;
}

export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
}
