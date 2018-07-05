package micro

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
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
		true,
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
	assert.Len(t, resp.Header.Get("X-Request-Id"), 36)

	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", httpPort))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 200, resp.StatusCode)
	assert.Len(t, resp.Header.Get("X-Request-Id"), 36)

	// another service
	s2 := NewService(
		[]grpc.StreamServerInterceptor{},
		[]grpc.UnaryServerInterceptor{},
		false,
	)

	// http port 8888 already in use
	err = s2.startGrpcGateway(httpPort, grpcPort, reverseProxyFunc)
	assert.Error(t, err)

	// grpc port 9999 alreday in use
	err = s2.startGrpcServer(grpcPort)
	assert.Error(t, err)

	// stop the first server
	s.Stop()

	// run a new service again
	s = NewService(
		[]grpc.StreamServerInterceptor{},
		[]grpc.UnaryServerInterceptor{},
		false,
	)
	go func() {
		if err := s.Start(httpPort, grpcPort, reverseProxyFunc); err != nil {
			t.Errorf("failed to serve: %v", err)
		}
	}()

	// wait 1 second for the server start
	time.Sleep(1 * time.Second)

	// the redoc is not up for the second server
	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/docs", httpPort))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	assert.Len(t, resp.Header.Get("X-Request-Id"), 36)
}

func TestErrorReverseProxyFunc(t *testing.T) {
	s := NewService(
		[]grpc.StreamServerInterceptor{},
		[]grpc.UnaryServerInterceptor{},
		false,
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

	err := s.startGrpcGateway(httpPort, grpcPort, reverseProxyFunc)
	assert.EqualError(t, err, errText)
}

func TestAnnotator(t *testing.T) {
	ctx := context.TODO()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-Id", "uuid")
	md := Annotator(ctx, req)
	id, ok := md["x-request-id"]
	assert.True(t, ok)
	assert.Equal(t, "uuid", id[0])
}
