package mw

import "time"

// 定制统一的返回体内容
func NewResMessage(code int, msg interface{}) map[string]interface{} {
	responseBody := make(map[string]interface{})

	if msg != nil {
		meta_info := make(map[string]interface{})
		meta_info["res_code"] = code
		meta_info["res_msg"] = msg
		meta_info["request_time"] = time.Now().UnixNano() / 1e6
		responseBody["meta_info"] = meta_info
	} else {
		meta_info := make(map[string]interface{})
		meta_info["res_code"] = code
		meta_info["request_time"] = time.Now().UnixNano() / 1e6
		responseBody["meta_info"] = meta_info
	}
	return responseBody
}
