package mysql

import (
	"apigw/src/cfgtypes"
	"apigw/src/slog"
	"context"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
	"strings"
	"time"
)

// gormLogLev 对齐 logrus 与 GORM 日志等级
func gormLogLev(lev string) logger.LogLevel {
	lowerLev := strings.ToLower(lev)
	switch lowerLev {
	case "debug":
		return logger.Info
	case "info":
		return logger.Info
	case "warn":
		return logger.Warn
	case "error", "fatal":
		return logger.Error
	default:
		return logger.Info
	}
}

// gormLogAdapter 适配器：把GORM日志转发到项目自定义slog
type gormLogAdapter struct {
	inner *logrus.Logger
	cfg   logger.Config
}

func NewGormLogger(cfg *cfgtypes.Apigw) logger.Interface {
	// 拿到项目全局唯一日志实例，不再使用StandardLogger
	logIns := slog.RawLogger()
	return &gormLogAdapter{
		inner: logIns,
		cfg: logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormLogLev(cfg.Log.Level),
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	}
}

func (g *gormLogAdapter) LogLevel() logger.LogLevel {
	return g.cfg.LogLevel
}

func (g *gormLogAdapter) LogMode(level logger.LogLevel) logger.Interface {
	newCfg := g.cfg
	newCfg.LogLevel = level
	return &gormLogAdapter{inner: g.inner, cfg: newCfg}
}

func (g *gormLogAdapter) Info(ctx context.Context, format string, args ...interface{}) {
	g.inner.WithContext(ctx).Infof(format, args...)
}

func (g *gormLogAdapter) Warn(ctx context.Context, format string, args ...interface{}) {
	g.inner.WithContext(ctx).Warnf(format, args...)
}

func (g *gormLogAdapter) Error(ctx context.Context, format string, args ...interface{}) {
	g.inner.WithContext(ctx).Errorf(format, args...)
}

func (g *gormLogAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rows int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	costMs := float64(elapsed.Nanoseconds()) / 1e6
	entry := g.inner.WithContext(ctx).WithField("_skip", 1)

	// GORM原生自动打印调用模型文件行号 + SQL耗时语句
	switch {
	case err != nil:
		entry.Errorf("[%.3fms] [rows:%d] %s , err: %v", costMs, rows, sql, err)
	case elapsed > g.cfg.SlowThreshold:
		entry.Warnf("[%.3fms] [rows:%d] %s", costMs, rows, sql)
	default:
		entry.Debugf("[%.3fms] [rows:%d] %s", costMs, rows, sql)
	}
}
