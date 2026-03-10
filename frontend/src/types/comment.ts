export interface Comment {
  id: number;
  content: string;
  post_id: number;
  author_id: number;
  parent_id?: number;
  author_username: string;
  author_display_name: string;
  post_slug?: string;
  post_title?: string;
  created_at: string;
  updated_at: string;
  upvotes: number;
  downvotes: number;
  user_vote?: number; // 1, -1, or undefined
}

export interface CreateCommentRequest {
  content: string;
  parent_id?: number;
}
