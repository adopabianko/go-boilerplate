package grpcgateway

import (
	"context"

	"go-boilerplate/internal/entity"
	pb "go-boilerplate/pkg/pb/payment"
)

type PaymentGateway interface {
	GetPaymentStatus(ctx context.Context, id string) (*entity.PaymentStatus, error)
}

type paymentWithGRPC struct {
	client pb.PaymentServiceClient
}

// Verify interface compliance
var _ PaymentGateway = (*paymentWithGRPC)(nil)

func NewPaymentGateway(client pb.PaymentServiceClient) PaymentGateway {
	return &paymentWithGRPC{client: client}
}

func (g *paymentWithGRPC) GetPaymentStatus(ctx context.Context, id string) (*entity.PaymentStatus, error) {
	req := &pb.CheckStatusRequest{TransactionId: id}

	resp, err := g.client.CheckStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	return &entity.PaymentStatus{
		ID:     resp.Id,
		Status: resp.Status,
		Amount: resp.Amount,
	}, nil
}
