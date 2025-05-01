package route

import (
	"apigw/src/cfgtypts"
	"apigw/src/service/proxy"
	"apigw/src/service/userauth"
	"apigw/src/slog"
	"context"
	"net/url"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

func Middleware(h *server.Hertz) {
	h.Use(RequestId())
	h.Use(RequestRecorder())
}

func HelloWorld() func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) {
		klog := slog.FromContext(ctx)
		klog.Info("hello world.")
		ctx.JSON(200, utils.H{"message": "hello world"})
	}
}

func LocalRouter(h *server.Hertz, auth *cfgtypts.Auth) {
	klog := slog.FromContext(nil)
	host := auth.Backend.Host
	path := "/internal/v1/uias/user/signin"
	klog.Infof("auth : /uias/v1/user/signin -> %s%s", host, path)

	h.HEAD("", HelloWorld())
	h.GET("/", HelloWorld())
	h.POST("/api/uias/v1/user/signin", userauth.UiasSignin(host, path))
	h.POST("/api/uias/v1/user/logout", userauth.UiasLogout())
}

func ProxyRouter(h *server.Hertz, cfgProxy *[]cfgtypts.Proxy) {
	klog := slog.FromContext(nil)
	for _, apigw := range *cfgProxy {
		for _, v := range apigw.Server {
			host := v.Location.Backend.Host // 后端服务域名
			tUrl := v.Location.Backend.Url  // 目标url
			rUrl := v.Location.Path         // 请求url
			klog.Infof("%s: %s -> %s", apigw.Name, v.Location.Path, host+tUrl)

			target, err := url.Parse(host)
			if err != nil {
				klog.Errorf("parse url fail: %v", err)
				panic(err)
			}

			klog.Infof("%s %s", rUrl, target)
			h.HEAD(rUrl+"/*path", proxy.ProxyUrl(host, rUrl))
			h.GET(rUrl+"/*path", proxy.ProxyUrl(host, rUrl))
			h.POST(rUrl+"/*path", proxy.ProxyUrl(host, rUrl))
			h.DELETE(rUrl+"/*path", proxy.ProxyUrl(host, rUrl))
			h.PATCH(rUrl+"/*path", proxy.ProxyUrl(host, rUrl))
		}
	}
}
