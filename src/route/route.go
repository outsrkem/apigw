package route

import (
	"apigw/src/cfgtypts"
	"apigw/src/service/proxy"
	"apigw/src/service/userauth"
	"apigw/src/slog"
	"github.com/cloudwego/hertz/pkg/app/server"
	"strings"
)

func Middleware(h *server.Hertz) {
	h.Use(RequestId())
	h.Use(RequestRecorder())
}

// LocalRouter 认证相关路由
func LocalRouter(h *server.Hertz, auth *cfgtypts.Auth) {
	klog := slog.FromContext(nil)
	host := auth.Backend.Host
	path := "/internal/v1/uias/user/signin"
	klog.Infof("Auth backend: %s", host)

	h.HEAD("", HelloWorld())
	h.GET("/", HelloWorld())
	h.POST("/api/uias/v1/user/signin", userauth.UiasSignin(host, path))
	h.POST("/api/uias/v1/user/logout", userauth.UiasLogout())

	h.POST("/api/uias/v1/uias/retpwd/gcode", proxy.NoAuthProxy(host, "/v1/uias/retpwd/gcode"))
	h.POST("/api/uias/v1/uias/retpwd/spwd", proxy.NoAuthProxy(host, "/v1/uias/retpwd/spwd"))
}

// 设置后端目标
func backend(backendHost, backendUrl string) string {
	host := strings.TrimSuffix(backendHost, "/")
	url := strings.TrimPrefix(backendUrl, "/")
	if url == "" {
		return host
	}
	return host + "/" + url
}

// ProxyRouter 外部注册接口的路由
func ProxyRouter(h *server.Hertz, cfgProxy *[]cfgtypts.Proxy) {
	klog := slog.FromContext(nil)
	for _, apigw := range *cfgProxy {
		for _, v := range apigw.Server {
			host := v.Location.Backend.Host // 后端服务域名
			tUrl := v.Location.Backend.Url  // 目标url
			rUrl := v.Location.Path         // 请求url
			auth := v.Location.Auth
			target := backend(host, tUrl)
			klog.Infof("%s: %s -> %s", apigw.Name, rUrl, target)
			method := NewProxyMethod(h, target, rUrl, auth, v.Location)

			switch v.Location.Method {
			case "Any":
				method.Any()
			case "Head":
				method.Head()
			case "Get":
				method.Get()
			case "Post":
				method.Post()
			case "Delete":
				method.Delete()
			case "Patch":
				method.Patch()
			default:
				method.Head()
			}
		}
	}
}
