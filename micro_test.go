package micro

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

var reverseProxyFunc ReverseProxyFunc
var httpPort, grpcPort uint16

func init() {
	reverseProxyFunc = func(
		ctx context.Context,
		mux *runtime.ServeMux,
		grpcHostAndPort string,
		opts []grpc.DialOption,
	) error {
		return nil
	}

	httpPort = 8888
	grpcPort = 9999

	SwaggerFile = "./swagger_demo.json"
}

func TestNewService(t *testing.T) {
	s := NewService(
		[]grpc.StreamServerInterceptor{},
		[]grpc.UnaryServerInterceptor{},
	)

	go func() {
		if err := s.Start(httpPort, grpcPort, reverseProxyFunc); err != nil {
			t.Errorf("failed to serve: %v", err)
		}
	}()

	// wait 1 second for the server start
	time.Sleep(1 * time.Second)

	// check if the http server is up
	httpHost := fmt.Sprintf(":%d", httpPort)
	_, err := net.Listen("tcp", httpHost)
	assert.Error(t, err)

	// check if the grpc server is up
	grpcHost := fmt.Sprintf(":%d", grpcPort)
	_, err = net.Listen("tcp", grpcHost)
	assert.Error(t, err)

	// check if the http endpoint works
	client := &http.Client{}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/swagger.json", httpPort))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 200, resp.StatusCode)
}

func TestPortsUnavailable(t *testing.T) {
	s := NewService(
		[]grpc.StreamServerInterceptor{},
		[]grpc.UnaryServerInterceptor{},
	)

	// ports 8888 and 9999 already in use
	err := s.Start(httpPort, grpcPort, reverseProxyFunc)
	assert.Error(t, err)
}

func TestErrorReverseProxyFunc(t *testing.T) {
	s := NewService(
		[]grpc.StreamServerInterceptor{},
		[]grpc.UnaryServerInterceptor{},
	)

	// mock error from reverseProxyFunc
	errText := "reverse proxy func error"
	reverseProxyFunc = func(
		ctx context.Context,
		mux *runtime.ServeMux,
		grpcHostAndPort string,
		opts []grpc.DialOption,
	) error {
		return errors.New(errText)
	}

	err := s.Start(httpPort, grpcPort, reverseProxyFunc)
	assert.Error(t, err)
}
