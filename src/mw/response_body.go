package mw

import (
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// 定制统一的返回体内容
func NewResMessage(code int, msg interface{}) map[string]interface{} {
	responseBody := make(map[string]interface{})

	if msg != nil {
		metadata := make(map[string]interface{})
		metadata["message"] = msg
		metadata["time"] = time.Now().UnixNano() / 1e6
		responseBody["metadata"] = metadata
	} else {
		metadata := make(map[string]interface{})
		metadata["message"] = "Successfully."
		metadata["time"] = time.Now().UnixNano() / 1e6
		responseBody["metadata"] = metadata
	}
	return responseBody
}

func ResponseBody(ctx *app.RequestContext, status int, msg interface{}) {
	ctx.JSON(status, NewResMessage(status, msg))
}
