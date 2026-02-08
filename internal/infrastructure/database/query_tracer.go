package database

import (
	"context"
	"time"

	"go-boilerplate/pkg/logger"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// QueryTracer implements pgx.QueryTracer to log all SQL queries
type QueryTracer struct{}

type queryContextKey struct{}

type queryContextValue struct {
	startTime time.Time
	sql       string
	args      []any
}

// TraceQueryStart is called at the beginning of Query, QueryRow, and Exec calls
func (t *QueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, queryContextKey{}, &queryContextValue{
		startTime: time.Now(),
		sql:       data.SQL,
		args:      data.Args,
	})
}

// TraceQueryEnd is called at the end of Query, QueryRow, and Exec calls
func (t *QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	queryCtx, ok := ctx.Value(queryContextKey{}).(*queryContextValue)
	if !ok {
		return
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

	if data.Err != nil {
		fields = append(fields, zap.Error(data.Err))
		logger.Error("SQL query failed", fields...)
	} else {
		logger.Info("SQL query executed", fields...)
	}
}
