package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-boilerplate/internal/delivery/http/handler"
	"go-boilerplate/internal/dto"
	"go-boilerplate/internal/repository"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUserIntegration(t *testing.T) {
	truncateTables()

	// Initialize layers
	userRepo := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, cfg, rdb)
	userHandler := handler.NewUserHandler(userUsecase)

	// Setup Router
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/register", userHandler.Register)
	r.POST("/login", userHandler.Login)
	r.GET("/users/:id", userHandler.GetUser)

	var accessToken string
	var userID string

	t.Run("Register Success", func(t *testing.T) {
		reqBody := dto.RegisterRequest{
			Email:    "integration@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code, "Response Body: %s", w.Body.String())
		
		var res response.Response
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		require.True(t, res.Success, "Register failed: %s, Body: %s", res.Message, w.Body.String())
	})

	t.Run("Register Duplicate Email Failure", func(t *testing.T) {
		reqBody := dto.RegisterRequest{
			Email:    "integration@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code, "Response Body: %s", w.Body.String())
		
		var res response.Response
		json.Unmarshal(w.Body.Bytes(), &res)
		require.False(t, res.Success, "Expected success=false, Body: %s", w.Body.String())
		require.Contains(t, res.Message, "Email already exists")
	})

	t.Run("Login Success", func(t *testing.T) {
		// Ensure user exists (sanity check)
		user, err := userRepo.GetByEmail(context.Background(), "integration@example.com")
		require.NoError(t, err, "User should exist before login")
		require.NotNil(t, user)

		reqBody := dto.LoginRequest{
			Email:    "integration@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		
		var res response.Response
		err = json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		require.True(t, res.Success, "Login failed: %s", res.Message)
		
		require.NotNil(t, res.Data, "Response data is nil")
		data := res.Data.(map[string]interface{})
		accessToken = data["access_token"].(string)
		require.NotEmpty(t, accessToken)
	})

	t.Run("Get User Success", func(t *testing.T) {
		// First get the user by email from DB to get the ID
		user, err := userRepo.GetByEmail(context.Background(), "integration@example.com")
		require.NoError(t, err)
		require.NotNil(t, user)
		userID = user.ID

		req, _ := http.NewRequest("GET", "/users/"+userID, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		
		var res response.Response
		err = json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		require.True(t, res.Success)
		
		require.NotNil(t, res.Data, "Response data is nil")
		data := res.Data.(map[string]interface{})
		require.Equal(t, "integration@example.com", data["email"])
	})
}
