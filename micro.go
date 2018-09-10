package micro

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/uuid"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

// Service - to represent the microservice
type Service struct {
	GRPCServer         *grpc.Server
	HTTPServer         *http.Server
	Mux                *runtime.ServeMux
	HTTPHandler        HTTPHandlerFunc
	ErrorHandler       runtime.ProtoErrorHandlerFunc
	Annotators         []AnnotatorFunc
	Redoc              *RedocOpts
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor
}

// ReverseProxyFunc - a callback that the caller should implement to steps to reverse-proxy the HTTP/1 requests to gRPC
type ReverseProxyFunc func(ctx context.Context, mux *runtime.ServeMux, grpcHostAndPort string, opts []grpc.DialOption) error

// HTTPHandlerFunc - http handler function
type HTTPHandlerFunc func(mux *runtime.ServeMux) http.Handler

// AnnotatorFunc - annotator function is for injecting meta data from http request into gRPC context
type AnnotatorFunc func(context.Context, *http.Request) metadata.MD

// DefaultHTTPHandler - default http handler which will set the http response header with X-Request-Id
func DefaultHTTPHandler(mux *runtime.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", RequestID(r))
		mux.ServeHTTP(w, r)
	})
}

// DefaultAnnotator - set the X-Request-Id into gRPC context
func DefaultAnnotator(c context.Context, req *http.Request) metadata.MD {
	return metadata.Pairs("x-request-id", RequestID(req))
}

// RequestID - get X-Request-Id from http request header, if it does not exist then generate one
func RequestID(req *http.Request) string {
	id := req.Header.Get("X-Request-Id")

	if id == "" {
		id = uuid.New().String()
	}

	// set it back into request header
	req.Header.Set("X-Request-Id", id)

	return id
}

// NewService - create a new microservice
func NewService(
	streamInterceptors []grpc.StreamServerInterceptor,
	unaryInterceptors []grpc.UnaryServerInterceptor,
	redoc *RedocOpts,
) *Service {
	s := Service{
		Redoc: redoc,
	}

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

// Start - start the microservice with listening on the ports
func (s *Service) Start(httpPort uint16, grpcPort uint16, reverseProxyFunc ReverseProxyFunc) error {

	errChan := make(chan error, 1)

	// start HTTP/1.0 gateway server
	go func() {
		errChan <- s.startGrpcGateway(httpPort, grpcPort, reverseProxyFunc)
	}()

	// start gRPC server
	go func() {
		errChan <- s.startGrpcServer(grpcPort)
	}()

	return <-errChan
}

func (s *Service) startGrpcServer(grpcPort uint16) error {
	// setup /metrics for prometheus
	grpc_prometheus.Register(s.GRPCServer)

	// register reflection service on gRPC server.
	reflection.Register(s.GRPCServer)

	grpcHost := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", grpcHost)
	if err != nil {
		return err
	}

	return s.GRPCServer.Serve(lis)
}

func (s *Service) startGrpcGateway(httpPort uint16, grpcPort uint16, reverseProxyFunc ReverseProxyFunc) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if s.ErrorHandler == nil {
		s.ErrorHandler = runtime.DefaultHTTPError
	}

	if s.Annotators == nil || len(s.Annotators) == 0 {
		s.Annotators = append(s.Annotators, DefaultAnnotator)
	}
	var muxOptions []runtime.ServeMuxOption
	muxOptions = append(muxOptions, runtime.WithMarshalerOption(
		runtime.MIMEWildcard,
		&runtime.JSONPb{OrigName: true, EmitDefaults: true},
	))
	muxOptions = append(muxOptions, runtime.WithProtoErrorHandler(s.ErrorHandler))

	for _, annotator := range s.Annotators {
		muxOptions = append(muxOptions, runtime.WithMetadata(annotator))
	}

	if s.Mux == nil { // set a default mux
		s.Mux = runtime.NewServeMux(muxOptions...)
	}

	if s.HTTPHandler == nil { // set a default http handler
		s.HTTPHandler = DefaultHTTPHandler
	}

	opts := []grpc.DialOption{grpc.WithInsecure()}

	// configure /metrics HTTP/1 endpoint
	patternMetrics := runtime.MustPattern(runtime.NewPattern(1, []int{int(utilities.OpLitPush), 0}, []string{"metrics"}, ""))
	s.Mux.Handle("GET", patternMetrics, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	if s.Redoc.Up {
		// configure /docs HTTP/1 endpoint
		patternRedoc := runtime.MustPattern(runtime.NewPattern(1, []int{int(utilities.OpLitPush), 0}, []string{"docs"}, ""))
		s.Mux.Handle("GET", patternRedoc, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			s.Redoc.Serve(w, r, pathParams)
		})
	}

	err := reverseProxyFunc(ctx, s.Mux, fmt.Sprintf("localhost:%d", grpcPort), opts)
	if err != nil {
		return err
	}

	// configure /swagger.json and /*.swagger.json HTTP/1 endpoints.
	// this is the fallback handler that will check if it's requesting swagger files,
	// if not matched or file not exists, then a 404 error will be returned.
	patternFallback := runtime.MustPattern(runtime.NewPattern(1, []int{int(utilities.OpPush), 0}, []string{""}, ""))
	s.Mux.Handle("GET", patternFallback, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		if matched, _ := regexp.MatchString(".*swagger\\.json", r.URL.Path); !matched {
			http.NotFound(w, r)
			return
		}

		dir, _ := os.Getwd()
		path := filepath.Join(dir, r.URL.Path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, path)
	})

	s.HTTPServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: s.HTTPHandler(s.Mux),
	}

	return s.HTTPServer.ListenAndServe()
}

// Stop - stop the microservice
func (s *Service) Stop() {
	s.GRPCServer.Stop()
	s.HTTPServer.Shutdown(context.Background())
}
