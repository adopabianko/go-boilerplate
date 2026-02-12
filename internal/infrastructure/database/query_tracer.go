package database

import (
	"context"
	"time"

	"go-boilerplate/pkg/logger"

	"github.com/jackc/pgx/v5"
	"go.elastic.co/apm/v2"
	"go.uber.org/zap"
)

// QueryTracer implements pgx.QueryTracer to log all SQL queries and create APM spans
type QueryTracer struct{}

type queryContextKey struct{}

type queryContextValue struct {
	startTime time.Time
	sql       string
	args      []any
	span      *apm.Span
}

// TraceQueryStart is called at the beginning of Query, QueryRow, and Exec calls
func (t *QueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	// Start APM span for the SQL query
	span, ctx := apm.StartSpan(ctx, "SQL", "db.postgresql.query")
	if span != nil && !span.Dropped() {
		span.Context.SetDatabase(apm.DatabaseSpanContext{
			Type:      "sql",
			Statement: data.SQL,
		})
	}

	return context.WithValue(ctx, queryContextKey{}, &queryContextValue{
		startTime: time.Now(),
		sql:       data.SQL,
		args:      data.Args,
		span:      span,
	})
}

// TraceQueryEnd is called at the end of Query, QueryRow, and Exec calls
func (t *QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	queryCtx, ok := ctx.Value(queryContextKey{}).(*queryContextValue)
	if !ok {
		return
	}

	// End APM span
	if queryCtx.span != nil {
		if data.Err != nil {
			e := apm.DefaultTracer().NewError(data.Err)
			e.SetSpan(queryCtx.span)
			e.Send()
		}
		queryCtx.span.End()
	}

	duration := time.Since(queryCtx.startTime)

	fields := []zap.Field{
		zap.String("type", "sql_query"),
		zap.String("sql", queryCtx.sql),
		zap.Any("args", queryCtx.args),
		zap.Duration("duration", duration),
		zap.Int64("duration_ms", duration.Milliseconds()),
		zap.String("command_tag", data.CommandTag.String()),
	}

	// Add trace context fields for log-trace correlation
	fields = append(fields, logger.TraceFields(ctx)...)

	if data.Err != nil {
		fields = append(fields, zap.Error(data.Err))
		logger.Error("SQL query failed", fields...)
	} else {
		logger.Info("SQL query executed", fields...)
	}
}

