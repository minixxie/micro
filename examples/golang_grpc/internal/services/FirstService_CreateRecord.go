package services

import (
	//"log"
	"golang.org/x/net/context"
	proto_golang_grpc_v1 "golang_grpc/proto/golang_grpc/v1"
)

func (s *FirstService) CreateRecord(ctx context.Context, req *proto_golang_grpc_v1.CreateRecordRequest) (*proto_golang_grpc_v1.CreateRecordResponse, error) {
	//log.Printf("Echo...")

	id, err := s.FirstModel.CreateRecord(req.Name)
	if err != nil {
		return nil, err
	}

	return &proto_golang_grpc_v1.CreateRecordResponse{Id: id}, nil
}
