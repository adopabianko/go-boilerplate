package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-boilerplate/internal/delivery/http/handler"
	"go-boilerplate/internal/dto"
	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserUsecase
type MockUserUsecase struct {
	mock.Mock
}

func (m *MockUserUsecase) Register(ctx context.Context, email, password string) error {
	args := m.Called(ctx, email, password)
	return args.Error(0)
}

func (m *MockUserUsecase) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockUserUsecase) ListUsers(ctx context.Context, page, limit int, order string) ([]entity.User, int64, error) {
	args := m.Called(ctx, page, limit, order)
	return args.Get(0).([]entity.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserUsecase) GetUser(ctx context.Context, id uint) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserUsecase) UpdateUser(ctx context.Context, id uint, email string) error {
	args := m.Called(ctx, id, email)
	return args.Error(0)
}

func (m *MockUserUsecase) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestUserHandler_Register(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	h := handler.NewUserHandler(mockUsecase)

	r := setupRouter()
	r.POST("/register", h.Register)

	reqBody := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	mockUsecase.On("Register", mock.Anything, reqBody.Email, reqBody.Password).Return(nil)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var res response.Response
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.True(t, res.Success)
}

func TestUserHandler_Login(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	h := handler.NewUserHandler(mockUsecase)

	r := setupRouter()
	r.POST("/login", h.Login)

	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	mockUsecase.On("Login", mock.Anything, reqBody.Email, reqBody.Password).Return("token-secret", nil)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.True(t, res.Success)
	data := res.Data.(map[string]interface{})
	assert.Equal(t, "token-secret", data["token"])
}

func TestUserHandler_ListUsers(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	h := handler.NewUserHandler(mockUsecase)

	r := setupRouter()
	r.GET("/users", h.ListUsers)

	users := []entity.User{
		{ID: 1, Email: "u1@example.com"},
	}
	mockUsecase.On("ListUsers", mock.Anything, 1, 10, "created_at desc").Return(users, int64(1), nil)

	req, _ := http.NewRequest("GET", "/users?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.True(t, res.Success)
	// We could verify data structure further if needed
}
