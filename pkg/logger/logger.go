package logger

import (
	"fmt"
	"go-boilerplate/internal/config"
	"net"
	"os"

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
