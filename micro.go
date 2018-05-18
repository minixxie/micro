package micro

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strconv"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Service - to represent the microservice
type Service struct {
	GRPCServer *grpc.Server

	upRedoc            bool
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor
}

// NewService - to create the microservice object
func NewService(streamInterceptors []grpc.StreamServerInterceptor, unaryInterceptors []grpc.UnaryServerInterceptor) *Service {
	s := Service{}
	s.upRedoc = false

	tracer := opentracing.GlobalTracer()

	s.streamInterceptors = []grpc.StreamServerInterceptor{}
	s.streamInterceptors = append(s.streamInterceptors, grpc_prometheus.StreamServerInterceptor)
	s.streamInterceptors = append(s.streamInterceptors, grpc_validator.StreamServerInterceptor())
	s.streamInterceptors = append(s.streamInterceptors, otgrpc.OpenTracingStreamServerInterceptor(tracer))
	s.streamInterceptors = append(s.streamInterceptors, streamInterceptors...)

	s.unaryInterceptors = []grpc.UnaryServerInterceptor{}
	s.unaryInterceptors = append(s.unaryInterceptors, grpc_prometheus.UnaryServerInterceptor)
	s.unaryInterceptors = append(s.unaryInterceptors, grpc_validator.UnaryServerInterceptor())
	s.unaryInterceptors = append(s.unaryInterceptors, otgrpc.OpenTracingServerInterceptor(tracer))
	s.unaryInterceptors = append(s.unaryInterceptors, unaryInterceptors...)

	s.GRPCServer = grpc.NewServer(
		grpc_middleware.WithStreamServerChain(s.streamInterceptors...),
		grpc_middleware.WithUnaryServerChain(s.unaryInterceptors...),
	)

	return &s
}

// UpRedoc - to configure redoc server to run up at /docs endpoint
func (s *Service) UpRedoc(up bool) *Service {
	s.upRedoc = up
	return s
}

// ReverseProxyFunc - a callback that the caller should implement to steps to reverse-proxy the HTTP/1 requests to gRPC
type ReverseProxyFunc func(ctx context.Context, mux *runtime.ServeMux, grpcHostAndPort string, opts []grpc.DialOption) error

// Start - to start the microservice with listening on the ports
func (s *Service) Start(httpPort uint16, grpcPort uint16, reverseProxyFunc ReverseProxyFunc) error {
	// start http1.0 server & swagger server in the background
	go func() {
		// Start swagger server at :8888
		if s.upRedoc {
			err := exec.Command(
				"/go/bin/swagger",
				"serve", "--no-open", "--base-path=/", "-F", "redoc", "-p", "8888", "/swagger.json",
			).Start()
			if err != nil {
				return
			}
		}
		// Start HTTP/1.0 server at :80
		if err := grpcGateway(grpcPort, httpPort, reverseProxyFunc); err != nil {
			return
		}
	}()

	// Setup /metrics for prometheus
	grpc_prometheus.Register(s.GRPCServer)

	// Register reflection service on gRPC server.
	reflection.Register(s.GRPCServer)

	grpcHost := strings.Join([]string{":", strconv.FormatUint(uint64(grpcPort), 10)}, "")
	lis, err := net.Listen("tcp", grpcHost)
	if err != nil {
		return err
	}
	if err = s.GRPCServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func grpcGateway(grpcPort uint16, httpPort uint16, reverseProxyFunc ReverseProxyFunc) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))

	opts := []grpc.DialOption{grpc.WithInsecure()}

	// configure /metrics HTTP/1 endpoint
	patternMetrics := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{"metrics"}, ""))
	mux.Handle("GET", patternMetrics, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	// configure /docs HTTP/1 endpoint
	url, _ := url.Parse("http://127.0.0.1:8888")
	proxyToSwaggerServer := httputil.NewSingleHostReverseProxy(url)
	patternRedoc := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{"docs"}, ""))
	mux.Handle("GET", patternRedoc, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		proxyToSwaggerServer.ServeHTTP(w, r)
	})
	// configure /swagger.json HTTP/1 endpoint
	patternSwaggerJSON := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{"swagger.json"}, ""))
	mux.Handle("GET", patternSwaggerJSON, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		proxyToSwaggerServer.ServeHTTP(w, r)
	})

	err := reverseProxyFunc(ctx, mux, strings.Join([]string{"localhost:", strconv.FormatUint(uint64(grpcPort), 10)}, ""), opts)
	if err != nil {
		return err
	}

	// var err error
	// // WalletService proxy
	// err = lalamove_walletService_v1.RegisterWalletServiceHandlerFromEndpoint(
	// 	ctx, mux, strings.Join([]string{"localhost:", strconv.FormatUint(grpcPort, 10)}, ""), opts)
	// if err != nil {
	// 	return err
	// }

	return http.ListenAndServe(strings.Join([]string{":", strconv.FormatUint(uint64(httpPort), 10)}, ""), mux)
}