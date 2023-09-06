package micro

import (
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
)

// Option - service functional option
//
// See this post about the "functional options" pattern:
// http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(s *Service)

// Debug - return an Option to set the service to debug mode or not
func Debug(flag bool) Option {
	return func(s *Service) {
		s.debug = flag
	}
}

// StaticDir - return an Option to set the staticDir
func StaticDir(staticDir string) Option {
	return func(s *Service) {
		s.staticDir = staticDir
	}
}

// Redoc - return an Option to set the Redoc
func Redoc(redoc *RedocOpts) Option {
	return func(s *Service) {
		s.redoc = redoc
	}
}

// Annotator - return an Option to append an annotator
func Annotator(annotator AnnotatorFunc) Option {
	return func(s *Service) {
		s.annotators = append(s.annotators, annotator)
	}
}

// HTTPHandler - return an Option to set the httpHandler
func HTTPHandler(httpHandler HTTPHandlerFunc) Option {
	return func(s *Service) {
		s.httpHandler = httpHandler
	}
}

// UnaryInterceptor - return an Option to append an unaryInterceptor
func UnaryInterceptor(unaryInterceptor grpc.UnaryServerInterceptor) Option {
	return func(s *Service) {
		s.unaryInterceptors = append(s.unaryInterceptors, unaryInterceptor)
	}
}

// StreamInterceptor - return an Option to append an streamInterceptor
func StreamInterceptor(streamInterceptor grpc.StreamServerInterceptor) Option {
	return func(s *Service) {
		s.streamInterceptors = append(s.streamInterceptors, streamInterceptor)
	}
}

// RouteOpt - return an Option to append a route
func RouteOpt(route Route) Option {
	return func(s *Service) {
		s.routes = append(s.routes, route)
	}
}

// ShutdownFunc - return an Option to register a function which will be called when server shutdown
func ShutdownFunc(f func()) Option {
	return func(s *Service) {
		s.shutdownFunc = f
	}
}

// ShutdownTimeout - return an Option to set the timeout before the server shutdown abruptly
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Service) {
		s.shutdownTimeout = timeout
	}
}

// PreShutdownDelay - return an Option to set the time waiting for running goroutines
// to finish their jobs before the shutdown starts
func PreShutdownDelay(timeout time.Duration) Option {
	return func(s *Service) {
		s.preShutdownDelay = timeout
	}
}

// InterruptSignal - return an Option to append a interrupt signal
func InterruptSignal(signal os.Signal) Option {
	return func(s *Service) {
		s.interruptSignals = append(s.interruptSignals, signal)
	}
}

// GRPCServerOption - return an Option to append a gRPC server option
func GRPCServerOption(serverOption grpc.ServerOption) Option {
	return func(s *Service) {
		s.grpcServerOptions = append(s.grpcServerOptions, serverOption)
	}
}

// GRPCDialOption - return an Option to append a gRPC dial option
func GRPCDialOption(dialOption grpc.DialOption) Option {
	return func(s *Service) {
		s.grpcDialOptions = append(s.grpcDialOptions, dialOption)
	}
}

// WithHTTPServer - return an Option to set the http server, note that the Addr and Handler will be
// reset in startGRPCGateway(), so you are not able to specify them
func WithHTTPServer(server *http.Server) Option {
	return func(s *Service) {
		s.HTTPServer = server
	}
}

func (s *Service) apply(opts ...Option) {
	for _, opt := range opts {
		opt(s)
	}
}
