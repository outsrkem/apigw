package answer

const (
	EcodeOkay                  = "APIGW.00000"
	EcodeNotLogIn              = "APIGW.01001" // 未登录token
	EcodeSaveSessionError      = "APIGW.05101" // 保存session错误
	EcodeReadUpstreamDataError = "APIGW.05102" // 读取上游返回体错误
	EcodeBackEndServiceError   = "APIGW.05001" // 后端错误
	EcodeSendingRequest        = "APIGW.05002" // 发送请求错误
)
