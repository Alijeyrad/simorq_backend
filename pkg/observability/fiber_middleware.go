package observability

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "github.com/Alijeyrad/simorq_backend/pkg/observability"
)

// FiberMiddleware returns a Fiber middleware that instruments HTTP requests with OpenTelemetry
func FiberMiddleware(serviceName string) fiber.Handler {
	tracer := otel.Tracer(tracerName)
	meter := otel.Meter(tracerName)

	// Create metrics
	requestCounter, _ := meter.Int64Counter(
		"http_server_request_count",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)

	requestDuration, _ := meter.Float64Histogram(
		"http_server_request_duration_ms",
		metric.WithDescription("HTTP request duration in milliseconds"),
		metric.WithUnit("ms"),
	)

	return func(c fiber.Ctx) error {
		// Extract trace context from incoming request headers
		ctx := otel.GetTextMapPropagator().Extract(
			c.Context(),
			propagation.HeaderCarrier(c.GetReqHeaders()),
		)

		// Start a new span
		spanName := c.Method() + " " + c.Route().Path
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.route", c.Route().Path),
				attribute.String("http.url", string(c.Request().URI().FullURI())),
				attribute.String("http.scheme", c.Protocol()),
				attribute.String("net.host.name", c.Hostname()),
				attribute.String("http.user_agent", c.Get("User-Agent")),
				attribute.String("http.client_ip", c.IP()),
			),
		)
		defer span.End()

		// Store context in Fiber context for downstream use
		c.SetContext(ctx)

		// Add trace ID to response headers for client correlation
		if span.SpanContext().HasTraceID() {
			c.Set("X-Trace-Id", span.SpanContext().TraceID().String())
		}

		// Record start time
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds() * 1000 // Convert to milliseconds

		// Set span status based on response
		statusCode := c.Response().StatusCode()
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Float64("http.duration_ms", duration),
		)

		// Record metrics
		attrs := metric.WithAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.route", c.Route().Path),
			attribute.Int("http.status_code", statusCode),
		)

		requestCounter.Add(ctx, 1, attrs)
		requestDuration.Record(ctx, duration, attrs)

		// Mark span as error if status code >= 500
		if statusCode >= 500 {
			span.SetStatus(codes.Error, "HTTP "+string(rune(statusCode)))
			if err != nil {
				span.RecordError(err)
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}
