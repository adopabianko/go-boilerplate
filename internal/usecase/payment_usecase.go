package usecase

import (
	"context"

	"go-boilerplate/internal/entity"
	grpcgateway "go-boilerplate/internal/gateway/grpc"
)

type PaymentUsecase interface {
	CheckStatus(ctx context.Context, transactionID string) (*entity.PaymentStatus, error)
}

type paymentUsecase struct {
	gateway grpcgateway.PaymentGateway
}

// Ensure gateway.PaymentGateway is defined in internal/gateway/grpc/payment_gateway.go

func NewPaymentUsecase(gw grpcgateway.PaymentGateway) PaymentUsecase {
	return &paymentUsecase{gateway: gw}
}

func (u *paymentUsecase) CheckStatus(ctx context.Context, transactionID string) (*entity.PaymentStatus, error) {
	return u.gateway.GetPaymentStatus(ctx, transactionID)
}
