package micro

import (
	"context"
	"log"
	"net"
	"os"
	"net/http"
	"strconv"
	"strings"

	// grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	// grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	// grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	// "google.golang.org/grpc/reflection"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// SwaggerFile is the swagger file (local path)
const SwaggerFile = "/swagger.json"

// Service - to represent the microservice
type Service struct {
	GRPCServer         *grpc.Server
        grpcServices       []grpcService

	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor

	upRedoc            bool
	grpcGatewayPort    uint16
	grpcPort           uint16

	// OTEL Meter
	meterProvider *sdkmetric.MeterProvider
	// OTEL Trace
	tracerProvider *sdktrace.TracerProvider

}

// ReverseProxyFunc - a callback that the caller should implement to steps to reverse-proxy the HTTP/1 requests to gRPC
// type ReverseProxyFunc func(ctx context.Context, mux *runtime.ServeMux, grpcHostAndPort string, opts []grpc.DialOption) error
type RegisterServiceHandlerFunc func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error

type grpcService struct {
	serviceDesc *grpc.ServiceDesc
	srv interface{}
	registerServiceHandlerFunc RegisterServiceHandlerFunc
}

// NewService - to create the microservice object
func NewService() *Service {
	s := Service{}

	s.upRedoc = os.Getenv("MICRO_REDOC") == "1"
	s.grpcGatewayPort = 80
	s.grpcPort = 9090

	s.initOpenTelemetry()

	/*
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
	)*/

	return &s
}

func (s *Service) SetGRPCGatewayPort(port uint16) {
	s.grpcGatewayPort = port
}

func (s *Service) SetGRPCPort(port uint16) {
	s.grpcPort = port
}

func (s *Service) AddService(serviceDesc *grpc.ServiceDesc, srv interface{}, registerServiceHandlerFunc RegisterServiceHandlerFunc) {
	s.grpcServices = append(s.grpcServices, grpcService{
		serviceDesc: serviceDesc,
		srv: srv,
		registerServiceHandlerFunc: registerServiceHandlerFunc,
	})
}

// UpRedoc - to configure redoc server to run up at /docs endpoint
func (s *Service) UpRedoc(up bool) *Service {
	s.upRedoc = up
	return s
}


// Start - to start the microservice with listening on the ports
func (s *Service) Start() error {
	s.GRPCServer = grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		// grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		// grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	// start http1.0 server & swagger server in the background
	go func() {
		// Start HTTP/1.0 server at :80
		if err := s.grpcGateway(); err != nil {
			log.Fatalf("failed to start gRPC gateway: %v", err)
		}
	}()

	// Setup /metrics for prometheus
	//grpc_prometheus.Register(s.GRPCServer)

	// Register reflection service on gRPC server.
	//reflection.Register(s.GRPCServer)

	for _, grpcService := range s.grpcServices {
		s.GRPCServer.RegisterService(grpcService.serviceDesc, grpcService.srv)
	}

	grpcHost := strings.Join([]string{":", strconv.FormatUint(uint64(s.grpcPort), 10)}, "")
	lis, err := net.Listen("tcp", grpcHost)
	if err != nil {
		return err
	}

	return s.GRPCServer.Serve(lis)
}

func (s *Service) Stop() {
	s.shutdownOpenTelemetry()
}

func (s *Service) grpcGateway() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}))

	// opts := []grpc.DialOption{grpc.WithInsecure()}

	// configure /metrics HTTP/1 endpoint
	patternMetrics := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{"metrics"}, ""))
	mux.Handle("GET", patternMetrics, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	// configure /docs HTTP/1 endpoint
	patternRedoc := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{"docs"}, ""))
	mux.Handle("GET", patternRedoc, redoc)

	// configure /swagger.json HTTP/1 endpoint
	patternSwaggerJSON := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{"swagger.json"}, ""))
	mux.Handle("GET", patternSwaggerJSON, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		http.ServeFile(w, r, SwaggerFile)
	})

	grpcHostAndPort := strings.Join([]string{":", strconv.FormatUint(uint64(s.grpcPort), 10)}, "")
	conn, err := grpc.DialContext(
		context.Background(),
		grpcHostAndPort,
		grpc.WithInsecure(),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		// grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		// grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		log.Printf("Cannot establish gRPC connection from gRPC gateway: %v\n", err)
		return err
	}

	for _, grpcService := range s.grpcServices {
		err := grpcService.registerServiceHandlerFunc(ctx, mux, conn)
		if err != nil {
			return err
		}
	}

	var handler http.Handler
	handler = otelhttp.NewHandler(mux, "grpc-gateway: mux.ServeHTTP()")
	handler = newTraceparentHandler(handler)

	return http.ListenAndServe(strings.Join([]string{":", strconv.FormatUint(uint64(s.grpcGatewayPort), 10)}, ""), handler)
}

// https://uptrace.dev/opentelemetry/opentelemetry-traceparent.html#injecting-traceparent-header
type traceparentHandler struct {
	next  http.Handler
	props propagation.TextMapPropagator
}

func newTraceparentHandler(next http.Handler) *traceparentHandler {
	return &traceparentHandler{
		next:  next,
		props: otel.GetTextMapPropagator(),
	}
}

func (h *traceparentHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// https://faun.pub/multi-hop-tracing-with-opentelemetry-in-golang-792df5feb37c
	// Extract Span/Trace info from the request header
	ctx := otel.GetTextMapPropagator().Extract(
		req.Context(), propagation.HeaderCarrier(req.Header),
	)

	tracer := otel.GetTracerProvider().Tracer("")
	ctx, span := tracer.Start(ctx, "grpc-gateway: " + req.Method + " " + req.URL.String(),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	h.props.Inject(ctx, propagation.HeaderCarrier(w.Header()))

	h.next.ServeHTTP(w, req.WithContext(ctx))
}
