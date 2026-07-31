package models

// OrmApiGroup API分组表
type OrmApiGroup struct {
	kid        int64  `gorm:"primaryKey;column:kid"`
	ID         string `gorm:"column:id"`          // uuid
	Name       string `gorm:"column:name"`        // 分组名称
	Remark     string `gorm:"column:remark"`      // 分组描述
	Status     int8   `gorm:"column:status"`      // 状态（0-禁用，1-启用）
	CreateTime int64  `gorm:"column:create_time"` // 创建时间戳
	UpdateTime int64  `gorm:"column:update_time"` // 更新时间戳
}

func (*OrmApiGroup) TableName() string {
	return "api_group"
}

// OrmApiDomain 域名配置表
type OrmApiDomain struct {
	kid        int64  `gorm:"primaryKey;column:kid"`
	ID         string `gorm:"column:id"`          // 分组id
	Name       string `gorm:"column:name"`        // 域名
	SslStatus  int8   `gorm:"column:ssl_status"`  // SSL状态（0-关闭，1-开启）
	SslCert    string `gorm:"column:ssl_cert"`    // SSL证书内容
	SslKey     string `gorm:"column:ssl_key"`     // SSL私钥内容
	Status     int8   `gorm:"column:status"`      // 状态（0-禁用，1-启用）
	CreateTime int64  `gorm:"column:create_time"` // 创建时间戳
	UpdateTime int64  `gorm:"column:update_time"` // 更新时间戳
}

func (*OrmApiDomain) TableName() string {
	return "api_domain"
}

// OrmLoadChannel 负载通道表
type OrmLoadChannel struct {
	kid        int64  `gorm:"primaryKey;column:kid"`
	ID         string `gorm:"column:id"`          // 负载通道id
	Name       string `gorm:"column:name"`        // 通道名称（如用户服务集群）
	Backend    string `gorm:"column:backend"`     // 后端地址列表（逗号分隔）
	Strategy   int8   `gorm:"column:strategy"`    // 负载策略（1-轮询，2-加权轮询，3-IP哈希）
	Timeout    int    `gorm:"column:timeout"`     // 后端超时时间（毫秒）
	HcInterval int    `gorm:"column:hcinterval"`  // 健康检查间隔（毫秒）
	Status     int8   `gorm:"column:status"`      // 状态（0-禁用，1-启用）
	CaCert     string `gorm:"column:ca_cert"`     // 后端CA证书ID
	CreateTime int64  `gorm:"column:create_time"` // 创建时间戳
	UpdateTime int64  `gorm:"column:update_time"` // 更新时间戳
}

func (*OrmLoadChannel) TableName() string {
	return "load_channel"
}

// OrmLcCa 后端负载的CA证书，https使用
type OrmLcCa struct {
	kid        int64  `gorm:"primaryKey;column:kid"`
	ID         string `gorm:"column:id"`          // 证书ID
	Name       string `gorm:"column:name"`        // 证书CN
	Cert       string `gorm:"column:cert"`        // 后端CA证书
	CreateTime int64  `gorm:"column:create_time"` // 创建时间戳
	UpdateTime int64  `gorm:"column:update_time"` // 更新时间戳
}

func (*OrmLcCa) TableName() string {
	return "lc_ca"
}

// OrmApiInterface API详情表
type OrmApiInterface struct {
	kid           int64  `gorm:"primaryKey;column:kid"`
	ID            string `gorm:"primaryKey;column:id"`  // API ID
	GroupID       string `gorm:"column:group_id"`       // 所属分组ID
	Protocol      string `gorm:"column:protocol"`       // 后端协议
	Method        string `gorm:"column:method"`         // HTTP方法
	ReqUri        string `gorm:"column:req_uri"`        // API路径
	BackendUri    string `gorm:"column:backend_uri"`    // 后端api
	Auth          string `gorm:"column:auth"`           // 认证类型
	Mode          string `gorm:"column:mode"`           // API的匹配方式prefix：前缀匹配,exact：精确匹配
	LcID          string `gorm:"column:lc_id"`          // 关联负载通道ID
	RateLimit     int    `gorm:"column:rate_limit"`     // 接口限流（QPS，0-不限流）
	Status        int8   `gorm:"column:status"`         // 状态（0-禁用，1-启用）
	PublishStatus int8   `gorm:"column:publish_status"` // 发布状态（0-未发布，1-测试中，2-已发布，3-已下线）
	CreateTime    int64  `gorm:"column:create_time"`    // 创建时间戳
	UpdateTime    int64  `gorm:"column:update_time"`    // 更新时间戳
}

func (*OrmApiInterface) TableName() string {
	return "api_interface"
}
