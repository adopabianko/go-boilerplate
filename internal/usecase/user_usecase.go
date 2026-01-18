package usecase

import (
	"context"

	"go-boilerplate/internal/config"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repository"
	"go-boilerplate/pkg/auth"
	appErrors "go-boilerplate/pkg/errors"

	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
	ListUsers(ctx context.Context, page, limit int, order string) ([]entity.User, int64, error)
}

type userUsecase struct {
	repo   repository.UserRepository
	config *config.Config
}

func NewUserUsecase(repo repository.UserRepository, cfg *config.Config) UserUsecase {
	return &userUsecase{repo: repo, config: cfg}
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

func (u *userUsecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", appErrors.New(401, "Invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", appErrors.New(401, "Invalid credentials")
	}

	token, err := auth.GenerateToken(user.ID, u.config.JWT)
	if err != nil {
		return "", appErrors.Wrap(err, 500, "Failed to generate token")
	}

	return token, nil
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
