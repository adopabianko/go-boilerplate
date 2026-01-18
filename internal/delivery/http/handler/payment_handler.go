package handler

import (
	"net/http"

	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/errors"
	"go-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	usecase usecase.PaymentUsecase
}

func NewPaymentHandler(u usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{usecase: u}
}

// CheckStatus godoc
// @Summary      Check payment status
// @Description  Check status via gRPC -> External Service
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Transaction ID"
// @Success      200  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/payments/{id} [get]
func (h *PaymentHandler) CheckStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.New(http.StatusBadRequest, "Transaction ID is required"))
		return
	}

	status, err := h.usecase.CheckStatus(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Payment status", status)
}
