package micro

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func unaryPanic(ctx context.Context, req interface{}) (interface{}, error) {
	panic("panic in unary handler")
}

func streamPanic(srv interface{}, stream grpc.ServerStream) error {
	panic("panic in steam handler")
}

func TestUnaryPanicHandler(t *testing.T) {
	ctx := context.TODO()
	_, err := UnaryPanicHandler(ctx, nil, nil, unaryPanic)
	assert.Error(t, err)
}

func TestStreamPanicHandler(t *testing.T) {
	err := StreamPanicHandler(nil, nil, nil, streamPanic)
	assert.Error(t, err)
}
