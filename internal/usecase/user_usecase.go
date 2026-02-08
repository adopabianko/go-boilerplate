package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-boilerplate/internal/config"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/infrastructure/redis"
	"go-boilerplate/internal/repository"
	"go-boilerplate/pkg/auth"
	appErrors "go-boilerplate/pkg/errors"

	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	ListUsers(ctx context.Context, page, limit int, order string) ([]entity.User, int64, error)
	GetUser(ctx context.Context, id uint) (*entity.User, error)
	UpdateUser(ctx context.Context, id uint, email string) error
	DeleteUser(ctx context.Context, id uint) error
}

type userUsecase struct {
	repo   repository.UserRepository
	config *config.Config
	redis  *redis.Client
}

func NewUserUsecase(repo repository.UserRepository, cfg *config.Config, rdb *redis.Client) UserUsecase {
	return &userUsecase{repo: repo, config: cfg, redis: rdb}
}

func (u *userUsecase) Register(ctx context.Context, email, password string) error {
	existingUser, _ := u.repo.GetByEmail(ctx, email)
	if existingUser != nil {
		return appErrors.New(400, "Email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return appErrors.Wrap(err, 500, "Failed to hash password")
	}

	user := &entity.User{
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := u.repo.Create(ctx, user); err != nil {
		return appErrors.Wrap(err, 500, "Failed to create user")
	}

	return nil
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", appErrors.New(401, "Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", appErrors.New(401, "Invalid credentials")
	}

	accessToken, refreshToken, err := auth.GenerateTokenPair(user.ID, u.config.JWT)
	if err != nil {
		return "", "", appErrors.Wrap(err, 500, "Failed to generate tokens")
	}

	return accessToken, refreshToken, nil
}

func (u *userUsecase) RefreshToken(ctx context.Context, tokenString string) (string, string, error) {
	claims, err := auth.ValidateRefreshToken(tokenString, u.config.JWT)
	if err != nil {
		return "", "", appErrors.New(401, "Invalid refresh token")
	}

	// Optional: Check if user still exists/is active
	user, err := u.repo.GetByID(ctx, claims.UserID)
	if err != nil {
		return "", "", appErrors.New(401, "User not found")
	}

	accessToken, refreshToken, err := auth.GenerateTokenPair(user.ID, u.config.JWT)
	if err != nil {
		return "", "", appErrors.Wrap(err, 500, "Failed to generate tokens")
	}

	return accessToken, refreshToken, nil
}

func (u *userUsecase) ListUsers(ctx context.Context, page, limit int, order string) ([]entity.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	users, total, err := u.repo.List(ctx, page, limit, order)
	if err != nil {
		return nil, 0, appErrors.Wrap(err, 500, "Failed to list users")
	}

	return users, total, nil
}

func (u *userUsecase) GetUser(ctx context.Context, id uint) (*entity.User, error) {
	// Check Redis Cache
	cacheKey := fmt.Sprintf("user:%d", id)
	cachedUser, err := u.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var user entity.User
		if err := json.Unmarshal([]byte(cachedUser), &user); err == nil {
			return &user, nil
		}
	}

	// Determine from DB
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.Wrap(err, 404, "User not found")
	}

	// Set Cache
	userJSON, _ := json.Marshal(user)
	u.redis.Set(ctx, cacheKey, string(userJSON), 1*time.Hour)

	return user, nil
}

func (u *userUsecase) UpdateUser(ctx context.Context, id uint, email string) error {
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return appErrors.Wrap(err, 404, "User not found")
	}

	user.Email = email
	if err := u.repo.Update(ctx, user); err != nil {
		return appErrors.Wrap(err, 500, "Failed to update user")
	}

	// Invalidate Cache
	cacheKey := fmt.Sprintf("user:%d", id)
	u.redis.Del(ctx, cacheKey)

	return nil
}

func (u *userUsecase) DeleteUser(ctx context.Context, id uint) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return appErrors.Wrap(err, 500, "Failed to delete user")
	}

	// Invalidate Cache
	cacheKey := fmt.Sprintf("user:%d", id)
	u.redis.Del(ctx, cacheKey)

	return nil
}
