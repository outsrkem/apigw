package answer

import (
	"strconv"
	"time"
)

type Metadata struct {
	Message interface{} `json:"message"`
	Time    string      `json:"time"`
	Ecode   string      `json:"ecode"`
}

func NewResMessage(ecode string, msg interface{}, payload interface{}) map[string]interface{} {
	body := make(map[string]interface{})
	if msg == "" {
		msg = "Successfully."
	}
	metadata := Metadata{
		Message: msg,
		Time:    strconv.FormatInt(time.Now().UnixNano()/1e6, 10),
		Ecode:   ecode,
	}
	body["metadata"] = metadata
	if payload != "" && payload != nil {
		body["payload"] = payload
	}
	return body
}

func ResBody(ecode string, msg interface{}, payload interface{}) map[string]interface{} {
	return NewResMessage(ecode, msg, payload)
}

type PageInfo struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

func SetPageInfo(pageSize, page int, total int64) *PageInfo {
	return &PageInfo{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
}
