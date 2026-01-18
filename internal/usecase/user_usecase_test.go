package usecase_test

import (
	"context"
	"testing"

	"go-boilerplate/internal/config"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, page, limit int, order string) ([]entity.User, int64, error) {
	args := m.Called(ctx, page, limit, order)
	return args.Get(0).([]entity.User), args.Get(1).(int64), args.Error(2)
}

func TestUserUsecase_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{}

	uc := usecase.NewUserUsecase(mockRepo, cfg)

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Email == "test@example.com"
	})).Return(nil)

	err := uc.Register(context.Background(), "test@example.com", "password123")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserUsecase_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{JWT: config.JWTConfig{SecretKey: "secret"}}

	uc := usecase.NewUserUsecase(mockRepo, cfg)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &entity.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	token, err := uc.Login(context.Background(), "test@example.com", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestUserUsecase_ListUsers(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{}

	uc := usecase.NewUserUsecase(mockRepo, cfg)

	users := []entity.User{
		{ID: 1, Email: "u1@example.com"},
	}

	mockRepo.On("List", mock.Anything, 1, 10, "").Return(users, int64(1), nil)

	res, total, err := uc.ListUsers(context.Background(), 1, 10, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, res, 1)
	mockRepo.AssertExpectations(t)
}
