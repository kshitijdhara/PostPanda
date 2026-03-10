package domain

import "time"

type User struct {
	ID           int64     `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	DisplayName  string    `db:"display_name" json:"display_name"`
	Bio          *string   `db:"bio" json:"bio,omitempty"`
	AvatarURL    *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	BannerURL    *string   `db:"banner_url" json:"banner_url,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Bio         string `json:"bio"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UserResponse struct {
	ID          int64   `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	DisplayName string  `json:"display_name"`
	Bio         *string `json:"bio,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	BannerURL   *string `json:"banner_url,omitempty"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Bio:         u.Bio,
		AvatarURL:   u.AvatarURL,
		BannerURL:   u.BannerURL,
	}
}
