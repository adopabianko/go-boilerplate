package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-boilerplate/internal/config"
	"go-boilerplate/internal/dto"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/infrastructure/redis"
	"go-boilerplate/internal/repository"
	"go-boilerplate/pkg/auth"
	appErrors "go-boilerplate/pkg/errors"
	"go-boilerplate/pkg/response"
	"go-boilerplate/pkg/tracer"

	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	ListUsers(ctx context.Context, req dto.ListUsersRequest, timezone string) ([]dto.UserResponse, response.Meta, error)
	GetUser(ctx context.Context, id string, timezone string) (*entity.User, error)
	UpdateUser(ctx context.Context, id string, email string) error
	DeleteUser(ctx context.Context, id string) error
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
	ctx, span := tracer.StartSpan(ctx, "UserUsecase.Register", "usecase")
	defer span.End()

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
	ctx, span := tracer.StartSpan(ctx, "UserUsecase.Login", "usecase")
	defer span.End()

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
	ctx, span := tracer.StartSpan(ctx, "UserUsecase.RefreshToken", "usecase")
	defer span.End()

	claims, err := auth.ValidateRefreshToken(tokenString, u.config.JWT)
	if err != nil {
		return "", "", appErrors.New(401, "Invalid refresh token")
	}

	// Optional: Check if user still exists/is active
	user, err := u.repo.GetByID(ctx, claims.UserID, "")
	if err != nil {
		return "", "", appErrors.New(401, "User not found")
	}

	accessToken, refreshToken, err := auth.GenerateTokenPair(user.ID, u.config.JWT)
	if err != nil {
		return "", "", appErrors.Wrap(err, 500, "Failed to generate tokens")
	}

	return accessToken, refreshToken, nil
}

func (u *userUsecase) ListUsers(ctx context.Context, req dto.ListUsersRequest, timezone string) ([]dto.UserResponse, response.Meta, error) {
	ctx, span := tracer.StartSpan(ctx, "UserUsecase.ListUsers", "usecase")
	defer span.End()

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Order == "" {
		req.Order = "created_at desc"
	}

	users, total, err := u.repo.List(ctx, req, timezone)
	if err != nil {
		return nil, response.Meta{}, appErrors.Wrap(err, 500, "Failed to list users")
	}

	userResponses := make([]dto.UserResponse, len(users))
	for i, u := range users {
		userResponses[i] = dto.UserResponse{
			ID:        u.ID,
			Email:     u.Email,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}

	offset := (req.Page - 1) * req.Limit
	meta := response.Meta{
		Offset: offset,
		Limit:  req.Limit,
		Total:  total,
		Order:  req.Order,
	}

	return userResponses, meta, nil
}

func (u *userUsecase) GetUser(ctx context.Context, id string, timezone string) (*entity.User, error) {
	ctx, span := tracer.StartSpan(ctx, "UserUsecase.GetUser", "usecase")
	defer span.End()

	if timezone == "" {
		timezone = "UTC"
	}

	// Check Redis Cache (timezone-specific)
	cacheKey := fmt.Sprintf("user:%s:%s", id, timezone)
	cachedUser, err := u.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var user entity.User
		if err := json.Unmarshal([]byte(cachedUser), &user); err == nil {
			return &user, nil
		}
	}

	// Determine from DB
	user, err := u.repo.GetByID(ctx, id, timezone)
	if err != nil {
		return nil, appErrors.Wrap(err, 404, "User not found")
	}

	// Set Cache
	userJSON, _ := json.Marshal(user)
	u.redis.Set(ctx, cacheKey, string(userJSON), 1*time.Hour)

	return user, nil
}

func (u *userUsecase) UpdateUser(ctx context.Context, id string, email string) error {
	ctx, span := tracer.StartSpan(ctx, "UserUsecase.UpdateUser", "usecase")
	defer span.End()

	user, err := u.repo.GetByID(ctx, id, "UTC") // Get original for update
	if err != nil {
		return appErrors.Wrap(err, 404, "User not found")
	}

	user.Email = email
	if err := u.repo.Update(ctx, user); err != nil {
		return appErrors.Wrap(err, 500, "Failed to update user")
	}

	// Invalidate Cache (all timezones for this user)
	u.redis.Del(ctx, fmt.Sprintf("user:%s:*", id))

	return nil
}

func (u *userUsecase) DeleteUser(ctx context.Context, id string) error {
	ctx, span := tracer.StartSpan(ctx, "UserUsecase.DeleteUser", "usecase")
	defer span.End()

	if err := u.repo.Delete(ctx, id); err != nil {
		return appErrors.Wrap(err, 500, "Failed to delete user")
	}

	// Invalidate Cache (all timezones for this user)
	u.redis.Del(ctx, fmt.Sprintf("user:%s:*", id))

	return nil
}
