package services

import (
	"golang.org/x/net/context"
	proto_golang_grpc_v1 "golang_grpc/proto/golang_grpc/v1"
	"log"
)

func (s *SecondService) Echo(ctx context.Context, req *proto_golang_grpc_v1.SecondServiceEchoRequest) (*proto_golang_grpc_v1.SecondServiceEchoResponse, error) {
	log.Printf("Echo...")

	return &proto_golang_grpc_v1.SecondServiceEchoResponse{Msg: "Second: " + req.Name}, nil
}
