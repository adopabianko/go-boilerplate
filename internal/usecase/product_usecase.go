package usecase

import (
	"context"

	"go-boilerplate/internal/entity"
	httpgateway "go-boilerplate/internal/gateway/http"
)

type ProductUsecase interface {
	ListProducts(ctx context.Context, page, limit int) ([]entity.Product, int64, error)
}

type productUsecase struct {
	gateway httpgateway.ProductGateway
}

func NewProductUsecase(gw httpgateway.ProductGateway) ProductUsecase {
	return &productUsecase{gateway: gw}
}

func (u *productUsecase) ListProducts(ctx context.Context, page, limit int) ([]entity.Product, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	skip := (page - 1) * limit

	return u.gateway.GetProducts(ctx, limit, skip)
}
