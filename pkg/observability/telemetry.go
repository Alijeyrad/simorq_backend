package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// Config holds observability configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string

	// OTLP Trace Exporter
	OTLPEndpoint string // e.g., "localhost:4318" for OTLP HTTP
	OTLPInsecure bool

	// Sampling
	SamplingRate float64 // 0.0 to 1.0, default 1.0 (sample all)
}

// Provider holds the OpenTelemetry providers
type Provider struct {
	TracerProvider     *trace.TracerProvider
	MeterProvider      *metric.MeterProvider
	PrometheusExporter *prometheus.Exporter
}

// InitTelemetry initializes OpenTelemetry tracing and metrics
func InitTelemetry(ctx context.Context, cfg Config) (*Provider, error) {
	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"", // Use empty schema URL to inherit from Default()
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentName(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize tracing
	tracerProvider, err := initTracing(ctx, res, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	// Initialize metrics with Prometheus
	meterProvider, promExporter, err := initMetrics(res)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Set global providers
	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)

	// Set global propagator for trace context propagation (W3C Trace Context)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Provider{
		TracerProvider:     tracerProvider,
		MeterProvider:      meterProvider,
		PrometheusExporter: promExporter,
	}, nil
}

// initTracing sets up the OTLP trace exporter
func initTracing(ctx context.Context, res *resource.Resource, cfg Config) (*trace.TracerProvider, error) {
	// Default sampling rate
	samplingRate := cfg.SamplingRate
	if samplingRate == 0 {
		samplingRate = 1.0
	}

	var exporter trace.SpanExporter
	var err error

	// Only configure OTLP exporter if endpoint is provided
	if cfg.OTLPEndpoint != "" {
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		}

		if cfg.OTLPInsecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}

		exporter, err = otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
	}

	// Create tracer provider with sampling
	tp := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(samplingRate)),
	)

	// Add batch span processor if exporter is configured
	if exporter != nil {
		tp.RegisterSpanProcessor(trace.NewBatchSpanProcessor(exporter))
	}

	return tp, nil
}

// initMetrics sets up Prometheus metrics exporter
func initMetrics(res *resource.Resource) (*metric.MeterProvider, *prometheus.Exporter, error) {
	// Create Prometheus exporter
	promExporter, err := prometheus.New()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create meter provider
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(promExporter),
	)

	return meterProvider, promExporter, nil
}

// Shutdown gracefully shuts down the telemetry providers
func (p *Provider) Shutdown(ctx context.Context) error {
	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Shutdown tracer provider
	if err := p.TracerProvider.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown tracer provider: %w", err)
	}

	// Shutdown meter provider
	if err := p.MeterProvider.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown meter provider: %w", err)
	}

	return nil
}
