package services

import (
	proto_golang_grpc_v1 "golang_grpc/proto/golang_grpc/v1"
)

type SecondService struct {
	proto_golang_grpc_v1.UnimplementedSecondServiceServer
}

var _ proto_golang_grpc_v1.SecondServiceServer = (*SecondService)(nil) // make sure this structimplements the interface
