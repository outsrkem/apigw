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

	h.POST("/api/uias/v1/uias/retpwd/gcode", proxy.NoAuthProxy(host, "/v1/uias/retpwd/gcode"))
	h.POST("/api/uias/v1/uias/retpwd/spwd", proxy.NoAuthProxy(host, "/v1/uias/retpwd/spwd"))
}

func HeadMethod(h *server.Hertz, host, rUrl, auth string) {
	switch auth {
	case "uias":
		h.HEAD(rUrl+"/*path", proxy.UiasAuthProxy(host, rUrl))
	case "off":
		h.HEAD(rUrl+"/*path", proxy.NoAuthProxy(host, rUrl))
	default:
		h.HEAD(rUrl+"/*path", proxy.UiasAuthProxy(host, rUrl))
	}
}

func GetMethod(h *server.Hertz, host, rUrl, auth string) {
	switch auth {
	case "uias":
		h.GET(rUrl+"/*path", proxy.UiasAuthProxy(host, rUrl))
	case "off":
		h.GET(rUrl+"/*path", proxy.NoAuthProxy(host, rUrl))
	default:
		h.GET(rUrl+"/*path", proxy.UiasAuthProxy(host, rUrl))
	}
}

func PostMethod(h *server.Hertz, host, rUrl string) {
	h.POST(rUrl+"/*path", proxy.ProxyUrl(host, rUrl))
}

func DeleteMethod(h *server.Hertz, host, rUrl string) {
	h.DELETE(rUrl+"/*path", proxy.ProxyUrl(host, rUrl))
}

func PatchMethod(h *server.Hertz, host, rUrl string) {
	h.PATCH(rUrl+"/*path", proxy.ProxyUrl(host, rUrl))
}

// ProxyRouter 代理路由
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
			switch v.Location.Method {
			case "Any":
				HeadMethod(h, host, rUrl, v.Location.Auth)
				GetMethod(h, host, rUrl, v.Location.Auth)
				PostMethod(h, host, rUrl)
				DeleteMethod(h, host, rUrl)
				PatchMethod(h, host, rUrl)
			case "Head":
				HeadMethod(h, host, rUrl, v.Location.Auth)
			case "Get":
				GetMethod(h, host, rUrl, v.Location.Auth)
			case "Post":
				PostMethod(h, host, rUrl)
			case "Delete":
				DeleteMethod(h, host, rUrl)
			case "Patch":
				PatchMethod(h, host, rUrl)
			default:
				HeadMethod(h, host, rUrl, v.Location.Auth)
			}
		}
	}
}
