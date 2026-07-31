package slog

import (
	"apigw/src/cfgtypes"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

// ctxKey private custom context key type to avoid cross-package key naming collision
type ctxKey struct{}

// requestIDKey global unexported context key for carrying X-Request-ID trace identifier
var requestIDKey = ctxKey{}

// Logger encapsulated log wrapper, isolates raw logrus global instance to avoid external contamination
type Logger struct {
	root *logrus.Logger
}

// globalLogger singleton global root logger instance of slog package
var globalLogger *Logger

// GetGlobal returns global root logger, used for initialization, cron tasks & non-request scenarios
func GetGlobal() *Logger {
	return globalLogger
}

// logLevel convert string level config to logrus.Level enumeration
func logLevel(lev string) logrus.Level {
	lowerLev := strings.ToLower(lev)
	switch lowerLev {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}

// MyFormatter custom log text formatter, implements logrus.Formatter interface
type MyFormatter struct{}

// Format implement logrus.Formatter, generate formatted log byte slice
func (MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	// UTC+8 standard datetime format
	asctime := entry.Time.Format("2006-01-02 15:04:05 +0800")
	level := entry.Level.String()

	caller := "?:?"
	if entry.Caller != nil {
		caller = fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
	}

	xRequestId := "-"
	if val, exists := entry.Data["xRequestId"]; exists {
		if id, ok := val.(string); ok {
			xRequestId = id
		}
	}

	_, err := fmt.Fprintf(b, "[%s] [%s] [%-7s] [%s] %s\n",
		asctime,
		xRequestId,
		level,
		caller,
		entry.Message,
	)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// InitLogger initialize global logger on program startup, configure output, rotation & log level
func InitLogger(cfg *cfgtypes.Log) {
	l := logrus.New()
	l.SetFormatter(&MyFormatter{})
	l.SetReportCaller(true)
	l.SetLevel(logLevel(cfg.Level))

	var writers []io.Writer
	if cfg.Output.File.Name != "" {
		l.Infof("log output file: %s", cfg.Output.File.Name)
		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.Output.File.Name,
			MaxSize:    cfg.Output.File.MaxSize,
			MaxBackups: cfg.Output.File.MaxBackups,
			MaxAge:     cfg.Output.File.MaxAge,
			Compress:   cfg.Output.File.Compress,
			LocalTime:  true,
		}
		writers = append(writers, lumberJackLogger)
		// print logs to stdout simultaneously when Stdout config equals "-"
		if cfg.Output.Stdout == "-" {
			writers = append(writers, os.Stdout)
		}
	} else {
		writers = append(writers, os.Stdout)
	}
	l.SetOutput(io.MultiWriter(writers...))

	globalLogger = &Logger{root: l}
}

// TraceIDMiddleware Hertz global middleware, extract X-Request-ID from hertz request context
func TraceIDMiddleware() app.HandlerFunc {
	return func(c context.Context, hzCtx *app.RequestContext) {
		reqID := "-"
		if val, exists := hzCtx.Keys["xRequestId"]; exists {
			if id, ok := val.(string); ok {
				reqID = id
			}
		}
		newCtx := context.WithValue(c, requestIDKey, reqID)
		hzCtx.Next(newCtx)
	}
}

// FromCtx get log entry bound with request trace ID from standard context.Context
func FromCtx(ctx context.Context) *logrus.Entry {
	reqID := "-"
	if v := ctx.Value(requestIDKey); v != nil {
		if id, ok := v.(string); ok {
			reqID = id
		}
	}
	return globalLogger.root.WithField("xRequestId", reqID)
}

// Debug raw debug log
func (l *Logger) Debug(args ...interface{}) {
	l.root.Debug(args...)
}

// Debugf formatted debug log
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.root.Debugf(format, args...)
}

// Info raw info log
func (l *Logger) Info(args ...interface{}) {
	l.root.Info(args...)
}

// Infof formatted info log
func (l *Logger) Infof(format string, args ...interface{}) {
	l.root.Infof(format, args...)
}

// Warn raw warning log
func (l *Logger) Warn(args ...interface{}) {
	l.root.Warn(args...)
}

// Warnf formatted warning log
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.root.Warnf(format, args...)
}

// Error raw error log
func (l *Logger) Error(args ...interface{}) {
	l.root.Error(args...)
}

// Errorf formatted error log
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.root.Errorf(format, args...)
}

// Fatal raw fatal log, exit program after print
func (l *Logger) Fatal(args ...interface{}) {
	l.root.Fatal(args...)
}

// Fatalf formatted fatal log, exit program after print
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.root.Fatalf(format, args...)
}

// RawLogger get raw underlying *logrus.Logger instance for advanced custom extension
func RawLogger() *logrus.Logger {
	if globalLogger == nil {
		globalLogger = &Logger{root: logrus.New()}
	}
	return globalLogger.root
}
