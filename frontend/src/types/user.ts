export interface User {
  id: number;
  username: string;
  email: string;
  display_name: string;
  bio?: string;
  avatar_url?: string;
  banner_url?: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  display_name: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface UpdateProfileRequest {
  display_name?: string;
  username?: string;
  bio?: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}
