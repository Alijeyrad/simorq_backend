package reqctx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// TraceInfo holds OpenTelemetry-compatible trace context.
type TraceInfo struct {
	// TraceID is a 32-character hex string (128-bit).
	// Identifies the entire distributed trace.
	TraceID string

	// SpanID is a 16-character hex string (64-bit).
	// Identifies this specific operation within the trace.
	SpanID string

	// ParentID is the parent span's ID, if this is a child span.
	ParentID string

	// Sampled indicates whether this trace should be recorded.
	Sampled bool
}

// WithTrace stores trace info in the context.
func WithTrace(ctx context.Context, trace *TraceInfo) context.Context {
	return context.WithValue(ctx, keyTrace, trace)
}

// TraceFromContext retrieves trace info from the context.
// Returns nil, false if not set.
func TraceFromContext(ctx context.Context) (*TraceInfo, bool) {
	v := ctx.Value(keyTrace)
	if v == nil {
		return nil, false
	}
	trace, ok := v.(*TraceInfo)
	return trace, ok
}

// MustTrace retrieves trace info from the context.
// Panics if not set.
func MustTrace(ctx context.Context) *TraceInfo {
	trace, ok := TraceFromContext(ctx)
	if !ok || trace == nil {
		panic("reqctx: TraceInfo not found in context")
	}
	return trace
}

// TraceIDFromContext returns the trace ID, or empty string if not set.
func TraceIDFromContext(ctx context.Context) string {
	trace, ok := TraceFromContext(ctx)
	if !ok || trace == nil {
		return ""
	}
	return trace.TraceID
}

// SpanIDFromContext returns the span ID, or empty string if not set.
func SpanIDFromContext(ctx context.Context) string {
	trace, ok := TraceFromContext(ctx)
	if !ok || trace == nil {
		return ""
	}
	return trace.SpanID
}

// GenerateTraceID creates a new random 128-bit trace ID as a 32-char hex string.
func GenerateTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateSpanID creates a new random 64-bit span ID as a 16-char hex string.
func GenerateSpanID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// NewTraceInfo creates a new TraceInfo with generated IDs.
func NewTraceInfo() *TraceInfo {
	return &TraceInfo{
		TraceID: GenerateTraceID(),
		SpanID:  GenerateSpanID(),
		Sampled: true,
	}
}

// NewChildSpan creates a child span from the current trace.
// If no trace exists in context, creates a new root trace.
func NewChildSpan(ctx context.Context) *TraceInfo {
	parent, ok := TraceFromContext(ctx)
	if !ok || parent == nil {
		return NewTraceInfo()
	}
	return &TraceInfo{
		TraceID:  parent.TraceID,
		SpanID:   GenerateSpanID(),
		ParentID: parent.SpanID,
		Sampled:  parent.Sampled,
	}
}
