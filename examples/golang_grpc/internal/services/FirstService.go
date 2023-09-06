package services

import (
	proto_golang_grpc_v1 "golang_grpc/proto/golang_grpc/v1"

	"golang_grpc/internal/models"
)

type FirstService struct {
	proto_golang_grpc_v1.UnimplementedFirstServiceServer
	FirstModel models.FirstModel
}

var _ proto_golang_grpc_v1.FirstServiceServer = (*FirstService)(nil) // make sure this struct implements the interface
