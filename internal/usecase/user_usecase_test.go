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

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUserUsecase_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{}

	uc := usecase.NewUserUsecase(mockRepo, cfg, nil)

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
	// Assuming tests are run from internal/usecase, we need to go up to root
	// Assuming tests are run from internal/usecase, we need to go up to root
	cfg := &config.Config{JWT: config.JWTConfig{
		PrivateKeyPath:   "../../certs/private.pem",
		PublicKeyPath:    "../../certs/public.pem",
		AccessExpiresIn:  15,
		RefreshExpiresIn: 10080,
	}}

	uc := usecase.NewUserUsecase(mockRepo, cfg, nil)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &entity.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	accessToken, refreshToken, err := uc.Login(context.Background(), "test@example.com", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	mockRepo.AssertExpectations(t)
}

func TestUserUsecase_ListUsers(t *testing.T) {
	mockRepo := new(MockUserRepository)
	cfg := &config.Config{}

	uc := usecase.NewUserUsecase(mockRepo, cfg, nil)

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
