package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}

type ZapLogger struct {
	logger *zap.Logger
}

func (z ZapLogger) Debug(msg string, fields ...interface{}) {
	//TODO implement me
	panic("implement me")
}

func (z ZapLogger) Info(msg string, fields ...interface{}) {
	//TODO implement me
	panic("implement me")
}

func (z ZapLogger) Warn(msg string, fields ...interface{}) {
	//TODO implement me
	panic("implement me")
}

func (z ZapLogger) Error(msg string, fields ...interface{}) {
	//TODO implement me
	panic("implement me")
}

func (z ZapLogger) Fatal(msg string, fields ...interface{}) {
	//TODO implement me
	panic("implement me")
}

func New(level string) Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(parseLevel(level))
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return &ZapLogger{logger: zapLogger}
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Implement Logger interface methods...
