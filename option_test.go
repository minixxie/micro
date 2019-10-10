package micro

import (
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestStaticDir(t *testing.T) {
	s := NewService(StaticDir("/a/b/c"))
	assert.Equal(t, "/a/b/c", s.staticDir)
}

func TestAnnotator(t *testing.T) {
	s := NewService(
		Annotator(func(ctx context.Context, req *http.Request) metadata.MD {
			md := metadata.New(nil)
			md.Set("key", "value")
			return md
		}),
	)

	assert.Len(t, s.annotators, 2)
}

func TestErrorHandler(t *testing.T) {
	s := NewService(ErrorHandler(nil))
	assert.Nil(t, s.errorHandler)
}

func TestHTTPHandler(t *testing.T) {
	s := NewService(HTTPHandler(nil))
	assert.Nil(t, s.httpHandler)
}

func TestUnaryInterceptor(t *testing.T) {
	s := NewService(
		UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			return nil, nil
		}),
	)

	assert.Len(t, s.unaryInterceptors, 5)
}

func TestStreamInterceptor(t *testing.T) {
	s := NewService(
		StreamInterceptor(func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return nil
		}),
	)

	assert.Len(t, s.streamInterceptors, 5)
}

func TestInterruptSignal(t *testing.T) {
	s := NewService(
		InterruptSignal(syscall.SIGKILL),
	)

	assert.Len(t, s.interruptSignals, 5)
}

func TestGRPCServerOption(t *testing.T) {
	s := NewService(
		GRPCServerOption(grpc.ConnectionTimeout(10 * time.Second)),
	)

	assert.Len(t, s.grpcServerOptions, 3)
}

func TestGRPCDialOption(t *testing.T) {
	s := NewService(
		GRPCDialOption(grpc.WithBlock()),
	)

	assert.Len(t, s.grpcDialOptions, 1)
}

func TestWithHTTPServer(t *testing.T) {
	s := NewService(WithHTTPServer(&http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}))
	assert.NotNil(t, s.HTTPServer)
	assert.Equal(t, 5*time.Second, s.HTTPServer.ReadTimeout)
}
