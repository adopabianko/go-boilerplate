package handler

import (
	"net/http"

	"go-boilerplate/internal/dto"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	usecase usecase.ProductUsecase
}

func NewProductHandler(u usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{usecase: u}
}

// ListProducts godoc
// @Summary      List products from external API
// @Description  Get list of products from DummyJSON
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        page  query     int     false  "Page number" default(1)
// @Param        limit query     int     false  "Items per page" default(10)
// @Success      200  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page := 1
	limit := 10

	// Reuse PaginationQuery dto if compatible, or just parse manually
	var q dto.PaginationQuery
	if err := c.ShouldBindQuery(&q); err == nil {
		if q.Page > 0 {
			page = q.Page
		}
		if q.Limit > 0 {
			limit = q.Limit
		}
	}

	products, total, err := h.usecase.ListProducts(c.Request.Context(), page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	totalPage := int(total) / limit
	if int(total)%limit != 0 {
		totalPage++
	}

	meta := response.Meta{
		Offset: (page - 1) * limit,
		Limit:  limit,
		Total:  total,
		// Order not supported by DummyJSON list, but we can set default or leave empty
	}

	response.SuccessWithPagination(c, http.StatusOK, "Product list", products, meta)
}
