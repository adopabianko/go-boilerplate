package main

import (
	"context"
	"log"
	"net"

	pb "go-boilerplate/pkg/pb/payment"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedPaymentServiceServer
}

// CheckStatus implements payment.PaymentServiceServer
func (s *server) CheckStatus(ctx context.Context, req *pb.CheckStatusRequest) (*pb.CheckStatusResponse, error) {
	log.Printf("Received CheckStatus request for Transaction ID: %s", req.TransactionId)

	// Simulate logic based on ID
	status := "SUCCESS"
	if req.TransactionId == "fail" {
		status = "FAILED"
	} else if req.TransactionId == "pending" {
		status = "PENDING"
	}

	return &pb.CheckStatusResponse{
		Id:     req.TransactionId,
		Status: status,
		Amount: 150000.00,
	}, nil
}

func main() {
	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPaymentServiceServer(s, &server{})

	log.Printf("Dummy Payment gRPC Server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
