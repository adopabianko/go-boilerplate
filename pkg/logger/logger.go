package logger

import (
	"context"
	"fmt"
	"net"
	"os"

	"go-boilerplate/internal/config"

	"go.elastic.co/apm/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger(cfg *config.Config) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Console Output
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel)

	// Logstash UDP Output
	var cores []zapcore.Core
	cores = append(cores, consoleCore)

	if cfg != nil && cfg.Logstash.Host != "" {
		logstashAddr := net.JoinHostPort(cfg.Logstash.Host, cfg.Logstash.Port)
		udpConn, err := net.Dial("udp", logstashAddr)
		if err == nil {
			logstashCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(udpConn), zapcore.InfoLevel)
			cores = append(cores, logstashCore)
		} else {
			fmt.Printf("Failed to connect to Logstash: %v\n", err)
		}
	}

	core := zapcore.NewTee(cores...)

	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// TraceFields extracts APM trace context from ctx and returns zap fields.
// These fields enable log-trace correlation in Kibana.
func TraceFields(ctx context.Context) []zap.Field {
	tx := apm.TransactionFromContext(ctx)
	if tx == nil {
		return nil
	}

	traceCtx := tx.TraceContext()
	fields := []zap.Field{
		zap.String("trace.id", traceCtx.Trace.String()),
		zap.String("transaction.id", traceCtx.Span.String()),
	}

	span := apm.SpanFromContext(ctx)
	if span != nil {
		spanCtx := span.TraceContext()
		fields = append(fields, zap.String("span.id", spanCtx.Span.String()))
	}

	return fields
}

// InfoCtx logs an info message with trace context fields.
func InfoCtx(ctx context.Context, message string, fields ...zap.Field) {
	fields = append(fields, TraceFields(ctx)...)
	Log.Info(message, fields...)
}

// ErrorCtx logs an error message with trace context fields.
func ErrorCtx(ctx context.Context, message string, fields ...zap.Field) {
	fields = append(fields, TraceFields(ctx)...)
	Log.Error(message, fields...)
}

func Info(message string, fields ...zap.Field) {
	Log.Info(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	Log.Error(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	Log.Debug(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	Log.Fatal(message, fields...)
}

