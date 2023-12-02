// ref: https://opentelemetry.io/docs/demo/services/checkout/
package micro

import (
	"context"
	"os"
	"log"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"sync"
)

var resource *sdkresource.Resource
var initResourcesOnce sync.Once

func initResource() *sdkresource.Resource {
	initResourcesOnce.Do(func() {
		extraResources, _ := sdkresource.New(
			context.Background(),
			sdkresource.WithOS(),
			sdkresource.WithProcess(),
			sdkresource.WithContainer(),
			sdkresource.WithHost(),
		)
		resource, _ = sdkresource.Merge(
			sdkresource.Default(),
			extraResources,
		)
	})
	return resource
}

func (s *Service) initOpenTelemetry() {
	s.meterProvider = initMeterProvider()
	s.tracerProvider = initTracerProvider()
}

func (s *Service) shutdownOpenTelemetry() {
	if err := s.meterProvider.Shutdown(context.Background()); err != nil {
		log.Printf("Error shutting down meter provider: %v\n", err)
	}
	if err := s.tracerProvider.Shutdown(context.Background()); err != nil {
		log.Printf("Error shutting down tracer provider: %v\n", err)
	}
}

func initMeterProvider() *sdkmetric.MeterProvider {
	log.Printf("initMeterProvider()...");
	ctx := context.Background()

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		otlpmetricgrpc.WithCompressor("gzip"),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("new otlp metric grpc exporter failed: %v\n", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(initResource()),
	)
	otel.SetMeterProvider(mp)
	return mp
}

func initTracerProvider() *sdktrace.TracerProvider {
	log.Printf("initTracerProvider()...");
	ctx := context.Background()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		otlptracegrpc.WithCompressor("gzip"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(initResource()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}
