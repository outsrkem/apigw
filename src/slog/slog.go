package slog

import (
	"apigw/src/cfgtypts"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

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

type MyFormatter struct {
}

func (MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	asctime := entry.Time.Format(time.DateTime + " -0700")
	level := entry.Level.String()
	var caller string
	if entry.Caller != nil {
		caller = fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
		//caller = fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
	} else {
		caller = "?:?"
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

func InitLogger(cfg *cfgtypts.Log) {
	logrus.SetFormatter(&MyFormatter{})
	logrus.SetReportCaller(true)
	logrus.SetLevel(logLevel(cfg.Level))
	var writers []io.Writer
	if cfg.Output.File.Name != "" {
		logrus.Infof("Output to %s", cfg.Output.File.Name)
		lumberJackLogger := &lumberjack.Logger{
			Filename:   cfg.Output.File.Name,
			MaxSize:    cfg.Output.File.MaxSize,
			MaxBackups: cfg.Output.File.MaxBackups,
			MaxAge:     cfg.Output.File.MaxAge,
			Compress:   cfg.Output.File.Compress,
			LocalTime:  true,
		}
		writers = append(writers, lumberJackLogger)
		if cfg.Output.Stdout == "-" {
			writers = append(writers, os.Stdout)
		}
	} else {
		writers = append(writers, os.Stdout)
	}

	logrus.SetOutput(io.MultiWriter(writers...))
}

func FromContext(ctx *app.RequestContext) *logrus.Entry {
	if ctx != nil {
		xRequestId := "-"
		if val, exists := ctx.Keys["xRequestId"]; exists {
			if id, ok := val.(string); ok {
				xRequestId = id
			}
		}
		return logrus.WithFields(logrus.Fields{"xRequestId": xRequestId})
	}
	return logrus.WithFields(logrus.Fields{})
}
