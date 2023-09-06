package services

import (
	//"log"
	"golang.org/x/net/context"
	proto_golang_grpc_v1 "golang_grpc/proto/golang_grpc/v1"
)

func (s *FirstService) Echo(ctx context.Context, req *proto_golang_grpc_v1.FirstServiceEchoRequest) (*proto_golang_grpc_v1.FirstServiceEchoResponse, error) {
	//log.Printf("Echo...")

	return &proto_golang_grpc_v1.FirstServiceEchoResponse{Msg: "First: " + req.Name}, nil
}
