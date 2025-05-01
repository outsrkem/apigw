package route

import (
	"apigw/src/slog"
	"context"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
)

func RequestId() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromContext(c)
		xRequestId := string(c.GetHeader("X-Request-Id"))
		if xRequestId == "" {
			xRequestId = strings.ReplaceAll(uuid.New().String(), "-", "")
			c.Response.Header.Set("X-Request-Id", xRequestId)
			klog.Warnf("request id is empty, Set a new request id: %s", xRequestId)
		}
		c.Set("xRequestId", xRequestId)
		c.Next(ctx)
	}
}

func RequestRecorder() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromContext(c)
		start := time.Now()
		c.Next(ctx)
		stop := time.Now()
		latency := stop.Sub(start)
		klog.Infof("|%14s | %d |%7s %s",
			latency, c.Response.StatusCode(), string(c.Request.Method()), c.Request.URI().String())
	}
}
