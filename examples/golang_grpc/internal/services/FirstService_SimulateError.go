package services

import (
	"errors"
	"golang.org/x/net/context"
	proto_golang_grpc_v1 "golang_grpc/proto/golang_grpc/v1"
	"log"
)

func (s *FirstService) SimulateError(ctx context.Context, req *proto_golang_grpc_v1.SimulateErrorRequest) (*proto_golang_grpc_v1.SimulateErrorResponse, error) {
	log.Printf("SimulateError...")

	return nil, errors.New("E_SERVER_ERROR")
}
