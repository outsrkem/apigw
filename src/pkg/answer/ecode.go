package answer

const (
	// ========== 00 成功 ==========
	EcodeOK       = "APIGW.00000" // 请求正常成功
	EcodeOkay     = "APIGW.00000" // 请求正常成功
	EcodeError    = "APIGW.11111" // 通用错误
	EcodeNotFound = "APIGW.11112" // 资源不存咋

	// ========== 01 登录&Token身份认证 ==========
	EcodeNotLogIn         = "APIGW.01001" // 未登录、Token不存在
	EcodeTokenExpire      = "APIGW.01002" // Token过期
	EcodeTokenInvalid     = "APIGW.01003" // Token格式非法、篡改
	EcodeRefreshTokenFail = "APIGW.01004" // 刷新Token失败

	// ========== 02 权限访问控制 ==========
	EcodeNoPermission = "APIGW.02001" // 已登录，但无接口访问权限
	EcodeIpForbidden  = "APIGW.02002" // IP黑名单拦截
	EcodeAccessDeny   = "APIGW.02003" // 访问被策略强制拒绝

	// ========== 03 路由&接口匹配（重点：接口不存在放这里） ==========
	EcodeStatusNotFound = "APIGW.03001" // 接口不存在、路由未匹配（对应HTTP404）
	EcodeMethodNotAllow = "APIGW.03002" // 请求方法不允许（对应HTTP405）
	EcodeRouteConfigErr = "APIGW.03003" // 路由配置解析错误

	// ========== 04 参数、配置、鉴权规则校验 ==========
	EcodeParamInvalid      = "APIGW.04010" // 接口鉴权配置类型错误
	EcodeParamMissing      = "APIGW.04011" // 必传参数缺失
	EcodeParamFormatErr    = "APIGW.04012" // 参数格式错误（JSON解析失败等）
	EcodeInvalidRequestErr = "APIGW.04013" // 参数无效
	EcodeApiGroupDuplicate = "APIGW.04014" // 参数无效
	EcodeInvalidApiId      = "APIGW.04015" // API ID 无效
	EcodeSignVerifyFail    = "APIGW.04020" // 请求签名校验失败
	EcodeGroupInUse        = "APIGW.04021" // API分组被引用占用，禁止删除

	// ========== 05 上游转发、后端服务异常 ==========
	EcodeBackEndServiceError   = "APIGW.05001" // 后端业务内部异常
	EcodeSendingRequest        = "APIGW.05002" // 向上游发送网络请求失败
	EcodeUpstreamUnavailable   = "APIGW.05003" // 上游服务无可用节点、下线熔断（HTTP503）
	EcodeGatewayTimeout        = "APIGW.05004" // 上游请求超时（HTTP504）
	EcodeReadUpstreamDataError = "APIGW.05102" // 读取上游响应体解析失败

	// ========== 06 网关内部存储异常 ==========
	EcodeSaveSessionError = "APIGW.05101" // 保存session错误
	EcodeRedisErr         = "APIGW.06001" // Redis读写异常
	EcodeMysqlErr         = "APIGW.06002" // Mysql数据库查询异常
	EcodeCacheLoadFail    = "APIGW.06003" // 路由/通道缓存加载失败

	// ========== 07 限流、熔断、风控 ==========
	EcodeRateLimit    = "APIGW.07001" // 请求频率限流拦截
	EcodeCircuitBreak = "APIGW.07002" // 上游熔断拦截
)
