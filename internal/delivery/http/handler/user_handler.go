package handler

import (
	"net/http"
	"strconv"

	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/errors"
	"go-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"

	"go-boilerplate/internal/dto"
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
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, err.Error()))
		return
	}

	if err := h.usecase.Register(c.Request.Context(), req.Email, req.Password); err != nil {
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
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, err.Error()))
		return
	}

	accessToken, refreshToken, err := h.usecase.Login(c.Request.Context(), req.Email, req.Password)
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
	var req dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, err.Error()))
		return
	}

	accessToken, refreshToken, err := h.usecase.RefreshToken(c.Request.Context(), req.RefreshToken)
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
// @Param        page  query     int     false  "Page number" default(1)
// @Param        limit query     int     false  "Items per page" default(10)
// @Param        order query     string  false  "Sort order (e.g. created_at desc)"
// @Success      200  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Security     BearerAuth
// @Router       /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Default values
	page := 1
	limit := 10
	order := "created_at desc"

	var q dto.PaginationQuery
	if err := c.ShouldBindQuery(&q); err == nil {
		if q.Page > 0 {
			page = q.Page
		}
		if q.Limit > 0 {
			limit = q.Limit
		}
		if q.Order != "" {
			order = q.Order
		}
	}

	users, total, err := h.usecase.ListUsers(c.Request.Context(), page, limit, order)
	if err != nil {
		response.Error(c, err)
		return
	}

	offset := (page - 1) * limit

	meta := response.Meta{
		Offset: offset,
		Limit:  limit,
		Total:  total,
		Order:  order,
	}

	response.SuccessWithPagination(c, http.StatusOK, "User list", users, meta)
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
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, "Invalid User ID"))
		return
	}

	user, err := h.usecase.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User retrieved successfully", user)
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
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, "Invalid User ID"))
		return
	}

	var req dto.RegisterRequest // Reusing RegisterRequest for simplicity
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, err.Error()))
		return
	}

	if err := h.usecase.UpdateUser(c.Request.Context(), uint(id), req.Email); err != nil {
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
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, errors.New(http.StatusBadRequest, "Invalid User ID"))
		return
	}

	if err := h.usecase.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, "User deleted successfully", nil)
}
