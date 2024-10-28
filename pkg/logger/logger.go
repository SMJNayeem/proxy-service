// package logger

// import (
// 	"context"

// 	"go.uber.org/zap"
// 	"go.uber.org/zap/zapcore"
// )

// type Logger struct {
// 	*zap.Logger
// }

// func NewLogger(cfg *config.LogConfig) (*Logger, error) {
// 	config := zap.Config{
// 		Level:            zap.NewAtomicLevelAt(getLogLevel(cfg.Level)),
// 		Development:      false,
// 		Encoding:         "json",
// 		EncoderConfig:    zap.NewProductionEncoderConfig(),
// 		OutputPaths:      []string{"stdout"},
// 		ErrorOutputPaths: []string{"stderr"},
// 	}

// 	logger, err := config.Build()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Logger{logger}, nil
// }

// func (l *Logger) WithContext(ctx context.Context) *Logger {
// 	if requestID, ok := ctx.Value("request_id").(string); ok {
// 		return &Logger{l.With(zap.String("request_id", requestID))}
// 	}
// 	return l
// }

// func getLogLevel(level string) zapcore.Level {
// 	switch level {
// 	case "debug":
// 		return zapcore.DebugLevel
// 	case "info":
// 		return zapcore.InfoLevel
// 	case "warn":
// 		return zapcore.WarnLevel
// 	case "error":
// 		return zapcore.ErrorLevel
// 	default:
// 		return zapcore.InfoLevel
// 	}
// }

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultLogger *Logger

type Logger struct {
	*zap.Logger
}

func NewLogger() *Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &Logger{
		Logger: logger,
	}
}

func Info(msg string, fields ...interface{}) {
	defaultLogger.Info(msg, fields...)
}

func Error(msg string, fields ...interface{}) {
	sugar := defaultLogger.Logger.Sugar()
	sugar.Errorw(msg, fields...)
}

func Fatal(msg string, fields ...interface{}) {
	defaultLogger.Fatal(msg, fields...)
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.Logger.Sugar().Infow(msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.Logger.Sugar().Errorw(msg, args...)
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Logger.Sugar().Fatalw(msg, args...)
}

func Sync() error {
	return defaultLogger.Logger.Sync()
}
