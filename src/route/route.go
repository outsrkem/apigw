package route

import (
	"apigw/src/cfgtypes"
	"apigw/src/core/userauth"
	"apigw/src/handler"
	"apigw/src/slog"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Middleware(h *server.Hertz) {
	h.Use(RequestId())
	h.Use(slog.TraceIDMiddleware())
	h.Use(RequestRecorder())
}

// AuthRouter registers routes related to authentication
func AuthRouter(h *server.Hertz, auth *cfgtypes.Auth) {
	var klog = slog.GetGlobal()
	host := auth.Backend.Host
	const logInPath = "/internal/v1/uias/user/signin"

	klog.Infof("Auth backend address: %s", host)

	h.POST("/api/uias/v1/user/signin", userauth.UiasSignin(host, logInPath))
	h.POST("/api/uias/v1/user/logout", userauth.UiasLogout())
}

// ========== API接口管理 处理器占位赋值 ==========
var (
	ApiListHandler   = HelloWorld()
	ApiDetailHandler = HelloWorld()
	ApiCreateHandler = HelloWorld()
	ApiUpdateHandler = HelloWorld()
	ApiStatusHandler = HelloWorld()
	ApiDeleteHandler = HelloWorld()

	// 分组管理
	GroupListHandler   = HelloWorld()
	GroupDetailHandler = HelloWorld()
	GroupCreateHandler = HelloWorld()
	GroupUpdateHandler = HelloWorld()
	GroupDeleteHandler = HelloWorld()

	// 域名绑定管理
	DomainListHandler   = HelloWorld()
	DomainDetailHandler = HelloWorld()
	DomainCreateHandler = HelloWorld()
	DomainUpdateHandler = HelloWorld()
	DomainDeleteHandler = HelloWorld()

	// 负载通道管理
	ChannelListHandler   = HelloWorld()
	ChannelDetailHandler = HelloWorld()
	ChannelCreateHandler = HelloWorld()
	ChannelUpdateHandler = HelloWorld()
	ChannelStatusHandler = HelloWorld()
	ChannelDeleteHandler = HelloWorld()
)

func ApigwRoute(h *server.Hertz) {
	h.HEAD("", HelloWorld())
	h.GET("/", HelloWorld())
	// -------------------------- 1.API接口路由管理（对应数据表 api_interface） --------------------------

	h.GET("/gateway/admin/api", ApiListHandler)                // ApiListHandler 查询接口路由分页列表、多条件筛选
	h.GET("/gateway/admin/api/:id", ApiDetailHandler)          // ApiDetailHandler 根据ID查询单条接口路由详情
	h.POST("/gateway/admin/api", ApiCreateHandler)             // ApiCreateHandler 新建一条API路由配置
	h.PUT("/gateway/admin/api/:id", ApiUpdateHandler)          // ApiUpdateHandler 根据ID全量覆盖修改整条路由配置
	h.PATCH("/gateway/admin/api/:id/status", ApiStatusHandler) // ApiStatusHandler 局部更新接口上下线状态，仅修改status/publishing_status字段
	h.DELETE("/gateway/admin/api/:id", ApiDeleteHandler)       // ApiDeleteHandler 根据ID执行接口逻辑/物理删除

	// -------------------------- 2.业务分组管理（对应分组数据表） --------------------------
	h.GET("/gateway/admin/group", GroupListHandler)          // GroupListHandler 查询业务分组分页列表、筛选查询
	h.GET("/gateway/admin/group/:id", GroupDetailHandler)    // GroupDetailHandler 根据分组ID查询分组详情
	h.POST("/gateway/admin/group", GroupCreateHandler)       // GroupCreateHandler 新建业务分组
	h.PUT("/gateway/admin/group/:id", GroupUpdateHandler)    // GroupUpdateHandler 全量修改分组基础信息
	h.DELETE("/gateway/admin/group/:id", GroupDeleteHandler) // GroupDeleteHandler 删除指定业务分组

	// -------------------------- 3.分组域名绑定管理（域名与分组绑定配置） --------------------------
	h.GET("/gateway/admin/domain", DomainListHandler)          // DomainListHandler 查询分组域名绑定列表
	h.GET("/gateway/admin/domain/:id", DomainDetailHandler)    // DomainDetailHandler 根据ID查询单条域名绑定详情
	h.POST("/gateway/admin/domain", DomainCreateHandler)       // DomainCreateHandler 新建分组-域名绑定关系
	h.PUT("/gateway/admin/domain/:id", DomainUpdateHandler)    // DomainUpdateHandler 全量修改域名绑定配置
	h.DELETE("/gateway/admin/domain/:id", DomainDeleteHandler) // DomainDeleteHandler 删除域名绑定记录

	// -------------------------- 4.负载通道/上游服务集群管理 --------------------------
	h.GET("/gateway/admin/channel", ChannelListHandler)                // ChannelListHandler 查询上游负载通道分页列表
	h.GET("/gateway/admin/channel/:id", ChannelDetailHandler)          // ChannelDetailHandler 根据ID查询单条负载通道详情
	h.POST("/gateway/admin/channel", ChannelCreateHandler)             // ChannelCreateHandler 创建新的上游负载集群通道
	h.PUT("/gateway/admin/channel/:id", ChannelUpdateHandler)          // ChannelUpdateHandler 全量修改通道节点、权重、协议等配置
	h.PATCH("/gateway/admin/channel/:id/status", ChannelStatusHandler) // ChannelStatusHandler 局部切换通道启用/禁用状态、熔断开关
	h.DELETE("/gateway/admin/channel/:id", ChannelDeleteHandler)       // ChannelDeleteHandler 删除无用负载通道

}

func ProxyRoute(h *server.Hertz, auth *cfgtypes.Auth) {
	h.Any("/*path", handler.ProxyDispatch())
}
