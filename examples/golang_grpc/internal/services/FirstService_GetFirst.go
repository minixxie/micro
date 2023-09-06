package services

import (
	// "log"
	"golang.org/x/net/context"
	proto_golang_grpc_v1 "golang_grpc/proto/golang_grpc/v1"
)

func (s *FirstService) GetFirst(ctx context.Context, req *proto_golang_grpc_v1.GetFirstRequest) (*proto_golang_grpc_v1.GetFirstResponse, error) {
	// log.Printf("GetFirst...")

	return &proto_golang_grpc_v1.GetFirstResponse{Msg: "GetFirst: " + req.Name}, nil
}
