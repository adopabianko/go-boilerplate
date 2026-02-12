package tracer

import (
	"context"

	"go.elastic.co/apm/v2"
)

// StartSpan starts a new APM span as a child of the current transaction/span in context.
// Returns a new context containing the span and the span itself.
// The caller MUST call span.End() when the operation completes (use defer).
// If there is no active transaction in the context, the span will be a no-op.
func StartSpan(ctx context.Context, name, spanType string) (context.Context, *apm.Span) {
	span, ctx := apm.StartSpan(ctx, name, spanType)
	return ctx, span
}

// SpanFromContext returns the current span from context, or nil.
func SpanFromContext(ctx context.Context) *apm.Span {
	return apm.SpanFromContext(ctx)
}

// TransactionFromContext returns the current transaction from context, or nil.
func TransactionFromContext(ctx context.Context) *apm.Transaction {
	return apm.TransactionFromContext(ctx)
}

// TraceContext extracts trace ID, span ID, and transaction ID strings from context.
// Returns empty strings if no active transaction is present.
func TraceContext(ctx context.Context) (traceID, spanID, transactionID string) {
	tx := apm.TransactionFromContext(ctx)
	if tx == nil {
		return "", "", ""
	}
	traceCtx := tx.TraceContext()
	traceID = traceCtx.Trace.String()
	transactionID = traceCtx.Span.String()

	span := apm.SpanFromContext(ctx)
	if span != nil {
		spanCtx := span.TraceContext()
		spanID = spanCtx.Span.String()
	}

	return traceID, spanID, transactionID
}
