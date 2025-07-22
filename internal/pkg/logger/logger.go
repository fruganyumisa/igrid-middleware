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
	z.logger.Debug(msg, convertToZapFields(fields...)...)
}

func (z ZapLogger) Info(msg string, fields ...interface{}) {
	z.logger.Info(msg, convertToZapFields(fields...)...)
}

func (z ZapLogger) Warn(msg string, fields ...interface{}) {
	z.logger.Warn(msg, convertToZapFields(fields...)...)
}

func (z ZapLogger) Error(msg string, fields ...interface{}) {
	z.logger.Error(msg, convertToZapFields(fields...)...)
}

func (z ZapLogger) Fatal(msg string, fields ...interface{}) {
	z.logger.Fatal(msg, convertToZapFields(fields...)...)
}

// convertToZapFields converts interface{} fields to zap.Field
func convertToZapFields(fields ...interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields)/2)

	for i := 0; i < len(fields)-1; i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		value := fields[i+1]

		switch v := value.(type) {
		case string:
			zapFields = append(zapFields, zap.String(key, v))
		case int:
			zapFields = append(zapFields, zap.Int(key, v))
		case int64:
			zapFields = append(zapFields, zap.Int64(key, v))
		case float64:
			zapFields = append(zapFields, zap.Float64(key, v))
		case bool:
			zapFields = append(zapFields, zap.Bool(key, v))
		case error:
			zapFields = append(zapFields, zap.Error(v))
		default:
			zapFields = append(zapFields, zap.Any(key, v))
		}
	}

	return zapFields
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
