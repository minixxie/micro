package micro

import (
	"context"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewService(t *testing.T) {
	s := NewService(
		[]grpc.StreamServerInterceptor{},
		[]grpc.UnaryServerInterceptor{},
	).UpRedoc(true)
	err = s.Start(80, 8080, func(ctx context.Context, mux *runtime.ServeMux, grpcHostAndPort string, opts []grpc.DialOption) error {
		// var err error
		// // FirstService Proxy
		// err = myservice_v1.RegisterFirstServiceHandlerFromEndpoint(
		// 	ctx, mux, grpcHostAndPort, opts)
		// if err != nil {
		// 	return err
		// }
		return nil
	})
	assert.Equal(t, err, nil)
}
