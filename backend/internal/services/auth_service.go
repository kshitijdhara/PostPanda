package services

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kshitijdhara/blog/internal/domain"
	"github.com/kshitijdhara/blog/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService struct {
	userRepo  repositories.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo repositories.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.User, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		DisplayName:  req.DisplayName,
	}

	user, err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, "", err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, req domain.LoginRequest) (*domain.User, string, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) ValidateToken(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, ErrInvalidToken
	}

	return int64(userID), nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *AuthService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID int64, req domain.UpdateProfileRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Bio != "" {
		user.Bio = &req.Bio
	}

	return s.userRepo.Update(ctx, user)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int64, req domain.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	_, err = s.userRepo.Update(ctx, user)
	return err
}

func (s *AuthService) generateToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
		"jti":      generateJTI(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateJTI() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
