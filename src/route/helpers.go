package route

import (
	"apigw/src/cfgtypts"
	"apigw/src/service/proxy"
	"apigw/src/slog"
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"net/http"
	"path"
)

func HelloWorld() func(c context.Context, ctx *app.RequestContext) {
	return func(c context.Context, ctx *app.RequestContext) {
		klog := slog.FromContext(ctx)
		klog.Info("hello world.")
		ctx.JSON(200, utils.H{"message": "hello world"})
	}
}

const (
	AuthOff    = "off"    // 不开启认证
	AuthUias   = "uias"   // 使用uias认证
	HttpAny    = "Any"    // http any method
	ModeExact  = "Exact"  // 绝对匹配
	ModePrefix = "Prefix" // 前缀匹配
)

// RegisterMethod 注册一个HTTP代理路由。
// Parameters:
//   - H: Hertz服务器实例。
//   - method: HTTP方法(GET/POST等)。
//   - Host: 后端服务的主机地址。
//   - RUrl: 路由路径。
//   - Auth: 认证类型(AuthOff,AuthUias)。
func RegisterMethod(h *server.Hertz, method, host, rUrl, auth string, location cfgtypts.Location) {
	var handlerFunc app.HandlerFunc
	switch auth {
	case AuthOff:
		handlerFunc = proxy.NoAuthProxy(host, rUrl)
	case AuthUias:
		handlerFunc = proxy.UiasAuthProxy(host, rUrl)
	default:
		handlerFunc = proxy.UiasAuthProxy(host, rUrl)
	}

	var fullPath string // apigw的路由路径
	switch location.Mode {
	case ModeExact:
		fullPath = path.Join(rUrl)
	case ModePrefix:

		fullPath = path.Join(rUrl, "/*path")
	default:
		fullPath = path.Join(rUrl)
	}

	switch method {
	case http.MethodHead:
		h.HEAD(fullPath, handlerFunc)
	case http.MethodGet:
		h.GET(fullPath, handlerFunc)
	case http.MethodPost:
		h.POST(fullPath, handlerFunc)
	case http.MethodDelete:
		h.DELETE(fullPath, handlerFunc)
	case http.MethodPatch:
		h.PATCH(fullPath, handlerFunc)
	case http.MethodPut:
		h.PUT(fullPath, handlerFunc)
	case http.MethodOptions:
		h.OPTIONS(fullPath, handlerFunc)
	case HttpAny:
		h.Any(fullPath, handlerFunc)
	}
}

type Method interface {
	Any()
	Head()
	Get()
	Post()
	Delete()
	Patch()
	Put()
	Options()
}

type method struct {
	H        *server.Hertz // hertz server
	Host     string        // 后端服务域名
	RUrl     string        // 目标url
	Auth     string        // 请求url或请求的接口前缀
	Location cfgtypts.Location
}

func NewProxyMethod(h *server.Hertz, host, rUrl, auth string, location cfgtypts.Location) Method {
	return &method{
		H:        h,
		Host:     host,
		RUrl:     rUrl,
		Auth:     auth,
		Location: location,
	}
}

func (t *method) Any() {
	RegisterMethod(t.H, HttpAny, t.Host, t.RUrl, t.Auth, t.Location)
}

func (t *method) Head() {
	RegisterMethod(t.H, http.MethodHead, t.Host, t.RUrl, t.Auth, t.Location)
}

func (t *method) Get() {
	RegisterMethod(t.H, http.MethodGet, t.Host, t.RUrl, t.Auth, t.Location)
}

func (t *method) Post() {
	RegisterMethod(t.H, http.MethodPost, t.Host, t.RUrl, t.Auth, t.Location)
}

func (t *method) Delete() {
	RegisterMethod(t.H, http.MethodDelete, t.Host, t.RUrl, t.Auth, t.Location)
}

func (t *method) Patch() {
	RegisterMethod(t.H, http.MethodPatch, t.Host, t.RUrl, t.Auth, t.Location)
}

func (t *method) Put() {
	RegisterMethod(t.H, http.MethodPut, t.Host, t.RUrl, t.Auth, t.Location)
}

func (t *method) Options() {
	RegisterMethod(t.H, http.MethodOptions, t.Host, t.RUrl, t.Auth, t.Location)
}
