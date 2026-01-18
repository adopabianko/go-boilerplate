package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
	)

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
