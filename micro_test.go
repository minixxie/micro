package micro

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

var reverseProxyFunc ReverseProxyFunc
var httpPort, grpcPort uint16
var shutdownFunc func()

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

	shutdownFunc = func() {
		fmt.Println("Server shutting down")
	}
}

func TestNewService(t *testing.T) {

	closer, _ := InitJaeger("micro", "localhost:6831", "localhost:6831", true)
	if closer != nil {
		defer closer.Close()
	}

	redoc := &RedocOpts{
		Route: "docs",
		Up:    true,
	}
	redoc.AddSpec("PetStore", "https://rebilly.github.io/ReDoc/swagger.yaml")

	// add the /test endpoint
	route := Route{
		Method:  "GET",
		Pattern: PathPattern("test"),
		Handler: func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			w.Write([]byte("Hello!"))
		},
	}

	s := NewService(
		Redoc(redoc),
		Debug(true),
		RouteOpt(route),
		ShutdownFunc(shutdownFunc),
		PreShutdownDelay(0),
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
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/", httpPort))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Len(t, resp.Header.Get("X-Request-Id"), 36)

	client = &http.Client{}
	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/fake.swagger.json", httpPort))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Len(t, resp.Header.Get("X-Request-Id"), 36)

	client = &http.Client{}
	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/demo.swagger.json", httpPort))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, resp.Header.Get("X-Request-Id"), 36)

	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/docs", httpPort))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, resp.Header.Get("X-Request-Id"), 36)

	resp, err = client.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", httpPort))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, resp.Header.Get("X-Request-Id"), 36)

	// create a root span and set uber-trace-id in header
	rootSpan := opentracing.StartSpan("root")
	client = &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/test", httpPort), nil)
	if err != nil {
		t.Error(err)
	}
	footprint := "1234567890"
	req.Header.Set("uber-trace-id", fmt.Sprintf("%+v", rootSpan))
	req.Header.Set("uberctx-footprint", footprint)
	resp, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, footprint, resp.Header.Get("X-Request-Id"))
	assert.Equal(t, "Hello!", string(body))
	rootSpan.Finish()

	// create service s2 to trigger errChan1
	s2 := NewService(
		Redoc(&RedocOpts{
			Up: false,
		}),
	)

	// grpc port 9999 alreday in use
	err = s2.Start(httpPort, grpcPort, reverseProxyFunc)
	assert.Error(t, err)

	// create service s3 to trigger errChan2
	s3 := NewService(
		Redoc(&RedocOpts{
			Up: false,
		}),
	)

	// http port 8888 already in use
	s.GRPCServer.Stop()
	err = s3.Start(httpPort, grpcPort, reverseProxyFunc)
	assert.Error(t, err)

	// wait 1 second for s3 gRPC server start
	time.Sleep(1 * time.Second)

	// close all previous services
	s.HTTPServer.Close()
	s3.GRPCServer.Stop()

	// run a new service, we use different ports to make sure ci not complain
	httpPort = 18888
	grpcPort = 19999
	s4 := NewService(
		Redoc(&RedocOpts{
			Up: false,
		}),
		ShutdownTimeout(10*time.Second),
	)
	go func() {
		if err := s4.Start(httpPort, grpcPort, reverseProxyFunc); err != nil {
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
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Len(t, resp.Header.Get("X-Request-Id"), 36)

	// send an interrupt signal to stop s4
	syscall.Kill(s4.Getpid(), syscall.SIGINT)

	// wait 3 second for the server shutdown
	time.Sleep(3 * time.Second)
}

func TestErrorReverseProxyFunc(t *testing.T) {
	s := NewService(
		Redoc(&RedocOpts{
			Up: true,
		}),
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

	err := s.startGRPCGateway(httpPort, grpcPort, reverseProxyFunc)
	assert.EqualError(t, err, errText)
}

func TestDefaultAnnotator(t *testing.T) {
	span := opentracing.StartSpan("root")
	ctx := opentracing.ContextWithSpan(context.TODO(), span)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-Id", "uuid")

	md := DefaultAnnotator(ctx, req)
	id, ok := md["x-request-id"]

	assert.True(t, ok)
	assert.Equal(t, "uuid", id[0])
}
