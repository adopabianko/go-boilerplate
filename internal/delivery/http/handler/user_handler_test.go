package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func (m *MockUserUsecase) Register(ctx context.Context, email, password string, timezone string) error {
	args := m.Called(ctx, email, password, timezone)
	return args.Error(0)
}

func (m *MockUserUsecase) Login(ctx context.Context, email, password string) (string, string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockUserUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockUserUsecase) ListUsers(ctx context.Context, page, limit int, order string, timezone string) ([]entity.User, int64, error) {
	args := m.Called(ctx, page, limit, order, timezone)
	return args.Get(0).([]entity.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserUsecase) GetUser(ctx context.Context, id string, timezone string) (*entity.User, error) {
	args := m.Called(ctx, id, timezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserUsecase) UpdateUser(ctx context.Context, id string, email string, timezone string) error {
	args := m.Called(ctx, id, email, timezone)
	return args.Error(0)
}

func (m *MockUserUsecase) DeleteUser(ctx context.Context, id string) error {
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

	mockUsecase.On("Register", mock.Anything, reqBody.Email, reqBody.Password, "UTC").Return(nil)

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

	mockUsecase.On("Login", mock.Anything, reqBody.Email, reqBody.Password).Return("access-token", "refresh-token", nil)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.True(t, res.Success)
	data := res.Data.(map[string]interface{})
	assert.Equal(t, "access-token", data["access_token"])
	assert.Equal(t, "refresh-token", data["refresh_token"])
}

func TestUserHandler_ListUsers(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	h := handler.NewUserHandler(mockUsecase)

	r := setupRouter()
	r.GET("/users", h.ListUsers)

	users := []entity.User{
		{ID: "019c514b-a933-74f2-8d08-a496675c66cf", Email: "u1@example.com"},
	}
	mockUsecase.On("ListUsers", mock.Anything, 1, 10, "created_at desc", "UTC").Return(users, int64(1), nil)

	req, _ := http.NewRequest("GET", "/users?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.True(t, res.Success)
}

func TestUserHandler_GetUser_Timezone(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	h := handler.NewUserHandler(mockUsecase)

	r := setupRouter()
	r.GET("/users/:id", h.GetUser)

	id := "019c514b-a933-74f2-8d08-a496675c66cf"
	now := time.Now().UTC()
	// In SQL-based approach, we simulate the DB returning a shifted time.
	localNow := now.In(time.FixedZone("WIB", 7*3600))
	user := &entity.User{
		ID:        id,
		Email:     "test@example.com",
		CreatedAt: localNow,
		UpdatedAt: localNow,
	}

	mockUsecase.On("GetUser", mock.Anything, id, "Asia/Jakarta").Return(user, nil)

	// Test with Asia/Jakarta (UTC+7)
	req, _ := http.NewRequest("GET", "/users/"+id, nil)
	req.Header.Set("X-Timezone", "Asia/Jakarta")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	json.Unmarshal(w.Body.Bytes(), &res)
	assert.True(t, res.Success)

	data := res.Data.(map[string]interface{})
	createdAtStr := data["created_at"].(string)

	// Parsing the returned string to check offset
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	assert.NoError(t, err)

	_, offset := createdAt.Zone()
	assert.Equal(t, 7*3600, offset, "Offset should be 7 hours for Asia/Jakarta")
}

