package route

import (
	"apigw/src/cfgtypes"
	"apigw/src/core/userauth"
	"apigw/src/handler"
	"apigw/src/service/apigroup"
	"apigw/src/service/apinterface"
	"apigw/src/service/channel"
	"apigw/src/service/domain"
	"apigw/src/service/lcca"
	"apigw/src/slog"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Middleware(h *server.Hertz) {
	h.Use(RequestId())
	h.Use(slog.TraceIDMiddleware())
	h.Use(RequestRecorder())
}

func ProxyRoute(h *server.Hertz) {
	h.Any("/*path", handler.ProxyDispatch())
}

func AuthRouter(h *server.Hertz, auth *cfgtypes.Auth) {
	var klog = slog.GetGlobal()
	host := auth.Backend.Host
	klog.Infof("Auth backend address: %s", host)

	const logInPath = "/internal/v1/uias/user/signin"
	h.POST("/api/uias/v1/user/signin", userauth.UiasSignin(host, logInPath))
	h.POST("/api/uias/v1/user/logout", userauth.UiasLogout())
}

func ApigwRoute(h *server.Hertz) {
	h.HEAD("", HelloWorld())
	h.GET("/", HelloWorld())

	internal := h.Group("/internal/v1/admin")

	// -------------------------- API分组 --------------------------
	internal.GET("/group", apigroup.GroupList())             // √ 获取api分组列表
	internal.GET("/group/:id", apigroup.GetApiGroupDetail()) // √ 查询分组详情
	internal.POST("/group", apigroup.CreateApiGroup())       // √ 创建api分组
	internal.DELETE("/group/:id", apigroup.DeleteApiGroup()) // √ 删除分组
	internal.PATCH("/group/:id", apigroup.UpdateApiGroup())  // √ 修改分组

	// -------------------------- API接口路由 --------------------------
	internal.GET("/api", apinterface.ListApi())                    // √ 查询API接口
	internal.GET("/api/:id", apinterface.GetApiDetail())           // √ 获取API接口详情
	internal.POST("/:groupId/api", apinterface.CreateApi())        // √ 新建API接口
	internal.DELETE("/api/:id", apinterface.DeleteApi())           // √ 删除API接口
	internal.PATCH("/api/:id", apinterface.UpdateApi())            // √ 更新API接口
	internal.PUT("/api/:id/lifecycle", apinterface.ApiLifecycle()) // √ API接口上下线
	internal.PUT("/api/:id/status", apinterface.ApiStatus())       // √ API接口启用,禁用

	// -------------------------- 分组域名绑定（域名与分组绑定配置） --------------------------
	internal.GET("/:groupId/domain", domain.ListDomain())          // √ 查询绑定列表
	internal.POST("/:groupId/domain", domain.BindDomain())         // √ 绑定域名
	internal.DELETE("/:groupId/domain/:id", domain.UnbindDomain()) // √ 解绑域名

	// -------------------------- 负载通道证书 --------------------------
	internal.GET("/lcca", lcca.ListLcCa())          // √ 查询证书列表
	internal.GET("/lcca/:id", lcca.GetLcCaDetail()) // √ 查询证书详情
	internal.POST("/lcca", lcca.CreateLcCa())       // √ 创建新证书
	internal.DELETE("/lcca/:id", lcca.DeleteLcCa()) // √ 删除证书
	internal.PATCH("/lcca/:id", lcca.UpdateLcCa())  // √ 更新证书信息

	// -------------------------- 负载通道 --------------------------
	internal.POST("/channel", channel.CreateChannel())               // √ 新建负载通道
	internal.GET("/channel", channel.ListChannel())                  // √ 查询分页列表
	internal.GET("/channel/:id", channel.GetChannelDetail())         // √ 通道详情
	internal.PATCH("/channel/:id", channel.UpdateChannel())          // √ 修改信息
	internal.DELETE("/channel/:id", channel.DeleteChannel())         // √ 删除无用负载通道
	internal.POST("/channel/:id/status", channel.SetChannelStatus()) // √ 熔断开关-通道启用/禁用
}
