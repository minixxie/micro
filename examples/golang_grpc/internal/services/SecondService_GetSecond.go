package services

import (
	"golang.org/x/net/context"
	proto_golang_grpc_v1 "golang_grpc/proto/golang_grpc/v1"
	"log"
)

func (s *SecondService) GetSecond(ctx context.Context, req *proto_golang_grpc_v1.GetSecondRequest) (*proto_golang_grpc_v1.GetSecondResponse, error) {
	log.Printf("GetSecond...")

	return &proto_golang_grpc_v1.GetSecondResponse{Msg: "GetSecond: " + req.Name}, nil
}
