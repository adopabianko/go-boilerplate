package handler

import (
	"net/http"

	"go-boilerplate/internal/dto"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/errors"
	"go-boilerplate/pkg/request"
	"go-boilerplate/pkg/response"
	"go-boilerplate/pkg/tracer"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	usecase usecase.UserUsecase
}

func NewUserHandler(u usecase.UserUsecase) *UserHandler {
	return &UserHandler{usecase: u}
}

// Register godoc
// @Summary      Register a new user
// @Description  Register a new user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Register Request"
// @Success      201  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	ctx, span := tracer.StartSpan(c.Request.Context(), "UserHandler.Register", "handler")
	defer span.End()

	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, err.Error()))
		return
	}

	if err := h.usecase.Register(ctx, req.Email, req.Password); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", nil)
}

// Login godoc
// @Summary      Login user
// @Description  Login with email and password to get JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login Request"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	ctx, span := tracer.StartSpan(c.Request.Context(), "UserHandler.Login", "handler")
	defer span.End()

	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, err.Error()))
		return
	}

	accessToken, refreshToken, err := h.usecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Login successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// RefreshToken godoc
// @Summary      Refresh Access Token
// @Description  Get new access token using refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshTokenRequest true "Refresh Token Request"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	ctx, span := tracer.StartSpan(c.Request.Context(), "UserHandler.RefreshToken", "handler")
	defer span.End()

	var req dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, err.Error()))
		return
	}

	accessToken, refreshToken, err := h.usecase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Token refreshed successfully", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// ListUsers godoc
// @Summary      List users
// @Description  Get list of users with pagination
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request query dto.ListUsersRequest true "List Users Request"
// @Success      200  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	ctx, span := tracer.StartSpan(c.Request.Context(), "UserHandler.ListUsers", "handler")
	defer span.End()

	var q dto.ListUsersRequest
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, err.Error()))
		return
	}

	tz := request.GetTimeLocation(c)
	userResponses, meta, err := h.usecase.ListUsers(ctx, q, tz)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, http.StatusOK, "User list", userResponses, meta)
}

// GetUser godoc
// @Summary      Get a user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	ctx, span := tracer.StartSpan(c.Request.Context(), "UserHandler.GetUser", "handler")
	defer span.End()

	idStr := c.Param("id")
	if idStr == "" {
		response.Error(c, errors.New(http.StatusBadRequest, "Invalid User ID"))
		return
	}

	tz := request.GetTimeLocation(c)
	user, err := h.usecase.GetUser(ctx, idStr, tz)
	if err != nil {
		response.Error(c, err)
		return
	}

	userResponse := dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	response.Success(c, http.StatusOK, "User retrieved successfully", userResponse)
}

// UpdateUser godoc
// @Summary      Update a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Param        request body dto.RegisterRequest true "Update Request"
// @Success      200  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	ctx, span := tracer.StartSpan(c.Request.Context(), "UserHandler.UpdateUser", "handler")
	defer span.End()

	idStr := c.Param("id")
	if idStr == "" {
		response.Error(c, errors.New(http.StatusBadRequest, "Invalid User ID"))
		return
	}

	var req dto.RegisterRequest // Reusing RegisterRequest for simplicity
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, err.Error()))
		return
	}

	if err := h.usecase.UpdateUser(ctx, idStr, req.Email); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User updated successfully", nil)
}

// DeleteUser godoc
// @Summary      Delete a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	ctx, span := tracer.StartSpan(c.Request.Context(), "UserHandler.DeleteUser", "handler")
	defer span.End()

	idStr := c.Param("id")
	if idStr == "" {
		response.Error(c, errors.New(http.StatusBadRequest, "Invalid User ID"))
		return
	}

	if err := h.usecase.DeleteUser(ctx, idStr); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User deleted successfully", nil)
}
