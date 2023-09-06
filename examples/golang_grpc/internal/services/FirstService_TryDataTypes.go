package services

import (
	// "log"
	"golang.org/x/net/context"
	proto_golang_grpc_v1 "golang_grpc/proto/golang_grpc/v1"
)

func (s *FirstService) TryDataTypes(ctx context.Context, req *proto_golang_grpc_v1.TryDataTypesRequest) (*proto_golang_grpc_v1.TryDataTypesResponse, error) {
	// log.Printf("TryDataTypes...")

	return &proto_golang_grpc_v1.TryDataTypesResponse{
		// Time: req.Time,
	}, nil
}
