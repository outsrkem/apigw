package route

import (
	"apigw/src/slog"
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

func HelloWorld() func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		klog := slog.FromCtx(ctx)
		klog.Info("hello world.")
		c.JSON(200, utils.H{"message": "hello world"})
	}
}
