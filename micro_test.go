package micro

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewService(t *testing.T) {
	s := NewService(
		[]grpc.StreamServerInterceptor{},
		[]grpc.UnaryServerInterceptor{},
	)

	SwaggerFile = "./swagger_demo.json"

	var httpPort, grpcPort uint16
	httpPort = 8888
	grpcPort = 9999
	go func() {
		if err := s.Start(httpPort, grpcPort, func(
			ctx context.Context,
			mux *runtime.ServeMux,
			grpcHostAndPort string,
			opts []grpc.DialOption,
		) error {
			return nil
		}); err != nil {
			t.Errorf("failed to serve: %v", err)
		}
	}()

	// wait 1 second for the server start
	time.Sleep(1 * time.Second)

	client := &http.Client{}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/swagger.json", httpPort))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 200, resp.StatusCode)
}
