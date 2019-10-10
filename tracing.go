package micro

import (
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

// InitSpan - initiate the tracing span and set the http response header with X-Request-Id
func InitSpan(mux *runtime.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var serverSpan opentracing.Span

		// Extracting B3 tracing context from the request.
		// This step is important to extract the actual request context from outside of the applications.
		// By default, Jaeger use "uber-trace-id" to propagate tracing context,
		// and use prefix "uberctx-" to propagate baggage in http headers.
		// See https://github.com/jaegertracing/jaeger-client-go/blob/master/constants.go
		var wireContext, err = opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header),
		)
		// We will be using method name as the span name (method above)
		var methodName = r.Method + " " + r.URL.Path
		if err != nil {
			// Found no span in headers, start a new span as root span
			Logger().Infof(err.Error())
			for k, h := range r.Header {
				for _, v := range h {
					Logger().Infof("Header: %s - %s", k, v)
				}
			}
			serverSpan = opentracing.StartSpan(methodName)
		} else {
			// Create span as a child of parent context
			Logger().Infof("Found parent span, start a child span: " + methodName)
			serverSpan = opentracing.StartSpan(
				methodName,
				opentracing.ChildOf(wireContext),
			)
		}
		serverSpan.SetTag("http.url.host", r.URL.Hostname())
		serverSpan.SetTag("peer.address", r.RemoteAddr)
		serverSpan.SetTag("http.url", r.URL.RequestURI())
		serverSpan.SetTag("http.url.query", r.URL.RawQuery)

		var footprint string
		if footprint = serverSpan.BaggageItem("footprint"); footprint != "" {
			Logger().Infof("Found baggage item footprint in span: " + footprint)
			serverSpan.SetTag("footprint", footprint)
		} else {
			footprint = RequestID(r)
			Logger().Infof("No baggage item footprint found in span, try to get from X-Request-Id: " + footprint)
			serverSpan.SetBaggageItem("footprint", footprint)
			serverSpan.SetTag("footprint", footprint)
		}

		// Set the http response header with X-Request-Id
		w.Header().Set("X-Request-Id", footprint)

		// We are passing the span as an item in Go context
		Logger().Infof("Passing span into context: %+v", serverSpan)
		var ctx = opentracing.ContextWithSpan(r.Context(), serverSpan)

		mux.ServeHTTP(w, r.WithContext(ctx))

		// Span needs to be finished in order to report it to Jaeger collector
		serverSpan.Finish()
	})
}

// InitJaeger - helper to initiate an instance of Jaeger Tracer as global tracer, if you need to
// customize your tracer, you can do it yourself instead of calling this function
func InitJaeger(service, samplingServerURL, localAgentHost string, debug bool) (io.Closer, error) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:              jaeger.SamplerTypeConst,
			Param:             1,
			SamplingServerURL: samplingServerURL,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: localAgentHost,
		},
	}

	l := config.Logger(jaeger.NullLogger)
	if debug { // only log to stdout in debug mode
		l = config.Logger(jaeger.StdLogger)
	}

	return cfg.InitGlobalTracer(service, l, config.ZipkinSharedRPCSpan(true))
}
