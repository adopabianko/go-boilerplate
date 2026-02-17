package response

import (
	"net/http"

	"go-boilerplate/pkg/errors"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
	Meta    *Meta        `json:"meta,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

type Meta struct {
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Total  int64  `json:"total"`
	Order  string `json:"order"`
}

type ErrorDetail struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func Success(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessWithPagination(c *gin.Context, code int, message string, data interface{}, meta Meta) {
	c.JSON(code, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

func Error(c *gin.Context, err error) {
	if customErr, ok := err.(*errors.CustomError); ok {
		c.JSON(customErr.Code, Response{
			Success: false,
			Message: customErr.Message,
			Error: &ErrorDetail{
				Code:    customErr.Code,
				Message: customErr.Error(),
			},
		})
		return
	}

	// Default to 500 Internal Server Error
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Message: "Internal Server Error",
		Error: &ErrorDetail{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		},
	})
}
